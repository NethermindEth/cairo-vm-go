package zero

import (
	"errors"
	"fmt"

	sn "github.com/NethermindEth/cairo-vm-go/pkg/parsers/starknet"
	"github.com/NethermindEth/cairo-vm-go/pkg/parsers/zero"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

type Program struct {
	// the bytecode in string format
	Bytecode []*f.Element
	// given a string it returns the pc for that function call
	Entrypoints map[string]uint64
	// it stores the start and end label pcs
	Labels map[string]uint64
	// builtins
	Builtins []sn.Builtin
}

func LoadCairoZeroProgram(cairoZeroJson *zero.ZeroProgram) (*Program, error) {
	// bytecode
	bytecode := make([]*f.Element, len(cairoZeroJson.Data))
	for i := range cairoZeroJson.Data {
		felt, err := new(f.Element).SetString(cairoZeroJson.Data[i])
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
		func(key string, typex string, value map[string]any) error {
			if typex == "function" {
				pc, ok := value["pc"].(float64)
				if !ok {
					return fmt.Errorf("%s: unknown entrypoint pc", key)
				}
				name := key[len(json.MainScope)+1:]
				result[name] = uint64(pc)
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
		func(key string, typex string, value map[string]any) error {
			if typex == "label" {
				pc, ok := value["pc"].(float64)
				if !ok {
					return fmt.Errorf("%s: unknown entrypoint pc", key)
				}
				name := key[len(json.MainScope)+1:]
				labels[name] = uint64(pc)
			}
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("extracting labels: %w", err)
	}

	return labels, nil
}

func scanIdentifiers(
	json *zero.ZeroProgram,
	f func(key string, typex string, value map[string]any) error,
) error {
	for key, value := range json.Identifiers {
		properties := value.(map[string]any)

		typex, ok := properties["type"].(string)
		if !ok {
			return errors.New("unnespecified identifier type")
		}
		if err := f(key, typex, properties); err != nil {
			return err
		}
	}
	return nil
}
