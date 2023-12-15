package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/obnahsgnaw/api/internal/middleware/authmid"
	"github.com/obnahsgnaw/api/pkg/corsmid"
	"github.com/obnahsgnaw/api/pkg/errobj"
	"github.com/obnahsgnaw/api/service"
	"github.com/obnahsgnaw/api/service/cors"
	"github.com/obnahsgnaw/application/pkg/debug"
	"io"
	"strings"
	"time"
)

type EngineConfig struct {
	Debug          bool
	AccessWriter   io.Writer
	ErrWriter      io.Writer
	TrustedProxies []string
	Cors           *cors.Config
}

type HttpConfig struct {
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
}

func NewEngine(cnf *EngineConfig) (*gin.Engine, error) {
	if cnf.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	if cnf.AccessWriter != nil {
		gin.DisableConsoleColor()
	} else {
		gin.ForceConsoleColor()
	}
	r := gin.New()

	if len(cnf.TrustedProxies) > 0 {
		if err := r.SetTrustedProxies(cnf.TrustedProxies); err != nil {
			return nil, err
		}
	}

	if cnf.AccessWriter != nil {
		r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
			Formatter: func(param gin.LogFormatterParams) string {
				return fmt.Sprintf("[ %s ] - %s %s %s %d %s %v %s %s\n",
					param.TimeStamp.Format(time.RFC3339),
					param.ClientIP,
					param.Method,
					param.Path,
					param.StatusCode,
					param.Latency,
					param.Request.Body,
					param.Request.UserAgent(),
					param.ErrorMessage,
				)
			},
			Output: cnf.AccessWriter,
		}))
	} else {
		r.Use(gin.Logger())
	}
	if cnf.ErrWriter != nil {
		r.Use(gin.RecoveryWithWriter(cnf.ErrWriter))
	} else {
		r.Use(gin.Recovery())
	}

	if cnf.Cors != nil {
		r.Use(corsmid.NewCorsMid(func() *cors.Config {
			return cnf.Cors
		}))
	}
	r.Use(authmid.NewRqIdMid())

	return r, nil
}

// NewRpcHttpProxyServer 创建一个rpc服务的http代理服务
func NewRpcHttpProxyServer(cnf *HttpConfig) (e *gin.Engine, mux *runtime.ServeMux, err error) {
	mux = getRpcApiProxyMux(cnf.MdProvider, cnf.MuxMiddleware, cnf.ErrObjProvider, cnf.Debugger)

	// 初始gin
	if e, err = NewEngine(&EngineConfig{
		Debug:          cnf.RouteDebug,
		AccessWriter:   cnf.AccessWriter,
		ErrWriter:      cnf.ErrWriter,
		TrustedProxies: cnf.TrustedProxies,
		Cors:           cnf.Cors,
	}); err != nil {
		return
	}
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
