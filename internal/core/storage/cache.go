package storage

import (
	"sync"
	"time"
)

// Cache defines the interface for blockchain data caching
type Cache interface {
	// Block operations
	GetBlock(hash [32]byte) (*Block, bool)
	SetBlock(block *Block)
	DeleteBlock(hash [32]byte)

	// Transaction operations
	GetTransaction(hash [32]byte) (*Transaction, bool)
	SetTransaction(tx *Transaction)
	DeleteTransaction(hash [32]byte)

	// UTXO operations
	GetUTXO(txHash [32]byte, outputIndex uint32) (*DBUTXO, bool)
	SetUTXO(txHash [32]byte, outputIndex uint32, utxo *DBUTXO)
	DeleteUTXO(txHash [32]byte, outputIndex uint32)

	// Chain state operations
	GetChainState() (*ChainState, bool)
	SetChainState(state *ChainState)

	// Maintenance operations
	Clear()
	Compact()
}

// LRUCache implements a least-recently-used cache
type LRUCache struct {
	mu sync.RWMutex

	// Configuration
	maxBlocks       int
	maxTransactions int
	maxUTXOs        int
	ttl             time.Duration

	// Block cache
	blockCache map[[32]byte]*blockEntry
	blockLRU   *lruList

	// Transaction cache
	txCache map[[32]byte]*txEntry
	txLRU   *lruList

	// UTXO cache
	utxoCache map[string]*utxoEntry
	utxoLRU   *lruList

	// Chain state cache
	chainState *chainStateEntry
}

// Entry types for different cache items
type blockEntry struct {
	value      *Block
	timestamp  time.Time
	prev, next *blockEntry
}

type txEntry struct {
	value      *Transaction
	timestamp  time.Time
	prev, next *txEntry
}

type utxoEntry struct {
	value      *DBUTXO
	timestamp  time.Time
	prev, next *utxoEntry
}

type chainStateEntry struct {
	value     *ChainState
	timestamp time.Time
}

// lruList represents a doubly-linked list for LRU tracking
type lruList struct {
	head, tail interface{}
}

// NewLRUCache creates a new LRU cache
func NewLRUCache(maxBlocks, maxTransactions, maxUTXOs int, ttl time.Duration) *LRUCache {
	return &LRUCache{
		maxBlocks:       maxBlocks,
		maxTransactions: maxTransactions,
		maxUTXOs:        maxUTXOs,
		ttl:             ttl,
		blockCache:      make(map[[32]byte]*blockEntry),
		blockLRU:        &lruList{},
		txCache:         make(map[[32]byte]*txEntry),
		txLRU:           &lruList{},
		utxoCache:       make(map[string]*utxoEntry),
		utxoLRU:         &lruList{},
	}
}

// GetBlock retrieves a block from the cache
func (c *LRUCache) GetBlock(hash [32]byte) (*Block, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.blockCache[hash]
	if !exists {
		return nil, false
	}

	if time.Since(entry.timestamp) > c.ttl {
		delete(c.blockCache, hash)
		return nil, false
	}

	c.moveBlockToFront(entry)
	return entry.value, true
}

// SetBlock adds a block to the cache
func (c *LRUCache) SetBlock(block *Block) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := &blockEntry{
		value:     block,
		timestamp: time.Now(),
	}

	if len(c.blockCache) >= c.maxBlocks {
		c.evictOldestBlock()
	}

	c.blockCache[block.Hash] = entry
	c.addBlockToFront(entry)
}

// DeleteBlock removes a block from the cache
func (c *LRUCache) DeleteBlock(hash [32]byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, exists := c.blockCache[hash]; exists {
		c.removeBlock(entry)
		delete(c.blockCache, hash)
	}
}

// GetTransaction retrieves a transaction from the cache
func (c *LRUCache) GetTransaction(hash [32]byte) (*Transaction, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.txCache[hash]
	if !exists {
		return nil, false
	}

	if time.Since(entry.timestamp) > c.ttl {
		delete(c.txCache, hash)
		return nil, false
	}

	c.moveTxToFront(entry)
	return entry.value, true
}

// SetTransaction adds a transaction to the cache
func (c *LRUCache) SetTransaction(tx *Transaction) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := &txEntry{
		value:     tx,
		timestamp: time.Now(),
	}

	if len(c.txCache) >= c.maxTransactions {
		c.evictOldestTransaction()
	}

	c.txCache[tx.Hash] = entry
	c.addTxToFront(entry)
}

// DeleteTransaction removes a transaction from the cache
func (c *LRUCache) DeleteTransaction(hash [32]byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, exists := c.txCache[hash]; exists {
		c.removeTx(entry)
		delete(c.txCache, hash)
	}
}

// GetUTXO retrieves a UTXO from the cache
func (c *LRUCache) GetUTXO(txHash [32]byte, outputIndex uint32) (*DBUTXO, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := utxoKey(txHash, outputIndex)
	entry, exists := c.utxoCache[key]
	if !exists {
		return nil, false
	}

	if time.Since(entry.timestamp) > c.ttl {
		delete(c.utxoCache, key)
		return nil, false
	}

	c.moveUTXOToFront(entry)
	return entry.value, true
}

// SetUTXO adds a UTXO to the cache
func (c *LRUCache) SetUTXO(txHash [32]byte, outputIndex uint32, utxo *DBUTXO) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := &utxoEntry{
		value:     utxo,
		timestamp: time.Now(),
	}

	if len(c.utxoCache) >= c.maxUTXOs {
		c.evictOldestUTXO()
	}

	key := utxoKey(txHash, outputIndex)
	c.utxoCache[key] = entry
	c.addUTXOToFront(entry)
}

// DeleteUTXO removes a UTXO from the cache
func (c *LRUCache) DeleteUTXO(txHash [32]byte, outputIndex uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := utxoKey(txHash, outputIndex)
	if entry, exists := c.utxoCache[key]; exists {
		c.removeUTXO(entry)
		delete(c.utxoCache, key)
	}
}

// GetChainState retrieves the chain state from the cache
func (c *LRUCache) GetChainState() (*ChainState, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.chainState == nil {
		return nil, false
	}

	if time.Since(c.chainState.timestamp) > c.ttl {
		c.chainState = nil
		return nil, false
	}

	return c.chainState.value, true
}

// SetChainState adds the chain state to the cache
func (c *LRUCache) SetChainState(state *ChainState) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.chainState = &chainStateEntry{
		value:     state,
		timestamp: time.Now(),
	}
}

// Clear removes all items from the cache
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.blockCache = make(map[[32]byte]*blockEntry)
	c.blockLRU = &lruList{}
	c.txCache = make(map[[32]byte]*txEntry)
	c.txLRU = &lruList{}
	c.utxoCache = make(map[string]*utxoEntry)
	c.utxoLRU = &lruList{}
	c.chainState = nil
}

// Compact removes expired items from the cache
func (c *LRUCache) Compact() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	// Compact block cache
	for hash, entry := range c.blockCache {
		if now.Sub(entry.timestamp) > c.ttl {
			c.removeBlock(entry)
			delete(c.blockCache, hash)
		}
	}

	// Compact transaction cache
	for hash, entry := range c.txCache {
		if now.Sub(entry.timestamp) > c.ttl {
			c.removeTx(entry)
			delete(c.txCache, hash)
		}
	}

	// Compact UTXO cache
	for key, entry := range c.utxoCache {
		if now.Sub(entry.timestamp) > c.ttl {
			c.removeUTXO(entry)
			delete(c.utxoCache, key)
		}
	}

	// Compact chain state
	if c.chainState != nil && now.Sub(c.chainState.timestamp) > c.ttl {
		c.chainState = nil
	}
}

// Helper functions

func (c *LRUCache) evictOldestBlock() {
	if c.blockLRU.tail != nil {
		entry := c.blockLRU.tail.(*blockEntry)
		c.removeBlock(entry)
		delete(c.blockCache, entry.value.Hash)
	}
}

func (c *LRUCache) evictOldestTransaction() {
	if c.txLRU.tail != nil {
		entry := c.txLRU.tail.(*txEntry)
		c.removeTx(entry)
		delete(c.txCache, entry.value.Hash)
	}
}

func (c *LRUCache) evictOldestUTXO() {
	if c.utxoLRU.tail != nil {
		entry := c.utxoLRU.tail.(*utxoEntry)
		c.removeUTXO(entry)
		// Note: We can't easily get the key from the entry, so we need to search
		for key, e := range c.utxoCache {
			if e == entry {
				delete(c.utxoCache, key)
				break
			}
		}
	}
}

func utxoKey(txHash [32]byte, outputIndex uint32) string {
	return string(txHash[:]) + string([]byte{byte(outputIndex >> 24), byte(outputIndex >> 16), byte(outputIndex >> 8), byte(outputIndex)})
}

// Block LRU operations

func (c *LRUCache) addBlockToFront(entry *blockEntry) {
	if c.blockLRU.head == nil {
		c.blockLRU.head = entry
		c.blockLRU.tail = entry
		return
	}

	entry.next = c.blockLRU.head.(*blockEntry)
	c.blockLRU.head.(*blockEntry).prev = entry
	c.blockLRU.head = entry
}

func (c *LRUCache) removeBlock(entry *blockEntry) {
	if entry.prev != nil {
		entry.prev.next = entry.next
	} else {
		c.blockLRU.head = entry.next
	}

	if entry.next != nil {
		entry.next.prev = entry.prev
	} else {
		c.blockLRU.tail = entry.prev
	}

	entry.prev = nil
	entry.next = nil
}

func (c *LRUCache) moveBlockToFront(entry *blockEntry) {
	if entry == c.blockLRU.head {
		return
	}

	c.removeBlock(entry)
	c.addBlockToFront(entry)
}

// Transaction LRU operations

func (c *LRUCache) addTxToFront(entry *txEntry) {
	if c.txLRU.head == nil {
		c.txLRU.head = entry
		c.txLRU.tail = entry
		return
	}

	entry.next = c.txLRU.head.(*txEntry)
	c.txLRU.head.(*txEntry).prev = entry
	c.txLRU.head = entry
}

func (c *LRUCache) removeTx(entry *txEntry) {
	if entry.prev != nil {
		entry.prev.next = entry.next
	} else {
		c.txLRU.head = entry.next
	}

	if entry.next != nil {
		entry.next.prev = entry.prev
	} else {
		c.txLRU.tail = entry.prev
	}

	entry.prev = nil
	entry.next = nil
}

func (c *LRUCache) moveTxToFront(entry *txEntry) {
	if entry == c.txLRU.head {
		return
	}

	c.removeTx(entry)
	c.addTxToFront(entry)
}

// UTXO LRU operations

func (c *LRUCache) addUTXOToFront(entry *utxoEntry) {
	if c.utxoLRU.head == nil {
		c.utxoLRU.head = entry
		c.utxoLRU.tail = entry
		return
	}

	entry.next = c.utxoLRU.head.(*utxoEntry)
	c.utxoLRU.head.(*utxoEntry).prev = entry
	c.utxoLRU.head = entry
}

func (c *LRUCache) removeUTXO(entry *utxoEntry) {
	if entry.prev != nil {
		entry.prev.next = entry.next
	} else {
		c.utxoLRU.head = entry.next
	}

	if entry.next != nil {
		entry.next.prev = entry.prev
	} else {
		c.utxoLRU.tail = entry.prev
	}

	entry.prev = nil
	entry.next = nil
}

func (c *LRUCache) moveUTXOToFront(entry *utxoEntry) {
	if entry == c.utxoLRU.head {
		return
	}

	c.removeUTXO(entry)
	c.addUTXOToFront(entry)
}

// LRUCache implements the Cache interface
var _ Cache = (*LRUCache)(nil)
