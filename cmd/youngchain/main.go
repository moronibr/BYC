package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/youngchain/internal/config"
	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/coin"
	"github.com/youngchain/internal/core/mining"
	"github.com/youngchain/internal/network"
	"github.com/youngchain/internal/storage"
)

func main() {
	// Parse command line flags
	nodeType := flag.String("type", "full", "Node type (full, miner, light)")
	port := flag.Int("port", 8333, "Port to listen on")
	flag.Parse()

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

	// Initialize supply tracker (will be used in future implementation)
	_ = coin.NewSupplyTracker()

	// Create network server
	cfg := &config.Config{
		ListenAddr: fmt.Sprintf(":%d", *port),
		MaxPeers:   10,
	}
	server := network.NewServer(cfg)
	server.SetDB(db)

	// Start network server
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start network server: %v", err)
	}
	defer server.Stop()

	// Create miner if this is a mining node
	var miner *mining.Miner
	if *nodeType == "miner" {
		miner = mining.NewMiner(goldenGenesis)
		miner.StartMining(coin.Leah)
		defer miner.StopMining()
	}

	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the node
	log.Printf("Starting Brigham Young Chain node (%s) on port %d", *nodeType, *port)

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
			if *nodeType == "miner" && miner != nil {
				// Check if the miner has found a block
				if miner.CurrentBlock.Hash != nil && len(miner.CurrentBlock.Hash) > 0 {
					// Save the mined block to the database
					if err := db.SaveBlock(miner.CurrentBlock); err != nil {
						log.Printf("Error saving mined block: %v", err)
					} else {
						log.Printf("Mined new block with hash: %x", miner.CurrentBlock.Hash)

						// Create a new block for mining
						newBlock := block.NewBlock(miner.CurrentBlock.Hash, miner.CurrentBlock.Header.Difficulty)
						miner.CurrentBlock = newBlock
						miner.StartMining(coin.Leah)
					}
				}
			}

			// Sleep briefly to avoid CPU spinning
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// createGenesisBlock creates the genesis block for either chain
func createGenesisBlock(blockType block.BlockType) *block.Block {
	// Create genesis block with empty previous hash and initial difficulty
	genesis := block.NewBlock([]byte{}, block.GetInitialDifficulty(blockType))

	// Add genesis message
	var message string
	if blockType == block.GoldenBlock {
		message = "In the beginning of the Golden Chain..."
	} else {
		message = "In the beginning of the Silver Chain..."
	}

	// Create genesis transaction
	tx := &block.Transaction{
		Version: 1,
		Inputs:  []block.TxInput{},
		Outputs: []block.TxOutput{
			{
				Value:    50, // Initial mining reward
				Script:   []byte(message),
				CoinType: coin.Leah, // Start with base unit
			},
		},
		LockTime: 0,
		CoinType: coin.Leah,
	}

	genesis.AddTransaction(tx)

	// Calculate merkle root
	genesis.UpdateMerkleRoot()

	// Calculate hash for the genesis block
	genesis.Hash = genesis.CalculateHash()

	return genesis
}
