package network

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/common"
)

// MessageType represents the type of network message
type MessageType uint32

const (
	// Message types
	MsgVersion         MessageType = 0x01
	MsgVerAck          MessageType = 0x02
	MsgInv             MessageType = 0x03
	MsgGetData         MessageType = 0x04
	MsgBlock           MessageType = 0x05
	MsgTx              MessageType = 0x06
	MsgGetBlocks       MessageType = 0x07
	MsgHeaders         MessageType = 0x08
	MsgCompactBlock    MessageType = 0x09
	MsgGetCompactBlock MessageType = 0x0A
	MsgCompactBlockTx  MessageType = 0x0B
)

// Message represents a network message
type Message struct {
	Magic    uint32
	Type     MessageType
	Length   uint32
	Checksum [4]byte
	Payload  []byte
}

// CompactBlock represents a compact block as per BIP152
type CompactBlock struct {
	Header       *common.Header
	Nonce        uint64
	ShortIDs     []uint64
	PrefilledTxs []PrefilledTx
}

// PrefilledTx represents a prefilled transaction in a compact block
type PrefilledTx struct {
	Index uint32
	Tx    *common.Transaction
}

// NewMessage creates a new network message
func NewMessage(magic uint32, msgType MessageType, payload []byte) *Message {
	msg := &Message{
		Magic:   magic,
		Type:    msgType,
		Length:  uint32(len(payload)),
		Payload: payload,
	}
	msg.Checksum = msg.CalculateChecksum()
	return msg
}

// Serialize serializes the message to bytes
func (m *Message) Serialize() []byte {
	var buf bytes.Buffer

	// Write magic
	binary.Write(&buf, binary.LittleEndian, m.Magic)

	// Write command
	cmd := make([]byte, 12)
	copy(cmd, []byte(m.Type.String()))
	buf.Write(cmd)

	// Write length
	binary.Write(&buf, binary.LittleEndian, m.Length)

	// Write checksum
	buf.Write(m.Checksum[:])

	// Write payload
	buf.Write(m.Payload)

	return buf.Bytes()
}

// CalculateChecksum calculates the checksum of the message
func (m *Message) CalculateChecksum() [4]byte {
	hash := sha256.Sum256(m.Payload)
	hash = sha256.Sum256(hash[:])
	var checksum [4]byte
	copy(checksum[:], hash[:4])
	return checksum
}

// String returns a string representation of the message type
func (t MessageType) String() string {
	switch t {
	case MsgVersion:
		return "version"
	case MsgVerAck:
		return "verack"
	case MsgInv:
		return "inv"
	case MsgGetData:
		return "getdata"
	case MsgBlock:
		return "block"
	case MsgTx:
		return "tx"
	case MsgGetBlocks:
		return "getblocks"
	case MsgHeaders:
		return "headers"
	case MsgCompactBlock:
		return "cmpctblock"
	case MsgGetCompactBlock:
		return "getcmpctblock"
	case MsgCompactBlockTx:
		return "cmpctblocktx"
	default:
		return fmt.Sprintf("unknown(%d)", t)
	}
}

// CreateCompactBlock creates a compact block from a full block
func CreateCompactBlock(block *block.Block, nonce uint64) *CompactBlock {
	cb := &CompactBlock{
		Header:   block.Header,
		Nonce:    nonce,
		ShortIDs: make([]uint64, 0, len(block.Transactions)),
	}

	// Calculate short IDs for each transaction
	for _, tx := range block.Transactions {
		shortID := CalculateShortID(tx.Hash, nonce)
		cb.ShortIDs = append(cb.ShortIDs, shortID)
	}

	return cb
}

// CalculateShortID calculates a short ID for a transaction as per BIP152
func CalculateShortID(txHash []byte, nonce uint64) uint64 {
	// Create a 64-bit hash using SHA-256 (since we don't have SipHash)
	key := make([]byte, 16)
	binary.LittleEndian.PutUint64(key[0:8], nonce)
	binary.LittleEndian.PutUint64(key[8:16], nonce>>8)

	// Use SHA-256 to create a 64-bit hash
	hash := sha256.Sum256(append(key, txHash...))
	return binary.LittleEndian.Uint64(hash[:8])
}

// Serialize serializes the compact block to bytes
func (cb *CompactBlock) Serialize() []byte {
	var buf bytes.Buffer

	// Write header fields
	binary.Write(&buf, binary.LittleEndian, cb.Header.Version)
	buf.Write(cb.Header.PrevBlockHash)
	buf.Write(cb.Header.MerkleRoot)
	binary.Write(&buf, binary.LittleEndian, cb.Header.Timestamp.Unix())
	binary.Write(&buf, binary.LittleEndian, cb.Header.Difficulty)
	binary.Write(&buf, binary.LittleEndian, cb.Header.Nonce)
	binary.Write(&buf, binary.LittleEndian, cb.Header.Height)
	buf.Write(cb.Header.Hash)

	// Write nonce
	binary.Write(&buf, binary.LittleEndian, cb.Nonce)

	// Write short IDs
	binary.Write(&buf, binary.LittleEndian, uint32(len(cb.ShortIDs)))
	for _, id := range cb.ShortIDs {
		binary.Write(&buf, binary.LittleEndian, id)
	}

	// Write prefilled transactions
	binary.Write(&buf, binary.LittleEndian, uint32(len(cb.PrefilledTxs)))
	for _, ptx := range cb.PrefilledTxs {
		binary.Write(&buf, binary.LittleEndian, ptx.Index)

		// Serialize transaction
		var txBuf bytes.Buffer
		binary.Write(&txBuf, binary.LittleEndian, ptx.Tx.Version)
		binary.Write(&txBuf, binary.LittleEndian, ptx.Tx.Timestamp.Unix())
		txBuf.Write(ptx.Tx.From)
		txBuf.Write(ptx.Tx.To)
		binary.Write(&txBuf, binary.LittleEndian, ptx.Tx.Amount)
		txBuf.Write(ptx.Tx.Data)

		// Write inputs
		binary.Write(&txBuf, binary.LittleEndian, uint32(len(ptx.Tx.Inputs)))
		for _, input := range ptx.Tx.Inputs {
			txBuf.Write(input.PreviousTxHash)
			binary.Write(&txBuf, binary.LittleEndian, input.PreviousTxIndex)
			txBuf.Write(input.ScriptSig)
			binary.Write(&txBuf, binary.LittleEndian, input.Sequence)
		}

		// Write outputs
		binary.Write(&txBuf, binary.LittleEndian, uint32(len(ptx.Tx.Outputs)))
		for _, output := range ptx.Tx.Outputs {
			binary.Write(&txBuf, binary.LittleEndian, output.Value)
			txBuf.Write(output.ScriptPubKey)
			binary.Write(&txBuf, binary.LittleEndian, uint32(len(output.Address)))
			txBuf.Write([]byte(output.Address))
		}

		// Write transaction size and data
		txBytes := txBuf.Bytes()
		binary.Write(&buf, binary.LittleEndian, uint32(len(txBytes)))
		buf.Write(txBytes)
	}

	return buf.Bytes()
}
