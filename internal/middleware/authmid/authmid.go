package authmid

import (
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/obnahsgnaw/api/internal/marshaler"
	"github.com/obnahsgnaw/api/pkg/apierr"
	"github.com/obnahsgnaw/api/service/autheduser"
	"net/http"
	"strconv"
)

// NewAuthMid  authentication  middleware
func NewAuthMid(manager *autheduser.Manager, debugCb func(msg string), errHandle func(err error, marshaler runtime.Marshaler, w http.ResponseWriter)) gin.HandlerFunc {
	if debugCb == nil {
		debugCb = func(msg string) {}
	}
	return func(c *gin.Context) {
		var err error
		var user autheduser.User
		rqId := c.Request.Header.Get("X-Request-ID")
		rqType := c.Request.Header.Get("X-Request-Type")
		logPrefix := "auth-middleware[" + rqType + "." + rqId + "]: "
		appId := c.Request.Header.Get(manager.AppIdHeaderKey())
		token := c.Request.Header.Get(manager.TokenHeaderKey())

		if !manager.Ignored(c) {
			if token != "" {
				if user, err = manager.Provider().GetTokenUser(rqId, appId, token); err != nil {
					debugCb(logPrefix + "validate failed, err=" + err.Error())
					c.Abort()
					errHandle(
						apierr.ToStatusError(apierr.NewUnauthorizedError(apierr.AuthMidInvalid, err).WithRequestTypeAndId(rqType, rqId)),
						marshaler.GetMarshaler(c.Request.Header.Get("Accept")),
						c.Writer,
					)
					return
				}

				debugCb(logPrefix + "accessed, user=" + strconv.Itoa(int(user.Id())))
				c.Request.Header.Set(manager.UserIdHeaderKey(), user.Uid())

				manager.Add(rqId, user)
			} else {
				debugCb(logPrefix + "token empty, ignored")
			}
		} else {
			debugCb(logPrefix + "validate ignored by ignorer")
		}

		c.Next()
		manager.Rm(rqId)
	}
}
