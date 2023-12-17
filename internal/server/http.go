package server

import (
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/obnahsgnaw/api/internal/middleware/authmid"
	"github.com/obnahsgnaw/api/pkg/errobj"
	"github.com/obnahsgnaw/api/service"
	"github.com/obnahsgnaw/application/pkg/debug"
	"github.com/obnahsgnaw/application/pkg/logging/logger"
	"github.com/obnahsgnaw/http"
	"github.com/obnahsgnaw/http/cors"
	"io"
	"strings"
)

type HttpConfig struct {
	Name           string
	PathPrefix     string
	AccessWriter   io.Writer
	ErrWriter      io.Writer
	TrustedProxies []string
	Cors           *cors.Config
	MdProvider     *service.MethodMdProvider
	Middlewares    []gin.HandlerFunc
	MuxMiddleware  []service.MuxRouteHandleFunc
	ExtRoutes      []service.RouteProvider
	ErrObjProvider errobj.Provider
	Debugger       debug.Debugger
	RouteDebug     bool
	LogCnf         *logger.Config
}

// NewRpcHttpProxyServer 创建一个rpc服务的http代理服务
func NewRpcHttpProxyServer(cnf *HttpConfig) (e *gin.Engine, mux *runtime.ServeMux, err error) {
	mux = getRpcApiProxyMux(cnf.MdProvider, cnf.MuxMiddleware, cnf.ErrObjProvider, cnf.Debugger)

	// 初始gin
	if e, err = http.New(&http.Config{
		Name:           cnf.Name,
		DebugMode:      cnf.RouteDebug,
		LogDebug:       cnf.Debugger.Debug(),
		AccessWriter:   cnf.AccessWriter,
		ErrWriter:      cnf.ErrWriter,
		TrustedProxies: cnf.TrustedProxies,
		Cors:           cnf.Cors,
		LogCnf:         cnf.LogCnf,
	}); err != nil {
		return
	}
	e.Use(authmid.NewRqIdMid())
	prefix := "/" + strings.TrimPrefix(cnf.PathPrefix, "/")
	// 设置路由
	e.GET(prefix, gin.WrapH(mux))
	// 设置其他路由
	AddExtRoute(e, cnf.ExtRoutes)
	// 代理到rpc
	e.Group(strings.TrimSuffix(prefix, "/")+"/*gw", cnf.Middlewares...).Any("", gin.WrapH(mux))

	return
}

func AddExtRoute(e *gin.Engine, routes []service.RouteProvider) {
	for _, rp := range routes {
		rp(e)
	}
}
