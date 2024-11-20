package deej

type ToggleOutoutDeviceButtonController interface {
	Start() error
	Stop()
	SubscribeToToggleOutoutDeviceClickEvents() chan ToggleOutoutDeviceClickEvent
}

// MuteButtonClickEvent represents a single MuteButton click captured by deej
type ToggleOutoutDeviceClickEvent struct {
	selectedOutputDevice int
}
