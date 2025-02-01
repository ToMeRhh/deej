package deej

import wca "github.com/moutend/go-wca/pkg/wca"

// SessionFinder represents an entity that can find all current audio sessions
type SessionFinder interface {
	GetAllSessions() ([]Session, error)
	getDefaultAudioEndpoints() (*wca.IMMDevice, *wca.IMMDevice, error)

	Release() error
}
