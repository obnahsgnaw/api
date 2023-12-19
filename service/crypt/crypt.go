package crypt

type Provider interface {
	Encrypt(appid, uid string, iv []byte, data []byte) ([]byte, error)
	Decrypt(appid, uid string, iv []byte, data []byte) ([]byte, error)
}

type Manager struct {
	provider        Provider
	appIdHeaderKey  string
	userIdHeaderKey string
	userIvHeaderKey string
}

func New(provider Provider, o ...Option) *Manager {
	s := &Manager{
		provider:        provider,
		appIdHeaderKey:  "X-App-Id",
		userIdHeaderKey: "X-User-Id",
		userIvHeaderKey: "X-User-Iv",
	}
	s.With(o...)
	return s
}

// Provider return crypt provider
func (m *Manager) Provider() Provider {
	return m.provider
}

func (m *Manager) AppIdHeaderKey() string {
	return m.appIdHeaderKey
}

func (m *Manager) UserIdHeaderKey() string {
	return m.userIdHeaderKey
}

func (m *Manager) UserIvHeaderKey() string {
	return m.userIvHeaderKey
}
