package authmid

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/obnahsgnaw/api/internal/marshaler"
	"github.com/obnahsgnaw/api/pkg/apierr"
	"github.com/obnahsgnaw/api/service/crypt"
	"io"
	"net/http"
)

type BodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w BodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func NewCryptMid(manager *crypt.Manager, debugCb func(msg string), errHandle func(err error, marshaler runtime.Marshaler, w http.ResponseWriter)) gin.HandlerFunc {
	if debugCb == nil {
		debugCb = func(msg string) {}
	}
	return func(c *gin.Context) {
		rqId := c.Request.Header.Get("X-Request-ID")
		rqType := c.Request.Header.Get("X-Request-Type")
		logPrefix := "crypt-middleware[" + rqType + "." + rqId + "]: "
		appId := c.GetHeader(manager.AppIdHeaderKey())
		userId := c.GetHeader(manager.UserIdHeaderKey())
		iv := c.GetHeader(manager.UserIvHeaderKey())
		body, _ := io.ReadAll(c.Request.Body)
		debugCb(logPrefix + "body in=" + string(body))
		// 解密
		decrypted, err := manager.Provider().Decrypt(appId, userId, []byte(iv), body)
		if err != nil {
			debugCb(logPrefix + "decrypt failed, err=" + err.Error())
			c.Abort()
			errHandle(
				apierr.ToStatusError(apierr.NewBadRequestError(apierr.CryptMidDecFailed, err).WithRequestTypeAndId(rqType, rqId)),
				marshaler.GetMarshaler(c.GetHeader("Accept")),
				c.Writer,
			)
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(decrypted))

		bdWriter := &BodyWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = bdWriter
		c.Next()
		// 加密
		encrypted, err := manager.Provider().Encrypt(appId, userId, []byte(iv), bdWriter.body.Bytes())
		if err != nil {
			bdWriter.body = bytes.NewBufferString("")
			debugCb(logPrefix + "encrypt failed, err=" + err.Error())
			c.Abort()
			errHandle(
				apierr.ToStatusError(apierr.NewInternalError(apierr.CryptMidEncFailed, err).WithRequestTypeAndId(rqType, rqId)),
				marshaler.GetMarshaler(c.GetHeader("Accept")),
				c.Writer,
			)
			return
		}
		bdWriter.body = bytes.NewBuffer(encrypted)
		debugCb(logPrefix + "body out=" + bdWriter.body.String())
	}
}
