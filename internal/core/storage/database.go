package storage

import (
	"context"
	"time"
)

// Database defines the interface for blockchain data storage
type Database interface {
	// Block operations
	StoreBlock(ctx context.Context, block *Block) error
	GetBlock(ctx context.Context, hash [32]byte) (*Block, error)
	GetBlockByHeight(ctx context.Context, height uint64) (*Block, error)
	GetLatestBlock(ctx context.Context) (*Block, error)
	DeleteBlock(ctx context.Context, hash [32]byte) error

	// Transaction operations
	StoreTransaction(ctx context.Context, tx *Transaction) error
	GetTransaction(ctx context.Context, hash [32]byte) (*Transaction, error)
	DeleteTransaction(ctx context.Context, hash [32]byte) error

	// UTXO operations
	StoreUTXO(ctx context.Context, txHash [32]byte, outputIndex uint32, utxo *DBUTXO) error
	GetUTXO(ctx context.Context, txHash [32]byte, outputIndex uint32) (*DBUTXO, error)
	GetAllUTXOs(ctx context.Context) ([]*DBUTXO, error)
	DeleteUTXO(ctx context.Context, txHash [32]byte, outputIndex uint32) error

	// Chain state operations
	StoreChainState(ctx context.Context, state *ChainState) error
	GetChainState(ctx context.Context) (*ChainState, error)

	// Batch operations
	BeginTx(ctx context.Context) (DBTx, error)
	Commit(ctx context.Context, tx DBTx) error
	Rollback(ctx context.Context, tx DBTx) error

	// Maintenance operations
	Close() error
	Compact() error
	Backup(ctx context.Context, path string) error
	Restore(ctx context.Context, path string) error
}

// DBTx represents a database transaction
type DBTx interface {
	// Block operations
	StoreBlock(block *Block) error
	GetBlock(hash [32]byte) (*Block, error)
	DeleteBlock(hash [32]byte) error

	// Transaction operations
	StoreTransaction(tx *Transaction) error
	GetTransaction(hash [32]byte) (*Transaction, error)
	DeleteTransaction(hash [32]byte) error

	// UTXO operations
	StoreUTXO(txHash [32]byte, outputIndex uint32, utxo *DBUTXO) error
	GetUTXO(txHash [32]byte, outputIndex uint32) (*DBUTXO, error)
	DeleteUTXO(txHash [32]byte, outputIndex uint32) error

	// Chain state operations
	StoreChainState(state *ChainState) error
	GetChainState() (*ChainState, error)
}

// ChainState represents the current state of the blockchain
type ChainState struct {
	CurrentHeight uint64    `json:"current_height"`
	CurrentHash   [32]byte  `json:"current_hash"`
	TotalWork     uint64    `json:"total_work"`
	LastUpdate    time.Time `json:"last_update"`
}

// Block represents a block in the blockchain
type Block struct {
	Version      uint32
	PrevHash     [32]byte
	MerkleRoot   [32]byte
	Timestamp    int64
	Difficulty   uint32
	Height       uint64
	Nonce        uint32
	Hash         [32]byte
	Transactions []*Transaction
}

// Transaction represents a transaction in the blockchain
type Transaction struct {
	Version   uint32
	Inputs    []*TxInput
	Outputs   []*TxOutput
	LockTime  uint32
	Hash      [32]byte
	Timestamp int64
}

// TxInput represents a transaction input
type TxInput struct {
	PrevHash  [32]byte
	PrevIndex uint32
	Script    []byte
	Sequence  uint32
}

// TxOutput represents a transaction output
type TxOutput struct {
	Value  uint64
	Script []byte
}

// DBUTXO represents an unspent transaction output in the database
type DBUTXO struct {
	TxHash      [32]byte `json:"tx_hash"`
	OutputIndex uint32   `json:"output_index"`
	Value       uint64   `json:"value"`
	Script      []byte   `json:"script"`
	Height      uint64   `json:"height"`
	IsCoinbase  bool     `json:"is_coinbase"`
}
