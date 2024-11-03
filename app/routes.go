package main

import (
	"github.com/gin-gonic/gin"
)

func initializeRoutes(router *gin.Engine) {
	router.GET("/", infoHandler)
	router.GET("/stats", statsHandler)

	router.GET("/image", handleImageFetch)
	router.GET("/generate", handleGenerateFetch)

	router.NoRoute(notFoundHandler)
}
