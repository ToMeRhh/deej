package deej

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tomerhh/deej/pkg/deej/util"
	"go.uber.org/zap"
)

// TcpIO provides a deej-aware abstraction layer to managing UDP connections
type TcpIO struct {
	deej   *Deej
	logger *zap.SugaredLogger

	stopChannel chan bool

	lastKnownNumMuteButtons int

	connection *net.TCPListener

	muteButtonClickConsumer    MuteButtonConsumer
	toggleOutputDeviceConsumer ToggleOutputDeviceConsumer
}

var expectedLinePattern = regexp.MustCompile(`\w+|^\d{1,4}(\|\d{1,4})*$`)

// NewTcpIO creates a TcpIO instance that uses the provided deej
// instance's connection info to establish communications with the controller
func NewTcpIO(deej *Deej, logger *zap.SugaredLogger) (*TcpIO, error) {
	logger = logger.Named("tcp")

	tcpio := &TcpIO{
		deej:        deej,
		logger:      logger,
		stopChannel: deej.stopChannel,
	}

	logger.Debug("Created TCP i/o instance")

	// respond to config changes
	tcpio.setupOnConfigReload()

	return tcpio, nil
}

// Start creates a UDP listener server
func (tcpio *TcpIO) Start() error {
	s, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf(":%d", tcpio.deej.config.TcpConnectionInfo.TcpPort))
	if err != nil {
		tcpio.logger.Warnw("Failed to resolve TCP address", "error", err)
		return fmt.Errorf("resolve tcp address: %w", err)
	}

	connection, err := net.ListenTCP("tcp4", s)
	if err != nil {
		tcpio.logger.Warnw("Failed to start UDP listener", "error", err)
		return fmt.Errorf("start udp listener: %w", err)
	}

	tcpio.connection = connection

	tcpio.logger.Infow("Connected", "conn", tcpio.connection)

	// read lines or await a stop
	go func() {
		packetChannel := tcpio.readPacket(tcpio.logger)

		for {
			select {
			case <-tcpio.stopChannel:
				tcpio.close(tcpio.logger)
			case packet := <-packetChannel:
				tcpio.handleConnection(tcpio.logger, packet)
			}
		}
	}()

	return nil
}

func (tcpio *TcpIO) readPacket(logger *zap.SugaredLogger) chan net.Conn {
	connChannel := make(chan net.Conn)

	go func() {
		for {
			conn, err := tcpio.connection.Accept()
			if err != nil {
				if tcpio.deej.Verbose() {
					logger.Warnw("Failed to read TCP packet", "error", err)
				}

				return
			}
			if tcpio.deej.Verbose() {
				logger.Debugw("Read new packet")
			}

			connChannel <- conn
		}
	}()

	return connChannel
}

func (tcpio *TcpIO) close(logger *zap.SugaredLogger) {
	if err := tcpio.connection.Close(); err != nil {
		logger.Warnw("Failed to close TCP connection", "error", err)
	} else {
		logger.Debug("TCP connection closed")
	}

	tcpio.connection = nil
}

// Stop signals us to shut down our slider connection, if one is active
func (tcpio *TcpIO) Stop() {
	tcpio.logger.Debug("Not currently connected, nothing to stop")
}

// Sets the given consumer as the one to receive mute button click events
func (tcpio *TcpIO) setMuteButtonClickEventConsumer(f MuteButtonConsumer) {
	tcpio.muteButtonClickConsumer = f
}

// Sets the given consumer as the one to receive toggle output device click events
func (tcpio *TcpIO) setToggleOutputDeviceEventConsumer(f ToggleOutputDeviceConsumer) {
	tcpio.toggleOutputDeviceConsumer = f
}

func (tcpio *TcpIO) setupOnConfigReload() {
	configReloadedChannel := tcpio.deej.config.SubscribeToChanges()

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
					tcpio.lastKnownNumMuteButtons = 0
				}()

				// if connection params have changed, attempt to stop and start the connection
			}
		}
	}()
}

func (tcpio *TcpIO) handleConnection(logger *zap.SugaredLogger, conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	conn.SetDeadline(time.Now().Add(5 * time.Second))
	request, err := reader.ReadString('\n') // Read the request (up to newline)
	if err != nil {
		log.Println("Error reading request:", err)
		return
	}
	request = strings.Trim(request, "\n")
	logger.Debug("Received request:", request)

	if !expectedLinePattern.MatchString(request) {
		logger.Info("bad syntax")
		return
	}

	parts := strings.Split(request, "|")
	if len(parts) < 2 {
		fmt.Println("Invalid input format")
		return
	}
	response := ""
	switch parts[0] {
	case "MuteButtons":
		logger.Debug("Got MuteButtons data: ", parts[1:])
		response = tcpio.handleMuteButtons(logger, parts[1:])
	case "SwitchOutput":
		logger.Debug("Got SwitchOutput data: ", parts[1:])
		if len(parts[1:]) != 1 {
			logger.Error("Invalid data, expected a single argument")
		}
		response = tcpio.handleSwitchOutput(logger, parts[1])
	case "GetCurrentOutputDevice":
		currentDeviceId, err := util.GetCurrentAudioDeviceID()
		if err != nil {
			logger.Error("Failed to get current audio device ID: ", err)
			return
		}

		currentDeviceFriendlyName, err := util.GetDeviceFriendlyNameByIdExec(currentDeviceId)
		if err != nil {
			logger.Error("Failed to get current audio device name: ", err)
			return
		}
		logger.Debug("Got GetCurrentOutputDevice data: ", currentDeviceId, currentDeviceFriendlyName)

		response = "-1"
		for deviceIndex, deviceName := range tcpio.deej.config.AvailableOutputDeviceMapping.m {
			if deviceName[0] == currentDeviceFriendlyName {
				logger.Debug("Found matching device: ", deviceName)
				response = fmt.Sprintf("%d", deviceIndex)
			}
		}
	default:
		logger.Warn("bad packet opcode")
	}
	conn.Write([]byte(response + "\n"))
}

func (tcpio *TcpIO) handleMuteButtons(logger *zap.SugaredLogger, data []string) string {
	numMuteButtons := len(data)

	// update our slider count, if needed - this will send slider move events for all
	if numMuteButtons != tcpio.lastKnownNumMuteButtons {
		logger.Infow("Detected mute buttons", "amount", numMuteButtons)
		tcpio.lastKnownNumMuteButtons = numMuteButtons
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
		logger.Debug("Mute button clicked: ", "event", clickEvents[len(clickEvents)-1])
	}

	// if we have no events, return an empty string
	if len(clickEvents) == 0 {
		return ""
	}

	// Run the consumer and convert the returned new state to string
	res, _ := tcpio.muteButtonClickConsumer(clickEvents)
	strSlice := make([]string, len(res.MuteButtons))
	for i, b := range res.MuteButtons {
		strSlice[i] = fmt.Sprint(b)
	}
	return strings.Join(strSlice, "|")
}

func (tcpio *TcpIO) handleSwitchOutput(logger *zap.SugaredLogger, data string) string {
	// convert output device ID to int
	outputId, _ := strconv.Atoi(string(data))
	event := ToggleOutoutDeviceClickEvent{
		selectedOutputDevice: outputId,
	}
	logger.Debugw("output device changed", "event", event.selectedOutputDevice)
	res, _ := tcpio.toggleOutputDeviceConsumer(event)
	logger.Debugw("output device changed", "newState", res)
	return strconv.Itoa(res.selectedOutputDevice)
}
