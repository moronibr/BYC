package network

import (
	"fmt"
	"sync"
	"time"
)

// PartitionManager manages network partitions
type PartitionManager struct {
	mu sync.RWMutex

	// Partition tracking
	partitions    map[string]*Partition
	activePeers   map[string]time.Time
	lastMergeTime time.Time

	// Configuration
	partitionTimeout time.Duration
	mergeTimeout     time.Duration
	maxPartitions    int
}

// Partition represents a network partition
type Partition struct {
	ID        string
	Peers     map[string]time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewPartitionManager creates a new partition manager
func NewPartitionManager(partitionTimeout, mergeTimeout time.Duration, maxPartitions int) *PartitionManager {
	return &PartitionManager{
		partitions:       make(map[string]*Partition),
		activePeers:      make(map[string]time.Time),
		partitionTimeout: partitionTimeout,
		mergeTimeout:     mergeTimeout,
		maxPartitions:    maxPartitions,
	}
}

// UpdatePeer updates peer activity and manages partitions
func (pm *PartitionManager) UpdatePeer(peerID string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	now := time.Now()
	pm.activePeers[peerID] = now

	// Check for partition formation
	pm.checkPartitions(now)
}

// checkPartitions checks for partition formation and merging
func (pm *PartitionManager) checkPartitions(now time.Time) {
	// Remove inactive peers
	for peerID, lastSeen := range pm.activePeers {
		if now.Sub(lastSeen) > pm.partitionTimeout {
			delete(pm.activePeers, peerID)
		}
	}

	// Check for new partitions
	if len(pm.partitions) < pm.maxPartitions {
		// Find disconnected peer groups
		groups := pm.findDisconnectedGroups()
		for _, group := range groups {
			if len(group) > 1 {
				pm.createPartition(group, now)
			}
		}
	}

	// Check for partition merging
	if now.Sub(pm.lastMergeTime) > pm.mergeTimeout {
		pm.mergePartitions(now)
	}
}

// findDisconnectedGroups finds groups of disconnected peers
func (pm *PartitionManager) findDisconnectedGroups() [][]string {
	var groups [][]string
	visited := make(map[string]bool)

	for peerID := range pm.activePeers {
		if visited[peerID] {
			continue
		}

		group := pm.findConnectedPeers(peerID, visited)
		if len(group) > 0 {
			groups = append(groups, group)
		}
	}

	return groups
}

// findConnectedPeers finds all peers connected to a given peer
func (pm *PartitionManager) findConnectedPeers(peerID string, visited map[string]bool) []string {
	var group []string
	queue := []string{peerID}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}

		visited[current] = true
		group = append(group, current)

		// Add connected peers to queue
		for peer := range pm.activePeers {
			if !visited[peer] && pm.arePeersConnected(current, peer) {
				queue = append(queue, peer)
			}
		}
	}

	return group
}

// arePeersConnected checks if two peers are connected
func (pm *PartitionManager) arePeersConnected(peer1, peer2 string) bool {
	// This is a simplified check. In a real implementation,
	// you would check actual network connectivity.
	return true
}

// createPartition creates a new partition
func (pm *PartitionManager) createPartition(peers []string, now time.Time) {
	partitionID := fmt.Sprintf("partition-%d", len(pm.partitions)+1)
	partition := &Partition{
		ID:        partitionID,
		Peers:     make(map[string]time.Time),
		CreatedAt: now,
		UpdatedAt: now,
	}

	for _, peer := range peers {
		partition.Peers[peer] = now
	}

	pm.partitions[partitionID] = partition
}

// mergePartitions merges partitions that can communicate
func (pm *PartitionManager) mergePartitions(now time.Time) {
	// Find partitions that can be merged
	merged := make(map[string]bool)
	for id1, p1 := range pm.partitions {
		if merged[id1] {
			continue
		}

		for id2, p2 := range pm.partitions {
			if id1 == id2 || merged[id2] {
				continue
			}

			if pm.canMergePartitions(p1, p2) {
				pm.mergePartition(p1, p2, now)
				merged[id2] = true
			}
		}
	}

	// Remove merged partitions
	for id := range merged {
		delete(pm.partitions, id)
	}

	pm.lastMergeTime = now
}

// canMergePartitions checks if two partitions can be merged
func (pm *PartitionManager) canMergePartitions(p1, p2 *Partition) bool {
	// Check if any peer from p1 can communicate with any peer from p2
	for peer1 := range p1.Peers {
		for peer2 := range p2.Peers {
			if pm.arePeersConnected(peer1, peer2) {
				return true
			}
		}
	}
	return false
}

// mergePartition merges two partitions
func (pm *PartitionManager) mergePartition(p1, p2 *Partition, now time.Time) {
	// Merge peers
	for peer, lastSeen := range p2.Peers {
		p1.Peers[peer] = lastSeen
	}

	// Update timestamp
	p1.UpdatedAt = now
}

// GetPartitions returns all current partitions
func (pm *PartitionManager) GetPartitions() map[string]*Partition {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	partitions := make(map[string]*Partition)
	for id, p := range pm.partitions {
		partitions[id] = p
	}
	return partitions
}

// GetPartition returns a specific partition
func (pm *PartitionManager) GetPartition(id string) *Partition {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return pm.partitions[id]
}

// RemovePartition removes a partition
func (pm *PartitionManager) RemovePartition(id string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	delete(pm.partitions, id)
}
