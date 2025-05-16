package sync

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/youngchain/internal/core/block"
)

// Message types
const (
	MsgVersion    = 0x01
	MsgVerAck     = 0x02
	MsgGetBlocks  = 0x03
	MsgInv        = 0x04
	MsgGetData    = 0x05
	MsgBlock      = 0x06
	MsgTx         = 0x07
	MsgPing       = 0x08
	MsgPong       = 0x09
	MsgGetHeaders = 0x0A
	MsgHeaders    = 0x0B
)

// Message represents a network message
type Message struct {
	// Message type
	Type uint32

	// Payload
	Payload []byte
}

// NewMessage creates a new message
func NewMessage(msgType uint32, payload []byte) *Message {
	return &Message{
		Type:    msgType,
		Payload: payload,
	}
}

// Serialize serializes the message
func (m *Message) Serialize() ([]byte, error) {
	// Create buffer
	buf := new(bytes.Buffer)

	// Write magic bytes
	if err := binary.Write(buf, binary.LittleEndian, uint32(0xD9B4BEF9)); err != nil {
		return nil, fmt.Errorf("failed to write magic bytes: %v", err)
	}

	// Write command
	command := make([]byte, 12)
	copy(command, []byte(fmt.Sprintf("%d", m.Type)))
	if err := binary.Write(buf, binary.LittleEndian, command); err != nil {
		return nil, fmt.Errorf("failed to write command: %v", err)
	}

	// Write payload length
	if err := binary.Write(buf, binary.LittleEndian, uint32(len(m.Payload))); err != nil {
		return nil, fmt.Errorf("failed to write payload length: %v", err)
	}

	// Write checksum
	checksum := block.Sha256d(m.Payload)[:4]
	if err := binary.Write(buf, binary.LittleEndian, checksum); err != nil {
		return nil, fmt.Errorf("failed to write checksum: %v", err)
	}

	// Write payload
	if err := binary.Write(buf, binary.LittleEndian, m.Payload); err != nil {
		return nil, fmt.Errorf("failed to write payload: %v", err)
	}

	return buf.Bytes(), nil
}

// Deserialize deserializes a message
func Deserialize(r io.Reader) (*Message, error) {
	// Read magic bytes
	var magic uint32
	if err := binary.Read(r, binary.LittleEndian, &magic); err != nil {
		return nil, fmt.Errorf("failed to read magic bytes: %v", err)
	}

	// Check magic bytes
	if magic != 0xD9B4BEF9 {
		return nil, fmt.Errorf("invalid magic bytes: %x", magic)
	}

	// Read command
	command := make([]byte, 12)
	if _, err := io.ReadFull(r, command); err != nil {
		return nil, fmt.Errorf("failed to read command: %v", err)
	}

	// Parse command
	var msgType uint32
	if _, err := fmt.Sscanf(string(command), "%d", &msgType); err != nil {
		return nil, fmt.Errorf("failed to parse command: %v", err)
	}

	// Read payload length
	var payloadLen uint32
	if err := binary.Read(r, binary.LittleEndian, &payloadLen); err != nil {
		return nil, fmt.Errorf("failed to read payload length: %v", err)
	}

	// Read checksum
	var checksum [4]byte
	if _, err := io.ReadFull(r, checksum[:]); err != nil {
		return nil, fmt.Errorf("failed to read checksum: %v", err)
	}

	// Read payload
	payload := make([]byte, payloadLen)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, fmt.Errorf("failed to read payload: %v", err)
	}

	// Verify checksum
	if !bytes.Equal(block.Sha256d(payload)[:4], checksum[:]) {
		return nil, fmt.Errorf("invalid checksum")
	}

	return &Message{
		Type:    msgType,
		Payload: payload,
	}, nil
}

// HandleMessage handles a message
func (p *Peer) HandleMessage(msg *Message) error {
	switch msg.Type {
	case MsgVersion:
		return p.handleVersion(msg)
	case MsgVerAck:
		return p.handleVerack(msg)
	case MsgGetBlocks:
		return p.handleGetBlocks(msg)
	case MsgInv:
		return p.handleInv(msg)
	case MsgGetData:
		return p.handleGetData(msg)
	case MsgBlock:
		return p.handleBlock(msg)
	case MsgTx:
		return p.handleTx(msg)
	case MsgPing:
		return p.handlePing(msg)
	case MsgPong:
		return p.handlePong(msg)
	case MsgGetHeaders:
		return p.handleGetHeaders(msg)
	case MsgHeaders:
		return p.handleHeaders(msg)
	default:
		return fmt.Errorf("unknown message type: %d", msg.Type)
	}
}

// handleVersion handles a version message
func (p *Peer) handleVersion(msg *Message) error {
	// Parse version
	var version uint32
	if err := binary.Read(bytes.NewReader(msg.Payload), binary.LittleEndian, &version); err != nil {
		return fmt.Errorf("failed to parse version: %v", err)
	}

	// Send verack
	if err := p.SendVerAck(); err != nil {
		return fmt.Errorf("failed to send verack: %v", err)
	}

	return nil
}

// handleVerack handles a verack message
func (p *Peer) handleVerack(msg *Message) error {
	// Mark peer as connected
	p.connected = true

	return nil
}

// handleGetBlocks handles a getblocks message
func (p *Peer) handleGetBlocks(msg *Message) error {
	// Parse getblocks
	var startHeight uint64
	if err := binary.Read(bytes.NewReader(msg.Payload), binary.LittleEndian, &startHeight); err != nil {
		return fmt.Errorf("failed to parse getblocks: %v", err)
	}

	// Get blocks
	blocks := p.blockSync.GetBlocks(startHeight)
	if blocks == nil {
		return nil
	}

	// Send blocks
	for _, block := range blocks {
		if err := p.SendBlock(block); err != nil {
			return fmt.Errorf("failed to send block: %v", err)
		}
	}

	return nil
}

// handleInv handles an inventory message
func (p *Peer) handleInv(msg *Message) error {
	// Parse inventory
	var count uint32
	if err := binary.Read(bytes.NewReader(msg.Payload), binary.LittleEndian, &count); err != nil {
		return fmt.Errorf("failed to parse inventory count: %v", err)
	}

	// Process inventory
	for i := uint32(0); i < count; i++ {
		var invType uint32
		if err := binary.Read(bytes.NewReader(msg.Payload[4+i*36:]), binary.LittleEndian, &invType); err != nil {
			return fmt.Errorf("failed to parse inventory type: %v", err)
		}

		var hash [32]byte
		copy(hash[:], msg.Payload[8+i*36:])

		// Request data
		if err := p.SendGetData(invType, hash); err != nil {
			return fmt.Errorf("failed to send getdata: %v", err)
		}
	}

	return nil
}

// handleGetData handles a getdata message
func (p *Peer) handleGetData(msg *Message) error {
	// Parse getdata
	var count uint32
	if err := binary.Read(bytes.NewReader(msg.Payload), binary.LittleEndian, &count); err != nil {
		return fmt.Errorf("failed to parse getdata count: %v", err)
	}

	// Process getdata
	for i := uint32(0); i < count; i++ {
		var invType uint32
		if err := binary.Read(bytes.NewReader(msg.Payload[4+i*36:]), binary.LittleEndian, &invType); err != nil {
			return fmt.Errorf("failed to parse getdata type: %v", err)
		}

		var hash [32]byte
		copy(hash[:], msg.Payload[8+i*36:])

		// Send data
		switch invType {
		case 2: // Block
			block := p.blockSync.GetBlock(hash)
			if block != nil {
				if err := p.SendBlock(block); err != nil {
					return fmt.Errorf("failed to send block: %v", err)
				}
			}
		case 1: // Transaction
			tx := p.blockSync.GetTransaction(hash)
			if tx != nil {
				if err := p.SendTx(tx); err != nil {
					return fmt.Errorf("failed to send transaction: %v", err)
				}
			}
		}
	}

	return nil
}

// handleBlock handles a block message
func (p *Peer) handleBlock(msg *Message) error {
	// Parse block
	block, err := block.DeserializeBlock(msg.Payload)
	if err != nil {
		return fmt.Errorf("failed to parse block: %v", err)
	}

	// Handle block
	if err := p.blockSync.HandleBlock(block); err != nil {
		return fmt.Errorf("failed to handle block: %v", err)
	}

	return nil
}

// handleTx handles a transaction message
func (p *Peer) handleTx(msg *Message) error {
	// Parse transaction
	tx, err := block.DeserializeTransaction(msg.Payload)
	if err != nil {
		return fmt.Errorf("failed to parse transaction: %v", err)
	}

	// Handle transaction
	if err := p.blockSync.HandleTransaction(tx); err != nil {
		return fmt.Errorf("failed to handle transaction: %v", err)
	}

	return nil
}

// handlePing handles a ping message
func (p *Peer) handlePing(msg *Message) error {
	// Parse nonce
	var nonce uint64
	if err := binary.Read(bytes.NewReader(msg.Payload), binary.LittleEndian, &nonce); err != nil {
		return fmt.Errorf("failed to parse nonce: %v", err)
	}

	// Send pong
	if err := p.SendPong(nonce); err != nil {
		return fmt.Errorf("failed to send pong: %v", err)
	}

	return nil
}

// handlePong handles a pong message
func (p *Peer) handlePong(msg *Message) error {
	// Parse nonce
	var nonce uint64
	if err := binary.Read(bytes.NewReader(msg.Payload), binary.LittleEndian, &nonce); err != nil {
		return fmt.Errorf("failed to parse nonce: %v", err)
	}

	// Update last seen
	p.lastSeen = time.Now()

	return nil
}

// handleGetHeaders handles a getheaders message
func (p *Peer) handleGetHeaders(msg *Message) error {
	// Parse getheaders
	var startHeight uint64
	if err := binary.Read(bytes.NewReader(msg.Payload), binary.LittleEndian, &startHeight); err != nil {
		return fmt.Errorf("failed to parse getheaders: %v", err)
	}

	// Get headers
	headers := p.blockSync.GetHeaders(startHeight)
	if headers == nil {
		return nil
	}

	// Send headers
	if err := p.SendHeaders(headers); err != nil {
		return fmt.Errorf("failed to send headers: %v", err)
	}

	return nil
}

// handleHeaders handles a headers message
func (p *Peer) handleHeaders(msg *Message) error {
	// Parse headers
	var count uint32
	if err := binary.Read(bytes.NewReader(msg.Payload), binary.LittleEndian, &count); err != nil {
		return fmt.Errorf("failed to parse headers count: %v", err)
	}

	// Process headers
	for i := uint32(0); i < count; i++ {
		header, err := block.DeserializeBlockHeader(msg.Payload[4+i*80:])
		if err != nil {
			return fmt.Errorf("failed to parse header: %v", err)
		}

		// Handle header
		if err := p.blockSync.HandleHeader(header); err != nil {
			return fmt.Errorf("failed to handle header: %v", err)
		}
	}

	return nil
}
