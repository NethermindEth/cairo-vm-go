package starknet

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	builtins "github.com/NethermindEth/cairo-vm-go/pkg/vm/builtins"
)

type EntryPointInfo struct {
	Selector fp.Element `json:"selector"`
	Offset   fp.Element `json:"offset"`
	Builtins []builtins.Builtin  `json:"builtins"`
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
