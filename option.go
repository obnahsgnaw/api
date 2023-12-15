package api

import (
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/obnahsgnaw/api/internal/middleware/authmid"
	"github.com/obnahsgnaw/api/internal/middleware/permmid"
	"github.com/obnahsgnaw/api/pkg/errobj"
	"github.com/obnahsgnaw/api/service"
	"github.com/obnahsgnaw/api/service/apidoc"
	"github.com/obnahsgnaw/api/service/authedapp"
	"github.com/obnahsgnaw/api/service/autheduser"
	"github.com/obnahsgnaw/api/service/cors"
	"github.com/obnahsgnaw/api/service/crypt"
	"github.com/obnahsgnaw/api/service/perm"
	"github.com/obnahsgnaw/api/service/sign"
	"github.com/obnahsgnaw/application/pkg/url"
	"github.com/obnahsgnaw/rpc"
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
		s.AddMiddleware(func() gin.HandlerFunc {
			s.logger.Debug("app middleware enabled")
			return authmid.NewAppMid(m, func(msg string) {
				s.logger.Debug(msg)
			}, s.ErrorHandler())
		})
	}
}
func CryptMiddleware(m *crypt.Manager) Option {
	return func(s *Server) {
		s.AddMiddleware(func() gin.HandlerFunc {
			s.logger.Debug("crypt middleware enabled")
			return authmid.NewCryptMid(m, func(msg string) {
				s.logger.Debug(msg)
			}, s.ErrorHandler())
		})
	}
}
func AuthMiddleware(m *autheduser.Manager) Option {
	return func(s *Server) {
		s.AddMiddleware(func() gin.HandlerFunc {
			s.logger.Debug("auth middleware enabled")
			return authmid.NewAuthMid(m, func(msg string) {
				s.logger.Debug(msg)
			}, s.ErrorHandler())
		})
	}
}
func SignMiddleware(m *sign.Manager) Option {
	return func(s *Server) {
		s.AddMiddleware(func() gin.HandlerFunc {
			s.logger.Debug("signature middleware enabled")
			return authmid.NewSignMid(m, func(msg string) {
				s.logger.Debug(msg)
			}, s.ErrorHandler())
		})
	}
}
func PermMiddleware(m *perm.Manager) Option {
	return func(s *Server) {
		s.AddMuxMiddleware(func() service.MuxRouteHandleFunc {
			s.logger.Debug("permission middleware enabled")
			return permmid.NewMuxPermissionMid(m, func(msg string) {
				s.logger.Debug(msg)
			}, s.ErrorHandler())
		})
	}
}
func Gateway(keyGen func() (string, error)) Option {
	return func(s *Server) {
		s.gatewayKeyGen = keyGen
	}
}
func RouteDebug(debug bool) Option {
	return func(s *Server) {
		s.routeDebug = debug
	}
}
func ErrObjProvider(p errobj.Provider) Option {
	return func(s *Server) {
		s.errObjProvider = p
	}
}
func Engine(e *gin.Engine, mux *runtime.ServeMux, ehost url.Host) Option {
	return func(s *Server) {
		s.engine = e
		s.mux = mux
		s.engineCus = true
		s.host = ehost
		s.initRegInfo()
	}
}
func DefaultEngine(host url.Host) Option {
	return func(s *Server) {
		s.host = host
		s.initRegInfo()
	}
}
func DocService(config *apidoc.Config) Option {
	return func(s *Server) {
		s.withDocService(config)
	}
}
func RpcServer(port int, autoAdd bool) Option {
	return func(s *Server) {
		s.withRpcServer(port, autoAdd)
	}
}
func RpcServerIns(ins *rpc.Server, autoAdd bool) Option {
	return func(s *Server) {
		s.withRpcServerIns(ins, autoAdd)
	}
}
