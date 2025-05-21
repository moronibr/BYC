package network

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/byc/internal/blockchain"
)

// DiscoveryConfig holds configuration for peer discovery
type DiscoveryConfig struct {
	BootstrapNodes   []string
	MaxPeers         int
	MinPeers         int
	PingInterval     time.Duration
	PingTimeout      time.Duration
	MaxPingLatency   time.Duration
	MaxConnections   int
	MaxInboundRate   int64
	MaxOutboundRate  int64
	CompressionLevel int
	EnableTLS        bool
	TLSConfig        *tls.Config
}

// PeerInfo represents information about a peer
type PeerInfo struct {
	Address      string
	LastSeen     time.Time
	Latency      time.Duration
	Version      string
	BlockHeight  int64
	HashRate     float64
	StakeWeight  float64
	IsBootstrap  bool
	IsOutbound   bool
	IsCompressed bool
	IsTLS        bool
}

// DiscoveryManager manages peer discovery and network security
type DiscoveryManager struct {
	config       *DiscoveryConfig
	blockchain   *blockchain.Blockchain
	peers        map[string]*PeerInfo
	connections  map[string]net.Conn
	rateLimiters map[string]*RateLimiter
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewDiscoveryConfig creates a new discovery configuration
func NewDiscoveryConfig() *DiscoveryConfig {
	return &DiscoveryConfig{
		BootstrapNodes:   []string{},
		MaxPeers:         50,
		MinPeers:         10,
		PingInterval:     30 * time.Second,
		PingTimeout:      5 * time.Second,
		MaxPingLatency:   1000 * time.Millisecond,
		MaxConnections:   100,
		MaxInboundRate:   1024 * 1024, // 1MB/s
		MaxOutboundRate:  1024 * 1024, // 1MB/s
		CompressionLevel: 6,
		EnableTLS:        true,
		TLSConfig:        &tls.Config{},
	}
}

// NewDiscoveryManager creates a new discovery manager
func NewDiscoveryManager(config *DiscoveryConfig, bc *blockchain.Blockchain) *DiscoveryManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &DiscoveryManager{
		config:       config,
		blockchain:   bc,
		peers:        make(map[string]*PeerInfo),
		connections:  make(map[string]net.Conn),
		rateLimiters: make(map[string]*RateLimiter),
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start starts the discovery manager
func (dm *DiscoveryManager) Start() error {
	// Connect to bootstrap nodes
	for _, addr := range dm.config.BootstrapNodes {
		if err := dm.connectToPeer(addr, true); err != nil {
			fmt.Printf("Failed to connect to bootstrap node %s: %v\n", addr, err)
		}
	}

	// Start peer discovery
	go dm.discoverPeers()
	go dm.monitorPeers()
	go dm.handleConnections()

	return nil
}

// Stop stops the discovery manager
func (dm *DiscoveryManager) Stop() {
	dm.cancel()
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Close all connections
	for _, conn := range dm.connections {
		conn.Close()
	}

	// Clear peers and connections
	dm.peers = make(map[string]*PeerInfo)
	dm.connections = make(map[string]net.Conn)
}

// connectToPeer connects to a peer
func (dm *DiscoveryManager) connectToPeer(addr string, isBootstrap bool) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Check if already connected
	if _, exists := dm.peers[addr]; exists {
		return errors.New("already connected to peer")
	}

	// Check connection limit
	if len(dm.connections) >= dm.config.MaxConnections {
		return errors.New("connection limit reached")
	}

	// Create connection
	var conn net.Conn
	var err error
	if dm.config.EnableTLS {
		conn, err = tls.Dial("tcp", addr, dm.config.TLSConfig)
	} else {
		conn, err = net.Dial("tcp", addr)
	}
	if err != nil {
		return err
	}

	// Create peer info
	peer := &PeerInfo{
		Address:      addr,
		LastSeen:     time.Now(),
		IsBootstrap:  isBootstrap,
		IsOutbound:   true,
		IsCompressed: true,
		IsTLS:        dm.config.EnableTLS,
	}

	// Add peer and connection
	dm.peers[addr] = peer
	dm.connections[addr] = conn

	// Create rate limiter
	dm.rateLimiters[addr] = NewRateLimiter(dm.config.MaxInboundRate, dm.config.MaxOutboundRate)

	return nil
}

// discoverPeers discovers new peers
func (dm *DiscoveryManager) discoverPeers() {
	ticker := time.NewTicker(dm.config.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-dm.ctx.Done():
			return
		case <-ticker.C:
			dm.mu.RLock()
			peerCount := len(dm.peers)
			dm.mu.RUnlock()

			if peerCount < dm.config.MinPeers {
				// Get random peers to ask for new peers
				peers := dm.getRandomPeers(3)
				for _, peer := range peers {
					// Request peer list
					if err := dm.requestPeerList(peer.Address); err != nil {
						fmt.Printf("Failed to request peer list from %s: %v\n", peer.Address, err)
					}
				}
			}
		}
	}
}

// monitorPeers monitors peer health
func (dm *DiscoveryManager) monitorPeers() {
	ticker := time.NewTicker(dm.config.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-dm.ctx.Done():
			return
		case <-ticker.C:
			dm.mu.RLock()
			peers := make([]*PeerInfo, 0, len(dm.peers))
			for _, peer := range dm.peers {
				peers = append(peers, peer)
			}
			dm.mu.RUnlock()

			for _, peer := range peers {
				// Ping peer
				latency, err := dm.pingPeer(peer.Address)
				if err != nil {
					fmt.Printf("Failed to ping peer %s: %v\n", peer.Address, err)
					dm.disconnectPeer(peer.Address)
					continue
				}

				// Update peer info
				dm.mu.Lock()
				peer.LastSeen = time.Now()
				peer.Latency = latency
				dm.mu.Unlock()

				// Check latency
				if latency > dm.config.MaxPingLatency {
					fmt.Printf("Peer %s latency too high: %v\n", peer.Address, latency)
					dm.disconnectPeer(peer.Address)
				}
			}
		}
	}
}

// handleConnections handles incoming connections
func (dm *DiscoveryManager) handleConnections() {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		fmt.Printf("Failed to start listener: %v\n", err)
		return
	}
	defer listener.Close()

	for {
		select {
		case <-dm.ctx.Done():
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				fmt.Printf("Failed to accept connection: %v\n", err)
				continue
			}

			// Handle connection
			go dm.handleConnection(conn)
		}
	}
}

// handleConnection handles an incoming connection
func (dm *DiscoveryManager) handleConnection(conn net.Conn) {
	addr := conn.RemoteAddr().String()

	// Check connection limit
	dm.mu.Lock()
	if len(dm.connections) >= dm.config.MaxConnections {
		dm.mu.Unlock()
		conn.Close()
		return
	}

	// Create peer info
	peer := &PeerInfo{
		Address:      addr,
		LastSeen:     time.Now(),
		IsBootstrap:  false,
		IsOutbound:   false,
		IsCompressed: true,
		IsTLS:        dm.config.EnableTLS,
	}

	// Add peer and connection
	dm.peers[addr] = peer
	dm.connections[addr] = conn
	dm.rateLimiters[addr] = NewRateLimiter(dm.config.MaxInboundRate, dm.config.MaxOutboundRate)
	dm.mu.Unlock()

	// Handle peer messages
	for {
		select {
		case <-dm.ctx.Done():
			return
		default:
			// Read message
			msg, err := dm.readMessage(conn)
			if err != nil {
				fmt.Printf("Failed to read message from %s: %v\n", addr, err)
				dm.disconnectPeer(addr)
				return
			}

			// Handle message
			if err := dm.handleMessage(addr, msg); err != nil {
				fmt.Printf("Failed to handle message from %s: %v\n", addr, err)
				dm.disconnectPeer(addr)
				return
			}
		}
	}
}

// disconnectPeer disconnects from a peer
func (dm *DiscoveryManager) disconnectPeer(addr string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Close connection
	if conn, exists := dm.connections[addr]; exists {
		conn.Close()
		delete(dm.connections, addr)
	}

	// Remove peer
	delete(dm.peers, addr)
	delete(dm.rateLimiters, addr)
}

// getRandomPeers returns a random selection of peers
func (dm *DiscoveryManager) getRandomPeers(count int) []*PeerInfo {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	peers := make([]*PeerInfo, 0, len(dm.peers))
	for _, peer := range dm.peers {
		peers = append(peers, peer)
	}

	// Shuffle peers
	rand.Shuffle(len(peers), func(i, j int) {
		peers[i], peers[j] = peers[j], peers[i]
	})

	// Return requested number of peers
	if count > len(peers) {
		count = len(peers)
	}
	return peers[:count]
}

// pingPeer pings a peer
func (dm *DiscoveryManager) pingPeer(addr string) (time.Duration, error) {
	start := time.Now()

	// Send ping message
	if err := dm.sendMessage(addr, "ping", nil); err != nil {
		return 0, err
	}

	// Wait for pong
	done := make(chan struct{})
	go func() {
		// Set timeout
		time.Sleep(dm.config.PingTimeout)
		close(done)
	}()

	// Wait for response
	select {
	case <-done:
		return 0, errors.New("ping timeout")
	default:
		// TODO: Implement actual pong handling
		return time.Since(start), nil
	}
}

// requestPeerList requests a list of peers from a peer
func (dm *DiscoveryManager) requestPeerList(addr string) error {
	return dm.sendMessage(addr, "getpeers", nil)
}

// readMessage reads a message from a connection
func (dm *DiscoveryManager) readMessage(conn net.Conn) ([]byte, error) {
	// TODO: Implement message reading with compression and rate limiting
	return nil, nil
}

// sendMessage sends a message to a peer
func (dm *DiscoveryManager) sendMessage(addr string, msgType string, payload interface{}) error {
	dm.mu.RLock()
	conn, exists := dm.connections[addr]
	limiter, exists := dm.rateLimiters[addr]
	dm.mu.RUnlock()

	if !exists {
		return errors.New("peer not found")
	}

	// Check rate limit
	if !limiter.AllowOutbound() {
		return errors.New("outbound rate limit exceeded")
	}

	// TODO: Implement message sending with compression
	return nil
}

// handleMessage handles a message from a peer
func (dm *DiscoveryManager) handleMessage(addr string, msg []byte) error {
	// TODO: Implement message handling
	return nil
}

// RateLimiter implements rate limiting
type RateLimiter struct {
	inboundRate   int64
	outboundRate  int64
	inboundCount  int64
	outboundCount int64
	lastReset     time.Time
	mu            sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(inboundRate, outboundRate int64) *RateLimiter {
	return &RateLimiter{
		inboundRate:  inboundRate,
		outboundRate: outboundRate,
		lastReset:    time.Now(),
	}
}

// AllowInbound checks if an inbound message is allowed
func (rl *RateLimiter) AllowInbound() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Reset counters if needed
	if time.Since(rl.lastReset) > time.Second {
		rl.inboundCount = 0
		rl.outboundCount = 0
		rl.lastReset = time.Now()
	}

	// Check rate limit
	if rl.inboundCount >= rl.inboundRate {
		return false
	}

	rl.inboundCount++
	return true
}

// AllowOutbound checks if an outbound message is allowed
func (rl *RateLimiter) AllowOutbound() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Reset counters if needed
	if time.Since(rl.lastReset) > time.Second {
		rl.inboundCount = 0
		rl.outboundCount = 0
		rl.lastReset = time.Now()
	}

	// Check rate limit
	if rl.outboundCount >= rl.outboundRate {
		return false
	}

	rl.outboundCount++
	return true
}
