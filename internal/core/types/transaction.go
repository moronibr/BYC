package types

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"time"

	"github.com/youngchain/internal/core/coin"
)

// Transaction represents a blockchain transaction
type Transaction struct {
	// Transaction hash
	Hash []byte

	// Transaction version
	Version uint32

	// Transaction timestamp
	Timestamp time.Time

	// Transaction inputs
	Inputs []*Input

	// Transaction outputs
	Outputs []*Output

	// Transaction lock time
	LockTime uint32

	// Transaction fee
	Fee uint64

	// Transaction coin type
	CoinType coin.CoinType

	// Transaction data
	Data []byte
}

// Input represents a transaction input
type Input struct {
	PreviousTxHash  []byte
	PreviousTxIndex uint32
	ScriptSig       []byte
	Sequence        uint32
	Address         string
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

// NewTransaction creates a new transaction
func NewTransaction(from, to []byte, amount uint64, data []byte) *Transaction {
	tx := &Transaction{
		Version:   1,
		Timestamp: time.Now(),
		Inputs:    make([]*Input, 0),
		Outputs:   make([]*Output, 0),
		LockTime:  0,
		Fee:       0,
		CoinType:  coin.Leah,
		Data:      data,
	}

	// Add input
	tx.Inputs = append(tx.Inputs, &Input{
		Address:         string(from),
		ScriptSig:       nil,
		Sequence:        0xffffffff,
		PreviousTxHash:  nil,
		PreviousTxIndex: 0,
	})

	// Add output
	tx.Outputs = append(tx.Outputs, &Output{
		Address:      string(to),
		Value:        amount,
		ScriptPubKey: nil,
	})

	// Calculate transaction hash
	tx.Hash = tx.CalculateHash()

	return tx
}

// CalculateHash calculates the transaction hash
func (tx *Transaction) CalculateHash() []byte {
	// Create a buffer to store transaction data
	buf := make([]byte, 0)

	// Add version
	versionBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(versionBytes, tx.Version)
	buf = append(buf, versionBytes...)

	// Add inputs
	for _, input := range tx.Inputs {
		buf = append(buf, input.PreviousTxHash...)
		indexBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(indexBytes, input.PreviousTxIndex)
		buf = append(buf, indexBytes...)
		buf = append(buf, input.ScriptSig...)
		seqBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(seqBytes, input.Sequence)
		buf = append(buf, seqBytes...)
	}

	// Add outputs
	for _, output := range tx.Outputs {
		valueBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(valueBytes, output.Value)
		buf = append(buf, valueBytes...)
		buf = append(buf, output.ScriptPubKey...)
	}

	// Add lock time and fee
	lockTimeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lockTimeBytes, tx.LockTime)
	buf = append(buf, lockTimeBytes...)
	feeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(feeBytes, tx.Fee)
	buf = append(buf, feeBytes...)

	// Add coin type
	buf = append(buf, []byte(tx.CoinType)...)

	// Add data
	buf = append(buf, tx.Data...)

	// Calculate hash
	hash := sha256.Sum256(buf)
	return hash[:]
}

// Copy creates a deep copy of the transaction
func (tx *Transaction) Copy() *Transaction {
	txCopy := &Transaction{
		Version:   tx.Version,
		Timestamp: tx.Timestamp,
		Inputs:    make([]*Input, len(tx.Inputs)),
		Outputs:   make([]*Output, len(tx.Outputs)),
		LockTime:  tx.LockTime,
		Fee:       tx.Fee,
		CoinType:  tx.CoinType,
		Data:      make([]byte, len(tx.Data)),
		Hash:      make([]byte, len(tx.Hash)),
	}

	// Copy inputs
	for i, input := range tx.Inputs {
		txCopy.Inputs[i] = &Input{
			PreviousTxHash:  make([]byte, len(input.PreviousTxHash)),
			PreviousTxIndex: input.PreviousTxIndex,
			ScriptSig:       make([]byte, len(input.ScriptSig)),
			Sequence:        input.Sequence,
			Address:         input.Address,
		}
		copy(txCopy.Inputs[i].PreviousTxHash, input.PreviousTxHash)
		copy(txCopy.Inputs[i].ScriptSig, input.ScriptSig)
	}

	// Copy outputs
	for i, output := range tx.Outputs {
		txCopy.Outputs[i] = &Output{
			Value:        output.Value,
			ScriptPubKey: make([]byte, len(output.ScriptPubKey)),
			Address:      output.Address,
		}
		copy(txCopy.Outputs[i].ScriptPubKey, output.ScriptPubKey)
	}

	copy(txCopy.Data, tx.Data)
	copy(txCopy.Hash, tx.Hash)

	return txCopy
}

// Size returns the size of the transaction in bytes
func (tx *Transaction) Size() int {
	size := 0

	size += 4 // Version
	size += 8 // Timestamp
	size += 4 // LockTime
	size += 8 // Fee
	size += len(tx.Data)

	// Add size of inputs
	for _, input := range tx.Inputs {
		size += len(input.PreviousTxHash)
		size += 4 // PreviousTxIndex
		size += len(input.ScriptSig)
		size += 4 // Sequence
		size += len(input.Address)
	}

	// Add size of outputs
	for _, output := range tx.Outputs {
		size += 8 // Value
		size += len(output.ScriptPubKey)
		size += len(output.Address)
	}

	// Add size of coin type
	size += len(tx.CoinType)

	return size
}

// Validate validates the transaction
func (tx *Transaction) Validate() error {
	// Validate version
	if tx.Version == 0 {
		return errors.New("invalid version")
	}

	// Validate inputs
	if len(tx.Inputs) == 0 {
		return errors.New("invalid inputs")
	}

	// Validate outputs
	if len(tx.Outputs) == 0 {
		return errors.New("invalid outputs")
	}

	// Validate timestamp
	if tx.Timestamp.IsZero() {
		return errors.New("invalid timestamp")
	}

	// Validate hash
	if len(tx.Hash) == 0 {
		return errors.New("invalid hash")
	}

	// Validate data
	if len(tx.Data) == 0 {
		return errors.New("invalid data")
	}

	// Validate coin type
	if tx.CoinType == "" {
		return errors.New("invalid coin type")
	}

	return nil
}
