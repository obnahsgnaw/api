package authmid

import (
	"github.com/gin-gonic/gin"
	"github.com/obnahsgnaw/api/internal/errhandler"
	"github.com/obnahsgnaw/api/internal/marshaler"
	"github.com/obnahsgnaw/api/pkg/apierr"
	"github.com/obnahsgnaw/api/service/autheduser"
	"strconv"
)

// NewAuthMid  authentication  middleware
func NewAuthMid(manager *autheduser.Manager, debugCb func(msg string)) gin.HandlerFunc {
	if debugCb == nil {
		debugCb = func(msg string) {}
	}
	return func(c *gin.Context) {
		rqId := c.Request.Header.Get("X-Request-ID")
		appId := c.Request.Header.Get(manager.AppIdHeaderKey())
		token := c.Request.Header.Get(manager.TokenHeaderKey())
		if token != "" {
			var err error
			var user autheduser.User
			// validate outside, fetch the uid user from provide
			if manager.OutsideValidate() {
				debugCb("auth-middleware: validate outside")
				uid := c.Request.Header.Get(manager.UserIdHeaderKey())
				user, err = manager.Provider().GetIdUser(appId, uid)
			} else {
				// validate inside, fetch the token user from provider
				debugCb("auth-middleware: validate inside")
				user, err = manager.Provider().GetTokenUser(appId, token)
			}

			if err != nil {
				debugCb("auth-middleware: validate failed, err=" + err.Error())
				c.Abort()
				errhandler.DefaultErrorHandler(
					apierr.ToStatusError(apierr.NewUnauthorizedError(apierr.AuthMidInvalid, err)),
					marshaler.GetMarshaler(c.Request.Header.Get("Accept")),
					c.Writer,
				)
				return
			}

			debugCb("auth-middleware: accessed, user=" + strconv.Itoa(int(user.Id())))
			c.Request.Header.Set(manager.UserIdHeaderKey(), user.Uid())

			manager.Add(rqId, user)
		} else {
			debugCb("auth-middleware: token empty, ignored")
		}

		c.Next()
		manager.Rm(rqId)
	}
}
