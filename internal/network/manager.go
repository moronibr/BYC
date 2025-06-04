package network

import (
	"fmt"
	"sync"
)

// NetworkManager manages the P2P network
type NetworkManager struct {
	config *NetworkConfig
	peers  map[string]*Peer
	mu     sync.RWMutex
}

// NewNetworkManager creates a new network manager
func NewNetworkManager(config *NetworkConfig) *NetworkManager {
	return &NetworkManager{
		config: config,
		peers:  make(map[string]*Peer),
	}
}

// AddPeer adds a peer to the network
func (nm *NetworkManager) AddPeer(peer *Peer) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	nm.peers[peer.Address] = peer
}

// RemovePeer removes a peer from the network
func (nm *NetworkManager) RemovePeer(address string) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	delete(nm.peers, address)
}

// GetPeer returns a peer by address
func (nm *NetworkManager) GetPeer(address string) *Peer {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	return nm.peers[address]
}

// GetPeers returns all peers
func (nm *NetworkManager) GetPeers() []*Peer {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	peers := make([]*Peer, 0, len(nm.peers))
	for _, peer := range nm.peers {
		peers = append(peers, peer)
	}
	return peers
}

// SendMessage sends a message to a peer
func (nm *NetworkManager) SendMessage(msg *NetworkMessage) error {
	peer := nm.GetPeer(msg.To)
	if peer == nil {
		return fmt.Errorf("peer %s not found", msg.To)
	}
	return peer.sendMessage(*msg)
}

// handleMessage handles a received message
func (nm *NetworkManager) handleMessage(msg *NetworkMessage) error {
	// TODO: Implement message handling
	return nil
}

// HasReceivedPong checks if the network manager has received a pong message
func (nm *NetworkManager) HasReceivedPong(peerAddr string) bool {
	// Implementation needed
	return false
}

// ConnectToPeer connects to a peer
func (nm *NetworkManager) ConnectToPeer(address string) error {
	// Implementation needed
	return nil
}
