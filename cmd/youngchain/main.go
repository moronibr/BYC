package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/youngchain/internal/config"
	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/coin"
	"github.com/youngchain/internal/core/common"
	"github.com/youngchain/internal/core/mining"
	"github.com/youngchain/internal/logger"
	"github.com/youngchain/internal/network"
	"github.com/youngchain/internal/storage"
)

// Mining statistics
var (
	hashrate    float64
	blocksFound int
	startTime   time.Time
	statsMutex  sync.RWMutex
)

func main() {
	// Display welcome message
	fmt.Println("Welcome to Brigham Young Chain!")
	fmt.Println("===============================")

	// Get user choice
	choice := getMenuChoice()

	// Initialize logger
	logConfig := logger.Config{
		Level:      "info",
		Filename:   "byc.log",
		MaxSize:    100, // 100MB
		MaxBackups: 3,
		MaxAge:     28, // 28 days
		Compress:   true,
	}
	if err := logger.Init(logConfig); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Initialize database
	db, err := storage.NewDB("byc.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create genesis blocks
	goldenGenesis := createGenesisBlock(block.GoldenBlock)
	silverGenesis := createGenesisBlock(block.SilverBlock)

	// Save genesis blocks to database
	if err := db.SaveBlock(goldenGenesis); err != nil {
		log.Fatalf("Failed to save golden genesis block: %v", err)
	}
	if err := db.SaveBlock(silverGenesis); err != nil {
		log.Fatalf("Failed to save silver genesis block: %v", err)
	}

	// Initialize supply tracker
	_ = coin.NewSupplyTracker()

	// Create network server
	cfg := &config.Config{
		ListenAddr: ":8333",
		MaxPeers:   10,
	}
	server := network.NewServer(cfg)
	server.SetDB(db)

	// Start network server
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start network server: %v", err)
	}
	defer server.Stop()

	var miner *mining.Miner
	var selectedCoinType coin.CoinType

	// Handle user choice
	switch choice {
	case 1: // Run Node
		fmt.Println("\nStarting node in full mode...")
		log.Printf("Starting Brigham Young Chain node (full) on port 8333")
	case 2: // Mine
		selectedCoinType = getCoinChoice()
		fmt.Printf("\nStarting mining node for %s coins...\n", selectedCoinType)

		// Initialize mining components
		miner = mining.NewMiner(goldenGenesis)
		miner.StartMining()
		defer miner.StopMining()
		log.Printf("Starting Brigham Young Chain node (miner) on port 8333")

		// Start mining statistics reporting
		startTime = time.Now()
		go reportMiningStats()
	default:
		fmt.Println("Invalid choice. Exiting...")
		return
	}

	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Main loop
	for {
		select {
		case sig := <-sigChan:
			log.Printf("Received signal %v, shutting down...", sig)
			return
		default:
			// Process any pending messages
			select {
			case msg := <-server.MessageChan:
				if err := server.HandleMessage(msg); err != nil {
					log.Printf("Error handling message: %v", err)
				}
			default:
				// No messages to process
			}

			// If this is a mining node, check if a block was mined
			if choice == 2 && miner != nil {
				// Check if the miner has found a block
				currentBlock := miner.GetCurrentBlock()
				if currentBlock != nil && currentBlock.Header.Hash != nil && len(currentBlock.Header.Hash) > 0 {
					// Save the mined block to the database
					if err := db.SaveBlock(currentBlock); err != nil {
						log.Printf("Error saving mined block: %v", err)
					} else {
						log.Printf("Mined new block with hash: %x", currentBlock.Header.Hash)

						// Update statistics
						statsMutex.Lock()
						blocksFound++
						statsMutex.Unlock()

						// Create a new block for mining
						newBlock := block.NewBlock(currentBlock.Header.Hash, uint64(currentBlock.Header.Difficulty))
						miner.SetCurrentBlock(newBlock)
						miner.StartMining()
					}
				}
			}

			// Sleep briefly to avoid CPU spinning
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func reportMiningStats() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	hashes := uint64(0)
	lastReport := time.Now()

	for {
		<-ticker.C
		statsMutex.Lock()
		uptime := time.Since(startTime).Seconds()
		hashrate = float64(hashes) / time.Since(lastReport).Seconds()
		hashes = 0
		lastReport = time.Now()
		statsMutex.Unlock()

		fmt.Printf("\rHashrate: %.2f H/s, Blocks found: %d, Uptime: %.0f seconds",
			hashrate, blocksFound, uptime)
	}
}

func getMenuChoice() int {
	for {
		fmt.Println("\nWhat would you like to do?")
		fmt.Println("1. Run Node")
		fmt.Println("2. Mine")
		fmt.Print("Enter your choice (1-2): ")

		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		choice, err := strconv.Atoi(input)
		if err == nil && (choice == 1 || choice == 2) {
			return choice
		}
		fmt.Println("Invalid choice. Please enter 1 or 2.")
	}
}

func getCoinChoice() coin.CoinType {
	for {
		fmt.Println("\nWhich coin would you like to mine?")
		fmt.Println("1. Leah (Easiest)")
		fmt.Println("2. Shiblum (Medium)")
		fmt.Println("3. Shiblon (Hardest)")
		fmt.Print("Enter your choice (1-3): ")

		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		choice, err := strconv.Atoi(input)
		if err == nil {
			switch choice {
			case 1:
				return coin.Leah
			case 2:
				return coin.Shiblum
			case 3:
				return coin.Shiblon
			}
		}
		fmt.Println("Invalid choice. Please enter 1, 2, or 3.")
	}
}

// createGenesisBlock creates the genesis block for either chain
func createGenesisBlock(blockType block.BlockType) *block.Block {
	// Create genesis block
	genesis := block.NewBlock(nil, 0)
	genesis.Header.Version = 1
	genesis.Header.Timestamp = time.Now()
	genesis.Header.Difficulty = block.GetInitialDifficulty(blockType)

	// Create genesis transaction
	tx := common.NewTransaction(
		nil,                              // From (empty for genesis)
		[]byte("genesis_reward_address"), // To
		1000000000,                       // Amount (1 billion)
		[]byte("Genesis transaction"),
	)

	// Add transaction to block
	if err := genesis.AddTransaction(tx); err != nil {
		log.Fatalf("Failed to add genesis transaction: %v", err)
	}

	// Calculate block hash
	genesis.Header.Hash = genesis.CalculateHash()

	return genesis
}
