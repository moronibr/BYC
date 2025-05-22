package blockchain

import (
	"crypto/sha256"
	"encoding/binary"
	"math"
	"math/big"
	"sync"
	"time"
)

// BlockHeader represents the header of a block
type BlockHeader struct {
	PreviousHash []byte
	MerkleRoot   []byte
	Timestamp    int64
	Difficulty   uint32
}

// MiningConfig holds mining configuration parameters
type MiningConfig struct {
	// Number of worker goroutines
	NumWorkers int
	// Target difficulty
	TargetDifficulty *big.Int
	// Block time target in seconds
	BlockTimeTarget int64
	// Maximum nonce value
	MaxNonce uint64
	// Target block time
	TargetBlockTime time.Duration
	// Difficulty window
	DifficultyWindow int
	// Maximum difficulty
	MaxDifficulty int
	// Minimum difficulty
	MinDifficulty int
	// Adjustment factor
	AdjustmentFactor float64
	// Pool share
	PoolShare float64
	// Pool fee
	PoolFee float64
	// Pool minimum payout
	PoolMinPayout float64
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

// MiningStats tracks mining statistics
type MiningStats struct {
	// Total hashes computed
	TotalHashes uint64
	// Current hash rate (hashes per second)
	HashRate float64
	// Time spent mining
	MiningTime time.Duration
	// Number of blocks found
	BlocksFound uint64
	// Last block found time
	LastBlockTime time.Time
}

// Miner represents a mining worker
type Miner struct {
	// Mining configuration
	config MiningConfig
	// Mining statistics
	stats MiningStats
	// Channel for stopping mining
	stopChan chan struct{}
	// Wait group for worker synchronization
	wg sync.WaitGroup
	// Mutex for thread-safe stats updates
	mu sync.RWMutex
	// ID of the miner
	ID string
	// Address of the miner
	Address string
	// Hashrate of the miner
	Hashrate float64
	// Shares mined by the miner
	Shares float64
	// Last share time
	LastShare time.Time
	// Pending payout for the miner
	PendingPayout float64
}

// NewMiningConfig creates a new mining configuration
func NewMiningConfig() *MiningConfig {
	return &MiningConfig{
		TargetBlockTime:  10 * time.Minute,
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
		if miner.Shares > 0 {
			miner.PendingPayout += miner.Shares * p.PoolShare
			miner.Shares = 0
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

// NewMiner creates a new miner instance
func NewMiner(config MiningConfig) *Miner {
	return &Miner{
		config:   config,
		stopChan: make(chan struct{}),
	}
}

// Start begins the mining process
func (m *Miner) Start(block *Block) {
	m.wg.Add(m.config.NumWorkers)
	startTime := time.Now()

	// Start worker goroutines
	for i := 0; i < m.config.NumWorkers; i++ {
		go m.worker(i, block, startTime)
	}
}

// Stop halts the mining process
func (m *Miner) Stop() {
	close(m.stopChan)
	m.wg.Wait()
}

// GetStats returns current mining statistics
func (m *Miner) GetStats() MiningStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stats
}

// worker performs the actual mining work
func (m *Miner) worker(id int, block *Block, startTime time.Time) {
	defer m.wg.Done()

	// Calculate nonce range for this worker
	startNonce := uint64(id) * m.config.MaxNonce / uint64(m.config.NumWorkers)
	endNonce := startNonce + m.config.MaxNonce/uint64(m.config.NumWorkers)

	// Create a copy of the block for this worker
	workerBlock := block.Copy()
	workerBlock.Nonce = startNonce

	// Pre-compute block header hash
	headerHash := workerBlock.HeaderHash()

	for {
		select {
		case <-m.stopChan:
			return
		default:
			// Update nonce
			workerBlock.Nonce++

			// Check if we've reached the end of our nonce range
			if workerBlock.Nonce >= endNonce {
				return
			}

			// Compute hash
			hash := sha256.Sum256(append(headerHash, binary.BigEndian.AppendUint64(nil, workerBlock.Nonce)...))
			hashInt := new(big.Int).SetBytes(hash[:])

			// Update statistics
			m.mu.Lock()
			m.stats.TotalHashes++
			m.stats.HashRate = float64(m.stats.TotalHashes) / time.Since(startTime).Seconds()
			m.mu.Unlock()

			// Check if we found a valid block
			if hashInt.Cmp(m.config.TargetDifficulty) <= 0 {
				m.mu.Lock()
				m.stats.BlocksFound++
				m.stats.LastBlockTime = time.Now()
				m.stats.MiningTime = time.Since(startTime)
				m.mu.Unlock()

				// Update the original block
				block.Nonce = workerBlock.Nonce
				block.Hash = hash[:]
				return
			}
		}
	}
}

// AdjustDifficulty adjusts the mining difficulty based on block time
func (m *Miner) AdjustDifficulty(actualBlockTime time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Calculate difficulty adjustment factor
	targetTime := time.Duration(m.config.BlockTimeTarget) * time.Second
	adjustmentFactor := float64(actualBlockTime) / float64(targetTime)

	// Adjust difficulty
	if adjustmentFactor > 1.1 {
		// Block time too high, decrease difficulty
		m.config.TargetDifficulty.Mul(m.config.TargetDifficulty, big.NewInt(9))
		m.config.TargetDifficulty.Div(m.config.TargetDifficulty, big.NewInt(10))
	} else if adjustmentFactor < 0.9 {
		// Block time too low, increase difficulty
		m.config.TargetDifficulty.Mul(m.config.TargetDifficulty, big.NewInt(10))
		m.config.TargetDifficulty.Div(m.config.TargetDifficulty, big.NewInt(9))
	}
}

// Copy creates a deep copy of a block
func (b *Block) Copy() *Block {
	return &Block{
		Timestamp:    b.Timestamp,
		Transactions: b.Transactions,
		PrevHash:     b.PrevHash,
		Hash:         b.Hash,
		Nonce:        b.Nonce,
		BlockType:    b.BlockType,
		Difficulty:   b.Difficulty,
	}
}

// HeaderHash computes the hash of the block header
func (b *Block) HeaderHash() []byte {
	data := make([]byte, 0)
	data = append(data, b.PrevHash...)
	data = append(data, b.Hash...)
	data = append(data, []byte(b.BlockType)...)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(b.Timestamp))
	data = append(data, buf...)
	binary.BigEndian.PutUint64(buf, b.Nonce)
	data = append(data, buf...)
	binary.BigEndian.PutUint32(buf[:4], uint32(b.Difficulty))
	data = append(data, buf[:4]...)
	hash := sha256.Sum256(data)
	return hash[:]
}
