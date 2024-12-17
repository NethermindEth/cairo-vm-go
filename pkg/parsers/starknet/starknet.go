package starknet

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/builtins"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

type EntryPointByTypeInfo struct {
	Selector fp.Element             `json:"selector"`
	Offset   fp.Element             `json:"offset"`
	Builtins []builtins.BuiltinType `json:"builtins"`
}

type EntryPointByType struct {
	External    []EntryPointByTypeInfo `json:"EXTERNAL"`
	L1Handler   []EntryPointByTypeInfo `json:"L1_HANDLER"`
	Constructor []EntryPointByTypeInfo `json:"CONSTRUCTOR"`
}

type Arg struct {
	GenericID string `json:"generic_id"`
	Size      int    `json:"size"`
	DebugName string `json:"debug_name"`
}

type EntryPointByFunction struct {
	Offset    int                    `json:"offset"`
	Builtins  []builtins.BuiltinType `json:"builtins"`
	InputArgs []Arg                  `json:"input_args"`
	ReturnArg []Arg                  `json:"return_arg"`
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
	Bytecode              []fp.Element                    `json:"bytecode"`
	CompilerVersion       string                          `json:"compiler_version"`
	EntryPointsByType     EntryPointByType                `json:"entry_points_by_type"`
	EntryPointsByFunction map[string]EntryPointByFunction `json:"entry_points_by_function"`
	Hints                 []Hints                         `json:"hints" validate:"required"`
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

type CairoFuncArgs struct {
	Single *fp.Element
	Array  []fp.Element
}

func ParseCairoProgramArgs(input string) ([]CairoFuncArgs, error) {
	re := regexp.MustCompile(`\[[^\]]*\]|\S+`)
	tokens := re.FindAllString(input, -1)
	var result []CairoFuncArgs
	for _, token := range tokens {
		if single, err := new(fp.Element).SetString(token); err == nil {
			result = append(result, CairoFuncArgs{
				Single: single,
				Array:  nil,
			})
		} else if strings.HasPrefix(token, "[") && strings.HasSuffix(token, "]") {
			arrayStr := strings.Trim(token, "[]")
			arrayParts := strings.Fields(arrayStr)
			var array []fp.Element
			for _, part := range arrayParts {
				single, err := new(fp.Element).SetString(part)
				if err != nil {
					return nil, fmt.Errorf("invalid felt value in array: %v", err)
				}
				array = append(array, *single)
			}
			result = append(result, CairoFuncArgs{
				Single: nil,
				Array:  array,
			})
		} else {
			return nil, fmt.Errorf("invalid token: %s", token)
		}
	}

	return result, nil
}
