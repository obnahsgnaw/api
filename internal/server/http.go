package server

import (
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/obnahsgnaw/api/internal/middleware/authmid"
	"github.com/obnahsgnaw/api/pkg/errobj"
	"github.com/obnahsgnaw/api/service"
	"github.com/obnahsgnaw/application/pkg/debug"
	"strings"
)

type MuxConfig struct {
	Version          string
	MdProvider       *service.MethodMdProvider
	MiddlewarePds    []gin.HandlerFunc
	MuxMiddlewarePds []service.MuxRouteHandleFunc
	ExtRoutes        []service.RouteProvider
	ErrObjProvider   errobj.Provider
	Debugger         debug.Debugger
}

func NewMux() *runtime.ServeMux {
	return newMux()
}

// InitRpcHttpProxyServer 创建一个rpc服务的http代理服务
func InitRpcHttpProxyServer(e *gin.Engine, mux *runtime.ServeMux, cnf *MuxConfig) {
	initMux(mux, cnf.MdProvider, cnf.MuxMiddlewarePds, cnf.ErrObjProvider, cnf.Debugger)
	e.Use(authmid.NewRqIdMid())
	version := "/" + strings.TrimPrefix(cnf.Version, "/")
	// 设置路由
	AddExtRoute(e, cnf.ExtRoutes)
	e.GET(version, gin.WrapH(mux))
	// 代理到rpc
	e.Group(strings.TrimSuffix(version, "/")+"/*gw", cnf.MiddlewarePds...).Any("", gin.WrapH(mux))
}

func AddExtRoute(e *gin.Engine, routes []service.RouteProvider) {
	for _, rp := range routes {
		rp(e)
	}
}
