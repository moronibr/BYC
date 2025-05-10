package witness

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

const (
	// WitnessVersion is the current witness version
	WitnessVersion = 0

	// MaxWitnessSize is the maximum size of a witness program
	MaxWitnessSize = 40

	// P2WPKHSize is the size of a P2WPKH witness program
	P2WPKHSize = 20

	// P2WSHSize is the size of a P2WSH witness program
	P2WSHSize = 32
)

// WitnessProgram represents a witness program
type WitnessProgram struct {
	Version byte
	Program []byte
}

// Witness represents transaction witness data
type Witness struct {
	ScriptSig    []byte
	ScriptPubKey []byte
	Sequence     uint32
	LockTime     uint32
	Timestamp    time.Time
}

// NewWitness creates a new witness
func NewWitness(scriptSig, scriptPubKey []byte) *Witness {
	return &Witness{
		ScriptSig:    scriptSig,
		ScriptPubKey: scriptPubKey,
		Sequence:     0xffffffff,
		LockTime:     0,
		Timestamp:    time.Now(),
	}
}

// Size returns the size of the witness in bytes
func (w *Witness) Size() int {
	size := 0

	// ScriptSig size
	size += len(w.ScriptSig)

	// ScriptPubKey size
	size += len(w.ScriptPubKey)

	// Sequence size
	size += 4

	// LockTime size
	size += 4

	// Timestamp size
	size += 8

	return size
}

// Hash returns the hash of the witness
func (w *Witness) Hash() []byte {
	hash := sha256.New()

	// Hash ScriptSig
	hash.Write(w.ScriptSig)

	// Hash ScriptPubKey
	hash.Write(w.ScriptPubKey)

	// Hash Sequence
	seqBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(seqBytes, w.Sequence)
	hash.Write(seqBytes)

	// Hash LockTime
	lockBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(lockBytes, w.LockTime)
	hash.Write(lockBytes)

	// Hash Timestamp
	timeBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(timeBytes, uint64(w.Timestamp.Unix()))
	hash.Write(timeBytes)

	return hash.Sum(nil)
}

// Validate validates the witness
func (w *Witness) Validate() error {
	// Validate ScriptSig
	if len(w.ScriptSig) == 0 {
		return ErrInvalidScriptSig
	}

	// Validate ScriptPubKey
	if len(w.ScriptPubKey) == 0 {
		return ErrInvalidScriptPubKey
	}

	// Validate Sequence
	if w.Sequence == 0 {
		return ErrInvalidSequence
	}

	// Validate LockTime
	if w.LockTime < 0 {
		return ErrInvalidLockTime
	}

	// Validate Timestamp
	if w.Timestamp.IsZero() {
		return ErrInvalidTimestamp
	}

	return nil
}

// Errors
var (
	ErrInvalidScriptSig    = errors.New("invalid script signature")
	ErrInvalidScriptPubKey = errors.New("invalid script public key")
	ErrInvalidSequence     = errors.New("invalid sequence")
	ErrInvalidLockTime     = errors.New("invalid lock time")
	ErrInvalidTimestamp    = errors.New("invalid timestamp")
)

// CreateP2WPKHWitness creates a P2WPKH witness
func CreateP2WPKHWitness(signature []byte, pubKey []byte) *Witness {
	return NewWitness(signature, pubKey)
}

// CreateP2WSHWitness creates a P2WSH witness
func CreateP2WSHWitness(signatures [][]byte, redeemScript []byte) *Witness {
	return NewWitness(signatures[0], redeemScript)
}

// CreateWitnessProgram creates a witness program
func CreateWitnessProgram(version byte, program []byte) (*WitnessProgram, error) {
	if version != WitnessVersion {
		return nil, fmt.Errorf("unsupported witness version: %d", version)
	}

	if len(program) > MaxWitnessSize {
		return nil, fmt.Errorf("program too large: %d bytes", len(program))
	}

	return &WitnessProgram{
		Version: version,
		Program: program,
	}, nil
}

// CreateP2WPKHProgram creates a P2WPKH witness program
func CreateP2WPKHProgram(pubKeyHash []byte) (*WitnessProgram, error) {
	if len(pubKeyHash) != P2WPKHSize {
		return nil, fmt.Errorf("invalid P2WPKH hash size: %d", len(pubKeyHash))
	}

	return CreateWitnessProgram(WitnessVersion, pubKeyHash)
}

// CreateP2WSHProgram creates a P2WSH witness program
func CreateP2WSHProgram(scriptHash []byte) (*WitnessProgram, error) {
	if len(scriptHash) != P2WSHSize {
		return nil, fmt.Errorf("invalid P2WSH hash size: %d", len(scriptHash))
	}

	return CreateWitnessProgram(WitnessVersion, scriptHash)
}

// Hash160 performs a RIPEMD160(SHA256(data)) hash
func Hash160(data []byte) []byte {
	hash := sha256.Sum256(data)
	// TODO: Implement RIPEMD160
	return hash[:]
}

// Serialize serializes the witness program
func (wp *WitnessProgram) Serialize() []byte {
	data := make([]byte, 0)
	data = append(data, wp.Version)
	data = append(data, byte(len(wp.Program)))
	data = append(data, wp.Program...)
	return data
}

// Deserialize deserializes a witness program
func Deserialize(data []byte) (*WitnessProgram, error) {
	if len(data) < 2 {
		return nil, errors.New("data too short")
	}

	version := data[0]
	programLen := int(data[1])
	if len(data) < 2+programLen {
		return nil, errors.New("data too short")
	}

	program := data[2 : 2+programLen]
	return CreateWitnessProgram(version, program)
}

// IsP2WPKH checks if the witness program is P2WPKH
func (wp *WitnessProgram) IsP2WPKH() bool {
	return wp.Version == WitnessVersion && len(wp.Program) == P2WPKHSize
}

// IsP2WSH checks if the witness program is P2WSH
func (wp *WitnessProgram) IsP2WSH() bool {
	return wp.Version == WitnessVersion && len(wp.Program) == P2WSHSize
}

// Clone creates a deep copy of the witness
func (w *Witness) Clone() *Witness {
	clone := &Witness{
		ScriptSig:    append([]byte{}, w.ScriptSig...),
		ScriptPubKey: append([]byte{}, w.ScriptPubKey...),
		Sequence:     w.Sequence,
		LockTime:     w.LockTime,
		Timestamp:    w.Timestamp,
	}
	return clone
}
