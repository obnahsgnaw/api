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

func NewCryptMid(manager *crypt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		appId := c.GetHeader("X-App-Id")
		userId := c.GetHeader("X-User-Id")
		iv := c.GetHeader("X-User-Iv")
		body, _ := io.ReadAll(c.Request.Body)
		if manager.Logger() != nil {
			manager.Logger().Debug("Middleware [Crypt]: Body In=" + string(body))
		}
		// 解密
		decrypted, err := manager.Provider().Decrypt(appId, userId, []byte(iv), body)
		if err != nil {
			if manager.Logger() != nil {
				manager.Logger().Debug("Middleware [Crypt ]: decrypt failed, err=" + err.Error())
			}
			c.Abort()
			errhandler.HandlerErr(
				apierr.ToStatusError(apierr.NewBadRequestError(apierr.CryptMidDecFailed, err)),
				marshaler.GetMarshaler(c.GetHeader("Accept")),
				c.Writer,
				nil,
				manager.ErrObjProvider(),
				manager,
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
			if manager.Logger() != nil {
				manager.Logger().Debug("Middleware [Crypt ]: encrypt failed, err=" + err.Error())
			}
			c.Abort()
			errhandler.HandlerErr(
				apierr.ToStatusError(apierr.NewInternalError(apierr.CryptMidEncFailed, err)),
				marshaler.GetMarshaler(c.GetHeader("Accept")),
				c.Writer,
				nil,
				manager.ErrObjProvider(),
				manager,
			)
			return
		}
		bdWriter.body = bytes.NewBuffer(encrypted)
		if manager.Logger() != nil {
			manager.Logger().Debug("Middleware [Crypt]: Body Out=" + bdWriter.body.String())
		}
	}
}
