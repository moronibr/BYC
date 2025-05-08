package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"

	"github.com/youngchain/internal/core/coin"
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

// CalculateHash calculates the transaction hash
func (tx *Transaction) CalculateHash() []byte {
	var buf bytes.Buffer

	// Write version
	binary.Write(&buf, binary.LittleEndian, tx.Version)

	// Write inputs
	binary.Write(&buf, binary.LittleEndian, uint32(len(tx.Inputs)))
	for _, input := range tx.Inputs {
		buf.Write(input.PreviousTxHash)
		binary.Write(&buf, binary.LittleEndian, input.PreviousTxIndex)
		binary.Write(&buf, binary.LittleEndian, uint32(len(input.ScriptSig)))
		buf.Write(input.ScriptSig)
		binary.Write(&buf, binary.LittleEndian, input.Sequence)
	}

	// Write outputs
	binary.Write(&buf, binary.LittleEndian, uint32(len(tx.Outputs)))
	for _, output := range tx.Outputs {
		binary.Write(&buf, binary.LittleEndian, output.Value)
		binary.Write(&buf, binary.LittleEndian, uint32(len(output.ScriptPubKey)))
		buf.Write(output.ScriptPubKey)
	}

	// Write lock time
	binary.Write(&buf, binary.LittleEndian, tx.LockTime)

	// Calculate hash
	hash := sha256.Sum256(buf.Bytes())
	return hash[:]
}

// Copy creates a deep copy of the transaction
func (tx *Transaction) Copy() *Transaction {
	// Create a new transaction
	txCopy := &Transaction{
		Version:  tx.Version,
		LockTime: tx.LockTime,
		Fee:      tx.Fee,
		CoinType: tx.CoinType,
		Inputs:   make([]*Input, len(tx.Inputs)),
		Outputs:  make([]*Output, len(tx.Outputs)),
	}

	// Copy inputs
	for i, input := range tx.Inputs {
		txCopy.Inputs[i] = &Input{
			PreviousTxHash:  make([]byte, len(input.PreviousTxHash)),
			PreviousTxIndex: input.PreviousTxIndex,
			ScriptSig:       make([]byte, len(input.ScriptSig)),
			Sequence:        input.Sequence,
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

	return txCopy
}

// Size returns the size of the transaction in bytes
func (tx *Transaction) Size() int {
	size := 4 // Version
	size += 1 // VarInt for input count
	for _, input := range tx.Inputs {
		size += len(input.PreviousTxHash)
		size += 4 // PreviousTxIndex
		size += len(input.ScriptSig)
		size += 4 // Sequence
	}
	size += 1 // VarInt for output count
	for _, output := range tx.Outputs {
		size += 8 // Value
		size += len(output.ScriptPubKey)
		size += len(output.Address)
	}
	size += 4 // LockTime
	size += 8 // Fee
	size += len(string(tx.CoinType))
	return size
}
