package network

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/youngchain/internal/core/types"
)

// NodeConfig holds the configuration for a P2P node
type NodeConfig struct {
	ListenAddr       string        // Address to listen on
	Port             int           // Port to listen on
	Bootstrap        []string      // List of bootstrap nodes
	MaxPeers         int           // Maximum number of peers
	HandshakeTimeout time.Duration // Timeout for handshake
	PingInterval     time.Duration // Interval for ping messages
}

// Node represents a P2P node in the network
type Node struct {
	config     *NodeConfig
	listener   net.Listener
	peers      map[string]*Peer
	peersMutex sync.RWMutex
	stopChan   chan struct{}
	msgChan    chan *types.Message
	handlers   map[types.MessageType]MessageHandler
}

// MessageHandler is a function that handles a specific message type
type MessageHandler func(*Peer, *types.Message) error

// NewNode creates a new P2P node
func NewNode(config *NodeConfig) (*Node, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Set default values
	if config.MaxPeers == 0 {
		config.MaxPeers = 50
	}
	if config.HandshakeTimeout == 0 {
		config.HandshakeTimeout = 10 * time.Second
	}
	if config.PingInterval == 0 {
		config.PingInterval = 30 * time.Second
	}

	node := &Node{
		config:   config,
		peers:    make(map[string]*Peer),
		stopChan: make(chan struct{}),
		msgChan:  make(chan *types.Message, 1000),
		handlers: make(map[types.MessageType]MessageHandler),
	}

	// Register default message handlers
	node.registerDefaultHandlers()

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

	// Start message processing
	go node.processMessages()

	// Start peer maintenance
	go node.maintainPeers()

	return node, nil
}

// registerDefaultHandlers registers the default message handlers
func (n *Node) registerDefaultHandlers() {
	n.handlers[types.MsgVersion] = n.handleVersion
	n.handlers[types.MsgVerAck] = n.handleVerAck
	n.handlers[types.MsgPing] = n.handlePing
	n.handlers[types.MsgPong] = n.handlePong
	n.handlers[types.MsgGetBlocks] = n.handleGetBlocks
	n.handlers[types.MsgBlock] = n.handleBlock
	n.handlers[types.MsgTx] = n.handleTx
}

// processMessages processes incoming messages
func (n *Node) processMessages() {
	for {
		select {
		case <-n.stopChan:
			return
		case msg := <-n.msgChan:
			if handler, ok := n.handlers[msg.Type]; ok {
				if err := handler(nil, msg); err != nil {
					fmt.Printf("Error handling message: %v\n", err)
				}
			}
		}
	}
}

// maintainPeers maintains the peer list
func (n *Node) maintainPeers() {
	ticker := time.NewTicker(n.config.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-n.stopChan:
			return
		case <-ticker.C:
			n.peersMutex.RLock()
			for _, peer := range n.peers {
				if time.Since(peer.LastPing()) > n.config.PingInterval*2 {
					peer.Disconnect()
				}
			}
			n.peersMutex.RUnlock()
		}
	}
}

// handleNewConnection handles a new incoming connection
func (n *Node) handleNewConnection(conn net.Conn) {
	peer := NewPeer(conn)

	// Perform handshake
	if err := n.performHandshake(peer); err != nil {
		peer.Disconnect()
		return
	}

	// Check if we have too many peers
	n.peersMutex.Lock()
	if len(n.peers) >= n.config.MaxPeers {
		n.peersMutex.Unlock()
		peer.Disconnect()
		return
	}
	n.addPeer(peer)
	n.peersMutex.Unlock()

	peer.Start()
}

// performHandshake performs the initial handshake with a peer
func (n *Node) performHandshake(peer *Peer) error {
	// Send version message
	version := &types.Message{
		Type: types.MsgVersion,
		Payload: &types.VersionPayload{
			Version:   1,
			Services:  0,
			Timestamp: time.Now().Unix(),
			AddrRecv:  peer.RemoteAddr().String(),
			AddrFrom:  n.listener.Addr().String(),
		},
	}

	if err := peer.SendMessage(version); err != nil {
		return fmt.Errorf("failed to send version: %v", err)
	}

	// Wait for verack
	select {
	case <-time.After(n.config.HandshakeTimeout):
		return fmt.Errorf("handshake timeout")
	case msg := <-peer.MsgChan():
		if msg.Type != types.MsgVerAck {
			return fmt.Errorf("unexpected message type: %v", msg.Type)
		}
	}

	return nil
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

// Message handlers
func (n *Node) handleVersion(peer *Peer, msg *types.Message) error {
	// Send verack
	verack := &types.Message{
		Type: types.MsgVerAck,
	}
	return peer.SendMessage(verack)
}

func (n *Node) handleVerAck(peer *Peer, msg *types.Message) error {
	// Handshake complete
	return nil
}

func (n *Node) handlePing(peer *Peer, msg *types.Message) error {
	// Send pong
	pong := &types.Message{
		Type:    types.MsgPong,
		Payload: msg.Payload,
	}
	return peer.SendMessage(pong)
}

func (n *Node) handlePong(peer *Peer, msg *types.Message) error {
	// Update last ping time
	peer.UpdateLastPing()
	return nil
}

func (n *Node) handleGetBlocks(peer *Peer, msg *types.Message) error {
	// Handle getblocks request
	return nil
}

func (n *Node) handleBlock(peer *Peer, msg *types.Message) error {
	// Handle new block
	return nil
}

func (n *Node) handleTx(peer *Peer, msg *types.Message) error {
	// Handle new transaction
	return nil
}
