package mining

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/byc/internal/blockchain"
)

// Miner represents a mining node
type Miner struct {
	Blockchain *blockchain.Blockchain
	BlockType  blockchain.BlockType
	CoinType   blockchain.CoinType
	Address    string
	stopChan   chan struct{}
	wg         sync.WaitGroup
}

// NewMiner creates a new miner
func NewMiner(bc *blockchain.Blockchain, blockType blockchain.BlockType, coinType blockchain.CoinType, address string) (*Miner, error) {
	if !blockchain.IsMineable(coinType) {
		return nil, fmt.Errorf("coin type %s is not mineable", coinType)
	}

	return &Miner{
		Blockchain: bc,
		BlockType:  blockType,
		CoinType:   coinType,
		Address:    address,
		stopChan:   make(chan struct{}),
	}, nil
}

// Start starts the mining process
func (m *Miner) Start(ctx context.Context) {
	m.wg.Add(1)
	go m.mine(ctx)
}

// Stop stops the mining process
func (m *Miner) Stop() {
	close(m.stopChan)
	m.wg.Wait()
}

// mine performs the mining process
func (m *Miner) mine(ctx context.Context) {
	defer m.wg.Done()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.mineBlock()
		}
	}
}

// mineBlock attempts to mine a new block
func (m *Miner) mineBlock() {
	// Create a new block with pending transactions
	block, err := m.Blockchain.MineBlock([]blockchain.Transaction{}, m.BlockType, m.CoinType)
	if err != nil {
		fmt.Printf("Error mining block: %v\n", err)
		return
	}

	// Add the block to the blockchain
	if err := m.Blockchain.AddBlock(block); err != nil {
		fmt.Printf("Error adding block: %v\n", err)
		return
	}

	fmt.Printf("Mined new block: %x\n", block.Hash)
}

// GetMiningDifficulty returns the current mining difficulty
func (m *Miner) GetMiningDifficulty() int {
	return m.Blockchain.Difficulty * blockchain.MiningDifficulty(m.CoinType)
}

// GetMiningStats returns current mining statistics
func (m *Miner) GetMiningStats() map[string]interface{} {
	return map[string]interface{}{
		"block_type": m.BlockType,
		"coin_type":  m.CoinType,
		"difficulty": m.GetMiningDifficulty(),
		"address":    m.Address,
		"is_mining":  true,
	}
}

// MiningPool represents a pool of miners
type MiningPool struct {
	Miners  map[string]*Miner
	mu      sync.RWMutex
	Address string
}

// NewMiningPool creates a new mining pool
func NewMiningPool(address string) *MiningPool {
	return &MiningPool{
		Miners:  make(map[string]*Miner),
		Address: address,
	}
}

// AddMiner adds a miner to the pool
func (p *MiningPool) AddMiner(miner *Miner) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Miners[miner.Address] = miner
}

// RemoveMiner removes a miner from the pool
func (p *MiningPool) RemoveMiner(address string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.Miners, address)
}

// GetMiner returns a miner by address
func (p *MiningPool) GetMiner(address string) *Miner {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Miners[address]
}

// GetPoolStats returns statistics for the entire pool
func (p *MiningPool) GetPoolStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total_miners"] = len(p.Miners)
	stats["address"] = p.Address

	minerStats := make(map[string]interface{})
	for addr, miner := range p.Miners {
		minerStats[addr] = miner.GetMiningStats()
	}
	stats["miners"] = minerStats

	return stats
}
