package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/youngchain/internal/core/coin"
	"github.com/youngchain/internal/core/types"
)

const (
	// TargetBits is the number of leading zero bits required for a valid block
	TargetBits = 24

	// MaxNonce is the maximum value a nonce can have
	MaxNonce = math.MaxInt64
)

// Block represents a block in the blockchain
type Block struct {
	Timestamp     int64
	Transactions  []*types.Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
	Height        int64
}

// String returns a string representation of the block
func (b *Block) String() string {
	return fmt.Sprintf(
		"Block:\n"+
			"  Timestamp: %d\n"+
			"  Transactions: %d\n"+
			"  Previous Hash: %x\n"+
			"  Hash: %x\n"+
			"  Nonce: %d\n"+
			"  Height: %d\n",
		b.Timestamp,
		len(b.Transactions),
		b.PrevBlockHash,
		b.Hash,
		b.Nonce,
		b.Height,
	)
}

// NewBlock creates a new block
func NewBlock(transactions []*types.Transaction, prevBlockHash []byte, height int64) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(),
		Transactions:  transactions,
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
		Nonce:         0,
		Height:        height,
	}

	// Run the proof of work algorithm
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash
	block.Nonce = nonce

	return block
}

// NewGenesisBlock creates the first block in the blockchain
func NewGenesisBlock(coinbase *types.Transaction) *Block {
	return NewBlock([]*types.Transaction{coinbase}, []byte{}, 0)
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

	return nil
}

// Validate validates the block
func (b *Block) Validate() error {
	// Check if block is nil
	if b == nil {
		return fmt.Errorf("block is nil")
	}

	// Check if hash is valid
	if err := b.CalculateHash(); err != nil {
		return fmt.Errorf("failed to calculate hash: %v", err)
	}

	// Check if previous hash matches
	if b.PrevBlockHash != nil && !bytes.Equal(b.PrevBlockHash, b.PrevBlockHash) {
		return fmt.Errorf("previous hash mismatch")
	}

	// Check if timestamp is valid
	if b.Timestamp > int64(time.Now().Unix()) {
		return fmt.Errorf("invalid timestamp")
	}

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

	// Validate proof of work
	pow := NewProofOfWork(b)
	if !pow.Validate() {
		return fmt.Errorf("invalid proof of work")
	}

	// Validate block weight
	if err := b.ValidateBlockWeight(); err != nil {
		return fmt.Errorf("invalid block weight: %v", err)
	}

	// Validate merkle root
	calculatedMerkleRoot := b.CalculateMerkleRoot()
	if !bytes.Equal(calculatedMerkleRoot, b.Hash) {
		return fmt.Errorf("invalid merkle root")
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

	return nil
}

// GetHash returns the block hash
func (b *Block) GetHash() []byte {
	return b.Hash
}

// GetPreviousHash returns the previous block hash
func (b *Block) GetPreviousHash() []byte {
	return b.PrevBlockHash
}

// GetTimestamp returns the block timestamp
func (b *Block) GetTimestamp() uint64 {
	return uint64(b.Timestamp)
}

// GetDifficulty returns the block difficulty
func (b *Block) GetDifficulty() uint32 {
	return 0 // Assuming difficulty is not stored in the block structure
}

// GetCoinType returns the block coin type
func (b *Block) GetCoinType() coin.Type {
	return coin.Type("") // Assuming coin type is not stored in the block structure
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
		Timestamp:     b.Timestamp,
		Transactions:  make([]*types.Transaction, len(b.Transactions)),
		PrevBlockHash: append([]byte{}, b.PrevBlockHash...),
		Hash:          append([]byte{}, b.Hash...),
		Nonce:         b.Nonce,
		Height:        b.Height,
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
		HashHex: hex.EncodeToString(b.Hash),
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
		b.Hash = hash
	}
	return nil
}

// UpdateMerkleRoot updates the merkle root of the block
func (b *Block) UpdateMerkleRoot() {
	b.Hash = b.CalculateMerkleRoot()
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

	// Timestamp size
	size += 8

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

// HashTransactions returns a hash of the transactions in the block
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.Hash())
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}

// Serialize serializes the block into a byte array
func (b *Block) Serialize() []byte {
	var result bytes.Buffer

	// Write timestamp
	binary.Write(&result, binary.LittleEndian, b.Timestamp)

	// Write transactions
	binary.Write(&result, binary.LittleEndian, int64(len(b.Transactions)))
	for _, tx := range b.Transactions {
		txBytes := tx.Serialize()
		binary.Write(&result, binary.LittleEndian, int64(len(txBytes)))
		result.Write(txBytes)
	}

	// Write previous block hash
	binary.Write(&result, binary.LittleEndian, int64(len(b.PrevBlockHash)))
	result.Write(b.PrevBlockHash)

	// Write nonce
	binary.Write(&result, binary.LittleEndian, int64(b.Nonce))

	// Write height
	binary.Write(&result, binary.LittleEndian, b.Height)

	return result.Bytes()
}

// DeserializeBlock deserializes a block from a byte array
func DeserializeBlock(data []byte) (*Block, error) {
	block := &Block{}
	reader := bytes.NewReader(data)

	// Read timestamp
	binary.Read(reader, binary.LittleEndian, &block.Timestamp)

	// Read transactions
	var txCount int64
	binary.Read(reader, binary.LittleEndian, &txCount)
	block.Transactions = make([]*types.Transaction, txCount)
	for i := int64(0); i < txCount; i++ {
		var txLen int64
		binary.Read(reader, binary.LittleEndian, &txLen)
		txBytes := make([]byte, txLen)
		reader.Read(txBytes)
		tx, err := types.DeserializeTransaction(txBytes)
		if err != nil {
			return nil, err
		}
		block.Transactions[i] = tx
	}

	// Read previous block hash
	var prevHashLen int64
	binary.Read(reader, binary.LittleEndian, &prevHashLen)
	block.PrevBlockHash = make([]byte, prevHashLen)
	reader.Read(block.PrevBlockHash)

	// Read nonce
	var nonce int64
	binary.Read(reader, binary.LittleEndian, &nonce)
	block.Nonce = int(nonce)

	// Read height
	binary.Read(reader, binary.LittleEndian, &block.Height)

	return block, nil
}
