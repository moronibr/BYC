package network

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	"byc/internal/logger"

	"go.uber.org/zap"
)

// DiscoveryService manages peer discovery and network topology
type DiscoveryService struct {
	node            *Node
	knownPeers      map[string]*DiscoveryPeerInfo
	bootstrapNodes  []string
	discoveryTicker *time.Ticker
	ctx             context.Context
	cancel          context.CancelFunc
	mu              sync.RWMutex
}

// DiscoveryPeerInfo contains information about a discovered peer
type DiscoveryPeerInfo struct {
	Address     string
	LastSeen    time.Time
	Latency     time.Duration
	Version     string
	BlockHeight int64
	IsActive    bool
	IsBootstrap bool
	LastPing    time.Time
	FailCount   int
}

// NewDiscoveryService creates a new discovery service
func NewDiscoveryService(node *Node, bootstrapNodes []string) *DiscoveryService {
	ctx, cancel := context.WithCancel(context.Background())
	return &DiscoveryService{
		node:           node,
		knownPeers:     make(map[string]*DiscoveryPeerInfo),
		bootstrapNodes: bootstrapNodes,
		ctx:            ctx,
		cancel:         cancel,
	}
}

// Start begins the discovery service
func (ds *DiscoveryService) Start() error {
	logger.Info("Starting discovery service")

	// Start periodic discovery
	ds.discoveryTicker = time.NewTicker(60 * time.Second)
	go ds.runDiscovery()

	// Start peer monitoring
	go ds.monitorPeers()

	// Initial peer discovery
	go ds.discoverPeers()

	return nil
}

// Stop stops the discovery service
func (ds *DiscoveryService) Stop() {
	if ds.discoveryTicker != nil {
		ds.discoveryTicker.Stop()
	}
	ds.cancel()
	logger.Info("Discovery service stopped")
}

// runDiscovery runs the periodic discovery process
func (ds *DiscoveryService) runDiscovery() {
	for {
		select {
		case <-ds.discoveryTicker.C:
			ds.discoverPeers()
		case <-ds.ctx.Done():
			return
		}
	}
}

// discoverPeers discovers new peers from bootstrap nodes and existing peers
func (ds *DiscoveryService) discoverPeers() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	// Try bootstrap nodes first
	for _, bootstrapAddr := range ds.bootstrapNodes {
		if ds.shouldConnectToPeer(bootstrapAddr) {
			go ds.connectToPeer(bootstrapAddr, true)
		}
	}

	// Try known peers
	for addr, peerInfo := range ds.knownPeers {
		if peerInfo.IsActive && ds.shouldConnectToPeer(addr) {
			go ds.connectToPeer(addr, false)
		}
	}
}

// shouldConnectToPeer determines if we should connect to a peer
func (ds *DiscoveryService) shouldConnectToPeer(addr string) bool {
	// Don't connect to ourselves
	if addr == ds.node.Config.Address {
		return false
	}

	// Check if already connected
	ds.node.mu.RLock()
	_, alreadyConnected := ds.node.Peers[addr]
	ds.node.mu.RUnlock()

	if alreadyConnected {
		return false
	}

	// Check if we have too many connections
	ds.node.mu.RLock()
	peerCount := len(ds.node.Peers)
	ds.node.mu.RUnlock()

	if peerCount >= 50 { // Max peers limit
		return false
	}

	return true
}

// connectToPeer attempts to connect to a peer
func (ds *DiscoveryService) connectToPeer(addr string, isBootstrap bool) {
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		ds.updatePeerFailCount(addr)
		logger.Debug("Failed to connect to peer", zap.String("peer", addr), zap.Error(err))
		return
	}
	defer conn.Close()

	// Create peer
	peer := NewPeer(fmt.Sprintf("peer_%d", rand.Int()), addr, 0)
	peer.conn = conn
	peer.Node = ds.node
	peer.IsBootstrap = isBootstrap

	// Add to node's peer list
	ds.node.mu.Lock()
	ds.node.Peers[addr] = peer
	ds.node.mu.Unlock()

	// Update peer info
	ds.updatePeerInfo(addr, &DiscoveryPeerInfo{
		Address:     addr,
		LastSeen:    time.Now(),
		IsActive:    true,
		IsBootstrap: isBootstrap,
		FailCount:   0,
	})

	// Start message handling
	go peer.handleMessages()

	// Send version message
	if err := peer.sendVersion(); err != nil {
		logger.Error("Failed to send version", zap.String("peer", addr), zap.Error(err))
		ds.disconnectPeer(addr)
		return
	}

	// Request peer list
	go ds.requestPeerList(peer)

	logger.Info("Connected to peer", zap.String("peer", addr))
}

// requestPeerList requests the peer list from a peer
func (ds *DiscoveryService) requestPeerList(peer *Peer) {
	// Send getaddr message
	getAddrMsg := NetworkMessage{
		Type:      MessageTypeGetAddr,
		From:      ds.node.Config.Address,
		To:        peer.Address,
		Payload:   []byte("{}"),
		Timestamp: time.Now(),
	}

	if err := peer.sendMessage(getAddrMsg); err != nil {
		logger.Error("Failed to request peer list", zap.String("peer", peer.Address), zap.Error(err))
	}
}

// updatePeerInfo updates peer information
func (ds *DiscoveryService) updatePeerInfo(addr string, info *DiscoveryPeerInfo) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.knownPeers[addr] = info
}

// updatePeerFailCount updates the fail count for a peer
func (ds *DiscoveryService) updatePeerFailCount(addr string) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if peerInfo, exists := ds.knownPeers[addr]; exists {
		peerInfo.FailCount++
		peerInfo.LastSeen = time.Now()
		if peerInfo.FailCount > 5 {
			peerInfo.IsActive = false
		}
	} else {
		ds.knownPeers[addr] = &DiscoveryPeerInfo{
			Address:   addr,
			LastSeen:  time.Now(),
			FailCount: 1,
			IsActive:  false,
		}
	}
}

// disconnectPeer disconnects from a peer
func (ds *DiscoveryService) disconnectPeer(addr string) {
	ds.node.mu.Lock()
	if peer, exists := ds.node.Peers[addr]; exists {
		if peer.conn != nil {
			peer.conn.Close()
		}
		delete(ds.node.Peers, addr)
	}
	ds.node.mu.Unlock()

	ds.updatePeerInfo(addr, &DiscoveryPeerInfo{
		Address:  addr,
		LastSeen: time.Now(),
		IsActive: false,
	})

	logger.Info("Disconnected from peer", zap.String("peer", addr))
}

// monitorPeers monitors peer health and removes inactive peers
func (ds *DiscoveryService) monitorPeers() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ds.cleanupInactivePeers()
		case <-ds.ctx.Done():
			return
		}
	}
}

// cleanupInactivePeers removes inactive peers
func (ds *DiscoveryService) cleanupInactivePeers() {
	ds.node.mu.Lock()
	defer ds.node.mu.Unlock()

	now := time.Now()
	for addr, peer := range ds.node.Peers {
		// Check if peer hasn't been seen for too long
		if now.Sub(peer.LastSeen) > 5*time.Minute {
			logger.Info("Removing inactive peer", zap.String("peer", addr))
			if peer.conn != nil {
				peer.conn.Close()
			}
			delete(ds.node.Peers, addr)
		}
	}
}

// GetPeerList returns the list of known peers
func (ds *DiscoveryService) GetPeerList() []*DiscoveryPeerInfo {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	peers := make([]*DiscoveryPeerInfo, 0, len(ds.knownPeers))
	for _, peer := range ds.knownPeers {
		peers = append(peers, peer)
	}
	return peers
}

// AddPeer adds a peer to the known peers list
func (ds *DiscoveryService) AddPeer(addr string, isBootstrap bool) {
	ds.updatePeerInfo(addr, &DiscoveryPeerInfo{
		Address:     addr,
		LastSeen:    time.Now(),
		IsActive:    true,
		IsBootstrap: isBootstrap,
		FailCount:   0,
	})
}

// HandlePeerList handles received peer list from other nodes
func (ds *DiscoveryService) HandlePeerList(peers []string) {
	for _, addr := range peers {
		if addr != ds.node.Config.Address {
			ds.AddPeer(addr, false)
		}
	}
}

// BroadcastPeerList broadcasts our peer list to connected peers
func (ds *DiscoveryService) BroadcastPeerList() {
	ds.node.mu.RLock()
	addresses := make([]string, 0, len(ds.node.Peers))
	for _, peer := range ds.node.Peers {
		addresses = append(addresses, peer.Address)
	}
	ds.node.mu.RUnlock()

	addrMsg := AddrMessage{
		Addresses: addresses,
	}
	payload, _ := json.Marshal(addrMsg)

	msg := NetworkMessage{
		Type:      MessageTypeAddr,
		From:      ds.node.Config.Address,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	ds.node.BroadcastMessage(msg)
}

// GetRandomPeers returns a random selection of peers
func (ds *DiscoveryService) GetRandomPeers(count int) []string {
	ds.node.mu.RLock()
	defer ds.node.mu.RUnlock()

	peers := make([]string, 0, len(ds.node.Peers))
	for _, peer := range ds.node.Peers {
		peers = append(peers, peer.Address)
	}

	if count >= len(peers) {
		return peers
	}

	// Shuffle and return first count
	rand.Shuffle(len(peers), func(i, j int) {
		peers[i], peers[j] = peers[j], peers[i]
	})

	return peers[:count]
}

// GetPeerStats returns statistics about peers
func (ds *DiscoveryService) GetPeerStats() map[string]interface{} {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	ds.node.mu.RLock()
	connectedCount := len(ds.node.Peers)
	ds.node.mu.RUnlock()

	totalKnown := len(ds.knownPeers)
	activeCount := 0
	bootstrapCount := 0

	for _, peer := range ds.knownPeers {
		if peer.IsActive {
			activeCount++
		}
		if peer.IsBootstrap {
			bootstrapCount++
		}
	}

	return map[string]interface{}{
		"connected_peers":  connectedCount,
		"known_peers":      totalKnown,
		"active_peers":     activeCount,
		"bootstrap_peers":  bootstrapCount,
		"discovery_active": ds.discoveryTicker != nil,
	}
}
