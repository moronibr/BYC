package types

import (
	"crypto/sha256"
	"encoding/binary"
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
	From []byte

	// Transaction outputs
	To []byte

	// Transaction amount
	Amount uint64

	// Transaction data
	Data []byte
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

// NewTransaction creates a new transaction
func NewTransaction(from, to []byte, amount uint64, data []byte) *Transaction {
	tx := &Transaction{
		Version:   1,
		Timestamp: time.Now(),
		From:      from,
		To:        to,
		Amount:    amount,
		Data:      data,
	}

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

	// Add timestamp
	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, uint64(tx.Timestamp.UnixNano()))
	buf = append(buf, timestampBytes...)

	// Add from address
	buf = append(buf, tx.From...)

	// Add to address
	buf = append(buf, tx.To...)

	// Add amount
	amountBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(amountBytes, tx.Amount)
	buf = append(buf, amountBytes...)

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
		From:      make([]byte, len(tx.From)),
		To:        make([]byte, len(tx.To)),
		Amount:    tx.Amount,
		Data:      make([]byte, len(tx.Data)),
		Timestamp: tx.Timestamp,
	}

	copy(txCopy.From, tx.From)
	copy(txCopy.To, tx.To)
	copy(txCopy.Data, tx.Data)

	return txCopy
}

// Size returns the size of the transaction in bytes
func (tx *Transaction) Size() int {
	size := 0

	size += 4            // Version
	size += len(tx.From) // From
	size += len(tx.To)   // To
	size += 8            // Amount
	size += len(tx.Data) // Data
	size += 8            // Timestamp

	return size
}
