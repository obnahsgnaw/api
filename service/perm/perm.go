package perm

type Provider interface {
	Can(appid, uid, method, pattern string) error
}

type Manager struct {
	provider        Provider
	appIdHeaderKey  string
	userIdHeaderKey string
}

func New(p Provider) *Manager {
	return &Manager{
		provider:        p,
		appIdHeaderKey:  "X-App-Id",
		userIdHeaderKey: "X-User-Id",
	}
}

func (m *Manager) Provider() Provider {
	return m.provider
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
