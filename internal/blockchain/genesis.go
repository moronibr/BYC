package blockchain

import (
	"encoding/hex"
	"time"
)

// GenesisBlock is the hardcoded first block of the BYC blockchain
var GenesisBlock = Block{
	Hash:         hexDecode("0000000000000000000000000000000000000000000000000000000000000000"), // Replace with your desired genesis block hash
	Timestamp:    time.Unix(1231006505, 0).Unix(),                                               // Convert to int64
	Transactions: []Transaction{
		// Add your genesis transaction(s) here
	},
	Nonce:    0, // Replace with your desired nonce
	PrevHash: hexDecode("0000000000000000000000000000000000000000000000000000000000000000"),
}

// GoldenGenesisBlock is the hardcoded first block of the Golden chain
var GoldenGenesisBlock = Block{
	Hash:         hexDecode("0000000000000000000000000000000000000000000000000000000000000000"),
	Timestamp:    time.Unix(1231006505, 0).Unix(),
	Transactions: []Transaction{
		// Add your genesis transaction(s) here
	},
	Nonce:     0,
	PrevHash:  hexDecode("0000000000000000000000000000000000000000000000000000000000000000"),
	BlockType: GoldenBlock,
}

// SilverGenesisBlock is the hardcoded first block of the Silver chain
var SilverGenesisBlock = Block{
	Hash:         hexDecode("0000000000000000000000000000000000000000000000000000000000000000"),
	Timestamp:    time.Unix(1231006505, 0).Unix(),
	Transactions: []Transaction{
		// Add your genesis transaction(s) here
	},
	Nonce:     0,
	PrevHash:  hexDecode("0000000000000000000000000000000000000000000000000000000000000000"),
	BlockType: SilverBlock,
}

// hexDecode converts a hex string to []byte
func hexDecode(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}
