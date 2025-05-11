package network

import (
	"fmt"
	"net"
	"sync"

	"github.com/youngchain/internal/core/types"
)

// NodeConfig holds the configuration for a P2P node
type NodeConfig struct {
	ListenAddr string   // Address to listen on
	Port       int      // Port to listen on
	Bootstrap  []string // List of bootstrap nodes
}

// Node represents a P2P node in the network
type Node struct {
	config     *NodeConfig
	listener   net.Listener
	peers      map[string]*Peer
	peersMutex sync.RWMutex
	stopChan   chan struct{}
}

// NewNode creates a new P2P node
func NewNode(config *NodeConfig) (*Node, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	node := &Node{
		config:   config,
		peers:    make(map[string]*Peer),
		stopChan: make(chan struct{}),
	}

	// Start listening for incoming connections
	addr := fmt.Sprintf("%s:%d", config.ListenAddr, config.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to start listener: %v", err)
	}
	node.listener = listener

	// Start accepting connections
	go node.acceptConnections()

	// Connect to bootstrap nodes
	go node.connectToBootstrapNodes()

	return node, nil
}

// Close stops the node and closes all connections
func (n *Node) Close() error {
	close(n.stopChan)
	if n.listener != nil {
		return n.listener.Close()
	}
	return nil
}

// acceptConnections accepts incoming connections
func (n *Node) acceptConnections() {
	for {
		select {
		case <-n.stopChan:
			return
		default:
			conn, err := n.listener.Accept()
			if err != nil {
				continue
			}
			go n.handleNewConnection(conn)
		}
	}
}

// handleNewConnection handles a new incoming connection
func (n *Node) handleNewConnection(conn net.Conn) {
	peer := NewPeer(conn)
	n.addPeer(peer)
	peer.Start()
}

// addPeer adds a new peer to the node
func (n *Node) addPeer(peer *Peer) {
	n.peersMutex.Lock()
	defer n.peersMutex.Unlock()

	addr := peer.RemoteAddr().String()
	n.peers[addr] = peer
}

// connectToBootstrapNodes connects to bootstrap nodes
func (n *Node) connectToBootstrapNodes() {
	for _, addr := range n.config.Bootstrap {
		go func(addr string) {
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				return
			}
			n.handleNewConnection(conn)
		}(addr)
	}
}

// BroadcastMessage broadcasts a message to all peers
func (n *Node) BroadcastMessage(msg *types.Message) {
	n.peersMutex.RLock()
	defer n.peersMutex.RUnlock()

	for _, peer := range n.peers {
		peer.SendMessage(msg)
	}
}

// GetPeerCount returns the number of connected peers
func (n *Node) GetPeerCount() int {
	n.peersMutex.RLock()
	defer n.peersMutex.RUnlock()
	return len(n.peers)
}
