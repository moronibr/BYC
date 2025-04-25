package network

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/youngchain/internal/config"
	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/coin"
)

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

// handleMessage handles a message from a peer
func (s *Server) handleMessage(msg *Message) error {
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
	// TODO: Implement block processing logic
	return nil
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
