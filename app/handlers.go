package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func handleImageFetch(c *gin.Context) {
	encryptedURL := c.Query("hash")
	if encryptedURL == "" {
		c.Header("X-Error", "Missing hash parameter")
		c.Header("Content-Type", fallbackContentType)
		c.Header("Cache-Control", "public, max-age=604800")
		c.Data(http.StatusOK, fallbackContentType, fallbackImageData)
		return
	}

	if cacheItem, found := getCachedImage(encryptedURL); found {
		c.Header("X-Cache", "Hit")
		c.Header("Content-Type", cacheItem.ContentType)
		c.Header("Cache-Control", "public, max-age=604800")
		c.Data(http.StatusOK, cacheItem.ContentType, cacheItem.ImageData)
		return
	}

	imageURL := decrypt(encryptedURL)
	if imageURL == "" {
		c.Header("X-Error", "Invalid hash parameter")
		c.Header("Content-Type", fallbackContentType)
		c.Header("Cache-Control", "public, max-age=604800")
		c.Data(http.StatusOK, fallbackContentType, fallbackImageData)
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

	c.Header("X-Cache", "Miss")
	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=604800")
	c.Data(http.StatusOK, contentType, imageData)
}

func handleGenerateFetch(c *gin.Context) {
	encryptedURL := c.Query("hash")
	if encryptedURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "error": "Missing hash parameter."})
		return
	}

	cacheItem, found := getCachedImage(encryptedURL)
	if found {
		c.Header("X-Cache", "Hit")
		c.Header("Content-Type", "image/png")
		c.Header("Cache-Control", "public, max-age=604800")
		c.Data(http.StatusOK, "image/png", cacheItem.ImageData)
		return
	}

	decryptedURL := decrypt(encryptedURL)
	if decryptedURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "error": "Invalid hash parameter."})
		return
	}

	cleanedURL := removeControlCharacters(decryptedURL)
	data := Data{}

	if err := json.Unmarshal([]byte(cleanedURL), &data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "error": "Malformed hash parameter."})
		return
	}

	validate := validator.New()

	if err := validate.Struct(data); err != nil {
		var validationErrors []string
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, fmt.Sprintf("Field '%s' is required.", err.Field()))
		}

		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "error": validationErrors})
		return
	}

	size := c.DefaultQuery("size", "1024")
	width, err := strconv.Atoi(size)
	if err != nil {
		width = 1024
	}

	allowedSizes := []int{256, 512, 1024, 2048, 4096}
	if !contains(allowedSizes, width) {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "error": "Invalid size parameter."})
		return
	}

	imageData := handleCard(c, data, width)
	if imageData == nil {
		return
	}

	cacheImage(encryptedURL, imageData, "image/webp")

	c.Header("X-Cache", "Miss")
	c.Header("Content-Type", "image/webp")
	c.Header("Cache-Control", "public, max-age=604800")
	c.Data(http.StatusOK, "image/webp", imageData)
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
