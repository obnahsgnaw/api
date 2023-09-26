package authmid

import (
	"github.com/gin-gonic/gin"
	"github.com/obnahsgnaw/api/internal/errhandler"
	"github.com/obnahsgnaw/api/internal/marshaler"
	"github.com/obnahsgnaw/api/pkg/apierr"
	"github.com/obnahsgnaw/api/service/authedapp"
	"github.com/obnahsgnaw/application/pkg/debug"
	"github.com/obnahsgnaw/application/pkg/dynamic"
)

// 1. 内部验证， 则通过header中的X-App-Id,去app服务获取相关的app信息和验证
// 2. 外部验证， 则通过header中的X-App-Id,去app服务获取相关的app信息不进行验证

// NewAppMid app middleware
func NewAppMid(manager *authedapp.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		rqId := c.Request.Header.Get("X-Request-Id")
		var app authedapp.App
		var err error
		var validate bool
		var appId string
		// validate outside, then decode the app stream
		if manager.OutsideValidate() {
			if manager.Logger() != nil && manager.Debug() {
				manager.Logger().Debug("Middleware [ App ]: outside validate")
			}
			validate = false
			appId = c.Request.Header.Get(manager.GetAuthedAppidHeaderKey())
		} else {
			// validate internal, fetch the app from provider
			if manager.Logger() != nil && manager.Debug() {
				manager.Logger().Debug("Middleware [ App ]: inside validate")
			}
			validate = true
			appId = c.Request.Header.Get(manager.GetAppidHeaderKey())
		}

		app, err = manager.Provider().GetValidApp(appId, manager.Project, validate, manager.Backend)

		if err != nil {
			if manager.Logger() != nil && manager.Debug() {
				manager.Logger().Debug("Middleware [ App ]: err=" + err.Error())
			}
			c.Abort()
			errhandler.HandlerErr(
				apierr.ToStatusError(apierr.NewUnauthorizedError(apierr.AppMidInvalid, err)),
				marshaler.GetMarshaler(c.GetHeader("Accept")),
				c.Writer,
				nil,
				manager.ErrObjProvider(),
				debug.New(dynamic.NewBool(func() bool {
					return manager.Debug()
				})),
			)
			return
		} else {
			if manager.Logger() != nil && manager.Debug() {
				manager.Logger().Debug("Middleware [ App ]: id=" + app.AppId())
			}
			c.Request.Header.Set(manager.GetAuthedAppidHeaderKey(), app.AppId())
			manager.Add(rqId, app)
		}

		c.Next()

		manager.Rm(rqId)
	}
}
