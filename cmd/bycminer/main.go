package main

import (
	"encoding/json"
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

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/coin"
	"github.com/youngchain/internal/core/mining"
	"github.com/youngchain/internal/network"
)

var (
	// Mining configuration
	miningType    = flag.String("type", "solo", "Mining type (solo, pool)")
	poolAddress   = flag.String("pool", "", "Pool address (required for pool mining)")
	walletAddress = flag.String("wallet", "", "Wallet address to receive mining rewards")
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
	transactions  []block.Transaction
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

	// Validate pool address for pool mining
	if *miningType == "pool" && *poolAddress == "" {
		log.Fatal("Pool address is required for pool mining")
	}

	// Validate wallet address
	if *walletAddress == "" {
		log.Fatal("Wallet address is required")
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
	networkServer = network.NewServer(host, port, 10)
	if err := networkServer.Start(); err != nil {
		log.Fatalf("Failed to start network server: %v", err)
	}
	defer networkServer.Stop()

	// Start network message handler
	go handleNetworkMessages()

	// Initialize mining with initial block
	latestBlock = &block.Block{
		Version:      1,
		PreviousHash: make([]byte, 32),
		Type:         block.GoldenBlock,
		Timestamp:    time.Now().Unix(),
		Difficulty:   calculateInitialDifficulty(),
	}
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
				transactions = []block.Transaction{}

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
				currentBlock.Nonce = nonce
				currentBlock.Hash = hash

				// Update statistics
				statsMutex.Lock()
				blocksFound++
				statsMutex.Unlock()

				// Submit block to network
				submitBlock(currentBlock)

				// Create new block for next mining round
				blockMutex.Lock()
				latestBlock = &block.Block{
					Version:      1,
					PreviousHash: currentBlock.Hash,
					Type:         block.GoldenBlock,
					Timestamp:    time.Now().Unix(),
					Difficulty:   calculateInitialDifficulty(),
				}
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

			// Increment nonce for next iteration
			nonce += uint64(*threads)
		}
	}
}

func calculateInitialDifficulty() uint64 {
	return 1 << 32 // Start with a relatively low difficulty
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

// handleNetworkMessages processes incoming network messages
func handleNetworkMessages() {
	for msg := range networkServer.MessageChan {
		switch msg.Command {
		case network.BlockMsg:
			// Process new block
			var blockMsg network.BlockMessage
			err := json.Unmarshal(msg.Payload, &blockMsg)
			if err != nil {
				log.Printf("Failed to unmarshal block message: %v", err)
				continue
			}

			// Update latest block
			blockMutex.Lock()
			latestBlock = blockMsg.Block
			blockMutex.Unlock()

			log.Printf("Received new block: %x", latestBlock.Hash)

		case network.TxMsg:
			// Process new transaction
			var txMsg network.TransactionMessage
			err := json.Unmarshal(msg.Payload, &txMsg)
			if err != nil {
				log.Printf("Failed to unmarshal transaction message: %v", err)
				continue
			}

			// Add transaction to pool
			txMutex.Lock()
			transactions = append(transactions, txMsg.Transaction)
			txMutex.Unlock()

			log.Printf("Received new transaction: %x", txMsg.Transaction.CalculateHash())
		}
	}
}

// submitBlock submits a mined block to the network
func submitBlock(block *block.Block) {
	// Create block message
	blockMsg := network.BlockMessage{
		Block:     block,
		BlockType: block.Type,
	}

	// Serialize message
	payload, err := json.Marshal(blockMsg)
	if err != nil {
		log.Printf("Failed to marshal block message: %v", err)
		return
	}

	// Create network message
	msg := network.NewMessage(network.BlockMsg, payload)

	// Broadcast message
	networkServer.BroadcastMessage(msg)

	log.Printf("Submitted block to network: %x", block.Hash)
}

func reportStats() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			statsMutex.RLock()
			duration := time.Since(startTime)
			fmt.Printf("\rHashrate: %.2f H/s | Blocks found: %d | Running time: %v",
				hashrate, blocksFound, duration.Round(time.Second))
			statsMutex.RUnlock()
		}
	}
}
