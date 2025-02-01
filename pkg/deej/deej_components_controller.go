package deej

type MuteButtonsState struct {
	MuteButtons []bool
}
type MuteButtonConsumer func(events []MuteButtonClickEvent) (newState MuteButtonsState, err error)

type OutputDeviceState struct {
	selectedOutputDevice int
}
type ToggleOutputDeviceConsumer func(event ToggleOutoutDeviceClickEvent) (newState OutputDeviceState, err error)

type DeejSlidersController interface {
	Start() error
	Stop()
	SubscribeToSliderMoveEvents() chan SliderMoveEvent
}

type DeejButtonsController interface {
	Start() error
	Stop()
	setMuteButtonClickEventConsumer(MuteButtonConsumer)
	setToggleOutputDeviceEventConsumer(ToggleOutputDeviceConsumer)
}

// SliderMoveEvent represents a single slider move captured by deej
type SliderMoveEvent struct {
	SliderID     int
	PercentValue float32
}

// ToggleOutoutDeviceClickEvent represents a single ToggleOutputDevice click captured by deej
type ToggleOutoutDeviceClickEvent struct {
	selectedOutputDevice int
}

// MuteButtonClickEvent represents a single MuteButton click captured by deej
type MuteButtonClickEvent struct {
	MuteButtonID int
	mute         bool
}
