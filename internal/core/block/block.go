package block

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"time"

	"github.com/youngchain/internal/core/coin"
)

// BlockType represents the type of block
type BlockType string

const (
	GoldenBlock BlockType = "golden"
	SilverBlock BlockType = "silver"
)

// Block represents a basic block in the blockchain
type Block struct {
	// Block header
	Version      uint32
	PreviousHash []byte
	MerkleRoot   []byte
	Timestamp    int64
	Type         BlockType

	// Mining information
	Difficulty uint64
	Nonce      uint64
	Hash       []byte

	// Transactions
	Transactions []Transaction
}

// Transaction represents a transaction in the block
type Transaction struct {
	Version    uint32
	Inputs     []TxInput
	Outputs    []TxOutput
	LockTime   uint32
	CoinType   coin.CoinType
	CrossBlock bool
	Signature  []byte
}

// TxInput represents a transaction input
type TxInput struct {
	PreviousTx  []byte
	OutputIndex uint32
	Script      []byte
	Sequence    uint32
}

// TxOutput represents a transaction output
type TxOutput struct {
	Value    uint64
	Script   []byte
	CoinType coin.CoinType
}

// NewBlock creates a new block
func NewBlock(blockType BlockType, previousHash []byte) *Block {
	return &Block{
		Version:      1,
		PreviousHash: previousHash,
		Timestamp:    time.Now().Unix(),
		Type:         blockType,
		Difficulty:   calculateInitialDifficulty(blockType),
	}
}

// CalculateHash calculates the block hash
func (b *Block) CalculateHash() []byte {
	header := append(b.PreviousHash, b.MerkleRoot...)
	header = append(header, []byte(string(b.Timestamp))...)
	header = append(header, []byte(string(b.Difficulty))...)
	header = append(header, []byte(string(b.Nonce))...)

	hash := sha256.Sum256(header)
	return hash[:]
}

// calculateInitialDifficulty sets the initial difficulty based on block type
func calculateInitialDifficulty(blockType BlockType) uint64 {
	switch blockType {
	case GoldenBlock:
		return 0x1d00ffff // Initial difficulty for golden block
	case SilverBlock:
		return 0x1d00ffff // Initial difficulty for silver block
	default:
		return 0x1d00ffff
	}
}

// String returns a string representation of the block
func (b *Block) String() string {
	return hex.EncodeToString(b.Hash)
}

// AddTransaction adds a transaction to the block
func (b *Block) AddTransaction(tx Transaction) {
	b.Transactions = append(b.Transactions, tx)
	b.UpdateMerkleRoot()
}

// UpdateMerkleRoot updates the merkle root of the block
func (b *Block) UpdateMerkleRoot() {
	// TODO: Implement merkle tree calculation
	// For now, we'll use a simple hash of all transaction hashes
	var txHashes [][]byte
	for _, tx := range b.Transactions {
		txHash := sha256.Sum256([]byte(string(tx.Version)))
		txHashes = append(txHashes, txHash[:])
	}

	combined := []byte{}
	for _, hash := range txHashes {
		combined = append(combined, hash...)
	}

	merkleRoot := sha256.Sum256(combined)
	b.MerkleRoot = merkleRoot[:]
}

// CalculateHash calculates the hash of a transaction
func (tx *Transaction) CalculateHash() []byte {
	// Serialize transaction data
	data := make([]byte, 0)

	// Add version
	versionBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(versionBytes, tx.Version)
	data = append(data, versionBytes...)

	// Add inputs
	for _, input := range tx.Inputs {
		data = append(data, input.PreviousTx...)
		indexBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(indexBytes, input.OutputIndex)
		data = append(data, indexBytes...)
		data = append(data, input.Script...)
		seqBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(seqBytes, input.Sequence)
		data = append(data, seqBytes...)
	}

	// Add outputs
	for _, output := range tx.Outputs {
		valueBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(valueBytes, output.Value)
		data = append(data, valueBytes...)
		data = append(data, output.Script...)
	}

	// Add lock time
	lockTimeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(lockTimeBytes, tx.LockTime)
	data = append(data, lockTimeBytes...)

	// Add coin type
	data = append(data, []byte(tx.CoinType)...)

	// Add cross block flag
	if tx.CrossBlock {
		data = append(data, 1)
	} else {
		data = append(data, 0)
	}

	// Calculate hash
	hash := sha256.Sum256(data)
	return hash[:]
}
