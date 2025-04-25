package block

import (
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

// Block represents a blockchain block
type Block struct {
	Header       BlockHeader    `json:"header"`
	Transactions []*Transaction `json:"transactions"`
	Hash         []byte         `json:"hash"`
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

// NewBlock creates a new block
func NewBlock(prevBlockHash []byte, difficulty uint32) *Block {
	return &Block{
		Header: BlockHeader{
			Version:       1,
			PrevBlockHash: prevBlockHash,
			Timestamp:     time.Now(),
			Difficulty:    difficulty,
		},
		Transactions: make([]*Transaction, 0),
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
		Header: BlockHeader{
			Version:       b.Header.Version,
			PrevBlockHash: make([]byte, len(b.Header.PrevBlockHash)),
			MerkleRoot:    make([]byte, len(b.Header.MerkleRoot)),
			Timestamp:     b.Header.Timestamp,
			Difficulty:    b.Header.Difficulty,
			Nonce:         b.Header.Nonce,
		},
		Transactions: make([]*Transaction, len(b.Transactions)),
		Hash:         make([]byte, len(b.Hash)),
	}

	// Copy byte slices
	copy(blockCopy.Header.PrevBlockHash, b.Header.PrevBlockHash)
	copy(blockCopy.Header.MerkleRoot, b.Header.MerkleRoot)
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

// String returns a string representation of the block
func (b *Block) String() string {
	return hex.EncodeToString(b.Hash)
}

// UpdateMerkleRoot updates the merkle root of the block
func (b *Block) UpdateMerkleRoot() {
	b.Header.MerkleRoot = b.CalculateMerkleRoot()
}

// GetInitialDifficulty returns the initial difficulty for a block type
func GetInitialDifficulty(blockType BlockType) uint32 {
	switch blockType {
	case GoldenBlock:
		return 0x1d00ffff // Initial difficulty for golden block
	case SilverBlock:
		return 0x1d00ffff // Initial difficulty for silver block
	default:
		return 0x1d00ffff
	}
}

// calculateInitialDifficulty sets the initial difficulty based on block type
func calculateInitialDifficulty(blockType BlockType) uint32 {
	return GetInitialDifficulty(blockType)
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
