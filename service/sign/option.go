package sign

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
func UserSignHeaderKey(key string) Option {
	return func(s *Manager) {
		s.signHeaderKey = key
	}
}
