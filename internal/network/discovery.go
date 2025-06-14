package network

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"sync"
	"time"

	"byc/internal/blockchain"
	"byc/internal/logger"

	"go.uber.org/zap"
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

// BootstrapNode represents a known bootstrap node
type BootstrapNode struct {
	Address   string
	LastSeen  time.Time
	IsActive  bool
	Version   string
	BlockType string
}

// DiscoveryManager manages peer discovery and network security
type DiscoveryManager struct {
	config         *DiscoveryConfig
	blockchain     *blockchain.Blockchain
	peers          map[string]*PeerInfo
	connections    map[string]net.Conn
	rateLimiters   map[string]*RateLimiter
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	bootstrapNodes map[string]*BootstrapNode
	knownPeers     map[string]*Peer
	node           *Node
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
func NewDiscoveryManager(node *Node, config *DiscoveryConfig) *DiscoveryManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &DiscoveryManager{
		config:         config,
		blockchain:     nil,
		peers:          make(map[string]*PeerInfo),
		connections:    make(map[string]net.Conn),
		rateLimiters:   make(map[string]*RateLimiter),
		mu:             sync.RWMutex{},
		ctx:            ctx,
		cancel:         cancel,
		bootstrapNodes: make(map[string]*BootstrapNode),
		knownPeers:     make(map[string]*Peer),
		node:           node,
	}
}

// Start starts the discovery manager
func (dm *DiscoveryManager) Start() error {
	// Load bootstrap nodes from config
	dm.loadBootstrapNodes()

	// Start periodic discovery
	go dm.startPeriodicDiscovery()

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

// loadBootstrapNodes loads bootstrap nodes from configuration
func (dm *DiscoveryManager) loadBootstrapNodes() {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Add default bootstrap nodes
	defaultNodes := []string{
		"bootstrap1.byc.network:3000",
		"bootstrap2.byc.network:3000",
		"bootstrap3.byc.network:3000",
	}

	for _, addr := range defaultNodes {
		dm.bootstrapNodes[addr] = &BootstrapNode{
			Address:   addr,
			LastSeen:  time.Now(),
			IsActive:  true,
			Version:   "1.0.0",
			BlockType: "golden", // Default block type
		}
	}

	// Add custom bootstrap nodes from config
	for _, addr := range dm.config.BootstrapNodes {
		dm.bootstrapNodes[addr] = &BootstrapNode{
			Address:   addr,
			LastSeen:  time.Now(),
			IsActive:  true,
			Version:   "1.0.0",
			BlockType: "golden", // Default block type
		}
	}
}

// startPeriodicDiscovery starts periodic peer discovery
func (dm *DiscoveryManager) startPeriodicDiscovery() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		dm.discoverPeers()
		dm.cleanupInactivePeers()
	}
}

// discoverPeers attempts to discover new peers
func (dm *DiscoveryManager) discoverPeers() {
	dm.mu.RLock()
	bootstrapNodes := make([]*BootstrapNode, 0, len(dm.bootstrapNodes))
	for _, node := range dm.bootstrapNodes {
		bootstrapNodes = append(bootstrapNodes, node)
	}
	dm.mu.RUnlock()

	// Try to connect to bootstrap nodes
	for _, node := range bootstrapNodes {
		if !node.IsActive {
			continue
		}

		// Send getaddr message to bootstrap node
		if err := dm.node.sendMessage(&Peer{Address: node.Address}, MessageTypeGetAddr, nil); err != nil {
			logger.Error("Failed to send getaddr to bootstrap node",
				zap.String("address", node.Address),
				zap.Error(err))
			continue
		}

		// Update last seen time
		dm.mu.Lock()
		node.LastSeen = time.Now()
		dm.mu.Unlock()
	}

	// Get random peers to request addresses from
	peers := dm.GetRandomPeers(10)
	for _, peer := range peers {
		if err := dm.node.sendMessage(peer, MessageTypeGetAddr, nil); err != nil {
			logger.Error("Failed to send getaddr to peer",
				zap.String("address", peer.Address),
				zap.Error(err))
			continue
		}
	}
}

// cleanupInactivePeers removes inactive peers
func (dm *DiscoveryManager) cleanupInactivePeers() {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	now := time.Now()
	for addr, peer := range dm.knownPeers {
		if now.Sub(peer.LastSeen) > 30*time.Minute {
			delete(dm.knownPeers, addr)
		}
	}
}

// AddPeer adds a new peer to the known peers list
func (dm *DiscoveryManager) AddPeer(peer *Peer) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	dm.knownPeers[peer.Address] = peer
}

// GetRandomPeers returns a random selection of peers
func (dm *DiscoveryManager) GetRandomPeers(count int) []*Peer {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	peers := make([]*Peer, 0, len(dm.knownPeers))
	for _, peer := range dm.knownPeers {
		peers = append(peers, peer)
	}

	// Shuffle peers
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(peers), func(i, j int) {
		peers[i], peers[j] = peers[j], peers[i]
	})

	// Return requested number of peers
	if count > len(peers) {
		count = len(peers)
	}
	return peers[:count]
}

// HandleAddr handles incoming addr messages
func (dm *DiscoveryManager) HandleAddr(addrs []string) {
	for _, addr := range addrs {
		// Validate address
		if _, err := net.ResolveTCPAddr("tcp", addr); err != nil {
			continue
		}

		// Add to known peers
		dm.AddPeer(&Peer{
			Address:  addr,
			LastSeen: time.Now(),
		})

		// Try to connect to new peer
		go dm.node.connectToPeer(addr)
	}
}

// GetBootstrapNodes returns the list of bootstrap nodes
func (dm *DiscoveryManager) GetBootstrapNodes() []*BootstrapNode {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	nodes := make([]*BootstrapNode, 0, len(dm.bootstrapNodes))
	for _, node := range dm.bootstrapNodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// UpdateBootstrapNode updates a bootstrap node's status
func (dm *DiscoveryManager) UpdateBootstrapNode(addr string, isActive bool) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if node, exists := dm.bootstrapNodes[addr]; exists {
		node.IsActive = isActive
		node.LastSeen = time.Now()
	}
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

// handleConnection handles a new connection
func (dm *DiscoveryManager) handleConnection(conn net.Conn) {
	defer conn.Close()

	addr := conn.RemoteAddr().String()

	// Check connection limit
	dm.mu.Lock()
	if len(dm.connections) >= dm.config.MaxConnections {
		dm.mu.Unlock()
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
	// Read message length (4 bytes)
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, lenBuf); err != nil {
		return nil, fmt.Errorf("failed to read message length: %v", err)
	}

	// Parse message length
	msgLen := binary.BigEndian.Uint32(lenBuf)
	if msgLen > 1024*1024 { // 1MB max message size
		return nil, fmt.Errorf("message too large: %d bytes", msgLen)
	}

	// Read message payload
	msgBuf := make([]byte, msgLen)
	if _, err := io.ReadFull(conn, msgBuf); err != nil {
		return nil, fmt.Errorf("failed to read message payload: %v", err)
	}

	return msgBuf, nil
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

	// Create message
	msg := struct {
		Type    string      `json:"type"`
		Payload interface{} `json:"payload"`
	}{
		Type:    msgType,
		Payload: payload,
	}

	// Marshal message
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	// Send message length
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(data)))
	if _, err := conn.Write(lenBuf); err != nil {
		return fmt.Errorf("failed to write message length: %v", err)
	}

	// Send message payload
	if _, err := conn.Write(data); err != nil {
		return fmt.Errorf("failed to write message payload: %v", err)
	}

	return nil
}

// handleMessage handles a message from a peer
func (dm *DiscoveryManager) handleMessage(addr string, msg []byte) error {
	// Unmarshal message
	var data struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(msg, &data); err != nil {
		return fmt.Errorf("failed to unmarshal message: %v", err)
	}

	// Update peer's last seen time
	dm.mu.Lock()
	if peer, exists := dm.peers[addr]; exists {
		peer.LastSeen = time.Now()
	}
	dm.mu.Unlock()

	// Handle message based on type
	switch data.Type {
	case "ping":
		return dm.sendMessage(addr, "pong", nil)
	case "pong":
		// Update peer latency
		dm.mu.Lock()
		if peer, exists := dm.peers[addr]; exists {
			peer.Latency = time.Since(peer.LastSeen)
		}
		dm.mu.Unlock()
		return nil
	case "getpeers":
		// Get random peers
		peers := dm.GetRandomPeers(10)
		peerAddrs := make([]string, len(peers))
		for i, peer := range peers {
			peerAddrs[i] = peer.Address
		}
		return dm.sendMessage(addr, "peers", peerAddrs)
	case "peers":
		// Parse peer list
		var peerAddrs []string
		if err := json.Unmarshal(data.Payload, &peerAddrs); err != nil {
			return fmt.Errorf("failed to unmarshal peer list: %v", err)
		}

		// Connect to new peers
		for _, peerAddr := range peerAddrs {
			if peerAddr != addr && !dm.isPeerConnected(peerAddr) {
				go dm.connectToPeer(peerAddr, false)
			}
		}
		return nil
	default:
		return fmt.Errorf("unknown message type: %s", data.Type)
	}
}

// isPeerConnected checks if a peer is already connected
func (dm *DiscoveryManager) isPeerConnected(addr string) bool {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	_, exists := dm.peers[addr]
	return exists
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
