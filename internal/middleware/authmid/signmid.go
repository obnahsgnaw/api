package authmid

import (
	"github.com/gin-gonic/gin"
	"github.com/obnahsgnaw/api/service/sign"
)

func NewSignMid(manager *sign.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		//appId := c.GetHeader("X-App-Id")
		//userId := c.GetHeader("X-User-Id")
		//sign := c.GetHeader("X-User-Signature")
		//method := c.Request.Method
		//uri := c.Request.URL.Path
		// TODO

		if manager.Logger() != nil {
			manager.Logger().Debug("Middleware [Sign ]: sign in=TODO")
		}
		// sign check TODO
		//c.AbortWithStatus(http.StatusUnauthorized)
		c.Next()
		// sign generate
		if manager.Logger() != nil {
			manager.Logger().Debug("Middleware [Sign ]: sign out=TODO")
		}
	}
}
