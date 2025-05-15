package types

import (
	"encoding/json"
)

// MessageType represents the type of a network message
type MessageType uint8

const (
	MsgVersion   MessageType = 1
	MsgVerAck    MessageType = 2
	MsgPing      MessageType = 3
	MsgPong      MessageType = 4
	MsgGetBlocks MessageType = 5
	MsgBlock     MessageType = 6
	MsgTx        MessageType = 7
)

// Message represents a network message
type Message struct {
	Type    MessageType
	Payload interface{}
}

// Encode encodes a message to bytes
func (m *Message) Encode() ([]byte, error) {
	return json.Marshal(m)
}

// Decode decodes a message from bytes
func (m *Message) Decode(data []byte) error {
	return json.Unmarshal(data, m)
}

// VersionPayload represents the payload of a version message
type VersionPayload struct {
	Version   uint32
	Services  uint64
	Timestamp int64
	AddrRecv  string
	AddrFrom  string
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
