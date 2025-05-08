package consensus

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/types"
	"github.com/youngchain/internal/interfaces"
)

const (
	// Target time between blocks in seconds
	TargetBlockTime = 10 * time.Minute

	// Difficulty adjustment interval in blocks
	DifficultyAdjustmentInterval = 2016

	// Minimum difficulty
	MinDifficulty = 0x1d00ffff

	// Maximum difficulty
	MaxDifficulty = 0x1d00ffff * 4

	// Initial block reward in satoshis
	InitialBlockReward = 50 * 100000000

	// Block reward halving interval
	BlockRewardHalvingInterval = 210000

	// Maximum block size in bytes
	MaxBlockSize = 1000000 // 1MB

	// Maximum future block time in seconds
	MaxFutureBlockTime = 7200 // 2 hours
)

// BlockChain defines the interface for blockchain operations
type BlockChain interface {
	GetLastBlock() (*block.Block, error)
	GetBlockByHeight(height uint64) (*block.Block, error)
	GetBlock(hash []byte) (*block.Block, error)
	AddBlock(block *block.Block) error
	GetPendingTransactions() []*types.Transaction
}

// Consensus represents the consensus engine
type Consensus struct {
	mu sync.RWMutex

	// Current difficulty
	difficulty uint32

	// Last difficulty adjustment time
	lastAdjustmentTime time.Time

	// Last difficulty adjustment block height
	lastAdjustmentHeight uint64

	// Block time history for difficulty adjustment
	blockTimes []time.Duration

	// Current block reward
	blockReward uint64
}

// NewConsensus creates a new consensus engine
func NewConsensus() *Consensus {
	return &Consensus{}
}

// MineBlock mines a new block
func (c *Consensus) MineBlock(block *block.Block) error {
	// Validate block time
	if err := c.validateBlockTime(block); err != nil {
		return err
	}

	// Add mining reward transaction
	if err := c.addMiningReward(block); err != nil {
		return err
	}

	// Set initial nonce
	block.Header.Nonce = 0

	// Calculate target hash
	target := c.calculateTarget()

	// Mine block
	for {
		// Calculate block hash
		hash := block.CalculateHash()

		// Check if hash meets difficulty target
		if bytes.Compare(hash, target) <= 0 {
			block.Hash = hash
			return nil
		}

		// Increment nonce
		if block.Header.Nonce == math.MaxUint32 {
			return fmt.Errorf("failed to find valid nonce")
		}
		block.Header.Nonce++
	}
}

// ValidateBlock validates a block
func (c *Consensus) ValidateBlock(block *block.Block) error {
	// Verify block hash
	calculatedHash := block.CalculateHash()
	if !bytes.Equal(calculatedHash, block.Hash) {
		return fmt.Errorf("invalid block hash")
	}

	// Verify difficulty
	target := c.calculateTarget()
	if bytes.Compare(block.Hash, target) > 0 {
		return fmt.Errorf("block hash does not meet difficulty target")
	}

	// Verify block time
	if err := c.validateBlockTime(block); err != nil {
		return err
	}

	// Verify block size
	if block.Size() > MaxBlockSize {
		return fmt.Errorf("block size exceeds maximum")
	}

	// Verify mining reward
	if err := c.verifyMiningReward(block); err != nil {
		return err
	}

	return nil
}

// validateBlockTime validates the block timestamp
func (c *Consensus) validateBlockTime(block *block.Block) error {
	now := time.Now()

	// Check if block is too far in the future
	if block.Header.Timestamp.After(now.Add(MaxFutureBlockTime * time.Second)) {
		return fmt.Errorf("block timestamp is too far in the future")
	}

	// Check if block is too old
	if block.Header.Timestamp.Before(now.Add(-MaxFutureBlockTime * time.Second)) {
		return fmt.Errorf("block timestamp is too old")
	}

	return nil
}

// addMiningReward adds the mining reward transaction to the block
func (c *Consensus) addMiningReward(block *block.Block) error {
	// Calculate current block reward
	reward := c.calculateBlockReward(block.Header.Height)

	// Create mining reward transaction
	rewardTx := types.NewTransaction(
		nil,                     // From (empty for mining reward)
		[]byte("mining_reward"), // To
		reward,                  // Value
		nil,                     // Data
	)

	// Add transaction to block
	block.Transactions = append([]*types.Transaction{rewardTx}, block.Transactions...)

	return nil
}

// verifyMiningReward verifies the mining reward transaction
func (c *Consensus) verifyMiningReward(block *block.Block) error {
	if len(block.Transactions) == 0 {
		return fmt.Errorf("block has no transactions")
	}

	// Get first transaction (should be mining reward)
	rewardTx := block.Transactions[0]

	// Verify it's a mining reward transaction
	if len(rewardTx.From) != 0 {
		return fmt.Errorf("mining reward transaction has inputs")
	}

	// Calculate expected reward
	expectedReward := c.calculateBlockReward(block.Header.Height)

	// Verify reward amount
	if rewardTx.Value != expectedReward {
		return fmt.Errorf("invalid mining reward amount")
	}

	return nil
}

// calculateBlockReward calculates the block reward for a given height
func (c *Consensus) calculateBlockReward(height uint64) uint64 {
	// Calculate number of halvings
	halvings := height / BlockRewardHalvingInterval

	// Calculate reward
	reward := InitialBlockReward
	for i := uint64(0); i < halvings; i++ {
		reward = reward / 2
	}

	return uint64(reward)
}

// SelectBestChain selects the best chain from multiple candidates
func (c *Consensus) SelectBestChain(chains []*block.Block) *block.Block {
	if len(chains) == 0 {
		return nil
	}

	var bestChain *block.Block
	var maxWork uint64

	for _, chain := range chains {
		// Calculate chain work
		work := c.calculateChainWork(chain)
		if work > maxWork {
			maxWork = work
			bestChain = chain
		}
	}

	return bestChain
}

// calculateChainWork calculates the total work of a chain
func (c *Consensus) calculateChainWork(block *block.Block) uint64 {
	var work uint64
	current := block

	for current != nil {
		// Add block work
		work += uint64(math.Pow(2, float64(32-c.difficulty)))
		current = current.Parent
	}

	return work
}

// GetDifficulty returns the current difficulty
func (c *Consensus) GetDifficulty() uint32 {
	return c.difficulty
}

// AdjustDifficulty adjusts the difficulty based on block time
func (c *Consensus) AdjustDifficulty(block *block.Block) error {
	// Add block time to history
	if len(c.blockTimes) > 0 {
		lastBlockTime := c.blockTimes[len(c.blockTimes)-1]
		blockTime := time.Duration(block.Header.Timestamp.UnixNano() - int64(lastBlockTime))
		c.blockTimes = append(c.blockTimes, blockTime)
	} else {
		c.blockTimes = append(c.blockTimes, TargetBlockTime)
	}

	// Check if we need to adjust difficulty
	if len(c.blockTimes) >= DifficultyAdjustmentInterval {
		// Calculate average block time
		var totalTime time.Duration
		for _, t := range c.blockTimes {
			totalTime += t
		}
		avgBlockTime := totalTime / time.Duration(len(c.blockTimes))

		// Calculate difficulty adjustment
		adjustment := float64(TargetBlockTime) / float64(avgBlockTime)
		if adjustment > 4 {
			adjustment = 4
		} else if adjustment < 0.25 {
			adjustment = 0.25
		}

		// Adjust difficulty
		newDifficulty := uint32(float64(c.difficulty) * adjustment)
		if newDifficulty < MinDifficulty {
			newDifficulty = MinDifficulty
		} else if newDifficulty > MaxDifficulty {
			newDifficulty = MaxDifficulty
		}

		c.difficulty = newDifficulty
		c.lastAdjustmentTime = time.Now()
		c.lastAdjustmentHeight = block.Header.Height
		c.blockTimes = c.blockTimes[:0]
	}

	return nil
}

// calculateTarget calculates the target hash based on difficulty
func (c *Consensus) calculateTarget() []byte {
	// Convert difficulty to target
	target := make([]byte, 32)
	bits := make([]byte, 4)
	binary.BigEndian.PutUint32(bits, c.difficulty)

	// Extract mantissa and exponent
	exponent := bits[0]
	mantissa := binary.BigEndian.Uint32(bits[1:])

	// Calculate target
	if exponent <= 3 {
		mantissa >>= 8 * (3 - exponent)
	} else {
		mantissa <<= 8 * (exponent - 3)
	}

	// Set target bytes
	binary.BigEndian.PutUint32(target[28:], mantissa)

	return target
}

// ValidateTransaction validates a transaction for mining
func (c *Consensus) ValidateTransaction(tx *types.Transaction) error {
	// Check transaction size
	if tx.Size() > 1000000 { // 1MB limit
		return fmt.Errorf("transaction too large")
	}

	// Check transaction fee
	if tx.Fee < c.calculateMinimumFee(tx) {
		return fmt.Errorf("transaction fee too low")
	}

	return nil
}

// calculateMinimumFee calculates the minimum fee for a transaction
func (c *Consensus) calculateMinimumFee(tx *types.Transaction) uint64 {
	// Base fee per byte
	baseFeePerByte := uint64(1)

	// Calculate size-based fee
	size := tx.Size()
	fee := uint64(size) * baseFeePerByte

	// Add priority fee
	priority := c.calculatePriority(tx)
	if priority < 0.5 {
		fee *= 2 // Double fee for low priority
	}

	return fee
}

// calculatePriority calculates the transaction priority
func (c *Consensus) calculatePriority(tx *types.Transaction) float64 {
	// Priority is based on the time until the transaction becomes valid
	now := time.Now().Unix()
	if tx.Timestamp.Unix() <= now {
		return 1.0 // Already valid
	}

	// Calculate time until valid in hours
	timeUntilValid := float64(tx.Timestamp.Unix()-now) / 3600.0

	// Priority decreases as time until valid increases
	return 1.0 / timeUntilValid
}

// SyncChain synchronizes the blockchain with the network
func (c *Consensus) SyncChain(chain interfaces.BlockChain, network interfaces.Network) error {
	// Get current chain state
	lastBlock, err := chain.GetLastBlock()
	if err != nil {
		return fmt.Errorf("failed to get last block: %v", err)
	}

	// Get peer heights
	peerHeights := network.GetPeerHeights()
	if len(peerHeights) == 0 {
		return fmt.Errorf("no peers available for sync")
	}

	// Find highest peer height
	var highestHeight uint64
	for _, height := range peerHeights {
		if height > highestHeight {
			highestHeight = height
		}
	}

	// If we're already at the highest height, no sync needed
	if lastBlock.Header.Height >= highestHeight {
		return nil
	}

	// Request blocks from peers
	blocks, err := network.RequestBlocks(lastBlock.Header.Height+1, highestHeight)
	if err != nil {
		return fmt.Errorf("failed to request blocks: %v", err)
	}

	// Validate and add blocks
	for _, block := range blocks {
		if err := c.ValidateBlock(block); err != nil {
			return fmt.Errorf("invalid block at height %d: %v", block.Header.Height, err)
		}

		// Add block to chain
		if err := chain.AddBlock(block); err != nil {
			return fmt.Errorf("failed to add block at height %d: %v", block.Header.Height, err)
		}

		// Adjust difficulty
		if err := c.AdjustDifficulty(block); err != nil {
			return fmt.Errorf("failed to adjust difficulty: %v", err)
		}
	}

	return nil
}

// GetMiningRewardAddress returns the address for mining rewards
func (c *Consensus) GetMiningRewardAddress() string {
	// TODO: Implement proper mining reward address generation
	return "mining_reward_address"
}

// ValidateChain validates the entire blockchain
func (c *Consensus) ValidateChain(chain interfaces.BlockChain) error {
	// Get genesis block
	genesisBlock, err := chain.GetBlockByHeight(0)
	if err != nil {
		return fmt.Errorf("failed to get genesis block: %v", err)
	}

	// Get current chain state
	lastBlock, err := chain.GetLastBlock()
	if err != nil {
		return fmt.Errorf("failed to get last block: %v", err)
	}

	// Validate chain from genesis to tip
	current := lastBlock
	for current.Header.Height > 0 {
		// Validate block
		if err := c.ValidateBlock(current); err != nil {
			return fmt.Errorf("invalid block at height %d: %v", current.Header.Height, err)
		}

		// Get previous block
		prevBlock, err := chain.GetBlock(current.Header.PrevBlock)
		if err != nil {
			return fmt.Errorf("failed to get previous block: %v", err)
		}

		// Verify block links
		if !bytes.Equal(prevBlock.Hash, current.Header.PrevBlock) {
			return fmt.Errorf("invalid block link at height %d", current.Header.Height)
		}

		current = prevBlock
	}

	// Verify genesis block
	if err := c.ValidateBlock(genesisBlock); err != nil {
		return fmt.Errorf("invalid genesis block: %v", err)
	}

	return nil
}

// GetBlockTemplate creates a new block template for mining
func (c *Consensus) GetBlockTemplate(chain interfaces.BlockChain, minerAddress string) (*block.Block, error) {
	// Get last block
	lastBlock, err := chain.GetLastBlock()
	if err != nil {
		return nil, fmt.Errorf("failed to get last block: %v", err)
	}

	// Create new block
	newBlock := block.NewBlock(lastBlock.Hash, c.GetDifficulty())
	newBlock.Header.Height = lastBlock.Header.Height + 1
	newBlock.Header.Timestamp = time.Now()

	// Add mining reward transaction
	reward := c.calculateBlockReward(newBlock.Header.Height)
	rewardTx := types.NewTransaction(
		nil,                  // From (empty for mining reward)
		[]byte(minerAddress), // To
		reward,               // Value
		nil,                  // Data
	)
	newBlock.Transactions = append([]*types.Transaction{rewardTx}, newBlock.Transactions...)

	// Add pending transactions
	pendingTxs := chain.GetPendingTransactions()
	for _, tx := range pendingTxs {
		if err := c.ValidateTransaction(tx); err != nil {
			continue // Skip invalid transactions
		}
		newBlock.AddTransaction(tx)
	}

	// Update merkle root
	newBlock.UpdateMerkleRoot()

	return newBlock, nil
}
