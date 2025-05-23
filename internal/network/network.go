package network

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/moroni/BYC/internal/blockchain"
	"github.com/moroni/BYC/internal/logger"
	"go.uber.org/zap"
)

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
	peer := &Peer{
		ID:       uuid.New().String(),
		Address:  conn.RemoteAddr().String(),
		LastSeen: time.Now(),
		conn:     conn,
		Node:     n,
		handlers: make(map[MessageType]MessageHandler),
	}
	n.mu.Lock()
	n.Peers[peer.ID] = peer
	n.mu.Unlock()
	go peer.handleMessages()
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := peer.sendPing(); err != nil {
				logger.Error("Failed to send ping", zap.Error(err))
				return
			}
		}
	}()
}

// connectToPeer connects to a peer
func (n *Node) connectToPeer(address string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		logger.Error("Failed to connect to peer", zap.String("address", address), zap.Error(err))
		return
	}

	peer := &Peer{
		Address:  address,
		LastSeen: time.Now(),
		conn:     conn,
		Node:     n,
		handlers: make(map[MessageType]MessageHandler),
	}

	n.mu.Lock()
	n.Peers[address] = peer
	n.mu.Unlock()

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
	msg := NetworkMessage{
		Type:    msgType,
		From:    n.Config.Address,
		To:      peer.Address,
		Payload: buf.Bytes(),
	}
	return gob.NewEncoder(peer.conn).Encode(msg)
}

// receiveMessage receives a message from a peer
func (n *Node) receiveMessage(peer *Peer) (*NetworkMessage, error) {
	var msg NetworkMessage
	if err := gob.NewDecoder(peer.conn).Decode(&msg); err != nil {
		return nil, fmt.Errorf("failed to decode message: %v", err)
	}
	return &msg, nil
}

// handleMessage handles a received message
func (n *Node) handleMessage(peer *Peer, msg *NetworkMessage) error {
	switch msg.Type {
	case MessageTypePing:
		return n.handlePing(peer, msg)
	case MessageTypePong:
		return n.handlePong(peer, msg)
	case MessageTypeGetBlocks:
		return n.handleGetBlocks(peer, msg)
	case MessageTypeBlocks:
		return n.handleBlocks(peer, msg)
	case MessageTypeGetData:
		return n.handleGetData(peer, msg)
	case MessageTypeInv:
		return n.handleInv(peer, msg)
	case MessageTypeTx:
		return n.handleTx(peer, msg)
	case MessageTypeBlock:
		return n.handleBlock(peer, msg)
	case MessageTypeAddr:
		return n.handleAddr(peer, msg)
	case MessageTypeGetAddr:
		return n.handleGetAddr(peer, msg)
	default:
		return fmt.Errorf("unknown message type: %v", msg.Type)
	}
}

// Message handlers
func (n *Node) handleVersion(peer *Peer, msg *NetworkMessage) error {
	var version int32
	if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&version); err != nil {
		return fmt.Errorf("failed to decode version: %v", err)
	}

	return n.sendMessage(peer, MessageTypeVerAck, nil)
}

func (n *Node) handleVerAck(peer *Peer, msg *NetworkMessage) error {
	n.mu.Lock()
	n.Peers[peer.Address] = peer
	n.mu.Unlock()

	// Request blocks
	return n.sendMessage(peer, MessageTypeGetBlocks, nil)
}

func (n *Node) handleGetBlocks(peer *Peer, msg *NetworkMessage) error {
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

	return n.sendMessage(peer, MessageTypeBlocks, blocks)
}

func (n *Node) handleBlocks(peer *Peer, msg *NetworkMessage) error {
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

func (n *Node) handleGetData(peer *Peer, msg *NetworkMessage) error {
	var inv []string
	if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&inv); err != nil {
		return fmt.Errorf("failed to decode inventory: %v", err)
	}

	for _, hash := range inv {
		if block, err := n.Blockchain.GetBlock([]byte(hash)); err == nil {
			return n.sendMessage(peer, MessageTypeBlock, block)
		}
		if tx, err := n.Blockchain.GetTransaction([]byte(hash)); err == nil {
			return n.sendMessage(peer, MessageTypeTx, tx)
		}
	}

	return nil
}

func (n *Node) handleInv(peer *Peer, msg *NetworkMessage) error {
	var inv []string
	if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&inv); err != nil {
		return fmt.Errorf("failed to decode inventory: %v", err)
	}

	return n.sendMessage(peer, MessageTypeGetData, inv)
}

func (n *Node) handleTx(peer *Peer, msg *NetworkMessage) error {
	var tx *blockchain.Transaction
	if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&tx); err != nil {
		return fmt.Errorf("failed to decode transaction: %v", err)
	}

	if err := n.Blockchain.AddTransaction(tx); err != nil {
		return fmt.Errorf("failed to add transaction: %v", err)
	}

	// Broadcast transaction to other peers
	n.broadcastMessage(MessageTypeTx, tx)
	return nil
}

func (n *Node) handleBlock(peer *Peer, msg *NetworkMessage) error {
	var block *blockchain.Block
	if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&block); err != nil {
		return fmt.Errorf("failed to decode block: %v", err)
	}

	if err := n.Blockchain.AddBlock(*block); err != nil {
		return fmt.Errorf("failed to add block: %v", err)
	}

	// Broadcast block to other peers
	n.broadcastMessage(MessageTypeBlock, block)
	return nil
}

func (n *Node) handleAddr(peer *Peer, msg *NetworkMessage) error {
	var addrs []string
	if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&addrs); err != nil {
		return fmt.Errorf("failed to decode addresses: %v", err)
	}

	for _, addr := range addrs {
		go n.connectToPeer(addr)
	}

	return nil
}

func (n *Node) handleGetAddr(peer *Peer, msg *NetworkMessage) error {
	var addrs []string
	n.mu.RLock()
	for addr := range n.Peers {
		addrs = append(addrs, addr)
	}
	n.mu.RUnlock()

	return n.sendMessage(peer, MessageTypeAddr, addrs)
}

func (n *Node) handlePing(peer *Peer, msg *NetworkMessage) error {
	return n.sendMessage(peer, MessageTypePong, nil)
}

func (n *Node) handlePong(peer *Peer, msg *NetworkMessage) error {
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
		conn := peer.conn
		conn.Close()
	}

	// Close server
	return n.server.Close()
}

// registerHandlers registers message handlers
func (p *Peer) registerHandlers() {
	p.handlers = map[MessageType]MessageHandler{
		MessageTypePing:      handlePing,
		MessageTypePong:      handlePong,
		MessageTypeGetBlocks: handleGetBlocks,
		MessageTypeBlocks:    handleBlocks,
		MessageTypeGetData:   handleGetData,
		MessageTypeBlock:     handleBlock,
		MessageTypeTx:        handleTx,
		MessageTypeInv:       handleInv,
		MessageTypeVerAck:    handleVerAck,
	}
}

// handleMessages handles incoming messages
func (p *Peer) handleMessages() {
	defer p.conn.Close()
	for {
		message, err := p.receiveMessage()
		if err != nil {
			if err == io.EOF {
				logger.Info("Peer disconnected", zap.String("peer", p.ID))
			} else {
				logger.Error("Error reading message", zap.Error(err))
			}
			return
		}

		handler, ok := p.handlers[message.Type]
		if !ok {
			logger.Error("Unknown message type", zap.String("type", string(message.Type)))
			continue
		}

		if err := handler(p, message.Payload); err != nil {
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

	msg := NetworkMessage{
		Type:      MessageTypeVersion,
		Payload:   payload,
		Timestamp: time.Now(),
	}
	return p.sendMessage(msg)
}

// receiveMessage receives a message from the peer
func (p *Peer) receiveMessage() (*NetworkMessage, error) {
	var msg NetworkMessage
	if err := gob.NewDecoder(p.conn).Decode(&msg); err != nil {
		return nil, fmt.Errorf("failed to decode message: %v", err)
	}
	return &msg, nil
}

// sendMessage sends a message to the peer
func (p *Peer) sendMessage(msg NetworkMessage) error {
	return gob.NewEncoder(p.conn).Encode(msg)
}

func handlePing(p *Peer, payload []byte) error      { return nil }
func handlePong(p *Peer, payload []byte) error      { return nil }
func handleGetBlocks(p *Peer, payload []byte) error { return nil }
func handleBlocks(p *Peer, payload []byte) error    { return nil }
func handleGetData(p *Peer, payload []byte) error   { return nil }
func handleBlock(p *Peer, payload []byte) error     { return nil }
func handleTx(p *Peer, payload []byte) error        { return nil }
func handleInv(p *Peer, payload []byte) error       { return nil }
func handleVerAck(p *Peer, payload []byte) error    { return nil }

// ConnectToPeer connects to a peer at the given address
func (n *Node) ConnectToPeer(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to connect to peer: %v", err)
	}

	peer := &Peer{
		Address:  address,
		LastSeen: time.Now(),
		conn:     conn,
		Node:     n,
		handlers: make(map[MessageType]MessageHandler),
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
func (n *Node) BroadcastMessage(msg NetworkMessage) error {
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
		if peer.conn != nil {
			peer.conn.Close()
		}
	}
}

// handlePeer handles messages from a peer
func (n *Node) handlePeer(peer *Peer) {
	defer func() {
		if peer.conn != nil {
			peer.conn.Close()
		}
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
		peer := Peer{
			Address:  addr,
			LastSeen: time.Now(),
			conn:     nil,
			Node:     nil,
			handlers: make(map[MessageType]MessageHandler),
		}
		if err := nm.connectToPeer(addr); err != nil {
			return fmt.Errorf("failed to connect to bootstrap peer %s: %v", addr, err)
		}
		nm.bootstrapPeers = append(nm.bootstrapPeers, &peer)
	}
	return nil
}

func (nm *NetworkManager) connectToPeer(address string) error {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", address, 0), nm.config.DialTimeout)
	if err != nil {
		return err
	}

	nm.mu.Lock()
	nm.connections[address] = conn
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
			Type:      MessageTypeGetAddr,
			From:      nm.config.NodeID,
			To:        peer.ID,
			Payload:   []byte("discovery"),
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

func (nm *NetworkManager) handleMessage(msg *NetworkMessage) error {
	switch msg.Type {
	case "discovery":
		return nm.handlePeerDiscovery(msg)
	case "peerlist":
		return nm.handlePeerList(msg)
	case "ping":
		return nm.handlePing(msg)
	case "pong":
		return nm.handlePong(msg)
	case "block":
		return nm.handleBlock(msg)
	case "transaction":
		return nm.handleTransaction(msg)
	default:
		return fmt.Errorf("unknown message type: %d", msg.Type)
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

func (nm *NetworkManager) handlePeerDiscovery(msg *NetworkMessage) error {
	// TODO: Implement peer discovery
	return nil
}

func (nm *NetworkManager) handlePeerList(msg *NetworkMessage) error {
	// TODO: Implement peer list handling
	return nil
}

func (nm *NetworkManager) handlePing(msg *NetworkMessage) error {
	pong := &NetworkMessage{
		Type:      MessageTypePong,
		From:      nm.config.NodeID,
		To:        msg.From,
		Payload:   []byte("pong"),
		Timestamp: time.Now(),
	}
	return nm.SendMessage(pong)
}

func (nm *NetworkManager) handlePong(msg *NetworkMessage) error {
	nm.mu.Lock()
	if peer, ok := nm.peers[msg.From]; ok {
		peer.LastSeen = time.Now()
	}
	nm.mu.Unlock()
	return nil
}

func (nm *NetworkManager) handleBlock(msg *NetworkMessage) error {
	// TODO: Implement block handling
	return nil
}

func (nm *NetworkManager) handleTransaction(msg *NetworkMessage) error {
	// TODO: Implement transaction handling
	return nil
}

// HasReceivedPong checks if a pong message was received from a peer
func (nm *NetworkManager) HasReceivedPong(peerAddr string) bool {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	peer, exists := nm.peers[peerAddr]
	if !exists {
		return false
	}

	// Check if we received a pong within the ping timeout
	return time.Since(peer.LastSeen) < nm.config.PingTimeout
}

// ConnectToPeer connects to a peer at the given address
func (nm *NetworkManager) ConnectToPeer(address string) error {
	return nm.connectToPeer(address)
}

func (p *Peer) sendPing() error {
	msg := NetworkMessage{
		Type:    MessageTypePing,
		Payload: []byte("ping"),
	}
	return p.sendMessage(msg)
}
