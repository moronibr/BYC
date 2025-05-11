package types

import (
	"encoding/json"
)

// MessageType represents the type of a network message
type MessageType string

const (
	// Version message type
	VersionMsg MessageType = "version"
	// VerAck message type
	VerAckMsg MessageType = "verack"
	// GetAddr message type
	GetAddrMsg MessageType = "getaddr"
	// Addr message type
	AddrMsg MessageType = "addr"
	// Block message type
	BlockMsg MessageType = "block"
	// Tx message type
	TxMsg MessageType = "tx"
	// Ping message type
	PingMsg MessageType = "ping"
	// Pong message type
	PongMsg MessageType = "pong"
)

// Message represents a network message
type Message struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload"`
}

// Encode encodes a message to bytes
func (m *Message) Encode() ([]byte, error) {
	return json.Marshal(m)
}

// Decode decodes a message from bytes
func (m *Message) Decode(data []byte) error {
	return json.Unmarshal(data, m)
}

// VersionMessage represents a version message
type VersionMessage struct {
	Version     uint32 `json:"version"`
	Services    uint64 `json:"services"`
	Timestamp   int64  `json:"timestamp"`
	AddrRecv    string `json:"addr_recv"`
	AddrFrom    string `json:"addr_from"`
	Nonce       uint64 `json:"nonce"`
	UserAgent   string `json:"user_agent"`
	StartHeight int32  `json:"start_height"`
	Relay       bool   `json:"relay"`
}

// AddrMessage represents an addr message
type AddrMessage struct {
	Addresses []string `json:"addresses"`
}

// BlockMessage represents a block message
type BlockMessage struct {
	Block *Block `json:"block"`
}

// TxMessage represents a transaction message
type TxMessage struct {
	Tx *Transaction `json:"tx"`
}

// PingMessage represents a ping message
type PingMessage struct {
	Nonce uint64 `json:"nonce"`
}

// PongMessage represents a pong message
type PongMessage struct {
	Nonce uint64 `json:"nonce"`
}
