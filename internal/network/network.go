package network

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"

	"byc/internal/blockchain"
	"byc/internal/logger"
	"byc/internal/utils"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// NewNode creates a new P2P node
func NewNode(config *Config) (*Node, error) {
	// Find available port for P2P server
	p2pAddress, err := utils.FindAvailableAddress(config.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to find available port for P2P server: %v", err)
	}
	config.Address = p2pAddress

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

	// Start accepting connections in a goroutine
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
					logger.Error("Temporary error accepting connection", zap.Error(err))
					time.Sleep(time.Second)
					continue
				}
				if err != io.EOF {
					logger.Error("Failed to accept connection", zap.Error(err))
				}
				return
			}

			go node.handleConnection(conn)
		}
	}()

	logger.Info("P2P server started", zap.String("address", config.Address))
	return node, nil
}

// Stop stops the node and closes all connections
func (n *Node) Stop() error {
	if n.server != nil {
		if err := n.server.Close(); err != nil {
			logger.Error("Error closing server", zap.Error(err))
		}
		n.server = nil
	}

	// Close all peer connections
	n.mu.Lock()
	defer n.mu.Unlock()
	for _, peer := range n.Peers {
		if peer.conn != nil {
			peer.conn.Close()
		}
	}
	n.Peers = make(map[string]*Peer)

	return nil
}

// handleConnection handles a new connection
func (n *Node) handleConnection(conn net.Conn) {
	defer conn.Close()

	peer := NewPeer(uuid.New().String(), conn.RemoteAddr().String(), 0)
	peer.conn = conn
	peer.Node = n
	peer.handlers = make(map[MessageType]MessageHandler)

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

	if err := n.Blockchain.AddTransaction(*tx); err != nil {
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

	if n.isMining {
		return fmt.Errorf("already mining")
	}

	n.isMining = true
	n.Config.BlockType = blockchain.GetBlockType(coinType)

	// Start mining in a goroutine
	go n.mineBlocks()

	return nil
}

// StopMining stops mining
func (n *Node) StopMining() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.isMining = false
	n.Config.BlockType = ""
}

// mineBlocks continuously mines new blocks
func (n *Node) mineBlocks() {
	for {
		n.mu.RLock()
		if !n.isMining {
			n.mu.RUnlock()
			return
		}
		blockType := n.Config.BlockType
		n.mu.RUnlock()

		// Get pending transactions
		pendingTxs := n.Blockchain.PendingTxs

		// Determine coin type based on block type
		var coinType blockchain.CoinType
		if blockType == blockchain.GoldenBlock {
			coinType = blockchain.Leah
		} else {
			coinType = blockchain.Senum
		}

		// Mine the block
		block, err := n.Blockchain.MineBlock(pendingTxs, blockType, coinType)
		if err != nil {
			logger.Error("Failed to mine block", zap.Error(err))
			continue
		}

		// Add the mined block
		if err := n.Blockchain.AddBlock(block); err != nil {
			logger.Error("Failed to add mined block", zap.Error(err))
			continue
		}

		// Broadcast the new block to peers
		n.broadcastMessage(MessageTypeBlock, &block)
	}
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

// sendPing sends a ping message to a peer
func (p *Peer) sendPing() error {
	msg := NetworkMessage{
		Type:    MessageTypePing,
		Payload: []byte("ping"),
	}
	return p.sendMessage(msg)
}

// GetAddress returns the node's address
func (n *Node) GetAddress() string {
	return n.Config.Address
}

// GetPeerAddresses returns a list of peer addresses
func (n *Node) GetPeerAddresses() []string {
	n.mu.RLock()
	defer n.mu.RUnlock()

	addresses := make([]string, 0, len(n.Peers))
	for _, peer := range n.Peers {
		addresses = append(addresses, peer.Address)
	}
	return addresses
}

// DisconnectPeer disconnects from a peer
func (n *Node) DisconnectPeer(address string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	peer, exists := n.Peers[address]
	if !exists {
		return fmt.Errorf("peer %s not found", address)
	}

	if peer.conn != nil {
		peer.conn.Close()
	}
	delete(n.Peers, address)
	return nil
}
