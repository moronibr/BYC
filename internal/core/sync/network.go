package sync

import (
	"fmt"
	"sync"
	"time"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/utxo"
)

// Network represents the network manager
type Network struct {
	mu sync.RWMutex

	// Block sync
	blockSync *BlockSync

	// Peer manager
	peerManager *PeerManager

	// Stop channel
	stopChan chan struct{}

	// Listen address
	listenAddr string

	// UTXO set
	utxoSet utxo.UTXOSetInterface

	// Blockchain
	blockchain *block.Blockchain
}

// NewNetwork creates a new network manager
func NewNetwork(listenAddr string, blockchain *block.Blockchain, utxoSet utxo.UTXOSetInterface) *Network {
	// Create block sync
	blockSync := NewBlockSync(blockchain, utxoSet)

	// Create peer manager
	peerManager := NewPeerManager(listenAddr, blockSync)

	return &Network{
		blockSync:   blockSync,
		peerManager: peerManager,
		stopChan:    make(chan struct{}),
		listenAddr:  listenAddr,
		utxoSet:     utxoSet,
		blockchain:  blockchain,
	}
}

// Start starts the network manager
func (n *Network) Start() {
	// Start block sync
	n.blockSync.Start()

	// Start peer manager
	n.peerManager.Start()

	// Start block broadcasting
	go n.broadcastBlocks()
}

// Stop stops the network manager
func (n *Network) Stop() {
	close(n.stopChan)

	// Stop block sync
	n.blockSync.Stop()

	// Stop peer manager
	n.peerManager.Stop()
}

// broadcastBlocks broadcasts new blocks to peers
func (n *Network) broadcastBlocks() {
	for {
		select {
		case <-n.stopChan:
			return
		default:
			// Get best block
			bestBlock := n.blockchain.GetBestBlock()
			if bestBlock == nil {
				time.Sleep(time.Second)
				continue
			}

			// Broadcast block
			n.peerManager.BroadcastBlock(bestBlock)

			// Wait before next broadcast
			time.Sleep(time.Second)
		}
	}
}

// HandleBlock handles an incoming block
func (n *Network) HandleBlock(block *block.Block) error {
	// Handle block
	if err := n.blockSync.HandleBlock(block); err != nil {
		return fmt.Errorf("failed to handle block: %v", err)
	}

	// Broadcast block
	n.peerManager.BroadcastBlock(block)

	return nil
}

// GetPeers returns the current peers
func (n *Network) GetPeers() []*Peer {
	return n.peerManager.GetPeers()
}

// GetSyncStatus returns the current synchronization status
func (n *Network) GetSyncStatus() (bool, uint64, time.Time) {
	return n.blockSync.GetSyncStatus()
}

// RequestBlocks requests blocks from peers
func (n *Network) RequestBlocks(startHeight uint64) error {
	return n.blockSync.RequestBlocks(startHeight)
}

// BroadcastInv broadcasts an inventory message to all peers
func (n *Network) BroadcastInv(inventory []byte) {
	n.peerManager.BroadcastInv(inventory)
}

// GetBlockSync returns the block synchronizer
func (n *Network) GetBlockSync() *BlockSync {
	return n.blockSync
}

// GetPeerManager returns the peer manager
func (n *Network) GetPeerManager() *PeerManager {
	return n.peerManager
}
