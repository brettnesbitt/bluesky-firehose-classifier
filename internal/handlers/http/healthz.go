package httphandlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterHealthzRoutes(router *gin.Engine) {
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "OK",
		})
	})
}
