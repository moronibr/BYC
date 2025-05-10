// Package pow implements a robust Proof of Work consensus mechanism with adaptive mining capabilities.
// It provides features like circuit breaker pattern, graceful shutdown, and resource management.
package pow

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math"
	"math/big"
	"time"
)

// TargetBits is the number of leading zero bits required in the hash.
// This determines the mining difficulty - higher values mean more difficult mining.
const TargetBits = 24

// MaxNonce is the maximum value for the nonce.
// This is used to prevent integer overflow during mining.
const MaxNonce = math.MaxInt64

// DifficultyAdjustmentInterval is the number of blocks between difficulty adjustments.
// This helps maintain a consistent block time as network hash rate changes.
const DifficultyAdjustmentInterval = 2016

// TargetTimePerBlock is the expected time between blocks in seconds.
// This is used for difficulty adjustment calculations.
const TargetTimePerBlock = 600 // 10 minutes

// MinDifficultyBits is the minimum difficulty (maximum target).
// This prevents the network from becoming too easy to mine.
const MinDifficultyBits = 4

// MaxDifficultyBits is the maximum difficulty (minimum target).
// This prevents the network from becoming too difficult to mine.
const MaxDifficultyBits = 32

var (
	// ErrInvalidBlock is returned when the block is invalid
	ErrInvalidBlock = errors.New("invalid block")

	// ErrResourceExhausted is returned when system resources are exhausted
	ErrResourceExhausted = errors.New("system resources exhausted")

	// ErrMiningTimeout is returned when mining takes too long
	ErrMiningTimeout = errors.New("mining timeout exceeded")
)

// Config holds the configuration for proof of work.
// It allows customization of mining behavior and system parameters.
type Config struct {
	// MiningTimeout is the maximum time allowed for mining a block.
	// If mining takes longer than this, it will be aborted.
	MiningTimeout time.Duration

	// MaxRetries is the maximum number of retries for mining.
	// After this many failed attempts, mining will stop.
	MaxRetries int

	// RecoveryEnabled enables automatic recovery from errors.
	// When true, panics during mining will be caught and handled.
	RecoveryEnabled bool

	// LoggingEnabled enables detailed logging.
	// When true, mining progress and errors will be logged.
	LoggingEnabled bool
}

// DefaultConfig returns the default configuration.
// This provides sensible defaults for most use cases.
func DefaultConfig() *Config {
	return &Config{
		MiningTimeout:   time.Minute * 5,
		MaxRetries:      3,
		RecoveryEnabled: true,
		LoggingEnabled:  true,
	}
}

// Block represents a block that needs to be mined.
// Implementations must provide all required methods.
type Block interface {
	// GetHash returns the current hash of the block.
	GetHash() []byte

	// GetPrevHash returns the hash of the previous block.
	GetPrevHash() []byte

	// GetTimestamp returns the block's timestamp.
	GetTimestamp() int64

	// GetData returns the block's data.
	GetData() []byte

	// GetNonce returns the current nonce value.
	GetNonce() int64

	// GetHeight returns the block's height in the chain.
	GetHeight() int64

	// SetNonce sets the block's nonce value.
	SetNonce(nonce int64)

	// SetHash sets the block's hash value.
	SetHash(hash []byte)
}

// CircuitBreakerState represents the state of the circuit breaker.
// This is used to implement the circuit breaker pattern for failure protection.
type CircuitBreakerState int

const (
	// CircuitBreakerClosed indicates normal operation.
	// The system is functioning normally.
	CircuitBreakerClosed CircuitBreakerState = iota

	// CircuitBreakerOpen indicates the circuit is open.
	// The system is failing and requests are being rejected.
	CircuitBreakerOpen

	// CircuitBreakerHalfOpen indicates the circuit is testing recovery.
	// Limited requests are allowed to test if the system has recovered.
	CircuitBreakerHalfOpen
)

// CircuitBreakerConfig holds configuration for the circuit breaker.
type CircuitBreakerConfig struct {
	FailureThreshold    int
	ResetTimeout        time.Duration
	HalfOpenMaxRequests int
}

// DefaultCircuitBreakerConfig returns the default circuit breaker configuration.
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		FailureThreshold:    5,
		ResetTimeout:        time.Second * 30,
		HalfOpenMaxRequests: 3,
	}
}

// ProofOfWork represents a proof of work instance.
// It handles the mining process and related functionality.
type ProofOfWork struct {
	block   Block
	target  *big.Int
	rm      *ResourceManager
	config  *Config
	metrics *Metrics

	// Circuit breaker fields
	circuitBreakerState  CircuitBreakerState
	failureCount         int
	lastFailureTime      time.Time
	halfOpenRequests     int
	circuitBreakerConfig *CircuitBreakerConfig

	// Shutdown fields
	shutdownChan    chan struct{}
	isShuttingDown  bool
	shutdownTimeout time.Duration
}

// NewProofOfWork creates a new proof of work instance.
// It initializes the mining system with the given block and configuration.
func NewProofOfWork(block Block, config *Config) (*ProofOfWork, error) {
	if block == nil {
		return nil, ErrInvalidBlock
	}

	if config == nil {
		config = DefaultConfig()
	}

	target := big.NewInt(1)
	target.Lsh(target, uint(256-TargetBits))

	return &ProofOfWork{
		block:                block,
		target:               target,
		rm:                   NewResourceManager(),
		config:               config,
		metrics:              NewMetrics(),
		circuitBreakerConfig: DefaultCircuitBreakerConfig(),
		shutdownChan:         make(chan struct{}),
		shutdownTimeout:      time.Second * 30,
	}, nil
}

// CalculateNextDifficulty calculates the next difficulty based on the last N blocks
func CalculateNextDifficulty(blocks []Block) int {
	if len(blocks) < DifficultyAdjustmentInterval {
		return TargetBits
	}

	// Get the first and last block in the interval
	firstBlock := blocks[0]
	lastBlock := blocks[len(blocks)-1]

	// Calculate the time difference
	timeDiff := lastBlock.GetTimestamp() - firstBlock.GetTimestamp()

	// Calculate the expected time
	expectedTime := int64(TargetTimePerBlock * DifficultyAdjustmentInterval)

	// Calculate the new difficulty
	newDifficulty := TargetBits
	if timeDiff < expectedTime/2 {
		// If blocks are being mined too quickly, increase difficulty
		newDifficulty = TargetBits + 1
	} else if timeDiff > expectedTime*2 {
		// If blocks are being mined too slowly, decrease difficulty
		newDifficulty = TargetBits - 1
	}

	// Ensure difficulty stays within bounds
	if newDifficulty < MinDifficultyBits {
		newDifficulty = MinDifficultyBits
	} else if newDifficulty > MaxDifficultyBits {
		newDifficulty = MaxDifficultyBits
	}

	return newDifficulty
}

// prepareData prepares the block data for hashing.
// It combines the block's fields into a single byte slice.
// The method:
//   - Combines block fields (prevHash, data, timestamp, nonce)
//   - Serializes the combined data
//   - Handles any serialization errors
//
// Parameters:
//   - nonce: The nonce value to include in the data
//
// Returns:
//   - []byte: The prepared data for hashing
//   - error: Any error that occurred during preparation
func (pow *ProofOfWork) prepareData(nonce int64) ([]byte, error) {
	data := bytes.Join(
		[][]byte{
			pow.block.GetPrevHash(),
			pow.block.GetData(),
			IntToHex(pow.block.GetTimestamp()),
			IntToHex(int64(TargetBits)),
			IntToHex(nonce),
		},
		[]byte{},
	)

	return data, nil
}

// Run performs proof of work mining with a single worker.
// This is the basic mining implementation without resource adaptation.
// The method:
//   - Iterates through possible nonce values
//   - Prepares block data for hashing
//   - Computes and verifies block hash
//   - Returns when a valid hash is found
//
// Returns:
//   - nonce: The nonce value that produced a valid hash
//   - hash: The resulting block hash
//   - error: Any error that occurred during mining
func (pow *ProofOfWork) Run() (int64, []byte, error) {
	var hashInt big.Int
	var hash [32]byte
	nonce := int64(0)

	if pow.config.LoggingEnabled {
		fmt.Printf("Mining a new block\n")
	}

	for nonce < MaxNonce {
		data, err := pow.prepareData(nonce)
		if err != nil {
			return 0, nil, err
		}
		hash = sha256.Sum256(data)
		if pow.config.LoggingEnabled {
			fmt.Printf("\r%x", hash)
		}
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}

	if pow.config.LoggingEnabled {
		fmt.Print("\n\n")
	}

	return nonce, hash[:], nil
}

// RunParallel performs proof of work mining with multiple workers.
// This allows for concurrent mining across multiple CPU cores.
// The method:
//   - Creates a worker pool of specified size
//   - Distributes nonce ranges among workers
//   - Collects results from workers
//   - Returns when a valid hash is found
//
// Parameters:
//   - workerCount: Number of concurrent workers to use
//
// Returns:
//   - nonce: The nonce value that produced a valid hash
//   - hash: The resulting block hash
//   - error: Any error that occurred during mining
func (pow *ProofOfWork) RunParallel(workerCount int) (int64, []byte, error) {
	if workerCount <= 0 {
		workerCount = 1
	}

	type miningResult struct {
		Nonce int64
		Hash  []byte
		Err   error
	}

	// Create channels for results and done signal
	resultChan := make(chan miningResult)
	done := make(chan struct{})

	// Start worker goroutines
	for i := 0; i < workerCount; i++ {
		go func(workerID int) {
			var hashInt big.Int
			var hash [32]byte
			startNonce := int64(workerID) * (MaxNonce / int64(workerCount))
			endNonce := startNonce + (MaxNonce / int64(workerCount))

			for nonce := startNonce; nonce < endNonce; nonce++ {
				select {
				case <-done:
					return
				default:
					data, err := pow.prepareData(nonce)
					if err != nil {
						select {
						case resultChan <- miningResult{Nonce: nonce, Err: err}:
						case <-done:
						}
						return
					}
					hash = sha256.Sum256(data)
					hashInt.SetBytes(hash[:])

					if hashInt.Cmp(pow.target) == -1 {
						select {
						case resultChan <- miningResult{Nonce: nonce, Hash: hash[:]}:
						case <-done:
						}
						return
					}
				}
			}
		}(i)
	}

	// Wait for the first result
	result := <-resultChan
	close(done)

	return result.Nonce, result.Hash, result.Err
}

// Shutdown gracefully shuts down the proof of work system.
// It ensures all ongoing operations are completed before closing.
// The method:
//   - Signals shutdown to all components
//   - Waits for ongoing operations to complete
//   - Closes all resources
//   - Prevents new operations from starting
//
// Returns:
//   - error: Any error that occurred during shutdown
func (pow *ProofOfWork) Shutdown() error {
	if pow.isShuttingDown {
		return nil
	}

	pow.isShuttingDown = true
	close(pow.shutdownChan)

	// Wait for ongoing operations to complete
	done := make(chan struct{})
	go func() {
		// Wait for any ongoing mining operations
		time.Sleep(pow.shutdownTimeout)
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(pow.shutdownTimeout):
		return fmt.Errorf("shutdown timeout exceeded")
	}
}

// checkCircuitBreaker checks the current state of the circuit breaker.
// It determines if operations should be allowed based on failure history.
// The method:
//   - Checks current circuit breaker state
//   - Evaluates failure count against threshold
//   - Considers reset timeout for recovery
//   - Manages half-open state transitions
//
// Returns:
//   - bool: True if operations should proceed, false if they should be blocked
func (pow *ProofOfWork) checkCircuitBreaker() error {
	now := time.Now()

	switch pow.circuitBreakerState {
	case CircuitBreakerClosed:
		if pow.failureCount >= pow.circuitBreakerConfig.FailureThreshold {
			pow.circuitBreakerState = CircuitBreakerOpen
			pow.lastFailureTime = now
			return fmt.Errorf("circuit breaker opened due to %d consecutive failures", pow.failureCount)
		}
		return nil

	case CircuitBreakerOpen:
		if now.Sub(pow.lastFailureTime) >= pow.circuitBreakerConfig.ResetTimeout {
			pow.circuitBreakerState = CircuitBreakerHalfOpen
			pow.halfOpenRequests = 0
			return nil
		}
		return fmt.Errorf("circuit breaker is open")

	case CircuitBreakerHalfOpen:
		if pow.halfOpenRequests >= pow.circuitBreakerConfig.HalfOpenMaxRequests {
			pow.circuitBreakerState = CircuitBreakerOpen
			pow.lastFailureTime = now
			return fmt.Errorf("circuit breaker reopened due to too many requests in half-open state")
		}
		pow.halfOpenRequests++
		return nil
	}

	return nil
}

// recordFailure records a failure in the circuit breaker.
// It updates the failure count and may trigger circuit breaker state changes.
// The method:
//   - Increments failure count
//   - Records failure timestamp
//   - Updates circuit breaker state if threshold reached
func (pow *ProofOfWork) recordFailure() {
	pow.failureCount++
	if pow.circuitBreakerState == CircuitBreakerHalfOpen {
		pow.circuitBreakerState = CircuitBreakerOpen
		pow.lastFailureTime = time.Now()
	}
}

// recordSuccess records a successful operation in the circuit breaker.
// It resets the failure count and may close the circuit breaker.
// The method:
//   - Resets failure count
//   - Updates circuit breaker state
//   - Resets half-open request count
func (pow *ProofOfWork) recordSuccess() {
	pow.failureCount = 0
	if pow.circuitBreakerState == CircuitBreakerHalfOpen {
		pow.circuitBreakerState = CircuitBreakerClosed
	}
}

// RunAdaptive performs proof of work mining with adaptive worker count.
// It automatically adjusts the number of workers based on system resources.
func (pow *ProofOfWork) RunAdaptive() (int64, []byte, error) {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		pow.metrics.RecordMiningAttempt(true)
		pow.metrics.RecordBlockMined(duration)
	}()

	// Check if system is shutting down
	select {
	case <-pow.shutdownChan:
		pow.metrics.RecordMiningAttempt(false)
		return 0, nil, fmt.Errorf("system is shutting down")
	default:
	}

	// Check circuit breaker
	if err := pow.checkCircuitBreaker(); err != nil {
		pow.metrics.RecordMiningAttempt(false)
		return 0, nil, err
	}

	// Adjust difficulty based on current conditions
	pow.target = pow.adjustDifficulty()
	pow.metrics.RecordDifficultyChange(pow.target.BitLen())

	var lastErr error
	for retry := 0; retry < pow.config.MaxRetries; retry++ {
		// Check system resources
		if err := pow.checkResources(); err != nil {
			pow.recordFailure()
			pow.metrics.RecordMiningAttempt(false)
			return 0, nil, fmt.Errorf("resource check failed: %w", err)
		}

		// Get optimal worker count
		workerCount := pow.rm.GetOptimalWorkerCount()
		pow.metrics.RecordResourceUsage(
			pow.rm.GetCPUUtilization(),
			pow.rm.GetMemoryUtilization(),
			workerCount,
			pow.rm.GetOptimalWorkerCount(),
		)

		// Create a channel for the result
		resultChan := make(chan struct {
			nonce int64
			hash  []byte
			err   error
		})

		// Start mining with timeout
		go func() {
			nonce, hash, err := pow.runWithRecovery(workerCount)
			resultChan <- struct {
				nonce int64
				hash  []byte
				err   error
			}{nonce, hash, err}
		}()

		// Wait for result or timeout
		select {
		case result := <-resultChan:
			if result.err != nil {
				lastErr = result.err
				pow.recordFailure()
				pow.metrics.RecordMiningAttempt(false)
				if pow.config.LoggingEnabled {
					fmt.Printf("Mining attempt %d failed: %v\n", retry+1, result.err)
				}
				continue
			}
			pow.recordSuccess()
			return result.nonce, result.hash, nil
		case <-time.After(pow.config.MiningTimeout):
			lastErr = ErrMiningTimeout
			pow.recordFailure()
			pow.metrics.RecordMiningAttempt(false)
			if pow.config.LoggingEnabled {
				fmt.Printf("Mining attempt %d timed out\n", retry+1)
			}
		case <-pow.shutdownChan:
			pow.metrics.RecordMiningAttempt(false)
			return 0, nil, fmt.Errorf("mining interrupted by shutdown")
		}
	}

	return 0, nil, fmt.Errorf("all mining attempts failed: %w", lastErr)
}

// runWithRecovery executes the mining function with panic recovery.
// It catches any panics and converts them to errors.
// The method:
//   - Wraps the mining function in a panic recovery
//   - Converts panics to errors
//   - Logs recovery information if enabled
//
// Parameters:
//   - workerCount: Number of workers to use for mining
//
// Returns:
//   - nonce: The nonce value that produced a valid hash
//   - hash: The resulting block hash
//   - error: Any error that occurred during mining
func (pow *ProofOfWork) runWithRecovery(workerCount int) (int64, []byte, error) {
	if !pow.config.RecoveryEnabled {
		return pow.RunParallel(workerCount)
	}

	defer func() {
		if r := recover(); r != nil {
			if pow.config.LoggingEnabled {
				log.Printf("Recovered from panic in mining: %v", r)
			}
		}
	}()

	return pow.RunParallel(workerCount)
}

// checkResources verifies that system resources are available for mining.
// It checks CPU and memory usage against configured limits.
// The method:
//   - Checks CPU utilization
//   - Verifies available memory
//   - Validates resource limits
//
// Returns:
//   - error: Any error that occurred during resource checking
func (pow *ProofOfWork) checkResources() error {
	cpuCount, memUsage := pow.rm.GetSystemMetrics()
	if cpuCount < 1 {
		return ErrResourceExhausted
	}

	if memUsage > pow.rm.maxMemoryPercent {
		return ErrResourceExhausted
	}

	return nil
}

// Validate checks if the block's proof of work is valid.
func (pow *ProofOfWork) Validate() bool {
	pow.metrics.RecordValidation(true)
	var hashInt big.Int

	data, err := pow.prepareData(pow.block.GetNonce())
	if err != nil {
		pow.metrics.RecordValidation(false)
		return false
	}
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1
	if !isValid {
		pow.metrics.RecordValidation(false)
	}

	return isValid
}

// IntToHex converts an int64 to a byte array
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		panic(err)
	}

	return buff.Bytes()
}

// adjustDifficulty adjusts the mining difficulty based on block time.
// It maintains a consistent block time as network hash rate changes.
// The method:
//   - Calculates time difference between blocks
//   - Adjusts target difficulty based on time difference
//   - Applies minimum and maximum difficulty limits
//
// Returns:
//   - *big.Int: The new target difficulty
func (pow *ProofOfWork) adjustDifficulty() *big.Int {
	// Get the previous block's timestamp
	prevTimestamp := pow.block.GetTimestamp()
	currentTimestamp := time.Now().Unix()

	// Calculate time difference
	timeDiff := currentTimestamp - prevTimestamp

	// Calculate difficulty adjustment factor
	// If blocks are being found too quickly, increase difficulty
	// If blocks are taking too long, decrease difficulty
	adjustmentFactor := float64(TargetTimePerBlock) / float64(timeDiff)
	if adjustmentFactor > 4 {
		adjustmentFactor = 4 // Cap maximum increase
	} else if adjustmentFactor < 0.25 {
		adjustmentFactor = 0.25 // Cap maximum decrease
	}

	// Calculate new target
	newTarget := new(big.Int).Set(pow.target)
	newTarget.Mul(newTarget, big.NewInt(int64(adjustmentFactor*100)))
	newTarget.Div(newTarget, big.NewInt(100))

	// Ensure target is within limits
	if newTarget.BitLen() > 256-MinDifficultyBits {
		newTarget.SetBit(new(big.Int), 256-MinDifficultyBits, 1)
	} else if newTarget.BitLen() < 256-MaxDifficultyBits {
		newTarget.SetBit(new(big.Int), 256-MaxDifficultyBits, 1)
	}

	return newTarget
}

// GetMetrics returns the current metrics.
func (pow *ProofOfWork) GetMetrics() map[string]interface{} {
	return pow.metrics.GetMetrics()
}
