package main

import (
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.DebugMode)
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard

	port, err := loadEnvVars()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	router := gin.Default()

	initializeRoutes(router)

	fmt.Printf("Starting server on port %s..\n", port)
	if err := router.Run(":" + port); err != nil {
		fmt.Printf("Failed to run server: %v\n", err)
		return
	}
}
