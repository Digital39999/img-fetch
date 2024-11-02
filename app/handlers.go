package main

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

func handleImageFetch(c *gin.Context) {
	encryptedURL := c.Query("hash")
	if encryptedURL == "" {
		c.Header("Cache-Control", "public, max-age=604800")
		c.Data(http.StatusOK, fallbackContentType, fallbackImageData)
		return
	}

	if cacheItem, found := getCachedImage(encryptedURL); found {
		c.Header("Cache-Control", "public, max-age=604800")
		c.Data(http.StatusOK, cacheItem.ContentType, cacheItem.ImageData)
		return
	}

	imageURL := decrypt(encryptedURL)
	if imageURL == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "error": "Invalid or corrupted encrypted URL."})
		return
	}

	cleanedURL := removeControlCharacters(imageURL)

	imageData, contentType, err := fetchImage(cleanedURL)
	if err != nil {
		imageData = fallbackImageData
		contentType = fallbackContentType
	} else {
		cacheImage(encryptedURL, imageData, contentType)
	}

	c.Header("Cache-Control", "public, max-age=604800")
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
