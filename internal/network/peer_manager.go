package network

import (
	"fmt"
	"sync"

	"github.com/youngchain/internal/config"
)

// Peer represents a network peer
type Peer struct {
	address string
	height  uint64
	mu      sync.RWMutex
}

// NewPeer creates a new peer
func NewPeer(address string) *Peer {
	return &Peer{
		address: address,
		height:  0,
	}
}

// GetHeight returns the peer's height
func (p *Peer) GetHeight() uint64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.height
}

// SetHeight sets the peer's height
func (p *Peer) SetHeight(height uint64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.height = height
}

// PeerManager manages network peers
type PeerManager struct {
	config *config.Config
	peers  map[string]*Peer
	mu     sync.RWMutex
}

// NewPeerManager creates a new peer manager
func NewPeerManager(config *config.Config) *PeerManager {
	return &PeerManager{
		config: config,
		peers:  make(map[string]*Peer),
	}
}

// AddPeer adds a peer
func (pm *PeerManager) AddPeer(address string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.peers[address] = NewPeer(address)
}

// RemovePeer removes a peer
func (pm *PeerManager) RemovePeer(address string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	delete(pm.peers, address)
}

// GetPeer gets a peer
func (pm *PeerManager) GetPeer(address string) (*Peer, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	peer, ok := pm.peers[address]
	if !ok {
		return nil, fmt.Errorf("peer not found: %s", address)
	}
	return peer, nil
}

// GetPeers gets all peers
func (pm *PeerManager) GetPeers() []*Peer {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	peers := make([]*Peer, 0, len(pm.peers))
	for _, peer := range pm.peers {
		peers = append(peers, peer)
	}
	return peers
}

// Broadcast broadcasts a message to all peers
func (pm *PeerManager) Broadcast(message []byte) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	for _, peer := range pm.peers {
		// TODO: Implement actual message sending
		_ = peer
	}
}

// Start starts the peer manager
func (pm *PeerManager) Start() error {
	// TODO: Implement peer discovery
	return nil
}

// Stop stops the peer manager
func (pm *PeerManager) Stop() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.peers = make(map[string]*Peer)
}
