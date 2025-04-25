package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/coin"
	"github.com/youngchain/internal/core/mining"
	"github.com/youngchain/internal/network"
)

func main() {
	// Parse command line flags
	nodeType := flag.String("type", "full", "Node type (full, miner, light)")
	port := flag.Int("port", 8333, "Port to listen on")
	flag.Parse()

	// Create genesis blocks
	goldenGenesis := createGenesisBlock(block.GoldenBlock)
	_ = createGenesisBlock(block.SilverBlock) // Will be used in future implementation

	// Initialize supply tracker (will be used in future implementation)
	_ = coin.NewSupplyTracker()

	// Create miner if this is a mining node
	var miner *mining.Miner
	if *nodeType == "miner" {
		miner = mining.NewMiner(goldenGenesis)
	}

	// Create network node (will be used in future implementation)
	_ = network.NewNode("localhost:" + string(*port))

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
			if miner != nil {
				miner.StopMining()
			}
			return
		default:
			// TODO: Implement node operation
			// - Handle incoming connections
			// - Process blocks and transactions
			// - Mine blocks if this is a mining node
		}
	}
}

// createGenesisBlock creates the genesis block for either chain
func createGenesisBlock(blockType block.BlockType) *block.Block {
	genesis := block.NewBlock(blockType, []byte{})

	// Add genesis message
	var message string
	if blockType == block.GoldenBlock {
		message = "In the beginning of the Golden Chain..."
	} else {
		message = "In the beginning of the Silver Chain..."
	}

	// Create genesis transaction
	tx := block.Transaction{
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
	return genesis
}
