package network

import (
	"net"
	"sync"
	"time"

	"github.com/youngchain/internal/logger"
)

// PeerInfo represents information about a peer
type PeerInfo struct {
	ID        string    `json:"id"`
	Address   string    `json:"address"`
	Version   string    `json:"version"`
	LastSeen  time.Time `json:"last_seen"`
	Connected bool      `json:"connected"`
}

// Peer represents a network peer
type Peer struct {
	info     PeerInfo
	conn     net.Conn
	sendChan chan []byte
	recvChan chan []byte
	quitChan chan struct{}
	wg       sync.WaitGroup
	mu       sync.RWMutex
}

// NewPeer creates a new peer instance
func NewPeer(conn net.Conn, info PeerInfo) *Peer {
	return &Peer{
		info:     info,
		conn:     conn,
		sendChan: make(chan []byte, 100),
		recvChan: make(chan []byte, 100),
		quitChan: make(chan struct{}),
	}
}

// Start starts the peer's send and receive loops
func (p *Peer) Start() {
	p.wg.Add(2)
	go p.sendLoop()
	go p.receiveLoop()
}

// Stop stops the peer's send and receive loops
func (p *Peer) Stop() {
	close(p.quitChan)
	p.wg.Wait()
	p.conn.Close()
}

// Send sends data to the peer
func (p *Peer) Send(data []byte) {
	select {
	case p.sendChan <- data:
	default:
		logger.Warn("Send channel full, dropping message", logger.String("peer", p.info.ID))
	}
}

func (p *Peer) sendLoop() {
	defer p.wg.Done()

	for {
		select {
		case data := <-p.sendChan:
			if err := p.write(data); err != nil {
				logger.Error("Failed to send data",
					logger.String("peer", p.info.ID),
					logger.Error2(err))
				return
			}
		case <-p.quitChan:
			return
		}
	}
}

func (p *Peer) receiveLoop() {
	defer p.wg.Done()

	buf := make([]byte, 4096)
	for {
		select {
		case <-p.quitChan:
			return
		default:
			n, err := p.conn.Read(buf)
			if err != nil {
				logger.Error("Failed to receive data",
					logger.String("peer", p.info.ID),
					logger.Error2(err))
				return
			}

			data := make([]byte, n)
			copy(data, buf[:n])

			select {
			case p.recvChan <- data:
			default:
				logger.Warn("Receive channel full, dropping message",
					logger.String("peer", p.info.ID))
			}
		}
	}
}

func (p *Peer) write(data []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	_, err := p.conn.Write(data)
	return err
}

// GetInfo returns the peer's information
func (p *Peer) GetInfo() PeerInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.info
}

// GetRecvChan returns the receive channel
func (p *Peer) GetRecvChan() <-chan []byte {
	return p.recvChan
}

// GetSendChan returns the send channel
func (p *Peer) GetSendChan() chan<- []byte {
	return p.sendChan
}

// UpdateLastSeen updates the peer's last seen time
func (p *Peer) UpdateLastSeen() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.info.LastSeen = time.Now()
}
