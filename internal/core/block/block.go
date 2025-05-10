package block

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/youngchain/internal/core/coin"
	"github.com/youngchain/internal/core/utxo"
	"github.com/youngchain/internal/core/witness"
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
	Version      uint32
	PrevHash     []byte
	MerkleRoot   []byte
	Timestamp    time.Time
	Difficulty   uint64
	Nonce        uint32
	Hash         []byte
	Height       uint64
	BlockSize    int
	BlockWeight  int
	TxCount      int
	Transactions []*Transaction
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
	Witness   *witness.Witness
	Inputs    []TxInput
	Outputs   []TxOutput
	LockTime  uint64
	Version   uint32
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
	return fmt.Sprintf("Block{Version: %d, PrevHash: %s, MerkleRoot: %s, Timestamp: %s, Difficulty: %d, Nonce: %d, Hash: %s, Height: %d, Size: %d, Weight: %d, TxCount: %d}",
		b.Version,
		hex.EncodeToString(b.PrevHash),
		hex.EncodeToString(b.MerkleRoot),
		b.Timestamp.Format(time.RFC3339),
		b.Difficulty,
		b.Nonce,
		hex.EncodeToString(b.Hash),
		b.Height,
		b.BlockSize,
		b.BlockWeight,
		b.TxCount,
	)
}

// String returns a string representation of the header
func (h *Header) String() string {
	return fmt.Sprintf("Header{Version: %d, Height: %d, Difficulty: %d, Nonce: %d}",
		h.Version, h.Height, h.Difficulty, h.Nonce)
}

// NewBlock creates a new block
func NewBlock(prevHash []byte, height uint64) *Block {
	return &Block{
		Version:      1,
		PrevHash:     prevHash,
		Timestamp:    time.Now(),
		Difficulty:   0x1d00ffff,
		Height:       height,
		Transactions: make([]*Transaction, 0),
	}
}

// AddTransaction adds a transaction to the block
func (b *Block) AddTransaction(tx *Transaction) error {
	if !b.CanAddTransaction(tx) {
		return errors.New("block is full")
	}

	b.Transactions = append(b.Transactions, tx)
	b.TxCount++
	b.BlockSize = b.GetBlockSize()
	b.BlockWeight = b.GetBlockWeight()

	return nil
}

// CalculateHash calculates the block hash
func (b *Block) CalculateHash() []byte {
	hash := sha256.New()

	// Hash version
	versionBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(versionBytes, b.Version)
	hash.Write(versionBytes)

	// Hash previous block hash
	hash.Write(b.PrevHash)

	// Hash merkle root
	hash.Write(b.MerkleRoot)

	// Hash timestamp
	timeBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(timeBytes, uint64(b.Timestamp.Unix()))
	hash.Write(timeBytes)

	// Hash difficulty
	diffBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(diffBytes, b.Difficulty)
	hash.Write(diffBytes)

	// Hash nonce
	nonceBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(nonceBytes, b.Nonce)
	hash.Write(nonceBytes)

	return hash.Sum(nil)
}

// Validate validates the block
func (b *Block) Validate() error {
	// Validate version
	if b.Version == 0 {
		return errors.New("invalid version")
	}

	// Validate previous block hash
	if len(b.PrevHash) != 32 {
		return errors.New("invalid previous block hash")
	}

	// Validate merkle root
	if len(b.MerkleRoot) != 32 {
		return errors.New("invalid merkle root")
	}

	// Validate timestamp
	if b.Timestamp.IsZero() {
		return errors.New("invalid timestamp")
	}

	// Validate difficulty
	if b.Difficulty == 0 {
		return errors.New("invalid difficulty")
	}

	// Validate hash
	if len(b.Hash) != 32 {
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

	return nil
}

// Clone creates a deep copy of the block
func (b *Block) Clone() *Block {
	clone := &Block{
		Version:      b.Version,
		PrevHash:     append([]byte{}, b.PrevHash...),
		MerkleRoot:   append([]byte{}, b.MerkleRoot...),
		Timestamp:    b.Timestamp,
		Difficulty:   b.Difficulty,
		Nonce:        b.Nonce,
		Hash:         append([]byte{}, b.Hash...),
		Height:       b.Height,
		BlockSize:    b.BlockSize,
		BlockWeight:  b.BlockWeight,
		TxCount:      b.TxCount,
		Transactions: make([]*Transaction, len(b.Transactions)),
	}

	for i, tx := range b.Transactions {
		clone.Transactions[i] = tx.Clone()
	}

	return clone
}

// NewTransaction creates a new transaction
func NewTransaction(from, to string, amount uint64, nonce uint64) *Transaction {
	return &Transaction{
		From:      from,
		To:        to,
		Amount:    amount,
		Nonce:     nonce,
		Timestamp: time.Now(),
	}
}

// Validate validates the transaction
func (tx *Transaction) Validate() error {
	// Validate version
	if tx.Version == 0 {
		return errors.New("invalid version")
	}

	// Validate from address
	if tx.From == "" {
		return errors.New("invalid from address")
	}

	// Validate to address
	if tx.To == "" {
		return errors.New("invalid to address")
	}

	// Validate amount
	if tx.Amount == 0 {
		return errors.New("invalid amount")
	}

	// Validate nonce
	if tx.Nonce == 0 {
		return errors.New("invalid nonce")
	}

	// Validate signature
	if len(tx.Signature) == 0 {
		return errors.New("invalid signature")
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

	// Validate lock time
	if !tx.IsLockTimeValid() {
		return errors.New("invalid lock time")
	}

	return nil
}

// GetTransactionSize returns the size of the transaction in bytes
func (tx *Transaction) GetTransactionSize() int {
	size := 0

	// From address size
	size += len(tx.From)

	// To address size
	size += len(tx.To)

	// Amount size
	size += 8

	// Nonce size
	size += 8

	// Signature size
	size += len(tx.Signature)

	// Hash size
	size += len(tx.Hash)

	// Timestamp size
	size += 8

	// Witness size if present
	if tx.Witness != nil {
		size += tx.Witness.Size()
	}

	return size
}

// Clone creates a deep copy of the transaction
func (tx *Transaction) Clone() *Transaction {
	clone := &Transaction{
		From:      tx.From,
		To:        tx.To,
		Amount:    tx.Amount,
		Nonce:     tx.Nonce,
		Signature: append([]byte{}, tx.Signature...),
		Hash:      append([]byte{}, tx.Hash...),
		Timestamp: tx.Timestamp,
		Version:   tx.Version,
	}

	if tx.Witness != nil {
		clone.Witness = tx.Witness.Clone()
	}

	clone.Inputs = make([]TxInput, len(tx.Inputs))
	copy(clone.Inputs, tx.Inputs)

	clone.Outputs = make([]TxOutput, len(tx.Outputs))
	copy(clone.Outputs, tx.Outputs)

	return clone
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

// Copy creates a deep copy of the block
func (b *Block) Copy() *Block {
	// Create a new block
	blockCopy := &Block{
		Version:      b.Version,
		PrevHash:     make([]byte, len(b.PrevHash)),
		MerkleRoot:   make([]byte, len(b.MerkleRoot)),
		Timestamp:    b.Timestamp,
		Difficulty:   b.Difficulty,
		Nonce:        b.Nonce,
		Hash:         make([]byte, len(b.Hash)),
		Height:       b.Height,
		BlockSize:    b.BlockSize,
		BlockWeight:  b.BlockWeight,
		TxCount:      b.TxCount,
		Transactions: make([]*Transaction, len(b.Transactions)),
	}

	// Copy byte slices
	copy(blockCopy.PrevHash, b.PrevHash)
	copy(blockCopy.MerkleRoot, b.MerkleRoot)
	copy(blockCopy.Hash, b.Hash)

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

// AddInput adds an input to the transaction
func (tx *Transaction) AddInput(input TxInput) {
	tx.Inputs = append(tx.Inputs, input)
}

// AddOutput adds an output to the transaction
func (tx *Transaction) AddOutput(output TxOutput) {
	tx.Outputs = append(tx.Outputs, output)
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

// IsDoubleSpend checks if a transaction is a double spend
func (tx *Transaction) IsDoubleSpend(utxoSet *utxo.UTXOSet) bool {
	for _, input := range tx.Inputs {
		// Check if the input's previous output exists and is unspent
		utxo, exists := utxoSet.GetUTXO(input.PreviousTx, input.OutputIndex)
		if !exists || utxo == nil {
			return true // Double spend detected
		}
	}
	return false
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
		size += tx.GetTransactionSize()
	}

	// Hash size
	size += len(b.Hash)

	return size
}
