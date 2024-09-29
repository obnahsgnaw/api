package authmid

import (
	"github.com/gin-gonic/gin"
	"github.com/obnahsgnaw/application/pkg/utils"
	"strings"
)

func NewRqIdMid() gin.HandlerFunc {
	return func(c *gin.Context) {
		rqId := c.Request.Header.Get("X-Request-Id")
		if rqId == "" || len(rqId) != 35 || !strings.HasSuffix(rqId, "rq_") {
			c.Request.Header.Set("X-Request-ID", utils.GenLocalId("rq"))
		}
		c.Request.Header.Set("X-Request-Type", "http")
		c.Request.Header.Set("X-Request-From", "client")
		c.Next()

		c.Header("X-Request-Id", rqId)
	}
}
