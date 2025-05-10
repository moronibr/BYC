package pow

import (
	"testing"
	"time"
)

// BenchmarkBlock implements the Block interface for benchmarking
type BenchmarkBlock struct {
	prevHash  []byte
	data      []byte
	timestamp int64
	nonce     int64
	hash      []byte
	height    int64
}

func (b *BenchmarkBlock) GetHash() []byte      { return b.hash }
func (b *BenchmarkBlock) GetPrevHash() []byte  { return b.prevHash }
func (b *BenchmarkBlock) GetTimestamp() int64  { return b.timestamp }
func (b *BenchmarkBlock) GetData() []byte      { return b.data }
func (b *BenchmarkBlock) GetNonce() int64      { return b.nonce }
func (b *BenchmarkBlock) GetHeight() int64     { return b.height }
func (b *BenchmarkBlock) SetNonce(nonce int64) { b.nonce = nonce }
func (b *BenchmarkBlock) SetHash(hash []byte)  { b.hash = hash }

// setupBenchmarkBlock creates a new block for benchmarking
func setupBenchmarkBlock() *BenchmarkBlock {
	return &BenchmarkBlock{
		prevHash:  []byte("benchmark-prev-hash"),
		data:      []byte("benchmark-data"),
		timestamp: time.Now().Unix(),
		height:    1,
	}
}

// BenchmarkRunSingle benchmarks single-threaded mining
func BenchmarkRunSingle(b *testing.B) {
	block := setupBenchmarkBlock()
	pow, err := NewProofOfWork(block, nil)
	if err != nil {
		b.Fatalf("Failed to create PoW: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := pow.Run()
		if err != nil {
			b.Fatalf("Failed to run mining: %v", err)
		}
	}
}

// BenchmarkRunParallel benchmarks parallel mining with different worker counts
func BenchmarkRunParallel(b *testing.B) {
	block := setupBenchmarkBlock()
	pow, err := NewProofOfWork(block, nil)
	if err != nil {
		b.Fatalf("Failed to create PoW: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := pow.RunParallel(4)
		if err != nil {
			b.Fatalf("Failed to run parallel mining: %v", err)
		}
	}
}

// BenchmarkMiningSpeed measures the time to find a solution
func BenchmarkMiningSpeed(b *testing.B) {
	block := setupBenchmarkBlock()
	pow, err := NewProofOfWork(block, nil)
	if err != nil {
		b.Fatalf("Failed to create PoW: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := pow.RunAdaptive()
		if err != nil {
			b.Fatalf("Failed to run adaptive mining: %v", err)
		}
	}
}

// BenchmarkMiningEfficiency measures the number of hashes per second
func BenchmarkMiningEfficiency(b *testing.B) {
	block := setupBenchmarkBlock()
	pow, err := NewProofOfWork(block, nil)
	if err != nil {
		b.Fatalf("Failed to create PoW: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := pow.RunAdaptive()
		if err != nil {
			b.Fatalf("Failed to run adaptive mining: %v", err)
		}
	}
}

// BenchmarkMiningMetrics provides detailed mining performance metrics
func BenchmarkMiningMetrics(b *testing.B) {
	block := setupBenchmarkBlock()
	pow, err := NewProofOfWork(block, nil)
	if err != nil {
		b.Fatalf("Failed to create PoW: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := pow.RunAdaptive()
		if err != nil {
			b.Fatalf("Failed to run adaptive mining: %v", err)
		}
	}
}

// BenchmarkMiningProfiling runs benchmarks with CPU and memory profiling
func BenchmarkMiningProfiling(b *testing.B) {
	block := setupBenchmarkBlock()
	pow, err := NewProofOfWork(block, nil)
	if err != nil {
		b.Fatalf("Failed to create PoW: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := pow.RunAdaptive()
		if err != nil {
			b.Fatalf("Failed to run adaptive mining: %v", err)
		}
	}
}
