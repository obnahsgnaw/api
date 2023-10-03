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
		s.AddMiddleware(authmid.NewAppMid(m, func(msg string) {
			s.debug(msg)
		}))
		s.debug("app middleware enabled")
	}
}
func CryptMiddleware(m *crypt.Manager) Option {
	return func(s *Server) {
		s.AddMiddleware(authmid.NewCryptMid(m, func(msg string) {
			s.debug(msg)
		}))
		s.debug("crypt middleware enabled")
	}
}
func AuthMiddleware(m *autheduser.Manager) Option {
	return func(s *Server) {
		s.AddMiddleware(authmid.NewAuthMid(m, func(msg string) {
			s.debug(msg)
		}))
		s.debug("auth middleware enabled")
	}
}
func SignMiddleware(m *sign.Manager) Option {
	return func(s *Server) {
		s.AddMiddleware(authmid.NewSignMid(m, func(msg string) {
			s.debug(msg)
		}))
		s.debug("signature middleware enabled")
	}
}
func PermMiddleware(m *perm.Manager) Option {
	return func(s *Server) {
		s.AddMuxMiddleware(permmid.NewMuxPermissionMid(m, func(msg string) {
			s.debug(msg)
		}))
		s.debug("permission middleware enabled")
	}
}
func Gateway(keyGen func() (string, error)) Option {
	return func(s *Server) {
		s.gatewayKeyGen = keyGen
		s.debug("gateway enabled")
	}
}
func RouteDebug(debug bool) Option {
	return func(s *Server) {
		s.debug("route debug enabled")
		s.routeDebug = debug
	}
}
