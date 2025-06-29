package main

import (
	// "net/http"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./static")

	SetupRoutes(router)

	router.Run(":8080")

}
