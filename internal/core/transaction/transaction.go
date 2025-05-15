package transaction

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/youngchain/internal/core/coin"
	"github.com/youngchain/internal/core/common"
	"github.com/youngchain/internal/core/types"
)

var (
	ErrInvalidTransaction = fmt.Errorf("invalid transaction")
	ErrInsufficientFunds  = fmt.Errorf("insufficient funds")
	ErrInvalidSignature   = fmt.Errorf("invalid signature")
	ErrDoubleSpend        = fmt.Errorf("double spend detected")
)

// Transaction represents a cryptocurrency transaction
type Transaction struct {
	Version   uint32    `json:"version"`
	Inputs    []Input   `json:"inputs"`
	Outputs   []Output  `json:"outputs"`
	LockTime  uint32    `json:"lockTime"`
	Fee       int64     `json:"fee"`
	CoinType  coin.Type `json:"coinType"`
	HashBytes []byte    `json:"hash"`
	Signature []byte    `json:"signature"`
	Data      []byte    `json:"data"`
	Witness   [][]byte  `json:"witness"`
}

// Input represents a transaction input
type Input struct {
	PreviousTxHash []byte `json:"previousTxHash"`
	OutputIndex    uint32 `json:"outputIndex"`
	ScriptSig      []byte `json:"scriptSig"`
	Sequence       uint32 `json:"sequence"`
}

// Output represents a transaction output
type Output struct {
	Value        int64  `json:"value"`
	ScriptPubKey []byte `json:"scriptPubKey"`
}

// UTXO represents an unspent transaction output
type UTXO struct {
	TxHash       []byte    `json:"txHash"`
	OutputIndex  uint32    `json:"txIndex"`
	Value        int64     `json:"value"`
	ScriptPubKey []byte    `json:"scriptPubKey"`
	Spent        bool      `json:"isSpent"`
	Address      string    `json:"address"`
	CoinType     coin.Type `json:"coinType"`
	BlockHash    []byte    `json:"blockHash"`
}

// TransactionPool manages pending transactions
type TransactionPool struct {
	transactions map[string]*types.Transaction
	utxos        map[string]*types.UTXO
	mu           sync.RWMutex
	maxSize      int
}

// NewTransactionPool creates a new transaction pool
func NewTransactionPool(maxSize int) *TransactionPool {
	return &TransactionPool{
		transactions: make(map[string]*types.Transaction),
		utxos:        make(map[string]*types.UTXO),
		maxSize:      maxSize,
	}
}

// NewTransaction creates a new transaction
func NewTransaction(version uint32, coinType coin.Type) *Transaction {
	return &Transaction{
		Version:  version,
		Inputs:   make([]Input, 0),
		Outputs:  make([]Output, 0),
		LockTime: uint32(time.Now().Unix()),
		CoinType: coinType,
		Data:     make([]byte, 0),
		Witness:  make([][]byte, 0),
	}
}

// AddInput adds an input to the transaction
func (tx *Transaction) AddInput(previousTxHash []byte, outputIndex uint32, scriptSig []byte) {
	tx.Inputs = append(tx.Inputs, Input{
		PreviousTxHash: previousTxHash,
		OutputIndex:    outputIndex,
		ScriptSig:      scriptSig,
		Sequence:       0xffffffff,
	})
}

// AddOutput adds an output to the transaction
func (tx *Transaction) AddOutput(value int64, scriptPubKey []byte) {
	tx.Outputs = append(tx.Outputs, Output{
		Value:        value,
		ScriptPubKey: scriptPubKey,
	})
}

// CalculateHash calculates the transaction hash
func (tx *Transaction) CalculateHash() {
	data, _ := json.Marshal(tx)
	hash := sha256.Sum256(data)
	tx.HashBytes = hash[:]
}

// Sign signs the transaction with the given private key
func (tx *Transaction) Sign(privateKey *ecdsa.PrivateKey) error {
	tx.CalculateHash()
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, tx.HashBytes)
	if err != nil {
		return err
	}
	signature := append(r.Bytes(), s.Bytes()...)
	tx.Signature = signature
	return nil
}

// Validate validates the transaction
func (tx *Transaction) Validate() (bool, error) {
	if len(tx.Inputs) == 0 || len(tx.Outputs) == 0 {
		return false, fmt.Errorf("transaction must have at least one input and one output")
	}

	// Validate inputs
	for _, input := range tx.Inputs {
		if len(input.PreviousTxHash) == 0 {
			return false, fmt.Errorf("invalid previous transaction hash")
		}
	}

	// Validate outputs
	var totalOutput int64
	for _, output := range tx.Outputs {
		if output.Value <= 0 {
			return false, fmt.Errorf("invalid output value")
		}
		totalOutput += output.Value
	}

	// Validate fee
	if tx.Fee < 0 {
		return false, fmt.Errorf("invalid fee")
	}

	return true, nil
}

// AddTransaction adds a transaction to the pool
func (tp *TransactionPool) AddTransaction(tx *types.Transaction) error {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	if len(tp.transactions) >= tp.maxSize {
		return fmt.Errorf("transaction pool is full")
	}

	// Validate transaction
	if err := tx.Validate(); err != nil {
		return err
	}

	// Add to pool
	tp.transactions[string(tx.GetHash())] = tx

	// Update UTXOs
	for _, input := range tx.GetInputs() {
		utxoKey := fmt.Sprintf("%x:%d", input.PreviousTxHash, input.OutputIndex)
		if utxo, exists := tp.utxos[utxoKey]; exists {
			utxo.Spent = true
		}
	}

	for i, output := range tx.GetOutputs() {
		utxo := types.NewUTXO(
			tx.GetHash(),
			uint32(i),
			output.Value,
			output.ScriptPubKey,
			"", // Address will be set by the wallet
			tx.GetCoinType(),
		)
		utxoKey := fmt.Sprintf("%x:%d", tx.GetHash(), i)
		tp.utxos[utxoKey] = utxo
	}

	return nil
}

// GetUTXOsByAddress returns all UTXOs for a given address
func (tp *TransactionPool) GetUTXOsByAddress(address string) []*types.UTXO {
	tp.mu.RLock()
	defer tp.mu.RUnlock()

	var utxos []*types.UTXO
	for _, utxo := range tp.utxos {
		if !utxo.Spent && utxo.Address == address {
			utxos = append(utxos, utxo)
		}
	}
	return utxos
}

// RemoveTransaction removes a transaction from the pool
func (tp *TransactionPool) RemoveTransaction(txHash []byte) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	delete(tp.transactions, string(txHash))
}

// GetTransaction returns a transaction from the pool
func (tp *TransactionPool) GetTransaction(txHash []byte) *types.Transaction {
	tp.mu.RLock()
	defer tp.mu.RUnlock()

	return tp.transactions[string(txHash)]
}

// GetTransactions returns all transactions in the pool
func (tp *TransactionPool) GetTransactions() []*types.Transaction {
	tp.mu.RLock()
	defer tp.mu.RUnlock()

	transactions := make([]*types.Transaction, 0, len(tp.transactions))
	for _, tx := range tp.transactions {
		transactions = append(transactions, tx)
	}
	return transactions
}

// Size returns the number of transactions in the pool
func (tp *TransactionPool) Size() int {
	tp.mu.RLock()
	defer tp.mu.RUnlock()

	return len(tp.transactions)
}

// Validate validates a transaction
func (tp *TransactionPool) Validate(tx *types.Transaction) error {
	// Check basic structure
	if len(tx.GetInputs()) == 0 || len(tx.GetOutputs()) == 0 {
		return fmt.Errorf("%w: empty inputs or outputs", ErrInvalidTransaction)
	}

	// Check for double spends
	spentOutputs := make(map[string]bool)
	var totalInput int64
	var totalOutput int64

	// Validate inputs
	for _, input := range tx.GetInputs() {
		utxoKey := fmt.Sprintf("%x:%d", input.PreviousTxHash, input.OutputIndex)

		// Check if UTXO exists
		utxo, exists := tp.utxos[utxoKey]
		if !exists {
			return fmt.Errorf("%w: UTXO not found", ErrInvalidTransaction)
		}

		// Check if already spent
		if spentOutputs[utxoKey] {
			return fmt.Errorf("%w: double spend detected", ErrInvalidTransaction)
		}

		// Check coin type
		if utxo.CoinType != tx.GetCoinType() {
			return fmt.Errorf("%w: coin type mismatch", ErrInvalidTransaction)
		}

		// Add to total input
		totalInput += utxo.Value
		spentOutputs[utxoKey] = true
	}

	// Validate outputs
	for _, output := range tx.GetOutputs() {
		if output.Value <= 0 {
			return fmt.Errorf("%w: invalid output value", ErrInvalidTransaction)
		}
		totalOutput += output.Value
	}

	// Check if inputs cover outputs plus fee
	if totalInput < totalOutput+tx.GetFee() {
		return fmt.Errorf("%w: insufficient funds", ErrInvalidTransaction)
	}

	return nil
}

// GetUTXO gets a UTXO from the pool
func (tp *TransactionPool) GetUTXO(txHash []byte, index uint32) *types.UTXO {
	utxoKey := fmt.Sprintf("%x:%d", txHash, index)
	return tp.utxos[utxoKey]
}

// AddUTXO adds a UTXO to the pool
func (tp *TransactionPool) AddUTXO(utxo *types.UTXO) {
	utxoKey := fmt.Sprintf("%x:%d", utxo.TxHash, utxo.OutputIndex)
	tp.utxos[utxoKey] = utxo
}

// RemoveUTXO removes a UTXO from the pool
func (tp *TransactionPool) RemoveUTXO(txHash []byte, index uint32) {
	utxoKey := fmt.Sprintf("%x:%d", txHash, index)
	delete(tp.utxos, utxoKey)
}

// GetBalance gets the balance for an address
func (tp *TransactionPool) GetBalance(address string, coinType coin.Type) int64 {
	var balance int64
	for _, utxo := range tp.utxos {
		if utxo.Address == address && utxo.CoinType == coinType && !utxo.Spent {
			balance += utxo.Value
		}
	}
	return balance
}

// Hash returns the transaction hash as a common.Hash
func (tx *Transaction) Hash() common.Hash {
	return common.BytesToHash(tx.HashBytes)
}

// FeeRate calculates the fee rate in satoshis per byte
func (tx *Transaction) FeeRate() int64 {
	size := tx.Size()
	if size == 0 {
		return 0
	}
	return tx.Fee / int64(size)
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
		size += 4  // OutputIndex
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
		Hash:  fmt.Sprintf("%x", tx.HashBytes),
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
