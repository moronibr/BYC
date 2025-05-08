package security

import (
	"fmt"
	"net"
	"sync"

	"golang.org/x/time/rate"
)

// PeerLimiter represents a rate limiter for peers
type PeerLimiter struct {
	mu sync.RWMutex

	// IP-based limits
	ipLimiters map[string]*rate.Limiter
	ipRate     rate.Limit
	ipBurst    int

	// Connection-based limits
	connLimiters map[string]*rate.Limiter
	connRate     rate.Limit
	connBurst    int

	// Global limits
	maxConnectionsPerIP int
	connectionsPerIP    map[string]int
}

// NewPeerLimiter creates a new peer limiter
func NewPeerLimiter(ipRate, connRate rate.Limit, ipBurst, connBurst, maxConnPerIP int) *PeerLimiter {
	return &PeerLimiter{
		ipLimiters:          make(map[string]*rate.Limiter),
		ipRate:              ipRate,
		ipBurst:             ipBurst,
		connLimiters:        make(map[string]*rate.Limiter),
		connRate:            connRate,
		connBurst:           connBurst,
		maxConnectionsPerIP: maxConnPerIP,
		connectionsPerIP:    make(map[string]int),
	}
}

// AllowConnection checks if a new connection is allowed
func (pl *PeerLimiter) AllowConnection(addr net.Addr) error {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	// Get IP address
	ip, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		return fmt.Errorf("invalid address: %v", err)
	}

	// Check connection limit per IP
	if pl.connectionsPerIP[ip] >= pl.maxConnectionsPerIP {
		return fmt.Errorf("too many connections from IP %s", ip)
	}

	// Get or create IP limiter
	ipLimiter, exists := pl.ipLimiters[ip]
	if !exists {
		ipLimiter = rate.NewLimiter(pl.ipRate, pl.ipBurst)
		pl.ipLimiters[ip] = ipLimiter
	}

	// Check IP rate limit
	if !ipLimiter.Allow() {
		return fmt.Errorf("rate limit exceeded for IP %s", ip)
	}

	// Get or create connection limiter
	connLimiter, exists := pl.connLimiters[addr.String()]
	if !exists {
		connLimiter = rate.NewLimiter(pl.connRate, pl.connBurst)
		pl.connLimiters[addr.String()] = connLimiter
	}

	// Check connection rate limit
	if !connLimiter.Allow() {
		return fmt.Errorf("rate limit exceeded for connection %s", addr.String())
	}

	// Increment connection count
	pl.connectionsPerIP[ip]++

	return nil
}

// RemoveConnection removes a connection
func (pl *PeerLimiter) RemoveConnection(addr net.Addr) {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	// Get IP address
	ip, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		return
	}

	// Decrement connection count
	if pl.connectionsPerIP[ip] > 0 {
		pl.connectionsPerIP[ip]--
	}

	// Remove connection limiter
	delete(pl.connLimiters, addr.String())
}

// Cleanup removes expired limiters
func (pl *PeerLimiter) Cleanup() {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	// Remove IPs with no connections
	for ip, count := range pl.connectionsPerIP {
		if count == 0 {
			delete(pl.connectionsPerIP, ip)
			delete(pl.ipLimiters, ip)
		}
	}
}

// GetConnectionCount gets the number of connections for an IP
func (pl *PeerLimiter) GetConnectionCount(ip string) int {
	pl.mu.RLock()
	defer pl.mu.RUnlock()
	return pl.connectionsPerIP[ip]
}

// GetTotalConnections gets the total number of connections
func (pl *PeerLimiter) GetTotalConnections() int {
	pl.mu.RLock()
	defer pl.mu.RUnlock()
	total := 0
	for _, count := range pl.connectionsPerIP {
		total += count
	}
	return total
}

// GetIPLimiters gets the number of IP limiters
func (pl *PeerLimiter) GetIPLimiters() int {
	pl.mu.RLock()
	defer pl.mu.RUnlock()
	return len(pl.ipLimiters)
}

// GetConnectionLimiters gets the number of connection limiters
func (pl *PeerLimiter) GetConnectionLimiters() int {
	pl.mu.RLock()
	defer pl.mu.RUnlock()
	return len(pl.connLimiters)
}
