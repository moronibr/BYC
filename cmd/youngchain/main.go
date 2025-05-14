package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/mining"
	"github.com/youngchain/internal/core/network"
	corestorage "github.com/youngchain/internal/core/storage"
	"github.com/youngchain/internal/core/transaction"
	"github.com/youngchain/internal/core/utxo"
	"github.com/youngchain/internal/logger"
	"github.com/youngchain/internal/storage"
)

var (
	// Mining statistics
	startTime     time.Time
	blocksFound   uint64
	lastHashCount uint64
	hashCount     uint64
	hashMutex     sync.Mutex
)

func main() {
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
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Initialize database
	db, err := storage.NewDB("byc.db")
	if err != nil {
		logger.Fatal("Failed to initialize database", logger.Error2(err))
	}
	defer db.Close()

	// Create genesis block
	genesisBlock := block.NewBlock(nil, 0)
	genesisBlock.Header.Version = 1
	genesisBlock.Header.Timestamp = time.Now()
	genesisBlock.Header.Difficulty = 1
	genesisBlock.Header.Hash = genesisBlock.CalculateHash()

	if err := db.SaveBlock(genesisBlock); err != nil {
		logger.Fatal("Failed to save genesis block", logger.Error2(err))
	}

	// Initialize chain state
	if err := db.SaveChainState(0, genesisBlock.Header.Hash); err != nil {
		logger.Fatal("Failed to initialize chain state", logger.Error2(err))
	}

	// Initialize P2P node
	port := 8333
	config := &network.NodeConfig{
		ListenAddr: "0.0.0.0",
		Port:       port,
		Bootstrap:  []string{"127.0.0.1:8334", "127.0.0.1:8335"}, // Default bootstrap nodes
	}
	node, err := network.NewNode(config)
	if err != nil {
		logger.Fatal("Failed to initialize P2P node", logger.Error2(err))
	}
	defer node.Close()

	// Initialize UTXO set
	utxoSet := utxo.NewUTXOSet()
	utxoAdapter := corestorage.NewUTXOAdapter(utxoSet)

	// Initialize transaction pool
	txPool := transaction.NewTxPool(1000, 1000, utxoAdapter)

	// Initialize miner
	miner := mining.NewMiner(txPool, corestorage.NewBlockStoreAdapter(db), utxoSet, "miner_address")

	// Handle user choice
	fmt.Println("Choose operation mode:")
	fmt.Println("1. Run as node")
	fmt.Println("2. Run as miner")
	fmt.Print("Enter choice (1 or 2): ")

	var choice int
	fmt.Scan(&choice)

	switch choice {
	case 1:
		// Run as node
		logger.Info("Running as node...")
		// Keep the node running
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		logger.Info("Shutting down node...")

	case 2:
		// Run as miner
		logger.Info("Running as miner...")
		startTime = time.Now()

		// Start mining
		miner.StartMining()
		defer miner.StopMining()
		go reportStats()

		// Keep the miner running
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		logger.Info("Shutting down miner...")

	default:
		logger.Fatal("Invalid choice")
	}
}

func reportStats() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		hashMutex.Lock()
		currentHashCount := hashCount
		hashMutex.Unlock()

		hashesPerSecond := float64(currentHashCount-lastHashCount) / 10.0
		lastHashCount = currentHashCount

		uptime := time.Since(startTime).Round(time.Second)
		totalHashes := currentHashCount

		logger.Info("Mining Statistics",
			logger.String("status", "active"),
			logger.Float64("hash_rate", hashesPerSecond),
			logger.String("hash_rate_unit", "H/s"),
			logger.Int64("blocks_mined", int64(blocksFound)),
			logger.Int64("total_hashes", int64(totalHashes)),
			logger.Duration("uptime", uptime),
			logger.String("mining_address", "miner_address"))
	}
}
