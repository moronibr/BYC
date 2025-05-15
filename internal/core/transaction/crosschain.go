package transaction

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/coin"
	"github.com/youngchain/internal/core/types"
)

// CrossChainTransfer represents a transfer between chains
type CrossChainTransfer struct {
	// Source chain transaction
	SourceTx *types.Transaction

	// Target chain transaction
	TargetTx *types.Transaction

	// Transfer amount
	Amount uint64

	// Source chain block
	SourceBlock *block.Block

	// Target chain block
	TargetBlock *block.Block

	// Transfer status
	Status TransferStatus

	// Creation timestamp
	CreatedAt time.Time

	// Completion timestamp
	CompletedAt time.Time
}

// TransferStatus represents the status of a cross-chain transfer
type TransferStatus string

const (
	TransferPending   TransferStatus = "pending"
	TransferConfirmed TransferStatus = "confirmed"
	TransferFailed    TransferStatus = "failed"
)

// CrossChainManager manages cross-chain transfers
type CrossChainManager struct {
	// Pending transfers
	pendingTransfers map[string]*CrossChainTransfer

	// Confirmed transfers
	confirmedTransfers map[string]*CrossChainTransfer

	// Failed transfers
	failedTransfers map[string]*CrossChainTransfer

	// Transfer mutex
	mu sync.RWMutex
}

// NewCrossChainManager creates a new cross-chain manager
func NewCrossChainManager() *CrossChainManager {
	return &CrossChainManager{
		pendingTransfers:   make(map[string]*CrossChainTransfer),
		confirmedTransfers: make(map[string]*CrossChainTransfer),
		failedTransfers:    make(map[string]*CrossChainTransfer),
	}
}

// CreateTransfer creates a new cross-chain transfer
func (m *CrossChainManager) CreateTransfer(sourceTx *types.Transaction, targetChain block.BlockType) (*CrossChainTransfer, error) {
	// Validate source transaction
	if err := m.validateSourceTransaction(sourceTx); err != nil {
		return nil, err
	}

	// Create transfer
	transfer := &CrossChainTransfer{
		SourceTx:  sourceTx,
		Amount:    sourceTx.Outputs[0].Value,
		Status:    TransferPending,
		CreatedAt: time.Now(),
	}

	// Generate transfer ID
	transferID := m.generateTransferID(transfer)

	// Store transfer
	m.mu.Lock()
	m.pendingTransfers[transferID] = transfer
	m.mu.Unlock()

	return transfer, nil
}

// validateSourceTransaction validates a source transaction
func (m *CrossChainManager) validateSourceTransaction(tx *types.Transaction) error {
	// Validate transaction structure
	if len(tx.Inputs) == 0 || len(tx.Outputs) == 0 {
		return fmt.Errorf("invalid transaction structure")
	}

	// Validate amount
	if tx.Outputs[0].Value == 0 {
		return fmt.Errorf("invalid transfer amount")
	}

	// Validate coin type
	if tx.CoinType != coin.Antion {
		return fmt.Errorf("only Antion coins can be transferred between chains")
	}

	return nil
}

// generateTransferID generates a unique ID for a transfer
func (m *CrossChainManager) generateTransferID(transfer *CrossChainTransfer) string {
	// Create a unique identifier using source transaction hash and timestamp
	data := append(transfer.SourceTx.Hash(), []byte(transfer.CreatedAt.String())...)
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

// ConfirmTransfer confirms a cross-chain transfer
func (m *CrossChainManager) ConfirmTransfer(transferID string, targetTx *types.Transaction, targetBlock *block.Block) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get transfer
	transfer, ok := m.pendingTransfers[transferID]
	if !ok {
		return fmt.Errorf("transfer not found")
	}

	// Validate target transaction
	if err := m.validateTargetTransaction(targetTx, transfer); err != nil {
		return err
	}

	// Update transfer
	transfer.TargetTx = targetTx
	transfer.TargetBlock = targetBlock
	transfer.Status = TransferConfirmed
	transfer.CompletedAt = time.Now()

	// Move transfer to confirmed
	delete(m.pendingTransfers, transferID)
	m.confirmedTransfers[transferID] = transfer

	return nil
}

// validateTargetTransaction validates a target transaction
func (m *CrossChainManager) validateTargetTransaction(targetTx *types.Transaction, transfer *CrossChainTransfer) error {
	// Validate transaction structure
	if len(targetTx.Inputs) == 0 || len(targetTx.Outputs) == 0 {
		return fmt.Errorf("invalid transaction structure")
	}

	// Validate amount
	if targetTx.Outputs[0].Value != transfer.Amount {
		return fmt.Errorf("invalid transfer amount")
	}

	// Validate coin type
	if targetTx.CoinType != coin.Antion {
		return fmt.Errorf("invalid coin type")
	}

	return nil
}

// FailTransfer marks a transfer as failed
func (m *CrossChainManager) FailTransfer(transferID string, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get transfer
	transfer, ok := m.pendingTransfers[transferID]
	if !ok {
		return fmt.Errorf("transfer not found")
	}

	// Update transfer
	transfer.Status = TransferFailed
	transfer.CompletedAt = time.Now()

	// Move transfer to failed
	delete(m.pendingTransfers, transferID)
	m.failedTransfers[transferID] = transfer

	return nil
}

// GetTransfer gets a transfer by ID
func (m *CrossChainManager) GetTransfer(transferID string) (*CrossChainTransfer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check pending transfers
	if transfer, ok := m.pendingTransfers[transferID]; ok {
		return transfer, nil
	}

	// Check confirmed transfers
	if transfer, ok := m.confirmedTransfers[transferID]; ok {
		return transfer, nil
	}

	// Check failed transfers
	if transfer, ok := m.failedTransfers[transferID]; ok {
		return transfer, nil
	}

	return nil, fmt.Errorf("transfer not found")
}

// GetPendingTransfers gets all pending transfers
func (m *CrossChainManager) GetPendingTransfers() []*CrossChainTransfer {
	m.mu.RLock()
	defer m.mu.RUnlock()

	transfers := make([]*CrossChainTransfer, 0, len(m.pendingTransfers))
	for _, transfer := range m.pendingTransfers {
		transfers = append(transfers, transfer)
	}

	return transfers
}

// GetConfirmedTransfers gets all confirmed transfers
func (m *CrossChainManager) GetConfirmedTransfers() []*CrossChainTransfer {
	m.mu.RLock()
	defer m.mu.RUnlock()

	transfers := make([]*CrossChainTransfer, 0, len(m.confirmedTransfers))
	for _, transfer := range m.confirmedTransfers {
		transfers = append(transfers, transfer)
	}

	return transfers
}

// GetFailedTransfers gets all failed transfers
func (m *CrossChainManager) GetFailedTransfers() []*CrossChainTransfer {
	m.mu.RLock()
	defer m.mu.RUnlock()

	transfers := make([]*CrossChainTransfer, 0, len(m.failedTransfers))
	for _, transfer := range m.failedTransfers {
		transfers = append(transfers, transfer)
	}

	return transfers
}
