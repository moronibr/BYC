package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"

	"github.com/youngchain/internal/core/common"
)

// TransactionWrapper wraps common.Transaction with additional methods
type TransactionWrapper struct {
	*common.Transaction
}

// NewTransactionWrapper creates a new transaction wrapper
func NewTransactionWrapper(tx *common.Transaction) *TransactionWrapper {
	return &TransactionWrapper{Transaction: tx}
}

// CalculateTxID calculates a transaction ID that is resistant to malleability
func (tx *TransactionWrapper) CalculateTxID() []byte {
	// Serialize the transaction
	data := tx.Serialize()

	// Calculate double SHA256 hash
	hash := sha256.Sum256(data)
	hash = sha256.Sum256(hash[:])

	return hash[:]
}

// VerifyTxID verifies that the transaction ID matches the calculated ID
func (tx *TransactionWrapper) VerifyTxID() bool {
	calculatedID := tx.CalculateTxID()
	return bytes.Equal(tx.Hash, calculatedID)
}

// IsMalleable checks if a transaction has been malleated
func (tx *TransactionWrapper) IsMalleable() bool {
	return !tx.VerifyTxID()
}

// Serialize serializes the transaction to bytes
func (tx *TransactionWrapper) Serialize() []byte {
	var buf bytes.Buffer

	// Write version
	binary.Write(&buf, binary.LittleEndian, tx.Version)

	// Write lock time
	binary.Write(&buf, binary.LittleEndian, tx.LockTime)

	// Write timestamp
	binary.Write(&buf, binary.LittleEndian, tx.Timestamp.Unix())

	// Write from address
	buf.Write(tx.From)

	// Write to address
	buf.Write(tx.To)

	// Write amount
	binary.Write(&buf, binary.LittleEndian, tx.Amount)

	// Write inputs
	binary.Write(&buf, binary.LittleEndian, uint32(len(tx.Inputs)))
	for _, input := range tx.Inputs {
		buf.Write(input.PreviousTxHash)
		binary.Write(&buf, binary.LittleEndian, input.PreviousTxIndex)
		buf.Write(input.ScriptSig)
	}

	// Write outputs
	binary.Write(&buf, binary.LittleEndian, uint32(len(tx.Outputs)))
	for _, output := range tx.Outputs {
		binary.Write(&buf, binary.LittleEndian, output.Value)
		buf.Write(output.ScriptPubKey)
	}

	// Write witness if present
	if tx.Witness != nil {
		for _, data := range tx.Witness.Data {
			buf.Write(data)
		}
	}

	return buf.Bytes()
}
