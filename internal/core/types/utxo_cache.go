package types

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

const (
	// DefaultCacheSize is the default maximum number of items in the cache
	DefaultCacheSize = 10000
	// DefaultCacheTTL is the default time-to-live for cache items
	DefaultCacheTTL = 5 * time.Minute
	// DefaultCleanupInterval is the default interval for cache cleanup
	DefaultCleanupInterval = 1 * time.Minute
)

// CacheItem represents an item in the cache
type CacheItem struct {
	// Key is the cache key
	Key string
	// Value is the cached value
	Value interface{}
	// Size is the size of the cached value in bytes
	Size int64
	// Created is the time when the item was created
	Created time.Time
	// LastAccessed is the time when the item was last accessed
	LastAccessed time.Time
	// Expires is the time when the item expires
	Expires time.Time
}

// UTXOCache handles caching of the UTXO set
type UTXOCache struct {
	utxoSet *UTXOSet
	mu      sync.RWMutex

	// Cache state
	maxSize     int
	ttl         time.Duration
	cleanup     time.Duration
	items       map[string]*list.Element
	queue       *list.List
	totalSize   int64
	stopCleanup chan struct{}
}

// NewUTXOCache creates a new UTXO cache handler
func NewUTXOCache(utxoSet *UTXOSet) *UTXOCache {
	cache := &UTXOCache{
		utxoSet:     utxoSet,
		maxSize:     DefaultCacheSize,
		ttl:         DefaultCacheTTL,
		cleanup:     DefaultCleanupInterval,
		items:       make(map[string]*list.Element),
		queue:       list.New(),
		stopCleanup: make(chan struct{}),
	}

	// Start cleanup goroutine
	go cache.cleanupLoop()

	return cache
}

// Get retrieves a value from the cache
func (uc *UTXOCache) Get(key string) (interface{}, bool) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	// Check if item exists
	element, exists := uc.items[key]
	if !exists {
		return nil, false
	}

	// Get item
	item := element.Value.(*CacheItem)

	// Check if item has expired
	if time.Now().After(item.Expires) {
		uc.mu.RUnlock()
		uc.mu.Lock()
		uc.removeItem(element)
		uc.mu.Unlock()
		uc.mu.RLock()
		return nil, false
	}

	// Update last accessed time
	item.LastAccessed = time.Now()

	// Move item to front of queue
	uc.queue.MoveToFront(element)

	return item.Value, true
}

// Set adds a value to the cache
func (uc *UTXOCache) Set(key string, value interface{}, size int64) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	// Check if item already exists
	if element, exists := uc.items[key]; exists {
		// Update existing item
		item := element.Value.(*CacheItem)
		item.Value = value
		item.Size = size
		item.LastAccessed = time.Now()
		item.Expires = time.Now().Add(uc.ttl)

		// Update total size
		uc.totalSize = uc.totalSize - item.Size + size

		// Move item to front of queue
		uc.queue.MoveToFront(element)

		return nil
	}

	// Create new item
	item := &CacheItem{
		Key:          key,
		Value:        value,
		Size:         size,
		Created:      time.Now(),
		LastAccessed: time.Now(),
		Expires:      time.Now().Add(uc.ttl),
	}

	// Check if cache is full
	if uc.queue.Len() >= uc.maxSize {
		// Remove least recently used item
		uc.removeItem(uc.queue.Back())
	}

	// Add item to cache
	element := uc.queue.PushFront(item)
	uc.items[key] = element
	uc.totalSize += size

	return nil
}

// Remove removes a value from the cache
func (uc *UTXOCache) Remove(key string) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	// Check if item exists
	if element, exists := uc.items[key]; exists {
		uc.removeItem(element)
	}
}

// Clear removes all values from the cache
func (uc *UTXOCache) Clear() {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	// Clear items
	uc.items = make(map[string]*list.Element)
	uc.queue.Init()
	uc.totalSize = 0
}

// GetCacheStats returns statistics about the cache
func (uc *UTXOCache) GetCacheStats() *CacheStats {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	stats := &CacheStats{
		MaxSize:     uc.maxSize,
		ItemCount:   uc.queue.Len(),
		TotalSize:   uc.totalSize,
		TTL:         uc.ttl,
		Cleanup:     uc.cleanup,
		HitCount:    0,
		MissCount:   0,
		EvictCount:  0,
		ExpireCount: 0,
	}

	// Calculate hit/miss ratio
	total := stats.HitCount + stats.MissCount
	if total > 0 {
		stats.HitRatio = float64(stats.HitCount) / float64(total)
	}

	return stats
}

// SetMaxSize sets the maximum number of items in the cache
func (uc *UTXOCache) SetMaxSize(size int) {
	uc.mu.Lock()
	uc.maxSize = size
	uc.mu.Unlock()
}

// SetTTL sets the time-to-live for cache items
func (uc *UTXOCache) SetTTL(ttl time.Duration) {
	uc.mu.Lock()
	uc.ttl = ttl
	uc.mu.Unlock()
}

// SetCleanupInterval sets the interval for cache cleanup
func (uc *UTXOCache) SetCleanupInterval(interval time.Duration) {
	uc.mu.Lock()
	uc.cleanup = interval
	uc.mu.Unlock()
}

// StopCleanup stops the cleanup goroutine
func (uc *UTXOCache) StopCleanup() {
	close(uc.stopCleanup)
}

// removeItem removes an item from the cache
func (uc *UTXOCache) removeItem(element *list.Element) {
	// Get item
	item := element.Value.(*CacheItem)

	// Remove item from map
	delete(uc.items, item.Key)

	// Remove item from queue
	uc.queue.Remove(element)

	// Update total size
	uc.totalSize -= item.Size
}

// cleanupLoop periodically removes expired items from the cache
func (uc *UTXOCache) cleanupLoop() {
	ticker := time.NewTicker(uc.cleanup)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			uc.cleanupExpired()
		case <-uc.stopCleanup:
			return
		}
	}
}

// cleanupExpired removes expired items from the cache
func (uc *UTXOCache) cleanupExpired() {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	now := time.Now()
	element := uc.queue.Back()

	for element != nil {
		item := element.Value.(*CacheItem)
		next := element.Prev()

		if now.After(item.Expires) {
			uc.removeItem(element)
		}

		element = next
	}
}

// CacheStats holds statistics about the cache
type CacheStats struct {
	// MaxSize is the maximum number of items in the cache
	MaxSize int
	// ItemCount is the current number of items in the cache
	ItemCount int
	// TotalSize is the total size of all items in the cache in bytes
	TotalSize int64
	// TTL is the time-to-live for cache items
	TTL time.Duration
	// Cleanup is the interval for cache cleanup
	Cleanup time.Duration
	// HitCount is the number of cache hits
	HitCount int64
	// MissCount is the number of cache misses
	MissCount int64
	// EvictCount is the number of items evicted from the cache
	EvictCount int64
	// ExpireCount is the number of items that expired
	ExpireCount int64
	// HitRatio is the ratio of cache hits to total requests
	HitRatio float64
}

// String returns a string representation of the cache statistics
func (cs *CacheStats) String() string {
	return fmt.Sprintf(
		"Max Size: %d, Items: %d\n"+
			"Total Size: %d bytes\n"+
			"TTL: %v, Cleanup: %v\n"+
			"Hits: %d, Misses: %d\n"+
			"Evictions: %d, Expirations: %d\n"+
			"Hit Ratio: %.2f",
		cs.MaxSize, cs.ItemCount,
		cs.TotalSize,
		cs.TTL, cs.Cleanup,
		cs.HitCount, cs.MissCount,
		cs.EvictCount, cs.ExpireCount,
		cs.HitRatio,
	)
}
