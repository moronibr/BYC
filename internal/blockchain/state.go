package blockchain

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/common"
	"github.com/youngchain/internal/core/script"
	"github.com/youngchain/internal/core/utxo"
)

// StateManager manages the blockchain state
type StateManager struct {
	state     *State
	storage   *Storage
	utxoSet   *utxo.UTXOSet
	processor *Processor
	mu        sync.RWMutex
}

// State represents the current blockchain state
type State struct {
	Balances     map[string]uint64
	Nonces       map[string]uint64
	ContractData map[string][]byte
	mu           sync.RWMutex
}

// NewStateManager creates a new state manager
func NewStateManager(storage *Storage, processor *Processor) *StateManager {
	return &StateManager{
		state: &State{
			Balances:     make(map[string]uint64),
			Nonces:       make(map[string]uint64),
			ContractData: make(map[string][]byte),
		},
		storage:   storage,
		utxoSet:   utxo.NewUTXOSet(),
		processor: processor,
	}
}

// ApplyBlock applies a block to the state
func (sm *StateManager) ApplyBlock(block *block.Block) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Process each transaction in the block
	for _, tx := range block.Transactions {
		if err := sm.processTransaction(tx, block.Header.Height); err != nil {
			return fmt.Errorf("failed to process transaction: %v", err)
		}
	}

	// Update state in storage
	if err := sm.persistState(); err != nil {
		return fmt.Errorf("failed to persist state: %v", err)
	}

	return nil
}

// processTransaction processes a single transaction
func (sm *StateManager) processTransaction(tx *common.Transaction, blockHeight uint64) error {
	// Validate transaction
	if err := tx.Validate(); err != nil {
		return fmt.Errorf("invalid transaction: %v", err)
	}

	// Skip coinbase transactions
	if tx.IsCoinbase() {
		return nil
	}

	// Calculate total input amount
	var totalInput uint64
	inputs := tx.Inputs()
	for _, input := range inputs {
		// Get UTXO
		utxo, exists := sm.utxoSet.GetUTXO(input.PreviousTxHash, input.PreviousTxIndex)
		if !exists {
			return fmt.Errorf("input not found: %x:%d", input.PreviousTxHash, input.PreviousTxIndex)
		}

		// Add to total input
		totalInput += utxo.Amount
	}

	// Calculate total output amount
	var totalOutput uint64
	outputs := tx.Outputs()
	for _, output := range outputs {
		totalOutput += output.Value
	}

	// Check if inputs cover outputs
	if totalInput < totalOutput {
		return fmt.Errorf("insufficient input amount")
	}

	// Update UTXO set
	for _, input := range inputs {
		// Mark input as spent
		sm.utxoSet.RemoveUTXO(input.PreviousTxHash, input.PreviousTxIndex)
	}

	// Add new UTXOs
	for i, output := range outputs {
		// Create script from public key
		script := script.CreateP2PKHScript(output.ScriptPubKey)
		if err := script.Validate(); err != nil {
			return fmt.Errorf("invalid script: %v", err)
		}

		utxo := &utxo.UTXO{
			TxHash:      tx.Hash(),
			OutIndex:    uint32(i),
			Amount:      output.Value,
			ScriptPub:   script,
			IsCoinbase:  tx.IsCoinbase(),
			BlockHeight: blockHeight,
		}
		if err := sm.utxoSet.AddUTXO(utxo); err != nil {
			return fmt.Errorf("failed to add UTXO: %v", err)
		}
	}

	return nil
}

// persistState persists the current state to storage
func (sm *StateManager) persistState() error {
	stateData, err := json.Marshal(sm.state)
	if err != nil {
		return err
	}

	return sm.storage.UpdateState([]byte("current_state"), stateData)
}

// LoadState loads the state from storage
func (sm *StateManager) LoadState() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	stateData, err := sm.storage.GetState([]byte("current_state"))
	if err != nil {
		return err
	}

	if stateData == nil {
		// Initialize empty state
		sm.state = &State{
			Balances:     make(map[string]uint64),
			Nonces:       make(map[string]uint64),
			ContractData: make(map[string][]byte),
		}
		return nil
	}

	return json.Unmarshal(stateData, sm.state)
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

// GetContractData returns the data for a contract
func (s *State) GetContractData(address string) []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.ContractData[address]
}

// UpdateContractData updates the data for a contract
func (s *State) UpdateContractData(address string, data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ContractData[address] = data
}

// GetState returns the current state
func (sm *StateManager) GetState() *State {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.state
}
