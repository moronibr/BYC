package network

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"encoding/binary"

	"crypto/sha256"

	"github.com/byc/internal/blockchain"
	"github.com/byc/internal/logger"
	"go.uber.org/zap"
)

// MessageType represents the type of network message
type MessageType int

const (
	// Message types
	VersionMsg MessageType = iota
	VerAckMsg
	GetBlocksMsg
	BlocksMsg
	GetDataMsg
	InvMsg
	TxMsg
	BlockMsg
	AddrMsg
	GetAddrMsg
	PingMsg
	PongMsg
	FilterLoadMsg
	MerkleBlockMsg
	RejectMsg
	NotFoundMsg
)

// Message represents a network message
type Message struct {
	Type    MessageType
	Payload []byte
}

// Node represents a P2P node
type Node struct {
	Config     *Config
	blockchain *blockchain.Blockchain
	peers      map[string]*Peer
	server     net.Listener
	mu         sync.RWMutex
}

// Peer represents a connected peer
type Peer struct {
	address  string
	conn     net.Conn
	node     *Node
	handlers map[MessageType]MessageHandler
	lastSeen time.Time
}

// Config represents the node configuration
type Config struct {
	Address        string
	BlockType      blockchain.BlockType
	BootstrapPeers []string
}

// NewNode creates a new P2P node
func NewNode(config *Config, bc *blockchain.Blockchain) (*Node, error) {
	node := &Node{
		Config:     config,
		blockchain: bc,
		peers:      make(map[string]*Peer),
	}

	// Start the server
	server, err := net.Listen("tcp", config.Address)
	if err != nil {
		return nil, err
	}
	node.server = server

	// Connect to bootstrap peers
	for _, peer := range config.BootstrapPeers {
		go node.ConnectToPeer(peer)
	}

	// Start accepting connections
	go node.acceptConnections()

	return node, nil
}

// acceptConnections accepts incoming connections
func (n *Node) acceptConnections() {
	for {
		conn, err := n.server.Accept()
		if err != nil {
			logger.Error("Failed to accept connection", zap.Error(err))
			continue
		}

		go n.handleConnection(conn)
	}
}

// handleConnection handles a new connection
func (n *Node) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Create new peer
	peer := &Peer{
		conn:     conn,
		address:  conn.RemoteAddr().String(),
		lastSeen: time.Now(),
		node:     n,
		handlers: make(map[MessageType]MessageHandler),
	}

	// Register message handlers
	peer.registerHandlers()

	// Start handling messages
	go peer.handleMessages()

	// Send version message
	peer.sendVersion()
}

// connectToPeer connects to a peer
func (n *Node) connectToPeer(address string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		logger.Error("Failed to connect to peer", zap.String("address", address), zap.Error(err))
		return
	}

	peer := &Peer{
		address:  address,
		conn:     conn,
		node:     n,
		handlers: make(map[MessageType]MessageHandler),
	}

	n.mu.Lock()
	n.peers[address] = peer
	n.mu.Unlock()

	// Register message handlers
	peer.registerHandlers()

	// Start handling messages
	go peer.handleMessages()

	// Send version message
	peer.sendVersion()
}

// sendMessage sends a message to a peer
func (n *Node) sendMessage(peer *Peer, msgType MessageType, payload interface{}) error {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(payload); err != nil {
		return fmt.Errorf("failed to encode message: %v", err)
	}

	msg := Message{
		Type:    msgType,
		Payload: buf.Bytes(),
	}

	return gob.NewEncoder(peer.conn).Encode(msg)
}

// receiveMessage receives a message from a peer
func (n *Node) receiveMessage(peer *Peer) (*Message, error) {
	var msg Message
	if err := gob.NewDecoder(peer.conn).Decode(&msg); err != nil {
		return nil, fmt.Errorf("failed to decode message: %v", err)
	}
	return &msg, nil
}

// handleMessage handles a received message
func (n *Node) handleMessage(peer *Peer, msg *Message) error {
	switch msg.Type {
	case VersionMsg:
		return n.handleVersion(peer, msg)
	case VerAckMsg:
		return n.handleVerAck(peer, msg)
	case GetBlocksMsg:
		return n.handleGetBlocks(peer, msg)
	case BlocksMsg:
		return n.handleBlocks(peer, msg)
	case GetDataMsg:
		return n.handleGetData(peer, msg)
	case InvMsg:
		return n.handleInv(peer, msg)
	case TxMsg:
		return n.handleTx(peer, msg)
	case BlockMsg:
		return n.handleBlock(peer, msg)
	case AddrMsg:
		return n.handleAddr(peer, msg)
	case GetAddrMsg:
		return n.handleGetAddr(peer, msg)
	case PingMsg:
		return n.handlePing(peer, msg)
	case PongMsg:
		return n.handlePong(peer, msg)
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

// Message handlers
func (n *Node) handleVersion(peer *Peer, msg *Message) error {
	var version int32
	if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&version); err != nil {
		return fmt.Errorf("failed to decode version: %v", err)
	}

	return n.sendMessage(peer, VerAckMsg, nil)
}

func (n *Node) handleVerAck(peer *Peer, msg *Message) error {
	n.mu.Lock()
	n.peers[peer.address] = peer
	n.mu.Unlock()

	// Request blocks
	return n.sendMessage(peer, GetBlocksMsg, nil)
}

func (n *Node) handleGetBlocks(peer *Peer, msg *Message) error {
	var blocks []*blockchain.Block
	if n.Config.BlockType == blockchain.GoldenBlock {
		for _, block := range n.blockchain.GoldenBlocks {
			blocks = append(blocks, &block)
		}
	} else {
		for _, block := range n.blockchain.SilverBlocks {
			blocks = append(blocks, &block)
		}
	}

	return n.sendMessage(peer, BlocksMsg, blocks)
}

func (n *Node) handleBlocks(peer *Peer, msg *Message) error {
	var blocks []*blockchain.Block
	if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&blocks); err != nil {
		return fmt.Errorf("failed to decode blocks: %v", err)
	}

	for _, block := range blocks {
		if err := n.blockchain.AddBlock(*block); err != nil {
			logger.Error("Failed to add block", zap.Error(err))
		}
	}

	return nil
}

func (n *Node) handleGetData(peer *Peer, msg *Message) error {
	var inv []string
	if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&inv); err != nil {
		return fmt.Errorf("failed to decode inventory: %v", err)
	}

	for _, hash := range inv {
		if block, err := n.blockchain.GetBlock([]byte(hash)); err == nil {
			return n.sendMessage(peer, BlockMsg, block)
		}
		if tx, err := n.blockchain.GetTransaction([]byte(hash)); err == nil {
			return n.sendMessage(peer, TxMsg, tx)
		}
	}

	return nil
}

func (n *Node) handleInv(peer *Peer, msg *Message) error {
	var inv []string
	if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&inv); err != nil {
		return fmt.Errorf("failed to decode inventory: %v", err)
	}

	return n.sendMessage(peer, GetDataMsg, inv)
}

func (n *Node) handleTx(peer *Peer, msg *Message) error {
	var tx *blockchain.Transaction
	if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&tx); err != nil {
		return fmt.Errorf("failed to decode transaction: %v", err)
	}

	if err := n.blockchain.AddTransaction(tx); err != nil {
		return fmt.Errorf("failed to add transaction: %v", err)
	}

	// Broadcast transaction to other peers
	n.broadcastMessage(TxMsg, tx)
	return nil
}

func (n *Node) handleBlock(peer *Peer, msg *Message) error {
	var block *blockchain.Block
	if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&block); err != nil {
		return fmt.Errorf("failed to decode block: %v", err)
	}

	if err := n.blockchain.AddBlock(*block); err != nil {
		return fmt.Errorf("failed to add block: %v", err)
	}

	// Broadcast block to other peers
	n.broadcastMessage(BlockMsg, block)
	return nil
}

func (n *Node) handleAddr(peer *Peer, msg *Message) error {
	var addrs []string
	if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&addrs); err != nil {
		return fmt.Errorf("failed to decode addresses: %v", err)
	}

	for _, addr := range addrs {
		go n.connectToPeer(addr)
	}

	return nil
}

func (n *Node) handleGetAddr(peer *Peer, msg *Message) error {
	var addrs []string
	n.mu.RLock()
	for addr := range n.peers {
		addrs = append(addrs, addr)
	}
	n.mu.RUnlock()

	return n.sendMessage(peer, AddrMsg, addrs)
}

func (n *Node) handlePing(peer *Peer, msg *Message) error {
	return n.sendMessage(peer, PongMsg, nil)
}

func (n *Node) handlePong(peer *Peer, msg *Message) error {
	peer.lastSeen = time.Now()
	return nil
}

// broadcastMessage broadcasts a message to all peers
func (n *Node) broadcastMessage(msgType MessageType, payload interface{}) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	for _, peer := range n.peers {
		if err := n.sendMessage(peer, msgType, payload); err != nil {
			logger.Error("Failed to broadcast message", zap.Error(err))
		}
	}
}

// StartMining starts mining
func (n *Node) StartMining(coinType blockchain.CoinType) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.Config.BlockType != "" {
		return fmt.Errorf("already mining")
	}

	n.Config.BlockType = blockchain.GetBlockType(coinType)
	return nil
}

// StopMining stops mining
func (n *Node) StopMining() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.Config.BlockType = ""
}

// GetPeers returns the list of connected peers
func (n *Node) GetPeers() []*Peer {
	n.mu.RLock()
	defer n.mu.RUnlock()

	peers := make([]*Peer, 0, len(n.peers))
	for _, peer := range n.peers {
		peers = append(peers, peer)
	}

	return peers
}

// Close closes the node
func (n *Node) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Close all peer connections
	for _, peer := range n.peers {
		peer.conn.Close()
	}

	// Close server
	return n.server.Close()
}

// MessageHandler is a function that handles a message
type MessageHandler func(*Peer, []byte) error

// registerHandlers registers message handlers
func (p *Peer) registerHandlers() {
	p.handlers = map[MessageType]MessageHandler{
		VersionMsg:     handleVersion,
		VerAckMsg:      handleVerAck,
		GetBlocksMsg:   handleGetBlocks,
		BlocksMsg:      handleBlocks,
		GetDataMsg:     handleGetData,
		BlockMsg:       handleBlock,
		TxMsg:          handleTx,
		InvMsg:         handleInv,
		NotFoundMsg:    handleNotFound,
		PingMsg:        handlePing,
		PongMsg:        handlePong,
		FilterLoadMsg:  handleFilterLoad,
		MerkleBlockMsg: handleMerkleBlock,
		RejectMsg:      handleReject,
	}
}

// handleMessages handles incoming messages
func (p *Peer) handleMessages() {
	for {
		msg, err := p.receiveMessage()
		if err != nil {
			logger.Error("Failed to receive message", zap.Error(err))
			return
		}

		handler, ok := p.handlers[msg.Type]
		if !ok {
			logger.Error("Unknown message type", zap.Int("type", int(msg.Type)))
			continue
		}

		if err := handler(p, msg.Payload); err != nil {
			logger.Error("Failed to handle message", zap.Error(err))
			return
		}
	}
}

// sendVersion sends a version message
func (p *Peer) sendVersion() error {
	payload, err := json.Marshal(p.node.Config)
	if err != nil {
		return err
	}

	msg := &Message{
		Type:    VersionMsg,
		Payload: payload,
	}
	return p.sendMessage(msg)
}

// receiveMessage receives a message from the peer
func (p *Peer) receiveMessage() (*Message, error) {
	// Read message header
	header := make([]byte, 24)
	if _, err := io.ReadFull(p.conn, header); err != nil {
		return nil, err
	}

	// Parse message header
	msg := &Message{
		Type: MessageType(binary.LittleEndian.Uint32(header[0:4])),
	}

	// Read message payload
	length := binary.LittleEndian.Uint32(header[4:8])
	if length > 0 {
		msg.Payload = make([]byte, length)
		if _, err := io.ReadFull(p.conn, msg.Payload); err != nil {
			return nil, err
		}

		// Verify checksum
		checksum := sha256.Sum256(msg.Payload)
		checksum = sha256.Sum256(checksum[:])
		if binary.LittleEndian.Uint32(checksum[:4]) != binary.LittleEndian.Uint32(header[8:12]) {
			return nil, fmt.Errorf("invalid checksum")
		}
	}

	return msg, nil
}

// sendMessage sends a message to the peer
func (p *Peer) sendMessage(msg *Message) error {
	// Create message header
	header := make([]byte, 24)
	binary.LittleEndian.PutUint32(header[0:4], uint32(msg.Type))

	// Calculate checksum
	checksum := sha256.Sum256(msg.Payload)
	checksum = sha256.Sum256(checksum[:])
	binary.LittleEndian.PutUint32(header[8:12], binary.LittleEndian.Uint32(checksum[:4]))

	// Set message length
	binary.LittleEndian.PutUint32(header[4:8], uint32(len(msg.Payload)))

	// Send message header
	if _, err := p.conn.Write(header); err != nil {
		return err
	}

	// Send message payload
	if len(msg.Payload) > 0 {
		if _, err := p.conn.Write(msg.Payload); err != nil {
			return err
		}
	}

	return nil
}

func handleVersion(p *Peer, payload []byte) error     { return nil }
func handleVerAck(p *Peer, payload []byte) error      { return nil }
func handleGetBlocks(p *Peer, payload []byte) error   { return nil }
func handleBlocks(p *Peer, payload []byte) error      { return nil }
func handleGetData(p *Peer, payload []byte) error     { return nil }
func handleBlock(p *Peer, payload []byte) error       { return nil }
func handleTx(p *Peer, payload []byte) error          { return nil }
func handleInv(p *Peer, payload []byte) error         { return nil }
func handleNotFound(p *Peer, payload []byte) error    { return nil }
func handlePing(p *Peer, payload []byte) error        { return nil }
func handlePong(p *Peer, payload []byte) error        { return nil }
func handleFilterLoad(p *Peer, payload []byte) error  { return nil }
func handleMerkleBlock(p *Peer, payload []byte) error { return nil }
func handleReject(p *Peer, payload []byte) error      { return nil }

// ConnectToPeer connects to a peer
func (n *Node) ConnectToPeer(address string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		logger.Error("Failed to connect to peer", zap.String("address", address), zap.Error(err))
		return
	}

	peer := &Peer{
		address:  address,
		conn:     conn,
		node:     n,
		handlers: make(map[MessageType]MessageHandler),
	}

	n.mu.Lock()
	n.peers[address] = peer
	n.mu.Unlock()

	// Register message handlers
	peer.registerHandlers()

	// Start handling messages
	go peer.handleMessages()

	// Send version message
	peer.sendVersion()
}
