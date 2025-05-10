package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
)

// CalculateTxID calculates a transaction ID that is resistant to malleability
func (tx *Transaction) CalculateTxID() []byte {
	// Create a copy of the transaction without the signature
	txCopy := *tx
	txCopy.Signature = nil

	// Serialize the transaction without the signature
	data := txCopy.Serialize()

	// Calculate double SHA256 hash
	hash := sha256.Sum256(data)
	hash = sha256.Sum256(hash[:])

	return hash[:]
}

// VerifyTxID verifies that the transaction ID matches the calculated ID
func (tx *Transaction) VerifyTxID() bool {
	calculatedID := tx.CalculateTxID()
	return bytes.Equal(tx.Hash, calculatedID)
}

// IsMalleable checks if a transaction has been malleated
func (tx *Transaction) IsMalleable() bool {
	return !tx.VerifyTxID()
}

// Serialize serializes the transaction to bytes
func (tx *Transaction) Serialize() []byte {
	var buf bytes.Buffer

	// Write version
	binary.Write(&buf, binary.LittleEndian, tx.Version)

	// Write lock time
	binary.Write(&buf, binary.LittleEndian, tx.LockTime)

	// Write timestamp
	binary.Write(&buf, binary.LittleEndian, tx.Timestamp.Unix())

	// Write from address
	buf.WriteString(tx.From)

	// Write to address
	buf.WriteString(tx.To)

	// Write amount
	binary.Write(&buf, binary.LittleEndian, tx.Amount)

	// Write nonce
	binary.Write(&buf, binary.LittleEndian, tx.Nonce)

	// Write inputs
	binary.Write(&buf, binary.LittleEndian, uint32(len(tx.Inputs)))
	for _, input := range tx.Inputs {
		buf.Write(input.PreviousTx)
		binary.Write(&buf, binary.LittleEndian, input.OutputIndex)
		buf.Write(input.Signature)
	}

	// Write outputs
	binary.Write(&buf, binary.LittleEndian, uint32(len(tx.Outputs)))
	for _, output := range tx.Outputs {
		buf.WriteString(output.Address)
		binary.Write(&buf, binary.LittleEndian, output.Amount)
	}

	// Write witness if present
	if tx.Witness != nil {
		buf.Write(tx.Witness.Serialize())
	}

	return buf.Bytes()
}

// TxInput represents a transaction input
type TxInput struct {
	PreviousTx  []byte
	OutputIndex uint32
	Signature   []byte
	Sequence    uint32
}

// TxOutput represents a transaction output
type TxOutput struct {
	Address string
	Amount  uint64
	Script  []byte
}
