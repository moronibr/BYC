package network

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/youngchain/internal/core/common"
	"github.com/youngchain/internal/network/messages"
)

// Peer represents a network peer
type Peer struct {
	conn   net.Conn
	addr   string
	mu     sync.RWMutex
	stopCh chan struct{}
}

// NewPeer creates a new peer
func NewPeer(conn net.Conn, addr string) *Peer {
	return &Peer{
		conn:   conn,
		addr:   addr,
		stopCh: make(chan struct{}),
	}
}

// ID returns the peer's address as its ID
func (p *Peer) ID() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.addr
}

// Send sends a message to the peer
func (p *Peer) Send(data []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Write message length
	length := uint32(len(data))
	if err := binary.Write(p.conn, binary.BigEndian, length); err != nil {
		return fmt.Errorf("failed to write message length: %v", err)
	}

	// Write message
	if _, err := p.conn.Write(data); err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}

	return nil
}

// SendTransaction sends a transaction to the peer
func (p *Peer) SendTransaction(tx *common.Transaction) error {
	msg := &messages.TransactionMessage{
		Transaction: tx,
		CoinType:    "default",
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction message: %v", err)
	}
	return p.Send(data)
}

// Close closes the peer connection
func (p *Peer) Close() error {
	close(p.stopCh)
	return p.conn.Close()
}
