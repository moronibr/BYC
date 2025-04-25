package network

import (
	"encoding/binary"
	"fmt"
)

// SerializeBinary serializes a message to binary format
func (m *Message) SerializeBinary() ([]byte, error) {
	// Serialize command
	commandBytes := []byte(m.Command)
	commandLength := len(commandBytes)

	// Create buffer for magic (4 bytes) + command length (4 bytes) + command + timestamp (8 bytes) + payload length (4 bytes) + payload
	buffer := make([]byte, 4+4+commandLength+8+4+len(m.Payload))

	// Write magic (4 bytes)
	binary.LittleEndian.PutUint32(buffer[0:4], m.Magic)

	// Write command length (4 bytes)
	binary.LittleEndian.PutUint32(buffer[4:8], uint32(commandLength))

	// Write command
	copy(buffer[8:8+commandLength], commandBytes)

	// Write timestamp (8 bytes)
	binary.LittleEndian.PutUint64(buffer[8+commandLength:16+commandLength], uint64(m.Timestamp))

	// Write payload length (4 bytes)
	binary.LittleEndian.PutUint32(buffer[16+commandLength:20+commandLength], uint32(len(m.Payload)))

	// Write payload
	copy(buffer[20+commandLength:], m.Payload)

	return buffer, nil
}

// DeserializeBinary deserializes a message from binary format
func DeserializeBinary(data []byte) (*Message, error) {
	if len(data) < 20 { // Minimum length for magic (4) + command length (4) + timestamp (8) + payload length (4)
		return nil, fmt.Errorf("message too short")
	}

	// Read magic
	magic := binary.LittleEndian.Uint32(data[0:4])

	// Read command length
	commandLength := binary.LittleEndian.Uint32(data[4:8])

	if len(data) < 8+int(commandLength) {
		return nil, fmt.Errorf("message too short for command")
	}

	// Read command
	command := MessageType(data[8 : 8+int(commandLength)])

	// Read timestamp
	timestamp := int64(binary.LittleEndian.Uint64(data[8+int(commandLength) : 16+int(commandLength)]))

	// Read payload length
	payloadLength := binary.LittleEndian.Uint32(data[16+int(commandLength) : 20+int(commandLength)])

	if len(data) < 20+int(commandLength)+int(payloadLength) {
		return nil, fmt.Errorf("message too short for payload")
	}

	// Read payload
	payload := data[20+int(commandLength) : 20+int(commandLength)+int(payloadLength)]

	return &Message{
		Magic:     magic,
		Command:   command,
		Payload:   payload,
		Timestamp: timestamp,
	}, nil
}
