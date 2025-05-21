package network

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/moroni/BYC/internal/logger"
	"go.uber.org/zap"
)

// Connection represents a network connection
type Connection struct {
	conn     net.Conn
	peer     *Peer
	lastSeen time.Time
	isActive bool
	mu       sync.RWMutex
}

// ConnectionPool manages network connections
type ConnectionPool struct {
	connections map[string]*Connection
	maxConns    int
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(maxConns int) *ConnectionPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &ConnectionPool{
		connections: make(map[string]*Connection),
		maxConns:    maxConns,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start begins connection pool management
func (cp *ConnectionPool) Start() {
	go cp.cleanupLoop()
}

// Stop stops connection pool management
func (cp *ConnectionPool) Stop() {
	cp.cancel()
	cp.closeAll()
}

// cleanupLoop periodically cleans up inactive connections
func (cp *ConnectionPool) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-cp.ctx.Done():
			return
		case <-ticker.C:
			cp.cleanup()
		}
	}
}

// cleanup removes inactive connections
func (cp *ConnectionPool) cleanup() {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	now := time.Now()
	for addr, conn := range cp.connections {
		if !conn.isActive || now.Sub(conn.lastSeen) > 5*time.Minute {
			cp.closeConnection(addr)
		}
	}
}

// closeAll closes all connections
func (cp *ConnectionPool) closeAll() {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	for addr := range cp.connections {
		cp.closeConnection(addr)
	}
}

// closeConnection closes a specific connection
func (cp *ConnectionPool) closeConnection(addr string) {
	if conn, exists := cp.connections[addr]; exists {
		conn.mu.Lock()
		conn.isActive = false
		conn.conn.Close()
		conn.mu.Unlock()
		delete(cp.connections, addr)
	}
}

// AddConnection adds a new connection to the pool
func (cp *ConnectionPool) AddConnection(conn net.Conn, peer *Peer) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	// Check if we've reached the maximum number of connections
	if len(cp.connections) >= cp.maxConns {
		return fmt.Errorf("connection pool is full")
	}

	// Create new connection
	connection := &Connection{
		conn:     conn,
		peer:     peer,
		lastSeen: time.Now(),
		isActive: true,
	}

	// Add to pool
	cp.connections[peer.Address] = connection

	logger.Info("Added new connection",
		zap.String("peer", peer.Address),
		zap.Int("total_connections", len(cp.connections)))

	return nil
}

// GetConnection returns a connection for a peer
func (cp *ConnectionPool) GetConnection(addr string) (*Connection, bool) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	conn, exists := cp.connections[addr]
	if !exists || !conn.isActive {
		return nil, false
	}

	// Update last seen time
	conn.mu.Lock()
	conn.lastSeen = time.Now()
	conn.mu.Unlock()

	return conn, true
}

// UpdateConnection updates a connection's status
func (cp *ConnectionPool) UpdateConnection(addr string, isActive bool) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if conn, exists := cp.connections[addr]; exists {
		conn.mu.Lock()
		conn.isActive = isActive
		conn.lastSeen = time.Now()
		conn.mu.Unlock()
	}
}

// GetActiveConnections returns the number of active connections
func (cp *ConnectionPool) GetActiveConnections() int {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	count := 0
	for _, conn := range cp.connections {
		conn.mu.RLock()
		if conn.isActive {
			count++
		}
		conn.mu.RUnlock()
	}
	return count
}

// GetConnectionStats returns connection pool statistics
func (cp *ConnectionPool) GetConnectionStats() map[string]interface{} {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total_connections"] = len(cp.connections)
	stats["active_connections"] = cp.GetActiveConnections()
	stats["max_connections"] = cp.maxConns

	// Get connection details
	connections := make([]map[string]interface{}, 0)
	for addr, conn := range cp.connections {
		conn.mu.RLock()
		connections = append(connections, map[string]interface{}{
			"address":   addr,
			"is_active": conn.isActive,
			"last_seen": conn.lastSeen,
			"peer_info": conn.peer,
		})
		conn.mu.RUnlock()
	}
	stats["connections"] = connections

	return stats
}
