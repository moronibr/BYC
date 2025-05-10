package types

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"
)

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

// Message represents a network message
type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// UnmarshalData unmarshals the message data into the provided value
func (m *Message) UnmarshalData(v interface{}) error {
	return json.Unmarshal(m.Data, v)
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
