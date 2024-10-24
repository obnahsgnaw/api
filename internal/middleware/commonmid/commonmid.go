package commonmid

import (
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/obnahsgnaw/api/internal/marshaler"
	"github.com/obnahsgnaw/api/pkg/apierr"
	"net/http"
)

func NewCommonMid(handler func(c *gin.Context, rqId, rqType string, debugger func(string)) error, debugCb func(msg string), errHandle func(err error, marshaler runtime.Marshaler, w http.ResponseWriter)) gin.HandlerFunc {
	if debugCb == nil {
		debugCb = func(msg string) {}
	}
	return func(c *gin.Context) {
		rqId := c.Request.Header.Get("X-Request-ID")
		rqType := c.Request.Header.Get("X-Request-Type")

		if err := handler(c, rqId, rqType, debugCb); err != nil {
			c.Abort()
			errHandle(
				apierr.ToStatusError(apierr.NewUnauthorizedError(apierr.AuthMidInvalid, err).WithRequestTypeAndId(rqType, rqId)),
				marshaler.GetMarshaler(c.Request.Header.Get("Accept")),
				c.Writer,
			)
			return
		}

		c.Next()
	}
}
