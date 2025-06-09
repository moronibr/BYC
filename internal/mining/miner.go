package mining

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"byc/internal/blockchain"
	"byc/internal/crypto"
	"byc/internal/wallet"
)

// WalletInfo stores wallet information for persistence
type WalletInfo struct {
	Address string
	Rewards map[string]float64 // map[coinType]amount
}

// Status represents the current mining status
type Status struct {
	HashRate         int64
	Shares           int64
	BlocksFound      int64
	Difficulty       int
	LastUpdate       time.Time
	MiningWallet     *wallet.Wallet
	Rewards          map[blockchain.CoinType]float64
	IsRunning        bool
	StartTime        time.Time
	EndTime          time.Time
	CurrentBlock     time.Time
	CurrentReward    float64
	TotalRewards     float64
	NetworkHashRate  int64
	AverageBlockTime float64
	ConnectedPeers   int
}

// Miner represents a mining node
type Miner struct {
	Blockchain *blockchain.Blockchain
	BlockType  blockchain.BlockType
	CoinType   blockchain.CoinType
	Address    string
	stopChan   chan struct{}
	wg         sync.WaitGroup
	status     Status
	mu         sync.RWMutex
	walletFile string
}

// NewMiner creates a new miner
func NewMiner(bc *blockchain.Blockchain, blockType blockchain.BlockType, coinType blockchain.CoinType, address string) (*Miner, error) {
	if !blockchain.IsMineable(coinType) {
		return nil, fmt.Errorf("coin type %s is not mineable", coinType)
	}

	// Create wallets directory if it doesn't exist
	walletsDir := "wallets"
	if err := os.MkdirAll(walletsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create wallets directory: %v", err)
	}

	// Try to load existing wallet
	walletFile := filepath.Join(walletsDir, "mining_wallet.json")
	var miningWallet *wallet.Wallet
	var rewards map[blockchain.CoinType]float64

	if _, err := os.Stat(walletFile); err == nil {
		// Load existing wallet
		data, err := os.ReadFile(walletFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read wallet file: %v", err)
		}

		var walletInfo WalletInfo
		if err := json.Unmarshal(data, &walletInfo); err != nil {
			return nil, fmt.Errorf("failed to parse wallet file: %v", err)
		}

		// Convert rewards map
		rewards = make(map[blockchain.CoinType]float64)
		for coinType, amount := range walletInfo.Rewards {
			rewards[blockchain.CoinType(coinType)] = amount
		}

		// Create wallet with existing address
		miningWallet = &wallet.Wallet{Address: walletInfo.Address}
	} else {
		// Create new wallet
		var err error
		miningWallet, err = wallet.NewWallet()
		if err != nil {
			return nil, fmt.Errorf("failed to create mining wallet: %v", err)
		}
		rewards = make(map[blockchain.CoinType]float64)
	}

	return &Miner{
		Blockchain: bc,
		BlockType:  blockType,
		CoinType:   coinType,
		Address:    address,
		stopChan:   make(chan struct{}),
		status: Status{
			Difficulty:   bc.Difficulty * blockchain.MiningDifficulty(coinType),
			MiningWallet: miningWallet,
			Rewards:      rewards,
		},
		walletFile: walletFile,
	}, nil
}

// saveWallet saves the wallet information to a file
func (m *Miner) saveWallet() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Convert rewards map to string keys for JSON
	rewards := make(map[string]float64)
	for coinType, amount := range m.status.Rewards {
		rewards[string(coinType)] = amount
	}

	walletInfo := WalletInfo{
		Address: m.status.MiningWallet.Address,
		Rewards: rewards,
	}

	data, err := json.MarshalIndent(walletInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal wallet info: %v", err)
	}

	if err := os.WriteFile(m.walletFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write wallet file: %v", err)
	}

	return nil
}

// calculateReward calculates the mining reward based on coin type and difficulty
func (m *Miner) calculateReward() float64 {
	baseReward := 1.0 // Base reward in the coin being mined

	// Adjust reward based on coin type
	switch m.CoinType {
	case blockchain.Leah:
		baseReward = 1.0
	case blockchain.Shiblum:
		baseReward = 0.5 // 1 Shiblum = 2 Leah
	case blockchain.Shiblon:
		baseReward = 0.25 // 1 Shiblon = 4 Leah
	case blockchain.Senum:
		baseReward = 0.125 // 1 Senum = 8 Leah
	case blockchain.Amnor:
		baseReward = 0.0625 // 1 Amnor = 16 Leah
	case blockchain.Ezrom:
		baseReward = 0.03125 // 1 Ezrom = 32 Leah
	case blockchain.Onti:
		baseReward = 0.015625 // 1 Onti = 64 Leah
	}

	// Adjust reward based on difficulty
	difficultyMultiplier := float64(m.status.Difficulty) / float64(m.Blockchain.Difficulty)
	reward := baseReward / difficultyMultiplier

	// Ensure minimum reward
	if reward < 0.0001 {
		reward = 0.0001
	}

	return reward
}

// mineBlock mines a new block
func (m *Miner) mineBlock() error {
	// Get pending transactions
	pendingTxs := m.Blockchain.GetPendingTransactions()

	// Create coinbase transaction
	coinbaseTx := blockchain.Transaction{
		ID:        []byte("coinbase"),
		Timestamp: time.Now(),
		Inputs:    []blockchain.TxInput{},
		Outputs: []blockchain.TxOutput{
			{
				Value:         m.calculateReward(),
				CoinType:      m.CoinType,
				PublicKeyHash: crypto.HashPublicKey(m.status.MiningWallet.PublicKey),
				Address:       m.status.MiningWallet.Address,
			},
		},
		BlockType: m.BlockType,
	}

	// Add coinbase transaction to pending transactions
	pendingTxs = append([]blockchain.Transaction{coinbaseTx}, pendingTxs...)

	// Mine block
	block, err := m.Blockchain.MineBlock(pendingTxs, m.BlockType, m.CoinType)
	if err != nil {
		return fmt.Errorf("failed to mine block: %v", err)
	}

	// Add block to blockchain
	if err := m.Blockchain.AddBlock(block); err != nil {
		return fmt.Errorf("failed to add block: %v", err)
	}

	// Update mining wallet with rewards
	m.status.Rewards[m.CoinType] += coinbaseTx.Outputs[0].Value

	// Save wallet
	if err := m.saveWallet(); err != nil {
		return fmt.Errorf("failed to save wallet: %v", err)
	}

	// Update status
	m.status.BlocksFound++
	m.status.CurrentBlock = time.Unix(block.Timestamp, 0)
	m.status.CurrentReward = coinbaseTx.Outputs[0].Value
	m.status.TotalRewards += coinbaseTx.Outputs[0].Value

	return nil
}

// Start starts the mining process
func (m *Miner) Start(ctx context.Context) {
	m.status.IsRunning = true
	m.status.StartTime = time.Now()

	go func() {
		for {
			select {
			case <-ctx.Done():
				m.Stop()
				return
			case <-m.stopChan:
				return
			default:
				if err := m.mineBlock(); err != nil {
					log.Printf("Mining error: %v", err)
					time.Sleep(time.Second)
					continue
				}
			}
		}
	}()
}

// Stop stops the mining process
func (m *Miner) Stop() {
	m.status.IsRunning = false
	m.status.EndTime = time.Now()
	close(m.stopChan)
}

// GetStatus returns the current mining status
func (m *Miner) GetStatus() Status {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Calculate hash rate
	if m.status.StartTime.IsZero() {
		m.status.HashRate = 0
	} else {
		elapsed := time.Since(m.status.StartTime).Seconds()
		if elapsed > 0 {
			m.status.HashRate = int64(float64(m.status.Shares) / elapsed)
		}
	}

	// Calculate network hash rate (placeholder - should be implemented)
	m.status.NetworkHashRate = m.status.HashRate * 100 // Placeholder

	// Calculate average block time
	if m.status.BlocksFound > 0 {
		elapsed := time.Since(m.status.StartTime).Seconds()
		m.status.AverageBlockTime = elapsed / float64(m.status.BlocksFound)
	}

	return m.status
}

// GetMiningDifficulty returns the current mining difficulty
func (m *Miner) GetMiningDifficulty() int {
	return m.Blockchain.Difficulty * blockchain.MiningDifficulty(m.CoinType)
}

// GetMiningStats returns current mining statistics
func (m *Miner) GetMiningStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"block_type":  m.BlockType,
		"coin_type":   m.CoinType,
		"difficulty":  m.status.Difficulty,
		"hash_rate":   m.status.HashRate,
		"shares":      m.status.Shares,
		"blocks":      m.status.BlocksFound,
		"address":     m.Address,
		"is_mining":   m.status.IsRunning,
		"wallet":      m.status.MiningWallet.Address,
		"rewards":     m.status.Rewards,
		"wallet_file": m.walletFile,
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
