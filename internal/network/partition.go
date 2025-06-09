package network

import (
	"fmt"
	"sync"
	"time"

	"byc/internal/logger"

	"go.uber.org/zap"
)

// PartitionState represents the state of a network partition
type PartitionState struct {
	DetectedAt    time.Time
	LastCheck     time.Time
	AffectedPeers map[string]bool
	RecoveryMode  bool
}

// PartitionManager manages network partition detection and recovery
type PartitionManager struct {
	networkManager *NetworkManager
	state          *PartitionState
	mu             sync.RWMutex
	checkInterval  time.Duration
	timeout        time.Duration
}

// NewPartitionManager creates a new partition manager
func NewPartitionManager(nm *NetworkManager) *PartitionManager {
	return &PartitionManager{
		networkManager: nm,
		state: &PartitionState{
			AffectedPeers: make(map[string]bool),
		},
		checkInterval: 30 * time.Second,
		timeout:       5 * time.Minute,
	}
}

// Start starts the partition manager
func (pm *PartitionManager) Start() {
	go pm.monitorLoop()
}

// Stop stops the partition manager
func (pm *PartitionManager) Stop() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.state.RecoveryMode = false
}

// IsPartitioned checks if the network is currently partitioned
func (pm *PartitionManager) IsPartitioned() bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return !pm.state.DetectedAt.IsZero() && pm.state.RecoveryMode
}

// GetAffectedPeers returns the list of affected peers
func (pm *PartitionManager) GetAffectedPeers() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	peers := make([]string, 0, len(pm.state.AffectedPeers))
	for peer := range pm.state.AffectedPeers {
		peers = append(peers, peer)
	}
	return peers
}

// monitorLoop continuously monitors the network for partitions
func (pm *PartitionManager) monitorLoop() {
	ticker := time.NewTicker(pm.checkInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := pm.checkPartition(); err != nil {
			logger.Error("failed to check partition", zap.Error(err))
		}
	}
}

// checkPartition checks for network partitions
func (pm *PartitionManager) checkPartition() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	peers := pm.networkManager.GetPeers()
	for _, peer := range peers {
		if err := pm.checkPeerConnectivity(peer.Address); err != nil {
			pm.state.AffectedPeers[peer.Address] = true
			if !pm.state.RecoveryMode {
				pm.state.DetectedAt = time.Now()
				pm.state.RecoveryMode = true
			}
		} else {
			delete(pm.state.AffectedPeers, peer.Address)
		}
	}

	pm.state.LastCheck = time.Now()
	return nil
}

// checkPeerConnectivity checks if a peer is reachable
func (pm *PartitionManager) checkPeerConnectivity(peerAddr string) error {
	// Send ping message
	msg := NewNetworkMessage(MessageTypePing, pm.networkManager.config.NodeID, peerAddr, []byte("ping"))

	if err := pm.networkManager.SendMessage(msg); err != nil {
		return fmt.Errorf("failed to send ping to %s: %v", peerAddr, err)
	}

	// Wait for pong response
	timeout := time.After(pm.timeout)
	for {
		select {
		case <-timeout:
			return fmt.Errorf("ping timeout for %s", peerAddr)
		default:
			// Check if we received a pong
			if pm.networkManager.HasReceivedPong(peerAddr) {
				return nil
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// RecoverPartition attempts to recover from a network partition
func (pm *PartitionManager) RecoverPartition() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if !pm.state.RecoveryMode {
		return fmt.Errorf("no active partition to recover from")
	}

	// Try to reconnect to affected peers
	for peer := range pm.state.AffectedPeers {
		if err := pm.networkManager.ConnectToPeer(peer); err != nil {
			logger.Error("failed to reconnect to peer",
				zap.String("peer", peer),
				zap.Error(err))
			continue
		}

		// Verify connection
		if err := pm.checkPeerConnectivity(peer); err == nil {
			delete(pm.state.AffectedPeers, peer)
			logger.Info("reconnected to peer",
				zap.String("peer", peer))
		}
	}

	// Check if partition is resolved
	if len(pm.state.AffectedPeers) == 0 {
		pm.state.DetectedAt = time.Time{}
		pm.state.RecoveryMode = false
		logger.Info("network partition resolved")
	}

	return nil
}

// GetPartitionState returns the current partition state
func (pm *PartitionManager) GetPartitionState() *PartitionState {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return &PartitionState{
		DetectedAt:    pm.state.DetectedAt,
		LastCheck:     pm.state.LastCheck,
		AffectedPeers: pm.state.AffectedPeers,
		RecoveryMode:  pm.state.RecoveryMode,
	}
}

// SetCheckInterval sets the partition check interval
func (pm *PartitionManager) SetCheckInterval(interval time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.checkInterval = interval
}

// SetTimeout sets the partition timeout
func (pm *PartitionManager) SetTimeout(timeout time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.timeout = timeout
}
