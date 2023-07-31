package vm

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	f "github.com/NethermindEth/juno/core/felt"
)

type Builtin uint8

const (
	Output Builtin = iota + 1
	RangeCheck
	Pedersen
	ECDSA
	Keccak
	Bitwise
	ECOP
	Poseidon
	SegmentArena
)

func (b Builtin) MarshalJSON() ([]byte, error) {
	switch b {
	case Output:
		return []byte("output"), nil
	case RangeCheck:
		return []byte("range_check"), nil
	case Pedersen:
		return []byte("pedersen"), nil
	case ECDSA:
		return []byte("ecdsa"), nil
	case Keccak:
		return []byte("keccak"), nil
	case Bitwise:
		return []byte("bitwise"), nil
	case ECOP:
		return []byte("ec_op"), nil
	case Poseidon:
		return []byte("poseidon"), nil
	case SegmentArena:
		return []byte("segment_arena"), nil

	}
	return nil, fmt.Errorf("Error marshaling builtin with unknow identifer: %d", uint8(b))
}

func (b *Builtin) UnmarshalJSON(data []byte) error {
	builtinName, err := strconv.Unquote(string(data))
    if err != nil {
        return fmt.Errorf("Error unmarsahling builtin: %w", err)
    }

	switch builtinName {
	case "output":
		*b = Output
	case "range_check":
		*b = RangeCheck
	case "pedersen":
		*b = Pedersen
	case "ecdsa":
		*b = ECDSA
	case "keccak":
		*b = Keccak
	case "bitwise":
		*b = Bitwise
	case "ec_op":
		*b = ECOP
	case "poseidon":
		*b = Poseidon
	case "segment_arena":
		*b = SegmentArena
    default:
	    return fmt.Errorf("Error unmarsahling unknwon builtin name: %s", builtinName)
	}
    return nil
}

type EntryPointInfo struct {
	Selector f.Felt    `json:"selector"`
	Offset   f.Felt    `json:"offset"`
	Builtins []Builtin `json:"builtins"`
}

type EntryPointByType struct {
	External    []EntryPointInfo `json:"EXTERNAL"`
	L1Handler   []EntryPointInfo `json:"L1_HANDLER"`
	Constructor []EntryPointInfo `json:"CONSTRUCTOR"`
}

type Program struct {
	// Prime is fixed to be 0x800000000000011000000000000000000000000000000000000000000000001 and wont fit in a f.Felt
	Bytecode        []f.Felt         `json:"bytecode"`
	CompilerVersion string           `json:"compiler_version"`
	EntryPoints     EntryPointByType `json:"entry_points_by_type"`
	// todo(rodro): Add remaining Json field
	// Hints
}

func ProgramFromFile(pathToFile string) (*Program, error) {
	content, error := os.ReadFile(pathToFile)
	if error != nil {
		return nil, error
	}
	return ProgramFromJSON(content)
}

func ProgramFromJSON(content json.RawMessage) (*Program, error) {
	var program Program
	return &program, json.Unmarshal(content, &program)
}
