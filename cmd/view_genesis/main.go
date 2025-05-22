package main

import (
	"github.com/moroni/BYC/internal/blockchain"
)

func main() {
	// Create new blockchain instance
	bc := blockchain.NewBlockchain()

	// Display Genesis block information
	bc.DisplayGenesisBlock()
}
