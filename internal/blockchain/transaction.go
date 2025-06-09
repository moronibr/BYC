package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"byc/internal/crypto"
)

// TransactionError represents a transaction-related error
type TransactionError struct {
	Operation string
	Reason    string
	Details   map[string]interface{}
}

func (e *TransactionError) Error() string {
	return fmt.Sprintf("transaction error during %s: %s", e.Operation, e.Reason)
}

// ValidationError represents a transaction validation error
type ValidationError struct {
	Field   string
	Reason  string
	Details map[string]interface{}
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error in field %s: %s", e.Field, e.Reason)
}

// Validate validates a transaction with improved error handling
func (tx *Transaction) Validate(utxoSet *UTXOSet) error {
	// Check if transaction is empty
	if len(tx.Inputs) == 0 && len(tx.Outputs) == 0 {
		return &ValidationError{
			Field:  "transaction",
			Reason: "empty transaction",
		}
	}

	// Verify transaction signature
	if !tx.Verify() {
		return &ValidationError{
			Field:  "signature",
			Reason: "invalid signature",
		}
	}

	// Check if the transaction is a coinbase transaction
	if tx.IsCoinbase() {
		return nil
	}

	// Validate inputs
	for i, input := range tx.Inputs {
		if len(input.TxID) == 0 {
			return &ValidationError{
				Field:  fmt.Sprintf("input[%d].TxID", i),
				Reason: "empty transaction ID",
			}
		}

		if input.OutputIndex < 0 {
			return &ValidationError{
				Field:  fmt.Sprintf("input[%d].OutputIndex", i),
				Reason: "invalid output index",
			}
		}

		// Check if input exists in UTXO set
		utxo := utxoSet.GetUTXO(input.TxID, input.OutputIndex)
		if len(utxo.TxID) == 0 {
			return &ValidationError{
				Field:  fmt.Sprintf("input[%d]", i),
				Reason: "UTXO not found",
			}
		}

		// Verify input ownership
		pubKey, err := crypto.BytesToPublicKey(input.PublicKey)
		if err != nil {
			return &ValidationError{
				Field:  fmt.Sprintf("input[%d]", i),
				Reason: "invalid public key",
			}
		}
		if !bytes.Equal(utxo.PublicKeyHash, crypto.HashPublicKey(pubKey)) {
			return &ValidationError{
				Field:  fmt.Sprintf("input[%d]", i),
				Reason: "unauthorized input",
			}
		}
	}

	// Validate outputs
	for i, output := range tx.Outputs {
		if output.Value <= 0 {
			return &ValidationError{
				Field:  fmt.Sprintf("output[%d].Value", i),
				Reason: "invalid amount",
			}
		}

		if len(output.PublicKeyHash) == 0 {
			return &ValidationError{
				Field:  fmt.Sprintf("output[%d].PublicKeyHash", i),
				Reason: "empty public key hash",
			}
		}

		if output.CoinType == "" {
			return &ValidationError{
				Field:  fmt.Sprintf("output[%d].CoinType", i),
				Reason: "invalid coin type",
			}
		}
	}

	// Check input/output balance for each coin type
	inputBalances := make(map[CoinType]float64)
	outputBalances := make(map[CoinType]float64)

	// Calculate input balances
	for _, input := range tx.Inputs {
		utxo := utxoSet.GetUTXO(input.TxID, input.OutputIndex)
		inputBalances[utxo.CoinType] += utxo.Amount
	}

	// Calculate output balances
	for _, output := range tx.Outputs {
		outputBalances[output.CoinType] += output.Value
	}

	// Validate balances and handle coin conversions
	for coinType, outputAmount := range outputBalances {
		inputAmount := inputBalances[coinType]

		// Check if we need to convert from other coins
		if inputAmount < outputAmount {
			// Try to convert from lower denomination coins
			switch coinType {
			case Shiblum:
				leahAmount := inputBalances[Leah]
				if leahAmount >= (outputAmount-inputAmount)*2 {
					continue
				}
			case Shiblon:
				shiblumAmount := inputBalances[Shiblum]
				if shiblumAmount >= (outputAmount-inputAmount)*2 {
					continue
				}
			case Senum:
				shiblonAmount := inputBalances[Shiblon]
				if shiblonAmount >= (outputAmount-inputAmount)/2 {
					continue
				}
			}

			return &ValidationError{
				Field:  "balance",
				Reason: fmt.Sprintf("insufficient balance for %s", coinType),
			}
		}
	}

	// Validate cross-block transfers
	if tx.BlockType != "" {
		for _, output := range tx.Outputs {
			if !CanTransferBetweenBlocks(output.CoinType) {
				return &ValidationError{
					Field:  "block_type",
					Reason: fmt.Sprintf("coin type %s cannot be transferred between blocks", output.CoinType),
				}
			}
		}
	}

	return nil
}

// Verify verifies the transaction signature
func (tx *Transaction) Verify() bool {
	txCopy := tx.TrimmedCopy()

	for i, input := range tx.Inputs {
		// Set the public key for this input
		txCopy.Inputs[i].PublicKey = input.PublicKey

		// Calculate the hash of the transaction
		hash := txCopy.CalculateHash()

		// Verify the signature
		if !crypto.Verify(hash, input.Signature, input.PublicKey) {
			return false
		}
	}

	return true
}

// TransactionBatch represents a batch of transactions
type TransactionBatch struct {
	Transactions []*Transaction
	Timestamp    time.Time
	BatchID      []byte
	mu           sync.RWMutex
}

// NewTransactionBatch creates a new transaction batch
func NewTransactionBatch() *TransactionBatch {
	return &TransactionBatch{
		Transactions: make([]*Transaction, 0),
		Timestamp:    time.Now(),
		BatchID:      make([]byte, 32),
	}
}

// AddTransaction adds a transaction to the batch
func (b *TransactionBatch) AddTransaction(tx *Transaction) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Validate transaction
	if tx == nil {
		return fmt.Errorf("nil transaction")
	}

	// Check batch size limit (e.g., 1000 transactions)
	if len(b.Transactions) >= 1000 {
		return fmt.Errorf("batch size limit reached")
	}

	// Add transaction to batch
	b.Transactions = append(b.Transactions, tx)
	return nil
}

// GetBatchSize returns the number of transactions in the batch
func (b *TransactionBatch) GetBatchSize() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.Transactions)
}

// GetBatchHash returns the hash of the batch
func (b *TransactionBatch) GetBatchHash() []byte {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Create a copy of transactions for hashing
	txs := make([]*Transaction, len(b.Transactions))
	copy(txs, b.Transactions)

	// Sort transactions by ID for consistent hashing
	sort.Slice(txs, func(i, j int) bool {
		return bytes.Compare(txs[i].ID, txs[j].ID) < 0
	})

	// Hash the batch
	data, _ := json.Marshal(txs)
	hash := sha256.Sum256(data)
	return hash[:]
}

// ValidateBatch validates all transactions in the batch
func (b *TransactionBatch) ValidateBatch(utxoSet *UTXOSet) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Create a map to track UTXO usage
	utxoUsage := make(map[string]bool)

	// Validate each transaction
	for i, tx := range b.Transactions {
		// Check for double spending within the batch
		for _, input := range tx.Inputs {
			utxoKey := fmt.Sprintf("%x:%d", input.TxID, input.OutputIndex)
			if utxoUsage[utxoKey] {
				return fmt.Errorf("double spending detected in batch: transaction %d", i)
			}
			utxoUsage[utxoKey] = true
		}

		// Validate transaction
		if err := tx.Validate(utxoSet); err != nil {
			return fmt.Errorf("invalid transaction %d: %v", i, err)
		}
	}

	return nil
}

// ProcessBatch processes the batch of transactions
func (b *TransactionBatch) ProcessBatch(utxoSet *UTXOSet) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Validate batch
	if err := b.ValidateBatch(utxoSet); err != nil {
		return err
	}

	// Process transactions in parallel
	var wg sync.WaitGroup
	errChan := make(chan error, len(b.Transactions))

	for _, tx := range b.Transactions {
		wg.Add(1)
		go func(tx *Transaction) {
			defer wg.Done()
			if err := utxoSet.ProcessTransaction(tx); err != nil {
				errChan <- err
			}
		}(tx)
	}

	// Wait for all transactions to be processed
	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// BatchProcessor handles transaction batching
type BatchProcessor struct {
	currentBatch *TransactionBatch
	utxoSet      *UTXOSet
	mu           sync.RWMutex
	batchSize    int
	batchTimeout time.Duration
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(utxoSet *UTXOSet) *BatchProcessor {
	return &BatchProcessor{
		currentBatch: NewTransactionBatch(),
		utxoSet:      utxoSet,
		batchSize:    1000,
		batchTimeout: 5 * time.Second,
	}
}

// AddTransaction adds a transaction to the current batch
func (p *BatchProcessor) AddTransaction(tx *Transaction) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Add transaction to batch
	if err := p.currentBatch.AddTransaction(tx); err != nil {
		return err
	}

	// Process batch if size limit reached
	if p.currentBatch.GetBatchSize() >= p.batchSize {
		return p.processCurrentBatch()
	}

	return nil
}

// processCurrentBatch processes the current batch
func (p *BatchProcessor) processCurrentBatch() error {
	if p.currentBatch.GetBatchSize() == 0 {
		return nil
	}

	// Process batch
	if err := p.currentBatch.ProcessBatch(p.utxoSet); err != nil {
		return err
	}

	// Create new batch
	p.currentBatch = NewTransactionBatch()
	return nil
}

// Start starts the batch processor
func (p *BatchProcessor) Start() {
	go func() {
		ticker := time.NewTicker(p.batchTimeout)
		defer ticker.Stop()

		for range ticker.C {
			p.mu.Lock()
			if err := p.processCurrentBatch(); err != nil {
				// Log error but continue processing
				fmt.Printf("Error processing batch: %v\n", err)
			}
			p.mu.Unlock()
		}
	}()
}
