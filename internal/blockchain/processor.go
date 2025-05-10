package blockchain

import (
	"fmt"
	"sync"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/wallet"
)

// Processor represents the transaction processor
type Processor struct {
	storage *Storage
	state   *State
	mu      sync.RWMutex
}

// State represents the blockchain state
type State struct {
	Balances map[string]uint64
	Nonces   map[string]uint64
	mu       sync.RWMutex
}

// NewProcessor creates a new transaction processor
func NewProcessor(storage *Storage) *Processor {
	return &Processor{
		storage: storage,
		state: &State{
			Balances: make(map[string]uint64),
			Nonces:   make(map[string]uint64),
		},
	}
}

// ProcessTransaction processes a single transaction
func (p *Processor) ProcessTransaction(tx *block.Transaction) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Verify transaction signature
	if !tx.VerifySignature() {
		return fmt.Errorf("invalid transaction signature")
	}

	// Check sender's balance
	senderBalance := p.state.GetBalance(tx.From)
	if senderBalance < tx.Amount {
		return fmt.Errorf("insufficient balance")
	}

	// Check nonce
	senderNonce := p.state.GetNonce(tx.From)
	if tx.Nonce != senderNonce {
		return fmt.Errorf("invalid nonce")
	}

	// Update balances
	p.state.UpdateBalance(tx.From, senderBalance-tx.Amount)
	receiverBalance := p.state.GetBalance(tx.To)
	p.state.UpdateBalance(tx.To, receiverBalance+tx.Amount)

	// Update nonce
	p.state.IncrementNonce(tx.From)

	return nil
}

// ProcessBlock processes a block of transactions
func (p *Processor) ProcessBlock(block *block.Block) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Process each transaction in the block
	for _, tx := range block.Transactions {
		if err := p.ProcessTransaction(tx); err != nil {
			return fmt.Errorf("failed to process transaction: %v", err)
		}
	}

	// Store the block
	if err := p.storage.StoreBlock(block); err != nil {
		return fmt.Errorf("failed to store block: %v", err)
	}

	return nil
}

// GetBalance returns the balance of an address
func (s *State) GetBalance(address string) uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	balance, exists := s.Balances[address]
	if !exists {
		return 0
	}
	return balance
}

// UpdateBalance updates the balance of an address
func (s *State) UpdateBalance(address string, balance uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Balances[address] = balance
}

// GetNonce returns the nonce of an address
func (s *State) GetNonce(address string) uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	nonce, exists := s.Nonces[address]
	if !exists {
		return 0
	}
	return nonce
}

// IncrementNonce increments the nonce of an address
func (s *State) IncrementNonce(address string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Nonces[address]++
}

// VerifyTransaction verifies a transaction
func (tx *block.Transaction) VerifySignature() bool {
	// Implement signature verification logic
	return true
}

// CreateTransaction creates a new transaction
func CreateTransaction(from, to string, amount, nonce uint64, wallet *wallet.Wallet) (*block.Transaction, error) {
	tx := &block.Transaction{
		From:   from,
		To:     to,
		Amount: amount,
		Nonce:  nonce,
	}

	// Sign the transaction
	if err := wallet.SignTransaction(tx); err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %v", err)
	}

	return tx, nil
}
