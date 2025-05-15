package consensus

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"time"
)

// BlockValidator manages block validation
type BlockValidator struct {
	// Configuration
	minBlockSize       int
	maxBlockSize       int
	maxTransactions    int
	minTimestamp       int64
	maxFutureTime      time.Duration
	difficultyAdjuster *DifficultyAdjuster
}

// NewBlockValidator creates a new block validator
func NewBlockValidator(minBlockSize, maxBlockSize, maxTransactions int, minTimestamp int64, maxFutureTime time.Duration, difficultyAdjuster *DifficultyAdjuster) *BlockValidator {
	return &BlockValidator{
		minBlockSize:       minBlockSize,
		maxBlockSize:       maxBlockSize,
		maxTransactions:    maxTransactions,
		minTimestamp:       minTimestamp,
		maxFutureTime:      maxFutureTime,
		difficultyAdjuster: difficultyAdjuster,
	}
}

// ValidateBlock validates a block
func (bv *BlockValidator) ValidateBlock(block *Block, prevBlock *Block) error {
	// Validate block size
	if err := bv.validateBlockSize(block); err != nil {
		return fmt.Errorf("invalid block size: %v", err)
	}

	// Validate block header
	if err := bv.validateBlockHeader(block, prevBlock); err != nil {
		return fmt.Errorf("invalid block header: %v", err)
	}

	// Validate transactions
	if err := bv.validateTransactions(block); err != nil {
		return fmt.Errorf("invalid transactions: %v", err)
	}

	// Validate merkle root
	if err := bv.validateMerkleRoot(block); err != nil {
		return fmt.Errorf("invalid merkle root: %v", err)
	}

	// Validate proof of work
	if err := bv.validateProofOfWork(block); err != nil {
		return fmt.Errorf("invalid proof of work: %v", err)
	}

	return nil
}

// validateBlockSize validates the block size
func (bv *BlockValidator) validateBlockSize(block *Block) error {
	size := block.Size()
	if size < bv.minBlockSize {
		return fmt.Errorf("block size %d is below minimum %d", size, bv.minBlockSize)
	}
	if size > bv.maxBlockSize {
		return fmt.Errorf("block size %d exceeds maximum %d", size, bv.maxBlockSize)
	}
	return nil
}

// validateBlockHeader validates the block header
func (bv *BlockValidator) validateBlockHeader(block, prevBlock *Block) error {
	// Validate version
	if block.Version == 0 {
		return fmt.Errorf("invalid version")
	}

	// Validate timestamp
	if block.Timestamp < bv.minTimestamp {
		return fmt.Errorf("timestamp %d is before minimum %d", block.Timestamp, bv.minTimestamp)
	}
	if block.Timestamp > time.Now().Unix()+int64(bv.maxFutureTime.Seconds()) {
		return fmt.Errorf("timestamp %d is too far in future", block.Timestamp)
	}

	// Validate previous block hash
	if prevBlock != nil {
		if block.PrevHash != prevBlock.Hash {
			return fmt.Errorf("previous block hash mismatch")
		}
		if block.Height != prevBlock.Height+1 {
			return fmt.Errorf("invalid block height")
		}
	} else {
		if block.PrevHash != [32]byte{} {
			return fmt.Errorf("genesis block must have zero previous hash")
		}
		if block.Height != 0 {
			return fmt.Errorf("genesis block must have height 0")
		}
	}

	// Validate difficulty
	if err := bv.difficultyAdjuster.ValidateDifficulty(
		time.Unix(block.Timestamp, 0),
		[]time.Time{time.Unix(prevBlock.Timestamp, 0)},
		block.Difficulty,
	); err != nil {
		return fmt.Errorf("invalid difficulty: %v", err)
	}

	return nil
}

// validateTransactions validates the block transactions
func (bv *BlockValidator) validateTransactions(block *Block) error {
	// Check transaction count
	if len(block.Transactions) == 0 {
		return fmt.Errorf("block must contain at least one transaction")
	}
	if len(block.Transactions) > bv.maxTransactions {
		return fmt.Errorf("block contains too many transactions")
	}

	// Validate coinbase transaction
	if err := bv.validateCoinbase(block.Transactions[0]); err != nil {
		return fmt.Errorf("invalid coinbase transaction: %v", err)
	}

	// Validate other transactions
	for i, tx := range block.Transactions[1:] {
		if err := bv.validateTransaction(tx); err != nil {
			return fmt.Errorf("invalid transaction at index %d: %v", i+1, err)
		}
	}

	return nil
}

// validateCoinbase validates the coinbase transaction
func (bv *BlockValidator) validateCoinbase(tx *Transaction) error {
	if !tx.IsCoinbase() {
		return fmt.Errorf("first transaction must be coinbase")
	}
	if len(tx.Inputs) != 1 {
		return fmt.Errorf("coinbase must have exactly one input")
	}
	if len(tx.Outputs) == 0 {
		return fmt.Errorf("coinbase must have at least one output")
	}
	return nil
}

// validateTransaction validates a transaction
func (bv *BlockValidator) validateTransaction(tx *Transaction) error {
	if tx.IsCoinbase() {
		return fmt.Errorf("non-first transaction cannot be coinbase")
	}
	if len(tx.Inputs) == 0 {
		return fmt.Errorf("transaction must have at least one input")
	}
	if len(tx.Outputs) == 0 {
		return fmt.Errorf("transaction must have at least one output")
	}
	return nil
}

// validateMerkleRoot validates the merkle root
func (bv *BlockValidator) validateMerkleRoot(block *Block) error {
	calculatedRoot := block.CalculateMerkleRoot()
	if calculatedRoot != block.MerkleRoot {
		return fmt.Errorf("merkle root mismatch")
	}
	return nil
}

// validateProofOfWork validates the proof of work
func (bv *BlockValidator) validateProofOfWork(block *Block) error {
	// Calculate target
	target := bv.difficultyAdjuster.CalculateTarget(block.Difficulty)

	// Calculate block hash
	hash := sha256.Sum256(block.HeaderBytes())

	// Check if hash is below target
	for i := 0; i < 32; i++ {
		if hash[i] < target[i] {
			return nil
		}
		if hash[i] > target[i] {
			return fmt.Errorf("block hash exceeds target")
		}
	}

	return nil
}

// Block represents a block in the blockchain
type Block struct {
	Version      uint32
	PrevHash     [32]byte
	MerkleRoot   [32]byte
	Timestamp    int64
	Difficulty   uint32
	Height       uint64
	Nonce        uint32
	Hash         [32]byte
	Transactions []*Transaction
}

// Size returns the size of the block in bytes
func (b *Block) Size() int {
	size := 32 + 32 + 8 + 4 + 8 + 4 // Fixed header size
	for _, tx := range b.Transactions {
		size += tx.Size()
	}
	return size
}

// HeaderBytes returns the block header bytes
func (b *Block) HeaderBytes() []byte {
	data := make([]byte, 0, 80)
	data = binary.BigEndian.AppendUint32(data, b.Version)
	data = append(data, b.PrevHash[:]...)
	data = append(data, b.MerkleRoot[:]...)
	data = binary.BigEndian.AppendUint64(data, uint64(b.Timestamp))
	data = binary.BigEndian.AppendUint32(data, b.Difficulty)
	data = binary.BigEndian.AppendUint32(data, b.Nonce)
	return data
}

// CalculateMerkleRoot calculates the merkle root of the block
func (b *Block) CalculateMerkleRoot() [32]byte {
	if len(b.Transactions) == 0 {
		return [32]byte{}
	}

	// Create leaf hashes
	hashes := make([][]byte, len(b.Transactions))
	for i, tx := range b.Transactions {
		hashes[i] = tx.Hash[:]
	}

	// Build merkle tree
	for len(hashes) > 1 {
		var newHashes [][]byte
		for i := 0; i < len(hashes); i += 2 {
			if i+1 == len(hashes) {
				newHashes = append(newHashes, hashes[i])
				continue
			}
			combined := append(hashes[i], hashes[i+1]...)
			hash := sha256.Sum256(combined)
			newHashes = append(newHashes, hash[:])
		}
		hashes = newHashes
	}

	return [32]byte(hashes[0])
}

// Transaction represents a transaction in the blockchain
type Transaction struct {
	Version   uint32
	Inputs    []*TxInput
	Outputs   []*TxOutput
	LockTime  uint32
	Hash      [32]byte
	Timestamp int64
}

// Size returns the size of the transaction in bytes
func (t *Transaction) Size() int {
	size := 4 + 4 + 4 + 4 // Version + LockTime + Input count + Output count
	for _, input := range t.Inputs {
		size += input.Size()
	}
	for _, output := range t.Outputs {
		size += output.Size()
	}
	return size
}

// IsCoinbase checks if the transaction is a coinbase transaction
func (t *Transaction) IsCoinbase() bool {
	return len(t.Inputs) == 1 && t.Inputs[0].IsCoinbase()
}

// TxInput represents a transaction input
type TxInput struct {
	PrevHash  [32]byte
	PrevIndex uint32
	Script    []byte
	Sequence  uint32
}

// Size returns the size of the input in bytes
func (i *TxInput) Size() int {
	return 32 + 4 + len(i.Script) + 4
}

// IsCoinbase checks if the input is a coinbase input
func (i *TxInput) IsCoinbase() bool {
	return i.PrevHash == [32]byte{} && i.PrevIndex == 0xFFFFFFFF
}

// TxOutput represents a transaction output
type TxOutput struct {
	Value  uint64
	Script []byte
}

// Size returns the size of the output in bytes
func (o *TxOutput) Size() int {
	return 8 + len(o.Script)
}
