package network

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/moroni/BYC/internal/blockchain"
	"github.com/moroni/BYC/internal/logger"
	"github.com/moroni/BYC/internal/security"
	"go.uber.org/zap"
)

// SyncManager handles blockchain synchronization
type SyncManager struct {
	node       *Node
	blockchain *blockchain.Blockchain
	security   *security.SecurityManager
	peers      map[string]*Peer
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewSyncManager creates a new sync manager
func NewSyncManager(node *Node, bc *blockchain.Blockchain, security *security.SecurityManager) *SyncManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &SyncManager{
		node:       node,
		blockchain: bc,
		security:   security,
		peers:      make(map[string]*Peer),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start begins the synchronization process
func (sm *SyncManager) Start() {
	go sm.syncLoop()
}

// Stop stops the synchronization process
func (sm *SyncManager) Stop() {
	sm.cancel()
}

// syncLoop continuously checks for new blocks and syncs with peers
func (sm *SyncManager) syncLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-ticker.C:
			sm.checkForNewBlocks()
		}
	}
}

// checkForNewBlocks checks for new blocks from peers
func (sm *SyncManager) checkForNewBlocks() {
	sm.mu.RLock()
	peers := make([]*Peer, 0, len(sm.peers))
	for _, peer := range sm.peers {
		peers = append(peers, peer)
	}
	sm.mu.RUnlock()

	// Get our current block height
	ourHeight := sm.blockchain.GetCurrentHeight()

	// Check each peer for new blocks
	for _, peer := range peers {
		// Request peer's block height
		if err := sm.node.sendMessage(peer, MessageTypeGetHeight, nil); err != nil {
			logger.Error("Failed to request block height",
				zap.String("peer", peer.Address),
				zap.Error(err))
			continue
		}

		// If peer has more blocks, request them
		if peer.Height > ourHeight {
			sm.requestBlocks(peer, uint64(ourHeight+1))
		}
	}
}

// requestBlocks requests blocks from a peer
func (sm *SyncManager) requestBlocks(peer *Peer, startHeight uint64) {
	// Request blocks in batches
	batchSize := uint64(100)
	endHeight := uint64(peer.Height)
	if endHeight > startHeight+batchSize {
		endHeight = startHeight + batchSize
	}

	// Create request message
	request := struct {
		StartHeight uint64 `json:"start_height"`
		EndHeight   uint64 `json:"end_height"`
	}{
		StartHeight: startHeight,
		EndHeight:   endHeight,
	}

	// Send request
	if err := sm.node.sendMessage(peer, MessageTypeGetBlocks, request); err != nil {
		logger.Error("Failed to request blocks",
			zap.String("peer", peer.Address),
			zap.Uint64("start_height", startHeight),
			zap.Uint64("end_height", endHeight),
			zap.Error(err))
		return
	}

	logger.Info("Requested blocks from peer",
		zap.String("peer", peer.Address),
		zap.Uint64("start_height", startHeight),
		zap.Uint64("end_height", endHeight))
}

// HandleBlocks handles incoming blocks from peers
func (sm *SyncManager) HandleBlocks(blocks []*blockchain.Block) error {
	for _, block := range blocks {
		// Validate block using security manager
		if err := sm.security.ValidateBlock(block); err != nil {
			return fmt.Errorf("invalid block: %v", err)
		}

		// Add block to chain
		if err := sm.blockchain.AddBlock(block); err != nil {
			return fmt.Errorf("failed to add block: %v", err)
		}

		logger.Info("Added new block",
			zap.String("hash", string(block.Hash)),
			zap.String("type", string(block.BlockType)))
	}

	return nil
}

// HandleHeight handles incoming height messages from peers
func (sm *SyncManager) HandleHeight(peer *Peer, height uint64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Update peer's height
	peer.Height = int64(height)

	// If peer has more blocks than us, request them
	ourHeight := sm.blockchain.GetCurrentHeight()
	if int64(height) > ourHeight {
		sm.requestBlocks(peer, uint64(ourHeight+1))
	}
}

// AddPeer adds a peer to the sync manager
func (sm *SyncManager) AddPeer(peer *Peer) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.peers[peer.Address] = peer
}

// RemovePeer removes a peer from the sync manager
func (sm *SyncManager) RemovePeer(addr string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.peers, addr)
}

// GetSyncStatus returns the current sync status
func (sm *SyncManager) GetSyncStatus() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	status := make(map[string]interface{})
	status["height"] = sm.blockchain.GetCurrentHeight()
	status["peers"] = len(sm.peers)
	status["syncing"] = false

	// Check if we're syncing
	for _, peer := range sm.peers {
		if peer.Height > sm.blockchain.GetCurrentHeight() {
			status["syncing"] = true
			break
		}
	}

	return status
}
