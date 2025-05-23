package api

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/moroni/BYC/internal/blockchain"
	"github.com/moroni/BYC/internal/logger"
	"github.com/moroni/BYC/internal/network"
	"github.com/moroni/BYC/internal/wallet"
	"go.uber.org/zap"
)

// Server represents the API server
type Server struct {
	blockchain *blockchain.Blockchain
	node       *network.Node
	router     *mux.Router
	upgrader   websocket.Upgrader
	config     *Config
	server     *http.Server
}

// NewServer creates a new API server
func NewServer(bc *blockchain.Blockchain, config *Config) *Server {
	server := &Server{
		blockchain: bc,
		router:     mux.NewRouter(),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins
			},
		},
		config: config,
	}

	// Register routes
	server.registerRoutes()

	return server
}

// registerRoutes registers all API routes
func (s *Server) registerRoutes() {
	// Blockchain routes
	s.router.HandleFunc("/api/blocks", s.getBlocks).Methods("GET")
	s.router.HandleFunc("/api/blocks/{hash}", s.getBlock).Methods("GET")
	s.router.HandleFunc("/api/blocks/latest", s.getLatestBlock).Methods("GET")

	// Transaction routes
	s.router.HandleFunc("/api/transactions", s.getTransactions).Methods("GET")
	s.router.HandleFunc("/api/transactions", s.createTransaction).Methods("POST")
	s.router.HandleFunc("/api/transactions/{id}", s.getTransaction).Methods("GET")

	// Wallet routes
	s.router.HandleFunc("/api/wallet", s.createWallet).Methods("POST")
	s.router.HandleFunc("/api/wallet/{address}/balance", s.getBalance).Methods("GET")
	s.router.HandleFunc("/api/wallet/{address}/balances", s.getAllBalances).Methods("GET")

	// Mining routes
	s.router.HandleFunc("/api/mining/start", s.startMining).Methods("POST")
	s.router.HandleFunc("/api/mining/stop", s.stopMining).Methods("POST")
	s.router.HandleFunc("/api/mining/status", s.getMiningStatus).Methods("GET")

	// Network routes
	s.router.HandleFunc("/api/peers", s.getPeers).Methods("GET")
	s.router.HandleFunc("/api/peers", s.addPeer).Methods("POST")

	// WebSocket route
	s.router.HandleFunc("/ws", s.handleWebSocket)

	// Node info route
	s.router.HandleFunc("/node/info", s.handleGetNodeInfo).Methods("GET")

	// Mine route
	s.router.HandleFunc("/mine", s.mine).Methods("POST")
}

// Start starts the API server
func (s *Server) Start() error {
	// Start the node
	node, err := network.NewNode(&network.Config{
		Address:        s.config.NodeAddress,
		BlockType:      s.config.BlockType,
		BootstrapPeers: s.config.BootstrapPeers,
	})
	if err != nil {
		return fmt.Errorf("failed to start node: %v", err)
	}
	s.node = node

	// Connect to bootstrap peers
	for _, peer := range s.config.BootstrapPeers {
		if err := s.node.ConnectToPeer(peer); err != nil {
			logger.Error("Failed to connect to bootstrap peer", zap.String("peer", peer), zap.Error(err))
		}
	}

	// Start the HTTP server
	s.server = &http.Server{
		Addr:    s.config.NodeAddress,
		Handler: s.router,
	}

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	return nil
}

// Response represents a generic API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// sendResponse sends a JSON response
func (s *Server) sendResponse(w http.ResponseWriter, status int, data interface{}, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := Response{
		Success: err == nil,
		Data:    data,
	}
	if err != nil {
		response.Error = err.Error()
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode response", zap.Error(err))
	}
}

// getBlocks returns all blocks
func (s *Server) getBlocks(w http.ResponseWriter, r *http.Request) {
	blockType := r.URL.Query().Get("type")
	var blocks []*blockchain.Block

	if blockType == "golden" {
		for _, block := range s.blockchain.GoldenBlocks {
			blocks = append(blocks, &block)
		}
	} else if blockType == "silver" {
		for _, block := range s.blockchain.SilverBlocks {
			blocks = append(blocks, &block)
		}
	} else {
		s.sendResponse(w, http.StatusBadRequest, nil, fmt.Errorf("invalid block type"))
		return
	}

	s.sendResponse(w, http.StatusOK, blocks, nil)
}

// getBlock returns a specific block
func (s *Server) getBlock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hashHex := vars["hash"]

	hash, err := hex.DecodeString(hashHex)
	if err != nil {
		s.sendResponse(w, http.StatusBadRequest, nil, fmt.Errorf("invalid hash encoding: %v", err))
		return
	}

	block, err := s.blockchain.GetBlock(hash)
	if err != nil {
		s.sendResponse(w, http.StatusNotFound, nil, err)
		return
	}

	s.sendResponse(w, http.StatusOK, block, nil)
}

// getLatestBlock returns the latest block
func (s *Server) getLatestBlock(w http.ResponseWriter, r *http.Request) {
	blockType := r.URL.Query().Get("type")
	var block *blockchain.Block

	if blockType == "golden" {
		if len(s.blockchain.GoldenBlocks) > 0 {
			block = &s.blockchain.GoldenBlocks[len(s.blockchain.GoldenBlocks)-1]
		}
	} else if blockType == "silver" {
		if len(s.blockchain.SilverBlocks) > 0 {
			block = &s.blockchain.SilverBlocks[len(s.blockchain.SilverBlocks)-1]
		}
	} else {
		s.sendResponse(w, http.StatusBadRequest, nil, fmt.Errorf("invalid block type"))
		return
	}

	if block == nil {
		s.sendResponse(w, http.StatusNotFound, nil, fmt.Errorf("no blocks found"))
		return
	}

	s.sendResponse(w, http.StatusOK, block, nil)
}

// getTransactions returns all transactions
func (s *Server) getTransactions(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	var transactions []*blockchain.Transaction

	if address != "" {
		txs, err := s.blockchain.GetTransactions(address)
		if err != nil {
			s.sendResponse(w, http.StatusInternalServerError, nil, err)
			return
		}
		transactions = txs
	}

	s.sendResponse(w, http.StatusOK, transactions, nil)
}

// createTransaction creates a new transaction
func (s *Server) createTransaction(w http.ResponseWriter, r *http.Request) {
	var tx blockchain.Transaction
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		s.sendResponse(w, http.StatusBadRequest, nil, err)
		return
	}

	if err := s.blockchain.AddTransaction(&tx); err != nil {
		s.sendResponse(w, http.StatusBadRequest, nil, err)
		return
	}

	s.sendResponse(w, http.StatusCreated, tx, nil)
}

// getTransaction returns a specific transaction
func (s *Server) getTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	tx, err := s.blockchain.GetTransaction([]byte(id))
	if err != nil {
		s.sendResponse(w, http.StatusNotFound, nil, err)
		return
	}

	s.sendResponse(w, http.StatusOK, tx, nil)
}

// createWallet creates a new wallet
func (s *Server) createWallet(w http.ResponseWriter, r *http.Request) {
	wlt, err := wallet.NewWallet()
	if err != nil {
		s.sendResponse(w, http.StatusInternalServerError, nil, err)
		return
	}

	s.sendResponse(w, http.StatusCreated, wlt, nil)
}

// getBalance returns the balance for a specific coin type
func (s *Server) getBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	coinType := r.URL.Query().Get("coin_type")

	balance := s.blockchain.GetBalance(address, blockchain.CoinType(coinType))
	s.sendResponse(w, http.StatusOK, balance, nil)
}

// getAllBalances returns balances for all coin types
func (s *Server) getAllBalances(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	wlt := &wallet.Wallet{Address: address}
	balances := wlt.GetAllBalances(s.blockchain)

	s.sendResponse(w, http.StatusOK, balances, nil)
}

// startMining starts mining
func (s *Server) startMining(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CoinType string `json:"coin_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendResponse(w, http.StatusBadRequest, nil, err)
		return
	}

	if err := s.node.StartMining(blockchain.CoinType(req.CoinType)); err != nil {
		s.sendResponse(w, http.StatusBadRequest, nil, err)
		return
	}

	s.sendResponse(w, http.StatusOK, nil, nil)
}

// stopMining stops mining
func (s *Server) stopMining(w http.ResponseWriter, r *http.Request) {
	s.node.StopMining()
	s.sendResponse(w, http.StatusOK, nil, nil)
}

// getMiningStatus returns the current mining status
func (s *Server) getMiningStatus(w http.ResponseWriter, r *http.Request) {
	status := struct {
		IsMining bool   `json:"is_mining"`
		CoinType string `json:"coin_type,omitempty"`
	}{
		IsMining: s.node.Config.BlockType != "",
		CoinType: string(s.node.Config.BlockType),
	}

	s.sendResponse(w, http.StatusOK, status, nil)
}

// getPeers returns the list of connected peers
func (s *Server) getPeers(w http.ResponseWriter, r *http.Request) {
	peers := s.node.GetPeers()
	s.sendResponse(w, http.StatusOK, peers, nil)
}

// addPeer adds a new peer
func (s *Server) addPeer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Address string `json:"address"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendResponse(w, http.StatusBadRequest, nil, err)
		return
	}

	go s.node.ConnectToPeer(req.Address)
	s.sendResponse(w, http.StatusOK, nil, nil)
}

// handleWebSocket handles WebSocket connections
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Failed to upgrade connection", zap.Error(err))
		return
	}
	defer conn.Close()

	// Handle WebSocket messages
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			logger.Error("Failed to read message", zap.Error(err))
			return
		}

		// Echo the message back
		if err := conn.WriteMessage(messageType, p); err != nil {
			logger.Error("Failed to write message", zap.Error(err))
			return
		}
	}
}

// handleGetNodeInfo handles the GET /node/info endpoint
func (s *Server) handleGetNodeInfo(w http.ResponseWriter, r *http.Request) {
	info := struct {
		Address        string   `json:"address"`
		BlockType      string   `json:"blockType"`
		BootstrapPeers []string `json:"bootstrapPeers"`
		Peers          []string `json:"peers"`
	}{
		Address:        s.node.Config.Address,
		BlockType:      string(s.node.Config.BlockType),
		BootstrapPeers: s.node.Config.BootstrapPeers,
		Peers:          make([]string, 0),
	}

	for _, peer := range s.node.Peers {
		info.Peers = append(info.Peers, peer.Address)
	}

	s.sendResponse(w, http.StatusOK, info, nil)
}

// mine starts mining
func (s *Server) mine(w http.ResponseWriter, r *http.Request) {
	if err := s.node.StartMining(blockchain.Leah); err != nil {
		s.sendResponse(w, http.StatusBadRequest, nil, err)
		return
	}
	s.sendResponse(w, http.StatusOK, nil, nil)
}

// ServeHTTP allows Server to be used as an http.Handler in tests
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
