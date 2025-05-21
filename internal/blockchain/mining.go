package blockchain

import (
	"math"
	"sync"
	"time"
)

// MiningConfig holds configuration for mining
type MiningConfig struct {
	TargetBlockTime  time.Duration
	DifficultyWindow int
	MaxDifficulty    int
	MinDifficulty    int
	AdjustmentFactor float64
	PoolShare        float64
	PoolFee          float64
	PoolMinPayout    float64
}

// MiningPool represents a mining pool
type MiningPool struct {
	ID            string
	Address       string
	TotalHashrate float64
	Miners        map[string]*Miner
	Shares        map[string]float64
	LastPayout    time.Time
	PoolShare     float64
	PoolFee       float64
	PoolMinPayout float64
	mu            sync.RWMutex
}

// Miner represents a miner in the pool
type Miner struct {
	ID            string
	Address       string
	Hashrate      float64
	Shares        float64
	LastShare     time.Time
	PendingPayout float64
}

// NewMiningConfig creates a new mining configuration
func NewMiningConfig() *MiningConfig {
	return &MiningConfig{
		TargetBlockTime:  2 * time.Minute,
		DifficultyWindow: 2016, // Similar to Bitcoin
		MaxDifficulty:    32,
		MinDifficulty:    1,
		AdjustmentFactor: 0.25, // 25% adjustment per window
		PoolShare:        0.95, // 95% to miners, 5% to pool
		PoolFee:          0.05, // 5% pool fee
		PoolMinPayout:    0.1,  // Minimum payout in base coin
	}
}

// NewMiningPool creates a new mining pool
func NewMiningPool(id, address string) *MiningPool {
	return &MiningPool{
		ID:            id,
		Address:       address,
		TotalHashrate: 0,
		Miners:        make(map[string]*Miner),
		Shares:        make(map[string]float64),
		LastPayout:    time.Now(),
		PoolShare:     0.95, // 95% to miners, 5% to pool
		PoolFee:       0.05, // 5% pool fee
		PoolMinPayout: 0.1,  // Minimum payout in base coin
	}
}

// AddMiner adds a miner to the pool
func (p *MiningPool) AddMiner(id, address string) *Miner {
	p.mu.Lock()
	defer p.mu.Unlock()

	miner := &Miner{
		ID:        id,
		Address:   address,
		Hashrate:  0,
		Shares:    0,
		LastShare: time.Now(),
	}
	p.Miners[id] = miner
	return miner
}

// RemoveMiner removes a miner from the pool
func (p *MiningPool) RemoveMiner(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.Miners, id)
}

// UpdateMinerStats updates a miner's statistics
func (p *MiningPool) UpdateMinerStats(id string, hashrate float64, shares float64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if miner, exists := p.Miners[id]; exists {
		miner.Hashrate = hashrate
		miner.Shares += shares
		miner.LastShare = time.Now()
		p.TotalHashrate += hashrate
	}
}

// CalculateDifficulty calculates the new difficulty based on recent block times
func (bc *Blockchain) CalculateDifficulty(blockType BlockType) int {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	var blocks []Block
	if blockType == GoldenBlock {
		blocks = bc.GoldenBlocks
	} else {
		blocks = bc.SilverBlocks
	}

	if len(blocks) < 2 {
		return bc.Difficulty
	}

	// Get the last N blocks for difficulty calculation
	windowSize := int(math.Min(float64(len(blocks)), float64(bc.MiningConfig.DifficultyWindow)))
	recentBlocks := blocks[len(blocks)-windowSize:]

	// Calculate average block time
	var totalTime int64
	for i := 1; i < len(recentBlocks); i++ {
		totalTime += recentBlocks[i].Timestamp - recentBlocks[i-1].Timestamp
	}
	avgBlockTime := float64(totalTime) / float64(len(recentBlocks)-1)

	// Calculate difficulty adjustment
	targetTime := float64(bc.MiningConfig.TargetBlockTime.Seconds())
	adjustment := targetTime / avgBlockTime

	// Apply adjustment factor to prevent large swings
	adjustment = 1 + (adjustment-1)*bc.MiningConfig.AdjustmentFactor

	// Calculate new difficulty
	newDifficulty := int(float64(bc.Difficulty) * adjustment)

	// Ensure difficulty stays within bounds
	if newDifficulty < bc.MiningConfig.MinDifficulty {
		newDifficulty = bc.MiningConfig.MinDifficulty
	} else if newDifficulty > bc.MiningConfig.MaxDifficulty {
		newDifficulty = bc.MiningConfig.MaxDifficulty
	}

	return newDifficulty
}

// CalculateMinerReward calculates the reward for a miner in the pool
func (p *MiningPool) CalculateMinerReward(minerID string, blockReward float64) float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()

	miner, exists := p.Miners[minerID]
	if !exists {
		return 0
	}

	// Calculate miner's share of the reward
	minerShare := (miner.Shares / p.TotalHashrate) * blockReward * p.PoolShare

	// Add to pending payout
	miner.PendingPayout += minerShare

	// Reset shares
	miner.Shares = 0

	return minerShare
}

// ProcessPayouts processes pending payouts to miners
func (p *MiningPool) ProcessPayouts() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, miner := range p.Miners {
		if miner.PendingPayout >= p.PoolMinPayout {
			// TODO: Implement actual payout transaction
			miner.PendingPayout = 0
		}
	}
	p.LastPayout = time.Now()
}

// GetPoolStats returns the current pool statistics
func (p *MiningPool) GetPoolStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return map[string]interface{}{
		"total_hashrate": p.TotalHashrate,
		"miner_count":    len(p.Miners),
		"last_payout":    p.LastPayout,
		"pool_fee":       p.PoolFee,
		"min_payout":     p.PoolMinPayout,
	}
}

// GetMinerStats returns statistics for a specific miner
func (p *MiningPool) GetMinerStats(minerID string) map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	miner, exists := p.Miners[minerID]
	if !exists {
		return nil
	}

	return map[string]interface{}{
		"hashrate":       miner.Hashrate,
		"shares":         miner.Shares,
		"last_share":     miner.LastShare,
		"pending_payout": miner.PendingPayout,
	}
}
