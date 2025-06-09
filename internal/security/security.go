package security

import (
	"fmt"
	"sync"
	"time"

	"byc/internal/blockchain"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// SecurityManager handles all security-related features
type SecurityManager struct {
	// Rate limiting
	apiRateLimiters  map[string]*rate.Limiter // endpoint -> limiter
	peerRateLimiters map[string]*rate.Limiter // peer address -> limiter
	rateLimitMu      sync.RWMutex

	// Connection limits
	maxPeers        int
	peerConnections map[string]time.Time // peer address -> connection time
	connectionMu    sync.RWMutex

	// Block size limits
	maxBlockSize            int64
	maxTransactionsPerBlock int

	// Transaction validation
	minTransactionFee  float64
	maxTransactionSize int64

	// Network message validation
	maxMessageSize      int64
	allowedMessageTypes map[string]bool

	// Logging
	logger *zap.Logger
}

// NewSecurityManager creates a new security manager with default settings
func NewSecurityManager() *SecurityManager {
	return &SecurityManager{
		apiRateLimiters:         make(map[string]*rate.Limiter),
		peerRateLimiters:        make(map[string]*rate.Limiter),
		peerConnections:         make(map[string]time.Time),
		maxPeers:                50,          // Default max peers
		maxBlockSize:            1024 * 1024, // 1MB default
		maxTransactionsPerBlock: 1000,
		minTransactionFee:       1000.0,
		maxTransactionSize:      1024 * 10,  // 10KB default
		maxMessageSize:          1024 * 100, // 100KB default
		allowedMessageTypes: map[string]bool{
			"block":       true,
			"transaction": true,
			"peer_list":   true,
			"get_blocks":  true,
			"get_peers":   true,
		},
		logger: zap.NewNop(), // Use a no-op logger by default
	}
}

// Rate Limiting

// AddAPIRateLimit adds a rate limit for an API endpoint
func (s *SecurityManager) AddAPIRateLimit(endpoint string, rps float64, burst int) {
	s.rateLimitMu.Lock()
	defer s.rateLimitMu.Unlock()

	s.apiRateLimiters[endpoint] = rate.NewLimiter(rate.Limit(rps), burst)
}

// AddPeerRateLimit adds a rate limit for a peer
func (s *SecurityManager) AddPeerRateLimit(peerAddr string, rps float64, burst int) {
	s.rateLimitMu.Lock()
	defer s.rateLimitMu.Unlock()

	s.peerRateLimiters[peerAddr] = rate.NewLimiter(rate.Limit(rps), burst)
}

// CheckAPIRateLimit checks if an API request is within rate limits
func (s *SecurityManager) CheckAPIRateLimit(endpoint string) error {
	s.rateLimitMu.RLock()
	limiter, exists := s.apiRateLimiters[endpoint]
	s.rateLimitMu.RUnlock()

	if !exists {
		return nil // No rate limit set for this endpoint
	}

	if !limiter.Allow() {
		return fmt.Errorf("rate limit exceeded for endpoint: %s", endpoint)
	}

	return nil
}

// CheckPeerRateLimit checks if a peer's request is within rate limits
func (s *SecurityManager) CheckPeerRateLimit(peerAddr string) error {
	s.rateLimitMu.RLock()
	limiter, exists := s.peerRateLimiters[peerAddr]
	s.rateLimitMu.RUnlock()

	if !exists {
		return nil // No rate limit set for this peer
	}

	if !limiter.Allow() {
		return fmt.Errorf("rate limit exceeded for peer: %s", peerAddr)
	}

	return nil
}

// Connection Limits

// CheckPeerConnection checks if a new peer connection is allowed
func (s *SecurityManager) CheckPeerConnection(peerAddr string) error {
	s.connectionMu.Lock()
	defer s.connectionMu.Unlock()

	// Check if we've reached the maximum number of peers
	if len(s.peerConnections) >= s.maxPeers {
		return fmt.Errorf("maximum number of peers reached")
	}

	// Check if this peer is already connected
	if _, exists := s.peerConnections[peerAddr]; exists {
		return fmt.Errorf("peer already connected: %s", peerAddr)
	}

	// Add the new peer connection
	s.peerConnections[peerAddr] = time.Now()
	return nil
}

// RemovePeerConnection removes a peer connection
func (s *SecurityManager) RemovePeerConnection(peerAddr string) {
	s.connectionMu.Lock()
	defer s.connectionMu.Unlock()

	delete(s.peerConnections, peerAddr)
}

// Block Validation

// ValidateBlock validates a block against security rules
func (s *SecurityManager) ValidateBlock(block *blockchain.Block) error {
	// Check block size
	blockSize := len(block.Hash) + len(block.PrevHash)
	for _, tx := range block.Transactions {
		blockSize += len(tx.ID)
	}
	if int64(blockSize) > s.maxBlockSize {
		return fmt.Errorf("block size exceeds maximum allowed size")
	}

	// Check number of transactions
	if len(block.Transactions) > s.maxTransactionsPerBlock {
		return fmt.Errorf("block contains too many transactions")
	}

	return nil
}

// Transaction Validation

// ValidateTransaction validates a transaction against security rules
func (s *SecurityManager) ValidateTransaction(tx *blockchain.Transaction) error {
	// Check transaction size
	txSize := len(tx.ID)
	for _, input := range tx.Inputs {
		txSize += len(input.TxID) + len(input.PublicKey) + len(input.Signature)
	}
	for _, output := range tx.Outputs {
		txSize += len(output.PublicKeyHash)
	}
	if int64(txSize) > s.maxTransactionSize {
		return fmt.Errorf("transaction size exceeds maximum allowed size")
	}

	// Check minimum transaction fee
	fee := tx.GetFee()
	if fee < s.minTransactionFee {
		return fmt.Errorf("transaction fee below minimum required")
	}

	return nil
}

// Network Message Validation

// ValidateMessage validates a network message
func (s *SecurityManager) ValidateMessage(msgType string, msgSize int64) error {
	// Check message type
	if !s.allowedMessageTypes[msgType] {
		return fmt.Errorf("invalid message type: %s", msgType)
	}

	// Check message size
	if msgSize > s.maxMessageSize {
		return fmt.Errorf("message size exceeds maximum allowed size")
	}

	return nil
}

// Cleanup removes old peer connections and rate limiters
func (s *SecurityManager) Cleanup() {
	s.connectionMu.Lock()
	defer s.connectionMu.Unlock()

	// Remove peer connections older than 24 hours
	cutoff := time.Now().Add(-24 * time.Hour)
	for addr, connTime := range s.peerConnections {
		if connTime.Before(cutoff) {
			delete(s.peerConnections, addr)
		}
	}

	// Remove rate limiters for disconnected peers
	s.rateLimitMu.Lock()
	defer s.rateLimitMu.Unlock()

	for addr := range s.peerRateLimiters {
		if _, exists := s.peerConnections[addr]; !exists {
			delete(s.peerRateLimiters, addr)
		}
	}
}

// StartCleanup starts the periodic cleanup routine
func (s *SecurityManager) StartCleanup() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			s.Cleanup()
		}
	}()
}
