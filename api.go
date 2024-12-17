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
	"strconv"
	"strings"
)

type Version int

func (v Version) String() string {
	return "v" + strconv.Itoa(int(v))
}

// Server API server
type Server struct {
	app                *application.Application
	id                 string // 模块
	name               string
	endType            endtype.EndType
	serverType         servertype.ServerType
	httpEngine         *engine.MuxHttp
	errFactory         *apierr.Factory
	rpcServer          *rpc.Server
	logger             *zap.Logger
	logCnf             *logger.Config
	services           []ServiceProvider
	regInfo            *regCenter.RegInfo
	docRegInfos        []*regCenter.RegInfo
	mdProvider         *service.MethodMdProvider
	errObjProvider     errobj.Provider
	version            Version
	regEnable          bool
	middlewarePds      map[string]func() gin.HandlerFunc
	muxMiddlewarePds   map[string]func() service.MuxRouteHandleFunc
	extRoutePds        []func() service.RouteProvider
	gatewayKeyGen      func() (string, error)
	gatewayKey         string
	running            bool
	muxRoutes          []func(*runtime.ServeMux) error
	staticRoutes       server.StaticRoute
	withoutRoutePrefix bool
}

// ServiceProvider api service provider
type ServiceProvider func(ctx context.Context, mux *runtime.ServeMux) (name string, err error)

func New(app *application.Application, e *engine.MuxHttp, id, name string, et endtype.EndType, v Version, options ...Option) *Server {
	s := &Server{
		app:        app,
		id:         id,
		name:       name,
		endType:    et,
		serverType: servertype.Api,
		httpEngine: e,
		version:    v,
		errFactory: apierr.New(1),
		errObjProvider: func(param errobj.Param) interface{} {
			return param
		},
		mdProvider:       service.NewMdProvider(),
		middlewarePds:    make(map[string]func() gin.HandlerFunc),
		muxMiddlewarePds: make(map[string]func() service.MuxRouteHandleFunc),
		logCnf:           app.LogConfig(),
		logger:           app.Logger().Named(utils.ToStr(id, "-", et.String(), "-", servertype.Api.String())),
		staticRoutes:     server.NewStaticRoute(),
	}
	e.Http().AddInitializer(s.initHttp)
	s.With(options...)
	return s
}

func logCnf(app *application.Application, et endtype.EndType, id string) *logger.Config {
	cnf := logger.CopyCnfWithLevel(app.LogConfig())
	if cnf != nil {
		cnf.SetFilename(utils.ToStr(servertype.Api.String(), "-", et.String(), "-", id))
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
	return url.Host{
		Ip:   s.httpEngine.Http().Ip(),
		Port: s.httpEngine.Http().Port(),
	}
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
	return s.httpEngine
}

func (s *Server) Rpc() *rpc.Server {
	return s.rpcServer
}

// RegisterApiService register a api service
func (s *Server) RegisterApiService(provider ServiceProvider) {
	s.services = append(s.services, provider)
}

// RegisterRpcService register a rcp service
func (s *Server) RegisterRpcService(provider rpc.ServiceInfo) {
	if s.rpcServer != nil {
		s.rpcServer.RegisterService(provider)
	}
}

func (s *Server) AddMiddleware(name string, mid func() gin.HandlerFunc, force bool) {
	if _, ok := s.middlewarePds[name]; ok && force || !ok {
		if s.logger != nil {
			s.logger.Debug(name + "middleware enabled")
		}
		s.middlewarePds[name] = mid
	}
}

func (s *Server) AddMuxMiddleware(name string, mid func() service.MuxRouteHandleFunc, force bool) {
	if _, ok := s.muxMiddlewarePds[name]; ok && force || !ok {
		if s.logger != nil {
			s.logger.Debug(name + "middleware enabled")
		}
		s.muxMiddlewarePds[name] = mid
	}
}

func (s *Server) AddRoute(route func() service.RouteProvider) {
	s.extRoutePds = append(s.extRoutePds, route)
}

// AddMuxRoute MuxRoute need add id prefix to access
func (s *Server) AddMuxRoute(meth string, pathPattern string, h func(w http.ResponseWriter, r *http.Request, pathParams map[string]string)) {
	s.muxRoutes = append(s.muxRoutes, func(mux *runtime.ServeMux) error {
		return mux.HandlePath(strings.ToUpper(meth), pathPattern, h)
	})
}

// AddMuxStaticRoute 文件路由，pathPattern 需要为 /xx/{path}
func (s *Server) AddMuxStaticRoute(meth string, pathPattern string, h func(w http.ResponseWriter, r *http.Request, pathParams map[string]string)) {
	s.staticRoutes.Add(meth, pathPattern)
	s.AddMuxRoute(meth, pathPattern+"/{path}", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		if _, ok := pathParams["path"]; ok {
			pathParams["path"] = s.staticRoutes.Decode(pathParams["path"])
		}
		h(w, r, pathParams)
	})
}

func (s *Server) AddDefIncomeMd(key string, valProvider service.MdValParser) {
	s.mdProvider.AddDefault(key, valProvider)
}

func (s *Server) AddAllIncomeMd() {
	s.mdProvider.AddAll()
}

func (s *Server) AddIncomeMd(method, key string, valProvider service.MdValParser) {
	s.mdProvider.Add(method, key, valProvider)
}

func (s *Server) AddMethodAllIncomeMd(method string) {
	s.mdProvider.AddMethodAll(method)
}

// ErrCode return err code factory
func (s *Server) ErrCode() *apierr.Factory {
	return s.errFactory
}

func (s *Server) Run(failedCb func(error)) {
	if s.running {
		return
	}
	var err error
	s.logger.Info("init start...")
	s.initRegInfo()
	if s.rpcServer != nil {
		s.app.AddServer(s.rpcServer)
		s.logger.Debug("api rpc service enabled")
		if s.Engine().Http().Config() != nil {
			if s.Engine().Http().Config().AccessWriter != nil {
				s.rpcServer.With(rpc.AccessWriter(s.Engine().Http().Config().AccessWriter))
			}
			if s.Engine().Http().Config().ErrWriter != nil {
				s.rpcServer.With(rpc.ErrorWriter(s.Engine().Http().Config().ErrWriter))
			}
		}
	}
	if err = s.initEngine(); err != nil {
		failedCb(err)
		return
	}
	for _, sp := range s.services {
		if n, err1 := sp(s.app.Context(), s.httpEngine.Mux()); err1 != nil {
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
	go func() {
		defer s.httpEngine.Http().CloseWithKey(s.id)
		s.httpEngine.Http().RunAndServWithKey(s.id, func(err error) {
			failedCb(s.apiServerError(s.msg("engine run failed, err="+err.Error()), nil))
		})
	}()
	s.logger.Info(utils.ToStr("server[", s.Host().String(), "] listen and serving..."))
	s.running = true
}

func (s *Server) initHttp() error {
	var mid []gin.HandlerFunc
	for _, m := range s.middlewarePds {
		mid = append(mid, m())
	}
	server.InitRpcHttpProxyServer(s.httpEngine.Http().Engine(), s.httpEngine.Mux(), s.id, s.version.String(), mid, s.staticRoutes, s.withoutRoutePrefix)

	var extRoutes []service.RouteProvider
	for _, m := range s.extRoutePds {
		extRoutes = append(extRoutes, m())
	}
	server.AddExtRoute(s.httpEngine.Http().Engine(), extRoutes)
	s.logger.Info("engine initialized")
	return nil
}

func (s *Server) initEngine() error {
	var mmid []service.MuxRouteHandleFunc
	for _, m := range s.muxMiddlewarePds {
		mmid = append(mmid, m())
	}
	server.InitMux(s.httpEngine.Mux(), s.mdProvider, mmid, s.errObjProvider, s.app.Debugger())

	for _, m := range s.muxRoutes {
		if err := m(s.httpEngine.Mux()); err != nil {
			return err
		}
	}
	s.logger.Info("mux initialized")
	return nil
}

func (s *Server) initRegInfo() {
	s.regInfo = &regCenter.RegInfo{
		AppId:   s.app.Cluster().Id(),
		RegType: regtype.Http,
		ServerInfo: regCenter.ServerInfo{
			Id:      s.id,
			Name:    s.name,
			EndType: s.endType.String(),
			Type:    s.serverType.String(),
		},
		Host:      s.httpEngine.Http().Host(),
		Val:       s.httpEngine.Http().Host(),
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
	s.running = false
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
		Host: url.Host{
			Ip:   s.httpEngine.Http().Ip(),
			Port: s.httpEngine.Http().Port(),
		},
	})
	docRegInfo := &regCenter.RegInfo{
		AppId:   s.app.Cluster().Id(),
		RegType: regtype.Doc,
		ServerInfo: regCenter.ServerInfo{
			Id:      s.id,
			Name:    s.name,
			Type:    s.serverType.String(),
			EndType: config.EndType.String(),
		},
		Host: s.httpEngine.Http().Host(),
		Val:  "",
		Values: map[string]string{
			"url":         config.Url(),
			"debugOrigin": config.DebugOrigin.String(),
			"title":       config.Title,
			"sort":        strconv.Itoa(config.Sort),
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
			Host: url.Host{
				Ip:   s.httpEngine.Http().Ip(),
				Port: s.httpEngine.Http().Port(),
			},
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
			if s.gatewayKey != "" {
				_ = s.app.Register().Unregister(s.app.Context(), s.gatewayKey)
			}
			s.gatewayKey = key
			if err = s.app.Register().Register(s.app.Context(), s.gatewayKey, url.Origin{
				Protocol: url.HTTP,
				Host: url.Host{
					Ip:   s.httpEngine.Http().Ip(),
					Port: s.httpEngine.Http().Port(),
				},
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
