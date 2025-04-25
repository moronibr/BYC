package network

import (
	"encoding/json"
	"time"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/coin"
)

// MessageType represents the type of network message
type MessageType string

const (
	// Chain identity
	ChainName    = "Brigham Young Chain"
	ChainVersion = "1.0.0"
	NetworkMagic = 0x42594348 // "BYCH" in hex

	// Message types
	VersionMsg   MessageType = "version"
	VerAckMsg    MessageType = "verack"
	BlockMsg     MessageType = "block"
	TxMsg        MessageType = "tx"
	GetBlocksMsg MessageType = "getblocks"
	GetDataMsg   MessageType = "getdata"
	InventoryMsg MessageType = "inventory"
)

// Message represents a network message
type Message struct {
	Magic     uint32      `json:"magic"`
	Command   MessageType `json:"command"`
	Payload   []byte      `json:"payload"`
	Timestamp int64       `json:"timestamp"`
}

// VersionMessage represents a version message
type VersionMessage struct {
	Version     uint32 `json:"version"`
	Services    uint64 `json:"services"`
	Timestamp   int64  `json:"timestamp"`
	AddrRecv    string `json:"addr_recv"`
	AddrTrans   string `json:"addr_trans"`
	Nonce       uint64 `json:"nonce"`
	UserAgent   string `json:"user_agent"`
	StartHeight int32  `json:"start_height"`
	Relay       bool   `json:"relay"`
}

// BlockMessage represents a block message
type BlockMessage struct {
	Block     *block.Block    `json:"block"`
	BlockType block.BlockType `json:"block_type"`
}

// TransactionMessage represents a transaction message
type TransactionMessage struct {
	Transaction block.Transaction `json:"transaction"`
	CoinType    coin.CoinType     `json:"coin_type"`
}

// NewMessage creates a new network message
func NewMessage(command MessageType, payload []byte) *Message {
	return &Message{
		Magic:     NetworkMagic,
		Command:   command,
		Payload:   payload,
		Timestamp: time.Now().Unix(),
	}
}

// Serialize serializes the message to JSON
func (m *Message) Serialize() ([]byte, error) {
	return json.Marshal(m)
}

// Deserialize deserializes the message from JSON
func DeserializeMessage(data []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// Node represents a network node
type Node struct {
	Address   string
	Version   uint32
	Services  uint64
	LastSeen  time.Time
	Connected bool
}

// NewNode creates a new network node
func NewNode(address string) *Node {
	return &Node{
		Address:   address,
		Version:   1,
		Services:  1,
		LastSeen:  time.Now(),
		Connected: false,
	}
}

// UpdateLastSeen updates the last seen timestamp
func (n *Node) UpdateLastSeen() {
	n.LastSeen = time.Now()
}

// IsActive checks if the node is active
func (n *Node) IsActive() bool {
	return time.Since(n.LastSeen) < time.Minute*5
}
