package cache

import (
	"context"
	"time"

	"github.com/allegro/bigcache/v3"
)

// Cache represents the caching system
type Cache struct {
	cache *bigcache.BigCache
}

// NewCache creates a new cache instance
func NewCache() (*Cache, error) {
	config := bigcache.Config{
		Shards:             1024,
		LifeWindow:         10 * time.Minute,
		CleanWindow:        5 * time.Minute,
		MaxEntriesInWindow: 1000 * 10 * 60,
		MaxEntrySize:       500,
		StatsEnabled:       true,
		Verbose:            true,
		HardMaxCacheSize:   8192,
		Logger:             bigcache.DefaultLogger(),
	}

	cache, err := bigcache.New(context.Background(), config)
	if err != nil {
		return nil, err
	}

	return &Cache{
		cache: cache,
	}, nil
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) ([]byte, error) {
	return c.cache.Get(key)
}

// Set stores a value in the cache
func (c *Cache) Set(key string, value []byte) error {
	return c.cache.Set(key, value)
}

// Delete removes a value from the cache
func (c *Cache) Delete(key string) error {
	return c.cache.Delete(key)
}

// Reset clears the cache
func (c *Cache) Reset() error {
	return c.cache.Reset()
}

// Len returns the number of entries in the cache
func (c *Cache) Len() int {
	return c.cache.Len()
}

// Capacity returns the capacity of the cache
func (c *Cache) Capacity() int {
	return c.cache.Capacity()
}

// Stats returns the cache statistics
func (c *Cache) Stats() *bigcache.Stats {
	return c.cache.Stats()
}
