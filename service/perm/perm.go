package perm

import "net/http"

type Provider interface {
	Can(appid, uid, method, pattern string) error
}

type Ignorer func(method, pattern string) bool
type Formatter func(r *http.Request, pattern string) string

type Manager struct {
	provider        Provider
	appIdHeaderKey  string
	userIdHeaderKey string
	ignoreChecker   Ignorer
	patternFormater Formatter
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

func (m *Manager) PatternFormat(r *http.Request, pattern string) string {
	if m.patternFormater != nil {
		return m.patternFormater(r, pattern)
	}
	return pattern
}
