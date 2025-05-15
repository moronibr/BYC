package types

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/youngchain/internal/core/coin"
)

// BlockType represents the type of a block
type BlockType string

const (
	GoldenBlock BlockType = "golden"
	SilverBlock BlockType = "silver"
)

// BlockHeader represents a block header
type BlockHeader struct {
	Version       uint32    `json:"version"`
	PrevBlockHash []byte    `json:"prevBlockHash"`
	MerkleRoot    []byte    `json:"merkleRoot"`
	Timestamp     uint64    `json:"timestamp"`
	Difficulty    uint32    `json:"difficulty"`
	Nonce         uint64    `json:"nonce"`
	Type          BlockType `json:"type"`
	CoinType      coin.Type `json:"coinType"`
	Height        uint64    `json:"height"`
	Hash          []byte    `json:"hash"`
}

// Block represents a block in the blockchain
type Block struct {
	Header       *BlockHeader   `json:"header"`
	Transactions []*Transaction `json:"transactions"`
	Size         int            `json:"size"`
	Timestamp    uint64         `json:"timestamp"`
	Difficulty   uint32         `json:"difficulty"`
	CoinType     coin.Type      `json:"coinType"`
	Hash         []byte         `json:"hash"`
	PreviousHash []byte         `json:"previousHash"`
}

// Transaction represents a cryptocurrency transaction
type Transaction struct {
	Version   uint32      `json:"version"`
	Inputs    []*TxInput  `json:"inputs"`
	Outputs   []*TxOutput `json:"outputs"`
	LockTime  uint32      `json:"lockTime"`
	Fee       int64       `json:"fee"`
	CoinType  coin.Type   `json:"coinType"`
	Hash      []byte      `json:"hash"`
	Signature []byte      `json:"signature"`
	Data      []byte      `json:"data"`
	Witness   [][]byte    `json:"witness"`
}

// TxInput represents a transaction input
type TxInput struct {
	PreviousTxHash []byte `json:"previousTxHash"`
	OutputIndex    uint32 `json:"outputIndex"`
	ScriptSig      []byte `json:"scriptSig"`
	Sequence       uint32 `json:"sequence"`
	Address        string `json:"address"`
}

// TxOutput represents a transaction output
type TxOutput struct {
	Value        int64  `json:"value"`
	ScriptPubKey []byte `json:"scriptPubKey"`
	Address      string `json:"address"`
}

// UTXO represents an unspent transaction output
type UTXO struct {
	TxHash       []byte    `json:"txHash"`
	OutputIndex  uint32    `json:"outputIndex"`
	Value        int64     `json:"value"`
	ScriptPubKey []byte    `json:"scriptPubKey"`
	Spent        bool      `json:"spent"`
	Address      string    `json:"address"`
	CoinType     coin.Type `json:"coinType"`
	BlockHash    []byte    `json:"blockHash"`
}

// UTXOSetInterface defines the interface for UTXO management
type UTXOSetInterface interface {
	GetUTXO(txHash []byte, index uint32) *UTXO
	AddUTXO(utxo *UTXO)
	RemoveUTXO(txHash []byte, index uint32)
	GetUTXOsByAddress(address string) []*UTXO
	GetBalance(address string, coinType coin.Type) int64
	SpendUTXO(txHash []byte, index uint32) error
	ValidateUTXO(txHash []byte, index uint32) bool
}

// BlockInterface defines the interface for block operations
type BlockInterface interface {
	GetHash() []byte
	GetPreviousHash() []byte
	GetTimestamp() uint64
	GetDifficulty() uint32
	GetCoinType() coin.Type
	GetTransactions() []*Transaction
	Validate() error
	CalculateHash() error
}

// TransactionInterface defines the interface for transaction operations
type TransactionInterface interface {
	GetHash() []byte
	GetInputs() []*TxInput
	GetOutputs() []*TxOutput
	GetFee() int64
	GetCoinType() coin.Type
	Validate() error
	CalculateHash() error
	Sign(privateKey *ecdsa.PrivateKey) error
	Verify() bool
}

// NewBlock creates a new block
func NewBlock(previousHash []byte, timestamp uint64) *Block {
	return &Block{
		Header: &BlockHeader{
			PrevBlockHash: previousHash,
			Timestamp:     timestamp,
		},
		Transactions: make([]*Transaction, 0),
		Timestamp:    timestamp,
		PreviousHash: previousHash,
	}
}

// NewTransaction creates a new transaction
func NewTransaction(version uint32, coinType coin.Type) *Transaction {
	return &Transaction{
		Version:  version,
		Inputs:   make([]*TxInput, 0),
		Outputs:  make([]*TxOutput, 0),
		LockTime: uint32(time.Now().Unix()),
		CoinType: coinType,
		Data:     make([]byte, 0),
		Witness:  make([][]byte, 0),
	}
}

// NewUTXO creates a new UTXO
func NewUTXO(txHash []byte, outputIndex uint32, value int64, scriptPubKey []byte, address string, coinType coin.Type) *UTXO {
	return &UTXO{
		TxHash:       txHash,
		OutputIndex:  outputIndex,
		Value:        value,
		ScriptPubKey: scriptPubKey,
		Spent:        false,
		Address:      address,
		CoinType:     coinType,
	}
}

// CalculateHash calculates the transaction hash
func (tx *Transaction) CalculateHash() error {
	data, err := json.Marshal(tx)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %v", err)
	}
	hash := sha256.Sum256(data)
	tx.Hash = hash[:]
	return nil
}

// Sign signs the transaction with the given private key
func (tx *Transaction) Sign(privateKey *ecdsa.PrivateKey) error {
	if err := tx.CalculateHash(); err != nil {
		return err
	}
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, tx.Hash)
	if err != nil {
		return err
	}
	signature := append(r.Bytes(), s.Bytes()...)
	tx.Signature = signature
	return nil
}

// Validate validates the transaction
func (tx *Transaction) Validate() error {
	if len(tx.Inputs) == 0 || len(tx.Outputs) == 0 {
		return fmt.Errorf("transaction must have at least one input and one output")
	}

	// Validate inputs
	for _, input := range tx.Inputs {
		if len(input.PreviousTxHash) == 0 {
			return fmt.Errorf("invalid previous transaction hash")
		}
	}

	// Validate outputs
	var totalOutput int64
	for _, output := range tx.Outputs {
		if output.Value <= 0 {
			return fmt.Errorf("invalid output value")
		}
		totalOutput += output.Value
	}

	// Validate fee
	if tx.Fee < 0 {
		return fmt.Errorf("invalid fee")
	}

	return nil
}

// GetHash returns the transaction hash
func (tx *Transaction) GetHash() []byte {
	return tx.Hash
}

// GetInputs returns the transaction inputs
func (tx *Transaction) GetInputs() []*TxInput {
	return tx.Inputs
}

// GetOutputs returns the transaction outputs
func (tx *Transaction) GetOutputs() []*TxOutput {
	return tx.Outputs
}

// GetFee returns the transaction fee
func (tx *Transaction) GetFee() int64 {
	return tx.Fee
}

// GetCoinType returns the transaction coin type
func (tx *Transaction) GetCoinType() coin.Type {
	return tx.CoinType
}

// Verify verifies the transaction signature
func (tx *Transaction) Verify() bool {
	// TODO: Implement signature verification
	return true
}

// Copy creates a deep copy of the transaction
func (tx *Transaction) Copy() *Transaction {
	copy := &Transaction{
		Version:   tx.Version,
		LockTime:  tx.LockTime,
		Fee:       tx.Fee,
		CoinType:  tx.CoinType,
		Hash:      make([]byte, len(tx.Hash)),
		Signature: make([]byte, len(tx.Signature)),
		Data:      make([]byte, len(tx.Data)),
		Witness:   make([][]byte, len(tx.Witness)),
	}

	copy.Hash = append(copy.Hash, tx.Hash...)
	copy.Signature = append(copy.Signature, tx.Signature...)
	copy.Data = append(copy.Data, tx.Data...)

	for i, witness := range tx.Witness {
		copy.Witness[i] = append([]byte{}, witness...)
	}

	copy.Inputs = make([]*TxInput, len(tx.Inputs))
	for i, input := range tx.Inputs {
		copy.Inputs[i] = &TxInput{
			PreviousTxHash: append([]byte{}, input.PreviousTxHash...),
			OutputIndex:    input.OutputIndex,
			ScriptSig:      append([]byte{}, input.ScriptSig...),
			Sequence:       input.Sequence,
			Address:        input.Address,
		}
	}

	copy.Outputs = make([]*TxOutput, len(tx.Outputs))
	for i, output := range tx.Outputs {
		copy.Outputs[i] = &TxOutput{
			Value:        output.Value,
			ScriptPubKey: append([]byte{}, output.ScriptPubKey...),
			Address:      output.Address,
		}
	}

	return copy
}

// IsCoinbase checks if the transaction is a coinbase transaction
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].PreviousTxHash) == 0
}

// Size returns the size of the transaction in bytes
func (tx *Transaction) Size() int {
	size := 0

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

	// Data
	size += 4 // Data length
	size += len(tx.Data)

	// Witness
	size += 4 // Witness count
	for _, witness := range tx.Witness {
		size += 4 // Witness length
		size += len(witness)
	}

	return size
}

// AddOutput adds an output to the transaction
func (tx *Transaction) AddOutput(value int64, scriptPubKey []byte) {
	tx.Outputs = append(tx.Outputs, &TxOutput{
		Value:        value,
		ScriptPubKey: scriptPubKey,
	})
}

// AddInput adds an input to the transaction
func (tx *Transaction) AddInput(previousTxHash []byte, outputIndex uint32, scriptSig []byte) {
	tx.Inputs = append(tx.Inputs, &TxInput{
		PreviousTxHash: previousTxHash,
		OutputIndex:    outputIndex,
		ScriptSig:      scriptSig,
		Sequence:       0xffffffff,
	})
}
