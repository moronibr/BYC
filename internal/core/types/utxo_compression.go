package types

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
)

const (
	// DefaultCompressionLevel is the default compression level (1-9)
	DefaultCompressionLevel = 6
	// DefaultCompressionThreshold is the default size threshold for compression in bytes
	DefaultCompressionThreshold = 1024 // 1KB
	// DefaultCompressionMethod is the default compression method
	DefaultCompressionMethod = CompressionMethodGzip
)

// CompressionMethod represents the compression method to use
type CompressionMethod byte

const (
	// CompressionMethodNone indicates no compression
	CompressionMethodNone CompressionMethod = iota
	// CompressionMethodGzip indicates gzip compression
	CompressionMethodGzip
	// CompressionMethodZlib indicates zlib compression
	CompressionMethodZlib
)

// UTXOCompression handles compression of the UTXO set
type UTXOCompression struct {
	utxoSet *UTXOSet
	mu      sync.RWMutex

	// Compression state
	method    CompressionMethod
	level     int
	threshold int64
}

// NewUTXOCompression creates a new UTXO compression handler
func NewUTXOCompression(utxoSet *UTXOSet) *UTXOCompression {
	return &UTXOCompression{
		utxoSet:   utxoSet,
		method:    DefaultCompressionMethod,
		level:     DefaultCompressionLevel,
		threshold: DefaultCompressionThreshold,
	}
}

// Compress compresses the UTXO set data
func (uc *UTXOCompression) Compress(data []byte) ([]byte, error) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	// Check if data is large enough to compress
	if int64(len(data)) < uc.threshold {
		return data, nil
	}

	// Create buffer for compressed data
	var buf bytes.Buffer

	// Write compression header
	header := make([]byte, 5)
	header[0] = byte(uc.method)
	binary.LittleEndian.PutUint32(header[1:5], uint32(len(data)))
	if _, err := buf.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write compression header: %v", err)
	}

	// Compress data based on method
	switch uc.method {
	case CompressionMethodNone:
		if _, err := buf.Write(data); err != nil {
			return nil, fmt.Errorf("failed to write uncompressed data: %v", err)
		}

	case CompressionMethodGzip:
		writer, err := gzip.NewWriterLevel(&buf, uc.level)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip writer: %v", err)
		}
		if _, err := writer.Write(data); err != nil {
			return nil, fmt.Errorf("failed to write gzip data: %v", err)
		}
		if err := writer.Close(); err != nil {
			return nil, fmt.Errorf("failed to close gzip writer: %v", err)
		}

	case CompressionMethodZlib:
		writer, err := zlib.NewWriterLevel(&buf, uc.level)
		if err != nil {
			return nil, fmt.Errorf("failed to create zlib writer: %v", err)
		}
		if _, err := writer.Write(data); err != nil {
			return nil, fmt.Errorf("failed to write zlib data: %v", err)
		}
		if err := writer.Close(); err != nil {
			return nil, fmt.Errorf("failed to close zlib writer: %v", err)
		}

	default:
		return nil, fmt.Errorf("unsupported compression method: %d", uc.method)
	}

	return buf.Bytes(), nil
}

// Decompress decompresses the UTXO set data
func (uc *UTXOCompression) Decompress(data []byte) ([]byte, error) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	// Check if data is compressed
	if len(data) < 5 {
		return data, nil
	}

	// Read compression header
	method := CompressionMethod(data[0])
	originalSize := binary.LittleEndian.Uint32(data[1:5])
	compressedData := data[5:]

	// Create buffer for decompressed data
	var buf bytes.Buffer

	// Decompress data based on method
	switch method {
	case CompressionMethodNone:
		if _, err := buf.Write(compressedData); err != nil {
			return nil, fmt.Errorf("failed to write uncompressed data: %v", err)
		}

	case CompressionMethodGzip:
		reader, err := gzip.NewReader(bytes.NewReader(compressedData))
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %v", err)
		}
		defer reader.Close()
		if _, err := io.Copy(&buf, reader); err != nil {
			return nil, fmt.Errorf("failed to read gzip data: %v", err)
		}

	case CompressionMethodZlib:
		reader, err := zlib.NewReader(bytes.NewReader(compressedData))
		if err != nil {
			return nil, fmt.Errorf("failed to create zlib reader: %v", err)
		}
		defer reader.Close()
		if _, err := io.Copy(&buf, reader); err != nil {
			return nil, fmt.Errorf("failed to read zlib data: %v", err)
		}

	default:
		return nil, fmt.Errorf("unsupported compression method: %d", method)
	}

	// Verify decompressed size
	if uint32(buf.Len()) != originalSize {
		return nil, fmt.Errorf("decompressed size mismatch: expected %d, got %d", originalSize, buf.Len())
	}

	return buf.Bytes(), nil
}

// GetCompressionStats returns statistics about the compression
func (uc *UTXOCompression) GetCompressionStats(data []byte) (*CompressionStats, error) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	// Compress data
	compressed, err := uc.Compress(data)
	if err != nil {
		return nil, err
	}

	// Calculate compression ratio
	ratio := float64(len(compressed)) / float64(len(data))

	// Calculate hash
	hash := sha256.Sum256(data)

	return &CompressionStats{
		OriginalSize:     int64(len(data)),
		CompressedSize:   int64(len(compressed)),
		CompressionRatio: ratio,
		Method:           uc.method,
		Level:            uc.level,
		Threshold:        uc.threshold,
		Hash:             hash[:],
	}, nil
}

// SetCompressionMethod sets the compression method
func (uc *UTXOCompression) SetCompressionMethod(method CompressionMethod) {
	uc.mu.Lock()
	uc.method = method
	uc.mu.Unlock()
}

// SetCompressionLevel sets the compression level
func (uc *UTXOCompression) SetCompressionLevel(level int) {
	uc.mu.Lock()
	if level < 1 {
		level = 1
	} else if level > 9 {
		level = 9
	}
	uc.level = level
	uc.mu.Unlock()
}

// SetCompressionThreshold sets the compression threshold
func (uc *UTXOCompression) SetCompressionThreshold(threshold int64) {
	uc.mu.Lock()
	uc.threshold = threshold
	uc.mu.Unlock()
}

// CompressionStats holds statistics about the compression
type CompressionStats struct {
	// OriginalSize is the size of the original data in bytes
	OriginalSize int64
	// CompressedSize is the size of the compressed data in bytes
	CompressedSize int64
	// CompressionRatio is the ratio of compressed size to original size
	CompressionRatio float64
	// Method is the compression method used
	Method CompressionMethod
	// Level is the compression level used
	Level int
	// Threshold is the compression threshold
	Threshold int64
	// Hash is the SHA-256 hash of the original data
	Hash []byte
}

// String returns a string representation of the compression statistics
func (cs *CompressionStats) String() string {
	return fmt.Sprintf(
		"Original Size: %d bytes\n"+
			"Compressed Size: %d bytes\n"+
			"Compression Ratio: %.2f\n"+
			"Method: %d, Level: %d\n"+
			"Threshold: %d bytes",
		cs.OriginalSize, cs.CompressedSize,
		cs.CompressionRatio, cs.Method, cs.Level,
		cs.Threshold,
	)
}
