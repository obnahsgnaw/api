package autheduser

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Ignorer func(*gin.Context) bool

// Manager authed user manager
type Manager struct {
	provider        UserProvider
	logger          *zap.Logger
	users           map[string]User
	appIdHeaderKey  string
	userIdHeaderKey string
	tokenHeaderKey  string
	ignoreChecker   Ignorer
}

// User interface
type User interface {
	Id() uint32
	Uid() string
	Name() string
	Backend() bool
	Attr(attr string) (string, bool)
	Attrs() map[string]string
	DefaultAttr(attr, defVal string) string
}

// UserProvider User provider interface
type UserProvider interface {
	GetTokenUser(appid, token string) (User, error)
	GetIdUser(appid, uid string) (User, error)
}

// New return an authed user manager
func New(provider UserProvider, o ...Option) *Manager {
	s := &Manager{
		users:           make(map[string]User),
		provider:        provider,
		appIdHeaderKey:  "X-App-Id",
		userIdHeaderKey: "X-User-Id",
		tokenHeaderKey:  "Authorization",
	}
	s.With(o...)
	return s
}

func NewManager(provider UserProvider, o ...Option) *Manager {
	return New(provider, o...)
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

func (m *Manager) AppIdHeaderKey() string {
	return m.appIdHeaderKey
}

func (m *Manager) UserIdHeaderKey() string {
	return m.userIdHeaderKey
}

func (m *Manager) TokenHeaderKey() string {
	return m.tokenHeaderKey
}

func (m *Manager) Ignored(c *gin.Context) bool {
	if m.ignoreChecker != nil {
		return m.ignoreChecker(c)
	}
	return false
}
