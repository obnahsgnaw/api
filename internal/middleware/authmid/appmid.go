package authmid

import (
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/obnahsgnaw/api/internal/marshaler"
	"github.com/obnahsgnaw/api/pkg/apierr"
	"github.com/obnahsgnaw/api/service/authedapp"
	"net/http"
)

// 1. 内部验证， 则通过header中的X-App-Id,去app服务获取相关的app信息和验证
// 2. 外部验证， 则通过header中的X-App-Id,去app服务获取相关的app信息不进行验证

// NewAppMid app middleware
func NewAppMid(manager *authedapp.Manager, debugCb func(msg string), errHandle func(err error, marshaler runtime.Marshaler, w http.ResponseWriter)) gin.HandlerFunc {
	if debugCb == nil {
		debugCb = func(msg string) {}
	}
	return func(c *gin.Context) {
		var app authedapp.App
		var err error
		var validate bool
		var appId string
		rqId := c.Request.Header.Get("X-Request-Id")
		// validate outside, then decode the app stream
		if manager.OutsideValidate() {
			debugCb("app-middleware: validate outside ")
			validate = false
			appId = c.Request.Header.Get(manager.AuthedAppidHeaderKey())
		} else {
			// validate internal, fetch the app from provider
			debugCb("app-middleware: validate inside")
			validate = true
			appId = c.Request.Header.Get(manager.AppidHeaderKey())
		}

		app, err = manager.Provider().GetValidApp(appId, manager.Project, validate)

		if err != nil {
			debugCb("app-middleware: validate failed,err=" + err.Error())
			c.Abort()
			errHandle(
				apierr.ToStatusError(apierr.NewUnauthorizedError(apierr.AppMidInvalid, err)),
				marshaler.GetMarshaler(c.GetHeader("Accept")),
				c.Writer,
			)
			return
		} else {
			debugCb("app-middleware: accessed, id=" + app.AppId())
			c.Request.Header.Set(manager.AuthedAppidHeaderKey(), app.AppId())
			manager.Add(rqId, app)
		}

		c.Next()

		manager.Rm(rqId)
	}
}
