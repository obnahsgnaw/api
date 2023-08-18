package authmid

import (
	"github.com/gin-gonic/gin"
	"github.com/obnahsgnaw/api/internal/errhandler"
	"github.com/obnahsgnaw/api/internal/marshaler"
	"github.com/obnahsgnaw/api/pkg/apierr"
	"github.com/obnahsgnaw/api/service"
	"github.com/obnahsgnaw/api/service/autheduser"
	"github.com/obnahsgnaw/application/pkg/debug"
	"github.com/obnahsgnaw/application/pkg/dynamic"
	"net/http"
	"strconv"
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
			if manager.Logger() != nil && manager.Debug() {
				manager.Logger().Debug("Middleware [Auth ]: " + method + ":" + uri)
			}
			var err error
			var user autheduser.User
			// validate outside, decode the user data
			if manager.OutsideValidate() {
				if manager.Logger() != nil && manager.Debug() {
					manager.Logger().Debug("Middleware [Auth ]: outside validate")
				}
				userStream := r.Header.Get(manager.OutsideHandler().Key)
				user, err = manager.OutsideHandler().Decode([]byte(userStream))
			} else {
				// validate internal, fetch the user from provider
				if manager.Logger() != nil && manager.Debug() {
					manager.Logger().Debug("Middleware [Auth ]: internal validate")
				}
				user, err = manager.Provider().GetValidTokenUser(appId, token)
			}

			if err != nil {
				if manager.Logger() != nil && manager.Debug() {
					manager.Logger().Debug("Middleware [Auth ]: user invalid, err=" + err.Error())
				}
				errhandler.HandlerErr(
					apierr.ToStatusError(apierr.NewUnauthorizedError(apierr.AuthMidInvalid, err)),
					marshaler.GetMarshaler(r.Header.Get("Accept")),
					w,
					nil,
					manager.ErrObjProvider(),
					debug.New(dynamic.NewBool(func() bool {
						return manager.Debug()
					})),
				)
				return false
			} else {
				if manager.Logger() != nil && manager.Debug() {
					manager.Logger().Debug("Middleware [Auth ]: user " + strconv.Itoa(int(user.Id())))
				}
				r.Header.Set("X-User-Id", user.Uid())
			}
			manager.Add(rqId, user)
		} else {
			if manager.Logger() != nil && manager.Debug() {
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
