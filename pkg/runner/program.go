package runner

import (
	"fmt"

	sn "github.com/NethermindEth/cairo-vm-go/pkg/parsers/starknet"
	"github.com/NethermindEth/cairo-vm-go/pkg/parsers/zero"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/builtins"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

type Program struct {
	// the bytecode in string format
	Bytecode []*fp.Element
	// given a string it returns the pc for that function call
	Entrypoints map[string]uint64
	// it stores the start and end label pcs
	Labels map[string]uint64
	// builtins
	Builtins []builtins.BuiltinType
}

type CairoProgram struct{}

func LoadCairoZeroProgram(cairoZeroJson *zero.ZeroProgram) (*Program, error) {
	// bytecode
	bytecode := make([]*fp.Element, len(cairoZeroJson.Data))
	for i := range cairoZeroJson.Data {
		felt, err := new(fp.Element).SetString(cairoZeroJson.Data[i])
		if err != nil {
			return nil, fmt.Errorf(
				"cannot read bytecode %s at position %d: %w",
				cairoZeroJson.Data[i], i, err,
			)
		}
		bytecode[i] = felt
	}

	entrypoints, labels := extractEntrypointsAndLabels(cairoZeroJson)

	return &Program{
		Bytecode:    bytecode,
		Entrypoints: entrypoints,
		Labels:      labels,
		Builtins:    cairoZeroJson.Builtins,
	}, nil
}

func extractEntrypointsAndLabels(json *zero.ZeroProgram) (map[string]uint64, map[string]uint64) {
	entrypoints := map[string]uint64{}
	for key, ident := range json.Identifiers {
		if ident.IdentifierType == "function" {
			name := key[len(json.MainScope)+1:]
			entrypoints[name] = uint64(ident.Pc)
		}
	}

	labels := make(map[string]uint64, 2)
	for key, ident := range json.Identifiers {
		if ident.IdentifierType == "label" {
			name := key[len(json.MainScope)+1:]
			labels[name] = uint64(ident.Pc)
		}
	}

	return entrypoints, labels
}

func LoadCairoProgram(cairoProgram *sn.StarknetProgram) (*Program, error) {
	bytecode := make([]*fp.Element, len(cairoProgram.Bytecode))
	for i, felt := range cairoProgram.Bytecode {
		f := felt
		bytecode[i] = &f
	}
	entrypoints := extractCairoEntryPoints(cairoProgram)
	builtins := extractCairoBuiltins(cairoProgram)

	return &Program{
		Bytecode:    bytecode,
		Entrypoints: entrypoints,
		Labels:      nil,
		Builtins:    builtins,
	}, nil
}

func extractCairoEntryPoints(cairoProgram *sn.StarknetProgram) map[string]uint64 {
	entrypoints := make(map[string]uint64)

	for name, entry := range cairoProgram.EntryPointsByFunction {
		entrypoints[name] = uint64(entry.Offset)
	}
	return entrypoints
}

func extractCairoBuiltins(cairoProgram *sn.StarknetProgram) []builtins.BuiltinType {
	uniqueBuiltins := make(map[builtins.BuiltinType]struct{})
	var builtins []builtins.BuiltinType
	for _, entry := range cairoProgram.EntryPointsByFunction {
		for _, builtin := range entry.Builtins {
			if _, exists := uniqueBuiltins[builtin]; !exists {
				uniqueBuiltins[builtin] = struct{}{}
				builtins = append(builtins, builtin)
			}
		}
	}
	return builtins
}
