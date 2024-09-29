package authedapp

import (
	"github.com/gin-gonic/gin"
	"sync"
)

/*
说明：
1. 内部验证， 即内部通过rpc验证X-App-Id的真实性，并得到详情
2. 外部验证， 即网关层先验证X-App-Id的真实性，内部只是查询该id的详情
*/

type Ignorer func(c *gin.Context) bool

// Manager authed app manager
type Manager struct {
	Project        string
	provider       AppProvider
	apps           sync.Map // rq id => App
	appIdHeaderKey string
	ignoreChecker  Ignorer
	ignoreApp      App
}

// AppProvider app provider interface
type AppProvider interface {
	GetValidApp(rqId, id, project string, validate bool) (app App, err error)
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
	Attrs() map[string]string
	DefaultAttr(attr, defVal string) string
}

// New return an authed app manager
func New(project string, provider AppProvider, o ...Option) *Manager {
	s := &Manager{
		Project:        project,
		provider:       provider,
		appIdHeaderKey: "X-App-Id",
	}
	s.With(o...)
	return s
}

func NewManager(project string, provider AppProvider, o ...Option) *Manager {
	return New(project, provider, o...)
}

// Add an authed app for request id
func (m *Manager) Add(rqId string, app App) {
	m.apps.Store(rqId, app)
}

// Rm remove an authed app for request id
func (m *Manager) Rm(rqId string) {
	m.apps.Delete(rqId)
}

// Get return an authed app for request id
func (m *Manager) Get(rqId string) (app App, exist bool) {
	var v interface{}
	v, exist = m.apps.Load(rqId)
	if exist {
		app = v.(App)
	}
	return
}

// Provider return app provider
func (m *Manager) Provider() AppProvider {
	return m.provider
}

func (m *Manager) AppidHeaderKey() string {
	return m.appIdHeaderKey
}

func (m *Manager) Ignored(c *gin.Context) (bool, App) {
	if m.ignoreChecker != nil {
		return m.ignoreChecker(c), m.ignoreApp
	}
	return false, nil
}
