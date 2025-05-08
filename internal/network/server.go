package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/youngchain/internal/config"
	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/types"
	"github.com/youngchain/internal/interfaces"
	"github.com/youngchain/internal/network/messages"
	"github.com/youngchain/internal/storage"
)

// ChainState represents the current state of the blockchain
type ChainState struct {
	BestBlockHash []byte
	BlockNumber   uint64
	Timestamp     time.Time
}

// Server implements both BlockChain and Network interfaces
type Server struct {
	mu sync.RWMutex

	// Server configuration
	config *config.Config

	// Peer management
	peerManager *PeerManager

	// Message handling
	MessageChan chan *messages.Message

	// Server state
	IsRunning bool
	StopChan  chan struct{}

	// Database
	db *storage.Database

	// Consensus
	consensus interfaces.Consensus

	// Transaction pool
	transactionPool *TransactionPool
}

// NewServer creates a new network server
func NewServer(config *config.Config, consensus interfaces.Consensus) *Server {
	return &Server{
		config:          config,
		peerManager:     NewPeerManager(config),
		MessageChan:     make(chan *messages.Message, 100),
		StopChan:        make(chan struct{}),
		consensus:       consensus,
		transactionPool: NewTransactionPool(),
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
func (s *Server) Stop() error {
	if !s.IsRunning {
		return nil
	}

	close(s.StopChan)
	s.IsRunning = false

	if err := s.peerManager.Stop(); err != nil {
		return fmt.Errorf("failed to stop peer manager: %v", err)
	}

	return nil
}

// SetDB sets the database for the server
func (s *Server) SetDB(db *storage.Database) {
	s.db = db
}

// HandleMessage handles a message from a peer
func (s *Server) HandleMessage(msg *messages.Message) error {
	switch msg.Type {
	case messages.BlockMsg:
		var blockMsg messages.BlockMessage
		if err := json.Unmarshal(msg.Data, &blockMsg); err != nil {
			return fmt.Errorf("failed to unmarshal block message: %v", err)
		}
		return s.processBlock(blockMsg.Block)

	case messages.TransactionMsg:
		var txMsg messages.TransactionMessage
		if err := json.Unmarshal(msg.Data, &txMsg); err != nil {
			return fmt.Errorf("failed to unmarshal transaction message: %v", err)
		}
		return s.processTransaction(txMsg.Transaction)

	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

// processBlock processes a block
func (s *Server) processBlock(block *block.Block) error {
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

	// Adjust difficulty
	if err := s.consensus.AdjustDifficulty(block); err != nil {
		return fmt.Errorf("failed to adjust difficulty: %v", err)
	}

	// Broadcast block to peers
	blockMsg := &messages.BlockMessage{
		Block:     block,
		BlockType: "new",
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
	for _, tx := range block.Transactions {
		if err := s.transactionPool.AddTransaction(tx); err != nil {
			return fmt.Errorf("transaction validation failed: %v", err)
		}
	}

	return nil
}

// UpdateChainState updates the chain state
func (s *Server) UpdateChainState(block *block.Block) error {
	// Get current chain state
	_, bestBlockHash, err := s.db.GetChainState()
	if err != nil {
		return fmt.Errorf("failed to get chain state: %v", err)
	}

	// Get current best block
	bestBlock, err := s.db.GetBlock(bestBlockHash)
	if err != nil {
		return fmt.Errorf("failed to get best block: %v", err)
	}

	// If new block has higher height, update chain state
	if block.Header.Height > bestBlock.Header.Height {
		if err := s.db.UpdateChainState(block.Hash, block.Header.Height); err != nil {
			return fmt.Errorf("failed to update chain state: %v", err)
		}
	}

	return nil
}

// calculateNextDifficulty calculates the next block difficulty
func (s *Server) calculateNextDifficulty(prevBlock *block.Block) uint32 {
	// TODO: Implement difficulty adjustment algorithm
	return prevBlock.Header.Difficulty
}

// processTransaction processes a transaction
func (s *Server) processTransaction(tx *types.Transaction) error {
	// Add transaction to pool
	if err := s.transactionPool.AddTransaction(tx); err != nil {
		return fmt.Errorf("failed to add transaction to pool: %v", err)
	}

	// Broadcast transaction to peers
	txMsg := &messages.TransactionMessage{
		Transaction: tx,
		CoinType:    "default",
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

// GetLastBlock implements interfaces.BlockChain
func (s *Server) GetLastBlock() (*block.Block, error) {
	_, bestBlockHash, err := s.db.GetChainState()
	if err != nil {
		return nil, fmt.Errorf("failed to get chain state: %v", err)
	}

	return s.db.GetBlock(bestBlockHash)
}

// GetBlockByHeight implements interfaces.BlockChain
func (s *Server) GetBlockByHeight(height uint64) (*block.Block, error) {
	return s.db.GetBlockByHeight(height)
}

// GetBlock implements interfaces.BlockChain
func (s *Server) GetBlock(hash []byte) (*block.Block, error) {
	return s.db.GetBlock(hash)
}

// GetPendingTransactions implements interfaces.BlockChain
func (s *Server) GetPendingTransactions() []*types.Transaction {
	return s.transactionPool.GetPendingTransactions()
}

// GetPeerHeights implements interfaces.Network
func (s *Server) GetPeerHeights() []uint64 {
	heights := make([]uint64, 0)
	for _, peer := range s.peerManager.peers {
		heights = append(heights, peer.GetInfo().Height)
	}
	return heights
}

// RequestBlocks implements interfaces.Network
func (s *Server) RequestBlocks(startHeight, endHeight uint64) ([]*block.Block, error) {
	blocks := make([]*block.Block, 0)
	for height := startHeight; height <= endHeight; height++ {
		block, err := s.GetBlockByHeight(height)
		if err != nil {
			return nil, fmt.Errorf("failed to get block at height %d: %v", height, err)
		}
		blocks = append(blocks, block)
	}
	return blocks, nil
}

// BroadcastBlock implements interfaces.Network
func (s *Server) BroadcastBlock(block *block.Block) error {
	// Create block message
	msg, err := messages.NewMessage(messages.BlockMsg, &messages.BlockMessage{
		Block:     block,
		BlockType: "new",
	})
	if err != nil {
		return err
	}

	// Broadcast message to all peers
	msgData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}
	s.peerManager.Broadcast(msgData)
	return nil
}

// BroadcastTransaction implements interfaces.Network
func (s *Server) BroadcastTransaction(tx interface{}) error {
	// Convert interface to transaction
	transaction, ok := tx.(*types.Transaction)
	if !ok {
		return fmt.Errorf("invalid transaction type")
	}

	// Create transaction message
	msg, err := messages.NewMessage(messages.TransactionMsg, &messages.TransactionMessage{
		Transaction: transaction,
		CoinType:    "default",
	})
	if err != nil {
		return err
	}

	// Broadcast message to all peers
	msgData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}
	s.peerManager.Broadcast(msgData)
	return nil
}
