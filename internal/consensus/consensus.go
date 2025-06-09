package consensus

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"sync"
	"time"

	"byc/internal/blockchain"
)

// ConsensusType represents the type of consensus mechanism
type ConsensusType string

const (
	ProofOfWork  ConsensusType = "POW"
	ProofOfStake ConsensusType = "POS"
	Hybrid       ConsensusType = "HYBRID"
)

// ConsensusConfig holds configuration for consensus
type ConsensusConfig struct {
	Type             ConsensusType
	MinStakeAmount   float64
	StakeLockTime    time.Duration
	StakeRewardRate  float64
	ForkThreshold    int
	ReorgThreshold   int
	MaxForkLength    int
	StakeValidation  bool
	HybridWeight     float64 // Weight for hybrid consensus (0-1)
	StakeEpochLength int     // Number of blocks per stake epoch
	StakeEpochReward float64 // Reward per stake epoch
	StakeMinAge      int     // Minimum age of stake in blocks
	StakeMaxAge      int     // Maximum age of stake in blocks
}

// Stake represents a stake in the network
type Stake struct {
	Address     string
	Amount      float64
	CoinType    blockchain.CoinType
	StartTime   time.Time
	EndTime     time.Time
	BlockHeight int64
	Reward      float64
}

// Fork represents a blockchain fork
type Fork struct {
	StartHeight int64
	EndHeight   int64
	Blocks      []*blockchain.Block
	StakeWeight float64
	HashRate    float64
}

// ConsensusManager manages the consensus mechanism
type ConsensusManager struct {
	config        *ConsensusConfig
	blockchain    *blockchain.Blockchain
	stakes        map[string]*Stake
	forks         []*Fork
	mu            sync.RWMutex
	lastEpochTime time.Time
	epochRewards  map[string]float64
}

// NewConsensusConfig creates a new consensus configuration
func NewConsensusConfig() *ConsensusConfig {
	return &ConsensusConfig{
		Type:             Hybrid,
		MinStakeAmount:   1000,
		StakeLockTime:    30 * 24 * time.Hour, // 30 days
		StakeRewardRate:  0.05,                // 5% annual reward
		ForkThreshold:    6,                   // Similar to Bitcoin
		ReorgThreshold:   100,                 // Maximum reorg depth
		MaxForkLength:    1000,                // Maximum fork length
		StakeValidation:  true,
		HybridWeight:     0.5,   // Equal weight for PoW and PoS
		StakeEpochLength: 1000,  // 1000 blocks per epoch
		StakeEpochReward: 50,    // 50 coins per epoch
		StakeMinAge:      100,   // Minimum 100 blocks
		StakeMaxAge:      10000, // Maximum 10000 blocks
	}
}

// NewConsensusManager creates a new consensus manager
func NewConsensusManager(config *ConsensusConfig, bc *blockchain.Blockchain) *ConsensusManager {
	return &ConsensusManager{
		config:        config,
		blockchain:    bc,
		stakes:        make(map[string]*Stake),
		forks:         make([]*Fork, 0),
		lastEpochTime: time.Now(),
		epochRewards:  make(map[string]float64),
	}
}

// AddStake adds a new stake to the network
func (cm *ConsensusManager) AddStake(address string, amount float64, coinType blockchain.CoinType) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if amount < cm.config.MinStakeAmount {
		return fmt.Errorf("stake amount below minimum: %f", cm.config.MinStakeAmount)
	}

	stake := &Stake{
		Address:     address,
		Amount:      amount,
		CoinType:    coinType,
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(cm.config.StakeLockTime),
		BlockHeight: cm.blockchain.GetCurrentHeight(),
	}

	cm.stakes[address] = stake
	return nil
}

// RemoveStake removes a stake from the network
func (cm *ConsensusManager) RemoveStake(address string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	stake, exists := cm.stakes[address]
	if !exists {
		return errors.New("stake not found")
	}

	if time.Now().Before(stake.EndTime) {
		return errors.New("stake is still locked")
	}

	delete(cm.stakes, address)
	return nil
}

// GetStakeWeight calculates the stake weight for an address
func (cm *ConsensusManager) GetStakeWeight(address string) float64 {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	stake, exists := cm.stakes[address]
	if !exists {
		return 0
	}

	// Calculate stake age
	age := cm.blockchain.GetCurrentHeight() - stake.BlockHeight
	if age < int64(cm.config.StakeMinAge) {
		return 0
	}

	// Cap stake age
	if age > int64(cm.config.StakeMaxAge) {
		age = int64(cm.config.StakeMaxAge)
	}

	// Calculate weight based on amount and age
	weight := stake.Amount * float64(age)
	return weight
}

// ValidateBlock validates a block based on the consensus mechanism
func (cm *ConsensusManager) ValidateBlock(block *blockchain.Block) error {
	switch cm.config.Type {
	case ProofOfWork:
		return cm.validatePoWBlock(block)
	case ProofOfStake:
		return cm.validatePoSBlock(block)
	case Hybrid:
		return cm.validateHybridBlock(block)
	default:
		return errors.New("invalid consensus type")
	}
}

// validatePoWBlock validates a Proof of Work block
func (cm *ConsensusManager) validatePoWBlock(block *blockchain.Block) error {
	// Verify proof of work
	if !cm.verifyProofOfWork(block) {
		return errors.New("invalid proof of work")
	}
	return nil
}

// validatePoSBlock validates a Proof of Stake block
func (cm *ConsensusManager) validatePoSBlock(block *blockchain.Block) error {
	// Verify stake
	if !cm.verifyStake(block) {
		return errors.New("invalid stake")
	}
	return nil
}

// validateHybridBlock validates a Hybrid consensus block
func (cm *ConsensusManager) validateHybridBlock(block *blockchain.Block) error {
	// Verify both PoW and PoS
	if !cm.verifyProofOfWork(block) {
		return errors.New("invalid proof of work")
	}
	if !cm.verifyStake(block) {
		return errors.New("invalid stake")
	}
	return nil
}

// verifyProofOfWork verifies the proof of work for a block
func (cm *ConsensusManager) verifyProofOfWork(block *blockchain.Block) bool {
	// Calculate target difficulty
	target := cm.calculateTarget(block.Difficulty)

	// Verify block hash
	hash := cm.calculateBlockHash(block)
	return bytes.Compare(hash, target) <= 0
}

// verifyStake verifies the stake for a block
func (cm *ConsensusManager) verifyStake(block *blockchain.Block) bool {
	// Get miner's stake weight
	stakeWeight := cm.GetStakeWeight(block.MinerAddress)
	if stakeWeight == 0 {
		return false
	}

	// Calculate required stake weight
	requiredWeight := cm.calculateRequiredStakeWeight(block)
	return stakeWeight >= requiredWeight
}

// calculateTarget calculates the target difficulty
func (cm *ConsensusManager) calculateTarget(difficulty int) []byte {
	target := make([]byte, 32)
	for i := 0; i < difficulty; i++ {
		target[i] = 0
	}
	return target
}

// calculateBlockHash calculates the hash of a block
func (cm *ConsensusManager) calculateBlockHash(block *blockchain.Block) []byte {
	data := bytes.Join([][]byte{
		block.PrevHash,
		[]byte(fmt.Sprintf("%d", block.Timestamp)),
		[]byte(fmt.Sprintf("%d", block.Nonce)),
	}, []byte{})

	hash := sha256.Sum256(data)
	return hash[:]
}

// calculateRequiredStakeWeight calculates the required stake weight for a block
func (cm *ConsensusManager) calculateRequiredStakeWeight(block *blockchain.Block) float64 {
	// Base weight on block difficulty
	baseWeight := float64(block.Difficulty) * 1000

	// Adjust for network stake
	totalStake := cm.getTotalStake()
	if totalStake > 0 {
		baseWeight *= (totalStake / cm.config.MinStakeAmount)
	}

	return baseWeight
}

// getTotalStake returns the total stake in the network
func (cm *ConsensusManager) getTotalStake() float64 {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var total float64
	for _, stake := range cm.stakes {
		total += stake.Amount
	}
	return total
}

// HandleFork handles a blockchain fork
func (cm *ConsensusManager) HandleFork(newBlock *blockchain.Block) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Check if block creates a fork
	if !cm.isFork(newBlock) {
		return nil
	}

	// Create new fork
	fork := &Fork{
		StartHeight: cm.blockchain.GetCurrentHeight(),
		Blocks:      []*blockchain.Block{newBlock},
		StakeWeight: cm.calculateForkStakeWeight(newBlock),
		HashRate:    cm.calculateForkHashRate(newBlock),
	}

	// Add fork to list
	cm.forks = append(cm.forks, fork)

	// Check if fork should be accepted
	if cm.shouldAcceptFork(fork) {
		return cm.acceptFork(fork)
	}

	return nil
}

// isFork checks if a block creates a fork
func (cm *ConsensusManager) isFork(block *blockchain.Block) bool {
	// Check if block's previous hash matches current chain
	currentBlock := cm.blockchain.GetLatestBlock()
	return !bytes.Equal(block.PrevHash, currentBlock.Hash)
}

// calculateForkStakeWeight calculates the stake weight of a fork
func (cm *ConsensusManager) calculateForkStakeWeight(block *blockchain.Block) float64 {
	var weight float64
	for _, tx := range block.Transactions {
		if tx.IsCoinbase() {
			weight += cm.GetStakeWeight(tx.Outputs[0].Address)
		}
	}
	return weight
}

// calculateForkHashRate calculates the hash rate of a fork
func (cm *ConsensusManager) calculateForkHashRate(block *blockchain.Block) float64 {
	// Calculate hash rate based on block difficulty and time
	difficulty := float64(block.Difficulty)
	timeDiff := float64(block.Timestamp - cm.blockchain.GetLatestBlock().Timestamp)
	if timeDiff == 0 {
		return 0
	}
	return difficulty / timeDiff
}

// shouldAcceptFork determines if a fork should be accepted
func (cm *ConsensusManager) shouldAcceptFork(fork *Fork) bool {
	// Check fork length
	if len(fork.Blocks) > cm.config.MaxForkLength {
		return false
	}

	// Check stake weight
	if fork.StakeWeight < cm.getTotalStake()*0.51 {
		return false
	}

	// Check hash rate
	if fork.HashRate < cm.calculateForkHashRate(cm.blockchain.GetLatestBlock())*0.51 {
		return false
	}

	return true
}

// acceptFork accepts a fork and updates the blockchain
func (cm *ConsensusManager) acceptFork(fork *Fork) error {
	// Revert to fork start height
	if err := cm.blockchain.RevertToHeight(fork.StartHeight); err != nil {
		return err
	}

	// Add fork blocks
	for _, block := range fork.Blocks {
		if err := cm.blockchain.AddBlock(*block); err != nil {
			return err
		}
	}

	// Clear forks
	cm.forks = make([]*Fork, 0)

	return nil
}

// ProcessEpochRewards processes stake epoch rewards
func (cm *ConsensusManager) ProcessEpochRewards() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	currentHeight := cm.blockchain.GetCurrentHeight()
	if currentHeight%int64(cm.config.StakeEpochLength) != 0 {
		return nil
	}

	// Calculate rewards for each stake
	for address, stake := range cm.stakes {
		// Calculate stake age
		age := currentHeight - stake.BlockHeight
		if age < int64(cm.config.StakeMinAge) {
			continue
		}

		// Calculate reward
		reward := cm.calculateStakeReward(stake, age)
		cm.epochRewards[address] += reward
	}

	// Process rewards
	for address, reward := range cm.epochRewards {
		if err := cm.processStakeReward(address, reward); err != nil {
			return err
		}
	}

	// Clear rewards
	cm.epochRewards = make(map[string]float64)
	cm.lastEpochTime = time.Now()

	return nil
}

// calculateStakeReward calculates the reward for a stake
func (cm *ConsensusManager) calculateStakeReward(stake *Stake, age int64) float64 {
	// Base reward
	reward := cm.config.StakeEpochReward

	// Adjust for stake amount
	reward *= (stake.Amount / cm.config.MinStakeAmount)

	// Adjust for stake age
	ageFactor := float64(age) / float64(cm.config.StakeMaxAge)
	reward *= (1 + ageFactor)

	return reward
}

// processStakeReward processes a stake reward
func (cm *ConsensusManager) processStakeReward(address string, reward float64) error {
	// Create reward transaction
	tx := &blockchain.Transaction{
		Inputs: []blockchain.TxInput{},
		Outputs: []blockchain.TxOutput{
			{
				Value:         reward,
				CoinType:      cm.stakes[address].CoinType,
				PublicKeyHash: []byte(address),
				Address:       address,
			},
		},
		Timestamp: time.Now(),
	}

	// Add transaction to blockchain
	return cm.blockchain.AddTransaction(tx)
}
