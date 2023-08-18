package authedapp

import (
	"github.com/obnahsgnaw/api/pkg/errobj"
	"github.com/obnahsgnaw/application/pkg/dynamic"
	"go.uber.org/zap"
)

// Manager authed app manager
type Manager struct {
	Project        string
	debug          dynamic.Bool
	Backend        bool
	outsideHandler *OutsideHandler
	apps           map[string]App
	provider       AppProvider
	errObjProvider errobj.Provider
	logger         *zap.Logger
}

type OutsideHandler struct {
	Key    string
	Decode func([]byte) (App, error)
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
	Attr(attr string) (interface{}, bool)
}

// New return an authed app manager
func New(project string, provider AppProvider, errObjProvider errobj.Provider, backend bool, debug dynamic.Bool) *Manager {
	return &Manager{
		Project:        project,
		Backend:        backend,
		debug:          debug,
		apps:           make(map[string]App),
		provider:       provider,
		errObjProvider: errObjProvider,
	}
}

// NewOutside return an outside authed app manager
func NewOutside(project string, provider *OutsideHandler, errObjProvider errobj.Provider, debug dynamic.Bool) *Manager {
	return &Manager{
		Project:        project,
		debug:          debug,
		apps:           make(map[string]App),
		errObjProvider: errObjProvider,
		outsideHandler: provider,
	}
}

// Add an authed app for request id
func (m *Manager) Add(rqId string, app App) {
	m.apps[rqId] = app
}

// Rm remove an authed app for request id
func (m *Manager) Rm(rqId string) {
	if _, ok := m.apps[rqId]; ok {
		delete(m.apps, rqId)
	}
}

// Get return an authed app for request id
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

func (m *Manager) Debug() bool {
	return m.debug.Val()
}

func (m *Manager) OutsideValidate() bool {
	return m.outsideHandler != nil
}

func (m *Manager) OutsideHandler() *OutsideHandler {
	return m.outsideHandler
}
