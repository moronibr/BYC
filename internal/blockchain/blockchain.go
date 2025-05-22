package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"
)

// Blockchain represents the BYC blockchain
type Blockchain struct {
	GoldenBlocks []Block
	SilverBlocks []Block
	PendingTxs   []Transaction
	UTXOSet      *UTXOSet
	Difficulty   int
	MiningConfig *MiningConfig
	MiningPool   *MiningPool
	Blocks       []*Block
	mu           sync.RWMutex
}

// NewBlockchain creates a new blockchain
func NewBlockchain() *Blockchain {
	bc := &Blockchain{
		GoldenBlocks: make([]Block, 0),
		SilverBlocks: make([]Block, 0),
		PendingTxs:   make([]Transaction, 0),
		UTXOSet:      NewUTXOSet(),
		Difficulty:   1,
		MiningConfig: NewMiningConfig(),
		MiningPool:   NewMiningPool("main", "pool.byc"),
	}
	return bc
}

// createGenesisBlock creates the first block in a chain
func createGenesisBlock(blockType BlockType) Block {
	block := Block{
		Timestamp:    time.Now().Unix(),
		Transactions: []Transaction{},
		PrevHash:     []byte{},
		Nonce:        0,
		BlockType:    blockType,
		Difficulty:   1, // Set to 1 to match base difficulty
	}
	block.Hash = calculateHash(block)
	return block
}

// AddBlock adds a block to the blockchain
func (bc *Blockchain) AddBlock(block interface{}) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	var b Block
	switch v := block.(type) {
	case Block:
		b = v
	case *Block:
		b = *v
	default:
		return fmt.Errorf("invalid block type: %T", block)
	}

	// Validate block
	if err := bc.validateBlock(b); err != nil {
		return err
	}

	// Update UTXO set
	for _, tx := range b.Transactions {
		if err := bc.UTXOSet.UpdateWithTransaction(&tx); err != nil {
			return err
		}
	}

	// Add block to the appropriate chain
	if b.BlockType == GoldenBlock {
		bc.GoldenBlocks = append(bc.GoldenBlocks, b)
	} else {
		bc.SilverBlocks = append(bc.SilverBlocks, b)
	}

	// Also add to the Blocks slice for backward compatibility
	bc.Blocks = append(bc.Blocks, &b)
	return nil
}

// validateBlock validates a block before adding it to the blockchain
func (bc *Blockchain) validateBlock(block Block) error {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	// Get the previous block
	var prevBlock Block
	if block.BlockType == GoldenBlock {
		if len(bc.GoldenBlocks) == 0 {
			return errors.New("no previous block found for golden chain")
		}
		prevBlock = bc.GoldenBlocks[len(bc.GoldenBlocks)-1]
	} else {
		if len(bc.SilverBlocks) == 0 {
			return errors.New("no previous block found for silver chain")
		}
		prevBlock = bc.SilverBlocks[len(bc.SilverBlocks)-1]
	}

	// 1. Validate block structure
	if block.Timestamp <= prevBlock.Timestamp {
		return errors.New("block timestamp must be greater than previous block")
	}

	if block.Timestamp > time.Now().Unix()+60 { // Allow 60 seconds of future time
		return errors.New("block timestamp is too far in the future")
	}

	// 2. Validate block hash
	if !bytes.Equal(block.PrevHash, prevBlock.Hash) {
		return errors.New("previous block hash mismatch")
	}

	// 3. Validate proof of work
	if !bc.isValidProof(block) {
		return errors.New("invalid proof of work")
	}

	// 4. Validate transactions
	if len(block.Transactions) == 0 {
		return errors.New("block must contain at least one transaction")
	}

	// 5. Validate coinbase transaction
	coinbaseFound := false
	for _, tx := range block.Transactions {
		if tx.IsCoinbase() {
			if coinbaseFound {
				return errors.New("multiple coinbase transactions found")
			}
			coinbaseFound = true
		}
	}
	if !coinbaseFound {
		return errors.New("block must contain exactly one coinbase transaction")
	}

	// 6. Validate transaction signatures and amounts
	for _, tx := range block.Transactions {
		if !tx.Verify() {
			return fmt.Errorf("invalid transaction signature: %x", tx.ID)
		}

		// Skip validation for coinbase transaction
		if !tx.IsCoinbase() {
			// Validate transaction against UTXO set
			if !tx.Validate(bc.UTXOSet) {
				return fmt.Errorf("invalid transaction: %x", tx.ID)
			}

			// Check for double spending
			for _, input := range tx.Inputs {
				if !bc.UTXOSet.HasUTXO(string(input.TxID), input.OutputIndex) {
					return fmt.Errorf("double spending detected in transaction: %x", tx.ID)
				}
			}
		}
	}

	// 7. Validate block size
	blockSize := bc.calculateBlockSize(block)
	if blockSize > MaxBlockSize {
		return fmt.Errorf("block size exceeds maximum allowed size: %d > %d", blockSize, MaxBlockSize)
	}

	return nil
}

// calculateBlockSize calculates the size of a block in bytes
func (bc *Blockchain) calculateBlockSize(block Block) int64 {
	var size int64

	// Add block header size
	size += 8  // Timestamp
	size += 32 // PrevHash
	size += 32 // Hash
	size += 8  // Nonce
	size += int64(len(block.BlockType))
	size += 4 // Difficulty

	// Add transactions size
	for _, tx := range block.Transactions {
		size += int64(len(tx.ID))
		for _, input := range tx.Inputs {
			size += int64(len(input.TxID))
			size += 8 // OutputIndex
			size += 8 // Amount
			size += int64(len(input.Signature))
			size += int64(len(input.PublicKey))
			size += int64(len(input.Address))
		}
		for _, output := range tx.Outputs {
			size += 8 // Value
			size += int64(len(output.CoinType))
			size += int64(len(output.PublicKeyHash))
			size += int64(len(output.Address))
		}
	}

	return size
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
	// Check if the hash has enough leading zeros
	for i := 0; i < block.Difficulty; i++ {
		if hash[i] != 0 {
			return false
		}
	}
	return true
}

// calculateHash calculates the hash of a block
func calculateHash(block Block) []byte {
	record := bytes.Join([][]byte{
		block.PrevHash,
		[]byte(string(block.BlockType)),
		[]byte(strconv.Itoa(block.Difficulty)),
		[]byte(strconv.FormatUint(block.Nonce, 10)),
		[]byte(strconv.FormatInt(block.Timestamp, 10)),
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
				if hex.EncodeToString(output.PublicKeyHash) == address && output.CoinType == coinType {
					balance += output.Value
				}
			}
		}
	}

	for _, block := range bc.SilverBlocks {
		for _, tx := range block.Transactions {
			for _, output := range tx.Outputs {
				if hex.EncodeToString(output.PublicKeyHash) == address && output.CoinType == coinType {
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

// AddTransaction adds a transaction to the pending transactions
func (bc *Blockchain) AddTransaction(tx *Transaction) error {
	// Validate transaction
	if !tx.Verify() {
		return errors.New("invalid transaction signature")
	}

	// Validate transaction against UTXO set
	if !tx.Validate(bc.UTXOSet) {
		return errors.New("invalid transaction")
	}

	// Add transaction to pending transactions
	bc.mu.Lock()
	bc.PendingTxs = append(bc.PendingTxs, *tx)
	bc.mu.Unlock()

	// Update UTXO set
	return bc.UTXOSet.UpdateWithTransaction(tx)
}

// GetBlock retrieves a block by its hash
func (bc *Blockchain) GetBlock(hash []byte) (*Block, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	// Search in golden blocks
	for _, block := range bc.GoldenBlocks {
		if bytes.Equal(block.Hash, hash) {
			return &block, nil
		}
	}

	// Search in silver blocks
	for _, block := range bc.SilverBlocks {
		if bytes.Equal(block.Hash, hash) {
			return &block, nil
		}
	}

	return nil, fmt.Errorf("block not found")
}

// GetTransaction retrieves a transaction by its ID
func (bc *Blockchain) GetTransaction(id []byte) (*Transaction, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	// Search in golden blocks
	for _, block := range bc.GoldenBlocks {
		for _, tx := range block.Transactions {
			if bytes.Equal(tx.ID, id) {
				return &tx, nil
			}
		}
	}

	// Search in silver blocks
	for _, block := range bc.SilverBlocks {
		for _, tx := range block.Transactions {
			if bytes.Equal(tx.ID, id) {
				return &tx, nil
			}
		}
	}

	return nil, fmt.Errorf("transaction not found")
}

// GetTransactions retrieves all transactions for a given address
func (bc *Blockchain) GetTransactions(address string) ([]*Transaction, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	var transactions []*Transaction

	// Search in golden blocks
	for _, block := range bc.GoldenBlocks {
		for _, tx := range block.Transactions {
			// Check inputs
			for _, input := range tx.Inputs {
				if input.Address == address {
					transactions = append(transactions, &tx)
					break
				}
			}
			// Check outputs
			for _, output := range tx.Outputs {
				if output.Address == address {
					transactions = append(transactions, &tx)
					break
				}
			}
		}
	}

	// Search in silver blocks
	for _, block := range bc.SilverBlocks {
		for _, tx := range block.Transactions {
			// Check inputs
			for _, input := range tx.Inputs {
				if input.Address == address {
					transactions = append(transactions, &tx)
					break
				}
			}
			// Check outputs
			for _, output := range tx.Outputs {
				if output.Address == address {
					transactions = append(transactions, &tx)
					break
				}
			}
		}
	}

	return transactions, nil
}

// Height returns the current height of the blockchain
func (bc *Blockchain) Height() int {
	return len(bc.GoldenBlocks) + len(bc.SilverBlocks)
}

// Size returns the total size of the blockchain in bytes
func (bc *Blockchain) Size() int64 {
	var size int64
	for _, block := range bc.GoldenBlocks {
		size += int64(len(block.Transactions))
	}
	for _, block := range bc.SilverBlocks {
		size += int64(len(block.Transactions))
	}
	return size
}

// GetTotalSupply returns the total supply of a specific coin type
func (bc *Blockchain) GetTotalSupply(coinType CoinType) float64 {
	var total float64

	// Check both chains for the balance
	for _, block := range bc.GoldenBlocks {
		for _, tx := range block.Transactions {
			for _, output := range tx.Outputs {
				if output.CoinType == coinType {
					total += output.Value
				}
			}
		}
	}

	for _, block := range bc.SilverBlocks {
		for _, tx := range block.Transactions {
			for _, output := range tx.Outputs {
				if output.CoinType == coinType {
					total += output.Value
				}
			}
		}
	}

	return total
}

// GetCurrentHeight returns the current height of the blockchain
func (bc *Blockchain) GetCurrentHeight() int64 {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return int64(len(bc.Blocks))
}

// GetLatestBlock returns the latest block in the blockchain
func (bc *Blockchain) GetLatestBlock() *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	if len(bc.Blocks) == 0 {
		return nil
	}
	return bc.Blocks[len(bc.Blocks)-1]
}

// RevertToHeight reverts the blockchain to a specific height
func (bc *Blockchain) RevertToHeight(height int64) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if height < 0 || int64(len(bc.Blocks)) <= height {
		return fmt.Errorf("invalid height: %d", height)
	}

	bc.Blocks = bc.Blocks[:height+1]
	return nil
}

// DisplayGenesisBlock displays information about the Genesis block
func (bc *Blockchain) DisplayGenesisBlock() {
	fmt.Println("\n=== Golden Chain Genesis Block ===")
	if len(bc.GoldenBlocks) > 0 {
		genesis := bc.GoldenBlocks[0]
		fmt.Printf("Timestamp: %s\n", time.Unix(genesis.Timestamp, 0).Format(time.RFC3339))
		fmt.Printf("Hash: %x\n", genesis.Hash)
		fmt.Printf("Previous Hash: %x\n", genesis.PrevHash)
		fmt.Printf("Difficulty: %d\n", genesis.Difficulty)
		fmt.Printf("Block Type: %s\n", genesis.BlockType)
		fmt.Printf("Number of Transactions: %d\n", len(genesis.Transactions))
	}

	fmt.Println("\n=== Silver Chain Genesis Block ===")
	if len(bc.SilverBlocks) > 0 {
		genesis := bc.SilverBlocks[0]
		fmt.Printf("Timestamp: %s\n", time.Unix(genesis.Timestamp, 0).Format(time.RFC3339))
		fmt.Printf("Hash: %x\n", genesis.Hash)
		fmt.Printf("Previous Hash: %x\n", genesis.PrevHash)
		fmt.Printf("Difficulty: %d\n", genesis.Difficulty)
		fmt.Printf("Block Type: %s\n", genesis.BlockType)
		fmt.Printf("Number of Transactions: %d\n", len(genesis.Transactions))
	}
}
