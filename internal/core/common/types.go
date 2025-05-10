package common

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

// Header represents a block header
type Header struct {
	Version       uint32
	PrevBlockHash []byte
	MerkleRoot    []byte
	Timestamp     time.Time
	Difficulty    uint32
	Nonce         uint32
	Height        uint64
	Hash          []byte
}

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

	// Transaction inputs
	Inputs []Input

	// Transaction outputs
	Outputs []Output

	// Transaction witness
	Witness *Witness

	// Transaction lock time
	LockTime uint64
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
	TxHash      []byte
	OutIndex    uint32
	Amount      uint64
	ScriptPub   []byte
	BlockHeight uint64
	IsCoinbase  bool
	IsSegWit    bool
	IsSpent     bool
	IsConfirmed bool
	CreatedAt   time.Time
	SpentAt     time.Time
}

// UTXOSet manages the set of unspent transaction outputs
type UTXOSet struct {
	utxos map[string]*UTXO // key: txHash:outIndex
}

// NewUTXOSet creates a new UTXO set
func NewUTXOSet() *UTXOSet {
	return &UTXOSet{
		utxos: make(map[string]*UTXO),
	}
}

// GetUTXO retrieves a UTXO from the set
func (us *UTXOSet) GetUTXO(txHash []byte, outIndex uint32) (*UTXO, bool) {
	key := utxoKey(txHash, outIndex)
	utxo, exists := us.utxos[key]
	return utxo, exists
}

// utxoKey creates a key for a UTXO
func utxoKey(txHash []byte, outIndex uint32) string {
	return string(txHash) + ":" + strconv.FormatUint(uint64(outIndex), 10)
}

// Witness represents a transaction witness
type Witness struct {
	Data [][]byte
}

// NewWitness creates a new witness
func NewWitness() *Witness {
	return &Witness{
		Data: make([][]byte, 0),
	}
}

// AddData adds data to the witness
func (w *Witness) AddData(data []byte) {
	w.Data = append(w.Data, data)
}

// Size returns the size of the witness in bytes
func (w *Witness) Size() int {
	size := 0
	for _, data := range w.Data {
		size += len(data)
	}
	return size
}

// Clone creates a deep copy of the witness
func (w *Witness) Clone() *Witness {
	clone := &Witness{
		Data: make([][]byte, len(w.Data)),
	}
	for i, data := range w.Data {
		clone.Data[i] = append([]byte{}, data...)
	}
	return clone
}

// Validate validates the witness
func (w *Witness) Validate() error {
	if len(w.Data) == 0 {
		return nil
	}
	return nil
}

// Validate validates the transaction
func (tx *Transaction) Validate() error {
	// Validate version
	if tx.Version == 0 {
		return errors.New("invalid version")
	}

	// Validate from address
	if len(tx.From) == 0 {
		return errors.New("invalid from address")
	}

	// Validate to address
	if len(tx.To) == 0 {
		return errors.New("invalid to address")
	}

	// Validate amount
	if tx.Amount == 0 {
		return errors.New("invalid amount")
	}

	// Validate hash
	if len(tx.Hash) == 0 {
		return errors.New("invalid hash")
	}

	// Validate timestamp
	if tx.Timestamp.IsZero() {
		return errors.New("invalid timestamp")
	}

	// Validate inputs
	if len(tx.Inputs) == 0 {
		return errors.New("no inputs")
	}

	// Validate outputs
	if len(tx.Outputs) == 0 {
		return errors.New("no outputs")
	}

	// Validate witness if present
	if tx.Witness != nil {
		if err := tx.Witness.Validate(); err != nil {
			return fmt.Errorf("invalid witness: %v", err)
		}
	}

	return nil
}

// Size returns the size of the transaction in bytes
func (tx *Transaction) Size() int {
	size := 0

	// Version size
	size += 4

	// From address size
	size += len(tx.From)

	// To address size
	size += len(tx.To)

	// Amount size
	size += 8

	// Data size
	size += len(tx.Data)

	// Hash size
	size += len(tx.Hash)

	// Timestamp size
	size += 8

	// Lock time size
	size += 8

	// Inputs size
	for _, input := range tx.Inputs {
		size += len(input.PreviousTxHash)
		size += 4 // PreviousTxIndex
		size += len(input.ScriptSig)
		size += 4 // Sequence
	}

	// Outputs size
	for _, output := range tx.Outputs {
		size += 8 // Value
		size += len(output.ScriptPubKey)
		size += len(output.Address)
	}

	// Witness size if present
	if tx.Witness != nil {
		size += tx.Witness.Size()
	}

	return size
}

// Copy creates a deep copy of the transaction
func (tx *Transaction) Copy() *Transaction {
	clone := &Transaction{
		Version:   tx.Version,
		From:      append([]byte{}, tx.From...),
		To:        append([]byte{}, tx.To...),
		Amount:    tx.Amount,
		Data:      append([]byte{}, tx.Data...),
		Hash:      append([]byte{}, tx.Hash...),
		Timestamp: tx.Timestamp,
		LockTime:  tx.LockTime,
	}

	if tx.Witness != nil {
		clone.Witness = tx.Witness.Clone()
	}

	clone.Inputs = make([]Input, len(tx.Inputs))
	for i, input := range tx.Inputs {
		clone.Inputs[i] = Input{
			PreviousTxHash:  append([]byte{}, input.PreviousTxHash...),
			PreviousTxIndex: input.PreviousTxIndex,
			ScriptSig:       append([]byte{}, input.ScriptSig...),
			Sequence:        input.Sequence,
		}
	}

	clone.Outputs = make([]Output, len(tx.Outputs))
	for i, output := range tx.Outputs {
		clone.Outputs[i] = Output{
			Value:        output.Value,
			ScriptPubKey: append([]byte{}, output.ScriptPubKey...),
			Address:      output.Address,
		}
	}

	return clone
}

// IsCoinbase checks if the transaction is a coinbase transaction
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].PreviousTxHash) == 0
}
