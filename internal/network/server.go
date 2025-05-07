package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/youngchain/internal/config"
	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/coin"
	"github.com/youngchain/internal/core/transaction"
	"github.com/youngchain/internal/core/types"
	"github.com/youngchain/internal/storage"
)

// MessageType represents the type of message
type MessageType string

const (
	BlockMsg       MessageType = "block"
	TransactionMsg MessageType = "transaction"
)

// Message represents a network message
type Message struct {
	Type MessageType     `json:"type"`
	Data json.RawMessage `json:"data"`
}

// BlockMessage represents a block message
type BlockMessage struct {
	Block     *block.Block    `json:"block"`
	BlockType block.BlockType `json:"block_type"`
}

// TransactionMessage represents a transaction message
type TransactionMessage struct {
	Transaction *types.Transaction `json:"transaction"`
	CoinType    coin.CoinType      `json:"coin_type"`
}

// NewMessage creates a new message
func NewMessage(msgType MessageType, data interface{}) (*Message, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message data: %v", err)
	}

	return &Message{
		Type: msgType,
		Data: jsonData,
	}, nil
}

// ChainState represents the current state of the blockchain
type ChainState struct {
	BestBlockHash []byte
	BlockNumber   uint64
	Timestamp     time.Time
}

// Server represents a network server for node communication
type Server struct {
	// Server configuration
	config *config.Config

	// Peer management
	peerManager *PeerManager

	// Message handling
	MessageChan chan *Message

	// Server state
	IsRunning bool
	StopChan  chan struct{}

	// Database
	db *storage.DB
}

// NewServer creates a new network server
func NewServer(config *config.Config) *Server {
	return &Server{
		config:      config,
		peerManager: NewPeerManager(config),
		MessageChan: make(chan *Message, 100),
		StopChan:    make(chan struct{}),
	}
}

// Start starts the server
func (s *Server) Start() error {
	if s.IsRunning {
		return fmt.Errorf("server is already running")
	}

	if err := s.peerManager.Start(); err != nil {
		return fmt.Errorf("failed to start peer manager: %v", err)
	}

	s.IsRunning = true
	go s.processMessages()

	return nil
}

// Stop stops the server
func (s *Server) Stop() {
	if !s.IsRunning {
		return
	}

	close(s.StopChan)
	s.IsRunning = false
	s.peerManager.Stop()
}

// SetDB sets the database for the server
func (s *Server) SetDB(db *storage.DB) {
	s.db = db
}

// HandleMessage handles a message from a peer
func (s *Server) HandleMessage(msg *Message) error {
	switch msg.Type {
	case BlockMsg:
		var blockMsg BlockMessage
		if err := json.Unmarshal(msg.Data, &blockMsg); err != nil {
			return fmt.Errorf("failed to unmarshal block message: %v", err)
		}
		return s.processBlock(blockMsg.Block, blockMsg.BlockType)

	case TransactionMsg:
		var txMsg TransactionMessage
		if err := json.Unmarshal(msg.Data, &txMsg); err != nil {
			return fmt.Errorf("failed to unmarshal transaction message: %v", err)
		}
		return s.processTransaction(txMsg.Transaction, txMsg.CoinType)

	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

// processBlock processes a block
func (s *Server) processBlock(block *block.Block, blockType block.BlockType) error {
	// Validate block hash
	calculatedHash := block.CalculateHash()
	if !bytes.Equal(calculatedHash, block.Hash) {
		return fmt.Errorf("invalid block hash")
	}

	// Validate block header
	if err := s.validateBlockHeader(block); err != nil {
		return fmt.Errorf("invalid block header: %v", err)
	}

	// Validate transactions
	if err := s.validateBlockTransactions(block); err != nil {
		return fmt.Errorf("invalid block transactions: %v", err)
	}

	// Save block to database
	if err := s.db.SaveBlock(block); err != nil {
		return fmt.Errorf("failed to save block: %v", err)
	}

	// Update chain state
	if err := s.UpdateChainState(block); err != nil {
		return fmt.Errorf("failed to update chain state: %v", err)
	}

	// Broadcast block to peers
	blockMsg := &BlockMessage{
		Block:     block,
		BlockType: blockType,
	}
	msgData, err := json.Marshal(blockMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal block message: %v", err)
	}
	s.peerManager.Broadcast(msgData)

	return nil
}

// validateBlockHeader validates the block header
func (s *Server) validateBlockHeader(block *block.Block) error {
	// Check version
	if block.Header.Version != 1 {
		return fmt.Errorf("invalid block version")
	}

	// Check timestamp
	if block.Header.Timestamp.After(time.Now()) {
		return fmt.Errorf("block timestamp is in the future")
	}

	// Validate block
	if len(block.Header.PrevBlockHash) > 0 {
		prevBlock, err := s.db.GetBlock(block.Header.PrevBlockHash)
		if err != nil {
			return fmt.Errorf("failed to get previous block: %v", err)
		}

		// Check difficulty
		if block.Header.Difficulty != s.calculateNextDifficulty(prevBlock) {
			return fmt.Errorf("invalid block difficulty")
		}
	}

	return nil
}

// validateBlockTransactions validates all transactions in the block
func (s *Server) validateBlockTransactions(block *block.Block) error {
	// Create transaction validator
	validator := transaction.NewTransactionPool(1000)

	// Validate each transaction
	for _, tx := range block.Transactions {
		if err := validator.Validate(tx); err != nil {
			return fmt.Errorf("transaction validation failed: %v", err)
		}
	}

	return nil
}

// UpdateChainState updates the chain state
func (s *Server) UpdateChainState(block *block.Block) error {
	height, _, err := s.db.GetChainState()
	if err != nil {
		return fmt.Errorf("failed to get chain state: %v", err)
	}

	// Update chain state
	if err := s.db.SaveChainState(height+1, block.Hash); err != nil {
		return fmt.Errorf("failed to save chain state: %v", err)
	}

	return nil
}

// calculateNextDifficulty calculates the next block difficulty
func (s *Server) calculateNextDifficulty(prevBlock *block.Block) uint32 {
	// TODO: Implement difficulty adjustment algorithm
	return prevBlock.Header.Difficulty
}

// processTransaction processes a transaction
func (s *Server) processTransaction(tx *types.Transaction, coinType coin.CoinType) error {
	// Create transaction validator
	validator := transaction.NewTransactionPool(1000)

	// Validate transaction
	if err := validator.Validate(tx); err != nil {
		return fmt.Errorf("transaction validation failed: %v", err)
	}

	// Add transaction to pool
	if err := validator.AddToPool(tx); err != nil {
		return fmt.Errorf("failed to add transaction to pool: %v", err)
	}

	// Broadcast transaction to peers
	txMsg := &TransactionMessage{
		Transaction: tx,
		CoinType:    coinType,
	}
	msgData, err := json.Marshal(txMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction message: %v", err)
	}
	s.peerManager.Broadcast(msgData)

	return nil
}

// processMessages processes incoming messages
func (s *Server) processMessages() {
	for {
		select {
		case msg := <-s.MessageChan:
			if err := s.HandleMessage(msg); err != nil {
				log.Printf("Failed to handle message: %v", err)
			}
		case <-s.StopChan:
			return
		}
	}
}

// BroadcastMessage broadcasts a message to all peers
func (s *Server) BroadcastMessage(msg *Message) {
	msgData, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}
	s.peerManager.Broadcast(msgData)
}

// ConnectToPeer connects to a peer
func (s *Server) ConnectToPeer(address string) error {
	return s.peerManager.ConnectToPeer(address)
}

// GetBlockByHash gets a block by its hash
func (s *Server) GetBlockByHash(hash []byte) (*block.Block, error) {
	return s.db.GetBlock(hash)
}

// GetLastBlock gets the last block in the chain
func (s *Server) GetLastBlock() (*block.Block, error) {
	_, bestBlockHash, err := s.db.GetChainState()
	if err != nil {
		return nil, fmt.Errorf("failed to get chain state: %v", err)
	}

	return s.db.GetBlock(bestBlockHash)
}
