package authedapp

type Option func(s *Manager)

func (m *Manager) With(o ...Option) {
	for _, oo := range o {
		if oo != nil {
			oo(m)
		}
	}
}

func AppIdHeaderKey(key string) Option {
	return func(s *Manager) {
		s.appIdHeaderKey = key
	}
}

func IgnoreChecker(i Ignorer, a App) Option {
	return func(s *Manager) {
		if i != nil && a != nil {
			s.ignoreChecker = i
			s.ignoreApp = a
		}
	}
}
