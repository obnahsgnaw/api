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
	"github.com/obnahsgnaw/application/pkg/logging/writer"
	"github.com/obnahsgnaw/application/pkg/url"
	"github.com/obnahsgnaw/application/pkg/utils"
	"github.com/obnahsgnaw/application/regtype"
	"github.com/obnahsgnaw/application/servertype"
	"github.com/obnahsgnaw/application/service/regCenter"
	"github.com/obnahsgnaw/rpc"
	"go.uber.org/zap"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

// Server API server
type Server struct {
	id             string // 模块
	name           string
	st             servertype.ServerType
	et             endtype.EndType
	app            *application.Application
	engine         *gin.Engine
	engineCus      bool
	mux            *runtime.ServeMux
	services       []ServiceProvider
	regEnable      bool
	regInfo        *regCenter.RegInfo
	docRegInfos    []*regCenter.RegInfo
	errFactory     *apierr.Factory
	rps            *rpc.Server
	rpsAdd         bool
	logger         *zap.Logger
	logCnf         *logger.Config
	errs           []error
	host           url.Host
	pathPrefix     string
	routeDebug     bool
	accessWriter   io.Writer
	errWriter      io.Writer
	trustedProxies []string
	cors           *cors.Config
	mdProvider     *service.MethodMdProvider
	middlewares    []func() gin.HandlerFunc
	muxMiddleware  []func() service.MuxRouteHandleFunc
	extRoutes      []func() service.RouteProvider
	errObjProvider errobj.Provider
	gatewayKeyGen  func() (string, error)
	gatewayKey     string
}

// ServiceProvider api service provider
type ServiceProvider func(ctx context.Context, mux *runtime.ServeMux) (name string, err error)

func New(app *application.Application, id, name string, et endtype.EndType, pathPrefix string, errCodePrefix int, options ...Option) *Server {
	var err error
	s := &Server{
		id:         id,
		name:       name,
		st:         servertype.Api,
		et:         et,
		app:        app,
		errFactory: apierr.New(errCodePrefix),
		pathPrefix: strings.TrimPrefix(pathPrefix, "/"),
		errObjProvider: func(param errobj.Param) interface{} {
			return param
		},
		mdProvider: service.NewMdProvider(),
	}
	if s.id == "" || s.name == "" {
		s.addErr(s.apiServerError(s.msg("id name invalid"), nil))
	}
	if s.pathPrefix == "" {
		s.addErr(s.apiServerError(s.msg("path prefix empty"), nil))
	}
	s.logCnf = logger.CopyCnfWithLevel(s.app.LogConfig())
	if s.logCnf != nil {
		s.logCnf.AddSubDir(filepath.Join(s.et.String(), "api-"+s.id))
		s.logCnf.SetFilename(s.id)
	}
	s.logger, err = logger.New(utils.ToStr("api:", s.et.String(), "-", s.id), s.logCnf, s.app.Debugger().Debug())
	s.addErr(err)
	s.With(options...)
	if s.accessWriter == nil && s.logCnf != nil {
		var wts []io.Writer
		if w, err := logger.NewAccessWriter(s.logCnf, s.app.Debugger().Debug()); err == nil {
			wts = append(wts, w)
		} else {
			s.addErr(err)
		}
		if s.app.Debugger().Debug() {
			wts = append(wts, writer.NewStdWriter())
		}
		s.accessWriter = newMultiWriter(wts...)
	}
	if s.errWriter == nil && s.logCnf != nil {
		var wts []io.Writer
		if w, err := logger.NewErrorWriter(s.logCnf, s.app.Debugger().Debug()); err == nil {
			wts = append(wts, w)
		} else {
			s.addErr(err)
		}
		if s.app.Debugger().Debug() {
			wts = append(wts, writer.NewStdWriter())
		}
		s.accessWriter = newMultiWriter(wts...)
	}
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

// LogConfig return
func (s *Server) LogConfig() *logger.Config {
	return s.logCnf
}

func (s *Server) App() *application.Application {
	return s.app
}

// RegisterService register a api service
func (s *Server) RegisterService(provider ServiceProvider) {
	s.services = append(s.services, provider)
}

// RegisterRpcService register a rcp service
func (s *Server) RegisterRpcService(provider rpc.ServiceInfo) {
	if s.rps != nil {
		s.rps.RegisterService(provider)
	}
}

func (s *Server) Rpc() *rpc.Server {
	return s.rps
}

func (s *Server) AddMiddleware(mid func() gin.HandlerFunc) {
	s.middlewares = append(s.middlewares, mid)
}

func (s *Server) AddMuxMiddleware(mid func() service.MuxRouteHandleFunc) {
	s.muxMiddleware = append(s.muxMiddleware, mid)
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
	if s.rps != nil && s.rpsAdd {
		s.app.AddServer(s.rps)
		s.logger.Debug("api doc service enabled")
	}
	if s.engine == nil {
		if s.host.Port <= 0 || s.host.Ip == "" {
			s.addErr(s.apiServerError(s.msg("host invalid"), nil))
			return
		}
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
		s.engine, s.mux, err = server.NewRpcHttpProxyServer(&server.HttpConfig{
			Name:           utils.ToStr(s.et.String(), "-", s.id),
			PathPrefix:     s.pathPrefix,
			AccessWriter:   s.accessWriter,
			ErrWriter:      s.errWriter,
			TrustedProxies: s.trustedProxies,
			Cors:           s.cors,
			MdProvider:     s.mdProvider,
			Middlewares:    mid,
			MuxMiddleware:  mmid,
			ExtRoutes:      extRoutes,
			ErrObjProvider: s.errObjProvider,
			Debugger:       s.app.Debugger(),
			RouteDebug:     s.routeDebug,
		})
		if err != nil {
			failedCb(s.apiServerError(s.msg("new engine failed"), err))
			return
		}
		s.logger.Info("engine initialized(default)")
	} else {
		var extRoutes []service.RouteProvider
		for _, m := range s.extRoutes {
			extRoutes = append(extRoutes, m())
		}
		server.AddExtRoute(s.engine, extRoutes)
		s.logger.Info("engine initialized(customer)")
	}
	for _, sp := range s.services {
		if n, err1 := sp(s.app.Context(), s.mux); err1 != nil {
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
	if !s.engineCus {
		go func(host string, engine *gin.Engine) {
			if err = engine.Run(host); err != nil {
				failedCb(s.apiServerError(s.msg("engine run failed, err="+err.Error()), nil))
			}
		}(s.host.String(), s.engine)
		s.logger.Info(utils.ToStr("server[", s.Host().String(), "] listen and serving..."))
	}
}

func (s *Server) initRegInfo() {
	s.regInfo = &regCenter.RegInfo{
		AppId:   s.app.ID(),
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

func (s *Server) withDocService(config *apidoc.Config) {
	if config.EndType == "" {
		config.EndType = endtype.Backend
	}
	if config.Path == "" {
		config.Path = "/doc"
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
	s.AddRoute(func() service.RouteProvider {
		s.logger.Debug("api doc service enabled")
		if s.engineCus {
			config.Path = "/" + s.id + "-docs/" + strings.TrimPrefix(config.Path, "/")
		}
		return server.DocRoute(config.Path, config.Provider)
	})
}

func (s *Server) withRpcServer(port int, autoAdd bool) *rpc.Server {
	s.rps = rpc.New(s.app, s.id, s.name, s.et, url.Host{Ip: s.host.Ip, Port: port}, rpc.Parent(s), rpc.RegEnable())
	s.rpsAdd = autoAdd

	return s.rps
}

func (s *Server) withRpcServerIns(ins *rpc.Server, autoAdd bool) *rpc.Server {
	s.rps = ins
	s.rpsAdd = autoAdd
	s.rps.AddRegInfo(s.id, s.name, s)

	return s.rps
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
