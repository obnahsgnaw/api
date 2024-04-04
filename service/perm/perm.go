package perm

type Provider interface {
	Can(appid, uid, method, pattern string) error
}

type Ignorer func(method, pattern string) bool

type Manager struct {
	provider        Provider
	appIdHeaderKey  string
	userIdHeaderKey string
	ignoreChecker   Ignorer
}

func New(p Provider, o ...Option) *Manager {
	s := &Manager{
		provider:        p,
		appIdHeaderKey:  "X-App-Id",
		userIdHeaderKey: "X-User-Id",
	}
	s.With(o...)
	return s
}

func (m *Manager) Provider() Provider {
	return m.provider
}

func (m *Manager) AppIdHeaderKey() string {
	return m.appIdHeaderKey
}

func (m *Manager) UserIdHeaderKey() string {
	return m.userIdHeaderKey
}

func (m *Manager) Ignored(method, pattern string) bool {
	if m.ignoreChecker != nil {
		return m.ignoreChecker(method, pattern)
	}
	return false
}
