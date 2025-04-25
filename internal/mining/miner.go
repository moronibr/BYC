package mining

import (
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"sync"
	"time"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/logger"
)

// Target difficulty (can be adjusted based on network hash rate)
var maxTarget = new(big.Int).Exp(big.NewInt(2), big.NewInt(234), nil) // Roughly 2 minutes per block

// Miner represents a mining worker
type Miner struct {
	sync.Mutex
	running    bool
	address    string
	difficulty *big.Int
	workers    int
	workChan   chan *block.Block
	resultChan chan *block.Block
	quitChan   chan struct{}
}

// NewMiner creates a new miner instance
func NewMiner(address string, workers int) *Miner {
	return &Miner{
		address:    address,
		difficulty: new(big.Int).Set(maxTarget),
		workers:    workers,
		workChan:   make(chan *block.Block),
		resultChan: make(chan *block.Block),
		quitChan:   make(chan struct{}),
	}
}

// Start starts the mining process
func (m *Miner) Start() {
	m.Lock()
	defer m.Unlock()

	if m.running {
		return
	}

	m.running = true
	for i := 0; i < m.workers; i++ {
		go m.mine(i)
	}
}

// Stop stops the mining process
func (m *Miner) Stop() {
	m.Lock()
	defer m.Unlock()

	if !m.running {
		return
	}

	close(m.quitChan)
	m.running = false
}

// SubmitBlock submits a new block for mining
func (m *Miner) SubmitBlock(block *block.Block) {
	if !m.running {
		return
	}

	select {
	case m.workChan <- block:
	default:
		logger.Warn("Mining queue full, dropping block")
	}
}

// GetResult returns the result channel
func (m *Miner) GetResult() <-chan *block.Block {
	return m.resultChan
}

func (m *Miner) mine(workerID int) {
	logger.Info("Starting mining worker", logger.Int("worker_id", workerID))

	var nonce uint64
	for {
		select {
		case <-m.quitChan:
			return
		case work := <-m.workChan:
			// Copy block to avoid modifying the original
			candidate := work.Copy()
			startTime := time.Now()

			for {
				select {
				case <-m.quitChan:
					return
				default:
					// Update nonce
					candidate.Header.Nonce = nonce
					nonce++

					// Calculate hash
					hash := m.calculateHash(candidate)
					hashInt := new(big.Int).SetBytes(hash)

					// Check if hash meets difficulty
					if hashInt.Cmp(m.difficulty) <= 0 {
						candidate.Hash = hash
						logger.Info("Block mined",
							logger.Int("worker_id", workerID),
							logger.String("hash", hex.EncodeToString(hash)),
							logger.Duration("duration", time.Since(startTime)))

						select {
						case m.resultChan <- candidate:
						default:
							logger.Warn("Result channel full")
						}
						break
					}

					// Check if we should move to next block
					if nonce%1000 == 0 {
						select {
						case newWork := <-m.workChan:
							candidate = newWork.Copy()
							startTime = time.Now()
						default:
						}
					}
				}
			}
		}
	}
}

func (m *Miner) calculateHash(b *block.Block) []byte {
	header := b.Header
	data := []byte(header.String())
	hash := sha256.Sum256(data)
	return hash[:]
}

// SetDifficulty sets the mining difficulty
func (m *Miner) SetDifficulty(difficulty *big.Int) {
	m.Lock()
	defer m.Unlock()
	m.difficulty = difficulty
}

// GetHashrate returns the current hashrate
func (m *Miner) GetHashrate() float64 {
	// TODO: Implement hashrate calculation
	return 0
}
