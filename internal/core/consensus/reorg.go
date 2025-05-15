package consensus

import (
	"fmt"
	"sync"
)

// ReorgManager manages chain reorganizations
type ReorgManager struct {
	mu sync.RWMutex

	// Chain state
	mainChain       []string            // Block hashes in main chain
	sideChains      map[string][]string // Block hashes in side chains
	blockParents    map[string]string   // Block hash to parent hash mapping
	blockHeights    map[string]uint64   // Block hash to height mapping
	blockTimestamps map[string]int64    // Block hash to timestamp mapping
}

// NewReorgManager creates a new reorg manager
func NewReorgManager() *ReorgManager {
	return &ReorgManager{
		mainChain:       make([]string, 0),
		sideChains:      make(map[string][]string),
		blockParents:    make(map[string]string),
		blockHeights:    make(map[string]uint64),
		blockTimestamps: make(map[string]int64),
	}
}

// AddBlock adds a block to the chain
func (rm *ReorgManager) AddBlock(blockHash, parentHash string, height uint64, timestamp int64) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Check if block already exists
	if _, exists := rm.blockParents[blockHash]; exists {
		return fmt.Errorf("block %s already exists", blockHash)
	}

	// Add block to chain
	rm.blockParents[blockHash] = parentHash
	rm.blockHeights[blockHash] = height
	rm.blockTimestamps[blockHash] = timestamp

	// Check if block extends main chain
	if len(rm.mainChain) > 0 && parentHash == rm.mainChain[len(rm.mainChain)-1] {
		rm.mainChain = append(rm.mainChain, blockHash)
		return nil
	}

	// Check if block extends a side chain
	for chainHash, chain := range rm.sideChains {
		if len(chain) > 0 && parentHash == chain[len(chain)-1] {
			rm.sideChains[chainHash] = append(chain, blockHash)
			return nil
		}
	}

	// Create new side chain
	rm.sideChains[blockHash] = []string{blockHash}

	return nil
}

// HandleReorg handles a chain reorganization
func (rm *ReorgManager) HandleReorg(newBlockHash string) ([]string, []string, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Find common ancestor
	commonAncestor, err := rm.findCommonAncestor(newBlockHash)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find common ancestor: %v", err)
	}

	// Get blocks to remove and add
	blocksToRemove := rm.getBlocksToRemove(commonAncestor)
	blocksToAdd := rm.getBlocksToAdd(newBlockHash, commonAncestor)

	// Update main chain
	rm.updateMainChain(commonAncestor, newBlockHash)

	return blocksToRemove, blocksToAdd, nil
}

// findCommonAncestor finds the common ancestor of two chains
func (rm *ReorgManager) findCommonAncestor(newBlockHash string) (string, error) {
	// Get chain for new block
	newChain := rm.getChain(newBlockHash)
	if newChain == nil {
		return "", fmt.Errorf("block %s not found in any chain", newBlockHash)
	}

	// Find first common block
	for _, block := range newChain {
		if rm.isInMainChain(block) {
			return block, nil
		}
	}

	return "", fmt.Errorf("no common ancestor found")
}

// getBlocksToRemove gets blocks to remove from main chain
func (rm *ReorgManager) getBlocksToRemove(commonAncestor string) []string {
	var blocks []string
	for i := len(rm.mainChain) - 1; i >= 0; i-- {
		if rm.mainChain[i] == commonAncestor {
			break
		}
		blocks = append(blocks, rm.mainChain[i])
	}
	return blocks
}

// getBlocksToAdd gets blocks to add to main chain
func (rm *ReorgManager) getBlocksToAdd(newBlockHash, commonAncestor string) []string {
	var blocks []string
	chain := rm.getChain(newBlockHash)
	if chain == nil {
		return blocks
	}

	found := false
	for _, block := range chain {
		if block == commonAncestor {
			found = true
			continue
		}
		if found {
			blocks = append(blocks, block)
		}
	}
	return blocks
}

// updateMainChain updates the main chain
func (rm *ReorgManager) updateMainChain(commonAncestor, newBlockHash string) {
	// Remove blocks after common ancestor
	for i := len(rm.mainChain) - 1; i >= 0; i-- {
		if rm.mainChain[i] == commonAncestor {
			rm.mainChain = rm.mainChain[:i+1]
			break
		}
	}

	// Add new blocks
	chain := rm.getChain(newBlockHash)
	if chain != nil {
		found := false
		for _, block := range chain {
			if block == commonAncestor {
				found = true
				continue
			}
			if found {
				rm.mainChain = append(rm.mainChain, block)
			}
		}
	}
}

// getChain gets the chain containing a block
func (rm *ReorgManager) getChain(blockHash string) []string {
	// Check main chain
	for _, block := range rm.mainChain {
		if block == blockHash {
			return rm.mainChain
		}
	}

	// Check side chains
	for _, chain := range rm.sideChains {
		for _, block := range chain {
			if block == blockHash {
				return chain
			}
		}
	}

	return nil
}

// isInMainChain checks if a block is in the main chain
func (rm *ReorgManager) isInMainChain(blockHash string) bool {
	for _, block := range rm.mainChain {
		if block == blockHash {
			return true
		}
	}
	return false
}

// GetMainChain returns the main chain
func (rm *ReorgManager) GetMainChain() []string {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	chain := make([]string, len(rm.mainChain))
	copy(chain, rm.mainChain)
	return chain
}

// GetSideChains returns all side chains
func (rm *ReorgManager) GetSideChains() map[string][]string {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	chains := make(map[string][]string)
	for hash, chain := range rm.sideChains {
		chainCopy := make([]string, len(chain))
		copy(chainCopy, chain)
		chains[hash] = chainCopy
	}
	return chains
}

// GetBlockHeight returns the height of a block
func (rm *ReorgManager) GetBlockHeight(blockHash string) (uint64, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	height, exists := rm.blockHeights[blockHash]
	return height, exists
}

// GetBlockTimestamp returns the timestamp of a block
func (rm *ReorgManager) GetBlockTimestamp(blockHash string) (int64, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	timestamp, exists := rm.blockTimestamps[blockHash]
	return timestamp, exists
}
