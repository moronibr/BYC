package network

import (
	"net/http"
	"sync"
	"time"
)

// TokenBucket implements a token bucket rate limiter
type TokenBucket struct {
	rate       float64
	burst      int
	tokens     float64
	lastUpdate time.Time
	mu         sync.Mutex
}

// NewTokenBucket creates a new token bucket rate limiter
func NewTokenBucket(rate float64, burst int) *TokenBucket {
	return &TokenBucket{
		rate:       rate,
		burst:      burst,
		tokens:     float64(burst),
		lastUpdate: time.Now(),
	}
}

// Allow checks if a request is allowed under the rate limit
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastUpdate).Seconds()
	tb.tokens = min(float64(tb.burst), tb.tokens+elapsed*tb.rate)
	tb.lastUpdate = now

	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}
	return false
}

// RateLimitMiddleware wraps an HTTP handler with rate limiting
type RateLimitMiddleware struct {
	limiter *TokenBucket
	next    http.Handler
}

// NewRateLimitMiddleware creates a new rate limit middleware
func NewRateLimitMiddleware(rate float64, burst int, next http.Handler) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter: NewTokenBucket(rate, burst),
		next:    next,
	}
}

// ServeHTTP implements the http.Handler interface
func (m *RateLimitMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !m.limiter.Allow() {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("Rate limit exceeded"))
		return
	}
	m.next.ServeHTTP(w, r)
}

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	GlobalRate     float64
	GlobalBurst    int
	PerIPRate      float64
	PerIPBurst     int
	EndpointRates  map[string]float64
	EndpointBursts map[string]int
}

// RateLimitManager manages rate limiters for different endpoints
type RateLimitManager struct {
	global           *TokenBucket
	ipLimiters       map[string]*TokenBucket
	endpointLimiters map[string]*TokenBucket
	config           RateLimitConfig
	mu               sync.RWMutex
}

// NewRateLimitManager creates a new rate limit manager
func NewRateLimitManager(config RateLimitConfig) *RateLimitManager {
	manager := &RateLimitManager{
		global:           NewTokenBucket(config.GlobalRate, config.GlobalBurst),
		ipLimiters:       make(map[string]*TokenBucket),
		endpointLimiters: make(map[string]*TokenBucket),
		config:           config,
	}

	// Initialize endpoint rate limiters
	for endpoint, rate := range config.EndpointRates {
		burst := config.EndpointBursts[endpoint]
		manager.endpointLimiters[endpoint] = NewTokenBucket(rate, burst)
	}

	return manager
}

// GetIPLimiter gets or creates a rate limiter for an IP address
func (m *RateLimitManager) GetIPLimiter(ip string) *TokenBucket {
	m.mu.Lock()
	defer m.mu.Unlock()

	limiter, exists := m.ipLimiters[ip]
	if !exists {
		limiter = NewTokenBucket(m.config.PerIPRate, m.config.PerIPBurst)
		m.ipLimiters[ip] = limiter
	}
	return limiter
}

// GetEndpointLimiter gets a rate limiter for an endpoint
func (m *RateLimitManager) GetEndpointLimiter(endpoint string) *TokenBucket {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.endpointLimiters[endpoint]
}

// AllowRequest checks if a request is allowed under all applicable rate limits
func (m *RateLimitManager) AllowRequest(ip, endpoint string) bool {
	// Check global rate limit
	if !m.global.Allow() {
		return false
	}

	// Check per-IP rate limit
	if !m.GetIPLimiter(ip).Allow() {
		return false
	}

	// Check per-endpoint rate limit
	if limiter := m.GetEndpointLimiter(endpoint); limiter != nil {
		if !limiter.Allow() {
			return false
		}
	}

	return true
}

// Cleanup removes old IP rate limiters
func (m *RateLimitManager) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Remove IP limiters that haven't been used in the last hour
	for ip, limiter := range m.ipLimiters {
		if time.Since(limiter.lastUpdate) > time.Hour {
			delete(m.ipLimiters, ip)
		}
	}
}
