package authmid

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/obnahsgnaw/api/internal/marshaler"
	"github.com/obnahsgnaw/api/pkg/apierr"
	"github.com/obnahsgnaw/api/service/sign"
	"net/http"
	"strings"
)

// if validate outside, then not user the mid

func NewSignMid(manager *sign.Manager, debugCb func(msg string), errHandle func(err error, marshaler runtime.Marshaler, w http.ResponseWriter)) gin.HandlerFunc {
	if debugCb == nil {
		debugCb = func(msg string) {}
	}
	return func(c *gin.Context) {
		appId := c.GetHeader(manager.AppIdHeaderKey())
		userId := c.GetHeader(manager.UserIdHeaderKey())
		signStr := c.GetHeader(manager.UserSignHeaderKey())
		method := c.Request.Method
		uri := c.Request.URL.Path

		debugCb("sign-middleware: sign in=" + signStr)
		s, t, n, err := parseSignStr(signStr)
		if err != nil {
			debugCb("sign-middleware: validate failed, err=" + err.Error())
			c.Abort()
			errHandle(
				apierr.ToStatusError(apierr.NewBadRequestError(apierr.SignMidInvalid, err)),
				marshaler.GetMarshaler(c.GetHeader("Accept")),
				c.Writer,
			)
			return
		}
		// sign check
		if err = manager.Provider().Validate(appId, userId, method, uri, s, t, n); err != nil {
			debugCb("sign-middleware: validate failed, err=" + err.Error())
			c.Abort()
			errHandle(
				apierr.ToStatusError(apierr.NewBadRequestError(apierr.SignMidInvalid, err)),
				marshaler.GetMarshaler(c.GetHeader("Accept")),
				c.Writer,
			)
			return
		}
		c.Next()
		// sign generate
		if s1, t1, n1, err1 := manager.Provider().Generate(appId, userId, method, uri); err1 != nil {
			debugCb("sign-middleware: gen failed, err=" + err1.Error())
			c.Abort()
			errHandle(
				apierr.ToStatusError(apierr.NewBadRequestError(apierr.SignMidGenFailed, err1)),
				marshaler.GetMarshaler(c.GetHeader("Accept")),
				c.Writer,
			)
			return
		} else {
			signStr1 := toSignStr(s1, t1, n1)
			debugCb("sign-middleware: sign out=" + signStr1)
			c.Header(manager.UserSignHeaderKey(), signStr1)
		}
	}
}

func parseSignStr(str string) (sign, timestamp, nonce string, err error) {
	if !strings.Contains(str, "-") {
		err = errors.New("signature format error")
		return
	}
	segments := strings.Split(str, "-")
	if len(segments) != 3 {
		err = errors.New("signature not contain 3 segment")
	}
	sign = segments[0]
	timestamp = segments[1]
	nonce = segments[2]
	return
}

func toSignStr(sign, timestamp, nonce string) string {
	return strings.Join([]string{sign, timestamp, nonce}, "-")
}
