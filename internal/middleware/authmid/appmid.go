package authmid

import (
	"github.com/gin-gonic/gin"
	"github.com/obnahsgnaw/api/internal/errhandler"
	"github.com/obnahsgnaw/api/internal/marshaler"
	"github.com/obnahsgnaw/api/pkg/apierr"
	"github.com/obnahsgnaw/api/service/authedapp"
)

// NewAppMid app middleware
func NewAppMid(manager *authedapp.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		rqId := c.Request.Header.Get("X-Request-Id")
		appId := c.Request.Header.Get("X-App-Id")
		if app, err := manager.Provider().GetValidApp(appId, manager.Project, !manager.OutsideValidate, manager.Backend); err != nil {
			if manager.Logger() != nil {
				manager.Logger().Debug("Middleware [ App ]: err=" + err.Error())
			}
			c.Abort()
			errhandler.HandlerErr(
				apierr.ToStatusError(apierr.NewUnauthorizedError(apierr.AppMidInvalid, err)),
				marshaler.GetMarshaler(c.GetHeader("Accept")),
				c.Writer,
				nil,
				manager.ErrObjProvider(),
				manager,
			)
			return
		} else {
			if manager.Logger() != nil {
				manager.Logger().Debug("Middleware [ App ]: id=" + app.AppId())
			}
			manager.Add(rqId, app)
		}

		c.Next()

		manager.Rm(rqId)
	}
}
