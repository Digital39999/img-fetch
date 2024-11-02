package main

import (
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

func handleImageFetch(c *gin.Context) {
	encodedURL := c.Query("hash")
	if encodedURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "error": "Missing image hash."})
		return
	}

	urlStore.RLock()
	imageURL, found := urlStore.store[encodedURL]
	urlStore.RUnlock()

	if !found {
		var err error
		imageURL, err = decodeURL(encodedURL)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "error": "Invalid image hash."})
			return
		}

		urlStore.Lock()
		urlStore.store[encodedURL] = imageURL
		urlStore.Unlock()
	}

	imageData, contentType, err := fetchImage(imageURL)
	if err != nil {
		fallbackURL := os.Getenv("FALLBACK_IMAGE_URL")
		if fallbackURL == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "error": "Failed to fetch image."})
			return
		}

		imageData, contentType, err = fetchImage(fallbackURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "error": "Failed to fetch fallback image."})
			return
		}
	}

	c.Data(http.StatusOK, contentType, imageData)
}

func statsHandler(c *gin.Context) {
	memoryBytes := getMemoryUsage()

	stats := gin.H{
		"cpu_usage": getCpuUsage(),
		"ram_usage": formatBytes(memoryBytes),

		"ram_usage_bytes": memoryBytes,

		"system_uptime": time.Since(startTime).String(),
		"go_routines":   runtime.NumGoroutine(),
	}

	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "data": stats})
}

func infoHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "data": "Image service is running."})
}

func notFoundHandler(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "error": "Route not found."})
}
