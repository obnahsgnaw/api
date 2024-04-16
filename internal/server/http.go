package server

import (
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/obnahsgnaw/api/internal/middleware/authmid"
	"github.com/obnahsgnaw/api/service"
	"strings"
)

type MuxConfig struct {
	Version       string
	PrefixReplace string // gin 路由前缀 但是 导mux后需要替换为空，用于多个服务板块以前缀区分进行注册路由，要不然都是v1注册不了
	MiddlewarePds []gin.HandlerFunc
}

// InitRpcHttpProxyServer 创建一个rpc服务的http代理服务
func InitRpcHttpProxyServer(e *gin.Engine, mux *runtime.ServeMux, cnf *MuxConfig) {
	e.Use(authmid.NewRqIdMid())
	cnf.Version = "/" + strings.Trim(cnf.Version, "/")
	version := cnf.Version
	if cnf.PrefixReplace != "" {
		version = version + "/" + cnf.PrefixReplace
		cnf.MiddlewarePds = append([]gin.HandlerFunc{replaceMid(version, cnf.Version)}, cnf.MiddlewarePds...)
	}
	e.GET(version, append(append([]gin.HandlerFunc{}, cnf.MiddlewarePds...), gin.WrapH(mux))...)
	// 代理到rpc
	e.Group(version+"/*gw", cnf.MiddlewarePds...).Any("", gin.WrapH(mux))
}

func AddExtRoute(e *gin.Engine, routes []service.RouteProvider) {
	for _, rp := range routes {
		rp(e)
	}
}
func replaceMid(prefixedVersion, rawVersion string) gin.HandlerFunc {
	return func(context *gin.Context) {
		if strings.HasPrefix(context.Request.RequestURI, prefixedVersion) {
			context.Request.RequestURI = strings.Replace(context.Request.RequestURI, prefixedVersion, rawVersion, 1)
		}
		if strings.HasPrefix(context.Request.URL.Path, prefixedVersion) {
			context.Request.URL.Path = strings.Replace(context.Request.URL.Path, prefixedVersion, rawVersion, 1)
		}
	}
}
