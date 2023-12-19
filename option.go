package api

import (
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/obnahsgnaw/api/internal/middleware/authmid"
	"github.com/obnahsgnaw/api/internal/middleware/permmid"
	"github.com/obnahsgnaw/api/internal/server"
	"github.com/obnahsgnaw/api/pkg/errobj"
	"github.com/obnahsgnaw/api/service"
	"github.com/obnahsgnaw/api/service/apidoc"
	"github.com/obnahsgnaw/api/service/authedapp"
	"github.com/obnahsgnaw/api/service/autheduser"
	"github.com/obnahsgnaw/api/service/crypt"
	"github.com/obnahsgnaw/api/service/perm"
	"github.com/obnahsgnaw/api/service/sign"
	"github.com/obnahsgnaw/application/pkg/url"
	"github.com/obnahsgnaw/application/pkg/utils"
	"github.com/obnahsgnaw/http"
	"github.com/obnahsgnaw/http/cors"
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
func TrustedProxies(proxies []string) Option { // 需要在引擎之前
	return func(s *Server) {
		s.trustedProxies = proxies
	}
}
func Cors(c *cors.Config) Option { // 需要在引擎之前
	return func(s *Server) {
		s.cors = c
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
func RouteDebug(debug bool) Option { // 需要在引擎之前
	return func(s *Server) {
		s.routeDebug = debug
	}
}
func ErrObjProvider(p errobj.Provider) Option {
	return func(s *Server) {
		s.errObjProvider = p
	}
}
func Engine(e *http.PortedEngine, mux *runtime.ServeMux) Option {
	return func(s *Server) {
		s.engine = e.Engine()
		s.mux = mux
		s.engineCus = true
		s.host = e.Host()
		s.initRegInfo()
	}
}
func NewEngine(host url.Host) Option {
	return func(s *Server) {
		var err error
		s.host = host
		s.engine, err = server.NewEngine(&http.Config{
			Name:           utils.ToStr(s.et.String(), "-", s.id),
			DebugMode:      s.routeDebug,
			LogDebug:       s.app.Debugger().Debug(),
			TrustedProxies: s.trustedProxies,
			Cors:           s.cors,
			LogCnf:         s.logCnf,
		})
		s.mux = server.NewMux()
		s.addErr(err)
		s.initRegInfo()
	}
}
func EngineInsOrNew(e *gin.Engine, mux *runtime.ServeMux, host url.Host) Option {
	if e == nil {
		return NewEngine(host)
	} else {
		return Engine(http.NewPortedEngine(e, host), mux)
	}
}
func Doc(config *apidoc.Config) Option {
	return func(s *Server) {
		s.addDoc(config)
	}
}
func RpcInsOrNew(ins *rpc.Server, host url.Host, autoAdd bool) Option {
	if ins != nil {
		return RpcIns(ins, autoAdd)
	}
	return RpcServer(host, autoAdd)
}
func RpcServer(host url.Host, autoAdd bool) Option {
	return func(s *Server) {
		s.withRpcServer(host, autoAdd)
	}
}
func RpcIns(ins *rpc.Server, autoAdd bool) Option {
	return func(s *Server) {
		s.withRpcServerIns(ins, autoAdd)
	}
}
