package main

import (
	"flag"
	"fmt"
	"log"
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
	"github.com/youngchain/internal/core/common"
	"github.com/youngchain/internal/core/mining"
	"github.com/youngchain/internal/core/storage"
	"github.com/youngchain/internal/core/transaction"
	"github.com/youngchain/internal/core/utxo"
	"github.com/youngchain/internal/network"
	"github.com/youngchain/internal/network/messages"
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
	minerPort   = flag.Int("port", 8334, "Port for the miner to listen on")

	// Statistics
	hashrate    float64
	blocksFound int
	startTime   time.Time
	statsMutex  sync.RWMutex

	// Network state
	networkServer *network.Server
	transactions  []*common.Transaction
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
	cfg := &config.Config{
		ListenAddr: fmt.Sprintf(":%d", *minerPort),
		MaxPeers:   10,
	}

	// Initialize network server
	networkServer = network.NewServer(cfg, nil)
	if err := networkServer.Start(); err != nil {
		log.Fatalf("Failed to start network server: %v", err)
	}
	defer networkServer.Stop()

	// Start network message handler
	go handleNetworkMessages()

	// Initialize storage
	blockStore := storage.NewBlockStore()
	utxoSet := utxo.NewUTXOSet()
	txPool := transaction.NewTxPool(1000, 1000, utxoSet)

	// Initialize mining with initial block
	latestBlock = block.NewBlock([]byte{}, uint64(block.GetInitialDifficulty(block.GoldenBlock)))
	miner := mining.NewMiner(txPool, blockStore, utxoSet, *walletAddress)

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

			if currentBlock == nil {
				// No block to mine yet, wait a bit
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// Create a copy of the block to avoid race conditions
			blockCopy := currentBlock.Copy()

			// Add pending transactions to block
			txMutex.Lock()
			if len(transactions) > 0 {
				// Add transactions to block
				for _, tx := range transactions {
					blockCopy.AddTransaction(tx)
				}

				// Update merkle root
				blockCopy.UpdateMerkleRoot()
			}
			txMutex.Unlock()

			// Start mining
			miner.StartMining()

			// Update statistics
			hashes++
			if time.Since(lastReport) >= time.Second {
				statsMutex.Lock()
				hashrate = float64(hashes) / time.Since(lastReport).Seconds()
				statsMutex.Unlock()
				hashes = 0
				lastReport = time.Now()
			}
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
	// Create block message
	blockMsg := &messages.BlockMessage{
		Block:     block,
		BlockType: "golden", // Use string literal for block type
	}

	// Create network message
	msg, err := messages.NewMessage(messages.BlockMsg, blockMsg)
	if err != nil {
		log.Printf("Error creating block message: %v", err)
		return
	}

	// Send message to network
	networkServer.MessageChan <- msg
	log.Printf("Submitted block with hash: %x", block.Header.Hash)
}

func reportStats() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		statsMutex.RLock()
		elapsed := time.Since(startTime).Seconds()
		log.Printf("Hashrate: %.2f H/s, Blocks found: %d, Uptime: %.0f seconds",
			hashrate, blocksFound, elapsed)
		statsMutex.RUnlock()
	}
}
