package network

import (
	"bytes"
	"context"
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

// Peer represents a network peer
type Peer struct {
	ID          string    `json:"id"`
	Address     string    `json:"address"`
	Port        int       `json:"port"`
	LastSeen    time.Time `json:"last_seen"`
	Latency     int64     `json:"latency"`
	Version     string    `json:"version"`
	IsActive    bool      `json:"is_active"`
	IsBootstrap bool      `json:"is_bootstrap"`
}

// NetworkMessage represents a message sent over the network
type NetworkMessage struct {
	Type      string          `json:"type"`
	From      string          `json:"from"`
	To        string          `json:"to,omitempty"`
	Data      json.RawMessage `json:"data"`
	Timestamp time.Time       `json:"timestamp"`
}

// NetworkManager handles network operations
type NetworkManager struct {
	mu             sync.RWMutex
	peers          map[string]*Peer
	bootstrapPeers []*Peer
	connections    map[string]net.Conn
	messageChan    chan *NetworkMessage
	stopChan       chan struct{}
	config         *NetworkConfig
	ctx            context.Context
	cancel         context.CancelFunc
}

// NetworkConfig represents network configuration
type NetworkConfig struct {
	ListenPort     int
	BootstrapPeers []string
	MaxPeers       int
	PingInterval   time.Duration
	DialTimeout    time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
}

// Node represents a P2P node
type Node struct {
	Config     *Config
	Blockchain *blockchain.Blockchain
	Peers      map[string]*Peer
	server     net.Listener
	mu         sync.RWMutex
}

// Config represents the node configuration
type Config struct {
	Address        string
	BlockType      blockchain.BlockType
	BootstrapPeers []string
}

// NewNode creates a new P2P node
func NewNode(config *Config) (*Node, error) {
	bc := blockchain.NewBlockchain()
	node := &Node{
		Config:     config,
		Blockchain: bc,
		Peers:      make(map[string]*Peer),
	}

	// Start listening for connections
	listener, err := net.Listen("tcp", config.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to start node: %v", err)
	}
	node.server = listener

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
		Conn:     conn,
		Address:  conn.RemoteAddr().String(),
		LastSeen: time.Now(),
		Node:     n,
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
		Address: address,
		Conn:    conn,
		Node:    n,
	}

	n.mu.Lock()
	n.Peers[address] = peer
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

	return gob.NewEncoder(peer.Conn).Encode(msg)
}

// receiveMessage receives a message from a peer
func (n *Node) receiveMessage(peer *Peer) (*Message, error) {
	var msg Message
	if err := gob.NewDecoder(peer.Conn).Decode(&msg); err != nil {
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
		return fmt.Errorf("unknown message type: %v", msg.Type)
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
	n.Peers[peer.Address] = peer
	n.mu.Unlock()

	// Request blocks
	return n.sendMessage(peer, GetBlocksMsg, nil)
}

func (n *Node) handleGetBlocks(peer *Peer, msg *Message) error {
	var blocks []*blockchain.Block
	if n.Config.BlockType == blockchain.GoldenBlock {
		for _, block := range n.Blockchain.GoldenBlocks {
			blocks = append(blocks, &block)
		}
	} else {
		for _, block := range n.Blockchain.SilverBlocks {
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
		if err := n.Blockchain.AddBlock(*block); err != nil {
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
		if block, err := n.Blockchain.GetBlock([]byte(hash)); err == nil {
			return n.sendMessage(peer, BlockMsg, block)
		}
		if tx, err := n.Blockchain.GetTransaction([]byte(hash)); err == nil {
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

	if err := n.Blockchain.AddTransaction(tx); err != nil {
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

	if err := n.Blockchain.AddBlock(*block); err != nil {
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
	for addr := range n.Peers {
		addrs = append(addrs, addr)
	}
	n.mu.RUnlock()

	return n.sendMessage(peer, AddrMsg, addrs)
}

func (n *Node) handlePing(peer *Peer, msg *Message) error {
	return n.sendMessage(peer, PongMsg, nil)
}

func (n *Node) handlePong(peer *Peer, msg *Message) error {
	peer.LastSeen = time.Now()
	return nil
}

// broadcastMessage broadcasts a message to all peers
func (n *Node) broadcastMessage(msgType MessageType, payload interface{}) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	for _, peer := range n.Peers {
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

	peers := make([]*Peer, 0, len(n.Peers))
	for _, peer := range n.Peers {
		peers = append(peers, peer)
	}

	return peers
}

// Close closes the node
func (n *Node) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Close all peer connections
	for _, peer := range n.Peers {
		peer.Conn.Close()
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
	payload, err := json.Marshal(p.Node.Config)
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
	if _, err := io.ReadFull(p.Conn, header); err != nil {
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
		if _, err := io.ReadFull(p.Conn, msg.Payload); err != nil {
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
	if _, err := p.Conn.Write(header); err != nil {
		return err
	}

	// Send message payload
	if len(msg.Payload) > 0 {
		if _, err := p.Conn.Write(msg.Payload); err != nil {
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

// ConnectToPeer connects to a peer at the given address
func (n *Node) ConnectToPeer(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to connect to peer: %v", err)
	}

	peer := &Peer{
		Address: address,
		Conn:    conn,
		Node:    n,
	}

	n.mu.Lock()
	n.Peers[address] = peer
	n.mu.Unlock()

	// Start handling messages from this peer
	go n.handlePeer(peer)

	// Send version message
	return peer.sendVersion()
}

// BroadcastMessage broadcasts a message to all connected peers
func (n *Node) BroadcastMessage(msg *Message) error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	for _, peer := range n.Peers {
		if err := n.sendMessage(peer, msg.Type, msg.Payload); err != nil {
			return fmt.Errorf("failed to send message to peer %s: %v", peer.Address, err)
		}
	}
	return nil
}

// Stop stops the node and closes all connections
func (n *Node) Stop() {
	if n.server != nil {
		n.server.Close()
	}

	// Close all peer connections
	for _, peer := range n.Peers {
		peer.Conn.Close()
	}
}

// handlePeer handles messages from a peer
func (n *Node) handlePeer(peer *Peer) {
	defer func() {
		peer.Conn.Close()
		n.mu.Lock()
		delete(n.Peers, peer.Address)
		n.mu.Unlock()
	}()

	for {
		msg, err := n.receiveMessage(peer)
		if err != nil {
			logger.Error("Failed to receive message", zap.Error(err))
			return
		}

		if handler, ok := peer.handlers[msg.Type]; ok {
			if err := handler(peer, msg.Payload); err != nil {
				logger.Error("Failed to handle message", zap.Error(err))
				return
			}
		}
	}
}

// NewNetworkManager creates a new network manager
func NewNetworkManager(config *NetworkConfig) *NetworkManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &NetworkManager{
		peers:          make(map[string]*Peer),
		bootstrapPeers: make([]*Peer, 0),
		connections:    make(map[string]net.Conn),
		messageChan:    make(chan *NetworkMessage, 100),
		stopChan:       make(chan struct{}),
		config:         config,
		ctx:            ctx,
		cancel:         cancel,
	}
}

// Start starts the network manager
func (nm *NetworkManager) Start() error {
	// Start listening for incoming connections
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", nm.config.ListenPort))
	if err != nil {
		return fmt.Errorf("failed to start listener: %v", err)
	}

	// Connect to bootstrap peers
	if err := nm.connectBootstrapPeers(); err != nil {
		return fmt.Errorf("failed to connect to bootstrap peers: %v", err)
	}

	// Start peer discovery
	go nm.startPeerDiscovery()

	// Start connection manager
	go nm.startConnectionManager()

	// Start message handler
	go nm.startMessageHandler()

	// Start listener
	go nm.startListener(listener)

	return nil
}

// Stop stops the network manager
func (nm *NetworkManager) Stop() {
	nm.cancel()
	close(nm.stopChan)

	// Close all connections
	nm.mu.Lock()
	defer nm.mu.Unlock()

	for _, conn := range nm.connections {
		conn.Close()
	}
}

// AddPeer adds a new peer
func (nm *NetworkManager) AddPeer(peer *Peer) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if len(nm.peers) >= nm.config.MaxPeers {
		return fmt.Errorf("maximum number of peers reached")
	}

	nm.peers[peer.ID] = peer
	return nil
}

// RemovePeer removes a peer
func (nm *NetworkManager) RemovePeer(peerID string) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if conn, exists := nm.connections[peerID]; exists {
		conn.Close()
		delete(nm.connections, peerID)
	}

	delete(nm.peers, peerID)
}

// GetPeers returns all known peers
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
	nm.mu.RLock()
	conn, exists := nm.connections[msg.To]
	nm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("peer %s not connected", msg.To)
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	conn.SetWriteDeadline(time.Now().Add(nm.config.WriteTimeout))
	_, err = conn.Write(data)
	return err
}

// Helper functions

func (nm *NetworkManager) connectBootstrapPeers() error {
	for _, addr := range nm.config.BootstrapPeers {
		peer := &Peer{
			Address:     addr,
			IsBootstrap: true,
		}
		if err := nm.connectToPeer(peer); err != nil {
			return fmt.Errorf("failed to connect to bootstrap peer %s: %v", addr, err)
		}
		nm.bootstrapPeers = append(nm.bootstrapPeers, peer)
	}
	return nil
}

func (nm *NetworkManager) connectToPeer(peer *Peer) error {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", peer.Address, peer.Port), nm.config.DialTimeout)
	if err != nil {
		return err
	}

	nm.mu.Lock()
	nm.connections[peer.ID] = conn
	nm.mu.Unlock()

	return nil
}

func (nm *NetworkManager) startPeerDiscovery() {
	ticker := time.NewTicker(nm.config.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			nm.discoverPeers()
		case <-nm.ctx.Done():
			return
		}
	}
}

func (nm *NetworkManager) discoverPeers() {
	nm.mu.RLock()
	peers := make([]*Peer, 0, len(nm.peers))
	for _, peer := range nm.peers {
		peers = append(peers, peer)
	}
	nm.mu.RUnlock()

	for _, peer := range peers {
		msg := &NetworkMessage{
			Type:      "get_peers",
			From:      "self",
			To:        peer.ID,
			Timestamp: time.Now(),
		}
		if err := nm.SendMessage(msg); err != nil {
			nm.RemovePeer(peer.ID)
		}
	}
}

func (nm *NetworkManager) startConnectionManager() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			nm.checkConnections()
		case <-nm.ctx.Done():
			return
		}
	}
}

func (nm *NetworkManager) checkConnections() {
	nm.mu.RLock()
	peers := make([]*Peer, 0, len(nm.peers))
	for _, peer := range nm.peers {
		peers = append(peers, peer)
	}
	nm.mu.RUnlock()

	for _, peer := range peers {
		if time.Since(peer.LastSeen) > nm.config.PingInterval*2 {
			nm.RemovePeer(peer.ID)
		}
	}
}

func (nm *NetworkManager) startMessageHandler() {
	for {
		select {
		case msg := <-nm.messageChan:
			nm.handleMessage(msg)
		case <-nm.ctx.Done():
			return
		}
	}
}

func (nm *NetworkManager) handleMessage(msg *NetworkMessage) {
	switch msg.Type {
	case "get_peers":
		nm.handleGetPeers(msg)
	case "peer_list":
		nm.handlePeerList(msg)
	case "ping":
		nm.handlePing(msg)
	case "pong":
		nm.handlePong(msg)
	}
}

func (nm *NetworkManager) startListener(listener net.Listener) {
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-nm.ctx.Done():
				return
			default:
				continue
			}
		}

		go nm.handleConnection(conn)
	}
}

func (nm *NetworkManager) handleConnection(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 4096)
	for {
		conn.SetReadDeadline(time.Now().Add(nm.config.ReadTimeout))
		n, err := conn.Read(buffer)
		if err != nil {
			return
		}

		var msg NetworkMessage
		if err := json.Unmarshal(buffer[:n], &msg); err != nil {
			continue
		}

		select {
		case nm.messageChan <- &msg:
		default:
			// Channel full, drop message
		}
	}
}

func (nm *NetworkManager) handleGetPeers(msg *NetworkMessage) {
	peers := nm.GetPeers()
	data, err := json.Marshal(peers)
	if err != nil {
		return
	}

	response := &NetworkMessage{
		Type:      "peer_list",
		From:      "self",
		To:        msg.From,
		Data:      data,
		Timestamp: time.Now(),
	}

	nm.SendMessage(response)
}

func (nm *NetworkManager) handlePeerList(msg *NetworkMessage) {
	var peers []*Peer
	if err := json.Unmarshal(msg.Data, &peers); err != nil {
		return
	}

	for _, peer := range peers {
		if peer.ID != "self" {
			nm.AddPeer(peer)
		}
	}
}

func (nm *NetworkManager) handlePing(msg *NetworkMessage) {
	response := &NetworkMessage{
		Type:      "pong",
		From:      "self",
		To:        msg.From,
		Timestamp: time.Now(),
	}

	nm.SendMessage(response)
}

func (nm *NetworkManager) handlePong(msg *NetworkMessage) {
	nm.mu.Lock()
	if peer, exists := nm.peers[msg.From]; exists {
		peer.LastSeen = time.Now()
		peer.Latency = time.Since(msg.Timestamp).Milliseconds()
	}
	nm.mu.Unlock()
}
