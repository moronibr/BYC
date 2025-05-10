package storage

import (
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/common"
	"go.etcd.io/bbolt"
)

// DB represents the blockchain database
type DB struct {
	db *bbolt.DB
}

// NewDB creates a new database instance
func NewDB(dbPath string) (*DB, error) {
	db, err := bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Create necessary buckets
	err = db.Update(func(tx *bbolt.Tx) error {
		// Blocks bucket
		if _, err := tx.CreateBucketIfNotExists([]byte("blocks")); err != nil {
			return fmt.Errorf("failed to create blocks bucket: %v", err)
		}

		// Chain state bucket
		if _, err := tx.CreateBucketIfNotExists([]byte("chainstate")); err != nil {
			return fmt.Errorf("failed to create chainstate bucket: %v", err)
		}

		// UTXO set bucket
		if _, err := tx.CreateBucketIfNotExists([]byte("utxo")); err != nil {
			return fmt.Errorf("failed to create utxo bucket: %v", err)
		}

		// Mempool bucket
		if _, err := tx.CreateBucketIfNotExists([]byte("mempool")); err != nil {
			return fmt.Errorf("failed to create mempool bucket: %v", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &DB{db: db}, nil
}

// Close closes the database
func (db *DB) Close() error {
	return db.db.Close()
}

// SaveBlock saves a block to the database
func (db *DB) SaveBlock(block *block.Block) error {
	return db.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("blocks"))
		if bucket == nil {
			return fmt.Errorf("blocks bucket not found")
		}

		// Serialize block
		data, err := json.Marshal(block)
		if err != nil {
			return fmt.Errorf("failed to serialize block: %v", err)
		}

		// Use block hash as key
		return bucket.Put(block.Header.Hash, data)
	})
}

// GetBlock retrieves a block by hash
func (db *DB) GetBlock(hash []byte) (*block.Block, error) {
	var block block.Block

	err := db.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("blocks"))
		if bucket == nil {
			return fmt.Errorf("blocks bucket not found")
		}

		data := bucket.Get(hash)
		if data == nil {
			return fmt.Errorf("block not found")
		}

		return json.Unmarshal(data, &block)
	})

	if err != nil {
		return nil, err
	}

	return &block, nil
}

// SaveUTXO saves a UTXO to the database
func (db *DB) SaveUTXO(txHash []byte, outputIndex uint32, utxo *block.UTXO) error {
	return db.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("utxo"))
		if bucket == nil {
			return fmt.Errorf("utxo bucket not found")
		}

		// Create composite key: txHash + outputIndex
		key := make([]byte, len(txHash)+4)
		copy(key, txHash)
		binary.LittleEndian.PutUint32(key[len(txHash):], outputIndex)

		// Serialize UTXO
		data, err := json.Marshal(utxo)
		if err != nil {
			return fmt.Errorf("failed to serialize UTXO: %v", err)
		}

		return bucket.Put(key, data)
	})
}

// GetUTXO retrieves a UTXO by transaction hash and output index
func (db *DB) GetUTXO(txHash []byte, outputIndex uint32) (*block.UTXO, error) {
	var utxo block.UTXO

	err := db.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("utxo"))
		if bucket == nil {
			return fmt.Errorf("utxo bucket not found")
		}

		// Create composite key
		key := make([]byte, len(txHash)+4)
		copy(key, txHash)
		binary.LittleEndian.PutUint32(key[len(txHash):], outputIndex)

		data := bucket.Get(key)
		if data == nil {
			return fmt.Errorf("UTXO not found")
		}

		return json.Unmarshal(data, &utxo)
	})

	if err != nil {
		return nil, err
	}

	return &utxo, nil
}

// DeleteUTXO removes a UTXO from the database
func (db *DB) DeleteUTXO(txHash []byte, outputIndex uint32) error {
	return db.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("utxo"))
		if bucket == nil {
			return fmt.Errorf("utxo bucket not found")
		}

		// Create composite key
		key := make([]byte, len(txHash)+4)
		copy(key, txHash)
		binary.LittleEndian.PutUint32(key[len(txHash):], outputIndex)

		return bucket.Delete(key)
	})
}

// SaveMempoolTx saves a transaction to the mempool
func (db *DB) SaveMempoolTx(transaction *common.Transaction) error {
	return db.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("mempool"))
		if bucket == nil {
			return fmt.Errorf("mempool bucket not found")
		}

		// Serialize transaction
		data, err := json.Marshal(transaction)
		if err != nil {
			return err
		}

		// Save transaction
		return bucket.Put(transaction.Hash(), data)
	})
}

// GetMempoolTx retrieves a transaction from the mempool
func (db *DB) GetMempoolTx(txHash []byte) (*common.Transaction, error) {
	var tx common.Transaction
	err := db.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("mempool"))
		if bucket == nil {
			return fmt.Errorf("mempool bucket not found")
		}

		// Get transaction data
		data := bucket.Get(txHash)
		if data == nil {
			return fmt.Errorf("transaction not found")
		}

		// Deserialize transaction
		return json.Unmarshal(data, &tx)
	})

	if err != nil {
		return nil, err
	}

	return &tx, nil
}

// GetAllMempoolTxs retrieves all transactions from the mempool
func (db *DB) GetAllMempoolTxs() ([]*common.Transaction, error) {
	var transactions []*common.Transaction
	err := db.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("mempool"))
		if bucket == nil {
			return fmt.Errorf("mempool bucket not found")
		}

		// Iterate over all transactions
		return bucket.ForEach(func(k, v []byte) error {
			var tx common.Transaction
			if err := json.Unmarshal(v, &tx); err != nil {
				return err
			}
			transactions = append(transactions, &tx)
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return transactions, nil
}

// SaveChainState saves the current chain state
func (db *DB) SaveChainState(height uint64, bestBlockHash []byte) error {
	return db.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("chainstate"))
		if bucket == nil {
			return fmt.Errorf("chainstate bucket not found")
		}

		// Save height
		heightBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(heightBytes, height)
		if err := bucket.Put([]byte("height"), heightBytes); err != nil {
			return err
		}

		// Save best block hash
		return bucket.Put([]byte("bestblock"), bestBlockHash)
	})
}

// GetChainState retrieves the current chain state
func (db *DB) GetChainState() (uint64, []byte, error) {
	var height uint64
	var bestBlockHash []byte

	err := db.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("chainstate"))
		if bucket == nil {
			return fmt.Errorf("chainstate bucket not found")
		}

		// Get height
		heightBytes := bucket.Get([]byte("height"))
		if heightBytes == nil {
			return fmt.Errorf("chain height not found")
		}
		height = binary.LittleEndian.Uint64(heightBytes)

		// Get best block hash
		bestBlockHash = bucket.Get([]byte("bestblock"))
		if bestBlockHash == nil {
			return fmt.Errorf("best block hash not found")
		}

		return nil
	})

	if err != nil {
		return 0, nil, err
	}

	return height, bestBlockHash, nil
}

// StoreTransaction stores a transaction in the database
func (db *DB) StoreTransaction(blockTx *common.Transaction) error {
	return db.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("transactions"))
		if bucket == nil {
			return fmt.Errorf("transactions bucket not found")
		}

		// Serialize transaction
		data, err := json.Marshal(blockTx)
		if err != nil {
			return err
		}

		// Store transaction
		return bucket.Put(blockTx.Hash(), data)
	})
}
