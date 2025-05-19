package storage

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/byc/internal/blockchain"
	"github.com/byc/internal/logger"
	"github.com/syndtr/goleveldb/leveldb"
	"go.uber.org/zap"
)

// Storage defines the interface for blockchain storage
type Storage interface {
	// Block operations
	PutBlock(block *blockchain.Block) error
	GetBlock(hash []byte) (*blockchain.Block, error)
	GetBlocks(blockType blockchain.BlockType) ([]*blockchain.Block, error)
	DeleteBlock(hash []byte) error

	// Transaction operations
	PutTransaction(tx *blockchain.Transaction) error
	GetTransaction(id []byte) (*blockchain.Transaction, error)
	GetTransactions(address string) ([]*blockchain.Transaction, error)
	DeleteTransaction(id []byte) error

	// UTXO operations
	PutUTXO(utxo *blockchain.UTXO) error
	GetUTXO(txID string, outputIndex int) (*blockchain.UTXO, error)
	GetUTXOs(address string) ([]*blockchain.UTXO, error)
	DeleteUTXO(txID string, outputIndex int) error

	// Close closes the storage
	Close() error
}

// LevelDBStorage implements Storage interface using LevelDB
type LevelDBStorage struct {
	db *leveldb.DB
}

// NewLevelDBStorage creates a new LevelDB storage
func NewLevelDBStorage(path string) (*LevelDBStorage, error) {
	// Create directory if it doesn't exist
	if err := filepath.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %v", err)
	}

	// Open LevelDB
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open LevelDB: %v", err)
	}

	return &LevelDBStorage{db: db}, nil
}

// PutBlock stores a block
func (s *LevelDBStorage) PutBlock(block *blockchain.Block) error {
	key := fmt.Sprintf("block:%x", block.Hash)
	data, err := json.Marshal(block)
	if err != nil {
		return fmt.Errorf("failed to marshal block: %v", err)
	}

	if err := s.db.Put([]byte(key), data, nil); err != nil {
		return fmt.Errorf("failed to store block: %v", err)
	}

	logger.Debug("Stored block", zap.String("hash", fmt.Sprintf("%x", block.Hash)))
	return nil
}

// GetBlock retrieves a block by its hash
func (s *LevelDBStorage) GetBlock(hash []byte) (*blockchain.Block, error) {
	key := fmt.Sprintf("block:%x", hash)
	data, err := s.db.Get([]byte(key), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get block: %v", err)
	}

	var block blockchain.Block
	if err := json.Unmarshal(data, &block); err != nil {
		return nil, fmt.Errorf("failed to unmarshal block: %v", err)
	}

	return &block, nil
}

// GetBlocks retrieves all blocks of a specific type
func (s *LevelDBStorage) GetBlocks(blockType blockchain.BlockType) ([]*blockchain.Block, error) {
	var blocks []*blockchain.Block
	iter := s.db.NewIterator(nil, nil)
	defer iter.Release()

	for iter.Next() {
		key := string(iter.Key())
		if len(key) < 6 || key[:6] != "block:" {
			continue
		}

		var block blockchain.Block
		if err := json.Unmarshal(iter.Value(), &block); err != nil {
			logger.Error("Failed to unmarshal block", zap.Error(err))
			continue
		}

		if block.BlockType == blockType {
			blocks = append(blocks, &block)
		}
	}

	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("failed to iterate blocks: %v", err)
	}

	return blocks, nil
}

// DeleteBlock deletes a block
func (s *LevelDBStorage) DeleteBlock(hash []byte) error {
	key := fmt.Sprintf("block:%x", hash)
	if err := s.db.Delete([]byte(key), nil); err != nil {
		return fmt.Errorf("failed to delete block: %v", err)
	}

	logger.Debug("Deleted block", zap.String("hash", fmt.Sprintf("%x", hash)))
	return nil
}

// PutTransaction stores a transaction
func (s *LevelDBStorage) PutTransaction(tx *blockchain.Transaction) error {
	key := fmt.Sprintf("tx:%x", tx.ID)
	data, err := json.Marshal(tx)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %v", err)
	}

	if err := s.db.Put([]byte(key), data, nil); err != nil {
		return fmt.Errorf("failed to store transaction: %v", err)
	}

	logger.Debug("Stored transaction", zap.String("id", fmt.Sprintf("%x", tx.ID)))
	return nil
}

// GetTransaction retrieves a transaction by its ID
func (s *LevelDBStorage) GetTransaction(id []byte) (*blockchain.Transaction, error) {
	key := fmt.Sprintf("tx:%x", id)
	data, err := s.db.Get([]byte(key), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get transaction: %v", err)
	}

	var tx blockchain.Transaction
	if err := json.Unmarshal(data, &tx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction: %v", err)
	}

	return &tx, nil
}

// GetTransactions retrieves all transactions for an address
func (s *LevelDBStorage) GetTransactions(address string) ([]*blockchain.Transaction, error) {
	var transactions []*blockchain.Transaction
	iter := s.db.NewIterator(nil, nil)
	defer iter.Release()

	for iter.Next() {
		key := string(iter.Key())
		if len(key) < 3 || key[:3] != "tx:" {
			continue
		}

		var tx blockchain.Transaction
		if err := json.Unmarshal(iter.Value(), &tx); err != nil {
			logger.Error("Failed to unmarshal transaction", zap.Error(err))
			continue
		}

		// Check if address is involved in transaction
		for _, input := range tx.Inputs {
			if string(input.PubKey) == address {
				transactions = append(transactions, &tx)
				break
			}
		}
		for _, output := range tx.Outputs {
			if output.Address == address {
				transactions = append(transactions, &tx)
				break
			}
		}
	}

	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("failed to iterate transactions: %v", err)
	}

	return transactions, nil
}

// DeleteTransaction deletes a transaction
func (s *LevelDBStorage) DeleteTransaction(id []byte) error {
	key := fmt.Sprintf("tx:%x", id)
	if err := s.db.Delete([]byte(key), nil); err != nil {
		return fmt.Errorf("failed to delete transaction: %v", err)
	}

	logger.Debug("Deleted transaction", zap.String("id", fmt.Sprintf("%x", id)))
	return nil
}

// PutUTXO stores a UTXO
func (s *LevelDBStorage) PutUTXO(utxo *blockchain.UTXO) error {
	key := fmt.Sprintf("utxo:%s:%d", utxo.TxID, utxo.OutputIndex)
	data, err := json.Marshal(utxo)
	if err != nil {
		return fmt.Errorf("failed to marshal UTXO: %v", err)
	}

	if err := s.db.Put([]byte(key), data, nil); err != nil {
		return fmt.Errorf("failed to store UTXO: %v", err)
	}

	logger.Debug("Stored UTXO", zap.String("txid", utxo.TxID), zap.Int("outputIndex", utxo.OutputIndex))
	return nil
}

// GetUTXO retrieves a UTXO by its transaction ID and output index
func (s *LevelDBStorage) GetUTXO(txID string, outputIndex int) (*blockchain.UTXO, error) {
	key := fmt.Sprintf("utxo:%s:%d", txID, outputIndex)
	data, err := s.db.Get([]byte(key), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get UTXO: %v", err)
	}

	var utxo blockchain.UTXO
	if err := json.Unmarshal(data, &utxo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal UTXO: %v", err)
	}

	return &utxo, nil
}

// GetUTXOs retrieves all UTXOs for an address
func (s *LevelDBStorage) GetUTXOs(address string) ([]*blockchain.UTXO, error) {
	var utxos []*blockchain.UTXO
	iter := s.db.NewIterator(nil, nil)
	defer iter.Release()

	for iter.Next() {
		key := string(iter.Key())
		if len(key) < 5 || key[:5] != "utxo:" {
			continue
		}

		var utxo blockchain.UTXO
		if err := json.Unmarshal(iter.Value(), &utxo); err != nil {
			logger.Error("Failed to unmarshal UTXO", zap.Error(err))
			continue
		}

		if utxo.Address == address {
			utxos = append(utxos, &utxo)
		}
	}

	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("failed to iterate UTXOs: %v", err)
	}

	return utxos, nil
}

// DeleteUTXO deletes a UTXO
func (s *LevelDBStorage) DeleteUTXO(txID string, outputIndex int) error {
	key := fmt.Sprintf("utxo:%s:%d", txID, outputIndex)
	if err := s.db.Delete([]byte(key), nil); err != nil {
		return fmt.Errorf("failed to delete UTXO: %v", err)
	}

	logger.Debug("Deleted UTXO", zap.String("txid", txID), zap.Int("outputIndex", outputIndex))
	return nil
}

// Close closes the storage
func (s *LevelDBStorage) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("failed to close LevelDB: %v", err)
	}
	return nil
}
