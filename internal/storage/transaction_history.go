package storage

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/youngchain/internal/core/coin"
	"go.etcd.io/bbolt"
)

const (
	historyBucket = "transaction_history"
)

// TransactionHistory represents a transaction history entry
type TransactionHistory struct {
	TxHash    []byte
	Address   string
	Amount    int64 // Positive for received, negative for sent
	Timestamp int64
	CoinType  coin.CoinType
	Status    string // "pending", "confirmed", "failed"
	BlockHash []byte
}

// TransactionHistoryStore manages transaction history
type TransactionHistoryStore struct {
	db *bbolt.DB
}

// NewTransactionHistoryStore creates a new transaction history store
func NewTransactionHistoryStore(db *bbolt.DB) *TransactionHistoryStore {
	return &TransactionHistoryStore{db: db}
}

// AddHistory adds a transaction to the history
func (ths *TransactionHistoryStore) AddHistory(history *TransactionHistory) error {
	return ths.db.Update(func(dbTx *bbolt.Tx) error {
		// Get or create history bucket
		bucket, err := dbTx.CreateBucketIfNotExists([]byte(historyBucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		// Create key: address + timestamp
		key := fmt.Sprintf("%s:%d", history.Address, history.Timestamp)

		// Marshal history
		data, err := json.Marshal(history)
		if err != nil {
			return fmt.Errorf("marshal history: %s", err)
		}

		// Save history
		if err := bucket.Put([]byte(key), data); err != nil {
			return fmt.Errorf("put history: %s", err)
		}

		return nil
	})
}

// GetHistory gets transaction history for an address
func (ths *TransactionHistoryStore) GetHistory(address string, coinType coin.CoinType) ([]*TransactionHistory, error) {
	var history []*TransactionHistory

	err := ths.db.View(func(dbTx *bbolt.Tx) error {
		// Get history bucket
		bucket := dbTx.Bucket([]byte(historyBucket))
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}

		// Iterate over all history entries
		return bucket.ForEach(func(k, v []byte) error {
			var entry TransactionHistory
			if err := json.Unmarshal(v, &entry); err != nil {
				return err
			}

			if entry.Address == address && entry.CoinType == coinType {
				history = append(history, &entry)
			}

			return nil
		})
	})

	return history, err
}

// GetHistoryByTimeRange gets transaction history for an address within a time range
func (ths *TransactionHistoryStore) GetHistoryByTimeRange(address string, coinType coin.CoinType, start, end time.Time) ([]*TransactionHistory, error) {
	var history []*TransactionHistory

	err := ths.db.View(func(dbTx *bbolt.Tx) error {
		// Get history bucket
		bucket := dbTx.Bucket([]byte(historyBucket))
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}

		// Iterate over all history entries
		return bucket.ForEach(func(k, v []byte) error {
			var entry TransactionHistory
			if err := json.Unmarshal(v, &entry); err != nil {
				return err
			}

			timestamp := time.Unix(entry.Timestamp, 0)
			if entry.Address == address &&
				entry.CoinType == coinType &&
				timestamp.After(start) &&
				timestamp.Before(end) {
				history = append(history, &entry)
			}

			return nil
		})
	})

	return history, err
}

// UpdateStatus updates the status of a transaction
func (ths *TransactionHistoryStore) UpdateStatus(txHash []byte, status string, blockHash []byte) error {
	return ths.db.Update(func(dbTx *bbolt.Tx) error {
		// Get history bucket
		bucket := dbTx.Bucket([]byte(historyBucket))
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}

		// Iterate over all history entries
		return bucket.ForEach(func(k, v []byte) error {
			var entry TransactionHistory
			if err := json.Unmarshal(v, &entry); err != nil {
				return err
			}

			if string(entry.TxHash) == string(txHash) {
				entry.Status = status
				entry.BlockHash = blockHash

				// Marshal updated entry
				data, err := json.Marshal(entry)
				if err != nil {
					return fmt.Errorf("marshal history: %s", err)
				}

				// Save updated entry
				if err := bucket.Put(k, data); err != nil {
					return fmt.Errorf("put history: %s", err)
				}
			}

			return nil
		})
	})
}

// GetPendingTransactions gets all pending transactions
func (ths *TransactionHistoryStore) GetPendingTransactions() ([]*TransactionHistory, error) {
	var pending []*TransactionHistory

	err := ths.db.View(func(dbTx *bbolt.Tx) error {
		// Get history bucket
		bucket := dbTx.Bucket([]byte(historyBucket))
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}

		// Iterate over all history entries
		return bucket.ForEach(func(k, v []byte) error {
			var entry TransactionHistory
			if err := json.Unmarshal(v, &entry); err != nil {
				return err
			}

			if entry.Status == "pending" {
				pending = append(pending, &entry)
			}

			return nil
		})
	})

	return pending, err
}
