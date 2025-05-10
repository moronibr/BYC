package common

import (
	"strconv"
	"time"

	"github.com/youngchain/internal/core/coin"
	"github.com/youngchain/internal/core/types"
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
	tx *types.Transaction
}

// NewTransaction creates a new transaction
func NewTransaction(from, to []byte, amount uint64, data []byte) *Transaction {
	return &Transaction{
		tx: types.NewTransaction(from, to, amount, data),
	}
}

// GetTransaction returns the underlying types.Transaction
func (t *Transaction) GetTransaction() *types.Transaction {
	return t.tx
}

// From returns the from address from the first input
func (t *Transaction) From() []byte {
	if len(t.tx.Inputs) > 0 {
		return []byte(t.tx.Inputs[0].Address)
	}
	return nil
}

// To returns the to address from the first output
func (t *Transaction) To() []byte {
	if len(t.tx.Outputs) > 0 {
		return []byte(t.tx.Outputs[0].Address)
	}
	return nil
}

// Amount returns the amount from the first output
func (t *Transaction) Amount() uint64 {
	if len(t.tx.Outputs) > 0 {
		return t.tx.Outputs[0].Value
	}
	return 0
}

// Hash returns the transaction hash
func (t *Transaction) Hash() []byte {
	return t.tx.Hash
}

// Version returns the transaction version
func (t *Transaction) Version() uint32 {
	return t.tx.Version
}

// Timestamp returns the transaction timestamp
func (t *Transaction) Timestamp() time.Time {
	return t.tx.Timestamp
}

// Inputs returns the transaction inputs
func (t *Transaction) Inputs() []*Input {
	inputs := make([]*Input, len(t.tx.Inputs))
	for i, input := range t.tx.Inputs {
		inputs[i] = &Input{
			PreviousTxHash:  input.PreviousTxHash,
			PreviousTxIndex: input.PreviousTxIndex,
			ScriptSig:       input.ScriptSig,
			Sequence:        input.Sequence,
			Address:         input.Address,
		}
	}
	return inputs
}

// Outputs returns the transaction outputs
func (t *Transaction) Outputs() []*Output {
	outputs := make([]*Output, len(t.tx.Outputs))
	for i, output := range t.tx.Outputs {
		outputs[i] = &Output{
			Value:        output.Value,
			ScriptPubKey: output.ScriptPubKey,
			Address:      output.Address,
		}
	}
	return outputs
}

// LockTime returns the transaction lock time
func (t *Transaction) LockTime() uint32 {
	return t.tx.LockTime
}

// Fee returns the transaction fee
func (t *Transaction) Fee() uint64 {
	return t.tx.Fee
}

// CoinType returns the transaction coin type
func (t *Transaction) CoinType() coin.CoinType {
	return t.tx.CoinType
}

// Data returns the transaction data
func (t *Transaction) Data() []byte {
	return t.tx.Data
}

// Witness returns the transaction witness data
func (t *Transaction) Witness() [][]byte {
	return t.tx.Witness
}

// Copy creates a deep copy of the transaction
func (t *Transaction) Copy() *Transaction {
	return &Transaction{
		tx: t.tx.Copy(),
	}
}

// Validate validates the transaction
func (t *Transaction) Validate() error {
	return t.tx.Validate()
}

// Size returns the size of the transaction in bytes
func (t *Transaction) Size() int {
	return t.tx.Size()
}

// IsCoinbase returns whether the transaction is a coinbase transaction
func (t *Transaction) IsCoinbase() bool {
	return t.tx.IsCoinbase()
}

// Input represents a transaction input
type Input struct {
	// Previous transaction hash
	PreviousTxHash []byte

	// Previous transaction output index
	PreviousTxIndex uint32

	// Input script
	ScriptSig []byte

	// Sequence number
	Sequence uint32

	// Input address
	Address string
}

// Output represents a transaction output
type Output struct {
	// Output value
	Value uint64

	// Output script
	ScriptPubKey []byte

	// Output address
	Address string
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

// Hash represents a 32-byte hash
type Hash [32]byte

// BytesToHash converts a byte slice to a Hash
func BytesToHash(b []byte) Hash {
	var h Hash
	copy(h[:], b)
	return h
}
