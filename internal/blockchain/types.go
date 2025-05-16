package blockchain

import (
	"time"
)

// CoinType represents the different types of coins in the system
type CoinType string

const (
	// Golden Block Coins
	Leah    CoinType = "LEAH"
	Shiblum CoinType = "SHIBLUM"
	Shiblon CoinType = "SHIBLON"
	Senine  CoinType = "SENINE"
	Seon    CoinType = "SEON"
	Shum    CoinType = "SHUM"
	Limnah  CoinType = "LIMNAH"
	Antion  CoinType = "ANTION"

	// Silver Block Coins
	Senum CoinType = "SENUM"
	Amnor CoinType = "AMNOR"
	Ezrom CoinType = "EZROM"
	Onti  CoinType = "ONTI"

	// Special Coins
	Ephraim  CoinType = "EPHRAIM"
	Manasseh CoinType = "MANASSEH"
	Joseph   CoinType = "JOSEPH"
)

// BlockType represents which blockchain a block belongs to
type BlockType string

const (
	GoldenBlock BlockType = "GOLDEN"
	SilverBlock BlockType = "SILVER"
)

// Block represents a block in the blockchain
type Block struct {
	Timestamp    int64
	Transactions []Transaction
	PrevHash     []byte
	Hash         []byte
	Nonce        int64
	BlockType    BlockType
	Difficulty   int
}

// Transaction represents a transaction in the blockchain
type Transaction struct {
	ID        []byte
	Inputs    []TxInput
	Outputs   []TxOutput
	Timestamp time.Time
	BlockType BlockType
}

// TxInput represents a transaction input
type TxInput struct {
	TxID      []byte
	Vout      int
	Signature []byte
	PubKey    []byte
}

// TxOutput represents a transaction output
type TxOutput struct {
	Value      float64
	CoinType   CoinType
	PubKeyHash []byte
}

// Wallet represents a user's wallet
type Wallet struct {
	PrivateKey []byte
	PublicKey  []byte
	Address    string
}

// Node represents a node in the P2P network
type Node struct {
	Address    string
	Peers      []string
	BlockType  BlockType
	IsMining   bool
	MiningCoin CoinType
}

// MiningDifficulty returns the difficulty multiplier for a given coin type
func MiningDifficulty(coinType CoinType) int {
	switch coinType {
	case Leah:
		return 1
	case Shiblum:
		return 2
	case Shiblon:
		return 4
	default:
		return 0 // Not mineable
	}
}

// IsMineable returns whether a coin type is mineable
func IsMineable(coinType CoinType) bool {
	return coinType == Leah || coinType == Shiblum || coinType == Shiblon
}

// CanTransferBetweenBlocks returns whether a coin can be transferred between blocks
func CanTransferBetweenBlocks(coinType CoinType) bool {
	return coinType == Antion
}

// GetBlockType returns the block type for a given coin
func GetBlockType(coinType CoinType) BlockType {
	switch coinType {
	case Senine, Seon, Shum, Limnah:
		return GoldenBlock
	case Senum, Amnor, Ezrom, Onti:
		return SilverBlock
	case Leah, Shiblum, Shiblon, Antion:
		return "" // These coins exist in both blocks
	case Ephraim:
		return GoldenBlock
	case Manasseh:
		return SilverBlock
	case Joseph:
		return "" // Special case, not tied to a specific block
	default:
		return ""
	}
}
