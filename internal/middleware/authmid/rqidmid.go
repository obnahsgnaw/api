package authmid

import (
	"github.com/gin-gonic/gin"
	"github.com/obnahsgnaw/application/pkg/utils"
)

func NewRqIdMid() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Header.Get("X-Request-ID") == "" {
			c.Request.Header.Set("X-Request-ID", utils.GenLocalId("rq"))
		}

		c.Next()
	}
}
