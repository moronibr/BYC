package network

import (
	"fmt"
	"net"
	"sync"
	"time"

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
	peers    map[string]*Peer
	mu       sync.RWMutex
	config   *config.Config
	listener net.Listener
	quitChan chan struct{}
}

// NewPeerManager creates a new peer manager
func NewPeerManager(config *config.Config) *PeerManager {
	return &PeerManager{
		peers:    make(map[string]*Peer),
		config:   config,
		quitChan: make(chan struct{}),
	}
}

// Start starts the peer manager
func (pm *PeerManager) Start() error {
	// Start listening for incoming connections
	listener, err := net.Listen("tcp", pm.config.ListenAddr)
	if err != nil {
		return err
	}
	pm.listener = listener

	// Start accepting connections
	go pm.acceptLoop()

	// Start peer discovery
	go pm.discoveryLoop()

	return nil
}

// Stop stops the peer manager
func (pm *PeerManager) Stop() error {
	close(pm.quitChan)
	if pm.listener != nil {
		if err := pm.listener.Close(); err != nil {
			return fmt.Errorf("failed to close listener: %v", err)
		}
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	for _, peer := range pm.peers {
		if err := peer.Stop(); err != nil {
			return fmt.Errorf("failed to stop peer: %v", err)
		}
	}

	return nil
}

func (pm *PeerManager) acceptLoop() {
	for {
		select {
		case <-pm.quitChan:
			return
		default:
			conn, err := pm.listener.Accept()
			if err != nil {
				fmt.Printf("Failed to accept connection: %v\n", err)
				continue
			}

			go pm.handleConnection(conn)
		}
	}
}

func (pm *PeerManager) handleConnection(conn net.Conn) {
	// Create peer info
	info := PeerInfo{
		ID:        conn.RemoteAddr().String(),
		Address:   conn.RemoteAddr().String(),
		Version:   "1.0.0", // TODO: Get from config
		LastSeen:  time.Now(),
		Connected: true,
	}

	// Create and start peer
	peer := NewPeer(conn, info)
	if err := peer.Start(); err != nil {
		fmt.Printf("Failed to start peer: %v\n", err)
		conn.Close()
		return
	}

	// Add peer to manager
	pm.addPeer(peer)
}

func (pm *PeerManager) discoveryLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-pm.quitChan:
			return
		case <-ticker.C:
			pm.discoverPeers()
		}
	}
}

func (pm *PeerManager) discoverPeers() {
	// Try to connect to bootstrap nodes
	for _, addr := range pm.config.BootstrapNodes {
		if pm.getPeerCount() >= pm.config.MaxPeers {
			return
		}

		// Skip if already connected
		if pm.hasPeer(addr) {
			continue
		}

		// Try to connect
		go func(addr string) {
			conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
			if err != nil {
				fmt.Printf("Failed to connect to peer %s: %v\n", addr, err)
				return
			}

			pm.handleConnection(conn)
		}(addr)
	}
}

func (pm *PeerManager) addPeer(peer *Peer) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.peers[peer.GetInfo().ID] = peer
}

func (pm *PeerManager) removePeer(id string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if peer, exists := pm.peers[id]; exists {
		if err := peer.Stop(); err != nil {
			fmt.Printf("Failed to stop peer %s: %v\n", id, err)
		}
		delete(pm.peers, id)
	}
}

func (pm *PeerManager) getPeer(id string) *Peer {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return pm.peers[id]
}

func (pm *PeerManager) hasPeer(addr string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for _, peer := range pm.peers {
		if peer.GetInfo().Address == addr {
			return true
		}
	}
	return false
}

func (pm *PeerManager) getPeerCount() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return len(pm.peers)
}

// Broadcast sends data to all connected peers
func (pm *PeerManager) Broadcast(data []byte) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for _, peer := range pm.peers {
		if err := peer.Send(data); err != nil {
			fmt.Printf("Failed to send data to peer %s: %v\n", peer.GetInfo().ID, err)
		}
	}
}

// ConnectToPeer connects to a peer at the given address
func (pm *PeerManager) ConnectToPeer(address string) error {
	if pm.getPeerCount() >= pm.config.MaxPeers {
		return fmt.Errorf("maximum number of peers reached")
	}

	if pm.hasPeer(address) {
		return fmt.Errorf("already connected to peer: %s", address)
	}

	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to peer: %v", err)
	}

	go pm.handleConnection(conn)
	return nil
}
