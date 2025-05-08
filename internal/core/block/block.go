package block

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/youngchain/internal/core/coin"
	"github.com/youngchain/internal/core/types"
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
	Header       *Header
	Transactions []*types.Transaction
	Hash         []byte
	Parent       *Block
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

// Transaction represents a transaction in the block
type Transaction struct {
	Version    uint32        `json:"version"`
	Inputs     []TxInput     `json:"inputs"`
	Outputs    []TxOutput    `json:"outputs"`
	LockTime   uint32        `json:"lock_time"`
	CoinType   coin.CoinType `json:"coin_type"`
	CrossBlock bool          `json:"cross_block"`
	Signature  []byte        `json:"signature"`
	Timestamp  time.Time     `json:"timestamp"`
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
		b.Header.Height, b.Hash, len(b.Transactions))
}

// String returns a string representation of the header
func (h *Header) String() string {
	return fmt.Sprintf("Header{Version: %d, Height: %d, Difficulty: %d, Nonce: %d}",
		h.Version, h.Height, h.Difficulty, h.Nonce)
}

// NewBlock creates a new block
func NewBlock(prevBlockHash []byte, difficulty uint32) *Block {
	return &Block{
		Header: &Header{
			Version:    1,
			PrevBlock:  prevBlockHash,
			Timestamp:  time.Now(),
			Difficulty: difficulty,
		},
		Transactions: make([]*types.Transaction, 0),
	}
}

// AddTransaction adds a transaction to the block
func (b *Block) AddTransaction(tx *types.Transaction) {
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
		hashes[i] = tx.CalculateHash()
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

// CalculateHash calculates the hash of the block
func (b *Block) CalculateHash() []byte {
	// Update Merkle root
	b.Header.MerkleRoot = b.CalculateMerkleRoot()

	// Serialize header
	headerBytes := []byte(b.Header.String())

	// Calculate hash
	hash := sha256.Sum256(headerBytes)
	return hash[:]
}

// Copy creates a deep copy of the block
func (b *Block) Copy() *Block {
	// Create a new block
	blockCopy := &Block{
		Header: &Header{
			Version:          b.Header.Version,
			PrevBlock:        make([]byte, len(b.Header.PrevBlock)),
			MerkleRoot:       make([]byte, len(b.Header.MerkleRoot)),
			Timestamp:        b.Header.Timestamp,
			Difficulty:       b.Header.Difficulty,
			Nonce:            b.Header.Nonce,
			Height:           b.Header.Height,
			TotalTxs:         b.Header.TotalTxs,
			StateRoot:        make([]byte, len(b.Header.StateRoot)),
			TransactionsRoot: make([]byte, len(b.Header.TransactionsRoot)),
			ReceiptsRoot:     make([]byte, len(b.Header.ReceiptsRoot)),
		},
		Transactions: make([]*types.Transaction, len(b.Transactions)),
		Hash:         make([]byte, len(b.Hash)),
	}

	// Copy byte slices
	copy(blockCopy.Header.PrevBlock, b.Header.PrevBlock)
	copy(blockCopy.Header.MerkleRoot, b.Header.MerkleRoot)
	copy(blockCopy.Header.StateRoot, b.Header.StateRoot)
	copy(blockCopy.Header.TransactionsRoot, b.Header.TransactionsRoot)
	copy(blockCopy.Header.ReceiptsRoot, b.Header.ReceiptsRoot)
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

// NewTransaction creates a new transaction
func NewTransaction(coinType coin.CoinType) *Transaction {
	return &Transaction{
		Version:    1,
		Inputs:     make([]TxInput, 0),
		Outputs:    make([]TxOutput, 0),
		LockTime:   0,
		CoinType:   coinType,
		CrossBlock: false,
		Signature:  make([]byte, 0),
		Timestamp:  time.Now(),
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

// Copy creates a deep copy of the transaction
func (tx *Transaction) Copy() *Transaction {
	txCopy := &Transaction{
		Version:    tx.Version,
		Inputs:     make([]TxInput, len(tx.Inputs)),
		Outputs:    make([]TxOutput, len(tx.Outputs)),
		LockTime:   tx.LockTime,
		CoinType:   tx.CoinType,
		CrossBlock: tx.CrossBlock,
		Signature:  make([]byte, len(tx.Signature)),
		Timestamp:  tx.Timestamp,
	}

	// Copy inputs
	for i, input := range tx.Inputs {
		txCopy.Inputs[i] = TxInput{
			PreviousTx:  make([]byte, len(input.PreviousTx)),
			OutputIndex: input.OutputIndex,
			Script:      make([]byte, len(input.Script)),
			Sequence:    input.Sequence,
		}
		copy(txCopy.Inputs[i].PreviousTx, input.PreviousTx)
		copy(txCopy.Inputs[i].Script, input.Script)
	}

	// Copy outputs
	for i, output := range tx.Outputs {
		txCopy.Outputs[i] = TxOutput{
			Value:    output.Value,
			Script:   make([]byte, len(output.Script)),
			CoinType: output.CoinType,
		}
		copy(txCopy.Outputs[i].Script, output.Script)
	}

	// Copy signature
	copy(txCopy.Signature, tx.Signature)

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
	size := 4 // Version
	size += 1 // VarInt for input count
	for _, input := range tx.Inputs {
		size += len(input.PreviousTx)
		size += 4 // OutputIndex
		size += len(input.Script)
		size += 4 // Sequence
	}
	size += 1 // VarInt for output count
	for _, output := range tx.Outputs {
		size += 8 // Value
		size += len(output.Script)
		size += len(string(output.CoinType))
	}
	size += 4 // LockTime
	size += len(string(tx.CoinType))
	size += 1 // CrossBlock
	size += len(tx.Signature)
	size += 8 // Timestamp
	return size
}

// Hash returns the hash of the transaction
func (tx *Transaction) Hash() []byte {
	// Create a byte slice to hold the serialized transaction
	data := make([]byte, 0, tx.Size())

	// Add version
	versionBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(versionBytes, tx.Version)
	data = append(data, versionBytes...)

	// Add inputs
	inputCountBytes := make([]byte, 1)
	inputCountBytes[0] = byte(len(tx.Inputs))
	data = append(data, inputCountBytes...)

	for _, input := range tx.Inputs {
		data = append(data, input.PreviousTx...)
		outputIndexBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(outputIndexBytes, input.OutputIndex)
		data = append(data, outputIndexBytes...)
		data = append(data, input.Script...)
		sequenceBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(sequenceBytes, input.Sequence)
		data = append(data, sequenceBytes...)
	}

	// Add outputs
	outputCountBytes := make([]byte, 1)
	outputCountBytes[0] = byte(len(tx.Outputs))
	data = append(data, outputCountBytes...)

	for _, output := range tx.Outputs {
		valueBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(valueBytes, output.Value)
		data = append(data, valueBytes...)
		data = append(data, output.Script...)
		data = append(data, string(output.CoinType)...)
	}

	// Add lock time
	lockTimeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(lockTimeBytes, tx.LockTime)
	data = append(data, lockTimeBytes...)

	// Add coin type
	data = append(data, string(tx.CoinType)...)

	// Add cross block flag
	if tx.CrossBlock {
		data = append(data, 1)
	} else {
		data = append(data, 0)
	}

	// Add signature
	data = append(data, tx.Signature...)

	// Add timestamp
	timestampBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(timestampBytes, uint64(tx.Timestamp.UnixNano()))
	data = append(data, timestampBytes...)

	// Calculate hash
	hash := sha256.Sum256(data)
	return hash[:]
}

// IsMature checks if the transaction is mature
func (tx *Transaction) IsMature() bool {
	// A transaction is mature if it's at least 1 hour old
	return time.Since(tx.Timestamp) >= time.Hour
}

// IsLockTimeValid checks if the transaction's lock time is valid
func (tx *Transaction) IsLockTimeValid() bool {
	// If lock time is 0, it's always valid
	if tx.LockTime == 0 {
		return true
	}

	// If lock time is a timestamp, check if it's in the past
	if tx.LockTime < 500000000 {
		return uint32(time.Now().Unix()) >= tx.LockTime
	}

	// If lock time is a block height, it's always valid (block height validation is done elsewhere)
	return true
}

// Size returns the size of the block in bytes
func (b *Block) Size() int {
	size := 0

	// Header size
	size += 4 // Version
	size += len(b.Header.PrevBlock)
	size += len(b.Header.MerkleRoot)
	size += 8 // Timestamp
	size += 4 // Difficulty
	size += 8 // Nonce
	size += 8 // Height
	size += 8 // TotalTxs
	size += len(b.Header.StateRoot)
	size += len(b.Header.TransactionsRoot)
	size += len(b.Header.ReceiptsRoot)

	// Transactions size
	for _, tx := range b.Transactions {
		size += tx.Size()
	}

	// Hash size
	size += len(b.Hash)

	return size
}
