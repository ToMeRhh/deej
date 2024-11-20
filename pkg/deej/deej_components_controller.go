package deej

type DeejComponentsController interface {
	Start() error
	Stop()
	SubscribeToSliderMoveEvents() chan SliderMoveEvent
	SubscribeToMuteButtonClickEvents() chan MuteButtonClickEvent
	SubscribeToToggleOutoutDeviceClickEvents() chan ToggleOutoutDeviceClickEvent
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
