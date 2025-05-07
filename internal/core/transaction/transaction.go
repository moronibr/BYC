package transaction

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/youngchain/internal/core/coin"
)

var (
	ErrInvalidTransaction = errors.New("invalid transaction")
	ErrInsufficientFunds  = errors.New("insufficient funds")
	ErrInvalidSignature   = errors.New("invalid signature")
	ErrDoubleSpend        = errors.New("double spend detected")
)

// Transaction represents a cryptocurrency transaction
type Transaction struct {
	Version  uint32
	Inputs   []*Input
	Outputs  []*Output
	LockTime uint32
	Fee      uint64
	CoinType coin.CoinType
}

// Input represents a transaction input
type Input struct {
	PreviousTxHash  []byte
	PreviousTxIndex uint32
	ScriptSig       []byte
	Sequence        uint32
}

// Output represents a transaction output
type Output struct {
	Value        uint64
	ScriptPubKey []byte
	Address      string
}

// UTXO represents an unspent transaction output
type UTXO struct {
	TxHash    []byte
	TxIndex   uint32
	Value     uint64
	Address   string
	CoinType  coin.CoinType
	IsSpent   bool
	BlockHash []byte
}

// TransactionPool manages pending transactions
type TransactionPool struct {
	transactions map[string]*Transaction
	utxoSet      map[string]*UTXO
	maxSize      int
}

// NewTransactionPool creates a new transaction pool
func NewTransactionPool(maxSize int) *TransactionPool {
	return &TransactionPool{
		transactions: make(map[string]*Transaction),
		utxoSet:      make(map[string]*UTXO),
		maxSize:      maxSize,
	}
}

// CreateTransaction creates a new transaction
func CreateTransaction(inputs []*Input, outputs []*Output, fee uint64, coinType coin.CoinType) *Transaction {
	return &Transaction{
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
func (tx *Transaction) Validate(utxoSet map[string]*UTXO) error {
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
		utxo, exists := utxoSet[utxoKey]
		if !exists {
			return fmt.Errorf("%w: UTXO not found", ErrInvalidTransaction)
		}

		// Check if already spent
		if spentOutputs[utxoKey] {
			return fmt.Errorf("%w: double spend detected", ErrDoubleSpend)
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
		return fmt.Errorf("%w: insufficient funds", ErrInsufficientFunds)
	}

	return nil
}

// AddToPool adds a transaction to the pool
func (tp *TransactionPool) AddToPool(tx *Transaction) error {
	// Check pool size
	if len(tp.transactions) >= tp.maxSize {
		return fmt.Errorf("transaction pool full")
	}

	// Validate transaction
	if err := tx.Validate(tp.utxoSet); err != nil {
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
func (tp *TransactionPool) GetTransaction(txHash []byte) *Transaction {
	return tp.transactions[string(txHash)]
}

// GetUTXO gets a UTXO from the set
func (tp *TransactionPool) GetUTXO(txHash []byte, index uint32) *UTXO {
	utxoKey := fmt.Sprintf("%x:%d", txHash, index)
	return tp.utxoSet[utxoKey]
}

// AddUTXO adds a UTXO to the set
func (tp *TransactionPool) AddUTXO(utxo *UTXO) {
	utxoKey := fmt.Sprintf("%x:%d", utxo.TxHash, utxo.TxIndex)
	tp.utxoSet[utxoKey] = utxo
}

// RemoveUTXO removes a UTXO from the set
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
