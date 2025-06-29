package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {

	router.GET("/", func(context *gin.Context) {
		context.HTML(http.StatusOK, "index.html", nil)
	})

	router.POST("/chat", HandleChatOPENROUTER)

	router.HEAD("/", func(c *gin.Context) { // for pings and stuff
		c.Status(http.StatusOK)
	})
}
