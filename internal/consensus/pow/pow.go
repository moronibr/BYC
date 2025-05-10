package pow

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
)

const (
	// TargetBits is the number of leading zero bits required in the hash
	TargetBits = 24

	// MaxNonce is the maximum value for the nonce
	MaxNonce = math.MaxInt64

	// DifficultyAdjustmentInterval is the number of blocks between difficulty adjustments
	DifficultyAdjustmentInterval = 2016

	// TargetTimePerBlock is the expected time between blocks in seconds
	TargetTimePerBlock = 600 // 10 minutes

	// MinDifficultyBits is the minimum difficulty (maximum target)
	MinDifficultyBits = 4

	// MaxDifficultyBits is the maximum difficulty (minimum target)
	MaxDifficultyBits = 32
)

// Block represents a block that needs to be mined
type Block interface {
	GetHash() []byte
	GetPrevHash() []byte
	GetTimestamp() int64
	GetData() []byte
	GetNonce() int64
	GetHeight() int64
	SetNonce(nonce int64)
	SetHash(hash []byte)
}

// ProofOfWork represents a proof of work
type ProofOfWork struct {
	block  Block
	target *big.Int
}

// NewProofOfWork creates a new proof of work
func NewProofOfWork(block Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-TargetBits))

	return &ProofOfWork{
		block:  block,
		target: target,
	}
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
	expectedTime := TargetTimePerBlock * DifficultyAdjustmentInterval

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

// prepareData prepares the data for hashing
func (pow *ProofOfWork) prepareData(nonce int64) []byte {
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

	return data
}

// Run performs the proof of work
func (pow *ProofOfWork) Run() (int64, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := int64(0)

	fmt.Printf("Mining a new block\n")
	for nonce < MaxNonce {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}
	fmt.Print("\n\n")

	return nonce, hash[:]
}

// RunParallel performs parallel proof of work
func (pow *ProofOfWork) RunParallel(numWorkers int) (int64, []byte) {
	if numWorkers <= 0 {
		numWorkers = 1
	}

	type result struct {
		nonce int64
		hash  []byte
	}

	// Create channels for results and done signal
	resultChan := make(chan result)
	done := make(chan struct{})

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			var hashInt big.Int
			var hash [32]byte
			startNonce := int64(workerID) * (MaxNonce / int64(numWorkers))
			endNonce := startNonce + (MaxNonce / int64(numWorkers))

			for nonce := startNonce; nonce < endNonce; nonce++ {
				select {
				case <-done:
					return
				default:
					data := pow.prepareData(nonce)
					hash = sha256.Sum256(data)
					hashInt.SetBytes(hash[:])

					if hashInt.Cmp(pow.target) == -1 {
						select {
						case resultChan <- result{nonce, hash[:]}:
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

	return result.nonce, result.hash
}

// Validate validates the proof of work
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.GetNonce())
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

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
