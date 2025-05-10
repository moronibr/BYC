package network

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"
)

// Peer represents a network peer
type Peer struct {
	ID        uint64
	Addr      net.Addr
	Version   uint32
	Services  uint64
	LastSeen  time.Time
	UserAgent string
	mu        sync.RWMutex
}

// PeerManager manages network peers
type PeerManager struct {
	peers     map[uint64]*Peer
	mu        sync.RWMutex
	maxPeers  int
	bootstrap []string
}

// NewPeerManager creates a new peer manager
func NewPeerManager(maxPeers int, bootstrap []string) *PeerManager {
	return &PeerManager{
		peers:     make(map[uint64]*Peer),
		maxPeers:  maxPeers,
		bootstrap: bootstrap,
	}
}

// AddPeer adds a peer to the manager
func (pm *PeerManager) AddPeer(peer *Peer) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if len(pm.peers) >= pm.maxPeers {
		return fmt.Errorf("maximum number of peers reached")
	}

	pm.peers[peer.ID] = peer
	return nil
}

// RemovePeer removes a peer from the manager
func (pm *PeerManager) RemovePeer(id uint64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	delete(pm.peers, id)
}

// GetPeer gets a peer by ID
func (pm *PeerManager) GetPeer(id uint64) (*Peer, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	peer, exists := pm.peers[id]
	return peer, exists
}

// GetRandomPeers gets a random subset of peers
func (pm *PeerManager) GetRandomPeers(count int) []*Peer {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if count > len(pm.peers) {
		count = len(pm.peers)
	}

	peers := make([]*Peer, 0, count)
	for _, peer := range pm.peers {
		peers = append(peers, peer)
		if len(peers) == count {
			break
		}
	}

	return peers
}

// DiscoverPeers discovers new peers
func (pm *PeerManager) DiscoverPeers() error {
	// Try bootstrap nodes first
	for _, addr := range pm.bootstrap {
		peer, err := pm.connectToPeer(addr)
		if err == nil {
			pm.AddPeer(peer)
		}
	}

	// Ask existing peers for more peers
	peers := pm.GetRandomPeers(5)
	for _, peer := range peers {
		// Send getaddr message
		msg := NewMessage(0, MsgGetAddr, nil)
		// TODO: Send message to peer
	}

	return nil
}

// connectToPeer connects to a peer
func (pm *PeerManager) connectToPeer(addr string) (*Peer, error) {
	// Generate random peer ID
	id := make([]byte, 8)
	rand.Read(id)
	peerID := binary.BigEndian.Uint64(id)

	// Create peer
	peer := &Peer{
		ID:       peerID,
		LastSeen: time.Now(),
	}

	// Parse address
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	peer.Addr = tcpAddr

	return peer, nil
}

// UpdatePeer updates peer information
func (pm *PeerManager) UpdatePeer(id uint64, version uint32, services uint64, userAgent string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if peer, exists := pm.peers[id]; exists {
		peer.Version = version
		peer.Services = services
		peer.UserAgent = userAgent
		peer.LastSeen = time.Now()
	}
}

// Cleanup removes inactive peers
func (pm *PeerManager) Cleanup(maxAge time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	now := time.Now()
	for id, peer := range pm.peers {
		if now.Sub(peer.LastSeen) > maxAge {
			delete(pm.peers, id)
		}
	}
}

// GetPeerCount returns the number of peers
func (pm *PeerManager) GetPeerCount() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return len(pm.peers)
}

// GetBootstrapPeers returns the bootstrap peers
func (pm *PeerManager) GetBootstrapPeers() []string {
	return pm.bootstrap
}

// SetBootstrapPeers sets the bootstrap peers
func (pm *PeerManager) SetBootstrapPeers(peers []string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.bootstrap = peers
}
