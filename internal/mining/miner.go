package mining

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/byc/internal/blockchain"
	"github.com/byc/internal/wallet"
)

// WalletInfo stores wallet information for persistence
type WalletInfo struct {
	Address string
	Rewards map[string]float64 // map[coinType]amount
}

// Status represents the current mining status
type Status struct {
	HashRate     int64
	Shares       int64
	BlocksFound  int64
	Difficulty   int
	LastUpdate   time.Time
	MiningWallet *wallet.Wallet
	Rewards      map[blockchain.CoinType]float64
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

// Start starts the mining process
func (m *Miner) Start(ctx context.Context) {
	m.wg.Add(1)
	go m.mine(ctx)
}

// Stop stops the mining process
func (m *Miner) Stop() {
	close(m.stopChan)
	m.wg.Wait()

	// Save wallet before stopping
	if err := m.saveWallet(); err != nil {
		fmt.Printf("Warning: Failed to save wallet: %v\n", err)
	}
}

// GetStatus returns the current mining status
func (m *Miner) GetStatus() Status {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

// mine performs the mining process
func (m *Miner) mine(ctx context.Context) {
	defer m.wg.Done()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	hashCount := int64(0)
	lastHashCount := int64(0)
	lastUpdate := time.Now()
	lastBlockTime := time.Now()

	fmt.Printf("\nMining wallet address: %s\n", m.status.MiningWallet.Address)
	fmt.Printf("Mining rewards will be sent to this wallet\n")
	fmt.Printf("Wallet file location: %s\n", m.walletFile)
	fmt.Println("--------------------------------------------------------")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nMining stopped by user.")
			return
		case <-m.stopChan:
			fmt.Println("\nMining stopped by user.")
			return
		default:
			// Calculate time since last block
			timeSinceLastBlock := time.Since(lastBlockTime)

			// Only attempt to mine if enough time has passed (target 10 minutes per block)
			if timeSinceLastBlock < 600*time.Second {
				time.Sleep(100 * time.Millisecond) // Small delay to prevent CPU overload
				continue
			}

			// Calculate reward based on coin type (mirroring Bitcoin's original 50 BTC reward)
			var reward float64
			switch m.CoinType {
			case blockchain.Leah:
				reward = 50 // 50 Leah per block (every 10 minutes)
			case blockchain.Shiblum:
				reward = 25 // 25 Shiblum per block (every 10 minutes)
			case blockchain.Shiblon:
				reward = 12.5 // 12.5 Shiblon per block (every 10 minutes)
			case blockchain.Ephraim:
				// Check if we've reached the maximum Ephraim supply
				if m.Blockchain.GetTotalSupply(blockchain.Ephraim) >= blockchain.MaxEphraimSupply {
					fmt.Printf("Maximum Ephraim supply reached\n")
					continue
				}
				reward = 0.5 // 0.5 Ephraim per block (every 10 minutes)
			case blockchain.Manasseh:
				// Check if we've reached the maximum Manasseh supply
				if m.Blockchain.GetTotalSupply(blockchain.Manasseh) >= blockchain.MaxManassehSupply {
					fmt.Printf("Maximum Manasseh supply reached\n")
					continue
				}
				reward = 0.5 // 0.5 Manasseh per block (every 10 minutes)
			default:
				reward = 1 // 1 coin for other types (every 10 minutes)
			}

			// Attempt to mine a block
			block, err := m.Blockchain.MineBlock([]blockchain.Transaction{}, m.BlockType, m.CoinType)
			if err != nil {
				fmt.Printf("Error mining block: %v\n", err)
				continue
			}

			// Update hash count
			hashCount++

			// Update mining stats
			m.mu.Lock()
			m.status.Shares++
			m.mu.Unlock()

			// Add the block to the blockchain
			if err := m.Blockchain.AddBlock(block); err != nil {
				fmt.Printf("Error adding block: %v\n", err)
				continue
			}

			// Update mining stats and rewards
			m.mu.Lock()
			m.status.BlocksFound++
			m.status.Rewards[m.CoinType] += reward
			m.mu.Unlock()

			// Update last block time
			lastBlockTime = time.Now()

			// Save wallet after each successful block
			if err := m.saveWallet(); err != nil {
				fmt.Printf("Warning: Failed to save wallet: %v\n", err)
			}

			fmt.Printf("\nMined new block: %x\n", block.Hash)
			fmt.Printf("Reward: %.2f %s\n", reward, m.CoinType)
			fmt.Printf("Total rewards: %.2f %s\n", m.status.Rewards[m.CoinType], m.CoinType)
			fmt.Printf("Wallet balance: %.2f %s\n", m.Blockchain.GetBalance(m.status.MiningWallet.Address, m.CoinType), m.CoinType)
			fmt.Println("--------------------------------------------------------")
			fmt.Println("Mining in progress. Press Esc or 'q' to stop.")
		}

		// Update hash rate every second
		select {
		case <-ticker.C:
			now := time.Now()
			elapsed := now.Sub(lastUpdate).Seconds()
			if elapsed >= 1.0 { // Only update if at least 1 second has passed
				hashRate := (hashCount - lastHashCount) / int64(elapsed)
				m.mu.Lock()
				m.status.HashRate = hashRate
				m.status.LastUpdate = now
				m.mu.Unlock()
				lastHashCount = hashCount
				lastUpdate = now
			}
		default:
			// Continue mining
		}
	}
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
		"is_mining":   true,
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
