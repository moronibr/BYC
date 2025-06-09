package blockchain

import (
	"encoding/hex"
	"time"
)

// GenesisBlock is the hardcoded first block of the BYC blockchain
var GenesisBlock = Block{
	Hash:      hexDecode("0000000000000000000000000000000000000000000000000000000000000000"),
	Timestamp: time.Unix(1231006505, 0).Unix(),
	Transactions: []Transaction{
		{
			ID:        []byte("genesis"),
			Timestamp: time.Unix(1231006505, 0),
			Inputs:    []TxInput{},
			Outputs: []TxOutput{
				{
					Value:         1000000, // Initial supply
					CoinType:      Leah,
					PublicKeyHash: []byte("genesis"),
					Address:       "genesis",
				},
			},
			BlockType: GoldenBlock,
		},
	},
	Nonce:    0,
	PrevHash: hexDecode("0000000000000000000000000000000000000000000000000000000000000000"),
}

// GoldenGenesisBlock is the hardcoded first block of the Golden chain
var GoldenGenesisBlock = Block{
	Hash:      hexDecode("0000000000000000000000000000000000000000000000000000000000000000"),
	Timestamp: time.Unix(1231006505, 0).Unix(),
	Transactions: []Transaction{
		{
			ID:        []byte("golden_genesis"),
			Timestamp: time.Unix(1231006505, 0),
			Inputs:    []TxInput{},
			Outputs: []TxOutput{
				{
					Value:         1000000, // Initial supply
					CoinType:      Leah,
					PublicKeyHash: []byte("golden_genesis"),
					Address:       "golden_genesis",
				},
				{
					Value:         500000, // Initial supply
					CoinType:      Shiblum,
					PublicKeyHash: []byte("golden_genesis"),
					Address:       "golden_genesis",
				},
				{
					Value:         250000, // Initial supply
					CoinType:      Shiblon,
					PublicKeyHash: []byte("golden_genesis"),
					Address:       "golden_genesis",
				},
			},
			BlockType: GoldenBlock,
		},
	},
	Nonce:     0,
	PrevHash:  hexDecode("0000000000000000000000000000000000000000000000000000000000000000"),
	BlockType: GoldenBlock,
}

// SilverGenesisBlock is the hardcoded first block of the Silver chain
var SilverGenesisBlock = Block{
	Hash:      hexDecode("0000000000000000000000000000000000000000000000000000000000000000"),
	Timestamp: time.Unix(1231006505, 0).Unix(),
	Transactions: []Transaction{
		{
			ID:        []byte("silver_genesis"),
			Timestamp: time.Unix(1231006505, 0),
			Inputs:    []TxInput{},
			Outputs: []TxOutput{
				{
					Value:         1000000, // Initial supply
					CoinType:      Senum,
					PublicKeyHash: []byte("silver_genesis"),
					Address:       "silver_genesis",
				},
				{
					Value:         500000, // Initial supply
					CoinType:      Amnor,
					PublicKeyHash: []byte("silver_genesis"),
					Address:       "silver_genesis",
				},
				{
					Value:         250000, // Initial supply
					CoinType:      Ezrom,
					PublicKeyHash: []byte("silver_genesis"),
					Address:       "silver_genesis",
				},
			},
			BlockType: SilverBlock,
		},
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
