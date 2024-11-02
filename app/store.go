package main

import (
	"sync"
	"time"
)

type CacheItem struct {
	ImageData    []byte
	ContentType  string
	LastAccessed time.Time
	Size         int
}

var (
	fallbackImageData   []byte
	fallbackContentType string
	once                sync.Once
	urlCache            sync.Map
	currentCacheSize    int
	maxCacheSize        int
	cacheMutex          sync.Mutex
)

func getCachedImage(key string) (*CacheItem, bool) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if cacheItem, found := urlCache.Load(key); found {
		item := cacheItem.(*CacheItem)
		item.LastAccessed = time.Now()
		return item, true
	}

	return nil, false
}

func cacheImage(key string, data []byte, contentType string) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	imageSize := len(data)
	if currentCacheSize+imageSize > maxCacheSize {
		evictCacheItems(imageSize)
	}

	cacheItem := &CacheItem{
		ImageData:    data,
		ContentType:  contentType,
		LastAccessed: time.Now(),
		Size:         imageSize,
	}

	urlCache.Store(key, cacheItem)
	currentCacheSize += imageSize
}

func evictCacheItems(requiredSpace int) {
	var oldestKey string
	var oldestTime time.Time

	urlCache.Range(func(key, value interface{}) bool {
		item := value.(*CacheItem)
		if oldestKey == "" || item.LastAccessed.Before(oldestTime) {
			oldestKey = key.(string)
			oldestTime = item.LastAccessed
		}

		return true
	})

	if oldestKey != "" {
		if item, ok := urlCache.Load(oldestKey); ok {
			cacheItem := item.(*CacheItem)

			urlCache.Delete(oldestKey)
			currentCacheSize -= cacheItem.Size

			if currentCacheSize+requiredSpace > maxCacheSize {
				evictCacheItems(requiredSpace)
			}
		}
	}
}
