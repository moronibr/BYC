package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/byc/internal/blockchain"
	"github.com/byc/internal/logger"
	"github.com/byc/internal/network"
	"github.com/byc/internal/wallet"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Server represents the API server
type Server struct {
	blockchain *blockchain.Blockchain
	node       *network.Node
	router     *mux.Router
	upgrader   websocket.Upgrader
}

// NewServer creates a new API server
func NewServer(bc *blockchain.Blockchain, node *network.Node) *Server {
	server := &Server{
		blockchain: bc,
		node:       node,
		router:     mux.NewRouter(),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins
			},
		},
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
}

// Start starts the API server
func (s *Server) Start(address string) error {
	logger.Info("Starting API server", zap.String("address", address))
	return http.ListenAndServe(address, s.router)
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
	hash := vars["hash"]

	block, err := s.blockchain.GetBlock([]byte(hash))
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
		IsMining: s.node.config.BlockType != "",
		CoinType: string(s.node.config.BlockType),
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

	go s.node.connectToPeer(req.Address)
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
