package peers

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/youngchain/internal/network/types"
)

// PeerManager manages peer connections and discovery
type PeerManager struct {
	peers     map[string]*types.Node
	mu        sync.RWMutex
	maxPeers  int
	bootstrap []string
}

// NewPeerManager creates a new peer manager
func NewPeerManager(maxPeers int, bootstrap []string) *PeerManager {
	return &PeerManager{
		peers:     make(map[string]*types.Node),
		maxPeers:  maxPeers,
		bootstrap: bootstrap,
	}
}

// AddPeer adds a new peer
func (pm *PeerManager) AddPeer(addr string) (*types.Node, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if len(pm.peers) >= pm.maxPeers {
		return nil, fmt.Errorf("maximum number of peers reached")
	}

	if _, exists := pm.peers[addr]; exists {
		return nil, fmt.Errorf("peer already exists")
	}

	peer := types.NewNode(addr)
	pm.peers[addr] = peer
	return peer, nil
}

// RemovePeer removes a peer
func (pm *PeerManager) RemovePeer(addr string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	delete(pm.peers, addr)
}

// GetPeer returns a peer by address
func (pm *PeerManager) GetPeer(addr string) *types.Node {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.peers[addr]
}

// GetPeers returns all peers
func (pm *PeerManager) GetPeers() []*types.Node {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	peers := make([]*types.Node, 0, len(pm.peers))
	for _, peer := range pm.peers {
		peers = append(peers, peer)
	}
	return peers
}

// GetActivePeers returns all active peers
func (pm *PeerManager) GetActivePeers() []*types.Node {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	peers := make([]*types.Node, 0)
	for _, peer := range pm.peers {
		if peer.IsActive() {
			peers = append(peers, peer)
		}
	}
	return peers
}

// StartPeerDiscovery starts the peer discovery process
func (pm *PeerManager) StartPeerDiscovery() {
	go pm.discoverPeers()
}

// discoverPeers discovers new peers
func (pm *PeerManager) discoverPeers() {
	ticker := time.NewTicker(time.Minute * 5)
	defer ticker.Stop()

	for range ticker.C {
		// Try to connect to bootstrap nodes
		for _, addr := range pm.bootstrap {
			if len(pm.peers) >= pm.maxPeers {
				break
			}

			if _, err := pm.AddPeer(addr); err == nil {
				go pm.connectToPeer(addr)
			}
		}

		// Check peer status
		for _, peer := range pm.GetActivePeers() {
			// Use peer variable to check status
			_ = peer.IsActive() // Use the peer variable to avoid unused variable warning
		}
	}
}

// connectToPeer attempts to connect to a peer
func (pm *PeerManager) connectToPeer(addr string) {
	conn, err := net.DialTimeout("tcp", addr, time.Second*5)
	if err != nil {
		pm.RemovePeer(addr)
		return
	}
	defer conn.Close()

	// Create connection handler
	if peer := pm.GetPeer(addr); peer != nil {
		connection := NewConnection(conn, peer, pm)
		connection.Start()
	}
}
