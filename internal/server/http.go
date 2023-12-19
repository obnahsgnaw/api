package server

import (
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/obnahsgnaw/api/internal/middleware/authmid"
	"github.com/obnahsgnaw/api/pkg/errobj"
	"github.com/obnahsgnaw/api/service"
	"github.com/obnahsgnaw/application/pkg/debug"
	"github.com/obnahsgnaw/http"
	"strings"
)

type MuxConfig struct {
	PathPrefix     string
	MdProvider     *service.MethodMdProvider
	Middlewares    []gin.HandlerFunc
	MuxMiddleware  []service.MuxRouteHandleFunc
	ExtRoutes      []service.RouteProvider
	ErrObjProvider errobj.Provider
	Debugger       debug.Debugger
}

func NewEngine(config *http.Config) (e *gin.Engine, err error) {
	return http.New(config)
}

func NewMux() *runtime.ServeMux {
	return newMux()
}

// InitRpcHttpProxyServer 创建一个rpc服务的http代理服务
func InitRpcHttpProxyServer(e *gin.Engine, mux *runtime.ServeMux, cnf *MuxConfig) {
	initMux(mux, cnf.MdProvider, cnf.MuxMiddleware, cnf.ErrObjProvider, cnf.Debugger)
	e.Use(authmid.NewRqIdMid())
	prefix := "/" + strings.TrimPrefix(cnf.PathPrefix, "/")
	// 设置路由
	e.GET(prefix, gin.WrapH(mux))
	// 设置其他路由
	AddExtRoute(e, cnf.ExtRoutes)
	// 代理到rpc
	e.Group(strings.TrimSuffix(prefix, "/")+"/*gw", cnf.Middlewares...).Any("", gin.WrapH(mux))
}

func AddExtRoute(e *gin.Engine, routes []service.RouteProvider) {
	for _, rp := range routes {
		rp(e)
	}
}
