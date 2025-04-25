package network

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/coin"
)

// Server represents a network server for node communication
type Server struct {
	// Server configuration
	Address  string
	Port     int
	MaxPeers int

	// Node management
	Peers      map[string]*Peer
	PeersMutex sync.RWMutex

	// Message handling
	MessageChan chan *Message

	// Server state
	IsRunning bool
	StopChan  chan struct{}
}

// Peer represents a connected peer
type Peer struct {
	// Connection information
	Address    string
	Connection net.Conn

	// Peer state
	Connected bool
	LastSeen  time.Time

	// Peer capabilities
	Version  uint32
	Services uint64

	// Message handling
	SendChan chan *Message
	RecvChan chan *Message
}

// NewServer creates a new network server
func NewServer(address string, port int, maxPeers int) *Server {
	return &Server{
		Address:     address,
		Port:        port,
		MaxPeers:    maxPeers,
		Peers:       make(map[string]*Peer),
		MessageChan: make(chan *Message, 100),
		StopChan:    make(chan struct{}),
	}
}

// Start starts the server
func (s *Server) Start() error {
	// Create listener
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Address, s.Port))
	if err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}

	s.IsRunning = true
	log.Printf("Server started on %s:%d", s.Address, s.Port)

	// Accept connections
	go s.acceptConnections(listener)

	// Process messages
	go s.processMessages()

	return nil
}

// Stop stops the server
func (s *Server) Stop() {
	s.IsRunning = false
	close(s.StopChan)

	// Close all peer connections
	s.PeersMutex.Lock()
	defer s.PeersMutex.Unlock()

	for _, peer := range s.Peers {
		peer.Connection.Close()
	}

	s.Peers = make(map[string]*Peer)
	log.Println("Server stopped")
}

// acceptConnections accepts incoming connections
func (s *Server) acceptConnections(listener net.Listener) {
	for s.IsRunning {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		// Check if we have too many peers
		s.PeersMutex.RLock()
		if len(s.Peers) >= s.MaxPeers {
			s.PeersMutex.RUnlock()
			conn.Close()
			log.Println("Rejected connection: too many peers")
			continue
		}
		s.PeersMutex.RUnlock()

		// Create new peer
		peer := &Peer{
			Address:    conn.RemoteAddr().String(),
			Connection: conn,
			Connected:  true,
			LastSeen:   time.Now(),
			SendChan:   make(chan *Message, 10),
			RecvChan:   make(chan *Message, 10),
		}

		// Add peer to list
		s.PeersMutex.Lock()
		s.Peers[peer.Address] = peer
		s.PeersMutex.Unlock()

		log.Printf("New peer connected: %s", peer.Address)

		// Handle peer
		go s.handlePeer(peer)
	}
}

// handlePeer handles communication with a peer
func (s *Server) handlePeer(peer *Peer) {
	defer func() {
		peer.Connection.Close()
		s.PeersMutex.Lock()
		delete(s.Peers, peer.Address)
		s.PeersMutex.Unlock()
		log.Printf("Peer disconnected: %s", peer.Address)
	}()

	// Send version message
	versionMsg := &VersionMessage{
		Version:     1,
		Services:    1,
		Timestamp:   time.Now().Unix(),
		AddrRecv:    peer.Address,
		AddrTrans:   s.Address,
		Nonce:       uint64(time.Now().UnixNano()),
		UserAgent:   "YoungChain/1.0.0",
		StartHeight: 0,
		Relay:       true,
	}

	versionPayload, _ := json.Marshal(versionMsg)
	versionNetworkMsg := NewMessage(VersionMsg, versionPayload)
	peer.SendChan <- versionNetworkMsg

	// Handle sending messages to peer
	go s.sendMessagesToPeer(peer)

	// Handle receiving messages from peer
	for s.IsRunning {
		// Read message length (4 bytes)
		lengthBuf := make([]byte, 4)
		_, err := peer.Connection.Read(lengthBuf)
		if err != nil {
			log.Printf("Error reading message length from %s: %v", peer.Address, err)
			return
		}

		// Convert length to int
		length := int(lengthBuf[0]) | int(lengthBuf[1])<<8 | int(lengthBuf[2])<<16 | int(lengthBuf[3])<<24

		// Read message
		messageBuf := make([]byte, length)
		_, err = peer.Connection.Read(messageBuf)
		if err != nil {
			log.Printf("Error reading message from %s: %v", peer.Address, err)
			return
		}

		// Deserialize message
		msg, err := DeserializeMessage(messageBuf)
		if err != nil {
			log.Printf("Error deserializing message from %s: %v", peer.Address, err)
			continue
		}

		// Update last seen time
		peer.LastSeen = time.Now()

		// Handle message
		s.handleMessage(peer, msg)
	}
}

// sendMessagesToPeer sends messages to a peer
func (s *Server) sendMessagesToPeer(peer *Peer) {
	for s.IsRunning {
		select {
		case msg := <-peer.SendChan:
			// Serialize message
			msgBytes, err := msg.Serialize()
			if err != nil {
				log.Printf("Error serializing message to %s: %v", peer.Address, err)
				continue
			}

			// Write message length (4 bytes)
			length := len(msgBytes)
			lengthBuf := []byte{
				byte(length & 0xFF),
				byte((length >> 8) & 0xFF),
				byte((length >> 16) & 0xFF),
				byte((length >> 24) & 0xFF),
			}

			_, err = peer.Connection.Write(lengthBuf)
			if err != nil {
				log.Printf("Error writing message length to %s: %v", peer.Address, err)
				return
			}

			// Write message
			_, err = peer.Connection.Write(msgBytes)
			if err != nil {
				log.Printf("Error writing message to %s: %v", peer.Address, err)
				return
			}

		case <-s.StopChan:
			return
		}
	}
}

// handleMessage handles a message from a peer
func (s *Server) handleMessage(peer *Peer, msg *Message) {
	switch msg.Command {
	case VersionMsg:
		// Handle version message
		var versionMsg VersionMessage
		err := json.Unmarshal(msg.Payload, &versionMsg)
		if err != nil {
			log.Printf("Error unmarshaling version message from %s: %v", peer.Address, err)
			return
		}

		peer.Version = versionMsg.Version
		peer.Services = versionMsg.Services

		// Send verack message
		verackMsg := NewMessage(VerAckMsg, []byte{})
		peer.SendChan <- verackMsg

	case VerAckMsg:
		// Handle verack message
		log.Printf("Received verack from %s", peer.Address)

	case BlockMsg:
		// Handle block message
		var blockMsg BlockMessage
		err := json.Unmarshal(msg.Payload, &blockMsg)
		if err != nil {
			log.Printf("Error unmarshaling block message from %s: %v", peer.Address, err)
			return
		}

		// Process block
		s.processBlock(blockMsg.Block, blockMsg.BlockType)

	case TxMsg:
		// Handle transaction message
		var txMsg TransactionMessage
		err := json.Unmarshal(msg.Payload, &txMsg)
		if err != nil {
			log.Printf("Error unmarshaling transaction message from %s: %v", peer.Address, err)
			return
		}

		// Process transaction
		s.processTransaction(txMsg.Transaction, txMsg.CoinType)

	default:
		// Forward message to message channel
		s.MessageChan <- msg
	}
}

// processBlock processes a block
func (s *Server) processBlock(block *block.Block, blockType block.BlockType) {
	// TODO: Implement block processing
	log.Printf("Processing %s block: %s", blockType, block.String())
}

// processTransaction processes a transaction
func (s *Server) processTransaction(tx block.Transaction, coinType coin.CoinType) {
	// TODO: Implement transaction processing
	log.Printf("Processing transaction with coin type: %s", coinType)
}

// processMessages processes messages from the message channel
func (s *Server) processMessages() {
	for s.IsRunning {
		select {
		case msg := <-s.MessageChan:
			// Process message
			log.Printf("Processing message: %s", msg.Command)

		case <-s.StopChan:
			return
		}
	}
}

// BroadcastMessage broadcasts a message to all peers
func (s *Server) BroadcastMessage(msg *Message) {
	s.PeersMutex.RLock()
	defer s.PeersMutex.RUnlock()

	for _, peer := range s.Peers {
		peer.SendChan <- msg
	}
}

// ConnectToPeer connects to a peer
func (s *Server) ConnectToPeer(address string) error {
	// Check if we already have this peer
	s.PeersMutex.RLock()
	if _, exists := s.Peers[address]; exists {
		s.PeersMutex.RUnlock()
		return fmt.Errorf("already connected to peer: %s", address)
	}
	s.PeersMutex.RUnlock()

	// Connect to peer
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to connect to peer %s: %v", address, err)
	}

	// Create new peer
	peer := &Peer{
		Address:    address,
		Connection: conn,
		Connected:  true,
		LastSeen:   time.Now(),
		SendChan:   make(chan *Message, 10),
		RecvChan:   make(chan *Message, 10),
	}

	// Add peer to list
	s.PeersMutex.Lock()
	s.Peers[address] = peer
	s.PeersMutex.Unlock()

	log.Printf("Connected to peer: %s", address)

	// Handle peer
	go s.handlePeer(peer)

	return nil
}
