package api

import (
	"github.com/obnahsgnaw/api/internal/middleware/authmid"
	"github.com/obnahsgnaw/api/internal/middleware/permmid"
	"github.com/obnahsgnaw/api/service/authedapp"
	"github.com/obnahsgnaw/api/service/autheduser"
	"github.com/obnahsgnaw/api/service/cors"
	"github.com/obnahsgnaw/api/service/crypt"
	"github.com/obnahsgnaw/api/service/perm"
	"github.com/obnahsgnaw/api/service/sign"
	"io"
)

type Option func(s *Server)

func RegEnable() Option {
	return func(s *Server) {
		s.regEnable = true
	}
}
func AccessWriter(w io.Writer) Option {
	return func(s *Server) {
		s.accessWriter = w
	}
}
func ErrorWriter(w io.Writer) Option {
	return func(s *Server) {
		s.errWriter = w
	}
}
func TrustedProxies(proxies []string) Option {
	return func(s *Server) {
		s.trustedProxies = proxies
	}
}
func Cors(c *cors.Config) Option {
	return func(s *Server) {
		s.cors = c
	}
}
func AppMiddleware(m *authedapp.Manager) Option {
	return func(s *Server) {
		s.debug("withed app middleware")
		s.AddMiddleware(authmid.NewAppMid(m))
	}
}
func CryptMiddleware(m *crypt.Manager) Option {
	return func(s *Server) {
		s.debug("withed crypt middleware")
		s.AddMiddleware(authmid.NewCryptMid(m))
	}
}
func AuthMiddleware(m *autheduser.Manager) Option {
	return func(s *Server) {
		s.debug("withed auth middleware")
		s.AddMuxMiddleware(authmid.NewMuxAuthBeforeMid(m))
		s.AddMiddleware(authmid.NewAuthAfterMid(m))
	}
}
func SignMiddleware(m *sign.Manager) Option {
	return func(s *Server) {
		s.debug("withed signature middleware")
		s.AddMiddleware(authmid.NewSignMid(m))
	}
}
func PermMiddleware(m *perm.Manager) Option {
	return func(s *Server) {
		s.debug("withed permission middleware")
		s.AddMuxMiddleware(permmid.NewMuxPermissionMid(m))
	}
}
func Gateway(keyGen func() (string, error)) Option {
	return func(s *Server) {
		s.debug("withed gateway")
		s.gatewayKeyGen = keyGen
	}
}
