package main

import (
	"byc/internal/blockchain"
)

func main() {
	// Create new blockchain instance
	bc := blockchain.NewBlockchain()

	// Display Genesis block information
	bc.DisplayGenesisBlock()
}
