package security

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/time/rate"
)

// RateLimiter manages rate limiting for RPC endpoints
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
	}
}

// AddEndpoint adds a new endpoint with rate limiting
func (rl *RateLimiter) AddEndpoint(endpoint string, requestsPerSecond float64, burst int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.limiters[endpoint] = rate.NewLimiter(rate.Limit(requestsPerSecond), burst)
}

// Allow checks if a request is allowed for the given endpoint
func (rl *RateLimiter) Allow(endpoint string) bool {
	rl.mu.RLock()
	limiter, exists := rl.limiters[endpoint]
	rl.mu.RUnlock()

	if !exists {
		return true // No rate limit for unknown endpoints
	}

	return limiter.Allow()
}

// Wait waits for rate limit to allow the request
func (rl *RateLimiter) Wait(ctx context.Context, endpoint string) error {
	rl.mu.RLock()
	limiter, exists := rl.limiters[endpoint]
	rl.mu.RUnlock()

	if !exists {
		return nil // No rate limit for unknown endpoints
	}

	return limiter.Wait(ctx)
}

// GetLimiter returns the rate limiter for an endpoint
func (rl *RateLimiter) GetLimiter(endpoint string) *rate.Limiter {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return rl.limiters[endpoint]
}

// UpdateLimits updates the rate limits for an endpoint
func (rl *RateLimiter) UpdateLimits(endpoint string, requestsPerSecond float64, burst int) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if limiter, exists := rl.limiters[endpoint]; exists {
		limiter.SetLimit(rate.Limit(requestsPerSecond))
		// Create a new limiter with updated burst
		rl.limiters[endpoint] = rate.NewLimiter(rate.Limit(requestsPerSecond), burst)
		return nil
	}

	return fmt.Errorf("endpoint %s not found", endpoint)
}

// RemoveEndpoint removes rate limiting for an endpoint
func (rl *RateLimiter) RemoveEndpoint(endpoint string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	delete(rl.limiters, endpoint)
}
