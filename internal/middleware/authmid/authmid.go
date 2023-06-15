package authmid

import (
	"github.com/gin-gonic/gin"
	"github.com/obnahsgnaw/api/internal/errhandler"
	"github.com/obnahsgnaw/api/internal/marshaler"
	"github.com/obnahsgnaw/api/pkg/apierr"
	"github.com/obnahsgnaw/api/service"
	"github.com/obnahsgnaw/api/service/autheduser"
	"net/http"
)

// NewMuxAuthBeforeMid auth middleware
func NewMuxAuthBeforeMid(manager *autheduser.Manager) service.MuxRouteHandleFunc {
	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string, pattern string) bool {
		method := r.Method
		uri := pattern
		rqId := r.Header.Get("X-Request-ID")
		appId := r.Header.Get("X-App-Id")
		token := r.Header.Get("Authorization")
		if manager.AuthedRouteManager().AuthMust(method, uri) {
			if manager.Logger() != nil {
				manager.Logger().Debug("Middleware [Auth ]: " + method + ":" + uri)
			}
			var err error
			var user autheduser.User
			// 外部验证 验证后在头写入用户数据
			if manager.OutsideValidate {
				if manager.Logger() != nil {
					manager.Logger().Debug("Middleware [Auth ]: auth outside")
				}
				user, err = manager.Provider().GetValidUser(r.Header.Get("X-User-Id"))
			} else {
				if manager.Logger() != nil {
					manager.Logger().Debug("Middleware [Auth]: auth inside")
				}

				if user, err = manager.Provider().GetValidTokenUser(appId, token); err == nil {
					r.Header.Set("X-User-Id", user.Uid())
				}
			}

			if err != nil {
				if manager.Logger() != nil {
					manager.Logger().Debug("Middleware [Auth ]: user invalid, err=" + err.Error())
				}
				errhandler.HandlerErr(
					apierr.ToStatusError(apierr.NewUnauthorizedError(apierr.AuthMidInvalid, err)),
					marshaler.GetMarshaler(r.Header.Get("Accept")),
					w,
					nil,
					manager.ErrObjProvider(),
					manager,
				)
				return false
			}
			manager.Add(rqId, user)

		} else {
			if manager.Logger() != nil {
				manager.Logger().Debug("Middleware [Auth ]: not need auth")
			}
		}
		return true
	}
}

// NewAuthAfterMid clear auth temp mid
func NewAuthAfterMid(manager *autheduser.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		rqId := c.Request.Header.Get("X-Request-ID")
		c.Next()
		manager.Rm(rqId)
	}
}
