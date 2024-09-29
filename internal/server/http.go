package server

import (
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/obnahsgnaw/api/internal/middleware/authmid"
	"github.com/obnahsgnaw/api/service"
	"strings"
)

type StaticRoute map[string]map[string]struct{}

func NewStaticRoute() StaticRoute {
	return make(map[string]map[string]struct{})
}

func (s StaticRoute) Add(meth string, path string) {
	meth = strings.ToUpper(meth)
	if _, ok := s[meth]; !ok {
		s[meth] = make(map[string]struct{})
	}
	s[meth][path] = struct{}{}
}
func (s StaticRoute) Encode(str string) string {
	if str == "" {
		return str
	}
	return base64.URLEncoding.EncodeToString([]byte(str))
}
func (s StaticRoute) Decode(str string) string {
	if str == "" {
		return str
	}
	b, _ := base64.URLEncoding.DecodeString(str)
	return string(b)
}

func (s StaticRoute) Match(c *gin.Context) {
	meth := strings.ToUpper(c.Request.Method)
	if v, ok := s[meth]; ok {
		for pt := range v {
			if strings.HasPrefix(c.Request.RequestURI, pt) {
				c.Request.RequestURI = pt + "/" + s.Encode(strings.TrimPrefix(c.Request.RequestURI, pt+"/"))
				c.Request.URL.Path = pt + "/" + s.Encode(strings.TrimPrefix(c.Request.URL.Path, pt+"/"))
				if c.Request.URL.RawPath != "" {
					c.Request.URL.RawPath = pt + "/" + s.Encode(strings.TrimPrefix(c.Request.URL.RawPath, pt+"/"))
				}
				return
			}
		}
	}
}

// InitRpcHttpProxyServer 创建一个rpc服务的http代理服务
func InitRpcHttpProxyServer(e *gin.Engine, mux *runtime.ServeMux, project string, version string, middlewares []gin.HandlerFunc, staticRoutes StaticRoute, withoutRoutePrefix bool) {
	version = "/" + strings.Trim(version, "/")
	prefix := version + "/" + project
	if withoutRoutePrefix {
		prefix = version
	}
	middlewares = append([]gin.HandlerFunc{authmid.NewRqIdMid(), replaceMid(prefix, version, staticRoutes, withoutRoutePrefix)}, middlewares...)
	e.GET(prefix, append(append([]gin.HandlerFunc{}, middlewares...), gin.WrapH(mux))...)
	e.Group(prefix+"/*gw", middlewares...).Any("", gin.WrapH(mux))
}

func AddExtRoute(e *gin.Engine, routes []service.RouteProvider) {
	for _, rp := range routes {
		rp(e)
	}
}
func replaceMid(prefix, version string, staticRoute StaticRoute, withoutRoutePrefix bool) gin.HandlerFunc {
	return func(context *gin.Context) {
		if !withoutRoutePrefix && strings.HasPrefix(context.Request.RequestURI, prefix) {
			context.Request.RequestURI = strings.Replace(context.Request.RequestURI, prefix, version, 1)
			context.Request.URL.Path = strings.Replace(context.Request.URL.Path, prefix, version, 1)
			if context.Request.URL.RawPath != "" {
				context.Request.URL.RawPath = strings.Replace(context.Request.URL.RawPath, prefix, version, 1)
			}
		}
		staticRoute.Match(context)
	}
}
