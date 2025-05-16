package network

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/byc/internal/blockchain"
)

// Message represents a message sent between nodes
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Node represents a P2P node
type Node struct {
	Address    string
	Peers      map[string]*Peer
	Blockchain *blockchain.Blockchain
	mu         sync.RWMutex
}

// Peer represents a connected peer
type Peer struct {
	Address string
	Conn    net.Conn
}

// handshakeMessage is sent immediately after connecting to share the listening address
const handshakeMessageType = "handshake"

// NewNode creates a new P2P node
func NewNode(address string, bc *blockchain.Blockchain) *Node {
	return &Node{
		Address:    address,
		Peers:      make(map[string]*Peer),
		Blockchain: bc,
	}
}

// Start starts the node and begins listening for connections
func (n *Node) Start() error {
	listener, err := net.Listen("tcp", n.Address)
	if err != nil {
		return fmt.Errorf("failed to start node: %v", err)
	}

	go n.acceptConnections(listener)
	return nil
}

// acceptConnections accepts incoming connections
func (n *Node) acceptConnections(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}

		go n.handleConnection(conn)
	}
}

// handleConnection handles an incoming connection
func (n *Node) handleConnection(conn net.Conn) {
	defer conn.Close()

	// 1. Receive handshake (peer's listening address)
	decoder := json.NewDecoder(conn)
	var handshake Message
	if err := decoder.Decode(&handshake); err != nil || handshake.Type != handshakeMessageType {
		return
	}
	var peerListenAddr string
	if err := json.Unmarshal(handshake.Payload, &peerListenAddr); err != nil {
		return
	}

	peer := &Peer{
		Address: peerListenAddr,
		Conn:    conn,
	}

	// 2. Deduplicate: if already connected, close new connection
	n.mu.Lock()
	if _, exists := n.Peers[peer.Address]; exists {
		n.mu.Unlock()
		return
	}
	n.Peers[peer.Address] = peer
	n.mu.Unlock()

	defer func() {
		n.mu.Lock()
		delete(n.Peers, peer.Address)
		n.mu.Unlock()
	}()

	// 3. Handle messages
	for {
		var msg Message
		if err := decoder.Decode(&msg); err != nil {
			return
		}
		n.handleMessage(msg, peer)
	}
}

// handleMessage handles incoming messages
func (n *Node) handleMessage(msg Message, peer *Peer) {
	switch msg.Type {
	case "new_block":
		var block blockchain.Block
		if err := json.Unmarshal(msg.Payload, &block); err != nil {
			return
		}
		n.handleNewBlock(block)
	case "get_blocks":
		n.sendBlocks(peer)
	case "get_peers":
		n.sendPeers(peer)
	}
}

// handleNewBlock handles a new block message
func (n *Node) handleNewBlock(block blockchain.Block) {
	if err := n.Blockchain.AddBlock(block); err != nil {
		return
	}

	// Broadcast the block to other peers
	n.broadcastBlock(block)
}

// broadcastBlock broadcasts a block to all peers
func (n *Node) broadcastBlock(block blockchain.Block) {
	payload, err := json.Marshal(block)
	if err != nil {
		return
	}

	msg := Message{
		Type:    "new_block",
		Payload: payload,
	}

	n.mu.RLock()
	defer n.mu.RUnlock()

	for _, peer := range n.Peers {
		n.sendMessage(msg, peer)
	}
}

// sendMessage sends a message to a peer
func (n *Node) sendMessage(msg Message, peer *Peer) {
	encoder := json.NewEncoder(peer.Conn)
	if err := encoder.Encode(msg); err != nil {
		return
	}
}

// Connect connects to a peer
func (n *Node) Connect(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to connect to peer: %v", err)
	}

	// 1. Send handshake (our listening address)
	encoder := json.NewEncoder(conn)
	payload, _ := json.Marshal(n.Address)
	handshake := Message{
		Type:    handshakeMessageType,
		Payload: payload,
	}
	if err := encoder.Encode(handshake); err != nil {
		conn.Close()
		return fmt.Errorf("failed to send handshake: %v", err)
	}

	peer := &Peer{
		Address: address,
		Conn:    conn,
	}

	// 2. Deduplicate: if already connected, close new connection
	n.mu.Lock()
	if _, exists := n.Peers[peer.Address]; exists {
		n.mu.Unlock()
		conn.Close()
		return nil
	}
	n.Peers[peer.Address] = peer
	n.mu.Unlock()

	go n.handleConnection(conn)
	return nil
}

// sendBlocks sends the blockchain to a peer
func (n *Node) sendBlocks(peer *Peer) {
	// TODO: Implement sending blocks to peer
}

// sendPeers sends the list of peers to a peer
func (n *Node) sendPeers(peer *Peer) {
	n.mu.RLock()
	peers := make([]string, 0, len(n.Peers))
	for addr := range n.Peers {
		peers = append(peers, addr)
	}
	n.mu.RUnlock()

	payload, err := json.Marshal(peers)
	if err != nil {
		return
	}

	msg := Message{
		Type:    "peers",
		Payload: payload,
	}

	n.sendMessage(msg, peer)
}

// BroadcastTransaction broadcasts a transaction to all peers
func (n *Node) BroadcastTransaction(tx blockchain.Transaction) {
	payload, err := json.Marshal(tx)
	if err != nil {
		return
	}

	msg := Message{
		Type:    "new_transaction",
		Payload: payload,
	}

	n.mu.RLock()
	defer n.mu.RUnlock()

	for _, peer := range n.Peers {
		n.sendMessage(msg, peer)
	}
}
