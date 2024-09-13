package starknet

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
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
	RangeCheck96
)

func (b Builtin) MarshalJSON() ([]byte, error) {
	switch b {
	case Output:
		return []byte("output"), nil
	case RangeCheck:
		return []byte("range_check"), nil
	case RangeCheck96:
		return []byte("range_check96"), nil
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
	return nil, fmt.Errorf("marshal unknown builtin: %d", uint8(b))
}

func (b *Builtin) UnmarshalJSON(data []byte) error {
	builtinName, err := strconv.Unquote(string(data))
	if err != nil {
		return fmt.Errorf("unmarshal builtin: %w", err)
	}

	switch builtinName {
	case "output":
		*b = Output
	case "range_check":
		*b = RangeCheck
	case "range_check96":
		*b = RangeCheck96
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
		return fmt.Errorf("unmarshal unknown builtin: %s", builtinName)
	}
	return nil
}

type EntryPointInfo struct {
	Selector fp.Element `json:"selector"`
	Offset   fp.Element `json:"offset"`
	Builtins []Builtin  `json:"builtins"`
}

type EntryPointByType struct {
	External    []EntryPointInfo `json:"EXTERNAL"`
	L1Handler   []EntryPointInfo `json:"L1_HANDLER"`
	Constructor []EntryPointInfo `json:"CONSTRUCTOR"`
}

type Hints struct {
	Index uint64
	Hints []Hint
}

// Hints are serialized as tuples of (index, []hint)
// https://github.com/starkware-libs/cairo/blob/main/crates/cairo-lang-starknet/src/casm_contract_class.rs#L90
func (hints *Hints) UnmarshalJSON(data []byte) error {
	var rawHints []any
	if err := json.Unmarshal(data, &rawHints); err != nil {
		return err
	}

	index, ok := rawHints[0].(float64)
	if !ok {
		return fmt.Errorf("unmarshal hints: index should be uint64")
	}
	hints.Index = uint64(index)

	rest, err := json.Marshal(rawHints[1])
	if err != nil {
		return err
	}

	var h []Hint
	if err := json.Unmarshal(rest, &h); err != nil {
		return err
	}
	hints.Hints = h
	return nil
}

func (hints *Hints) MarshalJSON() ([]byte, error) {
	var rawHints []any
	rawHints = append(rawHints, hints.Index)
	rawHints = append(rawHints, hints.Hints)

	return json.Marshal(rawHints)
}

type StarknetProgram struct {
	// Prime is fixed to be 0x800000000000011000000000000000000000000000000000000000000000001 and wont fit in a f.Felt
	Bytecode        []fp.Element     `json:"bytecode"`
	CompilerVersion string           `json:"compiler_version"`
	EntryPoints     EntryPointByType `json:"entry_points_by_type"`
	Hints           []Hints          `json:"hints" validate:"required"`
}

func StarknetProgramFromFile(pathToFile string) (*StarknetProgram, error) {
	content, error := os.ReadFile(pathToFile)
	if error != nil {
		return nil, error
	}
	return StarknetProgramFromJSON(content)
}

func StarknetProgramFromJSON(content json.RawMessage) (*StarknetProgram, error) {
	var starknet StarknetProgram
	return &starknet, json.Unmarshal(content, &starknet)
}
