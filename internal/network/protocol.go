package network

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
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
	Type MessageType     `json:"type"`
	Data json.RawMessage `json:"data"`
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

// SerializeBinary serializes a message to binary format
func SerializeBinary(msg *Message) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %v", err)
	}

	// Create buffer for length (4 bytes) + data
	buffer := make([]byte, 4+len(data))

	// Write length (4 bytes)
	binary.LittleEndian.PutUint32(buffer[0:4], uint32(len(data)))

	// Write data
	copy(buffer[4:], data)

	return buffer, nil
}

// DeserializeBinary deserializes a message from binary format
func DeserializeBinary(data []byte) (*Message, error) {
	if len(data) < 4 { // Minimum length for data length (4)
		return nil, fmt.Errorf("message too short")
	}

	// Read length
	length := binary.LittleEndian.Uint32(data[0:4])

	if len(data) < 4+int(length) {
		return nil, fmt.Errorf("message too short for payload")
	}

	// Read data
	var msg Message
	if err := json.Unmarshal(data[4:4+length], &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %v", err)
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
