package sign

import (
	"github.com/obnahsgnaw/api/pkg/errobj"
	"go.uber.org/zap"
)

type Provider interface {
	Validate(appid, uid, method, uri, signature, timestamp, nonce string) error
	Generate(appid, uid, method, uri string) (signature, timestamp, nonce string, err error)
}

type Manager struct {
	logger         *zap.Logger
	provider       Provider
	errObjProvider errobj.Provider
	debug          bool
}

func New(provider Provider, errObjProvider errobj.Provider) *Manager {
	return &Manager{
		provider:       provider,
		errObjProvider: errObjProvider,
		debug:          false,
	}
}

// Provider return crypt provider
func (m *Manager) Provider() Provider {
	return m.provider
}

func (m *Manager) SetDebug(enable bool) {
	m.debug = enable
}

func (m *Manager) Debug() bool {
	return m.debug
}

func (m *Manager) ErrObjProvider() errobj.Provider {
	return m.errObjProvider
}

func (m *Manager) SetLogger(logger *zap.Logger) {
	m.logger = logger
}

func (m *Manager) Logger() *zap.Logger {
	return m.logger
}
