package consensus

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/common"
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

	// Maximum transaction size in bytes
	MaxTransactionSize = 1000000 // 1MB

	// Maximum transaction data size in bytes
	MaxTransactionDataSize = 1000000 // 1MB
)

// BlockChain defines the interface for blockchain operations
type BlockChain interface {
	GetLastBlock() (*block.Block, error)
	GetBlockByHeight(height uint64) (*block.Block, error)
	GetBlock(hash []byte) (*block.Block, error)
	AddBlock(block *block.Block) error
	GetPendingTransactions() []*common.Transaction
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

	// Mining address
	miningAddress []byte

	// Blockchain
	blockchain BlockChain

	// Transaction pool
	txPool []*common.Transaction
}

// NewConsensus creates a new consensus engine
func NewConsensus() *Consensus {
	return &Consensus{
		txPool: make([]*common.Transaction, 0),
	}
}

// MineBlock mines a block
func (c *Consensus) MineBlock(block *block.Block) error {
	// Validate block time
	if err := c.validateBlockTime(block); err != nil {
		return err
	}

	// Add mining reward transaction
	if err := c.addMiningReward(block); err != nil {
		return err
	}

	// Mine block
	target := c.GetTarget()
	for block.Header.Nonce < math.MaxUint32 {
		// Calculate block hash
		block.Header.Hash = block.CalculateHash()

		// Check if hash meets target
		if bytes.Compare(block.Header.Hash, target) <= 0 {
			return nil
		}

		// Increment nonce
		block.Header.Nonce++
	}

	return errors.New("failed to find valid nonce")
}

// ValidateBlock validates a block
func (c *Consensus) ValidateBlock(block *block.Block) error {
	// Validate block time
	if err := c.validateBlockTime(block); err != nil {
		return err
	}

	// Validate block hash
	calculatedHash := block.CalculateHash()
	if !bytes.Equal(block.Header.Hash, calculatedHash) {
		return errors.New("invalid block hash")
	}

	// Validate block target
	target := c.GetTarget()
	if bytes.Compare(block.Header.Hash, target) > 0 {
		return errors.New("block hash does not meet target")
	}

	// Validate transactions
	for _, tx := range block.Transactions {
		if err := c.ValidateTransaction(tx); err != nil {
			return fmt.Errorf("invalid transaction: %v", err)
		}
	}

	// Verify mining reward
	if err := c.verifyMiningReward(block); err != nil {
		return fmt.Errorf("invalid mining reward: %v", err)
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

// addMiningReward adds a mining reward transaction to the block
func (c *Consensus) addMiningReward(block *block.Block) error {
	// Calculate block reward
	reward := c.calculateBlockReward(block.Header.Height)

	// Create mining reward transaction
	rewardTx := common.NewTransaction(
		nil, // From is nil for coinbase
		c.miningAddress,
		reward,
		nil, // Data is nil for coinbase
	)

	// Add transaction to block
	return block.AddTransaction(rewardTx)
}

// verifyMiningReward verifies the mining reward transaction
func (c *Consensus) verifyMiningReward(block *block.Block) error {
	if len(block.Transactions) == 0 {
		return fmt.Errorf("block has no transactions")
	}

	// Get first transaction (should be mining reward)
	rewardTx := block.Transactions[0]

	// Verify transaction amount
	expectedReward := c.calculateBlockReward(block.Header.Height)
	if rewardTx.Amount() != expectedReward {
		return fmt.Errorf("invalid mining reward amount: got %d, want %d", rewardTx.Amount(), expectedReward)
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

// SelectBestChain selects the best chain from multiple chains
func (c *Consensus) SelectBestChain(chains []*block.Block) *block.Block {
	if len(chains) == 0 {
		return nil
	}

	// Find chain with most work
	bestChain := chains[0]
	bestWork := c.calculateChainWork(bestChain)

	for _, chain := range chains[1:] {
		work := c.calculateChainWork(chain)
		if work > bestWork {
			bestChain = chain
			bestWork = work
		}
	}

	return bestChain
}

// calculateChainWork calculates the total work done in a chain
func (c *Consensus) calculateChainWork(block *block.Block) uint64 {
	work := uint64(0)
	current := block

	for current != nil {
		work += uint64(current.Header.Difficulty)
		// Get previous block
		prevBlock, err := c.blockchain.GetBlock(current.Header.PrevBlockHash)
		if err != nil {
			break
		}
		current = prevBlock
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

// ValidateTransaction validates a transaction
func (c *Consensus) ValidateTransaction(tx *common.Transaction) error {
	// Validate version
	if tx.Version() == 0 {
		return errors.New("invalid version")
	}

	// Validate timestamp
	if tx.Timestamp().IsZero() {
		return errors.New("invalid timestamp")
	}

	// Validate inputs
	if len(tx.Inputs()) == 0 {
		return errors.New("no inputs")
	}

	// Validate outputs
	if len(tx.Outputs()) == 0 {
		return errors.New("no outputs")
	}

	// Validate amount
	if tx.Amount() == 0 {
		return errors.New("invalid amount")
	}

	// Check minimum fee
	minFee := c.calculateMinimumFee(tx)
	if tx.Fee() < minFee {
		return fmt.Errorf("transaction fee too low: got %d, want at least %d", tx.Fee(), minFee)
	}

	return nil
}

// calculateMinimumFee calculates the minimum fee for a transaction
func (c *Consensus) calculateMinimumFee(tx *common.Transaction) uint64 {
	// Base fee per byte
	baseFeePerByte := uint64(1)

	// Calculate fee based on transaction size
	return uint64(tx.Size()) * baseFeePerByte
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

// ValidateChain validates a blockchain
func (c *Consensus) ValidateChain(chain interfaces.BlockChain) error {
	// Get last block
	lastBlock, err := chain.GetLastBlock()
	if err != nil {
		return err
	}

	// Validate each block
	current := lastBlock
	for current != nil {
		// Validate block
		if err := c.ValidateBlock(current); err != nil {
			return fmt.Errorf("invalid block at height %d: %v", current.Header.Height, err)
		}

		// Get previous block
		prevBlock, err := chain.GetBlock(current.Header.PrevBlockHash)
		if err != nil {
			break
		}
		current = prevBlock
	}

	return nil
}

// GetBlockTemplate creates a new block template
func (c *Consensus) GetBlockTemplate(chain interfaces.BlockChain, minerAddress string) (*block.Block, error) {
	// Get last block
	lastBlock, err := chain.GetLastBlock()
	if err != nil {
		return nil, err
	}

	// Create new block
	newBlock := block.NewBlock(lastBlock.Header.Hash, uint64(c.GetDifficulty()))
	newBlock.Header.Height = lastBlock.Header.Height + 1
	newBlock.Header.Timestamp = time.Now()

	// Add mining reward transaction
	reward := c.calculateBlockReward(newBlock.Header.Height)
	rewardTx := common.NewTransaction(
		nil,                  // From (empty for mining reward)
		[]byte(minerAddress), // To
		reward,               // Value
		nil,                  // Data
	)
	if err := newBlock.AddTransaction(rewardTx); err != nil {
		return nil, err
	}

	// Add pending transactions
	pendingTxs := chain.GetPendingTransactions()
	for _, tx := range pendingTxs {
		// Convert types.Transaction to common.Transaction
		commonTx := common.NewTransaction(
			[]byte(tx.Inputs[0].Address),  // From
			[]byte(tx.Outputs[0].Address), // To
			tx.Outputs[0].Value,           // Amount
			tx.Data,                       // Data
		)

		// Get the underlying transaction
		underlyingTx := commonTx.GetTransaction()

		// Copy inputs
		underlyingTx.Inputs = make([]*types.TxInput, len(tx.Inputs))
		for i, input := range tx.Inputs {
			underlyingTx.Inputs[i] = &types.TxInput{
				PreviousTxHash:  input.PreviousTxHash,
				PreviousTxIndex: input.PreviousTxIndex,
				ScriptSig:       input.ScriptSig,
				Sequence:        input.Sequence,
				Address:         input.Address,
			}
		}

		// Copy outputs
		underlyingTx.Outputs = make([]*types.TxOutput, len(tx.Outputs))
		for i, output := range tx.Outputs {
			underlyingTx.Outputs[i] = &types.TxOutput{
				Value:        output.Value,
				ScriptPubKey: output.ScriptPubKey,
				Address:      output.Address,
			}
		}

		if err := c.ValidateTransaction(commonTx); err != nil {
			continue
		}
		if err := newBlock.AddTransaction(commonTx); err != nil {
			continue
		}
	}

	return newBlock, nil
}

// AddBlock adds a block to the chain
func (c *Consensus) AddBlock(block *block.Block) error {
	// Validate block
	if err := c.ValidateBlock(block); err != nil {
		return err
	}

	// Add block to chain
	return c.blockchain.AddBlock(block)
}

// CreateBlock creates a new block
func (c *Consensus) CreateBlock() (*block.Block, error) {
	// Get last block
	lastBlock, err := c.blockchain.GetLastBlock()
	if err != nil {
		return nil, err
	}

	// Create new block
	newBlock := block.NewBlock(lastBlock.Header.Hash, uint64(c.GetDifficulty()))
	newBlock.Header.Timestamp = time.Now()
	newBlock.Header.Difficulty = uint32(c.GetDifficulty())

	// Add pending transactions
	for _, tx := range c.txPool {
		if err := newBlock.AddTransaction(tx); err != nil {
			continue
		}
	}

	return newBlock, nil
}

// GetTarget returns the current target difficulty
func (c *Consensus) GetTarget() []byte {
	return c.calculateTarget()
}
