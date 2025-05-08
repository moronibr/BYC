package server

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

const (
	maxMessageSize = 1024 * 1024 // 1MB
	maxRetries     = 3
	retryDelay     = time.Second
	readTimeout    = 30 * time.Second
	writeTimeout   = 30 * time.Second
)

// Server represents a network server
type Server struct {
	config    *config.Config
	consensus *consensus.Consensus
	peerMgr   *peers.Manager
	mu        sync.RWMutex
	logger    *log.Logger
}

// NewServer creates a new server
func NewServer(config *config.Config, consensus *consensus.Consensus) *Server {
	return &Server{
		config:    config,
		consensus: consensus,
		peerMgr:   peers.NewManager(config),
		logger:    log.New(log.Writer(), "[Server] ", log.LstdFlags),
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

	s.logger.Printf("Server started listening on %s", s.config.ListenAddr)
	go s.acceptConnections(listener)
	return nil
}

// Stop stops the server
func (s *Server) Stop() {
	s.logger.Println("Stopping server...")
	s.peerMgr.Stop()
}

// acceptConnections accepts incoming connections
func (s *Server) acceptConnections(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			s.logger.Printf("Failed to accept connection: %v", err)
			continue
		}

		s.logger.Printf("New connection from %s", conn.RemoteAddr().String())
		go s.handleConnection(conn)
	}
}

// handleConnection handles a connection
func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			s.logger.Printf("Error closing connection: %v", err)
		}
	}()

	// Set timeouts
	if err := conn.SetDeadline(time.Now().Add(readTimeout)); err != nil {
		s.logger.Printf("Failed to set read deadline: %v", err)
		return
	}

	// Create peer info
	info := peers.Info{
		Address:  conn.RemoteAddr().String(),
		Version:  "1.0.0", // TODO: Get from config
		LastSeen: time.Now(),
	}

	// Add peer
	s.peerMgr.AddPeer(conn, info)
	s.logger.Printf("Added peer %s", info.Address)

	// Handle messages
	for {
		// Reset read deadline
		if err := conn.SetDeadline(time.Now().Add(readTimeout)); err != nil {
			s.logger.Printf("Failed to set read deadline: %v", err)
			break
		}

		// Read message length
		var length uint32
		if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
			if err != io.EOF {
				s.logger.Printf("Error reading message length: %v", err)
			}
			break
		}

		// Validate message size
		if length > maxMessageSize {
			s.logger.Printf("Message too large: %d bytes (max: %d)", length, maxMessageSize)
			break
		}

		// Read message
		message := make([]byte, length)
		if _, err := io.ReadFull(conn, message); err != nil {
			s.logger.Printf("Error reading message: %v", err)
			break
		}

		// Handle message with retry
		var handleErr error
		for i := 0; i < maxRetries; i++ {
			if err := s.handleMessage(message); err != nil {
				handleErr = err
				s.logger.Printf("Failed to handle message (attempt %d/%d): %v", i+1, maxRetries, err)
				time.Sleep(retryDelay)
				continue
			}
			handleErr = nil
			break
		}
		if handleErr != nil {
			s.logger.Printf("Failed to handle message after %d attempts: %v", maxRetries, handleErr)
			break
		}
	}

	// Remove peer
	s.peerMgr.RemovePeer(info.Address)
	s.logger.Printf("Removed peer %s", info.Address)
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
	// Validate block with retry
	var validateErr error
	for i := 0; i < maxRetries; i++ {
		if err := s.consensus.ValidateBlock(msg.Block); err != nil {
			validateErr = err
			s.logger.Printf("Failed to validate block (attempt %d/%d): %v", i+1, maxRetries, err)
			time.Sleep(retryDelay)
			continue
		}
		validateErr = nil
		break
	}
	if validateErr != nil {
		return fmt.Errorf("failed to validate block after %d attempts: %v", maxRetries, validateErr)
	}

	// Add block to chain with retry
	var mineErr error
	for i := 0; i < maxRetries; i++ {
		if err := s.consensus.MineBlock(msg.Block); err != nil {
			mineErr = err
			s.logger.Printf("Failed to mine block (attempt %d/%d): %v", i+1, maxRetries, err)
			time.Sleep(retryDelay)
			continue
		}
		mineErr = nil
		break
	}
	if mineErr != nil {
		return fmt.Errorf("failed to mine block after %d attempts: %v", maxRetries, mineErr)
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
	// Validate transaction with retry
	var validateErr error
	for i := 0; i < maxRetries; i++ {
		if err := s.consensus.ValidateTransaction(msg.Transaction); err != nil {
			validateErr = err
			s.logger.Printf("Failed to validate transaction (attempt %d/%d): %v", i+1, maxRetries, err)
			time.Sleep(retryDelay)
			continue
		}
		validateErr = nil
		break
	}
	if validateErr != nil {
		return fmt.Errorf("failed to validate transaction after %d attempts: %v", maxRetries, validateErr)
	}

	// Add transaction to mempool with retry
	var addErr error
	for i := 0; i < maxRetries; i++ {
		if err := s.consensus.ValidateTransaction(msg.Transaction); err != nil {
			addErr = err
			s.logger.Printf("Failed to add transaction to mempool (attempt %d/%d): %v", i+1, maxRetries, err)
			time.Sleep(retryDelay)
			continue
		}
		addErr = nil
		break
	}
	if addErr != nil {
		return fmt.Errorf("failed to add transaction to mempool after %d attempts: %v", maxRetries, addErr)
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
func (s *Server) BroadcastBlock(b *block.Block) error {
	// Determine block type based on mining reward transaction
	blockType := block.GoldenBlock // Default to golden block
	if len(b.Transactions) > 0 {
		switch b.Transactions[0].CoinType {
		case coin.Leah:
			blockType = block.GoldenBlock
		case coin.Shiblum:
			blockType = block.SilverBlock
		}
	}

	msg := messages.BlockMessage{
		Block:     b,
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
