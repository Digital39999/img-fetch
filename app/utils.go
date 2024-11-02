package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/shirou/gopsutil/cpu"
)

var startTime = time.Now()
var urlStore = struct {
	sync.RWMutex
	store map[string]string
}{
	store: make(map[string]string),
}

func loadEnvVars() (string, error) {
	_ = godotenv.Load()

	port := os.Getenv("PORT")

	if port == "" {
		return "", errors.New("missing PORT in environment variables")
	}

	if os.Getenv("FALLBACK_IMAGE_URL") == "" {
		return "", errors.New("missing FALLBACK_IMAGE_URL in environment variables")
	}

	return port, nil
}

func getCpuUsage() float64 {
	percent, err := cpu.Percent(0, false)
	if err != nil {
		return 0
	}

	if len(percent) > 0 {
		return math.Round(percent[0]*100) / 100
	}

	return 0
}

func getMemoryUsage() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	totalAllocated := m.Alloc + m.TotalAlloc
	return totalAllocated
}

func formatBytes(bytes uint64) string {
	const (
		_         = iota
		KB uint64 = 1 << (10 * iota)
		MB
		GB
		TB
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2fTB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2fGB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2fMB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2fKB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

func fetchImage(url string) ([]byte, string, error) {
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("failed to fetch image: %v", err)
	}

	defer resp.Body.Close()

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read image data: %v", err)
	}

	return imageData, resp.Header.Get("Content-Type"), nil
}

func decodeURL(encodedURL string) (string, error) {
	data, err := base64.URLEncoding.DecodeString(encodedURL)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
