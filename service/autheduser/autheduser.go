package autheduser

import (
	"github.com/obnahsgnaw/api/internal/server/authroute"
	"github.com/obnahsgnaw/api/pkg/errobj"
	"go.uber.org/zap"
)

// Manager authed user manager
type Manager struct {
	Backend         bool
	OutsideValidate bool
	debug           bool
	users           map[string]User
	provider        UserProvider
	errObjProvider  errobj.Provider
	authManager     *authroute.Manager
	logger          *zap.Logger
}

// User interface
type User interface {
	Id() uint32
	Uid() string
	Name() string
}

// UserProvider User provider interface
type UserProvider interface {
	GetValidTokenUser(appid, token string) (User, error)
	GetValidUser(uid string) (User, error)
	GetJwtKey(appid, uid string) (string, error)
}

// New return an authed user manager
func New(provider UserProvider, errObjProvider errobj.Provider, authManager *authroute.Manager, outsideValidate, backend bool) *Manager {
	return &Manager{
		Backend:         backend,
		OutsideValidate: outsideValidate,
		users:           make(map[string]User),
		provider:        provider,
		errObjProvider:  errObjProvider,
		authManager:     authManager,
	}
}

// Add an authed app for request id
func (m *Manager) Add(rqId string, user User) {
	m.users[rqId] = user
}

// Rm remove an authed app for request id
func (m *Manager) Rm(rqId string) {
	if _, ok := m.users[rqId]; ok {
		delete(m.users, rqId)
	}
}

// Get return an authed app for request id
func (m *Manager) Get(rqId string) (user User, exist bool) {
	user, exist = m.users[rqId]
	return
}

// Provider return user provider
func (m *Manager) Provider() UserProvider {
	return m.provider
}

func (m *Manager) SetDebug(enable bool) {
	m.debug = enable
}

func (m *Manager) Debug() bool {
	return m.debug
}

// ErrObjProvider return err obj provider
func (m *Manager) ErrObjProvider() errobj.Provider {
	return m.errObjProvider
}

// AuthedRouteManager return auth route manager
func (m *Manager) AuthedRouteManager() *authroute.Manager {
	return m.authManager
}

func (m *Manager) SetLogger(logger *zap.Logger) {
	m.logger = logger
}

func (m *Manager) Logger() *zap.Logger {
	return m.logger
}
