package script

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/youngchain/internal/core/block"
)

// OpCode represents a Bitcoin Script operation code
type OpCode byte

const (
	// Stack operations
	OP_0                   OpCode = 0x00
	OP_1                   OpCode = 0x51
	OP_2                   OpCode = 0x52
	OP_3                   OpCode = 0x53
	OP_4                   OpCode = 0x54
	OP_5                   OpCode = 0x55
	OP_6                   OpCode = 0x56
	OP_7                   OpCode = 0x57
	OP_8                   OpCode = 0x58
	OP_9                   OpCode = 0x59
	OP_10                  OpCode = 0x5a
	OP_11                  OpCode = 0x5b
	OP_12                  OpCode = 0x5c
	OP_13                  OpCode = 0x5d
	OP_14                  OpCode = 0x5e
	OP_15                  OpCode = 0x5f
	OP_16                  OpCode = 0x60
	OP_DUP                 OpCode = 0x76
	OP_HASH160             OpCode = 0xa9
	OP_EQUAL               OpCode = 0x87
	OP_EQUALVERIFY         OpCode = 0x88
	OP_CHECKSIG            OpCode = 0xac
	OP_CHECKMULTISIG       OpCode = 0xae
	OP_PUSHDATA1           OpCode = 0x4c
	OP_PUSHDATA2           OpCode = 0x4d
	OP_PUSHDATA4           OpCode = 0x4e
	OP_DROP                OpCode = 0x75
	OP_CHECKLOCKTIMEVERIFY OpCode = 0xb1
)

// Script represents a Bitcoin Script
type Script struct {
	Instructions []Instruction
}

// Instruction represents a single script instruction
type Instruction struct {
	OpCode OpCode
	Data   []byte
}

// NewScript creates a new script
func NewScript() *Script {
	return &Script{
		Instructions: make([]Instruction, 0),
	}
}

// AddOp adds an operation to the script
func (s *Script) AddOp(op OpCode) {
	s.Instructions = append(s.Instructions, Instruction{OpCode: op})
}

// AddData adds data to the script
func (s *Script) AddData(data []byte) error {
	if len(data) <= 75 {
		s.Instructions = append(s.Instructions, Instruction{
			OpCode: OpCode(len(data)),
			Data:   data,
		})
	} else if len(data) <= 255 {
		s.Instructions = append(s.Instructions, Instruction{
			OpCode: OP_PUSHDATA1,
			Data:   append([]byte{byte(len(data))}, data...),
		})
	} else if len(data) <= 65535 {
		lenBytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(lenBytes, uint16(len(data)))
		s.Instructions = append(s.Instructions, Instruction{
			OpCode: OP_PUSHDATA2,
			Data:   append(lenBytes, data...),
		})
	} else {
		lenBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(lenBytes, uint32(len(data)))
		s.Instructions = append(s.Instructions, Instruction{
			OpCode: OP_PUSHDATA4,
			Data:   append(lenBytes, data...),
		})
	}
	return nil
}

// CreateP2PKHScript creates a Pay-to-Public-Key-Hash script
func CreateP2PKHScript(pubKeyHash []byte) *Script {
	script := NewScript()
	script.AddOp(OP_DUP)
	script.AddOp(OP_HASH160)
	script.AddData(pubKeyHash)
	script.AddOp(OP_EQUALVERIFY)
	script.AddOp(OP_CHECKSIG)
	return script
}

// CreateP2SHScript creates a Pay-to-Script-Hash script
func CreateP2SHScript(scriptHash []byte) *Script {
	script := NewScript()
	script.AddOp(OP_HASH160)
	script.AddData(scriptHash)
	script.AddOp(OP_EQUAL)
	return script
}

// CreateMultiSigScript creates a multi-signature script
func CreateMultiSigScript(required int, pubKeys []*ecdsa.PublicKey) (*Script, error) {
	if required <= 0 || required > len(pubKeys) {
		return nil, errors.New("invalid required signatures")
	}

	script := NewScript()
	script.AddOp(OpCode(required + 80)) // OP_1 to OP_16
	for _, pubKey := range pubKeys {
		pubKeyBytes := append(pubKey.X.Bytes(), pubKey.Y.Bytes()...)
		script.AddData(pubKeyBytes)
	}
	script.AddOp(OpCode(len(pubKeys) + 80)) // OP_1 to OP_16
	script.AddOp(OP_CHECKMULTISIG)
	return script, nil
}

// CreateTimeLockedScript creates a time-locked script
func CreateTimeLockedScript(lockTime uint32, script *Script) *Script {
	result := NewScript()
	result.AddOp(OpCode(lockTime))
	result.AddOp(OP_CHECKLOCKTIMEVERIFY)
	result.AddOp(OP_DROP)
	for _, inst := range script.Instructions {
		result.Instructions = append(result.Instructions, inst)
	}
	return result
}

// Execute executes the script
func (s *Script) Execute(stack [][]byte, tx *block.TransactionWrapper, inputIndex int) error {
	for _, inst := range s.Instructions {
		switch inst.OpCode {
		case OP_DUP:
			if len(stack) < 1 {
				return errors.New("stack underflow")
			}
			stack = append(stack, stack[len(stack)-1])

		case OP_HASH160:
			if len(stack) < 1 {
				return errors.New("stack underflow")
			}
			hash := sha256.Sum256(stack[len(stack)-1])
			hash160 := sha256.Sum256(hash[:])
			stack[len(stack)-1] = hash160[:]

		case OP_EQUAL:
			if len(stack) < 2 {
				return errors.New("stack underflow")
			}
			a := stack[len(stack)-2]
			b := stack[len(stack)-1]
			stack = stack[:len(stack)-2]
			if bytes.Equal(a, b) {
				stack = append(stack, []byte{1})
			} else {
				stack = append(stack, []byte{0})
			}

		case OP_EQUALVERIFY:
			if len(stack) < 2 {
				return errors.New("stack underflow")
			}
			a := stack[len(stack)-2]
			b := stack[len(stack)-1]
			stack = stack[:len(stack)-2]
			if !bytes.Equal(a, b) {
				return errors.New("OP_EQUALVERIFY failed")
			}

		case OP_CHECKSIG:
			if len(stack) < 2 {
				return errors.New("stack underflow")
			}
			stack = stack[:len(stack)-2]

			if !verifySignature() {
				return errors.New("signature verification failed")
			}

			stack = append(stack, []byte{1})

		case OP_CHECKMULTISIG:
			if len(stack) < 3 {
				return errors.New("stack underflow")
			}

			// Get the number of public keys and required signatures
			n := int(stack[len(stack)-1][0])
			m := int(stack[len(stack)-2][0])
			if m > n {
				return errors.New("invalid number of required signatures")
			}

			// Get the public keys and signatures
			pubKeys := make([]*ecdsa.PublicKey, n)
			sigs := make([][]byte, len(stack)-3)
			for i := 0; i < n; i++ {
				pubKey, err := bytesToPublicKey()
				if err != nil {
					return fmt.Errorf("invalid public key: %v", err)
				}
				pubKeys[i] = pubKey
			}
			for i := 0; i < len(stack)-3; i++ {
				sigs[i] = stack[i]
			}

			// Verify signatures
			validSigs := 0
			for range sigs {
				for range pubKeys {
					if verifySignature() {
						validSigs++
						break
					}
				}
			}

			// Clear the stack
			stack = stack[:0]

			if validSigs >= m {
				stack = append(stack, []byte{1})
			} else {
				stack = append(stack, []byte{0})
			}

		case OP_CHECKLOCKTIMEVERIFY:
			if len(stack) < 1 {
				return errors.New("stack underflow")
			}

			lockTime := binary.LittleEndian.Uint32(stack[len(stack)-1])
			if lockTime > tx.LockTime() {
				return errors.New("lock time not satisfied")
			}

		default:
			if inst.OpCode <= OP_16 {
				stack = append(stack, []byte{byte(inst.OpCode - 80)})
			} else if inst.OpCode == OP_PUSHDATA1 || inst.OpCode == OP_PUSHDATA2 || inst.OpCode == OP_PUSHDATA4 {
				stack = append(stack, inst.Data)
			} else {
				return fmt.Errorf("unsupported opcode: %x", inst.OpCode)
			}
		}
	}

	return nil
}

// bytesToPublicKey converts bytes to an ECDSA public key
func bytesToPublicKey() (*ecdsa.PublicKey, error) {
	// TODO: Implement proper public key parsing
	return nil, errors.New("not implemented")
}

// verifySignature verifies an ECDSA signature
func verifySignature() bool {
	// TODO: Implement proper signature verification
	return false
}

// Validate validates the script
func (s *Script) Validate() error {
	if len(s.Instructions) == 0 {
		return errors.New("empty script")
	}
	return nil
}

// Clone creates a deep copy of the script
func (s *Script) Clone() *Script {
	clone := &Script{
		Instructions: make([]Instruction, len(s.Instructions)),
	}
	for i, inst := range s.Instructions {
		clone.Instructions[i] = Instruction{
			OpCode: inst.OpCode,
			Data:   append([]byte{}, inst.Data...),
		}
	}
	return clone
}

// MatchesAddress checks if the script matches an address
func (s *Script) MatchesAddress(address string) bool {
	// TODO: Implement address matching
	return false
}

// Serialize serializes the script
func (s *Script) Serialize() []byte {
	data := make([]byte, 0)
	for _, inst := range s.Instructions {
		data = append(data, byte(inst.OpCode))
		if inst.Data != nil {
			data = append(data, inst.Data...)
		}
	}
	return data
}
