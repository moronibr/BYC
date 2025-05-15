package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/youngchain/internal/core/coin"
	"github.com/youngchain/internal/core/types"
)

// Block represents a block in the blockchain
type Block struct {
	Header       *types.BlockHeader
	Transactions []*types.Transaction
	Size         int
	Timestamp    uint64
	Difficulty   uint32
	CoinType     coin.Type
	Hash         []byte
	PreviousHash []byte
}

// String returns a string representation of the block
func (b *Block) String() string {
	return fmt.Sprintf("Block{Version: %d, PrevHash: %s, MerkleRoot: %s, Timestamp: %d, Difficulty: %d, Nonce: %d, Hash: %s, Height: %d, Size: %d, TxCount: %d}",
		b.Header.Version,
		hex.EncodeToString(b.Header.PrevBlockHash),
		hex.EncodeToString(b.Header.MerkleRoot),
		b.Header.Timestamp,
		b.Header.Difficulty,
		b.Header.Nonce,
		hex.EncodeToString(b.Header.Hash),
		b.Header.Height,
		b.Size,
		len(b.Transactions),
	)
}

// NewBlock creates a new block
func NewBlock(previousHash []byte, timestamp uint64) *Block {
	return &Block{
		Header: &types.BlockHeader{
			PrevBlockHash: previousHash,
			Timestamp:     timestamp,
		},
		Transactions: make([]*types.Transaction, 0),
		Timestamp:    timestamp,
		PreviousHash: previousHash,
	}
}

// CalculateHash calculates the block hash
func (b *Block) CalculateHash() error {
	// Create a copy of the block without the hash
	blockCopy := *b
	blockCopy.Hash = nil

	// Marshal the block to JSON
	data, err := json.Marshal(blockCopy)
	if err != nil {
		return fmt.Errorf("failed to marshal block: %v", err)
	}

	// Calculate SHA-256 hash
	hash := sha256.Sum256(data)
	b.Hash = hash[:]
	b.Header.Hash = hash[:]

	return nil
}

// Validate validates the block
func (b *Block) Validate() error {
	// Check if block is nil
	if b == nil {
		return fmt.Errorf("block is nil")
	}

	// Check if header is nil
	if b.Header == nil {
		return fmt.Errorf("block header is nil")
	}

	// Check if hash is valid
	if err := b.CalculateHash(); err != nil {
		return fmt.Errorf("failed to calculate hash: %v", err)
	}

	// Check if previous hash matches
	if b.PreviousHash != nil && !bytes.Equal(b.PreviousHash, b.Header.PrevBlockHash) {
		return fmt.Errorf("previous hash mismatch")
	}

	// Check if timestamp is valid
	if b.Timestamp > uint64(time.Now().Unix()) {
		return fmt.Errorf("invalid timestamp")
	}

	// Check if difficulty is valid
	if b.Difficulty == 0 {
		return fmt.Errorf("invalid difficulty")
	}

	// Check if coin type is valid
	if b.CoinType == "" {
		return fmt.Errorf("invalid coin type")
	}

	// Validate transactions
	for _, tx := range b.Transactions {
		if err := tx.Validate(); err != nil {
			return fmt.Errorf("invalid transaction: %v", err)
		}
	}

	return nil
}

// AddTransaction adds a transaction to the block
func (b *Block) AddTransaction(tx *types.Transaction) error {
	// Validate transaction
	if err := tx.Validate(); err != nil {
		return fmt.Errorf("invalid transaction: %v", err)
	}

	// Add transaction
	b.Transactions = append(b.Transactions, tx)

	// Update block size
	data, err := json.Marshal(tx)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %v", err)
	}
	b.Size += len(data)

	return nil
}

// GetHash returns the block hash
func (b *Block) GetHash() []byte {
	return b.Hash
}

// GetPreviousHash returns the previous block hash
func (b *Block) GetPreviousHash() []byte {
	return b.PreviousHash
}

// GetTimestamp returns the block timestamp
func (b *Block) GetTimestamp() uint64 {
	return b.Timestamp
}

// GetDifficulty returns the block difficulty
func (b *Block) GetDifficulty() uint32 {
	return b.Difficulty
}

// GetCoinType returns the block coin type
func (b *Block) GetCoinType() coin.Type {
	return b.CoinType
}

// GetTransactions returns the block transactions
func (b *Block) GetTransactions() []*types.Transaction {
	return b.Transactions
}

// CalculateMerkleRoot calculates the merkle root of the block's transactions
func (b *Block) CalculateMerkleRoot() []byte {
	if len(b.Transactions) == 0 {
		return nil
	}

	// Create a slice of transaction hashes
	hashes := make([][]byte, len(b.Transactions))
	for i, tx := range b.Transactions {
		hashes[i] = tx.GetHash()
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

// Clone creates a deep copy of the block
func (b *Block) Clone() *Block {
	clone := &Block{
		Header: &types.BlockHeader{
			Version:       b.Header.Version,
			PrevBlockHash: append([]byte{}, b.Header.PrevBlockHash...),
			MerkleRoot:    append([]byte{}, b.Header.MerkleRoot...),
			Timestamp:     b.Header.Timestamp,
			Difficulty:    b.Header.Difficulty,
			Nonce:         b.Header.Nonce,
			Height:        b.Header.Height,
		},
		Transactions: make([]*types.Transaction, len(b.Transactions)),
		Size:         b.Size,
		Timestamp:    b.Timestamp,
		Difficulty:   b.Difficulty,
		CoinType:     b.CoinType,
		Hash:         append([]byte{}, b.Hash...),
		PreviousHash: append([]byte{}, b.PreviousHash...),
	}

	for i, tx := range b.Transactions {
		clone.Transactions[i] = tx.Copy()
	}

	return clone
}

// Copy creates a deep copy of the block
func (b *Block) Copy() *Block {
	return b.Clone()
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
func GetInitialDifficulty(blockType types.BlockType) uint32 {
	return calculateInitialDifficulty(blockType)
}

// calculateInitialDifficulty calculates the initial difficulty for a block type
func calculateInitialDifficulty(blockType types.BlockType) uint32 {
	switch blockType {
	case types.GoldenBlock:
		return 0x1d00ffff // Bitcoin's initial difficulty
	case types.SilverBlock:
		return 0x1d00ffff / 2 // Half of Bitcoin's initial difficulty
	default:
		return 0x1d00ffff
	}
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
		return fmt.Errorf("block weight exceeds maximum")
	}
	return nil
}

// CanAddTransaction checks if a transaction can be added to the block
func (b *Block) CanAddTransaction(tx *types.Transaction) bool {
	// Check if adding the transaction would exceed the maximum block weight
	newWeight := b.GetBlockWeight() + tx.Size()*4
	maxWeight := 4 * 1024 * 1024 // 4MB
	return newWeight <= maxWeight
}

// ValidateTransactions validates all transactions in the block
func (b *Block) ValidateTransactions(utxoSet types.UTXOSetInterface) error {
	// Check if block has transactions
	if len(b.Transactions) == 0 {
		return fmt.Errorf("block has no transactions")
	}

	// Check if first transaction is coinbase
	if !b.Transactions[0].IsCoinbase() {
		return fmt.Errorf("first transaction is not coinbase")
	}

	// Check if block has more than one coinbase transaction
	for i := 1; i < len(b.Transactions); i++ {
		if b.Transactions[i].IsCoinbase() {
			return fmt.Errorf("block has more than one coinbase transaction")
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
