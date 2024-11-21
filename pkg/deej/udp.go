package deej

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/omriharel/deej/pkg/deej/util"
)

// UdpIO provides a deej-aware abstraction layer to managing UDP connections
type UdpIO struct {
	deej   *Deej
	logger *zap.SugaredLogger

	stopChannel chan bool

	currentSelectedOutputDevice int
	lastKnownNumSliders         int
	lastKnownNumMuteButtons     int
	currentSliderPercentValues  []float32
	currentMuteButtonStates     []*bool

	connection *net.UDPConn

	sliderMoveConsumers                  []chan SliderMoveEvent
	muteButtonClickConsumers             []chan MuteButtonClickEvent
	selectedOutoutDeviceChangedConsumers []chan ToggleOutoutDeviceClickEvent
}

var expectedUdpLinePattern = regexp.MustCompile(`\w+|^\d{1,4}(\|\d{1,4})*$`)

// NewUdpIO creates a UdpIO instance that uses the provided deej
// instance's connection info to establish communications with the controller
func NewUdpIO(deej *Deej, logger *zap.SugaredLogger) (*UdpIO, error) {
	logger = logger.Named("udp")

	udpio := &UdpIO{
		deej:                deej,
		logger:              logger,
		stopChannel:         make(chan bool),
		sliderMoveConsumers: []chan SliderMoveEvent{},
	}

	logger.Debug("Created UDP i/o instance")

	// respond to config changes
	udpio.setupOnConfigReload()

	return udpio, nil
}

// Start creates a UDP listener server
func (udpio *UdpIO) Start() error {
	s, err := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", udpio.deej.config.UdpConnectionInfo.UdpPort))
	if err != nil {
		udpio.logger.Warnw("Failed to resolve UDP address", "error", err)
		return fmt.Errorf("resolve udp address: %w", err)
	}

	connection, err := net.ListenUDP("udp4", s)
	if err != nil {
		udpio.logger.Warnw("Failed to start UDP listener", "error", err)
		return fmt.Errorf("start udp listener: %w", err)
	}

	udpio.connection = connection

	namedLogger := udpio.logger.Named(fmt.Sprintf(":%d", udpio.deej.config.UdpConnectionInfo.UdpPort))

	namedLogger.Infow("Connected", "conn", udpio.connection)

	// read lines or await a stop
	go func() {
		packetChannel := udpio.readPacket(namedLogger)

		for {
			select {
			case <-udpio.stopChannel:
				udpio.close(namedLogger)
			case packet := <-packetChannel:
				udpio.handlePacket(namedLogger, packet)
			}
		}
	}()

	return nil
}

func (udpio *UdpIO) readPacket(logger *zap.SugaredLogger) chan string {
	packetChannel := make(chan string)

	go func() {
		for {
			packet := make([]byte, 4096)
			bytesRead, _, err := udpio.connection.ReadFromUDP(packet)

			if err != nil {

				if udpio.deej.Verbose() {
					logger.Warnw("Failed to read UDP packet", "error", err)
				}

				return
			}

			stringData := string(packet[:bytesRead])

			if udpio.deej.Verbose() {
				logger.Debugw("Read new packet", "packet", stringData)
			}

			packetChannel <- stringData
		}
	}()

	return packetChannel
}

func (udpio *UdpIO) close(logger *zap.SugaredLogger) {
	if err := udpio.connection.Close(); err != nil {
		logger.Warnw("Failed to close UDP connection", "error", err)
	} else {
		logger.Debug("UDP connection closed")
	}

	udpio.connection = nil
}

// Stop signals us to shut down our slider connection, if one is active
func (udpio *UdpIO) Stop() {
	udpio.logger.Debug("Not currently connected, nothing to stop")
}

// SubscribeToSliderMoveEvents returns an unbuffered channel that receives
// a sliderMoveEvent struct every time a slider moves
func (udpio *UdpIO) SubscribeToSliderMoveEvents() chan SliderMoveEvent {
	ch := make(chan SliderMoveEvent)
	udpio.sliderMoveConsumers = append(udpio.sliderMoveConsumers, ch)
	return ch
}

// SubscribeToMuteButtonClickEvents returns an unbuffered channel that receives
// a sliderMoveEvent struct every time a slider moves
func (udpio *UdpIO) SubscribeToMuteButtonClickEvents() chan MuteButtonClickEvent {
	ch := make(chan MuteButtonClickEvent)
	udpio.muteButtonClickConsumers = append(udpio.muteButtonClickConsumers, ch)
	return ch
}

// SubscribeToToggleOutoutDeviceClickEvents returns an unbuffered channel that receives
// a sliderMoveEvent struct every time a slider moves
func (udpio *UdpIO) SubscribeToToggleOutoutDeviceClickEvents() chan ToggleOutoutDeviceClickEvent {
	ch := make(chan ToggleOutoutDeviceClickEvent)
	udpio.selectedOutoutDeviceChangedConsumers = append(udpio.selectedOutoutDeviceChangedConsumers, ch)
	return ch
}

func (udpio *UdpIO) setupOnConfigReload() {
	configReloadedChannel := udpio.deej.config.SubscribeToChanges()

	const stopDelay = 50 * time.Millisecond

	go func() {
		for {
			select {
			case <-configReloadedChannel:

				// make any config reload unset our slider number to ensure process volumes are being re-set
				// (the next read line will emit SliderMoveEvent instances for all sliders)\
				// this needs to happen after a small delay, because the session map will also re-acquire sessions
				// whenever the config file is reloaded, and we don't want it to receive these move events while the map
				// is still cleared. this is kind of ugly, but shouldn't cause any issues
				go func() {
					<-time.After(stopDelay)
					udpio.lastKnownNumSliders = 0
				}()

				// if connection params have changed, attempt to stop and start the connection
			}
		}
	}()
}

func (udpio *UdpIO) handlePacket(logger *zap.SugaredLogger, packet string) {
	logger.Info("got packet:", packet)
	if !expectedUdpLinePattern.MatchString(packet) {
		logger.Info("bad syntax")
		return
	}

	parts := strings.Split(packet, "|")
	if len(parts) < 2 {
		fmt.Println("Invalid input format")
		return
	}

	switch parts[0] {
	case "Sliders":
		logger.Debug("Got Sliders data: ", parts[1:])
		udpio.handleSliders(logger, parts[1:])
		return
	case "MuteButtons":
		logger.Debug("Got MuteButtons data: ", parts[1:])
		udpio.handleMuteButtons(logger, parts[1:])
		return
	case "SwitchOutput":
		logger.Debug("Got SwitchOutput data: ", parts[1:])
		if len(parts[1:]) != 1 {
			logger.Error("Invalid data, expected a single argument")
		}
		udpio.handleSwitchOutput(logger, parts[1])
		return
	default:
		logger.Warn("bad packet opcode")
		return
	}
}

func (udpio *UdpIO) handleSliders(logger *zap.SugaredLogger, data []string) {
	numSliders := len(data)

	// update our slider count, if needed - this will send slider move events for all
	if numSliders != udpio.lastKnownNumSliders {
		logger.Infow("Detected sliders", "amount", numSliders)
		udpio.lastKnownNumSliders = numSliders
		udpio.currentSliderPercentValues = make([]float32, numSliders)

		// reset everything to be an impossible value to force the slider move event later
		for idx := range udpio.currentSliderPercentValues {
			udpio.currentSliderPercentValues[idx] = -1.0
		}
	}

	// for each slider:
	moveEvents := []SliderMoveEvent{}
	for sliderIdx, stringValue := range data {

		// convert string values to integers ("1023" -> 1023)
		number, _ := strconv.Atoi(string(stringValue))

		// turns out the first line could come out dirty sometimes (i.e. "4558|925|41|643|220")
		// so let's check the first number for correctness just in case
		if sliderIdx == 0 && number > 1023 {
			udpio.logger.Debugw("Got malformed packet from UDP, ignoring", "packet", data)
			return
		}

		// map the value from raw to a "dirty" float between 0 and 1 (e.g. 0.15451...)
		dirtyFloat := float32(number) / 1023.0

		// normalize it to an actual volume scalar between 0.0 and 1.0 with 2 points of precision
		normalizedScalar := util.NormalizeScalar(dirtyFloat)

		// if sliders are inverted, take the complement of 1.0
		if udpio.deej.config.InvertSliders {
			normalizedScalar = 1 - normalizedScalar
		}

		// check if it changes the desired state (could just be a jumpy raw slider value)
		if util.SignificantlyDifferent(udpio.currentSliderPercentValues[sliderIdx], normalizedScalar, udpio.deej.config.NoiseReductionLevel) {

			// if it does, update the saved value and create a move event
			udpio.currentSliderPercentValues[sliderIdx] = normalizedScalar

			moveEvents = append(moveEvents, SliderMoveEvent{
				SliderID:     sliderIdx,
				PercentValue: normalizedScalar,
			})

			if udpio.deej.Verbose() {
				logger.Debugw("Slider moved", "event", moveEvents[len(moveEvents)-1])
			}
		}
	}

	// deliver move events if there are any, towards all potential consumers
	if len(moveEvents) > 0 {
		for _, consumer := range udpio.sliderMoveConsumers {
			for _, moveEvent := range moveEvents {
				consumer <- moveEvent
			}
		}
	}
}

func (udpio *UdpIO) handleMuteButtons(logger *zap.SugaredLogger, data []string) {
	numMuteButtons := len(data)

	// update our slider count, if needed - this will send slider move events for all
	if numMuteButtons != udpio.lastKnownNumMuteButtons {
		logger.Infow("Detected mute buttons", "amount", numMuteButtons)
		udpio.lastKnownNumMuteButtons = numMuteButtons
		udpio.currentMuteButtonStates = make([]*bool, numMuteButtons)
	}

	// for each button:
	clickEvents := []MuteButtonClickEvent{}
	for buttonIndex, muteStringValue := range data {
		muteValue, err := strconv.ParseBool(muteStringValue)
		if err != nil {
			logger.Error(err)
			continue
		}
		// Update the saved value and create a move event
		clickEvents = append(clickEvents, MuteButtonClickEvent{
			MuteButtonID: buttonIndex,
			mute:         muteValue,
		})
		logger.Debug("Mute button clicked", "event", clickEvents[len(clickEvents)-1])

	}

	// deliver move events if there are any, towards all potential consumers
	if len(clickEvents) > 0 {
		for _, consumer := range udpio.muteButtonClickConsumers {
			for _, clickEvent := range clickEvents {
				consumer <- clickEvent
			}
		}
	}
}

func (udpio *UdpIO) handleSwitchOutput(logger *zap.SugaredLogger, data string) {
	// convert output device ID to int
	outputId, _ := strconv.Atoi(string(data))
	event := ToggleOutoutDeviceClickEvent{
		selectedOutputDevice: outputId,
	}
	logger.Debugw("output device changed", "event", event)

	for _, consumer := range udpio.selectedOutoutDeviceChangedConsumers {
		consumer <- event
	}

}
