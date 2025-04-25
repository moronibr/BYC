package main

import (
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/youngchain/internal/config"
	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/coin"
	"github.com/youngchain/internal/core/mining"
	"github.com/youngchain/internal/network"
)

var (
	// Mining configuration
	miningType    = flag.String("type", "solo", "Mining type (solo, pool)")
	poolAddress   = flag.String("pool", "", "Pool address (required for pool mining)")
	walletAddress = flag.String("wallet", "", "Wallet address to receive mining rewards (required for pool mining)")
	threads       = flag.Int("threads", runtime.NumCPU(), "Number of mining threads")
	coinType      = flag.String("coin", "leah", "Coin type to mine (leah, shiblum, shiblon)")

	// Network configuration
	nodeAddress = flag.String("node", "localhost:8333", "BYC node address to connect to")

	// Statistics
	hashrate    float64
	blocksFound int
	startTime   time.Time
	statsMutex  sync.RWMutex

	// Network state
	networkServer *network.Server
	transactions  []*block.Transaction
	txMutex       sync.RWMutex
	latestBlock   *block.Block
	blockMutex    sync.RWMutex
)

func main() {
	flag.Parse()

	// Validate mining type
	if *miningType != "solo" && *miningType != "pool" {
		log.Fatal("Invalid mining type. Must be 'solo' or 'pool'")
	}

	// Validate pool address and wallet address for pool mining
	if *miningType == "pool" {
		if *poolAddress == "" {
			log.Fatal("Pool address is required for pool mining")
		}
		if *walletAddress == "" {
			log.Fatal("Wallet address is required for pool mining")
		}
	}

	// Parse coin type
	var parsedCoinType coin.CoinType
	switch strings.ToLower(*coinType) {
	case "leah":
		parsedCoinType = coin.Leah
	case "shiblum":
		parsedCoinType = coin.Shiblum
	case "shiblon":
		parsedCoinType = coin.Shiblon
	default:
		log.Fatal("Invalid coin type. Must be one of: leah, shiblum, shiblon")
	}

	// Initialize network connection
	host, port := parseAddress(*nodeAddress)
	cfg := &config.Config{
		ListenAddr: fmt.Sprintf("%s:%d", host, port),
		MaxPeers:   10,
	}
	networkServer = network.NewServer(cfg)
	if err := networkServer.Start(); err != nil {
		log.Fatalf("Failed to start network server: %v", err)
	}
	defer networkServer.Stop()

	// Start network message handler
	go handleNetworkMessages()

	// Initialize mining with initial block
	latestBlock = block.NewBlock([]byte{}, block.GetInitialDifficulty(block.GoldenBlock))
	miner := mining.NewMiner(latestBlock)

	// Start mining threads
	var wg sync.WaitGroup
	stopChan := make(chan struct{})

	for i := 0; i < *threads; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			mineThread(id, miner, parsedCoinType, stopChan)
		}(i)
	}

	// Start statistics reporting
	startTime = time.Now()
	go reportStats()

	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Stop mining
	close(stopChan)
	wg.Wait()
}

func mineThread(threadID int, miner *mining.Miner, coinType coin.CoinType, stopChan chan struct{}) {
	nonce := uint64(threadID)
	hashes := uint64(0)
	lastReport := time.Now()

	for {
		select {
		case <-stopChan:
			return
		default:
			// Get latest block from network
			blockMutex.RLock()
			currentBlock := latestBlock
			blockMutex.RUnlock()

			// Update miner's current block
			miner.CurrentBlock = currentBlock

			// Add pending transactions to block
			txMutex.Lock()
			if len(transactions) > 0 {
				// Add transactions to block
				for _, tx := range transactions {
					currentBlock.AddTransaction(tx)
				}

				// Clear transaction pool
				transactions = []*block.Transaction{}

				// Update merkle root
				currentBlock.UpdateMerkleRoot()
			}
			txMutex.Unlock()

			// Calculate hash
			hash := miner.CalculateHash(nonce)
			hashes++

			// Check if hash meets difficulty target
			target := miner.CalculateTarget(coinType)
			if new(big.Int).SetBytes(hash).Cmp(target) <= 0 {
				// Found a valid block!
				currentBlock.Header.Nonce = nonce
				currentBlock.Hash = hash

				// Update statistics
				statsMutex.Lock()
				blocksFound++
				statsMutex.Unlock()

				// Submit block to network
				submitBlock(currentBlock)

				// Create new block for next mining round
				blockMutex.Lock()
				latestBlock = block.NewBlock(currentBlock.Hash, block.GetInitialDifficulty(block.GoldenBlock))
				blockMutex.Unlock()

				log.Printf("Thread %d found a valid block! Hash: %x", threadID, hash)
			}

			// Report hashrate every second
			if time.Since(lastReport) >= time.Second {
				statsMutex.Lock()
				hashrate = float64(hashes) / time.Since(lastReport).Seconds()
				statsMutex.Unlock()
				hashes = 0
				lastReport = time.Now()
			}

			nonce += uint64(*threads)
		}
	}
}

func parseAddress(address string) (string, int) {
	parts := strings.Split(address, ":")
	if len(parts) != 2 {
		log.Fatal("Invalid address format. Must be host:port")
	}
	port := 0
	fmt.Sscanf(parts[1], "%d", &port)
	return parts[0], port
}

func handleNetworkMessages() {
	// TODO: Implement network message handling
	// - Handle incoming blocks
	// - Handle incoming transactions
	// - Handle mining pool messages
}

func submitBlock(block *block.Block) {
	// TODO: Implement block submission
	// - For solo mining: broadcast to network
	// - For pool mining: submit to pool
}

func reportStats() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		statsMutex.RLock()
		uptime := time.Since(startTime).Seconds()
		log.Printf("Hashrate: %.2f H/s, Blocks found: %d, Uptime: %.0f seconds",
			hashrate, blocksFound, uptime)
		statsMutex.RUnlock()
	}
}
