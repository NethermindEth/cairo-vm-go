package runner

import (
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/parsers/zero"
	builtins "github.com/NethermindEth/cairo-vm-go/pkg/vm/builtins"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

type ZeroProgram struct {
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

func LoadCairoZeroProgram(cairoZeroJson *zero.ZeroProgram) (*ZeroProgram, error) {
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

	return &ZeroProgram{
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
