package zero

import (
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/parsers/starknet"
	sn "github.com/NethermindEth/cairo-vm-go/pkg/parsers/starknet"
	"github.com/NethermindEth/cairo-vm-go/pkg/parsers/zero"
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
	Builtins []sn.Builtin
}

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

	entrypoints, err := extractEntrypoints(cairoZeroJson)
	if err != nil {
		return nil, err
	}

	labels, err := extractLabels(cairoZeroJson)
	if err != nil {
		return nil, err
	}

	return &Program{
		Bytecode:    bytecode,
		Entrypoints: entrypoints,
		Labels:      labels,
		Builtins:    cairoZeroJson.Builtins,
	}, nil
}

func extractEntrypoints(json *zero.ZeroProgram) (map[string]uint64, error) {
	result := make(map[string]uint64)
	err := scanIdentifiers(
		json,
		func(key string, ident *zero.Identifier) error {
			if ident.IdentifierType == "function" {
				name := key[len(json.MainScope)+1:]
				result[name] = uint64(ident.Pc)
			}
			return nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("extracting entrypoints: %w", err)
	}
	return result, nil
}

func extractLabels(json *zero.ZeroProgram) (map[string]uint64, error) {
	labels := make(map[string]uint64, 2)
	err := scanIdentifiers(
		json,
		func(key string, ident *zero.Identifier) error {
			if ident.IdentifierType == "label" {
				name := key[len(json.MainScope)+1:]
				labels[name] = uint64(ident.Pc)
			}
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("extracting labels: %w", err)
	}

	return labels, nil
}

func scanIdentifiers(json *zero.ZeroProgram, f func(key string, ident *zero.Identifier) error) error {
	for key, ident := range json.Identifiers {
		if err := f(key, ident); err != nil {
			return err
		}
	}
	return nil
}

func LoadCairoProgram(cairoProgram *sn.StarknetProgram) (*Program, error) {
	bytecode := make([]*fp.Element, len(cairoProgram.Bytecode))
	for i, felt := range cairoProgram.Bytecode {
		bytecode[i] = &felt
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

func extractCairoBuiltins(cairoProgram *sn.StarknetProgram) []sn.Builtin {
	uniqueBuiltins := make(map[starknet.Builtin]struct{})
	var builtins []starknet.Builtin
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
