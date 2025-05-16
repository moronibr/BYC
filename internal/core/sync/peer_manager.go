package sync

import (
	"net"
	"sync"
	"time"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/logger"
)

const (
	// Maximum number of peers
	MaxPeers = 125

	// Minimum number of peers
	MinPeers = 8

	// Peer discovery interval
	PeerDiscoveryInterval = 5 * time.Minute

	// Peer connection timeout
	PeerConnTimeout = 30 * time.Second

	// Peer handshake timeout
	PeerHandshakeTimeout = 30 * time.Second
)

// PeerManager represents a peer manager
type PeerManager struct {
	mu sync.RWMutex

	// Peers
	peers map[string]*Peer

	// Block sync
	blockSync *BlockSync

	// Stop channel
	stopChan chan struct{}

	// Listen address
	listenAddr string

	// Seed nodes
	seedNodes []string
}

// NewPeerManager creates a new peer manager
func NewPeerManager(listenAddr string, blockSync *BlockSync) *PeerManager {
	return &PeerManager{
		peers:      make(map[string]*Peer),
		blockSync:  blockSync,
		stopChan:   make(chan struct{}),
		listenAddr: listenAddr,
		seedNodes: []string{
			"seed.bitcoin.sipa.be:8333",
			"dnsseed.bluematt.me:8333",
			"dnsseed.bitcoin.dashjr.org:8333",
			"seed.bitcoinstats.com:8333",
			"seed.bitcoin.jonasschnelli.ch:8333",
			"seed.btc.petertodd.org:8333",
			"seed.bitcoin.sprovoost.nl:8333",
			"dnsseed.emzy.de:8333",
			"seed.bitcoin.wiz.biz:8333",
		},
	}
}

// Start starts the peer manager
func (pm *PeerManager) Start() {
	// Start listening for incoming connections
	go pm.listen()

	// Start peer discovery
	go pm.discoverPeers()

	// Start peer maintenance
	go pm.maintainPeers()
}

// Stop stops the peer manager
func (pm *PeerManager) Stop() {
	close(pm.stopChan)

	// Stop all peers
	pm.mu.Lock()
	for _, peer := range pm.peers {
		peer.Stop()
	}
	pm.mu.Unlock()
}

// listen listens for incoming connections
func (pm *PeerManager) listen() {
	// Create listener
	listener, err := net.Listen("tcp", pm.listenAddr)
	if err != nil {
		logger.Error("Failed to listen", logger.Error(err))
		return
	}
	defer listener.Close()

	for {
		select {
		case <-pm.stopChan:
			return
		default:
			// Accept connection
			conn, err := listener.Accept()
			if err != nil {
				logger.Error("Failed to accept connection", logger.Error(err))
				continue
			}

			// Handle connection
			go pm.handleConnection(conn)
		}
	}
}

// handleConnection handles an incoming connection
func (pm *PeerManager) handleConnection(conn net.Conn) {
	// Create peer
	peer := NewPeer(conn, conn.RemoteAddr().String(), pm.blockSync)

	// Add peer
	pm.mu.Lock()
	if len(pm.peers) >= MaxPeers {
		pm.mu.Unlock()
		conn.Close()
		return
	}
	pm.peers[peer.address] = peer
	pm.mu.Unlock()

	// Start peer
	peer.Start()

	// Send version
	if err := peer.SendVersion(); err != nil {
		logger.Error("Failed to send version", logger.Error(err))
		pm.removePeer(peer)
		return
	}

	// Wait for verack
	select {
	case <-time.After(PeerHandshakeTimeout):
		logger.Error("Handshake timeout")
		pm.removePeer(peer)
		return
	}
}

// discoverPeers discovers new peers
func (pm *PeerManager) discoverPeers() {
	ticker := time.NewTicker(PeerDiscoveryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pm.stopChan:
			return
		case <-ticker.C:
			// Check if we need more peers
			pm.mu.RLock()
			needPeers := len(pm.peers) < MinPeers
			pm.mu.RUnlock()

			if needPeers {
				// Connect to seed nodes
				for _, seed := range pm.seedNodes {
					go pm.connectToPeer(seed)
				}
			}
		}
	}
}

// maintainPeers maintains the peer list
func (pm *PeerManager) maintainPeers() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-pm.stopChan:
			return
		case <-ticker.C:
			// Check peers
			pm.mu.Lock()
			for addr, peer := range pm.peers {
				// Check if peer is still alive
				if time.Since(peer.lastSeen) > PeerConnTimeout {
					peer.Stop()
					delete(pm.peers, addr)
				}
			}
			pm.mu.Unlock()
		}
	}
}

// connectToPeer connects to a peer
func (pm *PeerManager) connectToPeer(addr string) {
	// Connect to peer
	conn, err := net.DialTimeout("tcp", addr, PeerConnTimeout)
	if err != nil {
		logger.Error("Failed to connect to peer", logger.Error(err))
		return
	}

	// Handle connection
	pm.handleConnection(conn)
}

// removePeer removes a peer
func (pm *PeerManager) removePeer(peer *Peer) {
	pm.mu.Lock()
	delete(pm.peers, peer.address)
	pm.mu.Unlock()
	peer.Stop()
}

// GetPeers returns the current peers
func (pm *PeerManager) GetPeers() []*Peer {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	peers := make([]*Peer, 0, len(pm.peers))
	for _, peer := range pm.peers {
		peers = append(peers, peer)
	}

	return peers
}

// BroadcastBlock broadcasts a block to all peers
func (pm *PeerManager) BroadcastBlock(block *block.Block) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for _, peer := range pm.peers {
		if err := peer.SendBlock(block); err != nil {
			logger.Error("Failed to send block", logger.Error(err))
		}
	}
}

// BroadcastInv broadcasts an inventory message to all peers
func (pm *PeerManager) BroadcastInv(inventory []byte) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for _, peer := range pm.peers {
		if err := peer.SendInv(inventory); err != nil {
			logger.Error("Failed to send inv", logger.Error(err))
		}
	}
}
