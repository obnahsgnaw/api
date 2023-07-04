package perm

import (
	"github.com/obnahsgnaw/api/pkg/errobj"
	"go.uber.org/zap"
)

type Provider interface {
	Can(appid, uid, method, pattern string) error
}

type Manager struct {
	debug          bool
	provider       Provider
	logger         *zap.Logger
	errObjProvider errobj.Provider
}

func New(p Provider, ebj errobj.Provider) *Manager {
	return &Manager{
		debug:          false,
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

func (m *Manager) SetDebug(enable bool) {
	m.debug = enable
}

func (m *Manager) Debug() bool {
	return m.debug
}

func (m *Manager) Provider() Provider {
	return m.provider
}

func (m *Manager) ErrObjProvider() errobj.Provider {
	return m.errObjProvider
}
