package types

import (
	"github.com/youngchain/internal/core/coin"
)

// If you need interfaces for UTXO operations, define them here. Do not define UTXO or UTXOSet structs here.

// BlockInterface defines the interface for block operations
// (Keep only if needed elsewhere in the codebase)
type BlockInterface interface {
	GetTransactions() []TransactionInterface
	GetHash() []byte
	GetHeight() uint64
}

type TransactionInterface interface {
	GetInputs() []InputInterface
	GetOutputs() []OutputInterface
	GetHash() []byte
	GetCoinType() coin.CoinType
	IsCoinbase() bool
	GetWitness() [][]byte
}

type InputInterface interface {
	GetPreviousTxHash() []byte
	GetPreviousTxIndex() uint32
}

type OutputInterface interface {
	GetValue() uint64
	GetScriptPubKey() []byte
}
