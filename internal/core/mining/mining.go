package mining

import (
	"crypto/sha256"
	"encoding/binary"
	"math/big"
	"time"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/coin"
)

// MiningDifficulty represents the three-tier mining difficulty
type MiningDifficulty struct {
	LeahDifficulty    uint64
	ShiblumDifficulty uint64
	ShiblonDifficulty uint64
}

// Miner represents a mining node
type Miner struct {
	CurrentBlock *block.Block
	Difficulty   MiningDifficulty
	Target       *big.Int
	IsMining     bool
	StopChan     chan struct{}
}

// NewMiner creates a new miner
func NewMiner(block *block.Block) *Miner {
	return &Miner{
		CurrentBlock: block,
		Difficulty: MiningDifficulty{
			LeahDifficulty:    0x1d00ffff,      // Base difficulty
			ShiblumDifficulty: 0x1d00ffff >> 1, // 2x harder
			ShiblonDifficulty: 0x1d00ffff >> 2, // 4x harder
		},
		Target:   new(big.Int),
		StopChan: make(chan struct{}),
	}
}

// StartMining starts the mining process
func (m *Miner) StartMining(coinType coin.CoinType) {
	m.IsMining = true
	go m.mine(coinType)
}

// StopMining stops the mining process
func (m *Miner) StopMining() {
	m.IsMining = false
	close(m.StopChan)
}

// mine performs the actual mining
func (m *Miner) mine(coinType coin.CoinType) {
	target := m.CalculateTarget(coinType)
	nonce := uint64(0)

	for m.IsMining {
		select {
		case <-m.StopChan:
			return
		default:
			hash := m.CalculateHash(nonce)
			if new(big.Int).SetBytes(hash).Cmp(target) <= 0 {
				m.CurrentBlock.Nonce = nonce
				m.CurrentBlock.Hash = hash
				// TODO: Broadcast the mined block
				return
			}
			nonce++
		}
	}
}

// CalculateTarget calculates the target based on coin type
func (m *Miner) CalculateTarget(coinType coin.CoinType) *big.Int {
	var difficulty uint64
	switch coinType {
	case coin.Leah:
		difficulty = m.Difficulty.LeahDifficulty
	case coin.Shiblum:
		difficulty = m.Difficulty.ShiblumDifficulty
	case coin.Shiblon:
		difficulty = m.Difficulty.ShiblonDifficulty
	default:
		difficulty = m.Difficulty.LeahDifficulty
	}

	target := new(big.Int).Lsh(big.NewInt(1), 256)
	target = target.Div(target, big.NewInt(int64(difficulty)))
	return target
}

// CalculateHash calculates the hash for mining
func (m *Miner) CalculateHash(nonce uint64) []byte {
	header := make([]byte, 80)
	binary.LittleEndian.PutUint32(header[0:4], m.CurrentBlock.Version)
	copy(header[4:36], m.CurrentBlock.PreviousHash)
	copy(header[36:68], m.CurrentBlock.MerkleRoot)
	binary.LittleEndian.PutUint32(header[68:72], uint32(m.CurrentBlock.Timestamp))
	binary.LittleEndian.PutUint32(header[72:76], uint32(m.CurrentBlock.Difficulty))
	binary.LittleEndian.PutUint64(header[76:80], nonce)

	hash := sha256.Sum256(header)
	return hash[:]
}

// AdjustDifficulty adjusts the mining difficulty based on block time
func (m *Miner) AdjustDifficulty(blockTime time.Duration) {
	// Target block time is 10 minutes
	targetTime := time.Minute * 10
	adjustmentFactor := float64(blockTime) / float64(targetTime)

	// Adjust difficulties
	m.Difficulty.LeahDifficulty = uint64(float64(m.Difficulty.LeahDifficulty) * adjustmentFactor)
	m.Difficulty.ShiblumDifficulty = uint64(float64(m.Difficulty.ShiblumDifficulty) * adjustmentFactor)
	m.Difficulty.ShiblonDifficulty = uint64(float64(m.Difficulty.ShiblonDifficulty) * adjustmentFactor)
}
