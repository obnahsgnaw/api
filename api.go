package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/obnahsgnaw/api/internal/errhandler"
	"github.com/obnahsgnaw/api/internal/server"
	"github.com/obnahsgnaw/api/pkg/apierr"
	"github.com/obnahsgnaw/api/pkg/errobj"
	"github.com/obnahsgnaw/api/service"
	"github.com/obnahsgnaw/api/service/apidoc"
	"github.com/obnahsgnaw/api/service/cors"
	"github.com/obnahsgnaw/application"
	"github.com/obnahsgnaw/application/endtype"
	"github.com/obnahsgnaw/application/pkg/logging/logger"
	"github.com/obnahsgnaw/application/pkg/url"
	"github.com/obnahsgnaw/application/pkg/utils"
	"github.com/obnahsgnaw/application/regtype"
	"github.com/obnahsgnaw/application/servertype"
	"github.com/obnahsgnaw/application/service/regCenter"
	"github.com/obnahsgnaw/rpc"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
)

// Server API server
type Server struct {
	id          string // 模块
	name        string
	st          servertype.ServerType
	et          endtype.EndType
	app         *application.Application
	engine      *gin.Engine
	mux         *runtime.ServeMux
	services    []ServiceProvider
	regEnable   bool
	regInfo     *regCenter.RegInfo
	docRegInfos []*regCenter.RegInfo
	errFactory  *apierr.Factory
	rps         *rpc.Server
	logger      *zap.Logger
	errs        []error

	host           url.Host
	pathPrefix     string
	routeDebug     bool
	accessWriter   io.Writer
	errWriter      io.Writer
	trustedProxies []string
	cors           *cors.Config
	mdProvider     *service.MethodMdProvider
	middlewares    []gin.HandlerFunc
	muxMiddleware  []service.MuxRouteHandleFunc
	extRoutes      []service.RouteProvider
	errObjProvider errobj.Provider
	gatewayKeyGen  func() (string, error)
	gatewayKey     string
}

// ServiceProvider api service provider
type ServiceProvider func(ctx context.Context, mux *runtime.ServeMux) (name string, err error)

func New(app *application.Application, id, name string, et endtype.EndType, pathPrefix string, host url.Host, errCodePrefix int, options ...Option) *Server {
	var err error
	s := &Server{
		id:         id,
		name:       name,
		st:         servertype.Api,
		et:         et,
		app:        app,
		errFactory: apierr.New(errCodePrefix),
		pathPrefix: strings.TrimPrefix(pathPrefix, "/"),
		host:       host,
		errObjProvider: func(param errobj.Param) interface{} {
			return param
		},
		mdProvider: service.NewMdProvider(),
	}
	if s.id == "" || s.name == "" {
		s.addErr(s.apiServerError(s.msg("id name invalid"), nil))
	}
	if s.host.Port <= 0 || s.host.Ip == "" {
		s.addErr(s.apiServerError(s.msg("host invalid"), nil))
	}
	if s.pathPrefix == "" {
		s.addErr(s.apiServerError(s.msg("path prefix empty"), nil))
	}
	s.logger, err = logger.New(utils.ToStr("Api[", s.et.String(), "][", id, "]"), s.app.LogConfig(), s.app.Debugger().Debug())
	s.addErr(err)
	s.regInfo = &regCenter.RegInfo{
		AppId:   app.ID(),
		RegType: regtype.Http,
		ServerInfo: regCenter.ServerInfo{
			Id:      s.id,
			Name:    s.name,
			EndType: s.et.String(),
			Type:    s.st.String(),
		},
		Host:      s.host.String(),
		Val:       s.host.String(),
		Ttl:       s.app.RegTtl(),
		KeyPreGen: regCenter.DefaultRegKeyPrefixGenerator(),
	}
	s.With(options...)
	return s
}

func (s *Server) With(options ...Option) {
	for _, o := range options {
		o(s)
	}
}

// ID return the api service id
func (s *Server) ID() string {
	return s.id
}

// Name return the api service name
func (s *Server) Name() string {
	return s.name
}

// Type return the api server end type
func (s *Server) Type() servertype.ServerType {
	return s.st
}

// EndType return the api server end type
func (s *Server) EndType() endtype.EndType {
	return s.et
}

// Host return the api service host
func (s *Server) Host() url.Host {
	return s.host
}

// RegEnabled reg http
func (s *Server) RegEnabled() bool {
	return s.regEnable
}

// RegInfo return the server register info
func (s *Server) RegInfo() *regCenter.RegInfo {
	return s.regInfo
}

// Logger return the logger
func (s *Server) Logger() *zap.Logger {
	return s.logger
}

func (s *Server) App() *application.Application {
	return s.app
}

// RegisterService register a api service
func (s *Server) RegisterService(provider ServiceProvider) {
	s.services = append(s.services, provider)
}

func (s *Server) AddMiddleware(mid gin.HandlerFunc) {
	s.middlewares = append(s.middlewares, mid)
}

func (s *Server) AddMuxMiddleware(mid service.MuxRouteHandleFunc) {
	s.muxMiddleware = append(s.muxMiddleware, mid)
}

func (s *Server) AddRoute(route service.RouteProvider) {
	s.extRoutes = append(s.extRoutes, route)
}

func (s *Server) AddDefIncomeMd(key string, valProvider service.MdValParser) {
	s.mdProvider.AddDefault(key, valProvider)
}

func (s *Server) AddIncomeMd(method, key string, valProvider service.MdValParser) {
	s.mdProvider.Add(method, key, valProvider)
}

// ErrCode return err code factory
func (s *Server) ErrCode() *apierr.Factory {
	return s.errFactory
}

func (s *Server) Run(failedCb func(error)) {
	if len(s.errs) != 0 {
		failedCb(s.errs[0])
		return
	}
	s.debug("start running...")
	engine, mux, err := server.NewRpcHttpProxyServer(&server.HttpConfig{
		PathPrefix:     s.pathPrefix,
		AccessWriter:   s.accessWriter,
		ErrWriter:      s.errWriter,
		TrustedProxies: s.trustedProxies,
		Cors:           s.cors,
		MdProvider:     s.mdProvider,
		Middlewares:    s.middlewares,
		MuxMiddleware:  s.muxMiddleware,
		ExtRoutes:      s.extRoutes,
		ErrObjProvider: s.errObjProvider,
		Debugger:       s.app.Debugger(),
		RouteDebug:     s.routeDebug,
	})
	if err != nil {
		failedCb(s.apiServerError(s.msg("new engine failed"), err))
		return
	}
	s.engine = engine
	s.mux = mux
	for _, sp := range s.services {
		if n, err := sp(s.app.Context(), s.mux); err != nil {
			failedCb(s.apiServerError(s.msg("service[", n, "] register failed"), err))
			return
		} else {
			s.debug("service[" + n + "] registered")
		}
	}
	if s.app.Register() != nil {
		if err = s.register(true); err != nil {
			failedCb(s.apiServerError(s.msg("register failed"), err))
			return
		}
		if err = s.regGateway(); err != nil {
			failedCb(err)
			return
		}
	}
	go func(host string, engine *gin.Engine) {
		if err = engine.Run(host); err != nil {
			failedCb(s.apiServerError(s.msg("engine run failed, err="+err.Error()), nil))
		}
	}(s.host.String(), s.engine)
	s.logger.Info(utils.ToStr("api server[", s.Host().String(), "] listen and serving..."))
}

func (s *Server) RefreshGateway() error {
	return s.regGateway()
}

func (s *Server) Release() {
	if s.app.Register() != nil {
		_ = s.register(false)
		if s.gatewayKey != "" {
			_ = s.app.Register().Unregister(s.app.Context(), s.gatewayKey)
			s.gatewayKey = ""
		}
	}
	_ = s.logger.Sync()
	s.debug("released")
}

func (s *Server) WithDocService(config *apidoc.Config) {
	s.debug("doc server enabled")
	if config.EndType == "" {
		config.EndType = endtype.Backend
	}
	config.SetOrigin(url.Origin{
		Protocol: config.Protocol,
		Host: url.Host{
			Ip:   s.host.Ip,
			Port: s.host.Port,
		},
	})
	docRegInfo := &regCenter.RegInfo{
		AppId:   s.app.ID(),
		RegType: regtype.Doc,
		ServerInfo: regCenter.ServerInfo{
			Id:      s.id,
			Name:    s.name,
			Type:    s.st.String(),
			EndType: config.EndType.String(),
		},
		Host: s.host.String(),
		Val:  "",
		Values: map[string]string{
			"url":         config.Url(),
			"debugOrigin": config.DebugOrigin.String(),
			"title":       config.Title,
		},
		Ttl:       config.RegTtl,
		KeyPreGen: config.RegKeyPreGen,
	}
	if docRegInfo.ServerInfo.EndType == "" {
		docRegInfo.ServerInfo.EndType = s.et.String()
	}
	s.docRegInfos = append(s.docRegInfos, docRegInfo)
	s.AddRoute(server.DocRoute(config.Path, config.Provider))
}

func (s *Server) WithRpcServer(port int, autoAdd bool) *rpc.Server {
	s.rps = rpc.New(s.app, s.id, s.name, s.et, url.Host{Ip: s.host.Ip, Port: port}, rpc.Parent(s), rpc.RegEnable())
	if autoAdd {
		s.debug("rpc server enabled")
		s.app.AddServer(s.rps)
	}

	return s.rps
}

func (s *Server) ErrorHandler() func(err error, marshaler runtime.Marshaler, w http.ResponseWriter) {
	return func(err error, marshaler runtime.Marshaler, w http.ResponseWriter) {
		errhandler.HandlerErr(err, marshaler, w, nil, s.errObjProvider, s.app.Debugger())
	}
}

func (s *Server) debug(msg string) {
	if s.app.Debugger().Debug() {
		s.logger.Debug(msg)
	}
}

func (s *Server) msg(msg ...string) string {
	return utils.ToStr("Api Server[", s.name, "]", utils.ToStr(msg...))
}

func (s *Server) register(reg bool) error {
	// reg http
	if s.regEnable {
		if reg {
			if err := s.app.DoRegister(s.regInfo); err != nil {
				return err
			}
		} else {
			if err := s.app.DoUnregister(s.regInfo); err != nil {
				return err
			}
		}
	}

	// reg doc
	if len(s.docRegInfos) > 0 {
		for _, docRegInfo := range s.docRegInfos {
			if reg {
				if err := s.app.DoRegister(docRegInfo); err != nil {
					return err
				}
			} else {
				if err := s.app.DoUnregister(docRegInfo); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *Server) watch() {
	// watch http
}

func (s *Server) regGateway() error {
	if s.gatewayKeyGen != nil {
		key, err := s.gatewayKeyGen()
		if err != nil {
			return s.apiServerError(s.msg("fetch gateway failed"), err)
		}

		if key != "" && key != s.gatewayKey {
			s.gatewayKey = key
			_ = s.app.Register().Unregister(s.app.Context(), s.gatewayKey)
			if err = s.app.Register().Register(s.app.Context(), s.gatewayKey, url.Origin{
				Protocol: url.HTTP,
				Host:     s.host,
			}.String(), s.app.RegTtl()); err != nil {
				return s.apiServerError(s.msg("register gateway failed"), err)
			}
		}
	}
	return nil
}

func (s *Server) apiServerError(msg string, err error) error {
	return utils.TitledError(utils.ToStr("api server[", s.name, "] error"), msg, err)
}

func (s *Server) addErr(err error) {
	if err != nil {
		s.errs = append(s.errs, err)
	}
}
