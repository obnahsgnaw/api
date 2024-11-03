package authmid

import (
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/obnahsgnaw/api/internal/marshaler"
	"github.com/obnahsgnaw/api/pkg/apierr"
	"github.com/obnahsgnaw/api/service/authedapp"
	"net/http"
)

// 通过header中的X-App-Id,去app服务获取相关的app信息和验证

// NewAppMid app middleware
func NewAppMid(manager *authedapp.Manager, debugCb func(msg string), errHandle func(err error, marshaler runtime.Marshaler, w http.ResponseWriter)) gin.HandlerFunc {
	if debugCb == nil {
		debugCb = func(msg string) {}
	}
	return func(c *gin.Context) {
		var app authedapp.App
		var err error
		rqId := c.Request.Header.Get("X-Request-Id")
		rqType := c.Request.Header.Get("X-Request-Type")
		logPrefix := "app-middleware[" + rqType + "." + rqId + "]: "
		appId := c.Request.Header.Get(manager.AppidHeaderKey())

		if ig, igApp := manager.Ignored(c); !ig {
			if app, err = manager.Provider().GetValidApp(rqId, appId, manager.Project, true); err != nil {
				debugCb(logPrefix + "validate failed,err=" + err.Error())
				c.Abort()
				errHandle(
					apierr.ToStatusError(apierr.NewUnauthorizedError(apierr.AppMidInvalid, err).WithRequestTypeAndId(rqType, rqId)),
					marshaler.GetMarshaler(c.GetHeader("Accept")),
					c.Writer,
				)
				return
			}
		} else {
			app = igApp
			debugCb(logPrefix + "validate ignored by ignorer")
			c.Request.Header.Set("X-App-ReplacedId", appId)
		}
		c.Request.Header.Set(manager.AppidHeaderKey(), app.AppId())

		debugCb(logPrefix + "accessed, id=" + app.AppId())
		manager.Add(rqId, app)

		c.Next()

		manager.Rm(rqId)
	}
}
