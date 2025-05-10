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
	"github.com/youngchain/internal/core/common"
	"github.com/youngchain/internal/interfaces"
	"github.com/youngchain/internal/network/messages"
	"github.com/youngchain/internal/network/peers"
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
	peerManager *peers.Manager

	// Message handling
	MessageChan chan interface{}

	// Server state
	IsRunning bool
	StopChan  chan struct{}

	// Database
	db *storage.Database

	// Consensus
	consensus interfaces.Consensus

	// Transaction pool
	txPool *TransactionPool

	// Connected peers
	peers map[string]*Peer
}

// NewServer creates a new network server
func NewServer(config *config.Config, consensus interfaces.Consensus) *Server {
	return &Server{
		config:      config,
		peerManager: peers.NewManager(config),
		MessageChan: make(chan interface{}, 100),
		StopChan:    make(chan struct{}),
		consensus:   consensus,
		txPool:      NewTransactionPool(),
		peers:       make(map[string]*Peer),
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

	s.peerManager.Stop()
	return nil
}

// SetDB sets the database for the server
func (s *Server) SetDB(db *storage.Database) {
	s.db = db
}

// HandleMessage handles incoming network messages
func (s *Server) HandleMessage(msg interface{}) error {
	if msg == nil {
		return fmt.Errorf("received nil message")
	}

	switch m := msg.(type) {
	case *common.Transaction:
		return s.ProcessTransaction(m)
	case *messages.BlockMessage:
		if m.Block == nil {
			return fmt.Errorf("received nil block in block message")
		}

		// Validate block header
		if err := s.validateBlockHeader(m.Block); err != nil {
			return fmt.Errorf("invalid block header: %v", err)
		}

		// Validate block and its transactions using consensus
		if err := s.validateBlockTransactions(m.Block); err != nil {
			return fmt.Errorf("invalid block: %v", err)
		}

		// Update chain state
		if err := s.UpdateChainState(m.Block); err != nil {
			return fmt.Errorf("failed to update chain state: %v", err)
		}

		// Broadcast block to other peers
		return s.BroadcastBlock(m.Block)
	case *messages.TransactionMessage:
		if m.Transaction == nil {
			return fmt.Errorf("received nil transaction in transaction message")
		}
		return s.ProcessTransaction(m.Transaction)
	default:
		return fmt.Errorf("unknown message type: %T", msg)
	}
}

// ProcessTransaction processes a new transaction
func (s *Server) ProcessTransaction(tx *common.Transaction) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Add to transaction pool
	if err := s.txPool.AddTransaction(tx); err != nil {
		return fmt.Errorf("failed to add transaction to pool: %v", err)
	}

	// Broadcast to peers
	for _, peer := range s.peers {
		if err := peer.SendTransaction(tx); err != nil {
			// Log error but continue with other peers
			fmt.Printf("Failed to send transaction to peer %s: %v\n", peer.ID(), err)
		}
	}

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
	// Use consensus engine to validate the entire block
	return s.consensus.ValidateBlock(block)
}

// UpdateChainState updates the chain state
func (s *Server) UpdateChainState(block *block.Block) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate block using consensus
	if err := s.consensus.ValidateBlock(block); err != nil {
		return fmt.Errorf("invalid block: %v", err)
	}

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
		// Verify block connects to our chain
		if !bytes.Equal(block.Header.PrevBlockHash, bestBlockHash) {
			// TODO: Handle chain reorganization
			return fmt.Errorf("block does not connect to our chain")
		}

		if err := s.db.UpdateChainState(block.Header.Hash, block.Header.Height); err != nil {
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

// GetPendingTransactions returns all pending transactions
func (s *Server) GetPendingTransactions() []*common.Transaction {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.txPool.GetPendingTransactions()
}

// GetPeerHeights implements interfaces.Network
func (s *Server) GetPeerHeights() []uint64 {
	heights := make([]uint64, 0)
	for _, peer := range s.peerManager.GetPeers() {
		heights = append(heights, peer.GetHeight())
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

// BroadcastTransaction broadcasts a transaction to all peers
func (s *Server) BroadcastTransaction(tx *common.Transaction) error {
	if tx == nil {
		return fmt.Errorf("cannot broadcast nil transaction")
	}

	// Create transaction message
	msg := &messages.TransactionMessage{
		Transaction: tx,
		CoinType:    "default",
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction message: %v", err)
	}

	// Broadcast using peer manager
	s.peerManager.Broadcast(data)
	return nil
}
