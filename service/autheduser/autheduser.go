package autheduser

import (
	"go.uber.org/zap"
)

// Manager authed user manager
type Manager struct {
	provider        UserProvider
	logger          *zap.Logger
	users           map[string]User
	outsideValidate func() bool
	appIdHeaderKey  string
	userIdHeaderKey string
	tokenHeaderKey  string
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
	GetIdUser(appid, uid string) (User, error)
}

// New return an authed user manager
func New(provider UserProvider) *Manager {
	return &Manager{
		users:    make(map[string]User),
		provider: provider,
		outsideValidate: func() bool {
			return false
		},
		appIdHeaderKey:  "X-App-Id",
		userIdHeaderKey: "X-User-Id",
		tokenHeaderKey:  "Authorization",
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

func (m *Manager) SetOutsideValidate(cb func() bool) {
	m.outsideValidate = cb
}

func (m *Manager) OutsideValidate() bool {
	return m.outsideValidate()
}

func (m *Manager) SetAppIdHeaderKey(key string) {
	m.appIdHeaderKey = key
}
func (m *Manager) AppIdHeaderKey() string {
	return m.appIdHeaderKey
}

func (m *Manager) SetUserIdHeaderKey(key string) {
	m.userIdHeaderKey = key
}
func (m *Manager) UserIdHeaderKey() string {
	return m.userIdHeaderKey
}

func (m *Manager) SetTokenHeaderKey(key string) {
	m.tokenHeaderKey = key
}
func (m *Manager) TokenHeaderKey() string {
	return m.tokenHeaderKey
}
