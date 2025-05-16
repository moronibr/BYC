package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"time"
)

// Blockchain represents the dual blockchain system
type Blockchain struct {
	GoldenBlocks []Block
	SilverBlocks []Block
	UTXOSet      *UTXOSet
	Difficulty   int
}

// NewBlockchain creates a new blockchain with genesis blocks
func NewBlockchain() *Blockchain {
	genesisGolden := createGenesisBlock(GoldenBlock)
	genesisSilver := createGenesisBlock(SilverBlock)

	return &Blockchain{
		GoldenBlocks: []Block{genesisGolden},
		SilverBlocks: []Block{genesisSilver},
		UTXOSet:      NewUTXOSet(),
		Difficulty:   4, // Initial difficulty
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

// AddBlock adds a block to the blockchain
func (bc *Blockchain) AddBlock(block Block) error {
	// Validate block
	if err := bc.validateBlock(block); err != nil {
		return err
	}

	// Update UTXO set
	for _, tx := range block.Transactions {
		// Remove spent UTXOs
		for _, input := range tx.Inputs {
			bc.UTXOSet.Remove(string(input.TxID), input.OutputIndex)
		}
		// Add new UTXOs
		for i, output := range tx.Outputs {
			utxo := UTXO{
				TxID:        string(tx.ID),
				OutputIndex: i,
				Amount:      output.Value,
				Address:     output.Address,
			}
			bc.UTXOSet.Add(utxo)
		}
	}

	// Add block to the appropriate chain
	if block.BlockType == GoldenBlock {
		bc.GoldenBlocks = append(bc.GoldenBlocks, block)
	} else {
		bc.SilverBlocks = append(bc.SilverBlocks, block)
	}
	return nil
}

// validateBlock validates a block before adding it to the blockchain
func (bc *Blockchain) validateBlock(block Block) error {
	// TODO: Implement block validation logic
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
		[]byte(strconv.Itoa(block.Difficulty)),
		[]byte(strconv.FormatInt(block.Nonce, 10)),
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
		prevBlock = bc.GoldenBlocks[len(bc.GoldenBlocks)-1]
	} else {
		prevBlock = bc.SilverBlocks[len(bc.SilverBlocks)-1]
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
	for _, block := range bc.GoldenBlocks {
		for _, tx := range block.Transactions {
			for _, output := range tx.Outputs {
				if hex.EncodeToString(output.PubKeyHash) == address && output.CoinType == coinType {
					balance += output.Value
				}
			}
		}
	}

	for _, block := range bc.SilverBlocks {
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
