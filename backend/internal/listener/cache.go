package listener

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// --- CACHE DEFINITIONS ---
const cacheTTL = 5 * time.Minute
const cleanupInterval = 10 * time.Minute // How often the sweeper runs

type cacheItem struct {
	data      any
	expiresAt time.Time
}

// Shared globally across the entire `listener` package
var deviceNameCache sync.Map
var deviceGroupNameCache sync.Map

// --- CACHE HELPERS ---

// getFromCache checks for a valid, non-expired cache entry (Lazy Eviction)
func getFromCache(cache *sync.Map, key string) (any, bool) {
	if item, ok := cache.Load(key); ok {
		cached := item.(cacheItem)
		if time.Now().Before(cached.expiresAt) {
			return cached.data, true // Cache HIT and still valid
		}
		// Cache EXPIRED: delete it
		cache.Delete(key)
	}
	return nil, false // Cache MISS or EXPIRED
}

// setToCache stores an entry with an expiration time
func setToCache(cache *sync.Map, key string, data any, duration time.Duration) {
	cache.Store(key, cacheItem{
		data:      data,
		expiresAt: time.Now().Add(duration),
	})
}

// --- ACTIVE MEMORY MANAGEMENT ---

// StartCacheSweeper runs in the background to prevent memory leaks from abandoned keys
func StartCacheSweeper(ctx context.Context) {
	ticker := time.NewTicker(cleanupInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				cleanMap(&deviceNameCache, "deviceNameCache")
				cleanMap(&deviceGroupNameCache, "deviceGroupNameCache")
			case <-ctx.Done():
				ticker.Stop()
				slog.Info("Shutting down cache listener sweeper...")
				return
			}
		}
	}()
}

// cleanMap loops through all items in a sync.Map and deletes expired ones
func cleanMap(cache *sync.Map, cacheName string) {
	now := time.Now()
	expiredCount := 0

	cache.Range(func(key, value any) bool {
		cached := value.(cacheItem)
		if now.After(cached.expiresAt) {
			cache.Delete(key)
			expiredCount++
		}
		return true // Return true to continue iterating to the next item
	})

	if expiredCount > 0 {
		slog.Debug("Cache swept", slog.String("cache", cacheName), slog.Int("cleared", expiredCount))
	}
}
