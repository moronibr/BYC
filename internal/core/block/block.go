package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/youngchain/internal/core/common"
)

// BlockType represents the type of block
type BlockType string

const (
	GoldenBlock BlockType = "golden"
	SilverBlock BlockType = "silver"
)

// Block represents a block in the blockchain
type Block struct {
	// Block header
	Header *common.Header

	// Block transactions
	Transactions []*common.Transaction

	// Block size in bytes
	BlockSize int

	// Block weight in weight units
	Weight int

	// Block validation status
	IsValid bool

	// Block validation error
	ValidationError error
}

// String returns a string representation of the block
func (b *Block) String() string {
	return fmt.Sprintf("Block{Version: %d, PrevHash: %s, MerkleRoot: %s, Timestamp: %s, Difficulty: %d, Nonce: %d, Hash: %s, Height: %d, Size: %d, Weight: %d, TxCount: %d}",
		b.Header.Version,
		hex.EncodeToString(b.Header.PrevBlockHash),
		hex.EncodeToString(b.Header.MerkleRoot),
		b.Header.Timestamp.Format(time.RFC3339),
		b.Header.Difficulty,
		b.Header.Nonce,
		hex.EncodeToString(b.Header.Hash),
		b.Header.Height,
		b.BlockSize,
		b.Weight,
		len(b.Transactions),
	)
}

// NewBlock creates a new block
func NewBlock(prevHash []byte, height uint64) *Block {
	return &Block{
		Header: &common.Header{
			Version:       1,
			PrevBlockHash: prevHash,
			Timestamp:     time.Now(),
			Difficulty:    0x1d00ffff,
			Height:        height,
		},
		Transactions: make([]*common.Transaction, 0),
	}
}

// AddTransaction adds a transaction to the block
func (b *Block) AddTransaction(tx *common.Transaction) error {
	if !b.CanAddTransaction(tx) {
		return errors.New("block is full")
	}

	b.Transactions = append(b.Transactions, tx)
	b.BlockSize = b.GetBlockSize()
	b.Weight = b.GetBlockWeight()

	return nil
}

// CalculateHash calculates the block hash
func (b *Block) CalculateHash() []byte {
	hash := sha256.New()

	// Hash version
	versionBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(versionBytes, b.Header.Version)
	hash.Write(versionBytes)

	// Hash previous block hash
	hash.Write(b.Header.PrevBlockHash)

	// Hash merkle root
	hash.Write(b.Header.MerkleRoot)

	// Hash timestamp
	timeBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(timeBytes, uint64(b.Header.Timestamp.Unix()))
	hash.Write(timeBytes)

	// Hash difficulty
	diffBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(diffBytes, uint64(b.Header.Difficulty))
	hash.Write(diffBytes)

	// Hash nonce
	nonceBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(nonceBytes, b.Header.Nonce)
	hash.Write(nonceBytes)

	return hash.Sum(nil)
}

// Validate validates the block
func (b *Block) Validate() error {
	// Validate version
	if b.Header.Version == 0 {
		return errors.New("invalid version")
	}

	// Validate previous block hash
	if len(b.Header.PrevBlockHash) != 32 {
		return errors.New("invalid previous block hash")
	}

	// Validate merkle root
	if len(b.Header.MerkleRoot) != 32 {
		return errors.New("invalid merkle root")
	}

	// Validate timestamp
	if b.Header.Timestamp.IsZero() {
		return errors.New("invalid timestamp")
	}

	// Validate difficulty
	if b.Header.Difficulty == 0 {
		return errors.New("invalid difficulty")
	}

	// Validate hash
	if len(b.Header.Hash) != 32 {
		return errors.New("invalid hash")
	}

	// Validate transactions
	for _, tx := range b.Transactions {
		if err := tx.Validate(); err != nil {
			return fmt.Errorf("invalid transaction: %v", err)
		}
	}

	// Validate block weight
	if err := b.ValidateBlockWeight(); err != nil {
		return err
	}

	// Validate merkle root matches transactions
	calculatedMerkleRoot := b.CalculateMerkleRoot()
	if !bytes.Equal(b.Header.MerkleRoot, calculatedMerkleRoot) {
		return errors.New("invalid merkle root")
	}

	return nil
}

// Clone creates a deep copy of the block
func (b *Block) Clone() *Block {
	clone := &Block{
		Header: &common.Header{
			Version:       b.Header.Version,
			PrevBlockHash: append([]byte{}, b.Header.PrevBlockHash...),
			MerkleRoot:    append([]byte{}, b.Header.MerkleRoot...),
			Timestamp:     b.Header.Timestamp,
			Difficulty:    b.Header.Difficulty,
			Nonce:         b.Header.Nonce,
			Height:        b.Header.Height,
		},
		Transactions:    make([]*common.Transaction, len(b.Transactions)),
		BlockSize:       b.BlockSize,
		Weight:          b.Weight,
		IsValid:         b.IsValid,
		ValidationError: b.ValidationError,
	}

	for i, tx := range b.Transactions {
		clone.Transactions[i] = tx.Copy()
	}

	return clone
}

// CalculateMerkleRoot calculates the merkle root of the block's transactions
func (b *Block) CalculateMerkleRoot() []byte {
	if len(b.Transactions) == 0 {
		return nil
	}

	// Create a slice of transaction hashes
	hashes := make([][]byte, len(b.Transactions))
	for i, tx := range b.Transactions {
		hashes[i] = tx.Hash()
	}

	// Calculate merkle root
	return calculateMerkleRoot(hashes)
}

// calculateMerkleRoot calculates the merkle root from a list of hashes
func calculateMerkleRoot(hashes [][]byte) []byte {
	if len(hashes) == 0 {
		return nil
	}

	if len(hashes) == 1 {
		return hashes[0]
	}

	// Create a new level of hashes
	newLevel := make([][]byte, 0)
	for i := 0; i < len(hashes); i += 2 {
		if i+1 == len(hashes) {
			// If there's an odd number of hashes, duplicate the last one
			newLevel = append(newLevel, calculateHash(append(hashes[i], hashes[i]...)))
		} else {
			newLevel = append(newLevel, calculateHash(append(hashes[i], hashes[i+1]...)))
		}
	}

	return calculateMerkleRoot(newLevel)
}

// calculateHash calculates the SHA-256 hash of the input
func calculateHash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// Copy creates a deep copy of the block
func (b *Block) Copy() *Block {
	// Create a new block
	blockCopy := &Block{
		Header: &common.Header{
			Version:       b.Header.Version,
			PrevBlockHash: append([]byte{}, b.Header.PrevBlockHash...),
			MerkleRoot:    append([]byte{}, b.Header.MerkleRoot...),
			Timestamp:     b.Header.Timestamp,
			Difficulty:    b.Header.Difficulty,
			Nonce:         b.Header.Nonce,
			Height:        b.Header.Height,
		},
		Transactions:    make([]*common.Transaction, len(b.Transactions)),
		BlockSize:       b.BlockSize,
		Weight:          b.Weight,
		IsValid:         b.IsValid,
		ValidationError: b.ValidationError,
	}

	// Copy byte slices
	copy(blockCopy.Header.PrevBlockHash, b.Header.PrevBlockHash)
	copy(blockCopy.Header.MerkleRoot, b.Header.MerkleRoot)
	copy(blockCopy.Header.Hash, b.Header.Hash)

	// Copy transactions
	for i, tx := range b.Transactions {
		blockCopy.Transactions[i] = tx.Copy()
	}

	return blockCopy
}

// MarshalJSON implements the json.Marshaler interface
func (b *Block) MarshalJSON() ([]byte, error) {
	type Alias Block
	return json.Marshal(&struct {
		*Alias
		HashHex string `json:"hash_hex"`
	}{
		Alias:   (*Alias)(b),
		HashHex: hex.EncodeToString(b.Header.Hash),
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (b *Block) UnmarshalJSON(data []byte) error {
	type Alias Block
	aux := &struct {
		*Alias
		HashHex string `json:"hash_hex"`
	}{
		Alias: (*Alias)(b),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.HashHex != "" {
		hash, err := hex.DecodeString(aux.HashHex)
		if err != nil {
			return err
		}
		b.Header.Hash = hash
	}
	return nil
}

// UpdateMerkleRoot updates the merkle root of the block
func (b *Block) UpdateMerkleRoot() {
	b.Header.MerkleRoot = b.CalculateMerkleRoot()
}

// GetInitialDifficulty returns the initial difficulty for a block type
func GetInitialDifficulty(blockType BlockType) uint32 {
	return calculateInitialDifficulty(blockType)
}

// calculateInitialDifficulty calculates the initial difficulty for a block type
func calculateInitialDifficulty(blockType BlockType) uint32 {
	switch blockType {
	case GoldenBlock:
		return 0x1d00ffff // Bitcoin's initial difficulty
	case SilverBlock:
		return 0x1d00ffff / 2 // Half of Bitcoin's initial difficulty
	default:
		return 0x1d00ffff
	}
}

// Size returns the size of the block in bytes
func (b *Block) Size() int {
	size := 0

	// Header size
	size += 8 // Height
	size += 8 // Timestamp
	size += len(b.Header.PrevBlockHash)
	size += 8 // Nonce
	size += 8 // Difficulty

	// Transactions size
	for _, tx := range b.Transactions {
		size += tx.Size()
	}

	// Hash size
	size += len(b.Header.Hash)

	return size
}

// GetBlockSize returns the size of the block in bytes
func (b *Block) GetBlockSize() int {
	size := 0

	// Version size
	size += 4

	// Previous block hash size
	size += len(b.Header.PrevBlockHash)

	// Merkle root size
	size += len(b.Header.MerkleRoot)

	// Timestamp size
	size += 8

	// Difficulty size
	size += 8

	// Nonce size
	size += 4

	// Transaction count size
	size += 4

	// Transaction sizes
	for _, tx := range b.Transactions {
		size += tx.Size()
	}

	return size
}

// GetBlockWeight returns the weight of the block
func (b *Block) GetBlockWeight() int {
	weight := 0

	// Base weight
	weight += b.GetBlockSize()

	// Additional weight for transactions
	for _, tx := range b.Transactions {
		weight += tx.Size() * 4 // Transactions are weighted 4x
	}

	return weight
}

// ValidateBlockWeight validates the block weight
func (b *Block) ValidateBlockWeight() error {
	maxWeight := 4 * 1024 * 1024 // 4MB
	if b.GetBlockWeight() > maxWeight {
		return errors.New("block weight exceeds maximum")
	}
	return nil
}

// CanAddTransaction checks if a transaction can be added to the block
func (b *Block) CanAddTransaction(tx *common.Transaction) bool {
	// Check if adding the transaction would exceed the maximum block weight
	newWeight := b.GetBlockWeight() + tx.Size()*4
	maxWeight := 4 * 1024 * 1024 // 4MB
	return newWeight <= maxWeight
}

// ValidateTransactions validates all transactions in the block
func (b *Block) ValidateTransactions(utxoSet *common.UTXOSet) error {
	// Check if block has transactions
	if len(b.Transactions) == 0 {
		return errors.New("block has no transactions")
	}

	// Check if first transaction is coinbase
	if !b.Transactions[0].IsCoinbase() {
		return errors.New("first transaction is not coinbase")
	}

	// Check if block has more than one coinbase transaction
	for i := 1; i < len(b.Transactions); i++ {
		if b.Transactions[i].IsCoinbase() {
			return errors.New("block has more than one coinbase transaction")
		}
	}

	// Validate each transaction
	for _, tx := range b.Transactions {
		if err := tx.Validate(); err != nil {
			return fmt.Errorf("invalid transaction: %v", err)
		}
	}

	return nil
}
