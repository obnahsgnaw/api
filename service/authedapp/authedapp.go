package authedapp

import (
	"github.com/obnahsgnaw/api/pkg/errobj"
	"go.uber.org/zap"
)

// Manager authed app manager
type Manager struct {
	Project         string
	debug           bool
	Backend         bool
	OutsideValidate bool
	apps            map[string]App
	provider        AppProvider
	errObjProvider  errobj.Provider
	logger          *zap.Logger
}

// AppProvider app provider interface
type AppProvider interface {
	GetValidApp(id, project string, validate, backend bool) (app App, err error)
}

// App interface
type App interface {
	Id() uint32
	AppId() string
	Name() string
	Backend() bool
	Scope() []string
}

// New return an authed app manager
func New(project string, provider AppProvider, errObjProvider errobj.Provider, outsideValidate, backend, debug bool) *Manager {
	return &Manager{
		Project:         project,
		Backend:         backend,
		debug:           debug,
		OutsideValidate: outsideValidate,
		apps:            make(map[string]App),
		provider:        provider,
		errObjProvider:  errObjProvider,
	}
}

// Add a authed app for request id
func (m *Manager) Add(rqId string, app App) {
	m.apps[rqId] = app
}

// Rm remove a authed app for request id
func (m *Manager) Rm(rqId string) {
	if _, ok := m.apps[rqId]; ok {
		delete(m.apps, rqId)
	}
}

// Get return a authed app for request id
func (m *Manager) Get(rqId string) (app App, exist bool) {
	app, exist = m.apps[rqId]
	return
}

// Provider return app provider
func (m *Manager) Provider() AppProvider {
	return m.provider
}

// ErrObjProvider return err obj provider
func (m *Manager) ErrObjProvider() errobj.Provider {
	return m.errObjProvider
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
