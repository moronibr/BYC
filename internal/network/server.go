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
	"github.com/youngchain/internal/storage"
)

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
	case MessageType("block"):
		var blockMsg struct {
			Block     *block.Block    `json:"block"`
			BlockType block.BlockType `json:"block_type"`
		}
		if err := json.Unmarshal(msg.Data, &blockMsg); err != nil {
			return fmt.Errorf("failed to unmarshal block message: %v", err)
		}
		return s.processBlock(blockMsg.Block, blockMsg.BlockType)

	case MessageType("transaction"):
		var txMsg struct {
			Transaction block.Transaction `json:"transaction"`
			CoinType    coin.CoinType     `json:"coin_type"`
		}
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
	msg, err := NewMessage(BlockMsg, blockMsg)
	if err != nil {
		return fmt.Errorf("failed to create block message: %v", err)
	}
	s.BroadcastMessage(msg)

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
	validator := transaction.NewValidator(s.db)

	// Validate each transaction
	for _, tx := range block.Transactions {
		if err := validator.ValidateTransaction(tx, nil); err != nil {
			return fmt.Errorf("transaction validation failed: %v", err)
		}
	}

	return nil
}

// UpdateChainState updates the chain state
func (s *Server) UpdateChainState(block *block.Block) error {
	height, _, err := s.db.GetChainState()
	if err != nil {
		return err
	}
	return s.db.SaveChainState(height+1, block.Hash)
}

// calculateNextDifficulty calculates the next block difficulty
func (s *Server) calculateNextDifficulty(prevBlock *block.Block) uint32 {
	// Since we don't have a GetLastBlocks method, we'll need to implement this differently
	// For now, we'll return the initial difficulty
	return block.GetInitialDifficulty(block.GoldenBlock)
}

// processTransaction processes a transaction
func (s *Server) processTransaction(tx block.Transaction, coinType coin.CoinType) error {
	// TODO: Implement transaction processing logic
	return nil
}

// processMessages processes messages from the message channel
func (s *Server) processMessages() {
	for {
		select {
		case <-s.StopChan:
			return
		case msg := <-s.MessageChan:
			s.BroadcastMessage(msg)
		}
	}
}

// BroadcastMessage broadcasts a message to all peers
func (s *Server) BroadcastMessage(msg *Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	s.peerManager.Broadcast(data)
}

// ConnectToPeer connects to a peer
func (s *Server) ConnectToPeer(address string) error {
	return s.peerManager.ConnectToPeer(address)
}

// GetBlockByHash retrieves a block by its hash
func (s *Server) GetBlockByHash(hash []byte) (*block.Block, error) {
	return s.db.GetBlock(hash)
}

// GetLastBlock retrieves the last block in the chain
func (s *Server) GetLastBlock() (*block.Block, error) {
	_, bestBlockHash, err := s.db.GetChainState()
	if err != nil {
		return nil, err
	}
	return s.db.GetBlock(bestBlockHash)
}
