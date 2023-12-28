package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/obnahsgnaw/api/engine"
	"github.com/obnahsgnaw/api/internal/errhandler"
	"github.com/obnahsgnaw/api/internal/server"
	"github.com/obnahsgnaw/api/pkg/apierr"
	"github.com/obnahsgnaw/api/pkg/errobj"
	"github.com/obnahsgnaw/api/service"
	"github.com/obnahsgnaw/api/service/apidoc"
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
	"net/http"
	"path/filepath"
	"strings"
)

// Server API server
type Server struct {
	id             string // 模块
	name           string
	app            *application.Application
	endType        endtype.EndType
	serverType     servertype.ServerType
	engine         *engine.MuxHttp
	errFactory     *apierr.Factory
	rps            *rpc.Server
	logger         *zap.Logger
	logCnf         *logger.Config
	services       []ServiceProvider
	regInfo        *regCenter.RegInfo
	docRegInfos    []*regCenter.RegInfo
	mdProvider     *service.MethodMdProvider
	errObjProvider errobj.Provider
	engineIgRun    bool
	engineIgInit   bool
	regEnable      bool
	rpsIgRun       bool
	errs           []error
	pathPrefix     string
	middlewares    map[string]func() gin.HandlerFunc
	muxMiddleware  map[string]func() service.MuxRouteHandleFunc
	extRoutes      []func() service.RouteProvider
	gatewayKeyGen  func() (string, error)
	gatewayKey     string
}

// ServiceProvider api service provider
type ServiceProvider func(ctx context.Context, mux *runtime.ServeMux) (name string, err error)

func New(app *application.Application, id, name string, et endtype.EndType, e *engine.MuxHttp, pathPrefix string, errCodePrefix int, options ...Option) *Server {
	var err error
	s := &Server{
		id:         id,
		name:       name,
		app:        app,
		serverType: servertype.Api,
		endType:    et,
		engine:     e,
		errFactory: apierr.New(errCodePrefix),
		pathPrefix: strings.TrimPrefix(pathPrefix, "/"),
		errObjProvider: func(param errobj.Param) interface{} {
			return param
		},
		mdProvider:    service.NewMdProvider(),
		middlewares:   make(map[string]func() gin.HandlerFunc),
		muxMiddleware: make(map[string]func() service.MuxRouteHandleFunc),
	}
	if s.id == "" || s.name == "" {
		s.addErr(s.apiServerError(s.msg("id name invalid"), nil))
	}
	if s.pathPrefix == "" {
		s.addErr(s.apiServerError(s.msg("path prefix empty"), nil))
	}
	s.logCnf = logCnf(app, et, id)
	s.logger, err = logger.New(utils.ToStr(s.serverType.String(), ":", s.endType.String(), "-", s.id), s.logCnf, s.app.Debugger().Debug())
	s.addErr(err)
	s.initRegInfo()
	s.With(options...)
	return s
}

func logCnf(app *application.Application, et endtype.EndType, id string) *logger.Config {
	cnf := logger.CopyCnfWithLevel(app.LogConfig())
	if cnf != nil {
		cnf.AddSubDir(filepath.Join(et.String(), utils.ToStr(servertype.Api.String(), "-", id)))
		cnf.SetFilename(utils.ToStr(servertype.Api.String(), "-", id))
	}
	return cnf
}

func (s *Server) With(options ...Option) {
	for _, o := range options {
		if o != nil {
			o(s)
		}
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
	return s.serverType
}

// EndType return the api server end type
func (s *Server) EndType() endtype.EndType {
	return s.endType
}

// Host return the api service host
func (s *Server) Host() url.Host {
	return s.engine.Http().Host()
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

// LogConfig return
func (s *Server) LogConfig() *logger.Config {
	return s.logCnf
}

func (s *Server) App() *application.Application {
	return s.app
}

func (s *Server) Engine() *engine.MuxHttp {
	return s.engine
}

func (s *Server) Rpc() *rpc.Server {
	return s.rps
}

// RegisterApiService register a api service
func (s *Server) RegisterApiService(provider ServiceProvider) {
	s.services = append(s.services, provider)
}

// RegisterRpcService register a rcp service
func (s *Server) RegisterRpcService(provider rpc.ServiceInfo) {
	if s.rps != nil {
		s.rps.RegisterService(provider)
	}
}

func (s *Server) AddMiddleware(name string, mid func() gin.HandlerFunc, force bool) {
	if _, ok := s.middlewares[name]; ok && force || !ok {
		if s.logger != nil {
			s.logger.Debug(name + "middleware enabled")
		}
		s.middlewares[name] = mid
	}
}

func (s *Server) AddMuxMiddleware(name string, mid func() service.MuxRouteHandleFunc, force bool) {
	if _, ok := s.muxMiddleware[name]; ok && force || !ok {
		if s.logger != nil {
			s.logger.Debug(name + "middleware enabled")
		}
		s.muxMiddleware[name] = mid
	}
}

func (s *Server) AddRoute(route func() service.RouteProvider) {
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
	var err error
	if len(s.errs) != 0 {
		failedCb(s.errs[0])
		return
	}
	s.logger.Info("init start...")
	if s.rps != nil && !s.rpsIgRun {
		s.app.AddServer(s.rps)
		s.logger.Debug("api rpc service enabled")
	}
	s.initEngine()
	for _, sp := range s.services {
		if n, err1 := sp(s.app.Context(), s.engine.Mux()); err1 != nil {
			failedCb(s.apiServerError(s.msg("service[", n, "] register failed"), err1))
			return
		} else {
			s.logger.Debug("service[" + n + "] registered")
		}
	}
	s.logger.Info("services initialized")
	if s.app.Register() != nil {
		if err = s.register(true); err != nil {
			failedCb(s.apiServerError(s.msg("register failed"), err))
			return
		}
		s.logger.Debug("server registered")
		if err = s.regGateway(); err != nil {
			failedCb(err)
			return
		}
		s.logger.Debug("gateway registered")
	}
	s.logger.Info("register initialized")
	s.logger.Info("initialized")
	if !s.engineIgRun {
		go func() {
			if err = s.engine.Http().RunAndServ(); err != nil {
				failedCb(s.apiServerError(s.msg("engine run failed, err="+err.Error()), nil))
			}
		}()
		s.logger.Info(utils.ToStr("server[", s.Host().String(), "] listen and serving..."))
	}
}

func (s *Server) initEngine() {
	if !s.engineIgInit {
		var mid []gin.HandlerFunc
		for _, m := range s.middlewares {
			mid = append(mid, m())
		}
		var mmid []service.MuxRouteHandleFunc
		for _, m := range s.muxMiddleware {
			mmid = append(mmid, m())
		}
		var extRoutes []service.RouteProvider
		for _, m := range s.extRoutes {
			extRoutes = append(extRoutes, m())
		}
		server.InitRpcHttpProxyServer(s.engine.Http().Engine(), s.engine.Mux(), &server.MuxConfig{
			PathPrefix:     s.pathPrefix,
			MdProvider:     s.mdProvider,
			Middlewares:    mid,
			MuxMiddleware:  mmid,
			ExtRoutes:      extRoutes,
			ErrObjProvider: s.errObjProvider,
			Debugger:       s.app.Debugger(),
		})
		s.logger.Info("engine initialized(default)")
	} else {
		var extRoutes []service.RouteProvider
		for _, m := range s.extRoutes {
			extRoutes = append(extRoutes, m())
		}
		server.AddExtRoute(s.engine.Http().Engine(), extRoutes)
		s.logger.Info("engine initialized(customer)")
	}
}

func (s *Server) initRegInfo() {
	s.regInfo = &regCenter.RegInfo{
		AppId:   s.app.ID(),
		RegType: regtype.Http,
		ServerInfo: regCenter.ServerInfo{
			Id:      s.id,
			Name:    s.name,
			EndType: s.endType.String(),
			Type:    s.serverType.String(),
		},
		Host:      s.engine.Http().Host().String(),
		Val:       s.engine.Http().Host().String(),
		Ttl:       s.app.RegTtl(),
		KeyPreGen: regCenter.DefaultRegKeyPrefixGenerator(),
	}
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
	if s.logger != nil {
		_ = s.logger.Sync()
		s.logger.Info("released")
	}
}

func (s *Server) addDoc(config *apidoc.Config) {
	if config.EndType == "" {
		config.EndType = endtype.Backend
	}
	if config.Path == "" {
		config.Path = "/" + s.id + "-docs/" + config.EndType.String() + "-doc"
	} else {
		config.Path = "/" + s.id + "-docs/" + strings.TrimPrefix(config.Path, "/")
	}
	config.SetOrigin(url.Origin{
		Protocol: config.Protocol,
		Host:     s.engine.Http().Host(),
	})
	docRegInfo := &regCenter.RegInfo{
		AppId:   s.app.ID(),
		RegType: regtype.Doc,
		ServerInfo: regCenter.ServerInfo{
			Id:      s.id,
			Name:    s.name,
			Type:    s.serverType.String(),
			EndType: config.EndType.String(),
		},
		Host: s.engine.Http().Host().String(),
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
		docRegInfo.ServerInfo.EndType = s.endType.String()
	}
	s.docRegInfos = append(s.docRegInfos, docRegInfo)
	s.AddRoute(func() service.RouteProvider {
		docUrl := url.Origin{
			Protocol: "http",
			Host:     s.engine.Http().Host(),
		}.String() + config.Path
		if s.logger != nil {
			s.logger.Debug("api " + config.EndType.String() + "doc service enabled, url=" + docUrl)
		}
		return server.DocRoute(config.Path, config.Provider)
	})
}

func (s *Server) ErrorHandler() func(err error, marshaler runtime.Marshaler, w http.ResponseWriter) {
	return func(err error, marshaler runtime.Marshaler, w http.ResponseWriter) {
		errhandler.HandlerErr(err, marshaler, w, nil, s.errObjProvider, s.app.Debugger())
	}
}

func (s *Server) msg(msg ...string) string {
	return utils.ToStr("Api Server[", s.name, "]", utils.ToStr(msg...))
}

func (s *Server) register(reg bool) error {
	var cb = func(msg string) { s.logger.Debug(msg) }
	// reg http
	if s.regEnable {
		if reg {
			if err := s.app.DoRegister(s.regInfo, cb); err != nil {
				return err
			}
		} else {
			if err := s.app.DoUnregister(s.regInfo, cb); err != nil {
				return err
			}
		}
	}

	// reg doc
	if len(s.docRegInfos) > 0 {
		for _, docRegInfo := range s.docRegInfos {
			if reg {
				if err := s.app.DoRegister(docRegInfo, cb); err != nil {
					return err
				}
			} else {
				if err := s.app.DoUnregister(docRegInfo, cb); err != nil {
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
				Host:     s.engine.Http().Host(),
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
