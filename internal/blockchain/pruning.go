package blockchain

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"byc/internal/storage"
)

// PruningConfig holds configuration for block pruning
type PruningConfig struct {
	MinBlocksToKeep    int
	PruningInterval    time.Duration
	BatchSize          int
	CompressionEnabled bool
}

// PruningManager manages block pruning and UTXO optimization
type PruningManager struct {
	config     PruningConfig
	storage    *storage.Storage
	blockchain *Blockchain
	mu         sync.RWMutex
}

// UTXOCache implements a cache for UTXOs
type UTXOCache struct {
	cache   map[string]*UTXO
	maxSize int
	mu      sync.RWMutex
}

// NewPruningManager creates a new pruning manager
func NewPruningManager(config PruningConfig, storage *storage.Storage, blockchain *Blockchain) *PruningManager {
	return &PruningManager{
		config:     config,
		storage:    storage,
		blockchain: blockchain,
	}
}

// Start begins the pruning process
func (pm *PruningManager) Start() {
	go func() {
		ticker := time.NewTicker(pm.config.PruningInterval)
		defer ticker.Stop()

		for range ticker.C {
			if err := pm.PruneOldBlocks(); err != nil {
				fmt.Printf("Error pruning blocks: %v\n", err)
			}
		}
	}()
}

// PruneOldBlocks removes old blocks while maintaining the minimum required
func (pm *PruningManager) PruneOldBlocks() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Get current height
	currentHeight := len(pm.blockchain.GoldenBlocks) + len(pm.blockchain.SilverBlocks)
	if currentHeight <= pm.config.MinBlocksToKeep {
		return nil
	}

	// Calculate blocks to prune
	blocksToPrune := currentHeight - pm.config.MinBlocksToKeep

	// Prune blocks in batches
	for i := 0; i < blocksToPrune; i += pm.config.BatchSize {
		end := i + pm.config.BatchSize
		if end > blocksToPrune {
			end = blocksToPrune
		}

		if err := pm.pruneBlockBatch(i, end); err != nil {
			return fmt.Errorf("failed to prune block batch: %v", err)
		}
	}

	return nil
}

// pruneBlockBatch prunes a batch of blocks
func (pm *PruningManager) pruneBlockBatch(start, end int) error {
	for i := start; i < end; i++ {
		// Get block to prune
		var block Block
		if i < len(pm.blockchain.GoldenBlocks) {
			block = pm.blockchain.GoldenBlocks[i]
		} else {
			block = pm.blockchain.SilverBlocks[i-len(pm.blockchain.GoldenBlocks)]
		}

		// Remove block from memory
		if i < len(pm.blockchain.GoldenBlocks) {
			pm.blockchain.GoldenBlocks = append(pm.blockchain.GoldenBlocks[:i], pm.blockchain.GoldenBlocks[i+1:]...)
		} else {
			idx := i - len(pm.blockchain.GoldenBlocks)
			pm.blockchain.SilverBlocks = append(pm.blockchain.SilverBlocks[:idx], pm.blockchain.SilverBlocks[idx+1:]...)
		}

		// Update UTXO set
		for _, tx := range block.Transactions {
			for _, input := range tx.Inputs {
				pm.blockchain.UTXOSet.Remove(fmt.Sprintf("%s:%d", string(input.TxID), input.OutputIndex))
			}
		}
	}

	return nil
}

// OptimizeUTXOSet optimizes the UTXO set by removing spent outputs
func (pm *PruningManager) OptimizeUTXOSet() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Create new optimized UTXO set
	optimized := NewUTXOSet()

	// Process UTXOs in batches
	batchSize := pm.config.BatchSize
	utxos := pm.blockchain.UTXOSet.GetAll()

	for i := 0; i < len(utxos); i += batchSize {
		end := i + batchSize
		if end > len(utxos) {
			end = len(utxos)
		}

		if err := pm.processUTXOBatch(utxos[i:end], optimized); err != nil {
			return fmt.Errorf("failed to process UTXO batch: %v", err)
		}
	}

	// Update UTXO set
	pm.blockchain.UTXOSet = optimized
	return nil
}

// processUTXOBatch processes a batch of UTXOs
func (pm *PruningManager) processUTXOBatch(utxos []UTXO, optimized *UTXOSet) error {
	for _, utxo := range utxos {
		// Check if UTXO is spent
		spent := false
		for _, block := range pm.blockchain.GoldenBlocks {
			for _, tx := range block.Transactions {
				for _, input := range tx.Inputs {
					if string(input.TxID) == utxo.TxID && input.OutputIndex == utxo.Index {
						spent = true
						break
					}
				}
				if spent {
					break
				}
			}
			if spent {
				break
			}
		}

		if !spent {
			for _, block := range pm.blockchain.SilverBlocks {
				for _, tx := range block.Transactions {
					for _, input := range tx.Inputs {
						if string(input.TxID) == utxo.TxID && input.OutputIndex == utxo.Index {
							spent = true
							break
						}
					}
					if spent {
						break
					}
				}
				if spent {
					break
				}
			}
		}

		if !spent {
			optimized.Add(utxo)
		}
	}

	return nil
}

// NewUTXOCache creates a new UTXO cache
func NewUTXOCache(maxSize int) *UTXOCache {
	return &UTXOCache{
		cache:   make(map[string]*UTXO),
		maxSize: maxSize,
	}
}

// Get gets a UTXO from the cache
func (c *UTXOCache) Get(key string) (*UTXO, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	utxo, exists := c.cache[key]
	return utxo, exists
}

// Put puts a UTXO in the cache
func (c *UTXOCache) Put(key string, utxo *UTXO) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check cache size
	if len(c.cache) >= c.maxSize {
		// Remove oldest entry
		var oldestKey string
		var oldestTime int64
		for k, v := range c.cache {
			if oldestTime == 0 || v.Timestamp < oldestTime {
				oldestKey = k
				oldestTime = v.Timestamp
			}
		}
		delete(c.cache, oldestKey)
	}

	c.cache[key] = utxo
}

// Clear clears the cache
func (c *UTXOCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*UTXO)
}

// CompressBlock compresses a block for storage
func (pm *PruningManager) CompressBlock(block *Block) ([]byte, error) {
	if !pm.config.CompressionEnabled {
		return json.Marshal(block)
	}

	// TODO: Implement compression
	return json.Marshal(block)
}

// DecompressBlock decompresses a block from storage
func (pm *PruningManager) DecompressBlock(data []byte) (*Block, error) {
	if !pm.config.CompressionEnabled {
		var block Block
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, err
		}
		return &block, nil
	}

	// TODO: Implement decompression
	var block Block
	if err := json.Unmarshal(data, &block); err != nil {
		return nil, err
	}
	return &block, nil
}

// Serialize serializes a block
func (b *Block) Serialize() ([]byte, error) {
	var buf bytes.Buffer

	// Write block header
	if err := binary.Write(&buf, binary.LittleEndian, b.Timestamp); err != nil {
		return nil, err
	}
	if err := binary.Write(&buf, binary.LittleEndian, b.Nonce); err != nil {
		return nil, err
	}
	if err := binary.Write(&buf, binary.LittleEndian, b.Difficulty); err != nil {
		return nil, err
	}

	// Write block hash
	if _, err := buf.Write(b.Hash); err != nil {
		return nil, err
	}

	// Write previous block hash
	if _, err := buf.Write(b.PrevHash); err != nil {
		return nil, err
	}

	// Write block type
	if _, err := buf.Write([]byte(b.BlockType)); err != nil {
		return nil, err
	}

	// Write transactions
	if err := binary.Write(&buf, binary.LittleEndian, int32(len(b.Transactions))); err != nil {
		return nil, err
	}
	for _, tx := range b.Transactions {
		txData, err := tx.Serialize()
		if err != nil {
			return nil, err
		}
		if err := binary.Write(&buf, binary.LittleEndian, int32(len(txData))); err != nil {
			return nil, err
		}
		if _, err := buf.Write(txData); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// Deserialize deserializes a block
func (b *Block) Deserialize(data []byte) error {
	buf := bytes.NewReader(data)

	// Read block header
	if err := binary.Read(buf, binary.LittleEndian, &b.Timestamp); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &b.Nonce); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &b.Difficulty); err != nil {
		return err
	}

	// Read block hash
	b.Hash = make([]byte, 32)
	if _, err := io.ReadFull(buf, b.Hash); err != nil {
		return err
	}

	// Read previous block hash
	b.PrevHash = make([]byte, 32)
	if _, err := io.ReadFull(buf, b.PrevHash); err != nil {
		return err
	}

	// Read block type
	blockType := make([]byte, 8)
	if _, err := io.ReadFull(buf, blockType); err != nil {
		return err
	}
	b.BlockType = BlockType(blockType)

	// Read transactions
	var txCount int32
	if err := binary.Read(buf, binary.LittleEndian, &txCount); err != nil {
		return err
	}
	b.Transactions = make([]Transaction, txCount)
	for i := int32(0); i < txCount; i++ {
		var txSize int32
		if err := binary.Read(buf, binary.LittleEndian, &txSize); err != nil {
			return err
		}
		txData := make([]byte, txSize)
		if _, err := io.ReadFull(buf, txData); err != nil {
			return err
		}
		if err := b.Transactions[i].Deserialize(txData); err != nil {
			return err
		}
	}

	return nil
}

// Serialize serializes a transaction
func (tx *Transaction) Serialize() ([]byte, error) {
	var buf bytes.Buffer

	// Write transaction ID
	if _, err := buf.Write(tx.ID); err != nil {
		return nil, err
	}

	// Write timestamp
	if err := binary.Write(&buf, binary.LittleEndian, tx.Timestamp.Unix()); err != nil {
		return nil, err
	}

	// Write block type
	if _, err := buf.Write([]byte(tx.BlockType)); err != nil {
		return nil, err
	}

	// Write inputs
	if err := binary.Write(&buf, binary.LittleEndian, int32(len(tx.Inputs))); err != nil {
		return nil, err
	}
	for _, input := range tx.Inputs {
		if err := input.Serialize(&buf); err != nil {
			return nil, err
		}
	}

	// Write outputs
	if err := binary.Write(&buf, binary.LittleEndian, int32(len(tx.Outputs))); err != nil {
		return nil, err
	}
	for _, output := range tx.Outputs {
		if err := output.Serialize(&buf); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// Deserialize deserializes a transaction
func (tx *Transaction) Deserialize(data []byte) error {
	buf := bytes.NewReader(data)

	// Read transaction ID
	tx.ID = make([]byte, 32)
	if _, err := io.ReadFull(buf, tx.ID); err != nil {
		return err
	}

	// Read timestamp
	var timestamp int64
	if err := binary.Read(buf, binary.LittleEndian, &timestamp); err != nil {
		return err
	}
	tx.Timestamp = time.Unix(timestamp, 0)

	// Read block type
	blockType := make([]byte, 8)
	if _, err := io.ReadFull(buf, blockType); err != nil {
		return err
	}
	tx.BlockType = BlockType(blockType)

	// Read inputs
	var inputCount int32
	if err := binary.Read(buf, binary.LittleEndian, &inputCount); err != nil {
		return err
	}
	tx.Inputs = make([]TxInput, inputCount)
	for i := int32(0); i < inputCount; i++ {
		if err := tx.Inputs[i].Deserialize(buf); err != nil {
			return err
		}
	}

	// Read outputs
	var outputCount int32
	if err := binary.Read(buf, binary.LittleEndian, &outputCount); err != nil {
		return err
	}
	tx.Outputs = make([]TxOutput, outputCount)
	for i := int32(0); i < outputCount; i++ {
		if err := tx.Outputs[i].Deserialize(buf); err != nil {
			return err
		}
	}

	return nil
}

// Serialize serializes a transaction input
func (input *TxInput) Serialize(w io.Writer) error {
	// Write transaction ID
	if _, err := w.Write(input.TxID); err != nil {
		return err
	}

	// Write output index
	if err := binary.Write(w, binary.LittleEndian, input.OutputIndex); err != nil {
		return err
	}

	// Write amount
	if err := binary.Write(w, binary.LittleEndian, input.Amount); err != nil {
		return err
	}

	// Write signature
	if err := binary.Write(w, binary.LittleEndian, int32(len(input.Signature))); err != nil {
		return err
	}
	if _, err := w.Write(input.Signature); err != nil {
		return err
	}

	// Write public key
	if err := binary.Write(w, binary.LittleEndian, int32(len(input.PublicKey))); err != nil {
		return err
	}
	if _, err := w.Write(input.PublicKey); err != nil {
		return err
	}

	// Write address
	if err := binary.Write(w, binary.LittleEndian, int32(len(input.Address))); err != nil {
		return err
	}
	if _, err := w.Write([]byte(input.Address)); err != nil {
		return err
	}

	return nil
}

// Deserialize deserializes a transaction input
func (input *TxInput) Deserialize(r io.Reader) error {
	// Read transaction ID
	input.TxID = make([]byte, 32)
	if _, err := io.ReadFull(r, input.TxID); err != nil {
		return err
	}

	// Read output index
	if err := binary.Read(r, binary.LittleEndian, &input.OutputIndex); err != nil {
		return err
	}

	// Read amount
	if err := binary.Read(r, binary.LittleEndian, &input.Amount); err != nil {
		return err
	}

	// Read signature
	var sigLen int32
	if err := binary.Read(r, binary.LittleEndian, &sigLen); err != nil {
		return err
	}
	input.Signature = make([]byte, sigLen)
	if _, err := io.ReadFull(r, input.Signature); err != nil {
		return err
	}

	// Read public key
	var keyLen int32
	if err := binary.Read(r, binary.LittleEndian, &keyLen); err != nil {
		return err
	}
	input.PublicKey = make([]byte, keyLen)
	if _, err := io.ReadFull(r, input.PublicKey); err != nil {
		return err
	}

	// Read address
	var addrLen int32
	if err := binary.Read(r, binary.LittleEndian, &addrLen); err != nil {
		return err
	}
	addr := make([]byte, addrLen)
	if _, err := io.ReadFull(r, addr); err != nil {
		return err
	}
	input.Address = string(addr)

	return nil
}

// Serialize serializes a transaction output
func (output *TxOutput) Serialize(w io.Writer) error {
	// Write value
	if err := binary.Write(w, binary.LittleEndian, output.Value); err != nil {
		return err
	}

	// Write coin type
	if err := binary.Write(w, binary.LittleEndian, int32(len(output.CoinType))); err != nil {
		return err
	}
	if _, err := w.Write([]byte(output.CoinType)); err != nil {
		return err
	}

	// Write public key hash
	if err := binary.Write(w, binary.LittleEndian, int32(len(output.PublicKeyHash))); err != nil {
		return err
	}
	if _, err := w.Write(output.PublicKeyHash); err != nil {
		return err
	}

	// Write address
	if err := binary.Write(w, binary.LittleEndian, int32(len(output.Address))); err != nil {
		return err
	}
	if _, err := w.Write([]byte(output.Address)); err != nil {
		return err
	}

	return nil
}

// Deserialize deserializes a transaction output
func (output *TxOutput) Deserialize(r io.Reader) error {
	// Read value
	if err := binary.Read(r, binary.LittleEndian, &output.Value); err != nil {
		return err
	}

	// Read coin type
	var coinTypeLen int32
	if err := binary.Read(r, binary.LittleEndian, &coinTypeLen); err != nil {
		return err
	}
	coinType := make([]byte, coinTypeLen)
	if _, err := io.ReadFull(r, coinType); err != nil {
		return err
	}
	output.CoinType = CoinType(coinType)

	// Read public key hash
	var hashLen int32
	if err := binary.Read(r, binary.LittleEndian, &hashLen); err != nil {
		return err
	}
	output.PublicKeyHash = make([]byte, hashLen)
	if _, err := io.ReadFull(r, output.PublicKeyHash); err != nil {
		return err
	}

	// Read address
	var addrLen int32
	if err := binary.Read(r, binary.LittleEndian, &addrLen); err != nil {
		return err
	}
	addr := make([]byte, addrLen)
	if _, err := io.ReadFull(r, addr); err != nil {
		return err
	}
	output.Address = string(addr)

	return nil
}
