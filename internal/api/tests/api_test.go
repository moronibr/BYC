package tests

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"byc/internal/api"
	"byc/internal/blockchain"
	"byc/internal/wallet"

	"github.com/stretchr/testify/assert"
)

func TestGetBalance(t *testing.T) {
	bc := blockchain.NewBlockchain()
	w, err := wallet.NewWallet()
	assert.NoError(t, err)

	config := &api.Config{
		NodeAddress:    ":0",
		BlockType:      blockchain.GoldenBlock,
		BootstrapPeers: []string{},
	}
	server := api.NewServer(bc, config)

	url := "/api/wallet/" + w.Address + "/balance"
	req := httptest.NewRequest("GET", url, nil)
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp api.Response
	err = json.NewDecoder(rr.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestGetBlocks(t *testing.T) {
	bc := blockchain.NewBlockchain()
	config := &api.Config{
		NodeAddress:    ":0",
		BlockType:      blockchain.GoldenBlock,
		BootstrapPeers: []string{},
	}
	server := api.NewServer(bc, config)

	req := httptest.NewRequest("GET", "/api/blocks?type=golden", nil)
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp api.Response
	err := json.NewDecoder(rr.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestCreateTransaction(t *testing.T) {
	bc := blockchain.NewBlockchain()
	config := &api.Config{
		NodeAddress:    ":0",
		BlockType:      blockchain.GoldenBlock,
		BootstrapPeers: []string{},
	}
	server := api.NewServer(bc, config)

	tx := blockchain.Transaction{
		ID:        []byte("testid"),
		Inputs:    []blockchain.TxInput{},
		Outputs:   []blockchain.TxOutput{},
		Timestamp: time.Now(),
		BlockType: blockchain.GoldenBlock,
	}
	body, _ := json.Marshal(tx)
	req := httptest.NewRequest("POST", "/api/transactions", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp api.Response
	err := json.NewDecoder(rr.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestGetBlock(t *testing.T) {
	bc := blockchain.NewBlockchain()
	config := &api.Config{
		NodeAddress:    ":0",
		BlockType:      blockchain.GoldenBlock,
		BootstrapPeers: []string{},
	}
	server := api.NewServer(bc, config)

	genesis := bc.GoldenBlocks[0]
	hash := hex.EncodeToString(genesis.Hash)
	t.Logf("Genesis Block: %+v", genesis)
	t.Logf("Hash: %s", hash)
	url := "/api/blocks/" + hash
	req := httptest.NewRequest("GET", url, nil)
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp api.Response
	err := json.NewDecoder(rr.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
}
