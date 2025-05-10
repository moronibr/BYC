package storage

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/youngchain/internal/core/coin"
	"github.com/youngchain/internal/core/common"
	"github.com/youngchain/internal/core/types"
	"go.etcd.io/bbolt"
)

const (
	transactionBucket = "transactions"
	utxoBucket        = "utxos"
)

// TransactionStore handles storage of transactions
type TransactionStore struct {
	db StorageInterface
}

// NewTransactionStore creates a new transaction store
func NewTransactionStore(db StorageInterface) *TransactionStore {
	return &TransactionStore{
		db: db,
	}
}

// SaveTransaction saves a transaction to storage
func (ts *TransactionStore) SaveTransaction(tx *common.Transaction) error {
	if tx == nil {
		return errors.New("cannot save nil transaction")
	}

	data, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	return ts.db.Put(tx.Hash(), data)
}

// GetTransaction retrieves a transaction by its hash
func (ts *TransactionStore) GetTransaction(txHash []byte) (*common.Transaction, error) {
	data, err := ts.db.Get(txHash)
	if err != nil {
		return nil, err
	}

	var tx common.Transaction
	if err := json.Unmarshal(data, &tx); err != nil {
		return nil, err
	}

	return &tx, nil
}

// SaveUTXO saves a UTXO to the database
func (ts *TransactionStore) SaveUTXO(utxo *types.UTXO) error {
	return ts.db.Update(func(dbTx *bbolt.Tx) error {
		// Get or create UTXO bucket
		bucket, err := dbTx.CreateBucketIfNotExists([]byte(utxoBucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		// Marshal UTXO
		data, err := json.Marshal(utxo)
		if err != nil {
			return fmt.Errorf("marshal UTXO: %s", err)
		}

		// Create key
		key := make([]byte, 36)
		copy(key[:32], utxo.TxHash)
		binary.LittleEndian.PutUint32(key[32:], utxo.TxIndex)

		// Save UTXO
		if err := bucket.Put(key, data); err != nil {
			return fmt.Errorf("put UTXO: %s", err)
		}

		return nil
	})
}

// GetUTXO gets a UTXO from the database
func (ts *TransactionStore) GetUTXO(txHash []byte, index uint32) (*types.UTXO, error) {
	var utxo *types.UTXO

	err := ts.db.View(func(dbTx *bbolt.Tx) error {
		// Get UTXO bucket
		bucket := dbTx.Bucket([]byte(utxoBucket))
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}

		// Create key
		key := make([]byte, 36)
		copy(key[:32], txHash)
		binary.LittleEndian.PutUint32(key[32:], index)

		// Get UTXO data
		data := bucket.Get(key)
		if data == nil {
			return fmt.Errorf("UTXO not found")
		}

		// Unmarshal UTXO
		if err := json.Unmarshal(data, &utxo); err != nil {
			return fmt.Errorf("unmarshal UTXO: %s", err)
		}

		return nil
	})

	return utxo, err
}

// DeleteUTXO deletes a UTXO from the database
func (ts *TransactionStore) DeleteUTXO(txHash []byte, index uint32) error {
	return ts.db.Update(func(dbTx *bbolt.Tx) error {
		// Get UTXO bucket
		bucket := dbTx.Bucket([]byte(utxoBucket))
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}

		// Create key
		key := make([]byte, 36)
		copy(key[:32], txHash)
		binary.LittleEndian.PutUint32(key[32:], index)

		// Delete UTXO
		if err := bucket.Delete(key); err != nil {
			return fmt.Errorf("delete UTXO: %s", err)
		}

		return nil
	})
}

// GetUTXOsByAddress gets all UTXOs for an address
func (ts *TransactionStore) GetUTXOsByAddress(address string) ([]*types.UTXO, error) {
	var utxos []*types.UTXO

	err := ts.db.View(func(dbTx *bbolt.Tx) error {
		// Get UTXO bucket
		bucket := dbTx.Bucket([]byte(utxoBucket))
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}

		// Iterate over all UTXOs
		return bucket.ForEach(func(k, v []byte) error {
			var utxo types.UTXO
			if err := json.Unmarshal(v, &utxo); err != nil {
				return err
			}

			if utxo.Address == address && !utxo.IsSpent {
				utxos = append(utxos, &utxo)
			}

			return nil
		})
	})

	return utxos, err
}

// GetBalance gets the balance for an address
func (ts *TransactionStore) GetBalance(address string, coinType coin.CoinType) (uint64, error) {
	var balance uint64

	err := ts.db.View(func(dbTx *bbolt.Tx) error {
		// Get UTXO bucket
		bucket := dbTx.Bucket([]byte(utxoBucket))
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}

		// Iterate over all UTXOs
		return bucket.ForEach(func(k, v []byte) error {
			var utxo types.UTXO
			if err := json.Unmarshal(v, &utxo); err != nil {
				return err
			}

			if utxo.Address == address && utxo.CoinType == coinType && !utxo.IsSpent {
				balance += utxo.Value
			}

			return nil
		})
	})

	return balance, err
}

// calculateHash calculates the hash of a transaction
func calculateHash(tx *common.Transaction) []byte {
	return tx.Hash()
}
