package network

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"

	"byc/internal/logger"

	"go.uber.org/zap"
)

// SecurityManager handles network security
type SecurityManager struct {
	rateLimiters map[string]*RateLimiter
	blacklist    map[string]time.Time
	whitelist    map[string]bool
	mu           sync.RWMutex
}

// NewSecurityManager creates a new security manager
func NewSecurityManager() *SecurityManager {
	return &SecurityManager{
		rateLimiters: make(map[string]*RateLimiter),
		blacklist:    make(map[string]time.Time),
		whitelist:    make(map[string]bool),
	}
}

// AddToBlacklist adds an address to the blacklist
func (sm *SecurityManager) AddToBlacklist(addr string, duration time.Duration) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.blacklist[addr] = time.Now().Add(duration)
	logger.Info("Added address to blacklist",
		zap.String("address", addr),
		zap.Duration("duration", duration))
}

// IsBlacklisted checks if an address is blacklisted
func (sm *SecurityManager) IsBlacklisted(addr string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if expiry, exists := sm.blacklist[addr]; exists {
		if time.Now().Before(expiry) {
			return true
		}
		// Remove expired blacklist entry
		delete(sm.blacklist, addr)
	}
	return false
}

// AddToWhitelist adds an address to the whitelist
func (sm *SecurityManager) AddToWhitelist(addr string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.whitelist[addr] = true
	logger.Info("Added address to whitelist",
		zap.String("address", addr))
}

// IsWhitelisted checks if an address is whitelisted
func (sm *SecurityManager) IsWhitelisted(addr string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.whitelist[addr]
}

// GetRateLimiter gets or creates a rate limiter for an address
func (sm *SecurityManager) GetRateLimiter(addr string) *RateLimiter {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if limiter, exists := sm.rateLimiters[addr]; exists {
		return limiter
	}

	// Create new rate limiter (100 requests per minute)
	limiter := NewRateLimiter(100, 1024*1024) // 1MB/s
	sm.rateLimiters[addr] = limiter
	return limiter
}

// CheckConnection checks if a connection is allowed
func (sm *SecurityManager) CheckConnection(addr string) error {
	// Check whitelist first
	if sm.IsWhitelisted(addr) {
		return nil
	}

	// Check blacklist
	if sm.IsBlacklisted(addr) {
		return fmt.Errorf("address is blacklisted")
	}

	// Check rate limit
	if !sm.GetRateLimiter(addr).AllowInbound() {
		sm.AddToBlacklist(addr, 5*time.Minute)
		return fmt.Errorf("rate limit exceeded")
	}

	return nil
}

// ConfigureTLS configures TLS for a connection
func ConfigureTLS(conn net.Conn, config *tls.Config) (net.Conn, error) {
	if config == nil {
		return conn, nil
	}

	tlsConn := tls.Server(conn, config)
	if err := tlsConn.Handshake(); err != nil {
		return nil, fmt.Errorf("TLS handshake failed: %v", err)
	}

	return tlsConn, nil
}

// GetSecurityStats returns security statistics
func (sm *SecurityManager) GetSecurityStats() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["blacklisted"] = len(sm.blacklist)
	stats["whitelisted"] = len(sm.whitelist)
	stats["rate_limited"] = len(sm.rateLimiters)

	// Get blacklist details
	blacklist := make([]map[string]interface{}, 0)
	for addr, expiry := range sm.blacklist {
		blacklist = append(blacklist, map[string]interface{}{
			"address": addr,
			"expiry":  expiry,
		})
	}
	stats["blacklist"] = blacklist

	return stats
}

// CleanupExpiredEntries removes expired blacklist entries
func (sm *SecurityManager) CleanupExpiredEntries() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	for addr, expiry := range sm.blacklist {
		if now.After(expiry) {
			delete(sm.blacklist, addr)
			logger.Info("Removed expired blacklist entry",
				zap.String("address", addr))
		}
	}
}
