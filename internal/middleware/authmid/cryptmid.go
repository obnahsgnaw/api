package authmid

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/obnahsgnaw/api/internal/errhandler"
	"github.com/obnahsgnaw/api/internal/marshaler"
	"github.com/obnahsgnaw/api/pkg/apierr"
	"github.com/obnahsgnaw/api/service/crypt"
	"io"
)

type BodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w BodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func NewCryptMid(manager *crypt.Manager, debugCb func(msg string)) gin.HandlerFunc {
	if debugCb == nil {
		debugCb = func(msg string) {}
	}
	return func(c *gin.Context) {
		appId := c.GetHeader(manager.AppIdHeaderKey())
		userId := c.GetHeader(manager.UserIdHeaderKey())
		iv := c.GetHeader(manager.UserIvHeaderKey())
		body, _ := io.ReadAll(c.Request.Body)
		debugCb("crypt-middleware: body in=" + string(body))
		// 解密
		decrypted, err := manager.Provider().Decrypt(appId, userId, []byte(iv), body)
		if err != nil {
			debugCb("crypt-middleware: decrypt failed, err=" + err.Error())
			c.Abort()
			errhandler.DefaultErrorHandler(
				apierr.ToStatusError(apierr.NewBadRequestError(apierr.CryptMidDecFailed, err)),
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
			debugCb("crypt-middleware: encrypt failed, err=" + err.Error())
			c.Abort()
			errhandler.DefaultErrorHandler(
				apierr.ToStatusError(apierr.NewInternalError(apierr.CryptMidEncFailed, err)),
				marshaler.GetMarshaler(c.GetHeader("Accept")),
				c.Writer,
			)
			return
		}
		bdWriter.body = bytes.NewBuffer(encrypted)
		debugCb("crypt-middleware: body out=" + bdWriter.body.String())
	}
}
