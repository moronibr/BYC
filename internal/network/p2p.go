package network

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// Node represents a P2P node
type Node struct {
	host     host.Host
	peers    map[peer.ID]*P2PPeer
	config   *NodeConfig
	ctx      context.Context
	cancel   context.CancelFunc
	stopChan chan struct{}
	mu       sync.RWMutex
}

// Peer represents a connected peer
type P2PPeer struct {
	ID       peer.ID
	Address  multiaddr.Multiaddr
	LastSeen time.Time
	Latency  time.Duration
	IsActive bool
	mu       sync.RWMutex
}

// UpdateLastSeen updates the last seen timestamp
func (p *P2PPeer) UpdateLastSeen() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.LastSeen = time.Now()
}

// SetActive sets the peer's active status
func (p *P2PPeer) SetActive(active bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.IsActive = active
}

// UpdateLatency updates the peer's latency
func (p *P2PPeer) UpdateLatency(latency time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Latency = latency
}

// IsPeerActive returns whether the peer is active
func (p *P2PPeer) IsPeerActive() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.IsActive
}

// Config holds the node configuration
type Config struct {
	ListenAddr string
	Port       int
	Bootstrap  []string
}

// NodeConfig holds configuration for a P2P node
type NodeConfig struct {
	ListenAddr     string
	BootstrapPeers []string
	MaxPeers       int
	NetworkID      uint32
}

// NewNode creates a new P2P node
func NewNode(config *NodeConfig) (*Node, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Create libp2p host
	h, err := libp2p.New(
		libp2p.ListenAddrStrings(config.ListenAddr),
	)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create host: %v", err)
	}

	node := &Node{
		host:     h,
		peers:    make(map[peer.ID]*P2PPeer),
		config:   config,
		ctx:      ctx,
		cancel:   cancel,
		stopChan: make(chan struct{}),
	}

	// Start peer discovery
	go node.discoverPeers()

	return node, nil
}

// start initializes and starts the node
func (n *Node) start() error {
	// Set up stream handlers
	n.host.SetStreamHandler("/blockchain/1.0.0", n.handleBlockchainStream)
	n.host.SetStreamHandler("/transaction/1.0.0", n.handleTransactionStream)

	// Connect to bootstrap nodes
	if err := n.connectBootstrapNodes(); err != nil {
		return fmt.Errorf("failed to connect to bootstrap nodes: %v", err)
	}

	// Start peer discovery
	go n.discoverPeers()

	return nil
}

// connectBootstrapNodes connects to the bootstrap nodes
func (n *Node) connectBootstrapNodes() error {
	for _, addr := range n.config.BootstrapPeers {
		ma, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			continue
		}

		peerInfo, err := peer.AddrInfoFromP2pAddr(ma)
		if err != nil {
			continue
		}

		if err := n.host.Connect(n.ctx, *peerInfo); err != nil {
			continue
		}

		n.addPeer(peerInfo.ID, ma)
	}

	return nil
}

// discoverPeers continuously discovers new peers
func (n *Node) discoverPeers() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			// Implement peer discovery logic here
			// This could use DHT, mDNS, or other discovery mechanisms
		}
	}
}

// addPeer adds a new peer to the node's peer list
func (n *Node) addPeer(id peer.ID, addr multiaddr.Multiaddr) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.peers[id] = &P2PPeer{
		ID:       id,
		Address:  addr,
		LastSeen: time.Now(),
		IsActive: true,
	}
}

// removePeer removes a peer from the node's peer list
func (n *Node) removePeer(id peer.ID) {
	n.mu.Lock()
	defer n.mu.Unlock()

	delete(n.peers, id)
}

// cleanupInactivePeers removes peers that haven't been seen recently
func (n *Node) cleanupInactivePeers() {
	n.mu.Lock()
	defer n.mu.Unlock()

	now := time.Now()
	for id, peer := range n.peers {
		if now.Sub(peer.LastSeen) > 30*time.Minute {
			n.removePeer(id)
		}
	}
}

// BroadcastBlock broadcasts a new block to all peers
func (n *Node) BroadcastBlock(block []byte) error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	for _, peer := range n.peers {
		if !peer.IsPeerActive() {
			continue
		}

		stream, err := n.host.NewStream(n.ctx, peer.ID, "/blockchain/1.0.0")
		if err != nil {
			peer.SetActive(false)
			continue
		}

		if _, err := stream.Write(block); err != nil {
			stream.Reset()
			peer.SetActive(false)
			continue
		}

		peer.UpdateLastSeen()
		stream.Close()
	}

	// Clean up inactive peers
	go n.cleanupInactivePeers()

	return nil
}

// BroadcastTransaction broadcasts a new transaction to all peers
func (n *Node) BroadcastTransaction(tx []byte) error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	for _, peer := range n.peers {
		if !peer.IsPeerActive() {
			continue
		}

		stream, err := n.host.NewStream(n.ctx, peer.ID, "/transaction/1.0.0")
		if err != nil {
			peer.SetActive(false)
			continue
		}

		if _, err := stream.Write(tx); err != nil {
			stream.Reset()
			peer.SetActive(false)
			continue
		}

		peer.UpdateLastSeen()
		stream.Close()
	}

	// Clean up inactive peers
	go n.cleanupInactivePeers()

	return nil
}

// handleBlockchainStream handles incoming blockchain streams
func (n *Node) handleBlockchainStream(stream network.Stream) {
	// Implement blockchain stream handling logic
}

// handleTransactionStream handles incoming transaction streams
func (n *Node) handleTransactionStream(stream network.Stream) {
	// Implement transaction stream handling logic
}

// Close shuts down the node
func (n *Node) Close() error {
	n.cancel()
	if err := n.host.Close(); err != nil {
		return fmt.Errorf("error closing host: %v", err)
	}
	close(n.stopChan)
	return nil
}
