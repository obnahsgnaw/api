package api

import (
	"github.com/gin-gonic/gin"
	"github.com/obnahsgnaw/api/internal/middleware/authmid"
	"github.com/obnahsgnaw/api/internal/middleware/commonmid"
	"github.com/obnahsgnaw/api/internal/middleware/permmid"
	"github.com/obnahsgnaw/api/pkg/apierr"
	"github.com/obnahsgnaw/api/pkg/errobj"
	"github.com/obnahsgnaw/api/service"
	"github.com/obnahsgnaw/api/service/apidoc"
	"github.com/obnahsgnaw/api/service/authedapp"
	"github.com/obnahsgnaw/api/service/autheduser"
	"github.com/obnahsgnaw/api/service/crypt"
	"github.com/obnahsgnaw/api/service/perm"
	"github.com/obnahsgnaw/api/service/sign"
	"github.com/obnahsgnaw/rpc"
)

type Option func(s *Server)

func RegEnable() Option {
	return func(s *Server) {
		s.regEnable = true
	}
}
func AppMiddleware(m *authedapp.Manager) Option {
	return func(s *Server) {
		s.AddMiddleware("app", func() gin.HandlerFunc {
			return authmid.NewAppMid(m, func(msg string) {
				s.logger.Debug(msg)
			}, s.ErrorHandler())
		}, false)
	}
}
func CryptMiddleware(m *crypt.Manager) Option {
	return func(s *Server) {
		s.AddMiddleware("crypt", func() gin.HandlerFunc {
			return authmid.NewCryptMid(m, func(msg string) {
				s.logger.Debug(msg)
			}, s.ErrorHandler())
		}, false)
	}
}
func AuthMiddleware(m *autheduser.Manager) Option {
	return func(s *Server) {
		s.AddMiddleware("auth", func() gin.HandlerFunc {
			return authmid.NewAuthMid(m, func(msg string) {
				s.logger.Debug(msg)
			}, s.ErrorHandler())
		}, false)
	}
}
func SignMiddleware(m *sign.Manager) Option {
	return func(s *Server) {
		s.AddMiddleware("sign", func() gin.HandlerFunc {
			return authmid.NewSignMid(m, func(msg string) {
				s.logger.Debug(msg)
			}, s.ErrorHandler())
		}, false)
	}
}
func PermMiddleware(m *perm.Manager) Option {
	return func(s *Server) {
		s.AddMuxMiddleware("perm", func() service.MuxRouteHandleFunc {
			return permmid.NewMuxPermissionMid(m, func(msg string) {
				s.logger.Debug(msg)
			}, s.ErrorHandler())
		}, false)
	}
}
func Gateway(keyGen func() (string, error)) Option {
	return func(s *Server) {
		s.gatewayKeyGen = keyGen
	}
}
func ErrObjProvider(p errobj.Provider) Option {
	return func(s *Server) {
		s.errObjProvider = p
	}
}
func DocServer(config *apidoc.Config) Option {
	return func(s *Server) {
		s.addDoc(config)
	}
}
func RpcServer() Option {
	return func(s *Server) {
		s.rpcServer = rpc.New(s.app, s.httpEngine.Http().Listener(), s.id, s.name, s.endType, rpc.NewPServer(s.id, s.serverType), rpc.RegEnable())
		s.initRpcError()
	}
}
func RpcIns(ins *rpc.Server) Option {
	return func(s *Server) {
		s.rpcServer = ins
		s.rpcServer.AddRegInfo(s.id, s.name, rpc.NewPServer(s.id, s.serverType))
		s.initRpcError()
	}
}
func ErrCodePrefix(p int) Option {
	return func(s *Server) {
		s.errFactory = apierr.New(p)
	}
}
func WithoutRoutePrefix() Option {
	return func(s *Server) {
		s.withoutRoutePrefix = true
	}
}
func CommonMiddleware(name string, handler func(c *gin.Context, rqId, rqType string, debugger func(string)) error) Option {
	return func(s *Server) {
		s.AddMiddleware(name, func() gin.HandlerFunc {
			return commonmid.NewCommonMid(handler, func(msg string) { s.logger.Debug(msg) }, s.ErrorHandler())
		}, false)
	}
}
