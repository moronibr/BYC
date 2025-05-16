package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"
)

// Blockchain represents the dual blockchain system
type Blockchain struct {
	GoldenChain []Block
	SilverChain []Block
	Difficulty  int
}

// NewBlockchain creates a new blockchain with genesis blocks
func NewBlockchain() *Blockchain {
	genesisGolden := createGenesisBlock(GoldenBlock)
	genesisSilver := createGenesisBlock(SilverBlock)

	return &Blockchain{
		GoldenChain: []Block{genesisGolden},
		SilverChain: []Block{genesisSilver},
		Difficulty:  4, // Initial difficulty
	}
}

// createGenesisBlock creates the first block in a chain
func createGenesisBlock(blockType BlockType) Block {
	return Block{
		Timestamp:    time.Now().Unix(),
		Transactions: []Transaction{},
		PrevHash:     []byte{},
		Hash:         []byte{},
		Nonce:        0,
		BlockType:    blockType,
		Difficulty:   4,
	}
}

// AddBlock adds a new block to the appropriate chain
func (bc *Blockchain) AddBlock(block Block) error {
	if block.BlockType == GoldenBlock {
		if !bc.isValidBlock(block, bc.GoldenChain[len(bc.GoldenChain)-1]) {
			return errors.New("invalid golden block")
		}
		bc.GoldenChain = append(bc.GoldenChain, block)
	} else if block.BlockType == SilverBlock {
		if !bc.isValidBlock(block, bc.SilverChain[len(bc.SilverChain)-1]) {
			return errors.New("invalid silver block")
		}
		bc.SilverChain = append(bc.SilverChain, block)
	} else {
		return errors.New("invalid block type")
	}
	return nil
}

// isValidBlock checks if a block is valid
func (bc *Blockchain) isValidBlock(block, prevBlock Block) bool {
	if !bytes.Equal(block.PrevHash, prevBlock.Hash) {
		return false
	}

	if !bc.isValidProof(block) {
		return false
	}

	return true
}

// isValidProof checks if the block's proof of work is valid
func (bc *Blockchain) isValidProof(block Block) bool {
	hash := calculateHash(block)
	target := make([]byte, 32)
	for i := 0; i < block.Difficulty; i++ {
		target[i] = 0
	}
	return bytes.Compare(hash, target) == -1
}

// calculateHash calculates the hash of a block
func calculateHash(block Block) []byte {
	record := bytes.Join([][]byte{
		block.PrevHash,
		[]byte(string(block.BlockType)),
		[]byte(string(block.Difficulty)),
		[]byte(string(block.Nonce)),
	}, []byte{})

	h := sha256.New()
	h.Write(record)
	return h.Sum(nil)
}

// MineBlock mines a new block with the given transactions
func (bc *Blockchain) MineBlock(transactions []Transaction, blockType BlockType, coinType CoinType) (Block, error) {
	if !IsMineable(coinType) {
		return Block{}, errors.New("coin type is not mineable")
	}

	var prevBlock Block
	if blockType == GoldenBlock {
		prevBlock = bc.GoldenChain[len(bc.GoldenChain)-1]
	} else {
		prevBlock = bc.SilverChain[len(bc.SilverChain)-1]
	}

	block := Block{
		Timestamp:    time.Now().Unix(),
		Transactions: transactions,
		PrevHash:     prevBlock.Hash,
		Nonce:        0,
		BlockType:    blockType,
		Difficulty:   bc.Difficulty * MiningDifficulty(coinType),
	}

	// Proof of work
	for {
		block.Hash = calculateHash(block)
		if bc.isValidProof(block) {
			break
		}
		block.Nonce++
	}

	return block, nil
}

// GetBalance returns the balance of a wallet for a specific coin type
func (bc *Blockchain) GetBalance(address string, coinType CoinType) float64 {
	var balance float64

	// Check both chains for the balance
	for _, block := range bc.GoldenChain {
		for _, tx := range block.Transactions {
			for _, output := range tx.Outputs {
				if hex.EncodeToString(output.PubKeyHash) == address && output.CoinType == coinType {
					balance += output.Value
				}
			}
		}
	}

	for _, block := range bc.SilverChain {
		for _, tx := range block.Transactions {
			for _, output := range tx.Outputs {
				if hex.EncodeToString(output.PubKeyHash) == address && output.CoinType == coinType {
					balance += output.Value
				}
			}
		}
	}

	return balance
}

// CreateTransaction creates a new transaction
func (bc *Blockchain) CreateTransaction(from, to string, amount float64, coinType CoinType) (Transaction, error) {
	if amount <= 0 {
		return Transaction{}, errors.New("amount must be positive")
	}

	// Check if the coin can be transferred between blocks
	if !CanTransferBetweenBlocks(coinType) {
		blockType := GetBlockType(coinType)
		if blockType == "" {
			return Transaction{}, errors.New("invalid coin type")
		}
	}

	// Create transaction
	tx := Transaction{
		ID:        []byte{},
		Inputs:    []TxInput{},
		Outputs:   []TxOutput{},
		Timestamp: time.Now(),
		BlockType: GetBlockType(coinType),
	}

	// TODO: Implement input/output creation logic
	// This would involve finding unspent transaction outputs
	// and creating new outputs for the recipient

	return tx, nil
}
