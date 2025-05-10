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
)

// Block represents a block that needs to be mined
type Block interface {
	GetHash() []byte
	GetPrevHash() []byte
	GetTimestamp() int64
	GetData() []byte
	GetNonce() int64
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
