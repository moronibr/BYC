package server

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"errors"
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
	"github.com/youngchain/internal/core/common"
	"github.com/youngchain/internal/core/types"
	"github.com/youngchain/internal/network/messages"
	"github.com/youngchain/internal/network/peers"
	"github.com/youngchain/internal/security"
	"golang.org/x/time/rate"
)

const (
	maxMessageSize  = 1024 * 1024 // 1MB
	maxRetries      = 3
	retryDelay      = time.Second
	readTimeout     = 30 * time.Second
	writeTimeout    = 30 * time.Second
	shutdownTimeout = 30 * time.Second
)

// Error types
var (
	ErrInvalidMessage     = errors.New("invalid message")
	ErrMessageTooLarge    = errors.New("message too large")
	ErrConnectionTimeout  = errors.New("connection timeout")
	ErrPeerNotFound       = errors.New("peer not found")
	ErrInvalidBlock       = errors.New("invalid block")
	ErrInvalidTransaction = errors.New("invalid transaction")
	ErrServerShutdown     = errors.New("server is shutting down")
)

// Server represents a network server
type Server struct {
	config    *config.Config
	consensus *consensus.Consensus
	peerMgr   *peers.Manager
	mu        sync.RWMutex
	logger    *log.Logger
	metrics   *ServerMetrics

	// Security components
	peerAuth    *security.PeerAuth
	msgSigner   *security.MessageSigner
	peerLimiter *security.PeerLimiter

	// Shutdown related fields
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	shutdownCh chan struct{}
	isShutdown bool
	tlsConfig  *security.TLSServerConfig
}

// ServerMetrics tracks server metrics
type ServerMetrics struct {
	mu                sync.RWMutex
	activeConnections int64
	totalConnections  int64
	messagesReceived  int64
	messagesSent      int64
	errors            int64
	lastError         error
	lastErrorTime     time.Time
}

// NewServerMetrics creates new server metrics
func NewServerMetrics() *ServerMetrics {
	return &ServerMetrics{}
}

// NewServer creates a new server
func NewServer(config *config.Config, consensus *consensus.Consensus) *Server {
	ctx, cancel := context.WithCancel(context.Background())

	// Generate key pair for message signing
	publicKey, privateKey, err := security.GenerateKeyPair()
	if err != nil {
		log.Fatalf("Failed to generate key pair: %v", err)
	}

	// Initialize TLS configuration
	tlsConfig := security.DefaultTLSServerConfig()
	tlsConfig.CertFile = config.TLS.CertFile
	tlsConfig.KeyFile = config.TLS.KeyFile
	tlsConfig.ClientCAs = config.TLS.ClientCAs
	tlsConfig.ServerName = config.TLS.ServerName

	return &Server{
		config:     config,
		consensus:  consensus,
		peerMgr:    peers.NewManager(config),
		logger:     log.New(log.Writer(), "[Server] ", log.LstdFlags),
		metrics:    NewServerMetrics(),
		ctx:        ctx,
		cancel:     cancel,
		shutdownCh: make(chan struct{}),
		tlsConfig:  tlsConfig,

		// Initialize security components
		peerAuth:  security.NewPeerAuth(),
		msgSigner: security.NewMessageSigner(privateKey, publicKey),
		peerLimiter: security.NewPeerLimiter(
			rate.Limit(100),  // 100 connections per second per IP
			rate.Limit(1000), // 1000 messages per second per connection
			100,              // 100 burst per IP
			1000,             // 1000 burst per connection
			10,               // max 10 connections per IP
		),
	}
}

// Start starts the server
func (s *Server) Start() error {
	s.logger.Printf("Starting server on %s", s.config.ListenAddr)

	// Start peer manager
	if err := s.peerMgr.Start(); err != nil {
		s.recordError(fmt.Errorf("failed to start peer manager: %v", err))
		return err
	}

	// Create TLS listener
	tlsConfig, err := security.LoadTLSServerConfig(s.tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to load TLS config: %v", err)
	}

	listener, err := tls.Listen("tcp", s.config.ListenAddr, tlsConfig)
	if err != nil {
		s.recordError(fmt.Errorf("failed to create TLS listener: %v", err))
		return err
	}
	defer listener.Close()

	s.logger.Printf("Server started successfully on %s", s.config.ListenAddr)

	// Start connection acceptor in a goroutine
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.acceptConnections(listener)
	}()

	return nil
}

// Stop stops the server gracefully
func (s *Server) Stop() error {
	s.mu.Lock()
	if s.isShutdown {
		s.mu.Unlock()
		return nil
	}
	s.isShutdown = true
	s.mu.Unlock()

	s.logger.Println("Initiating graceful shutdown...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Signal shutdown to all components
	s.cancel()
	close(s.shutdownCh)

	// Stop accepting new connections
	s.logger.Println("Stopping peer manager...")
	s.peerMgr.Stop()

	// Wait for all goroutines to finish
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	// Wait for either shutdown timeout or all goroutines to finish
	select {
	case <-shutdownCtx.Done():
		s.logger.Printf("Shutdown timed out after %v", shutdownTimeout)
		return fmt.Errorf("shutdown timed out after %v", shutdownTimeout)
	case <-done:
		s.logger.Printf("Server stopped. Final metrics: connections=%d, messages=%d, errors=%d",
			s.metrics.activeConnections, s.metrics.messagesReceived, s.metrics.errors)
		return nil
	}
}

// recordError records an error in metrics
func (s *Server) recordError(err error) {
	s.metrics.mu.Lock()
	defer s.metrics.mu.Unlock()
	s.metrics.errors++
	s.metrics.lastError = err
	s.metrics.lastErrorTime = time.Now()
	s.logger.Printf("Error: %v", err)
}

// acceptConnections accepts incoming connections
func (s *Server) acceptConnections(listener net.Listener) {
	defer func() {
		if err := listener.Close(); err != nil {
			s.recordError(fmt.Errorf("error closing listener: %v", err))
		}
	}()

	for {
		select {
		case <-s.ctx.Done():
			s.logger.Println("Stopping connection acceptor")
			return
		default:
			// Set deadline for accept
			if err := listener.(*net.TCPListener).SetDeadline(time.Now().Add(1 * time.Second)); err != nil {
				s.recordError(fmt.Errorf("failed to set accept deadline: %v", err))
				continue
			}

			conn, err := listener.Accept()
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				s.recordError(fmt.Errorf("failed to accept connection: %v", err))
				continue
			}

			s.metrics.mu.Lock()
			s.metrics.activeConnections++
			s.metrics.totalConnections++
			s.metrics.mu.Unlock()

			s.logger.Printf("New connection from %s (active: %d, total: %d)",
				conn.RemoteAddr().String(), s.metrics.activeConnections, s.metrics.totalConnections)

			s.wg.Add(1)
			go func() {
				defer s.wg.Done()
				s.handleConnection(conn)
			}()
		}
	}
}

// handleConnection handles a connection
func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			s.recordError(fmt.Errorf("panic in connection handler: %v", r))
		}
		if err := conn.Close(); err != nil {
			s.recordError(fmt.Errorf("error closing connection: %v", err))
		}
		s.metrics.mu.Lock()
		s.metrics.activeConnections--
		s.metrics.mu.Unlock()
		s.logger.Printf("Connection closed from %s (active: %d)",
			conn.RemoteAddr().String(), s.metrics.activeConnections)

		// Remove connection from limiter
		s.peerLimiter.RemoveConnection(conn.RemoteAddr())
	}()

	// Get TLS connection state
	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		log.Printf("Warning: Non-TLS connection from %s", conn.RemoteAddr())
		return
	}

	state := tlsConn.ConnectionState()
	log.Printf("TLS connection established with %s using %s and %s",
		conn.RemoteAddr(),
		security.GetTLSVersionString(state.Version),
		security.GetCipherSuiteString(state.CipherSuite))

	// Check connection limits
	if err := s.peerLimiter.AllowConnection(conn.RemoteAddr()); err != nil {
		s.recordError(fmt.Errorf("connection limit exceeded: %v", err))
		return
	}

	// Set timeouts
	if err := conn.SetDeadline(time.Now().Add(readTimeout)); err != nil {
		s.recordError(fmt.Errorf("failed to set read deadline: %v", err))
		return
	}

	// Create peer info
	info := peers.Info{
		Address:  conn.RemoteAddr().String(),
		Version:  1, // Changed from string to uint32
		LastSeen: time.Now(),
	}

	// Add peer
	s.peerMgr.AddPeer(conn, info)
	s.logger.Printf("Added peer %s (version: %d)", info.Address, info.Version)

	// Handle messages
	for {
		select {
		case <-s.ctx.Done():
			s.logger.Printf("Connection handler shutting down for peer %s", info.Address)
			return
		default:
			// Reset read deadline
			if err := conn.SetDeadline(time.Now().Add(readTimeout)); err != nil {
				s.recordError(fmt.Errorf("failed to set read deadline: %v", err))
				return
			}

			// Read message length
			var length uint32
			if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
				if err != io.EOF {
					s.recordError(fmt.Errorf("error reading message length: %v", err))
				}
				return
			}

			// Validate message size
			if length > maxMessageSize {
				s.recordError(fmt.Errorf("%w: %d bytes (max: %d)", ErrMessageTooLarge, length, maxMessageSize))
				return
			}

			// Read message
			message := make([]byte, length)
			if _, err := io.ReadFull(conn, message); err != nil {
				s.recordError(fmt.Errorf("error reading message: %v", err))
				return
			}

			s.metrics.mu.Lock()
			s.metrics.messagesReceived++
			s.metrics.mu.Unlock()

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
				s.recordError(fmt.Errorf("failed to handle message after %d attempts: %v", maxRetries, handleErr))
				return
			}
		}
	}
}

// handleMessage handles a message
func (s *Server) handleMessage(message []byte) error {
	// Unmarshal signed message
	var signedMsg security.SignedMessage
	if err := json.Unmarshal(message, &signedMsg); err != nil {
		return fmt.Errorf("failed to unmarshal signed message: %v", err)
	}

	// Verify message
	valid, err := s.msgSigner.VerifyMessage(&signedMsg)
	if err != nil {
		return fmt.Errorf("failed to verify message: %v", err)
	}
	if !valid {
		return fmt.Errorf("invalid message signature")
	}

	// Unmarshal actual message
	var msg messages.Message
	if err := json.Unmarshal(signedMsg.Message, &msg); err != nil {
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
		return fmt.Errorf("%w: %s", ErrInvalidMessage, msg.Type)
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
		return fmt.Errorf("%w: %v", ErrInvalidBlock, validateErr)
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
	s.metrics.mu.Lock()
	s.metrics.messagesSent++
	s.metrics.mu.Unlock()
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
		return fmt.Errorf("%w: %v", ErrInvalidTransaction, validateErr)
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
	s.metrics.mu.Lock()
	s.metrics.messagesSent++
	s.metrics.mu.Unlock()
	return nil
}

// BroadcastBlock broadcasts a block
func (s *Server) BroadcastBlock(b *block.Block) error {
	// Determine block type based on mining reward transaction
	blockType := block.GoldenBlock // Default to golden block
	if len(b.Transactions) > 0 {
		// Get the coin type from the first transaction
		coinType := b.Transactions[0].CoinType()
		switch coinType {
		case coin.Leah:
			blockType = block.GoldenBlock
		case coin.Shiblum:
			blockType = block.SilverBlock
		}
	}

	msg := messages.BlockMessage{
		Block:     b,
		BlockType: string(blockType), // Convert BlockType to string
	}

	// Sign message
	signedMsg, err := s.msgSigner.SignMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to sign block message: %v", err)
	}

	// Marshal signed message
	messageData, err := json.Marshal(signedMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal signed message: %v", err)
	}

	s.peerMgr.Broadcast(messageData)
	return nil
}

// BroadcastTransaction broadcasts a transaction
func (s *Server) BroadcastTransaction(tx *types.Transaction) error {
	// Create a common.Transaction wrapper
	commonTx := common.NewTransaction(
		[]byte(tx.Inputs[0].Address),  // From address
		[]byte(tx.Outputs[0].Address), // To address
		tx.Outputs[0].Value,           // Amount
		tx.Data,                       // Data
	)

	msg := messages.TransactionMessage{
		Transaction: commonTx,
		CoinType:    string(tx.CoinType), // Convert CoinType to string
	}

	// Sign message
	signedMsg, err := s.msgSigner.SignMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to sign transaction message: %v", err)
	}

	// Marshal signed message
	messageData, err := json.Marshal(signedMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal signed message: %v", err)
	}

	s.peerMgr.Broadcast(messageData)
	return nil
}
