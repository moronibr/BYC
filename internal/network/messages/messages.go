package messages

import (
	"encoding/json"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/types"
)

// MessageType represents the type of network message
type MessageType string

const (
	// Message types
	PingMsg        MessageType = "ping"
	PongMsg        MessageType = "pong"
	BlockMsg       MessageType = "block"
	TransactionMsg MessageType = "transaction"
)

// Message represents a network message
type Message struct {
	Type MessageType     `json:"type"`
	Data json.RawMessage `json:"data"`
}

// BlockMessage represents a block message
type BlockMessage struct {
	Block     *block.Block `json:"block"`
	BlockType string       `json:"block_type"`
}

// TransactionMessage represents a transaction message
type TransactionMessage struct {
	Transaction *types.Transaction `json:"transaction"`
	CoinType    string             `json:"coin_type"`
}

// NewMessage creates a new message
func NewMessage(msgType MessageType, data interface{}) (*Message, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &Message{
		Type: msgType,
		Data: jsonData,
	}, nil
}
