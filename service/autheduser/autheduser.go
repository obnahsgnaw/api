package autheduser

import (
	"github.com/obnahsgnaw/api/internal/server/authroute"
	"github.com/obnahsgnaw/api/pkg/errobj"
	"github.com/obnahsgnaw/application/pkg/dynamic"
	"go.uber.org/zap"
)

// Manager authed user manager
type Manager struct {
	Backend         bool
	outsideValidate bool
	debug           dynamic.Bool
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
	Backend() bool
	Attr(attr string) (string, bool)
}

// UserProvider User provider interface
type UserProvider interface {
	GetTokenUser(appid, token string) (User, error)
	GetIdUser(uid string) (User, error)
}

// New return an authed user manager
func New(provider UserProvider, errObjProvider errobj.Provider, authManager *authroute.Manager, backend bool, debug dynamic.Bool) *Manager {
	return &Manager{
		Backend:        backend,
		users:          make(map[string]User),
		provider:       provider,
		errObjProvider: errObjProvider,
		authManager:    authManager,
		debug:          debug,
	}
}

func (m *Manager) Outside() {
	m.outsideValidate = true
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

func (m *Manager) Debug() bool {
	return m.debug.Val()
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

func (m *Manager) OutsideValidate() bool {
	return m.outsideValidate
}
