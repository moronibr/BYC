package blockchain

import (
	"fmt"
	"sync"
	"time"
)

// CacheEntry represents a cache entry with expiration
type CacheEntry struct {
	Value      interface{}
	Expiration time.Time
}

// BlockchainCache handles caching of frequently accessed blockchain data
type BlockchainCache struct {
	// Cache for block headers
	blockHeaders map[string]*CacheEntry
	// Cache for transaction data
	transactions map[string]*CacheEntry
	// Cache for UTXO data
	utxos map[string]*CacheEntry
	// Cache for address balances
	balances map[string]*CacheEntry
	// Mutex for thread safety
	mu sync.RWMutex
	// Cache expiration duration
	expiration time.Duration
}

// NewBlockchainCache creates a new blockchain cache
func NewBlockchainCache() *BlockchainCache {
	cache := &BlockchainCache{
		blockHeaders: make(map[string]*CacheEntry),
		transactions: make(map[string]*CacheEntry),
		utxos:        make(map[string]*CacheEntry),
		balances:     make(map[string]*CacheEntry),
		expiration:   5 * time.Minute,
	}

	// Start cache cleanup goroutine
	go cache.cleanup()

	return cache
}

// GetBlockHeader retrieves a block header from cache
func (c *BlockchainCache) GetBlockHeader(hash string) (*Block, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if entry, ok := c.blockHeaders[hash]; ok {
		if time.Now().Before(entry.Expiration) {
			return entry.Value.(*Block), true
		}
		delete(c.blockHeaders, hash)
	}
	return nil, false
}

// SetBlockHeader adds a block header to cache
func (c *BlockchainCache) SetBlockHeader(hash string, block *Block) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.blockHeaders[hash] = &CacheEntry{
		Value:      block,
		Expiration: time.Now().Add(c.expiration),
	}
}

// GetTransaction retrieves a transaction from cache
func (c *BlockchainCache) GetTransaction(txID string) (*Transaction, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if entry, ok := c.transactions[txID]; ok {
		if time.Now().Before(entry.Expiration) {
			return entry.Value.(*Transaction), true
		}
		delete(c.transactions, txID)
	}
	return nil, false
}

// SetTransaction adds a transaction to cache
func (c *BlockchainCache) SetTransaction(txID string, tx *Transaction) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.transactions[txID] = &CacheEntry{
		Value:      tx,
		Expiration: time.Now().Add(c.expiration),
	}
}

// GetUTXO retrieves a UTXO from cache
func (c *BlockchainCache) GetUTXO(txID string, outputIndex int) (*UTXO, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := fmt.Sprintf("%s:%d", txID, outputIndex)
	if entry, ok := c.utxos[key]; ok {
		if time.Now().Before(entry.Expiration) {
			return entry.Value.(*UTXO), true
		}
		delete(c.utxos, key)
	}
	return nil, false
}

// SetUTXO adds a UTXO to cache
func (c *BlockchainCache) SetUTXO(txID string, outputIndex int, utxo *UTXO) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := fmt.Sprintf("%s:%d", txID, outputIndex)
	c.utxos[key] = &CacheEntry{
		Value:      utxo,
		Expiration: time.Now().Add(c.expiration),
	}
}

// GetBalance retrieves an address balance from cache
func (c *BlockchainCache) GetBalance(address string, coinType CoinType) (float64, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", address, coinType)
	if entry, ok := c.balances[key]; ok {
		if time.Now().Before(entry.Expiration) {
			return entry.Value.(float64), true
		}
		delete(c.balances, key)
	}
	return 0, false
}

// SetBalance adds an address balance to cache
func (c *BlockchainCache) SetBalance(address string, coinType CoinType, balance float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := fmt.Sprintf("%s:%s", address, coinType)
	c.balances[key] = &CacheEntry{
		Value:      balance,
		Expiration: time.Now().Add(c.expiration),
	}
}

// cleanup periodically removes expired cache entries
func (c *BlockchainCache) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()

		// Clean block headers
		for hash, entry := range c.blockHeaders {
			if now.After(entry.Expiration) {
				delete(c.blockHeaders, hash)
			}
		}

		// Clean transactions
		for txID, entry := range c.transactions {
			if now.After(entry.Expiration) {
				delete(c.transactions, txID)
			}
		}

		// Clean UTXOs
		for key, entry := range c.utxos {
			if now.After(entry.Expiration) {
				delete(c.utxos, key)
			}
		}

		// Clean balances
		for key, entry := range c.balances {
			if now.After(entry.Expiration) {
				delete(c.balances, key)
			}
		}

		c.mu.Unlock()
	}
}

// Clear clears all cache entries
func (c *BlockchainCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.blockHeaders = make(map[string]*CacheEntry)
	c.transactions = make(map[string]*CacheEntry)
	c.utxos = make(map[string]*CacheEntry)
	c.balances = make(map[string]*CacheEntry)
}
