package perm

import (
	"github.com/obnahsgnaw/api/pkg/errobj"
	"github.com/obnahsgnaw/application/pkg/dynamic"
	"go.uber.org/zap"
)

type Provider interface {
	Can(appid, uid, method, pattern string) error
}

type Manager struct {
	debug          dynamic.Bool
	provider       Provider
	logger         *zap.Logger
	errObjProvider errobj.Provider
}

func New(p Provider, ebj errobj.Provider, debug dynamic.Bool) *Manager {
	return &Manager{
		debug:          debug,
		provider:       p,
		logger:         nil,
		errObjProvider: ebj,
	}
}

func (m *Manager) SetLogger(logger *zap.Logger) {
	m.logger = logger
}

func (m *Manager) Logger() *zap.Logger {
	return m.logger
}

func (m *Manager) Debug() bool {
	return m.debug.Val()
}

func (m *Manager) Provider() Provider {
	return m.provider
}

func (m *Manager) ErrObjProvider() errobj.Provider {
	return m.errObjProvider
}
