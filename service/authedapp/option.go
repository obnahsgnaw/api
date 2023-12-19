package authedapp

type Option func(s *Manager)

func (m *Manager) With(o ...Option) {
	for _, oo := range o {
		if oo != nil {
			oo(m)
		}
	}
}

func OutsideValidate(cb func() bool) Option {
	return func(s *Manager) {
		s.outsideValidate = cb
	}
}

func AppIdHeaderKey(key string) Option {
	return func(s *Manager) {
		s.appIdBfHeaderKey = key
	}
}

func AuthedAppidHeaderKey(key string) Option {
	return func(s *Manager) {
		s.appIdAfHeaderKey = key
	}
}
