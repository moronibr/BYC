package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/byc/internal/api"
	"github.com/byc/internal/blockchain"
	"github.com/byc/internal/wallet"
	"github.com/stretchr/testify/assert"
)

func setupTestAPI(t *testing.T) (*api.Server, *wallet.Wallet) {
	// Create test wallet
	w, err := wallet.NewWallet()
	assert.NoError(t, err)

	// Create test server
	server := api.NewServer(":0", w)
	assert.NotNil(t, server)

	return server, w
}

func TestGetBalance(t *testing.T) {
	server, w := setupTestAPI(t)
	defer server.Close()

	// Create test request
	req, err := http.NewRequest("GET", "/balance/"+w.Address+"/Leah", nil)
	assert.NoError(t, err)

	// Create response recorder
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.GetBalance)

	// Serve request
	handler.ServeHTTP(rr, req)

	// Check response
	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]float64
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "balance")
}

func TestCreateTransaction(t *testing.T) {
	server, w := setupTestAPI(t)
	defer server.Close()

	// Create test transaction request
	txReq := api.TransactionRequest{
		To:       "recipient_address",
		Amount:   10.0,
		CoinType: blockchain.Leah,
	}

	// Convert request to JSON
	jsonData, err := json.Marshal(txReq)
	assert.NoError(t, err)

	// Create test request
	req, err := http.NewRequest("POST", "/transaction", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.CreateTransaction)

	// Serve request
	handler.ServeHTTP(rr, req)

	// Check response
	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "transaction_id")
}

func TestGetBlock(t *testing.T) {
	server, _ := setupTestAPI(t)
	defer server.Close()

	// Create test request with a known block hash
	req, err := http.NewRequest("GET", "/block/test_hash", nil)
	assert.NoError(t, err)

	// Create response recorder
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.GetBlock)

	// Serve request
	handler.ServeHTTP(rr, req)

	// Check response
	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "hash")
}

func TestGetTransaction(t *testing.T) {
	server, _ := setupTestAPI(t)
	defer server.Close()

	// Create test request with a known transaction ID
	req, err := http.NewRequest("GET", "/transaction/test_tx", nil)
	assert.NoError(t, err)

	// Create response recorder
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.GetTransaction)

	// Serve request
	handler.ServeHTTP(rr, req)

	// Check response
	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "id")
}

func TestCreateSpecialCoin(t *testing.T) {
	server, w := setupTestAPI(t)
	defer server.Close()

	// Test creating Ephraim coin
	req, err := http.NewRequest("POST", "/special/ephraim", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.CreateEphraimCoin)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Test creating Manasseh coin
	req, err = http.NewRequest("POST", "/special/manasseh", nil)
	assert.NoError(t, err)

	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(server.CreateManassehCoin)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Test creating Joseph coin
	req, err = http.NewRequest("POST", "/special/joseph", nil)
	assert.NoError(t, err)

	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(server.CreateJosephCoin)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestInvalidRequests(t *testing.T) {
	server, _ := setupTestAPI(t)
	defer server.Close()

	// Test invalid balance request
	req, err := http.NewRequest("GET", "/balance/invalid_address/InvalidCoin", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.GetBalance)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// Test invalid transaction request
	invalidTxReq := api.TransactionRequest{
		To:       "recipient_address",
		Amount:   -10.0, // Invalid amount
		CoinType: "InvalidCoin",
	}

	jsonData, err := json.Marshal(invalidTxReq)
	assert.NoError(t, err)

	req, err = http.NewRequest("POST", "/transaction", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(server.CreateTransaction)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetAllBalances(t *testing.T) {
	server, w := setupTestAPI(t)
	defer server.Close()

	// Create test request
	req, err := http.NewRequest("GET", "/balances/"+w.Address, nil)
	assert.NoError(t, err)

	// Create response recorder
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.GetAllBalances)

	// Serve request
	handler.ServeHTTP(rr, req)

	// Check response
	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]float64
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, blockchain.Leah)
	assert.Contains(t, response, blockchain.Shiblum)
	assert.Contains(t, response, blockchain.Senum)
}
