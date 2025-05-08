package server

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/youngchain/internal/config"
	"github.com/youngchain/internal/consensus"
	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/coin"
	"github.com/youngchain/internal/core/types"
	"github.com/youngchain/internal/network/messages"
	"github.com/youngchain/internal/network/peers"
)

// Server represents a network server
type Server struct {
	config    *config.Config
	consensus *consensus.Consensus
	peerMgr   *peers.Manager
	mu        sync.RWMutex
}

// NewServer creates a new server
func NewServer(config *config.Config, consensus *consensus.Consensus) *Server {
	return &Server{
		config:    config,
		consensus: consensus,
		peerMgr:   peers.NewManager(config),
	}
}

// Start starts the server
func (s *Server) Start() error {
	// Start peer manager
	if err := s.peerMgr.Start(); err != nil {
		return fmt.Errorf("failed to start peer manager: %v", err)
	}

	// Start listening for connections
	listener, err := net.Listen("tcp", s.config.ListenAddr)
	if err != nil {
		return fmt.Errorf("failed to start listener: %v", err)
	}

	go s.acceptConnections(listener)
	return nil
}

// Stop stops the server
func (s *Server) Stop() {
	s.peerMgr.Stop()
}

// acceptConnections accepts incoming connections
func (s *Server) acceptConnections(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		go s.handleConnection(conn)
	}
}

// handleConnection handles a connection
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Create peer info
	info := peers.Info{
		Address:  conn.RemoteAddr().String(),
		Version:  "1.0.0", // TODO: Get from config
		LastSeen: time.Now(),
	}

	// Add peer
	s.peerMgr.AddPeer(conn, info)

	// Handle messages
	for {
		// Read message length
		var length uint32
		if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
			break
		}

		// Read message
		message := make([]byte, length)
		if _, err := io.ReadFull(conn, message); err != nil {
			break
		}

		// Handle message
		if err := s.handleMessage(message); err != nil {
			break
		}
	}

	// Remove peer
	s.peerMgr.RemovePeer(info.Address)
}

// handleMessage handles a message
func (s *Server) handleMessage(message []byte) error {
	var msg messages.Message
	if err := json.Unmarshal(message, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %v", err)
	}

	switch msg.Type {
	case messages.BlockMsg:
		var blockMsg messages.BlockMessage
		if err := json.Unmarshal(msg.Data, &blockMsg); err != nil {
			return fmt.Errorf("failed to unmarshal block message: %v", err)
		}
		return s.handleBlockMessage(blockMsg)

	case messages.TransactionMsg:
		var txMsg messages.TransactionMessage
		if err := json.Unmarshal(msg.Data, &txMsg); err != nil {
			return fmt.Errorf("failed to unmarshal transaction message: %v", err)
		}
		return s.handleTransactionMessage(txMsg)

	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

// handleBlockMessage handles a block message
func (s *Server) handleBlockMessage(msg messages.BlockMessage) error {
	// Validate block
	if err := s.consensus.ValidateBlock(msg.Block); err != nil {
		return fmt.Errorf("invalid block: %v", err)
	}

	// Add block to chain
	if err := s.consensus.MineBlock(msg.Block); err != nil {
		return fmt.Errorf("failed to mine block: %v", err)
	}

	// Broadcast block
	messageData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal block message: %v", err)
	}
	s.peerMgr.Broadcast(messageData)
	return nil
}

// handleTransactionMessage handles a transaction message
func (s *Server) handleTransactionMessage(msg messages.TransactionMessage) error {
	// Validate transaction
	if err := s.consensus.ValidateTransaction(msg.Transaction); err != nil {
		return fmt.Errorf("invalid transaction: %v", err)
	}

	// Add transaction to mempool
	if err := s.consensus.ValidateTransaction(msg.Transaction); err != nil {
		return fmt.Errorf("failed to validate transaction: %v", err)
	}

	// Broadcast transaction
	messageData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction message: %v", err)
	}
	s.peerMgr.Broadcast(messageData)
	return nil
}

// BroadcastBlock broadcasts a block
func (s *Server) BroadcastBlock(block *block.Block) error {
	// Determine block type based on mining reward transaction
	blockType := block.GoldenBlock // Default to golden block
	if len(block.Transactions) > 0 {
		switch block.Transactions[0].CoinType {
		case coin.Leah:
			blockType = block.GoldenBlock
		case coin.Shiblum:
			blockType = block.SilverBlock
		}
	}

	msg := messages.BlockMessage{
		Block:     block,
		BlockType: blockType,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal block message: %v", err)
	}

	message, err := messages.NewMessage(messages.BlockMsg, data)
	if err != nil {
		return fmt.Errorf("failed to create message: %v", err)
	}

	messageData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	s.peerMgr.Broadcast(messageData)
	return nil
}

// BroadcastTransaction broadcasts a transaction
func (s *Server) BroadcastTransaction(tx *types.Transaction) error {
	msg := messages.TransactionMessage{
		Transaction: tx,
		CoinType:    tx.CoinType,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction message: %v", err)
	}

	message, err := messages.NewMessage(messages.BlockMsg, data)
	if err != nil {
		return fmt.Errorf("failed to create message: %v", err)
	}

	messageData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	s.peerMgr.Broadcast(messageData)
	return nil
}
