package security

import (
	"sync"
	"time"
)

const (
	// Default rate limits
	DefaultRateLimitPerMinute = 100
	DefaultRateLimitPerHour   = 1000
	DefaultBurstSize          = 50
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	mu            sync.Mutex
	rate          float64
	burst         int
	tokens        float64
	lastRefill    time.Time
	ipLimits      map[string]*IPLimit
	cleanupTicker *time.Ticker
	done          chan bool
}

// IPLimit tracks rate limits for a specific IP
type IPLimit struct {
	MinuteTokens float64
	HourTokens   float64
	LastMinute   time.Time
	LastHour     time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		rate:          rate,
		burst:         burst,
		tokens:        float64(burst),
		lastRefill:    time.Now(),
		ipLimits:      make(map[string]*IPLimit),
		cleanupTicker: time.NewTicker(1 * time.Hour),
		done:          make(chan bool),
	}

	go rl.cleanupLoop()
	return rl
}

// cleanupLoop periodically removes old IP entries
func (rl *RateLimiter) cleanupLoop() {
	for {
		select {
		case <-rl.cleanupTicker.C:
			rl.cleanup()
		case <-rl.done:
			rl.cleanupTicker.Stop()
			return
		}
	}
}

// cleanup removes IP entries older than 24 hours
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, limit := range rl.ipLimits {
		if now.Sub(limit.LastHour) > 24*time.Hour {
			delete(rl.ipLimits, ip)
		}
	}
}

// Allow checks if a request from an IP is allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	limit, exists := rl.ipLimits[ip]
	if !exists {
		limit = &IPLimit{
			MinuteTokens: float64(DefaultBurstSize),
			HourTokens:   float64(DefaultRateLimitPerHour),
			LastMinute:   now,
			LastHour:     now,
		}
		rl.ipLimits[ip] = limit
	}

	// Refill minute tokens
	elapsed := now.Sub(limit.LastMinute).Seconds()
	limit.MinuteTokens = min(float64(DefaultBurstSize),
		limit.MinuteTokens+elapsed*float64(DefaultRateLimitPerMinute)/60)
	limit.LastMinute = now

	// Refill hour tokens
	elapsed = now.Sub(limit.LastHour).Seconds()
	limit.HourTokens = min(float64(DefaultRateLimitPerHour),
		limit.HourTokens+elapsed*float64(DefaultRateLimitPerHour)/3600)
	limit.LastHour = now

	// Check if we have enough tokens
	if limit.MinuteTokens < 1 || limit.HourTokens < 1 {
		return false
	}

	// Consume tokens
	limit.MinuteTokens--
	limit.HourTokens--

	return true
}

// GetRemainingTokens returns the remaining tokens for an IP
func (rl *RateLimiter) GetRemainingTokens(ip string) (minuteTokens, hourTokens float64) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limit, exists := rl.ipLimits[ip]
	if !exists {
		return float64(DefaultBurstSize), float64(DefaultRateLimitPerHour)
	}

	return limit.MinuteTokens, limit.HourTokens
}

// Stop stops the rate limiter
func (rl *RateLimiter) Stop() {
	close(rl.done)
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
