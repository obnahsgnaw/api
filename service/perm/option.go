package perm

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

func UserIdHeaderKey(key string) Option {
	return func(s *Manager) {
		s.userIdHeaderKey = key
	}
}

func IgnoreChecker(i Ignorer) Option {
	return func(s *Manager) {
		if i != nil {
			s.ignoreChecker = i
		}
	}
}
