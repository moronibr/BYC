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
	state   *StateManager
	mu      sync.RWMutex
}

// NewProcessor creates a new transaction processor
func NewProcessor(storage *Storage, state *StateManager) *Processor {
	return &Processor{
		storage: storage,
		state:   state,
	}
}

// ProcessTransaction processes a single transaction
func (p *Processor) ProcessTransaction(tx *block.Transaction) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Verify transaction signature
	if !tx.VerifySignature(nil) { // TODO: Get public key from address
		return fmt.Errorf("invalid transaction signature")
	}

	// Check sender's balance
	senderBalance := p.state.GetState().GetBalance(tx.From)
	if senderBalance < tx.Amount {
		return fmt.Errorf("insufficient balance")
	}

	// Check nonce
	senderNonce := p.state.GetState().GetNonce(tx.From)
	if tx.Nonce != senderNonce {
		return fmt.Errorf("invalid nonce")
	}

	// Update balances
	p.state.GetState().UpdateBalance(tx.From, senderBalance-tx.Amount)
	receiverBalance := p.state.GetState().GetBalance(tx.To)
	p.state.GetState().UpdateBalance(tx.To, receiverBalance+tx.Amount)

	// Update nonce
	p.state.GetState().IncrementNonce(tx.From)

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

// CreateTransaction creates a new transaction
func CreateTransaction(from, to string, amount, nonce uint64, wallet *wallet.Wallet) (*block.Transaction, error) {
	return wallet.CreateTransaction(to, amount, nonce)
}
