package network

import (
	"encoding/json"
	"fmt"
	"time"

	"byc/internal/blockchain"
	"byc/internal/logger"

	"go.uber.org/zap"
)

// ProtocolHandler handles P2P protocol messages
type ProtocolHandler struct {
	node *Node
}

// NewProtocolHandler creates a new protocol handler
func NewProtocolHandler(node *Node) *ProtocolHandler {
	return &ProtocolHandler{
		node: node,
	}
}

// HandleMessage handles incoming protocol messages
func (ph *ProtocolHandler) HandleMessage(peer *Peer, msg *NetworkMessage) error {
	switch msg.Type {
	case MessageTypeVersion:
		return ph.handleVersion(peer, msg)
	case MessageTypeVerAck:
		return ph.handleVerAck(peer, msg)
	case MessageTypePing:
		return ph.handlePing(peer, msg)
	case MessageTypePong:
		return ph.handlePong(peer, msg)
	case MessageTypeGetBlocks:
		return ph.handleGetBlocks(peer, msg)
	case MessageTypeBlocks:
		return ph.handleBlocks(peer, msg)
	case MessageTypeGetData:
		return ph.handleGetData(peer, msg)
	case MessageTypeInv:
		return ph.handleInv(peer, msg)
	case MessageTypeBlock:
		return ph.handleBlock(peer, msg)
	case MessageTypeTx:
		return ph.handleTx(peer, msg)
	case MessageTypeAddr:
		return ph.handleAddr(peer, msg)
	case MessageTypeGetAddr:
		return ph.handleGetAddr(peer, msg)
	case MessageTypeGetHeight:
		return ph.handleGetHeight(peer, msg)
	default:
		logger.Warn("Unknown message type", zap.String("type", string(msg.Type)))
		return nil
	}
}

// handleVersion handles version messages
func (ph *ProtocolHandler) handleVersion(peer *Peer, msg *NetworkMessage) error {
	var version VersionMessage
	if err := json.Unmarshal(msg.Payload, &version); err != nil {
		return fmt.Errorf("failed to unmarshal version message: %v", err)
	}

	peer.mu.Lock()
	peer.Version = version.Version
	peer.Height = version.Height
	peer.mu.Unlock()

	// Send verack
	verack := VerAckMessage{
		Accepted: true,
	}
	payload, _ := json.Marshal(verack)
	response := NetworkMessage{
		Type:      MessageTypeVerAck,
		From:      ph.node.Config.Address,
		To:        peer.Address,
		Payload:   payload,
		Timestamp: time.Now(),
	}
	return peer.sendMessage(response)
}

// handleVerAck handles verack messages
func (ph *ProtocolHandler) handleVerAck(peer *Peer, msg *NetworkMessage) error {
	var verack VerAckMessage
	if err := json.Unmarshal(msg.Payload, &verack); err != nil {
		return fmt.Errorf("failed to unmarshal verack message: %v", err)
	}

	if verack.Accepted {
		peer.mu.Lock()
		peer.IsActive = true
		peer.mu.Unlock()
		logger.Info("Peer connection established", zap.String("peer", peer.Address))
	} else {
		logger.Warn("Peer rejected connection", zap.String("peer", peer.Address))
	}

	return nil
}

// handlePing handles ping messages
func (ph *ProtocolHandler) handlePing(peer *Peer, msg *NetworkMessage) error {
	// Send pong response
	pong := PongMessage{
		Timestamp: time.Now(),
	}
	payload, _ := json.Marshal(pong)
	response := NetworkMessage{
		Type:      MessageTypePong,
		From:      ph.node.Config.Address,
		To:        peer.Address,
		Payload:   payload,
		Timestamp: time.Now(),
	}
	return peer.sendMessage(response)
}

// handlePong handles pong messages
func (ph *ProtocolHandler) handlePong(peer *Peer, msg *NetworkMessage) error {
	var pong PongMessage
	if err := json.Unmarshal(msg.Payload, &pong); err != nil {
		return fmt.Errorf("failed to unmarshal pong message: %v", err)
	}

	// Update peer latency
	latency := time.Since(pong.Timestamp)
	peer.mu.Lock()
	peer.Latency = latency
	peer.LastSeen = time.Now()
	peer.mu.Unlock()

	return nil
}

// handleGetBlocks handles getblocks messages
func (ph *ProtocolHandler) handleGetBlocks(peer *Peer, msg *NetworkMessage) error {
	var getBlocks GetBlocksMessage
	if err := json.Unmarshal(msg.Payload, &getBlocks); err != nil {
		return fmt.Errorf("failed to unmarshal getblocks message: %v", err)
	}

	// Get blocks from blockchain
	ph.node.Blockchain.LockForRead()
	var blocks []blockchain.Block

	// Determine which chain to use based on block type
	if getBlocks.BlockType == blockchain.GoldenBlock {
		blocks = ph.node.Blockchain.GoldenBlocks
	} else {
		blocks = ph.node.Blockchain.SilverBlocks
	}
	ph.node.Blockchain.UnlockForRead()

	// Send blocks response
	blocksMsg := BlocksMessage{
		Blocks:    blocks,
		BlockType: getBlocks.BlockType,
	}
	payload, _ := json.Marshal(blocksMsg)
	response := NetworkMessage{
		Type:      MessageTypeBlocks,
		From:      ph.node.Config.Address,
		To:        peer.Address,
		Payload:   payload,
		Timestamp: time.Now(),
	}
	return peer.sendMessage(response)
}

// handleBlocks handles blocks messages
func (ph *ProtocolHandler) handleBlocks(peer *Peer, msg *NetworkMessage) error {
	var blocksMsg BlocksMessage
	if err := json.Unmarshal(msg.Payload, &blocksMsg); err != nil {
		return fmt.Errorf("failed to unmarshal blocks message: %v", err)
	}

	// Process received blocks
	ph.node.Blockchain.LockForWrite()
	defer ph.node.Blockchain.UnlockForWrite()

	for _, block := range blocksMsg.Blocks {
		if blocksMsg.BlockType == blockchain.GoldenBlock {
			ph.node.Blockchain.GoldenBlocks = append(ph.node.Blockchain.GoldenBlocks, block)
		} else {
			ph.node.Blockchain.SilverBlocks = append(ph.node.Blockchain.SilverBlocks, block)
		}
	}

	logger.Info("Received blocks from peer",
		zap.String("peer", peer.Address),
		zap.Int("count", len(blocksMsg.Blocks)),
		zap.String("type", string(blocksMsg.BlockType)))

	return nil
}

// handleGetData handles getdata messages
func (ph *ProtocolHandler) handleGetData(peer *Peer, msg *NetworkMessage) error {
	var getData GetDataMessage
	if err := json.Unmarshal(msg.Payload, &getData); err != nil {
		return fmt.Errorf("failed to unmarshal getdata message: %v", err)
	}

	// Send requested data
	switch getData.DataType {
	case "block":
		return ph.sendBlock(peer, getData.Hash)
	case "tx":
		return ph.sendTransaction(peer, getData.Hash)
	default:
		return fmt.Errorf("unknown data type: %s", getData.DataType)
	}
}

// handleInv handles inventory messages
func (ph *ProtocolHandler) handleInv(peer *Peer, msg *NetworkMessage) error {
	var inv InvMessage
	if err := json.Unmarshal(msg.Payload, &inv); err != nil {
		return fmt.Errorf("failed to unmarshal inv message: %v", err)
	}

	// Request data for items we don't have
	for _, item := range inv.Items {
		if !ph.node.Blockchain.HasBlock(item.Hash, item.BlockType) {
			getData := GetDataMessage{
				DataType:  item.Type,
				Hash:      item.Hash,
				BlockType: item.BlockType,
			}
			payload, _ := json.Marshal(getData)
			request := NetworkMessage{
				Type:      MessageTypeGetData,
				From:      ph.node.Config.Address,
				To:        peer.Address,
				Payload:   payload,
				Timestamp: time.Now(),
			}
			if err := peer.sendMessage(request); err != nil {
				logger.Error("Failed to request data", zap.Error(err))
			}
		}
	}

	return nil
}

// handleBlock handles block messages
func (ph *ProtocolHandler) handleBlock(peer *Peer, msg *NetworkMessage) error {
	var blockMsg BlockMessage
	if err := json.Unmarshal(msg.Payload, &blockMsg); err != nil {
		return fmt.Errorf("failed to unmarshal block message: %v", err)
	}

	// Validate and add block
	ph.node.Blockchain.LockForWrite()
	defer ph.node.Blockchain.UnlockForWrite()

	// Add block to appropriate chain
	if blockMsg.BlockType == blockchain.GoldenBlock {
		ph.node.Blockchain.GoldenBlocks = append(ph.node.Blockchain.GoldenBlocks, blockMsg.Block)
	} else {
		ph.node.Blockchain.SilverBlocks = append(ph.node.Blockchain.SilverBlocks, blockMsg.Block)
	}

	logger.Info("Received new block from peer",
		zap.String("peer", peer.Address),
		zap.String("hash", string(blockMsg.Block.Hash)),
		zap.String("type", string(blockMsg.BlockType)))

	// Broadcast to other peers
	ph.broadcastBlock(blockMsg.Block, blockMsg.BlockType, peer.Address)

	return nil
}

// handleTx handles transaction messages
func (ph *ProtocolHandler) handleTx(peer *Peer, msg *NetworkMessage) error {
	var txMsg TxMessage
	if err := json.Unmarshal(msg.Payload, &txMsg); err != nil {
		return fmt.Errorf("failed to unmarshal tx message: %v", err)
	}

	// Add transaction to pending pool
	ph.node.Blockchain.LockForWrite()
	ph.node.Blockchain.PendingTxs = append(ph.node.Blockchain.PendingTxs, txMsg.Transaction)
	ph.node.Blockchain.UnlockForWrite()

	logger.Info("Received new transaction from peer",
		zap.String("peer", peer.Address),
		zap.String("txid", string(txMsg.Transaction.ID)))

	// Broadcast to other peers
	ph.broadcastTransaction(txMsg.Transaction, peer.Address)

	return nil
}

// handleAddr handles addr messages
func (ph *ProtocolHandler) handleAddr(peer *Peer, msg *NetworkMessage) error {
	var addrMsg AddrMessage
	if err := json.Unmarshal(msg.Payload, &addrMsg); err != nil {
		return fmt.Errorf("failed to unmarshal addr message: %v", err)
	}

	// Add new peers to discovery
	for _, addr := range addrMsg.Addresses {
		if addr != ph.node.Config.Address {
			ph.node.ConnectToPeer(addr)
		}
	}

	return nil
}

// handleGetAddr handles getaddr messages
func (ph *ProtocolHandler) handleGetAddr(peer *Peer, msg *NetworkMessage) error {
	// Send our peer list
	ph.node.mu.RLock()
	addresses := make([]string, 0, len(ph.node.Peers))
	for _, p := range ph.node.Peers {
		addresses = append(addresses, p.Address)
	}
	ph.node.mu.RUnlock()

	addrMsg := AddrMessage{
		Addresses: addresses,
	}
	payload, _ := json.Marshal(addrMsg)
	response := NetworkMessage{
		Type:      MessageTypeAddr,
		From:      ph.node.Config.Address,
		To:        peer.Address,
		Payload:   payload,
		Timestamp: time.Now(),
	}
	return peer.sendMessage(response)
}

// handleGetHeight handles getheight messages
func (ph *ProtocolHandler) handleGetHeight(peer *Peer, msg *NetworkMessage) error {
	ph.node.Blockchain.LockForRead()
	goldenHeight := int64(len(ph.node.Blockchain.GoldenBlocks))
	silverHeight := int64(len(ph.node.Blockchain.SilverBlocks))
	ph.node.Blockchain.UnlockForRead()

	heightMsg := HeightMessage{
		GoldenHeight: goldenHeight,
		SilverHeight: silverHeight,
	}
	payload, _ := json.Marshal(heightMsg)
	response := NetworkMessage{
		Type:      MessageTypeGetHeight,
		From:      ph.node.Config.Address,
		To:        peer.Address,
		Payload:   payload,
		Timestamp: time.Now(),
	}
	return peer.sendMessage(response)
}

// sendBlock sends a block to a peer
func (ph *ProtocolHandler) sendBlock(peer *Peer, hash string) error {
	ph.node.Blockchain.LockForRead()
	defer ph.node.Blockchain.UnlockForRead()

	// Find block by hash
	var block blockchain.Block
	var blockType blockchain.BlockType
	found := false

	for _, b := range ph.node.Blockchain.GoldenBlocks {
		if string(b.Hash) == hash {
			block = b
			blockType = blockchain.GoldenBlock
			found = true
			break
		}
	}

	if !found {
		for _, b := range ph.node.Blockchain.SilverBlocks {
			if string(b.Hash) == hash {
				block = b
				blockType = blockchain.SilverBlock
				found = true
				break
			}
		}
	}

	if !found {
		return fmt.Errorf("block not found: %s", hash)
	}

	blockMsg := BlockMessage{
		Block:     block,
		BlockType: blockType,
	}
	payload, _ := json.Marshal(blockMsg)
	response := NetworkMessage{
		Type:      MessageTypeBlock,
		From:      ph.node.Config.Address,
		To:        peer.Address,
		Payload:   payload,
		Timestamp: time.Now(),
	}
	return peer.sendMessage(response)
}

// sendTransaction sends a transaction to a peer
func (ph *ProtocolHandler) sendTransaction(peer *Peer, txid string) error {
	ph.node.Blockchain.LockForRead()
	defer ph.node.Blockchain.UnlockForRead()

	// Find transaction by ID
	for _, tx := range ph.node.Blockchain.PendingTxs {
		if string(tx.ID) == txid {
			txMsg := TxMessage{
				Transaction: tx,
			}
			payload, _ := json.Marshal(txMsg)
			response := NetworkMessage{
				Type:      MessageTypeTx,
				From:      ph.node.Config.Address,
				To:        peer.Address,
				Payload:   payload,
				Timestamp: time.Now(),
			}
			return peer.sendMessage(response)
		}
	}

	return fmt.Errorf("transaction not found: %s", txid)
}

// broadcastBlock broadcasts a block to all peers except the sender
func (ph *ProtocolHandler) broadcastBlock(block blockchain.Block, blockType blockchain.BlockType, senderAddr string) {
	blockMsg := BlockMessage{
		Block:     block,
		BlockType: blockType,
	}
	payload, _ := json.Marshal(blockMsg)
	msg := NetworkMessage{
		Type:      MessageTypeBlock,
		From:      ph.node.Config.Address,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	ph.node.mu.RLock()
	defer ph.node.mu.RUnlock()

	for _, peer := range ph.node.Peers {
		if peer.Address != senderAddr {
			msg.To = peer.Address
			if err := peer.sendMessage(msg); err != nil {
				logger.Error("Failed to broadcast block to peer",
					zap.String("peer", peer.Address),
					zap.Error(err))
			}
		}
	}
}

// broadcastTransaction broadcasts a transaction to all peers except the sender
func (ph *ProtocolHandler) broadcastTransaction(tx blockchain.Transaction, senderAddr string) {
	txMsg := TxMessage{
		Transaction: tx,
	}
	payload, _ := json.Marshal(txMsg)
	msg := NetworkMessage{
		Type:      MessageTypeTx,
		From:      ph.node.Config.Address,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	ph.node.mu.RLock()
	defer ph.node.mu.RUnlock()

	for _, peer := range ph.node.Peers {
		if peer.Address != senderAddr {
			msg.To = peer.Address
			if err := peer.sendMessage(msg); err != nil {
				logger.Error("Failed to broadcast transaction to peer",
					zap.String("peer", peer.Address),
					zap.Error(err))
			}
		}
	}
}

// Message structures for protocol communication
type VersionMessage struct {
	Version   string `json:"version"`
	Height    int64  `json:"height"`
	Address   string `json:"address"`
	Timestamp int64  `json:"timestamp"`
}

type VerAckMessage struct {
	Accepted bool `json:"accepted"`
}

type PingMessage struct {
	Timestamp time.Time `json:"timestamp"`
}

type PongMessage struct {
	Timestamp time.Time `json:"timestamp"`
}

type GetBlocksMessage struct {
	BlockType blockchain.BlockType `json:"block_type"`
	StartHash string               `json:"start_hash"`
	EndHash   string               `json:"end_hash"`
}

type BlocksMessage struct {
	Blocks    []blockchain.Block   `json:"blocks"`
	BlockType blockchain.BlockType `json:"block_type"`
}

type GetDataMessage struct {
	DataType  string               `json:"data_type"`
	Hash      string               `json:"hash"`
	BlockType blockchain.BlockType `json:"block_type"`
}

type InvItem struct {
	Type      string               `json:"type"`
	Hash      string               `json:"hash"`
	BlockType blockchain.BlockType `json:"block_type"`
}

type InvMessage struct {
	Items []InvItem `json:"items"`
}

type BlockMessage struct {
	Block     blockchain.Block     `json:"block"`
	BlockType blockchain.BlockType `json:"block_type"`
}

type TxMessage struct {
	Transaction blockchain.Transaction `json:"transaction"`
}

type AddrMessage struct {
	Addresses []string `json:"addresses"`
}

type HeightMessage struct {
	GoldenHeight int64 `json:"golden_height"`
	SilverHeight int64 `json:"silver_height"`
}
