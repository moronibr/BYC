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
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/multiformats/go-multiaddr"
)

// Node represents a P2P node in the network
type Node struct {
	host   host.Host
	peers  map[peer.ID]*Peer
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
	config *Config
}

// Peer represents a connected peer
type Peer struct {
	ID       peer.ID
	Address  multiaddr.Multiaddr
	LastSeen time.Time
	Latency  time.Duration
	IsActive bool
	mu       sync.RWMutex
}

// Config holds the node configuration
type Config struct {
	ListenAddr string
	Port       int
	Bootstrap  []string
}

// NewNode creates a new P2P node
func NewNode(config *Config) (*Node, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Create a new libp2p host
	h, err := libp2p.New(
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", config.Port)),
		libp2p.Security(noise.ID, noise.New),
	)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create host: %v", err)
	}

	node := &Node{
		host:   h,
		peers:  make(map[peer.ID]*Peer),
		ctx:    ctx,
		cancel: cancel,
		config: config,
	}

	// Start the node
	if err := node.start(); err != nil {
		cancel()
		return nil, err
	}

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
	for _, addr := range n.config.Bootstrap {
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

	n.peers[id] = &Peer{
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

// BroadcastBlock broadcasts a new block to all peers
func (n *Node) BroadcastBlock(block []byte) error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	for _, peer := range n.peers {
		if !peer.IsActive {
			continue
		}

		stream, err := n.host.NewStream(n.ctx, peer.ID, "/blockchain/1.0.0")
		if err != nil {
			continue
		}

		if _, err := stream.Write(block); err != nil {
			stream.Reset()
			continue
		}

		stream.Close()
	}

	return nil
}

// BroadcastTransaction broadcasts a new transaction to all peers
func (n *Node) BroadcastTransaction(tx []byte) error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	for _, peer := range n.peers {
		if !peer.IsActive {
			continue
		}

		stream, err := n.host.NewStream(n.ctx, peer.ID, "/transaction/1.0.0")
		if err != nil {
			continue
		}

		if _, err := stream.Write(tx); err != nil {
			stream.Reset()
			continue
		}

		stream.Close()
	}

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
	return n.host.Close()
}
