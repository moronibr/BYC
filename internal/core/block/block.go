package block

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/youngchain/internal/core/coin"
)

// BlockType represents the type of block
type BlockType string

const (
	GoldenBlock BlockType = "golden"
	SilverBlock BlockType = "silver"
)

// BlockHeader represents the header of a block
type BlockHeader struct {
	Version       uint32    `json:"version"`
	PrevBlockHash []byte    `json:"prev_block_hash"`
	MerkleRoot    []byte    `json:"merkle_root"`
	Timestamp     time.Time `json:"timestamp"`
	Difficulty    uint32    `json:"difficulty"`
	Nonce         uint64    `json:"nonce"`
	Height        uint64    `json:"height"`
}

// String returns a string representation of the block header
func (h *BlockHeader) String() string {
	return fmt.Sprintf("%d:%x:%x:%d:%d:%d",
		h.Version,
		h.PrevBlockHash,
		h.MerkleRoot,
		h.Timestamp.Unix(),
		h.Difficulty,
		h.Nonce)
}

// Block represents a block in the blockchain
type Block struct {
	Height       uint64
	Timestamp    time.Time
	Transactions []*Transaction
	PrevHash     []byte
	Hash         []byte
	Nonce        uint64
	Difficulty   uint64
	MerkleRoot   []byte
}

// Header represents a block header
type Header struct {
	Version          uint32
	PrevBlock        []byte
	MerkleRoot       []byte
	Timestamp        time.Time
	Difficulty       uint32
	Nonce            uint32
	Height           uint64
	TotalTxs         uint64
	StateRoot        []byte
	TransactionsRoot []byte
	ReceiptsRoot     []byte
}

// Transaction represents a blockchain transaction
type Transaction struct {
	From      string
	To        string
	Amount    uint64
	Nonce     uint64
	Signature []byte
	Hash      []byte
	Timestamp time.Time
}

// TxInput represents a transaction input
type TxInput struct {
	PreviousTx  []byte `json:"previous_tx"`
	OutputIndex uint32 `json:"output_index"`
	Script      []byte `json:"script"`
	Sequence    uint32 `json:"sequence"`
}

// TxOutput represents a transaction output
type TxOutput struct {
	Value    uint64        `json:"value"`
	Script   []byte        `json:"script"`
	CoinType coin.CoinType `json:"coin_type"`
}

// String returns a string representation of the block
func (b *Block) String() string {
	return fmt.Sprintf("Block{Height: %d, Hash: %x, Transactions: %d}",
		b.Height, b.Hash, len(b.Transactions))
}

// String returns a string representation of the header
func (h *Header) String() string {
	return fmt.Sprintf("Header{Version: %d, Height: %d, Difficulty: %d, Nonce: %d}",
		h.Version, h.Height, h.Difficulty, h.Nonce)
}

// NewBlock creates a new block
func NewBlock(prevBlockHash []byte, difficulty uint32) *Block {
	return &Block{
		Height:       0,
		Timestamp:    time.Now(),
		Transactions: make([]*Transaction, 0),
		PrevHash:     prevBlockHash,
		Nonce:        0,
		Difficulty:   uint64(difficulty),
	}
}

// AddTransaction adds a transaction to the block
func (b *Block) AddTransaction(tx *Transaction) {
	b.Transactions = append(b.Transactions, tx)
}

// CalculateMerkleRoot calculates the Merkle root of the block's transactions
func (b *Block) CalculateMerkleRoot() []byte {
	if len(b.Transactions) == 0 {
		hash := sha256.Sum256([]byte{})
		return hash[:]
	}

	// Create a list of transaction hashes
	hashes := make([][]byte, len(b.Transactions))
	for i, tx := range b.Transactions {
		hashes[i] = tx.Hash
	}

	// Calculate Merkle root
	for len(hashes) > 1 {
		// If odd number of hashes, duplicate the last one
		if len(hashes)%2 != 0 {
			hashes = append(hashes, hashes[len(hashes)-1])
		}

		// Create new level of hashes
		newHashes := make([][]byte, len(hashes)/2)
		for i := 0; i < len(hashes); i += 2 {
			// Concatenate hashes
			combined := append(hashes[i], hashes[i+1]...)
			// Hash the concatenated hashes
			hash := sha256.Sum256(combined)
			newHashes[i/2] = hash[:]
		}
		hashes = newHashes
	}

	return hashes[0]
}

// CalculateHash computes the block hash
func (b *Block) CalculateHash() []byte {
	// Combine block data
	data := make([]byte, 0)
	data = append(data, binary.BigEndian.AppendUint64(nil, b.Height)...)

	// Convert timestamp to bytes
	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, uint64(b.Timestamp.UnixNano()))
	data = append(data, timestampBytes...)

	data = append(data, b.PrevHash...)
	data = append(data, binary.BigEndian.AppendUint64(nil, b.Nonce)...)
	data = append(data, binary.BigEndian.AppendUint64(nil, b.Difficulty)...)
	data = append(data, b.MerkleRoot...)

	// Add transaction hashes
	for _, tx := range b.Transactions {
		data = append(data, tx.Hash...)
	}

	// Calculate hash
	hash := sha256.Sum256(data)
	return hash[:]
}

// Copy creates a deep copy of the block
func (b *Block) Copy() *Block {
	// Create a new block
	blockCopy := &Block{
		Height:       b.Height,
		Timestamp:    b.Timestamp,
		Transactions: make([]*Transaction, len(b.Transactions)),
		PrevHash:     make([]byte, len(b.PrevHash)),
		Nonce:        b.Nonce,
		Difficulty:   b.Difficulty,
		MerkleRoot:   make([]byte, len(b.MerkleRoot)),
	}

	// Copy byte slices
	copy(blockCopy.PrevHash, b.PrevHash)
	copy(blockCopy.MerkleRoot, b.MerkleRoot)

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
	b.MerkleRoot = b.CalculateMerkleRoot()
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

// NewTransaction creates a new transaction
func NewTransaction(coinType coin.CoinType) *Transaction {
	return &Transaction{
		From:      "",
		To:        "",
		Amount:    0,
		Nonce:     0,
		Signature: nil,
		Hash:      nil,
	}
}

// AddInput adds an input to the transaction
func (tx *Transaction) AddInput(input TxInput) {
	// Implementation needed
}

// AddOutput adds an output to the transaction
func (tx *Transaction) AddOutput(output TxOutput) {
	// Implementation needed
}

// CalculateHash computes the transaction hash
func (tx *Transaction) CalculateHash() []byte {
	// Combine transaction data
	data := make([]byte, 0)
	data = append(data, []byte(tx.From)...)
	data = append(data, []byte(tx.To)...)
	data = append(data, binary.BigEndian.AppendUint64(nil, tx.Amount)...)
	data = append(data, binary.BigEndian.AppendUint64(nil, tx.Nonce)...)

	// Add timestamp
	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, uint64(tx.Timestamp.UnixNano()))
	data = append(data, timestampBytes...)

	// Calculate hash
	hash := sha256.Sum256(data)
	return hash[:]
}

// Copy creates a deep copy of the transaction
func (tx *Transaction) Copy() *Transaction {
	txCopy := &Transaction{
		From:      tx.From,
		To:        tx.To,
		Amount:    tx.Amount,
		Nonce:     tx.Nonce,
		Signature: make([]byte, len(tx.Signature)),
		Hash:      make([]byte, len(tx.Hash)),
		Timestamp: tx.Timestamp,
	}

	// Copy signature
	copy(txCopy.Signature, tx.Signature)

	// Copy hash
	copy(txCopy.Hash, tx.Hash)

	return txCopy
}

// MarshalJSON implements the json.Marshaler interface
func (tx *Transaction) MarshalJSON() ([]byte, error) {
	type Alias Transaction
	return json.Marshal(&struct {
		*Alias
		SignatureHex string `json:"signature_hex"`
	}{
		Alias:        (*Alias)(tx),
		SignatureHex: hex.EncodeToString(tx.Signature),
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (tx *Transaction) UnmarshalJSON(data []byte) error {
	type Alias Transaction
	aux := &struct {
		*Alias
		SignatureHex string `json:"signature_hex"`
	}{
		Alias: (*Alias)(tx),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.SignatureHex != "" {
		signature, err := hex.DecodeString(aux.SignatureHex)
		if err != nil {
			return err
		}
		tx.Signature = signature
	}
	return nil
}

// Size returns the size of the transaction in bytes
func (tx *Transaction) Size() int {
	size := 0
	size += len(tx.From)
	size += len(tx.To)
	size += 8 // Amount
	size += 8 // Nonce
	size += len(tx.Signature)
	return size
}

// VerifySignature verifies the transaction signature
func (tx *Transaction) VerifySignature(publicKey *ecdsa.PublicKey) bool {
	if tx.Signature == nil {
		return false
	}

	// Verify the signature using the transaction hash
	return ecdsa.VerifyASN1(publicKey, tx.Hash, tx.Signature)
}

// IsMature checks if the transaction is mature
func (tx *Transaction) IsMature() bool {
	// A transaction is mature if it's at least 1 hour old
	return time.Since(tx.Timestamp) >= time.Hour
}

// IsLockTimeValid checks if the transaction's lock time is valid
func (tx *Transaction) IsLockTimeValid() bool {
	// If lock time is 0, it's always valid
	if tx.Nonce == 0 {
		return true
	}

	// If lock time is a timestamp, check if it's in the past
	if tx.Nonce < 500000000 {
		return uint64(time.Now().Unix()) >= tx.Nonce
	}

	// If lock time is a block height, it's always valid (block height validation is done elsewhere)
	return true
}

// Size returns the size of the block in bytes
func (b *Block) Size() int {
	size := 0

	// Header size
	size += 8 // Height
	size += 8 // Timestamp
	size += len(b.PrevHash)
	size += 8 // Nonce
	size += 8 // Difficulty

	// Transactions size
	for _, tx := range b.Transactions {
		size += tx.Size()
	}

	// Hash size
	size += len(b.Hash)

	return size
}
