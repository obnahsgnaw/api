package sign

import (
	"github.com/obnahsgnaw/api/pkg/errobj"
	"github.com/obnahsgnaw/application/pkg/dynamic"
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
	debug          dynamic.Bool
}

func New(provider Provider, errObjProvider errobj.Provider, debug dynamic.Bool) *Manager {
	return &Manager{
		provider:       provider,
		errObjProvider: errObjProvider,
		debug:          debug,
	}
}

// Provider return crypt provider
func (m *Manager) Provider() Provider {
	return m.provider
}

func (m *Manager) Debug() bool {
	return m.debug.Val()
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
