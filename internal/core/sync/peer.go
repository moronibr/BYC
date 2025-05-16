package sync

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/logger"
)

const (
	// Message types
	MsgVersion   = 0x01
	MsgVerAck    = 0x02
	MsgGetBlocks = 0x03
	MsgBlock     = 0x04
	MsgInv       = 0x05
	MsgGetData   = 0x06
	MsgNotFound  = 0x07
	MsgPing      = 0x08
	MsgPong      = 0x09
	MsgReject    = 0x0A

	// Protocol version
	ProtocolVersion = 70001

	// Network magic bytes
	NetworkMagic = 0xD9B4BEF9

	// Default port
	DefaultPort = 8333

	// Connection timeout
	ConnTimeout = 30 * time.Second

	// Keep-alive interval
	KeepAliveInterval = 30 * time.Second

	// Maximum message size
	MaxMessageSize = 4 * 1024 * 1024 // 4MB
)

// Peer represents a network peer
type Peer struct {
	mu sync.RWMutex

	// Connection
	conn net.Conn

	// Address
	address string

	// Version
	version uint32

	// Services
	services uint64

	// Last seen
	lastSeen time.Time

	// Send queue
	sendQueue chan []byte

	// Stop channel
	stopChan chan struct{}

	// Block sync
	blockSync *BlockSync
}

// NewPeer creates a new peer
func NewPeer(conn net.Conn, address string, blockSync *BlockSync) *Peer {
	return &Peer{
		conn:      conn,
		address:   address,
		sendQueue: make(chan []byte, 1000),
		stopChan:  make(chan struct{}),
		blockSync: blockSync,
		lastSeen:  time.Now(),
	}
}

// Start starts the peer
func (p *Peer) Start() {
	go p.readLoop()
	go p.writeLoop()
	go p.keepAlive()
}

// Stop stops the peer
func (p *Peer) Stop() {
	close(p.stopChan)
	p.conn.Close()
}

// readLoop reads messages from the peer
func (p *Peer) readLoop() {
	for {
		select {
		case <-p.stopChan:
			return
		default:
			// Read message header
			header := make([]byte, 24)
			if _, err := p.conn.Read(header); err != nil {
				logger.Error("Failed to read message header", logger.Error(err))
				return
			}

			// Parse message header
			magic := binary.LittleEndian.Uint32(header[0:4])
			if magic != NetworkMagic {
				logger.Error("Invalid network magic")
				return
			}

			command := string(bytes.TrimRight(header[4:16], "\x00"))
			length := binary.LittleEndian.Uint32(header[16:20])
			checksum := binary.LittleEndian.Uint32(header[20:24])

			// Check message size
			if length > MaxMessageSize {
				logger.Error("Message too large")
				return
			}

			// Read message payload
			payload := make([]byte, length)
			if _, err := p.conn.Read(payload); err != nil {
				logger.Error("Failed to read message payload", logger.Error(err))
				return
			}

			// Verify checksum
			if !verifyChecksum(payload, checksum) {
				logger.Error("Invalid checksum")
				return
			}

			// Handle message
			if err := p.handleMessage(command, payload); err != nil {
				logger.Error("Failed to handle message", logger.Error(err))
				continue
			}

			// Update last seen
			p.mu.Lock()
			p.lastSeen = time.Now()
			p.mu.Unlock()
		}
	}
}

// writeLoop writes messages to the peer
func (p *Peer) writeLoop() {
	for {
		select {
		case <-p.stopChan:
			return
		case msg := <-p.sendQueue:
			if _, err := p.conn.Write(msg); err != nil {
				logger.Error("Failed to write message", logger.Error(err))
				return
			}
		}
	}
}

// keepAlive sends periodic pings to keep the connection alive
func (p *Peer) keepAlive() {
	ticker := time.NewTicker(KeepAliveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopChan:
			return
		case <-ticker.C:
			p.SendPing()
		}
	}
}

// handleMessage handles a received message
func (p *Peer) handleMessage(command string, payload []byte) error {
	switch command {
	case "version":
		return p.handleVersion(payload)
	case "verack":
		return p.handleVerAck(payload)
	case "getblocks":
		return p.handleGetBlocks(payload)
	case "block":
		return p.handleBlock(payload)
	case "inv":
		return p.handleInv(payload)
	case "getdata":
		return p.handleGetData(payload)
	case "notfound":
		return p.handleNotFound(payload)
	case "ping":
		return p.handlePing(payload)
	case "pong":
		return p.handlePong(payload)
	case "reject":
		return p.handleReject(payload)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

// SendMessage sends a message to the peer
func (p *Peer) SendMessage(command string, payload []byte) error {
	// Create message header
	header := make([]byte, 24)
	binary.LittleEndian.PutUint32(header[0:4], NetworkMagic)
	copy(header[4:16], command)
	binary.LittleEndian.PutUint32(header[16:20], uint32(len(payload)))
	binary.LittleEndian.PutUint32(header[20:24], calculateChecksum(payload))

	// Combine header and payload
	msg := append(header, payload...)

	// Send message
	select {
	case p.sendQueue <- msg:
		return nil
	default:
		return fmt.Errorf("send queue is full")
	}
}

// SendVersion sends a version message
func (p *Peer) SendVersion() error {
	// Create version message
	version := make([]byte, 86)
	binary.LittleEndian.PutUint32(version[0:4], ProtocolVersion)
	binary.LittleEndian.PutUint64(version[4:12], 1) // Services
	binary.LittleEndian.PutInt64(version[12:20], time.Now().Unix())
	binary.LittleEndian.PutUint64(version[20:28], 0) // AddrRecv services
	copy(version[28:48], make([]byte, 16))           // AddrRecv IP
	binary.LittleEndian.PutUint16(version[48:50], DefaultPort)
	binary.LittleEndian.PutUint64(version[50:58], 0) // AddrFrom services
	copy(version[58:78], make([]byte, 16))           // AddrFrom IP
	binary.LittleEndian.PutUint16(version[78:80], DefaultPort)
	binary.LittleEndian.PutUint64(version[80:88], 0) // Nonce

	return p.SendMessage("version", version)
}

// SendVerAck sends a verack message
func (p *Peer) SendVerAck() error {
	return p.SendMessage("verack", nil)
}

// SendGetBlocks sends a getblocks message
func (p *Peer) SendGetBlocks(startHash []byte, stopHash []byte) error {
	// Create getblocks message
	msg := make([]byte, 37)
	binary.LittleEndian.PutUint32(msg[0:4], ProtocolVersion)
	binary.LittleEndian.PutUint8(msg[4:5], 1) // Hash count
	copy(msg[5:37], startHash)                // Start hash
	copy(msg[37:69], stopHash)                // Stop hash

	return p.SendMessage("getblocks", msg)
}

// SendBlock sends a block message
func (p *Peer) SendBlock(block *block.Block) error {
	// Serialize block
	data, err := block.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize block: %v", err)
	}

	return p.SendMessage("block", data)
}

// SendInv sends an inv message
func (p *Peer) SendInv(inventory []byte) error {
	return p.SendMessage("inv", inventory)
}

// SendGetData sends a getdata message
func (p *Peer) SendGetData(inventory []byte) error {
	return p.SendMessage("getdata", inventory)
}

// SendNotFound sends a notfound message
func (p *Peer) SendNotFound(inventory []byte) error {
	return p.SendMessage("notfound", inventory)
}

// SendPing sends a ping message
func (p *Peer) SendPing() error {
	nonce := make([]byte, 8)
	binary.LittleEndian.PutUint64(nonce, uint64(time.Now().UnixNano()))
	return p.SendMessage("ping", nonce)
}

// SendPong sends a pong message
func (p *Peer) SendPong(nonce []byte) error {
	return p.SendMessage("pong", nonce)
}

// SendReject sends a reject message
func (p *Peer) SendReject(command string, code uint8, reason string) error {
	// Create reject message
	msg := make([]byte, 1+len(command)+1+len(reason))
	msg[0] = code
	copy(msg[1:], command)
	copy(msg[1+len(command):], reason)

	return p.SendMessage("reject", msg)
}

// handleVersion handles a version message
func (p *Peer) handleVersion(payload []byte) error {
	// Parse version message
	version := binary.LittleEndian.Uint32(payload[0:4])
	services := binary.LittleEndian.Uint64(payload[4:12])
	timestamp := binary.LittleEndian.Int64(payload[12:20])

	// Update peer info
	p.mu.Lock()
	p.version = version
	p.services = services
	p.lastSeen = time.Unix(timestamp, 0)
	p.mu.Unlock()

	// Send verack
	return p.SendVerAck()
}

// handleVerAck handles a verack message
func (p *Peer) handleVerAck(payload []byte) error {
	// Update last seen
	p.mu.Lock()
	p.lastSeen = time.Now()
	p.mu.Unlock()

	return nil
}

// handleGetBlocks handles a getblocks message
func (p *Peer) handleGetBlocks(payload []byte) error {
	// Parse getblocks message
	version := binary.LittleEndian.Uint32(payload[0:4])
	hashCount := binary.LittleEndian.Uint8(payload[4:5])
	startHash := payload[5:37]
	stopHash := payload[37:69]

	// Check version
	if version < ProtocolVersion {
		return p.SendReject("getblocks", 0x01, "version too old")
	}

	// Get blocks
	blocks, err := p.blockSync.GetBlocks(startHash, stopHash, int(hashCount))
	if err != nil {
		return fmt.Errorf("failed to get blocks: %v", err)
	}

	// Send blocks
	for _, block := range blocks {
		if err := p.SendBlock(block); err != nil {
			return fmt.Errorf("failed to send block: %v", err)
		}
	}

	return nil
}

// handleBlock handles a block message
func (p *Peer) handleBlock(payload []byte) error {
	// Deserialize block
	block, err := block.Deserialize(payload)
	if err != nil {
		return fmt.Errorf("failed to deserialize block: %v", err)
	}

	// Handle block
	return p.blockSync.HandleBlock(block)
}

// handleInv handles an inv message
func (p *Peer) handleInv(payload []byte) error {
	// Parse inv message
	count := binary.LittleEndian.Uint32(payload[0:4])
	offset := 4

	// Process inventory items
	for i := uint32(0); i < count; i++ {
		// Get inventory type
		invType := binary.LittleEndian.Uint32(payload[offset : offset+4])
		offset += 4

		// Get inventory hash
		invHash := payload[offset : offset+32]
		offset += 32

		// Handle inventory item
		switch invType {
		case 2: // Block
			if err := p.SendGetData(invHash); err != nil {
				return fmt.Errorf("failed to send getdata: %v", err)
			}
		}
	}

	return nil
}

// handleGetData handles a getdata message
func (p *Peer) handleGetData(payload []byte) error {
	// Parse getdata message
	count := binary.LittleEndian.Uint32(payload[0:4])
	offset := 4

	// Process inventory items
	for i := uint32(0); i < count; i++ {
		// Get inventory type
		invType := binary.LittleEndian.Uint32(payload[offset : offset+4])
		offset += 4

		// Get inventory hash
		invHash := payload[offset : offset+32]
		offset += 32

		// Handle inventory item
		switch invType {
		case 2: // Block
			block, err := p.blockSync.GetBlock(invHash)
			if err != nil {
				return p.SendNotFound(invHash)
			}
			if err := p.SendBlock(block); err != nil {
				return fmt.Errorf("failed to send block: %v", err)
			}
		}
	}

	return nil
}

// handleNotFound handles a notfound message
func (p *Peer) handleNotFound(payload []byte) error {
	// Parse notfound message
	count := binary.LittleEndian.Uint32(payload[0:4])
	offset := 4

	// Process inventory items
	for i := uint32(0); i < count; i++ {
		// Get inventory type
		invType := binary.LittleEndian.Uint32(payload[offset : offset+4])
		offset += 4

		// Get inventory hash
		invHash := payload[offset : offset+32]
		offset += 32

		// Handle inventory item
		switch invType {
		case 2: // Block
			logger.Info("Block not found", logger.String("hash", fmt.Sprintf("%x", invHash)))
		}
	}

	return nil
}

// handlePing handles a ping message
func (p *Peer) handlePing(payload []byte) error {
	// Send pong
	return p.SendPong(payload)
}

// handlePong handles a pong message
func (p *Peer) handlePong(payload []byte) error {
	// Update last seen
	p.mu.Lock()
	p.lastSeen = time.Now()
	p.mu.Unlock()

	return nil
}

// handleReject handles a reject message
func (p *Peer) handleReject(payload []byte) error {
	// Parse reject message
	code := payload[0]
	command := string(payload[1 : len(payload)-1])
	reason := string(payload[len(payload)-1:])

	logger.Info("Message rejected",
		logger.Uint8("code", code),
		logger.String("command", command),
		logger.String("reason", reason))

	return nil
}

// verifyChecksum verifies a message checksum
func verifyChecksum(payload []byte, checksum uint32) bool {
	return calculateChecksum(payload) == checksum
}

// calculateChecksum calculates a message checksum
func calculateChecksum(payload []byte) uint32 {
	hash := sha256.Sum256(payload)
	hash = sha256.Sum256(hash[:])
	return binary.LittleEndian.Uint32(hash[:4])
}
