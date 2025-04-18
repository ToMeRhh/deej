package deej

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	ole "github.com/go-ole/go-ole"
	"github.com/thoas/go-funk"
	"github.com/tomerhh/deej/pkg/deej/util"
	"go.uber.org/zap"
)

type sessionMap struct {
	deej   *Deej
	logger *zap.SugaredLogger

	m    map[string][]Session
	lock sync.Locker

	sessionFinder SessionFinder

	lastSessionRefresh time.Time
	unmappedSessions   []Session
}

const (
	masterSessionName = "master" // master device volume
	systemSessionName = "system" // system sounds volume
	inputSessionName  = "mic"    // microphone input level

	// some targets need to be transformed before their correct audio sessions can be accessed.
	// this prefix identifies those targets to ensure they don't contradict with another similarly-named process
	specialTargetTransformPrefix = "deej."

	// targets the currently active window (Windows-only, experimental)
	specialTargetCurrentWindow = "current"

	// targets all currently unmapped sessions (experimental)
	specialTargetAllUnmapped = "unmapped"

	// this threshold constant assumes that re-acquiring all sessions is a kind of expensive operation,
	// and needs to be limited in some manner. this value was previously user-configurable through a config
	// key "process_refresh_frequency", but exposing this type of implementation detail seems wrong now
	minTimeBetweenSessionRefreshes = time.Second * 5

	// determines whether the map should be refreshed when a slider moves.
	// this is a bit greedy but allows us to ensure sessions are always re-acquired, which is
	// especially important for process groups (because you can have one ongoing session
	// always preventing lookup of other processes bound to its slider, which forces the user
	// to manually refresh sessions). a cleaner way to do this down the line is by registering to notifications
	// whenever a new session is added, but that's too hard to justify for how easy this solution is
	maxTimeBetweenSessionRefreshes = time.Second * 45
)

// this matches friendly device names (on Windows), e.g. "Headphones (Realtek Audio)"
var deviceSessionKeyPattern = regexp.MustCompile(`^.+ \(.+\)$`)

func newSessionMap(deej *Deej, logger *zap.SugaredLogger, sessionFinder SessionFinder) (*sessionMap, error) {
	logger = logger.Named("sessions")

	m := &sessionMap{
		deej:          deej,
		logger:        logger,
		m:             make(map[string][]Session),
		lock:          &sync.Mutex{},
		sessionFinder: sessionFinder,
	}

	logger.Debug("Created session map instance")

	return m, nil
}

func (m *sessionMap) initialize() error {
	if err := m.getAndAddSessions(); err != nil {
		m.logger.Warnw("Failed to get all sessions during session map initialization", "error", err)
		return fmt.Errorf("get all sessions during init: %w", err)
	}

	m.setupOnConfigReload()
	m.setupOnSliderMove()
	m.setupOnMuteButtonClicked()
	m.setupOnToggleOutputDeviceButtonClicked()

	return nil
}

func (m *sessionMap) release() error {
	if err := m.sessionFinder.Release(); err != nil {
		m.logger.Warnw("Failed to release session finder during session map release", "error", err)
		return fmt.Errorf("release session finder during release: %w", err)
	}

	return nil
}

// assumes the session map is clean!
// only call on a new session map or as part of refreshSessions which calls reset
func (m *sessionMap) getAndAddSessions() error {

	// mark that we're refreshing before anything else
	m.lastSessionRefresh = time.Now()
	m.unmappedSessions = nil

	sessions, err := m.sessionFinder.GetAllSessions()
	if err != nil {
		m.logger.Warnw("Failed to get sessions from session finder", "error", err)
		return fmt.Errorf("get sessions from SessionFinder: %w", err)
	}

	for _, session := range sessions {
		m.add(session)

		if !m.sessionMapped(session) {
			m.unmappedSessions = append(m.unmappedSessions, session)
		}
	}

	m.logger.Infow("Got all audio sessions successfully", "sessionMap", m)

	return nil
}

func (m *sessionMap) setupOnConfigReload() {
	configReloadedChannel := m.deej.config.SubscribeToChanges()

	go func() {
		for {
			select {
			case <-configReloadedChannel:
				m.logger.Info("Detected config reload, attempting to re-acquire all audio sessions")
				m.refreshSessions(false)
			}
		}
	}()
}

func (m *sessionMap) setupOnSliderMove() {
	eventsChannel := m.deej.deejSlidersController.SubscribeToSliderMoveEvents()

	go func() {
		for event := range eventsChannel {
			m.handleSliderMoveEvent(event)
		}
	}()
}

func (m *sessionMap) setupOnMuteButtonClicked() {
	m.deej.deejButtonsController.setMuteButtonClickEventConsumer(m.handleMuteButtonClickedEventsAndGetState)
}

func (m *sessionMap) setupOnToggleOutputDeviceButtonClicked() {
	m.deej.deejButtonsController.setToggleOutputDeviceEventConsumer(m.handleToggleOutputDeviceClickedEventAndGetState)
}

// performance: explain why force == true at every such use to avoid unintended forced refresh spams
func (m *sessionMap) refreshSessions(force bool) {

	// make sure enough time passed since the last refresh, unless force is true in which case always clear
	if !force && m.lastSessionRefresh.Add(minTimeBetweenSessionRefreshes).After(time.Now()) {
		return
	}

	// clear and release sessions first
	m.clear()

	if err := m.getAndAddSessions(); err != nil {
		m.logger.Warnw("Failed to re-acquire all audio sessions", "error", err)
	} else {
		m.logger.Debug("Re-acquired sessions successfully")
	}
}

// returns true if a session is not currently mapped to any slider, false otherwise
// special sessions (master, system, mic) and device-specific sessions always count as mapped,
// even when absent from the config. this makes sense for every current feature that uses "unmapped sessions"
func (m *sessionMap) sessionMapped(session Session) bool {

	// count master/system/mic as mapped
	if funk.ContainsString([]string{masterSessionName, systemSessionName, inputSessionName}, session.Key()) {
		return true
	}

	// count device sessions as mapped
	if deviceSessionKeyPattern.MatchString(session.Key()) {
		return true
	}

	matchFound := false

	// look through the actual mappings
	m.deej.config.SliderMapping.iterate(func(sliderIdx int, targets []string) {
		for _, target := range targets {

			// ignore special transforms
			if m.targetHasSpecialTransform(target) {
				continue
			}

			// safe to assume this has a single element because we made sure there's no special transform
			target = m.resolveTarget(target)[0]

			if target == session.Key() {
				matchFound = true
				return
			}
		}
	})

	return matchFound
}

func (m *sessionMap) maybeRefreshSessions() {
	// first of all, ensure our session map isn't moldy
	if m.lastSessionRefresh.Add(maxTimeBetweenSessionRefreshes).Before(time.Now()) {
		m.logger.Debug("Stale session map detected on slider move, refreshing")
		m.refreshSessions(true)
	}
}

func (m *sessionMap) handleSliderMoveEvent(event SliderMoveEvent) {

	m.maybeRefreshSessions()

	// get the targets mapped to this slider from the config
	targets, ok := m.deej.config.SliderMapping.get(event.SliderID)

	// if slider not found in config, silently ignore
	if !ok {
		m.logger.Warn("Ignoring data for unmapped slider (%d)", event.SliderID)
		return
	}

	targetFound := false
	adjustmentFailed := false

	// for each possible target for this slider...
	for _, target := range targets {

		// resolve the target name by cleaning it up and applying any special transformations.
		// depending on the transformation applied, this can result in more than one target name
		resolvedTargets := m.resolveTarget(target)

		// for each resolved target...
		for _, resolvedTarget := range resolvedTargets {

			// check the map for matching sessions
			sessions, ok := m.get(resolvedTarget)

			// no sessions matching this target - move on
			if !ok {
				continue
			}

			targetFound = true

			// iterate all matching sessions and adjust the volume of each one
			for _, session := range sessions {
				if session.GetVolume() != event.PercentValue {
					if err := session.SetVolume(event.PercentValue); err != nil {
						m.logger.Warnw("Failed to set target session volume", "error", err)
						adjustmentFailed = true
					}
				}
			}
		}
	}

	// if we still haven't found a target or the volume adjustment failed, maybe look for the target again.
	// processes could've opened since the last time this slider moved.
	// if they haven't, the cooldown will take care to not spam it up
	if !targetFound {
		m.refreshSessions(false)
	} else if adjustmentFailed {

		// performance: the reason that forcing a refresh here is okay is that we'll only get here
		// when a session's SetVolume call errored, such as in the case of a stale master session
		// (or another, more catastrophic failure happens)
		m.refreshSessions(true)
	}
}

func (m *sessionMap) handleMuteButtonClickedEventsAndGetState(events []MuteButtonClickEvent) (newState MuteButtonsState, err error) {
	m.maybeRefreshSessions()
	m.logger.Infow("Handling mute events", "events", events, "len", len(events))

	// get the targets mapped to this buttons from the config
	targets_arr := make([][]string, len(events))
	for event_index, event := range events {
		targets, ok := m.deej.config.MuteButtonMapping.get(event.MuteButtonID)
		if !ok {
			// if a button is not found in config, silently ignore
			m.logger.Warn("Ignoring data for unmapped button (%d)", event.MuteButtonID)
			continue
		}
		targets_arr[event_index] = targets
	}
	m.logger.Infow("targets:", "targets", targets_arr)

	targetFound := false
	adjustmentFailed := false
	ret := MuteButtonsState{MuteButtons: make([]bool, len(events))}

	// for each possible target for this slider...
	for event_index, targets := range targets_arr {
		for _, target := range targets {

			// resolve the target name by cleaning it up and applying any special transformations.
			// depending on the transformation applied, this can result in more than one target name
			resolvedTargets := m.resolveTarget(target)

			// for each resolved target...
			for _, resolvedTarget := range resolvedTargets {

				// check the map for matching sessions
				sessions, ok := m.get(resolvedTarget)
				m.logger.Infof("testing target: %s", sessions)

				// no sessions matching this target - move on
				if !ok {
					continue
				}

				targetFound = true

				// iterate all matching sessions and adjust the mute state of each one
				for _, session := range sessions {
					if err := session.SetMute(events[event_index].mute); err != nil {
						m.logger.Warnw("Failed to set target session mute state", "error", err)
						adjustmentFailed = true
						ret.MuteButtons[event_index] = session.GetMute()
					} else {
						ret.MuteButtons[event_index] = events[event_index].mute
					}
				}
			}
		}
	}

	// if we still haven't found a target or the volume adjustment failed, maybe look for the target again.
	// processes could've opened since the last time this slider moved.
	// if they haven't, the cooldown will take care to not spam it up
	if !targetFound {
		m.refreshSessions(false)
	} else if adjustmentFailed {

		// performance: the reason that forcing a refresh here is okay is that we'll only get here
		// when a session's SetVolume call errored, such as in the case of a stale master session
		// (or another, more catastrophic failure happens)
		m.refreshSessions(true)
	}
	return ret, nil
}

func (m *sessionMap) handleToggleOutputDeviceClickedEventAndGetState(event ToggleOutoutDeviceClickEvent) (newState OutputDeviceState, err error) {
	m.maybeRefreshSessions()

	// get the device friendly name of the target device to toggle to
	selectedDeviceFriendlyName, ok := m.deej.config.AvailableOutputDeviceMapping.get(event.selectedOutputDevice)
	if !ok {
		m.logger.Warn("Ignoring data for unknown output device (%d)", event.selectedOutputDevice)
		return OutputDeviceState{}, fmt.Errorf("Ignoring data for unknown output device (%d) %w", event.selectedOutputDevice, err)
	} else if len(selectedDeviceFriendlyName) != 1 {
		m.logger.Warn("Multiple output device toggeling is not supported (%d), %s", event.selectedOutputDevice, selectedDeviceFriendlyName)
		return OutputDeviceState{}, fmt.Errorf("config consists of multiple output devices to toggle %w", err)
	}

	// get the UUID of the target device to toggle to
	selectedDevice, err := util.GetDeviceIDByNameWinAPI(selectedDeviceFriendlyName[0])

	// if the error is "Incorrect function" that corresponds to 0x00000001,
	// which represents E_FALSE in COM error handling. this is fine for this function,
	// and just means that the call was redundant.
	const eFalse = 1
	oleError := &ole.OleError{}

	if errors.As(err, &oleError) {
		if oleError.Code() == eFalse {
			m.logger.Warnf("CoInitializeEx failed with E_FALSE due to redundant invocation %w", err)
		} else {
			m.logger.Warnw("Failed to call CoInitializeEx",
				"isOleError", true,
				"error", err,
				"oleError", oleError,
				"err", err)

			return OutputDeviceState{}, fmt.Errorf("call CoInitializeEx: %w", err)
		}
	}
	if err != nil {
		m.logger.Warnw("Failed to get device ID by name", "error", err)
		return OutputDeviceState{}, fmt.Errorf("failed to get device ID by name: %w", err)
	}

	m.logger.Infof("Changing selected device to: %s (%s)", selectedDevice, selectedDeviceFriendlyName[0])
	res := util.SetAudioDeviceByID(selectedDevice, m.logger)
	m.refreshSessions(true)
	if res {
		return OutputDeviceState(event), nil
	}
	out, _, err := m.sessionFinder.getDefaultAudioEndpoints()
	if err != nil {
		return OutputDeviceState{selectedOutputDevice: -1}, nil
	}
	var outDeviceId string
	out.GetId(&outDeviceId)
	for key, ids := range m.deej.config.AvailableOutputDeviceMapping.m {
		for _, id := range ids {
			if id == outDeviceId {
				return OutputDeviceState{selectedOutputDevice: key}, nil
			}
		}
	}
	return OutputDeviceState{selectedOutputDevice: -1}, nil
}

func (m *sessionMap) targetHasSpecialTransform(target string) bool {
	return strings.HasPrefix(target, specialTargetTransformPrefix)
}

func (m *sessionMap) resolveTarget(target string) []string {

	// start by ignoring the case
	target = strings.ToLower(target)

	// look for any special targets first, by examining the prefix
	if m.targetHasSpecialTransform(target) {
		return m.applyTargetTransform(strings.TrimPrefix(target, specialTargetTransformPrefix))
	}

	return []string{target}
}

func (m *sessionMap) applyTargetTransform(specialTargetName string) []string {

	// select the transformation based on its name
	switch specialTargetName {

	// get current active window
	case specialTargetCurrentWindow:
		currentWindowProcessNames, err := util.GetCurrentWindowProcessNames()

		// silently ignore errors here, as this is on deej's "hot path" (and it could just mean the user's running linux)
		if err != nil {
			return nil
		}

		// we could have gotten a non-lowercase names from that, so let's ensure we return ones that are lowercase
		for targetIdx, target := range currentWindowProcessNames {
			currentWindowProcessNames[targetIdx] = strings.ToLower(target)
		}

		// remove dupes
		return funk.UniqString(currentWindowProcessNames)

	// get currently unmapped sessions
	case specialTargetAllUnmapped:
		targetKeys := make([]string, len(m.unmappedSessions))
		for sessionIdx, session := range m.unmappedSessions {
			targetKeys[sessionIdx] = session.Key()
		}

		return targetKeys
	}

	return nil
}

func (m *sessionMap) add(value Session) {
	m.lock.Lock()
	defer m.lock.Unlock()

	key := value.Key()

	existing, ok := m.m[key]
	if !ok {
		m.m[key] = []Session{value}
	} else {
		m.m[key] = append(existing, value)
	}
}

func (m *sessionMap) get(key string) ([]Session, bool) {
	m.lock.Lock()
	defer m.lock.Unlock()

	value, ok := m.m[key]
	return value, ok
}

func (m *sessionMap) clear() {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.logger.Debug("Releasing and clearing all audio sessions")

	for key, sessions := range m.m {
		for _, session := range sessions {
			session.Release()
		}

		delete(m.m, key)
	}

	m.logger.Debug("Session map cleared")
}

func (m *sessionMap) String() string {
	m.lock.Lock()
	defer m.lock.Unlock()

	sessionCount := 0

	for _, value := range m.m {
		sessionCount += len(value)
	}

	return fmt.Sprintf("<%d audio sessions>", sessionCount)
}
