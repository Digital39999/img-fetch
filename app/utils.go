package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/shirou/gopsutil/cpu"
)

var startTime = time.Now()

func loadEnvVars() (string, error) {
	_ = godotenv.Load()

	port := os.Getenv("PORT")

	if port == "" {
		return "", errors.New("missing PORT in environment variables")
	}

	if os.Getenv("FALLBACK_IMAGE_URL") == "" {
		return "", errors.New("missing FALLBACK_IMAGE_URL in environment variables")
	}

	if os.Getenv("SECRET_KEY") == "" {
		return "", errors.New("missing SECRET_KEY in environment variables")
	}

	if os.Getenv("MAX_CACHE_SIZE_MB") == "" {
		return "", errors.New("missing MAX_CACHE_SIZE_MB in environment variables")
	}

	return port, nil
}

func initCacheSettings() {
	maxSizeStr := os.Getenv("MAX_CACHE_SIZE_MB")
	maxCacheSizeMB, _ := strconv.Atoi(maxSizeStr)
	maxCacheSize = maxCacheSizeMB * 1024 * 1024
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
	url = strings.TrimSpace(url)

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

func decrypt(text string) (string, error) {
	encryptedText, err := hex.DecodeString(text)
	if err != nil {
		return "", err
	}

	iv := make([]byte, aes.BlockSize)
	key := validateKey()

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(encryptedText))
	mode.CryptBlocks(decrypted, encryptedText)

	return string(decrypted), nil
}

func validateKey() string {
	key := os.Getenv("SECRET_KEY")

	if len(key) < 32 {
		key += "00000000000000000000000000000000"[:32-len(key)]
	} else if len(key) > 32 {
		key = key[:32]
	}

	return key
}

func initFallbackImage() error {
	once.Do(func() {
		fallbackURL := os.Getenv("FALLBACK_IMAGE_URL")
		if fallbackURL == "" {
			fmt.Println("No FALLBACK_IMAGE_URL set.")
			return
		}

		data, contentType, err := fetchImage(fallbackURL)
		if err != nil {
			fmt.Println("Error fetching fallback image:", err)
			return
		}

		fallbackImageData = data
		fallbackContentType = contentType
	})

	return nil
}

func removeControlCharacters(url string) string {
	re := regexp.MustCompile(`[\x00-\x1F\x7F]`)
	return re.ReplaceAllString(url, "")
}
