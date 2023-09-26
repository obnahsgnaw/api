package authedapp

import (
	"github.com/obnahsgnaw/api/pkg/errobj"
	"github.com/obnahsgnaw/application/pkg/dynamic"
	"go.uber.org/zap"
)

/*
说明：
1. 内部验证， 即内部通过rpc验证X-App-Id的真实性，并得到详情
2. 外部验证， 即网关层先验证X-App-Id的真实性，内部只是查询该id的详情
*/

// Manager authed app manager
type Manager struct {
	Project          string
	debug            dynamic.Bool
	Backend          bool
	outsideValidate  func() bool
	apps             map[string]App
	provider         AppProvider
	errObjProvider   errobj.Provider
	logger           *zap.Logger
	appIdBfHeaderKey string
	appIdAfHeaderKey string
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
	Manage() bool
	Attr(attr string) (string, bool)
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

func (m *Manager) Outside(cb func() bool) {
	m.outsideValidate = cb
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
	if m.outsideValidate == nil {
		return false
	}
	return m.outsideValidate()
}

func (m *Manager) SetAppidHeaderKey(key string) {
	m.appIdBfHeaderKey = key
}
func (m *Manager) SetAuthedAppidHeaderKey(key string) {
	m.appIdAfHeaderKey = key
}
func (m *Manager) GetAppidHeaderKey() string {
	if m.appIdBfHeaderKey == "" {
		return "X-App-Id"
	}

	return m.appIdBfHeaderKey
}
func (m *Manager) GetAuthedAppidHeaderKey() string {
	if m.appIdAfHeaderKey == "" {
		return "X-App-Id"
	}

	return m.appIdAfHeaderKey
}
