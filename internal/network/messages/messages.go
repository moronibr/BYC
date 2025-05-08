package messages

import (
	"encoding/json"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/coin"
	"github.com/youngchain/internal/core/types"
)

// Type represents the type of message
type Type string

const (
	BlockMsg       Type = "block"
	TransactionMsg Type = "transaction"
)

// Message represents a network message
type Message struct {
	Type Type            `json:"type"`
	Data json.RawMessage `json:"data"`
}

// BlockMessage represents a block message
type BlockMessage struct {
	Block     *block.Block    `json:"block"`
	BlockType block.BlockType `json:"block_type"`
}

// TransactionMessage represents a transaction message
type TransactionMessage struct {
	Transaction *types.Transaction `json:"transaction"`
	CoinType    coin.CoinType      `json:"coin_type"`
}

// NewMessage creates a new message
func NewMessage(msgType Type, data interface{}) (*Message, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &Message{
		Type: msgType,
		Data: jsonData,
	}, nil
}
