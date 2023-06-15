package server

import (
	"github.com/gin-gonic/gin"
	"github.com/obnahsgnaw/api/service"
	"net/http"
)

func DocRoute(url string, contentProvider func() ([]byte, error)) service.RouteProvider {
	return func(engine *gin.Engine) {
		engine.GET(url, func(c *gin.Context) {
			tmpl, err := contentProvider()
			if err != nil {
				c.String(404, err.Error())
			} else {
				c.Header("Content-Type", "application/json; charset=utf-8")
				c.Status(http.StatusOK)
				_, _ = c.Writer.Write(tmpl)
			}
		})
	}
}
