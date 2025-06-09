package network

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"sync"
	"time"

	"byc/internal/logger"

	"go.uber.org/zap"
)

// Stream represents a multiplexed stream
type Stream struct {
	ID        uint32
	Priority  uint8
	Direction bool // true for outgoing, false for incoming
	Data      chan []byte
	Error     chan error
	Closed    chan struct{}
}

// MultiplexedConn represents a multiplexed connection
type MultiplexedConn struct {
	conn        net.Conn
	streams     map[uint32]*Stream
	mu          sync.RWMutex
	nextID      uint32
	compression bool
}

// NewMultiplexedConn creates a new multiplexed connection
func NewMultiplexedConn(conn net.Conn, compression bool) *MultiplexedConn {
	return &MultiplexedConn{
		conn:        conn,
		streams:     make(map[uint32]*Stream),
		nextID:      1,
		compression: compression,
	}
}

// CreateStream creates a new stream
func (mc *MultiplexedConn) CreateStream(priority uint8) *Stream {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	stream := &Stream{
		ID:        mc.nextID,
		Priority:  priority,
		Direction: true,
		Data:      make(chan []byte, 100),
		Error:     make(chan error, 1),
		Closed:    make(chan struct{}),
	}

	mc.streams[stream.ID] = stream
	mc.nextID++

	return stream
}

// AcceptStream accepts a new stream
func (mc *MultiplexedConn) AcceptStream(id uint32, priority uint8) *Stream {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	stream := &Stream{
		ID:        id,
		Priority:  priority,
		Direction: false,
		Data:      make(chan []byte, 100),
		Error:     make(chan error, 1),
		Closed:    make(chan struct{}),
	}

	mc.streams[id] = stream
	return stream
}

// CloseStream closes a stream
func (mc *MultiplexedConn) CloseStream(id uint32) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if stream, ok := mc.streams[id]; ok {
		close(stream.Closed)
		delete(mc.streams, id)
	}
}

// Write writes data to a stream
func (mc *MultiplexedConn) Write(streamID uint32, data []byte) error {
	mc.mu.RLock()
	stream, ok := mc.streams[streamID]
	mc.mu.RUnlock()

	if !ok {
		return fmt.Errorf("stream %d not found", streamID)
	}

	if mc.compression {
		var buf bytes.Buffer
		zw := gzip.NewWriter(&buf)
		if _, err := zw.Write(data); err != nil {
			return fmt.Errorf("failed to compress data: %v", err)
		}
		if err := zw.Close(); err != nil {
			return fmt.Errorf("failed to close gzip writer: %v", err)
		}
		data = buf.Bytes()
	}

	header := make([]byte, 9)
	binary.BigEndian.PutUint32(header[0:4], streamID)
	header[4] = stream.Priority
	binary.BigEndian.PutUint32(header[5:9], uint32(len(data)))

	if _, err := mc.conn.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %v", err)
	}

	if _, err := mc.conn.Write(data); err != nil {
		return fmt.Errorf("failed to write data: %v", err)
	}

	return nil
}

// Read reads data from a stream
func (mc *MultiplexedConn) Read(streamID uint32) ([]byte, error) {
	mc.mu.RLock()
	stream, ok := mc.streams[streamID]
	mc.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("stream %d not found", streamID)
	}

	select {
	case data := <-stream.Data:
		if mc.compression {
			zr, err := gzip.NewReader(bytes.NewReader(data))
			if err != nil {
				return nil, fmt.Errorf("failed to create gzip reader: %v", err)
			}
			defer zr.Close()

			uncompressed, err := ioutil.ReadAll(zr)
			if err != nil {
				return nil, fmt.Errorf("failed to decompress data: %v", err)
			}
			return uncompressed, nil
		}
		return data, nil
	case err := <-stream.Error:
		return nil, err
	case <-stream.Closed:
		return nil, io.EOF
	}
}

// Start starts the multiplexer
func (mc *MultiplexedConn) Start() {
	go mc.readLoop()
}

// readLoop reads data from the connection and routes it to the appropriate stream
func (mc *MultiplexedConn) readLoop() {
	for {
		header := make([]byte, 9)
		if _, err := io.ReadFull(mc.conn, header); err != nil {
			if err == io.EOF {
				return
			}
			logger.Error("failed to read header", zap.Error(err))
			continue
		}

		streamID := binary.BigEndian.Uint32(header[0:4])
		priority := header[4]
		length := binary.BigEndian.Uint32(header[5:9])

		data := make([]byte, length)
		if _, err := io.ReadFull(mc.conn, data); err != nil {
			logger.Error("failed to read data", zap.Error(err))
			continue
		}

		mc.mu.RLock()
		stream, ok := mc.streams[streamID]
		mc.mu.RUnlock()

		if !ok {
			stream = mc.AcceptStream(streamID, priority)
		}

		select {
		case stream.Data <- data:
		case <-stream.Closed:
			continue
		default:
			logger.Warn("stream buffer full, dropping message",
				zap.Uint32("streamID", streamID))
		}
	}
}

// Close closes the multiplexed connection
func (mc *MultiplexedConn) Close() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	for _, stream := range mc.streams {
		close(stream.Closed)
	}

	return mc.conn.Close()
}

// SetReadDeadline sets the read deadline
func (mc *MultiplexedConn) SetReadDeadline(t time.Time) error {
	return mc.conn.SetReadDeadline(t)
}

// SetWriteDeadline sets the write deadline
func (mc *MultiplexedConn) SetWriteDeadline(t time.Time) error {
	return mc.conn.SetWriteDeadline(t)
}

// LocalAddr returns the local address
func (mc *MultiplexedConn) LocalAddr() net.Addr {
	return mc.conn.LocalAddr()
}

// RemoteAddr returns the remote address
func (mc *MultiplexedConn) RemoteAddr() net.Addr {
	return mc.conn.RemoteAddr()
}
