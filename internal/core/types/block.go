package types

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"time"
)

// Block represents a block in the blockchain
type Block struct {
	Header       *BlockHeader
	Transactions []*Transaction
}

// BlockHeader represents the header of a block
type BlockHeader struct {
	Version    uint32
	PrevHash   string
	MerkleRoot string
	Timestamp  time.Time
	Difficulty uint32
	Nonce      uint32
	Hash       string
}

// NewBlock creates a new block
func NewBlock(prevHash string, difficulty uint32) *Block {
	return &Block{
		Header: &BlockHeader{
			Version:    1,
			PrevHash:   prevHash,
			Timestamp:  time.Now(),
			Difficulty: difficulty,
		},
		Transactions: make([]*Transaction, 0),
	}
}

// CalculateHash calculates the hash of the block
func (b *Block) CalculateHash() string {
	// Create a buffer to hold the header data
	data := make([]byte, 0)

	// Add version
	versionBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(versionBytes, b.Header.Version)
	data = append(data, versionBytes...)

	// Add previous hash
	prevHashBytes, _ := hex.DecodeString(b.Header.PrevHash)
	data = append(data, prevHashBytes...)

	// Add merkle root
	merkleRootBytes, _ := hex.DecodeString(b.Header.MerkleRoot)
	data = append(data, merkleRootBytes...)

	// Add timestamp
	timestampBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(timestampBytes, uint64(b.Header.Timestamp.Unix()))
	data = append(data, timestampBytes...)

	// Add difficulty
	difficultyBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(difficultyBytes, b.Header.Difficulty)
	data = append(data, difficultyBytes...)

	// Add nonce
	nonceBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(nonceBytes, b.Header.Nonce)
	data = append(data, nonceBytes...)

	// Calculate SHA-256 hash
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// AddTransaction adds a transaction to the block
func (b *Block) AddTransaction(tx *Transaction) {
	b.Transactions = append(b.Transactions, tx)
}

// CalculateMerkleRoot calculates the merkle root of the block's transactions
func (b *Block) CalculateMerkleRoot() string {
	if len(b.Transactions) == 0 {
		return ""
	}

	// Create a list of transaction hashes
	hashes := make([][]byte, len(b.Transactions))
	for i, tx := range b.Transactions {
		hash := sha256.Sum256([]byte(tx.Hash))
		hashes[i] = hash[:]
	}

	// Calculate merkle root
	for len(hashes) > 1 {
		var newHashes [][]byte
		for i := 0; i < len(hashes); i += 2 {
			var hash []byte
			if i+1 < len(hashes) {
				// Combine two hashes
				combined := append(hashes[i], hashes[i+1]...)
				hashSum := sha256.Sum256(combined)
				hash = hashSum[:]
			} else {
				// Duplicate the last hash
				combined := append(hashes[i], hashes[i]...)
				hashSum := sha256.Sum256(combined)
				hash = hashSum[:]
			}
			newHashes = append(newHashes, hash)
		}
		hashes = newHashes
	}

	return hex.EncodeToString(hashes[0])
}
