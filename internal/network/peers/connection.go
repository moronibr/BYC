package peers

import (
	"bufio"
	"fmt"
	"net"

	"github.com/youngchain/internal/core/block"
	coretypes "github.com/youngchain/internal/core/types"
	networktypes "github.com/youngchain/internal/network/types"
)

// Connection represents a peer connection
type Connection struct {
	conn      net.Conn
	reader    *bufio.Reader
	writer    *bufio.Writer
	peer      *networktypes.Node
	manager   *PeerManager
	handshake bool
}

// NewConnection creates a new peer connection
func NewConnection(conn net.Conn, peer *networktypes.Node, manager *PeerManager) *Connection {
	return &Connection{
		conn:      conn,
		reader:    bufio.NewReader(conn),
		writer:    bufio.NewWriter(conn),
		peer:      peer,
		manager:   manager,
		handshake: false,
	}
}

// Start starts the connection handler
func (c *Connection) Start() {
	go c.handleMessages()
}

// handleMessages handles incoming messages
func (c *Connection) handleMessages() {
	defer c.Close()

	for {
		// Read message length
		lengthBytes := make([]byte, 4)
		if _, err := c.reader.Read(lengthBytes); err != nil {
			return
		}
		length := uint32(lengthBytes[0]) | uint32(lengthBytes[1])<<8 | uint32(lengthBytes[2])<<16 | uint32(lengthBytes[3])<<24

		// Read message data
		data := make([]byte, length)
		if _, err := c.reader.Read(data); err != nil {
			return
		}

		// Parse message
		msg, err := networktypes.DeserializeBinary(append(lengthBytes, data...))
		if err != nil {
			continue
		}

		// Handle message
		if err := c.handleMessage(msg); err != nil {
			return
		}
	}
}

// handleMessage handles a single message
func (c *Connection) handleMessage(msg *networktypes.Message) error {
	switch msg.Type {
	case "version":
		return c.handleVersion(msg)
	case "verack":
		return c.handleVerAck(msg)
	case "block":
		return c.handleBlock(msg)
	case "tx":
		return c.handleTransaction(msg)
	case "getblocks":
		return c.handleGetBlocks(msg)
	case "getdata":
		return c.handleGetData(msg)
	case "inventory":
		return c.handleInventory(msg)
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

// handleVersion handles a version message
func (c *Connection) handleVersion(msg *networktypes.Message) error {
	var versionMsg struct {
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

	if err := msg.UnmarshalData(&versionMsg); err != nil {
		return err
	}

	// Update peer information
	c.peer.Version = versionMsg.Version
	c.peer.Services = versionMsg.Services
	c.peer.UpdateLastSeen()

	// Send verack
	verAckMsg := &networktypes.Message{
		Type: "verack",
		Data: nil,
	}

	return c.SendMessage(verAckMsg)
}

// handleVerAck handles a verack message
func (c *Connection) handleVerAck(msg *networktypes.Message) error {
	c.handshake = true
	return nil
}

// handleBlock handles a block message
func (c *Connection) handleBlock(msg *networktypes.Message) error {
	if !c.handshake {
		return fmt.Errorf("handshake not completed")
	}

	var blockMsg struct {
		Block *block.Block `json:"block"`
	}
	if err := msg.UnmarshalData(&blockMsg); err != nil {
		return fmt.Errorf("failed to unmarshal block message: %v", err)
	}

	// Process block
	if blockMsg.Block == nil {
		return fmt.Errorf("block is nil")
	}

	// TODO: Validate and process block
	return nil
}

// handleTransaction handles a transaction message
func (c *Connection) handleTransaction(msg *networktypes.Message) error {
	if !c.handshake {
		return fmt.Errorf("handshake not completed")
	}

	var txMsg struct {
		Transaction *coretypes.Transaction `json:"transaction"`
	}
	if err := msg.UnmarshalData(&txMsg); err != nil {
		return fmt.Errorf("failed to unmarshal transaction message: %v", err)
	}

	// Process transaction
	if txMsg.Transaction == nil {
		return fmt.Errorf("transaction is nil")
	}

	// TODO: Validate and process transaction
	return nil
}

// handleGetBlocks handles a getblocks message
func (c *Connection) handleGetBlocks(msg *networktypes.Message) error {
	if !c.handshake {
		return fmt.Errorf("handshake not completed")
	}

	var getBlocksMsg struct {
		Version     uint32   `json:"version"`
		BlockHashes [][]byte `json:"block_hashes"`
		StopHash    []byte   `json:"stop_hash"`
	}
	if err := msg.UnmarshalData(&getBlocksMsg); err != nil {
		return fmt.Errorf("failed to unmarshal getblocks message: %v", err)
	}

	// TODO: Implement getblocks handler
	return nil
}

// handleGetData handles a getdata message
func (c *Connection) handleGetData(msg *networktypes.Message) error {
	if !c.handshake {
		return fmt.Errorf("handshake not completed")
	}

	var getDataMsg struct {
		Inventory []struct {
			Type uint32 `json:"type"`
			Hash []byte `json:"hash"`
		} `json:"inventory"`
	}
	if err := msg.UnmarshalData(&getDataMsg); err != nil {
		return fmt.Errorf("failed to unmarshal getdata message: %v", err)
	}

	// TODO: Implement getdata handler
	return nil
}

// handleInventory handles an inventory message
func (c *Connection) handleInventory(msg *networktypes.Message) error {
	if !c.handshake {
		return fmt.Errorf("handshake not completed")
	}

	var invMsg struct {
		Inventory []struct {
			Type uint32 `json:"type"`
			Hash []byte `json:"hash"`
		} `json:"inventory"`
	}
	if err := msg.UnmarshalData(&invMsg); err != nil {
		return fmt.Errorf("failed to unmarshal inventory message: %v", err)
	}

	// TODO: Implement inventory handler
	return nil
}

// SendMessage sends a message to the peer
func (c *Connection) SendMessage(msg *networktypes.Message) error {
	data, err := networktypes.SerializeBinary(msg)
	if err != nil {
		return err
	}

	if _, err := c.writer.Write(data); err != nil {
		return err
	}

	return c.writer.Flush()
}

// Close closes the connection
func (c *Connection) Close() {
	c.conn.Close()
	c.manager.RemovePeer(c.peer.Address)
}
