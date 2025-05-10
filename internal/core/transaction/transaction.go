package transaction

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/youngchain/internal/core/coin"
	"github.com/youngchain/internal/core/common"
	"github.com/youngchain/internal/core/types"
)

var (
	ErrInvalidTransaction = errors.New("invalid transaction")
	ErrInsufficientFunds  = errors.New("insufficient funds")
	ErrInvalidSignature   = errors.New("invalid signature")
	ErrDoubleSpend        = errors.New("double spend detected")
)

// Transaction represents a cryptocurrency transaction
type Transaction struct {
	Version  uint32        `json:"version"`
	Inputs   []*Input      `json:"inputs"`
	Outputs  []*Output     `json:"outputs"`
	LockTime uint32        `json:"lockTime"`
	Fee      uint64        `json:"fee"`
	CoinType coin.CoinType `json:"coinType"`
}

// Input represents a transaction input
type Input struct {
	PreviousTxHash  []byte `json:"previousTxHash"`
	PreviousTxIndex uint32 `json:"previousTxIndex"`
	ScriptSig       []byte `json:"scriptSig"`
	Sequence        uint32 `json:"sequence"`
}

// Output represents a transaction output
type Output struct {
	Value        uint64 `json:"value"`
	ScriptPubKey []byte `json:"scriptPubKey"`
	Address      string `json:"address"`
}

// UTXO represents an unspent transaction output
type UTXO struct {
	TxHash    []byte        `json:"txHash"`
	TxIndex   uint32        `json:"txIndex"`
	Value     uint64        `json:"value"`
	Address   string        `json:"address"`
	CoinType  coin.CoinType `json:"coinType"`
	IsSpent   bool          `json:"isSpent"`
	BlockHash []byte        `json:"blockHash"`
}

// TransactionPool manages pending transactions
type TransactionPool struct {
	transactions map[string]*types.Transaction
	utxoSet      map[string]*types.UTXO
	maxSize      int
}

// NewTransactionPool creates a new transaction pool
func NewTransactionPool(maxSize int) *TransactionPool {
	return &TransactionPool{
		transactions: make(map[string]*types.Transaction),
		utxoSet:      make(map[string]*types.UTXO),
		maxSize:      maxSize,
	}
}

// CreateTransaction creates a new transaction
func CreateTransaction(inputs []*types.Input, outputs []*types.Output, fee uint64, coinType coin.CoinType) *types.Transaction {
	return &types.Transaction{
		Version:  1,
		Inputs:   inputs,
		Outputs:  outputs,
		LockTime: uint32(time.Now().Unix()),
		Fee:      fee,
		CoinType: coinType,
	}
}

// CalculateHash calculates the transaction hash
func (tx *Transaction) CalculateHash() []byte {
	var buf bytes.Buffer

	// Write version
	binary.Write(&buf, binary.LittleEndian, tx.Version)

	// Write inputs
	for _, input := range tx.Inputs {
		buf.Write(input.PreviousTxHash)
		binary.Write(&buf, binary.LittleEndian, input.PreviousTxIndex)
		buf.Write(input.ScriptSig)
		binary.Write(&buf, binary.LittleEndian, input.Sequence)
	}

	// Write outputs
	for _, output := range tx.Outputs {
		binary.Write(&buf, binary.LittleEndian, output.Value)
		buf.Write(output.ScriptPubKey)
	}

	// Write lock time and fee
	binary.Write(&buf, binary.LittleEndian, tx.LockTime)
	binary.Write(&buf, binary.LittleEndian, tx.Fee)

	// Write coin type
	buf.Write([]byte(tx.CoinType))

	// Calculate hash
	hash := sha256.Sum256(buf.Bytes())
	return hash[:]
}

// Validate validates a transaction
func (tp *TransactionPool) Validate(tx *types.Transaction) error {
	// Check basic structure
	if len(tx.Inputs) == 0 || len(tx.Outputs) == 0 {
		return fmt.Errorf("%w: empty inputs or outputs", ErrInvalidTransaction)
	}

	// Check for double spends
	spentOutputs := make(map[string]bool)
	var totalInput uint64
	var totalOutput uint64

	// Validate inputs
	for _, input := range tx.Inputs {
		utxoKey := fmt.Sprintf("%x:%d", input.PreviousTxHash, input.PreviousTxIndex)

		// Check if UTXO exists
		utxo, exists := tp.utxoSet[utxoKey]
		if !exists {
			return fmt.Errorf("%w: UTXO not found", ErrInvalidTransaction)
		}

		// Check if already spent
		if spentOutputs[utxoKey] {
			return fmt.Errorf("%w: double spend detected", ErrInvalidTransaction)
		}

		// Check coin type
		if utxo.CoinType != tx.CoinType {
			return fmt.Errorf("%w: coin type mismatch", ErrInvalidTransaction)
		}

		// Add to total input
		totalInput += utxo.Value
		spentOutputs[utxoKey] = true
	}

	// Validate outputs
	for _, output := range tx.Outputs {
		if output.Value == 0 {
			return fmt.Errorf("%w: zero value output", ErrInvalidTransaction)
		}
		totalOutput += output.Value
	}

	// Check if inputs cover outputs plus fee
	if totalInput < totalOutput+tx.Fee {
		return fmt.Errorf("%w: insufficient funds", ErrInvalidTransaction)
	}

	return nil
}

// AddToPool adds a transaction to the pool
func (tp *TransactionPool) AddToPool(tx *types.Transaction) error {
	// Check pool size
	if len(tp.transactions) >= tp.maxSize {
		return fmt.Errorf("transaction pool full")
	}

	// Validate transaction
	if err := tp.Validate(tx); err != nil {
		return err
	}

	// Add to pool
	txHash := tx.CalculateHash()
	tp.transactions[string(txHash)] = tx

	// Update UTXO set
	for _, input := range tx.Inputs {
		utxoKey := fmt.Sprintf("%x:%d", input.PreviousTxHash, input.PreviousTxIndex)
		if utxo, exists := tp.utxoSet[utxoKey]; exists {
			utxo.IsSpent = true
		}
	}

	return nil
}

// RemoveFromPool removes a transaction from the pool
func (tp *TransactionPool) RemoveFromPool(txHash []byte) {
	delete(tp.transactions, string(txHash))
}

// GetTransaction gets a transaction from the pool
func (tp *TransactionPool) GetTransaction(txHash []byte) *types.Transaction {
	return tp.transactions[string(txHash)]
}

// GetUTXO gets a UTXO from the pool
func (tp *TransactionPool) GetUTXO(txHash []byte, index uint32) *types.UTXO {
	utxoKey := fmt.Sprintf("%x:%d", txHash, index)
	return tp.utxoSet[utxoKey]
}

// AddUTXO adds a UTXO to the pool
func (tp *TransactionPool) AddUTXO(utxo *types.UTXO) {
	utxoKey := fmt.Sprintf("%x:%d", utxo.TxHash, utxo.TxIndex)
	tp.utxoSet[utxoKey] = utxo
}

// RemoveUTXO removes a UTXO from the pool
func (tp *TransactionPool) RemoveUTXO(txHash []byte, index uint32) {
	utxoKey := fmt.Sprintf("%x:%d", txHash, index)
	delete(tp.utxoSet, utxoKey)
}

// GetBalance gets the balance for an address
func (tp *TransactionPool) GetBalance(address string, coinType coin.CoinType) uint64 {
	var balance uint64
	for _, utxo := range tp.utxoSet {
		if utxo.Address == address && utxo.CoinType == coinType && !utxo.IsSpent {
			balance += utxo.Value
		}
	}
	return balance
}

// Hash returns the transaction hash
func (tx *Transaction) Hash() common.Hash {
	return common.BytesToHash(tx.CalculateHash())
}

// FeeRate calculates the fee rate in satoshis per byte
func (tx *Transaction) FeeRate() uint64 {
	size := tx.Size()
	if size == 0 {
		return 0
	}
	return tx.Fee / uint64(size)
}

// Size returns the transaction size in bytes
func (tx *Transaction) Size() int {
	var size int

	// Version
	size += 4

	// Input count and inputs
	size += 4 // Input count
	for _, input := range tx.Inputs {
		size += 32 // PreviousTxHash
		size += 4  // PreviousTxIndex
		size += 4  // ScriptSig length
		size += len(input.ScriptSig)
		size += 4 // Sequence
	}

	// Output count and outputs
	size += 4 // Output count
	for _, output := range tx.Outputs {
		size += 8 // Value
		size += 4 // ScriptPubKey length
		size += len(output.ScriptPubKey)
	}

	// LockTime and Fee
	size += 4 // LockTime
	size += 8 // Fee

	// CoinType
	size += len(tx.CoinType)

	return size
}

// MarshalJSON implements custom JSON marshaling for Transaction
func (tx *Transaction) MarshalJSON() ([]byte, error) {
	type Alias Transaction
	return json.Marshal(&struct {
		*Alias
		Hash string `json:"hash"`
	}{
		Alias: (*Alias)(tx),
		Hash:  fmt.Sprintf("%x", tx.CalculateHash()),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for Transaction
func (tx *Transaction) UnmarshalJSON(data []byte) error {
	type Alias Transaction
	aux := &struct {
		*Alias
		Hash string `json:"hash"`
	}{
		Alias: (*Alias)(tx),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	return nil
}
