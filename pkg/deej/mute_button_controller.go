package deej

type MuteButtonController interface {
	Start() error
	Stop()
	SubscribeToMuteButtonClickEvents() chan MuteButtonClickEvent
}

// MuteButtonClickEvent represents a single MuteButton click captured by deej
type MuteButtonClickEvent struct {
	MuteButtonID int
	mute         bool
}
