package sign

type Provider interface {
	Validate(appid, uid, method, uri, signature, timestamp, nonce string) error
	Generate(appid, uid, method, uri string) (signature, timestamp, nonce string, err error)
}

type Manager struct {
	provider          Provider
	appIdHeaderKey    string
	userIdHeaderKey   string
	userSignHeaderKey string
}

func New(provider Provider, o ...Option) *Manager {
	s := &Manager{
		provider:          provider,
		appIdHeaderKey:    "X-App-Id",
		userIdHeaderKey:   "X-User-Id",
		userSignHeaderKey: "X-User-Signature",
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

func (m *Manager) UserSignHeaderKey() string {
	return m.userSignHeaderKey
}
