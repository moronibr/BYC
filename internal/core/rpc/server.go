package rpc

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/coin"
	"github.com/youngchain/internal/core/transaction"
	"github.com/youngchain/internal/core/wallet"
)

// RPCServer represents the RPC server
type RPCServer struct {
	blockchain *block.Blockchain
	txPool     *transaction.TxPool
	wallets    map[string]*wallet.Wallet
	upgrader   websocket.Upgrader
	mu         sync.RWMutex
}

// NewRPCServer creates a new RPC server
func NewRPCServer(blockchain *block.Blockchain, txPool *transaction.TxPool) *RPCServer {
	return &RPCServer{
		blockchain: blockchain,
		txPool:     txPool,
		wallets:    make(map[string]*wallet.Wallet),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins
			},
		},
	}
}

// Start starts the RPC server
func (s *RPCServer) Start(addr string) error {
	http.HandleFunc("/rpc", s.handleRPC)
	http.HandleFunc("/ws", s.handleWebSocket)
	return http.ListenAndServe(addr, nil)
}

// handleRPC handles HTTP RPC requests
func (s *RPCServer) handleRPC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		JSONRPC string          `json:"jsonrpc"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params"`
		ID      interface{}     `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.sendError(w, nil, -32700, "Parse error", err)
		return
	}

	if request.JSONRPC != "2.0" {
		s.sendError(w, request.ID, -32600, "Invalid Request", nil)
		return
	}

	response := s.handleRequest(request.Method, request.Params, request.ID)
	json.NewEncoder(w).Encode(response)
}

// handleWebSocket handles WebSocket connections
func (s *RPCServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		var request struct {
			JSONRPC string          `json:"jsonrpc"`
			Method  string          `json:"method"`
			Params  json.RawMessage `json:"params"`
			ID      interface{}     `json:"id"`
		}

		if err := conn.ReadJSON(&request); err != nil {
			break
		}

		if request.JSONRPC != "2.0" {
			s.sendWebSocketError(conn, request.ID, -32600, "Invalid Request", nil)
			continue
		}

		response := s.handleRequest(request.Method, request.Params, request.ID)
		conn.WriteJSON(response)
	}
}

// handleRequest handles RPC requests
func (s *RPCServer) handleRequest(method string, params json.RawMessage, id interface{}) interface{} {
	switch method {
	case "getblockcount":
		return s.handleGetBlockCount(id)
	case "getbestblockhash":
		return s.handleGetBestBlockHash(id)
	case "getblock":
		return s.handleGetBlock(params, id)
	case "gettransaction":
		return s.handleGetTransaction(params, id)
	case "sendtransaction":
		return s.handleSendTransaction(params, id)
	case "getbalance":
		return s.handleGetBalance(params, id)
	case "createwallet":
		return s.handleCreateWallet(params, id)
	case "loadwallet":
		return s.handleLoadWallet(params, id)
	default:
		return s.createError(id, -32601, "Method not found", nil)
	}
}

// handleGetBlockCount handles getblockcount RPC
func (s *RPCServer) handleGetBlockCount(id interface{}) interface{} {
	count := s.blockchain.GetBlockCount()
	return s.createResponse(id, count)
}

// handleGetBestBlockHash handles getbestblockhash RPC
func (s *RPCServer) handleGetBestBlockHash(id interface{}) interface{} {
	bestBlock := s.blockchain.GetBestBlock()
	if bestBlock == nil {
		return s.createError(id, -32603, "Internal error", nil)
	}
	return s.createResponse(id, bestBlock.Header.Hash)
}

// handleGetBlock handles getblock RPC
func (s *RPCServer) handleGetBlock(params json.RawMessage, id interface{}) interface{} {
	var request struct {
		Hash string `json:"hash"`
	}
	if err := json.Unmarshal(params, &request); err != nil {
		return s.createError(id, -32602, "Invalid params", err)
	}

	block := s.blockchain.GetBlockByHash([]byte(request.Hash))
	if block == nil {
		return s.createError(id, -32603, "Block not found", nil)
	}

	return s.createResponse(id, block)
}

// handleGetTransaction handles gettransaction RPC
func (s *RPCServer) handleGetTransaction(params json.RawMessage, id interface{}) interface{} {
	var request struct {
		Hash string `json:"hash"`
	}
	if err := json.Unmarshal(params, &request); err != nil {
		return s.createError(id, -32602, "Invalid params", err)
	}

	tx := s.blockchain.GetTransactionByHash([]byte(request.Hash))
	if tx == nil {
		return s.createError(id, -32603, "Transaction not found", nil)
	}

	return s.createResponse(id, tx)
}

// handleSendTransaction handles sendtransaction RPC
func (s *RPCServer) handleSendTransaction(params json.RawMessage, id interface{}) interface{} {
	var request struct {
		Transaction *transaction.Transaction `json:"transaction"`
	}
	if err := json.Unmarshal(params, &request); err != nil {
		return s.createError(id, -32602, "Invalid params", err)
	}

	if err := s.txPool.AddTransaction(request.Transaction); err != nil {
		return s.createError(id, -32603, "Failed to add transaction", err)
	}

	return s.createResponse(id, request.Transaction.Hash)
}

// handleGetBalance handles getbalance RPC
func (s *RPCServer) handleGetBalance(params json.RawMessage, id interface{}) interface{} {
	var request struct {
		Address string `json:"address"`
	}
	if err := json.Unmarshal(params, &request); err != nil {
		return s.createError(id, -32602, "Invalid params", err)
	}

	wallet, ok := s.wallets[request.Address]
	if !ok {
		return s.createError(id, -32603, "Wallet not found", nil)
	}

	balance, err := wallet.GetBalance(s.txPool)
	if err != nil {
		return s.createError(id, -32603, "Failed to get balance", err)
	}

	return s.createResponse(id, balance)
}

// handleCreateWallet handles createwallet RPC
func (s *RPCServer) handleCreateWallet(params json.RawMessage, id interface{}) interface{} {
	var request struct {
		CoinType string `json:"coin_type"`
	}
	if err := json.Unmarshal(params, &request); err != nil {
		return s.createError(id, -32602, "Invalid params", err)
	}

	wallet, err := wallet.NewWallet(coin.CoinType(request.CoinType))
	if err != nil {
		return s.createError(id, -32603, "Failed to create wallet", err)
	}

	s.mu.Lock()
	s.wallets[wallet.Address] = wallet
	s.mu.Unlock()

	return s.createResponse(id, wallet)
}

// handleLoadWallet handles loadwallet RPC
func (s *RPCServer) handleLoadWallet(params json.RawMessage, id interface{}) interface{} {
	var request struct {
		Filename string `json:"filename"`
	}
	if err := json.Unmarshal(params, &request); err != nil {
		return s.createError(id, -32602, "Invalid params", err)
	}

	wallet, err := wallet.LoadWallet(request.Filename)
	if err != nil {
		return s.createError(id, -32603, "Failed to load wallet", err)
	}

	s.mu.Lock()
	s.wallets[wallet.Address] = wallet
	s.mu.Unlock()

	return s.createResponse(id, wallet)
}

// createResponse creates a JSON-RPC response
func (s *RPCServer) createResponse(id interface{}, result interface{}) interface{} {
	return struct {
		JSONRPC string      `json:"jsonrpc"`
		Result  interface{} `json:"result"`
		ID      interface{} `json:"id"`
	}{
		JSONRPC: "2.0",
		Result:  result,
		ID:      id,
	}
}

// createError creates a JSON-RPC error response
func (s *RPCServer) createError(id interface{}, code int, message string, err error) interface{} {
	var data interface{}
	if err != nil {
		data = err.Error()
	}

	return struct {
		JSONRPC string      `json:"jsonrpc"`
		Error   ErrorObject `json:"error"`
		ID      interface{} `json:"id"`
	}{
		JSONRPC: "2.0",
		Error: ErrorObject{
			Code:    code,
			Message: message,
			Data:    data,
		},
		ID: id,
	}
}

// sendError sends an HTTP error response
func (s *RPCServer) sendError(w http.ResponseWriter, id interface{}, code int, message string, err error) {
	response := s.createError(id, code, message, err)
	json.NewEncoder(w).Encode(response)
}

// sendWebSocketError sends a WebSocket error response
func (s *RPCServer) sendWebSocketError(conn *websocket.Conn, id interface{}, code int, message string, err error) {
	response := s.createError(id, code, message, err)
	conn.WriteJSON(response)
}

// ErrorObject represents a JSON-RPC error object
type ErrorObject struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
