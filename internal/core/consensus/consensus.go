package consensus

import (
	"bytes"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/types"
)

// ConsensusConfig holds the configuration for the consensus mechanism
type ConsensusConfig struct {
	// Chain-specific parameters
	GoldenChainParams *ChainParams
	SilverChainParams *ChainParams

	// Consensus parameters
	TargetTimePerBlock   time.Duration
	DifficultyAdjustment int64
	MaxBlockSize         int
	MaxBlockWeight       int
}

// ChainParams holds parameters specific to each chain
type ChainParams struct {
	InitialDifficulty uint32
	MinDifficulty     uint32
	MaxDifficulty     uint32
	BlockReward       uint64
	HalvingInterval   int64
}

// Consensus represents the consensus mechanism
type Consensus struct {
	config      *ConsensusConfig
	goldenChain *ChainConsensus
	silverChain *ChainConsensus
	mu          sync.RWMutex
}

// ChainConsensus represents consensus for a single chain
type ChainConsensus struct {
	params     *ChainParams
	lastBlock  *block.Block
	difficulty uint32
	mu         sync.RWMutex
}

// NewConsensus creates a new consensus instance
func NewConsensus(config *ConsensusConfig) *Consensus {
	if config == nil {
		config = &ConsensusConfig{
			TargetTimePerBlock:   10 * time.Minute,
			DifficultyAdjustment: 2016,
			MaxBlockSize:         1000000,
			MaxBlockWeight:       4000000,
			GoldenChainParams: &ChainParams{
				InitialDifficulty: 0x1d00ffff,
				MinDifficulty:     0x1d00ffff,
				MaxDifficulty:     0x1d00ffff,
				BlockReward:       50,
				HalvingInterval:   210000,
			},
			SilverChainParams: &ChainParams{
				InitialDifficulty: 0x1d00ffff,
				MinDifficulty:     0x1d00ffff,
				MaxDifficulty:     0x1d00ffff,
				BlockReward:       50,
				HalvingInterval:   210000,
			},
		}
	}

	return &Consensus{
		config: config,
		goldenChain: &ChainConsensus{
			params:     config.GoldenChainParams,
			difficulty: config.GoldenChainParams.InitialDifficulty,
		},
		silverChain: &ChainConsensus{
			params:     config.SilverChainParams,
			difficulty: config.SilverChainParams.InitialDifficulty,
		},
	}
}

// ValidateBlock validates a block according to consensus rules
func (c *Consensus) ValidateBlock(block *block.Block) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Get the appropriate chain consensus
	var chainConsensus *ChainConsensus
	if block.Header.Type == block.GoldenBlock {
		chainConsensus = c.goldenChain
	} else if block.Header.Type == block.SilverBlock {
		chainConsensus = c.silverChain
	} else {
		return fmt.Errorf("invalid block type: %s", block.Header.Type)
	}

	// Validate block header
	if err := c.validateBlockHeader(block, chainConsensus); err != nil {
		return err
	}

	// Validate block size and weight
	if err := c.validateBlockSize(block); err != nil {
		return err
	}

	// Validate transactions
	if err := c.validateTransactions(block); err != nil {
		return err
	}

	// Validate proof of work
	if err := c.validateProofOfWork(block, chainConsensus); err != nil {
		return err
	}

	return nil
}

// validateBlockHeader validates the block header
func (c *Consensus) validateBlockHeader(block *block.Block, chainConsensus *ChainConsensus) error {
	// Validate version
	if block.Header.Version < 1 {
		return fmt.Errorf("invalid block version: %d", block.Header.Version)
	}

	// Validate timestamp
	if block.Header.Timestamp.After(time.Now().Add(2 * time.Hour)) {
		return fmt.Errorf("block timestamp too far in future")
	}

	// Validate previous block
	if chainConsensus.lastBlock != nil {
		if !bytes.Equal(block.Header.PrevBlockHash, chainConsensus.lastBlock.Header.Hash) {
			return fmt.Errorf("invalid previous block hash")
		}
		if block.Header.Height != chainConsensus.lastBlock.Header.Height+1 {
			return fmt.Errorf("invalid block height")
		}
	}

	return nil
}

// validateBlockSize validates the block size and weight
func (c *Consensus) validateBlockSize(block *block.Block) error {
	if block.BlockSize > c.config.MaxBlockSize {
		return fmt.Errorf("block size exceeds maximum")
	}
	if block.Weight > c.config.MaxBlockWeight {
		return fmt.Errorf("block weight exceeds maximum")
	}
	return nil
}

// validateTransactions validates the block's transactions
func (c *Consensus) validateTransactions(block *block.Block) error {
	if len(block.Transactions) == 0 {
		return fmt.Errorf("block has no transactions")
	}

	// Validate coinbase transaction
	if err := c.validateCoinbaseTx(block.Transactions[0], block.Header.Height); err != nil {
		return err
	}

	// Validate other transactions
	for i, tx := range block.Transactions[1:] {
		if err := c.validateTransaction(tx, i+1); err != nil {
			return fmt.Errorf("invalid transaction %d: %v", i+1, err)
		}
	}

	return nil
}

// validateProofOfWork validates the block's proof of work
func (c *Consensus) validateProofOfWork(block *block.Block, chainConsensus *ChainConsensus) error {
	// Calculate target difficulty
	target := new(big.Int).Lsh(big.NewInt(1), uint(256-chainConsensus.difficulty))

	// Calculate block hash
	hash := block.CalculateHash()
	hashInt := new(big.Int).SetBytes(hash)

	// Check if hash is less than target
	if hashInt.Cmp(target) > 0 {
		return fmt.Errorf("block hash does not meet difficulty requirement")
	}

	return nil
}

// UpdateDifficulty updates the difficulty for the next block
func (c *Consensus) UpdateDifficulty(block *block.Block) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var chainConsensus *ChainConsensus
	if block.Header.Type == block.GoldenBlock {
		chainConsensus = c.goldenChain
	} else if block.Header.Type == block.SilverBlock {
		chainConsensus = c.silverChain
	} else {
		return fmt.Errorf("invalid block type: %s", block.Header.Type)
	}

	// Only adjust difficulty every DifficultyAdjustment blocks
	if block.Header.Height%uint64(c.config.DifficultyAdjustment) != 0 {
		return nil
	}

	// Get the first block of the previous difficulty period
	firstBlock, err := c.getFirstBlockOfPeriod(block)
	if err != nil {
		return err
	}

	// Calculate time difference
	timeDiff := block.Header.Timestamp.Sub(firstBlock.Header.Timestamp)
	expectedTime := time.Duration(c.config.DifficultyAdjustment) * c.config.TargetTimePerBlock

	// Calculate new difficulty
	newDifficulty := chainConsensus.difficulty
	if timeDiff > expectedTime*2 {
		newDifficulty = newDifficulty / 2
	} else if timeDiff < expectedTime/2 {
		newDifficulty = newDifficulty * 2
	}

	// Ensure difficulty is within bounds
	if newDifficulty < chainConsensus.params.MinDifficulty {
		newDifficulty = chainConsensus.params.MinDifficulty
	} else if newDifficulty > chainConsensus.params.MaxDifficulty {
		newDifficulty = chainConsensus.params.MaxDifficulty
	}

	chainConsensus.difficulty = newDifficulty
	return nil
}

// getFirstBlockOfPeriod gets the first block of the previous difficulty period
func (c *Consensus) getFirstBlockOfPeriod(block *block.Block) (*block.Block, error) {
	// This is a placeholder - in a real implementation, you would need to
	// retrieve the block from storage
	return nil, fmt.Errorf("not implemented")
}

// validateCoinbaseTx validates the coinbase transaction
func (c *Consensus) validateCoinbaseTx(tx *types.Transaction, height uint64) error {
	// Validate coinbase transaction structure
	if len(tx.Inputs) != 1 || len(tx.Outputs) != 1 {
		return fmt.Errorf("invalid coinbase transaction structure")
	}

	// Validate coinbase input
	if tx.Inputs[0].PreviousTxIndex != 0xffffffff {
		return fmt.Errorf("invalid coinbase input")
	}

	// Validate block reward
	expectedReward := c.calculateBlockReward(height)
	if tx.Outputs[0].Value != expectedReward {
		return fmt.Errorf("invalid block reward: expected %d, got %d", expectedReward, tx.Outputs[0].Value)
	}

	return nil
}

// validateTransaction validates a regular transaction
func (c *Consensus) validateTransaction(tx *types.Transaction, index int) error {
	// Validate transaction structure
	if len(tx.Inputs) == 0 || len(tx.Outputs) == 0 {
		return fmt.Errorf("invalid transaction structure")
	}

	// Validate transaction size
	if tx.Size() > c.config.MaxBlockSize {
		return fmt.Errorf("transaction size exceeds maximum")
	}

	// Validate transaction weight
	if tx.Weight() > c.config.MaxBlockWeight {
		return fmt.Errorf("transaction weight exceeds maximum")
	}

	// Validate transaction hash
	if !bytes.Equal(tx.Hash(), tx.CalculateHash()) {
		return fmt.Errorf("invalid transaction hash")
	}

	return nil
}

// calculateBlockReward calculates the block reward for a given height
func (c *Consensus) calculateBlockReward(height uint64) uint64 {
	// This is a placeholder - in a real implementation, you would need to
	// implement the halving logic
	return 50
}
