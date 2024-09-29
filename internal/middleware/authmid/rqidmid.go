package authmid

import (
	"github.com/gin-gonic/gin"
	"github.com/obnahsgnaw/application/pkg/utils"
	"strings"
)

func NewRqIdMid() gin.HandlerFunc {
	return func(c *gin.Context) {
		rqId := c.Request.Header.Get("X-Request-Id")
		if rqId == "" || len(rqId) != 35 || !strings.HasPrefix(rqId, "rq_") {
			rqId = utils.GenLocalId("rq")
			c.Request.Header.Set("X-Request-ID", rqId)
		}
		c.Request.Header.Set("X-Request-Type", "http")
		c.Request.Header.Set("X-Request-From", "client")
		c.Header("X-Request-Id", rqId)
		c.Next()
	}
}
