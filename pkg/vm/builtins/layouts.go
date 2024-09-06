package builtins

import (
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/parsers/starknet"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

type LayoutBuiltin struct {
	// Runner for the builtin
	Runner memory.BuiltinRunner
	// Builtin id from starknet parser
	Builtin starknet.Builtin
}

type Layout struct {
	// Name of the layout
	Name string
	// Number of range check units allowed per step
	RcUnits uint64
	// List of builtins to be included in given layout
	Builtins []LayoutBuiltin
}

func getPlainLayout() Layout {
	return Layout{Name: "plain", RcUnits: 16, Builtins: []LayoutBuiltin{}}
}

func getSmallLayout() Layout {
	return Layout{Name: "small", RcUnits: 16, Builtins: []LayoutBuiltin{
		{Runner: &Output{}, Builtin: starknet.Output},
		{Runner: &Pedersen{ratio: 8}, Builtin: starknet.Pedersen},
		{Runner: &RangeCheck{ratio: 8, RangeCheckNParts: 8}, Builtin: starknet.RangeCheck},
		{Runner: &ECDSA{ratio: 512}, Builtin: starknet.ECDSA},
	}}
}

func getDexLayout() Layout {
	return Layout{Name: "dex", RcUnits: 4, Builtins: []LayoutBuiltin{
		{Runner: &Output{}, Builtin: starknet.Output},
		{Runner: &Pedersen{ratio: 8}, Builtin: starknet.Pedersen},
		{Runner: &RangeCheck{ratio: 8, RangeCheckNParts: 8}, Builtin: starknet.RangeCheck},
		{Runner: &ECDSA{ratio: 512}, Builtin: starknet.ECDSA},
	}}
}

func getRecursiveLayout() Layout {
	return Layout{Name: "recursive", RcUnits: 4, Builtins: []LayoutBuiltin{
		{Runner: &Output{}, Builtin: starknet.Output},
		{Runner: &Pedersen{ratio: 128}, Builtin: starknet.Pedersen},
		{Runner: &RangeCheck{ratio: 8, RangeCheckNParts: 8}, Builtin: starknet.RangeCheck},
		{Runner: &Bitwise{ratio: 8}, Builtin: starknet.Bitwise},
	}}
}

func getStarknetLayout() Layout {
	return Layout{Name: "starknet", RcUnits: 4, Builtins: []LayoutBuiltin{
		{Runner: &Output{}, Builtin: starknet.Output},
		{Runner: &Pedersen{ratio: 32}, Builtin: starknet.Pedersen},
		{Runner: &RangeCheck{ratio: 16, RangeCheckNParts: 8}, Builtin: starknet.RangeCheck},
		{Runner: &ECDSA{ratio: 2048}, Builtin: starknet.ECDSA},
		{Runner: &Bitwise{ratio: 64}, Builtin: starknet.Bitwise},
		{Runner: &EcOp{ratio: 1024, cache: make(map[uint64]fp.Element)}, Builtin: starknet.ECOP},
		{Runner: &Poseidon{ratio: 32, cache: make(map[uint64]fp.Element)}, Builtin: starknet.Poseidon},
	}}
}

func getStarknetWithKeccakLayout() Layout {
	return Layout{Name: "starknet_with_keccak", RcUnits: 4, Builtins: []LayoutBuiltin{
		{Runner: &Output{}, Builtin: starknet.Output},
		{Runner: &Pedersen{ratio: 32}, Builtin: starknet.Pedersen},
		{Runner: &RangeCheck{ratio: 16, RangeCheckNParts: 8}, Builtin: starknet.RangeCheck},
		{Runner: &ECDSA{ratio: 2048}, Builtin: starknet.ECDSA},
		{Runner: &Bitwise{ratio: 64}, Builtin: starknet.Bitwise},
		{Runner: &EcOp{ratio: 1024, cache: make(map[uint64]fp.Element)}, Builtin: starknet.ECOP},
		{Runner: &Keccak{ratio: 2048, cache: make(map[uint64]fp.Element)}, Builtin: starknet.Keccak},
		{Runner: &Poseidon{ratio: 32, cache: make(map[uint64]fp.Element)}, Builtin: starknet.Poseidon},
	}}
}

func getRecursiveLargeOutputLayout() Layout {
	return Layout{Name: "recursive_large_output", RcUnits: 4, Builtins: []LayoutBuiltin{
		{Runner: &Output{}, Builtin: starknet.Output},
		{Runner: &Pedersen{ratio: 128}, Builtin: starknet.Pedersen},
		{Runner: &RangeCheck{ratio: 8, RangeCheckNParts: 8}, Builtin: starknet.RangeCheck},
		{Runner: &Bitwise{ratio: 8}, Builtin: starknet.Bitwise},
		{Runner: &Poseidon{ratio: 8, cache: make(map[uint64]fp.Element)}, Builtin: starknet.Poseidon},
	}}
}

func getRecursiveWithPoseidonLayout() Layout {
	return Layout{Name: "recursive_with_poseidon", RcUnits: 4, Builtins: []LayoutBuiltin{
		{Runner: &Output{}, Builtin: starknet.Output},
		{Runner: &Pedersen{ratio: 256}, Builtin: starknet.Pedersen},
		{Runner: &RangeCheck{ratio: 16, RangeCheckNParts: 8}, Builtin: starknet.RangeCheck},
		{Runner: &Bitwise{ratio: 16}, Builtin: starknet.Bitwise},
		{Runner: &Poseidon{ratio: 64, cache: make(map[uint64]fp.Element)}, Builtin: starknet.Poseidon},
	}}
}

func getAllSolidityLayout() Layout {
	return Layout{Name: "recursive_with_poseidon", RcUnits: 8, Builtins: []LayoutBuiltin{
		{Runner: &Output{}, Builtin: starknet.Output},
		{Runner: &Pedersen{ratio: 8}, Builtin: starknet.Pedersen},
		{Runner: &RangeCheck{ratio: 8, RangeCheckNParts: 8}, Builtin: starknet.RangeCheck},
		{Runner: &ECDSA{ratio: 512}, Builtin: starknet.ECDSA},
		{Runner: &Bitwise{ratio: 256}, Builtin: starknet.Bitwise},
		{Runner: &EcOp{ratio: 256, cache: make(map[uint64]fp.Element)}, Builtin: starknet.ECOP},
	}}
}

func GetLayout(layout string) (Layout, error) {
	switch layout {
	case "plain":
		return getPlainLayout(), nil
	case "small":
		return getSmallLayout(), nil
	case "dex":
		return getDexLayout(), nil
	case "recursive":
		return getRecursiveLayout(), nil
	case "starknet":
		return getStarknetLayout(), nil
	case "starknet_with_keccak":
		return getStarknetWithKeccakLayout(), nil
	case "recursive_large_output":
		return getRecursiveLargeOutputLayout(), nil
	case "recursive_with_poseidon":
		return getRecursiveWithPoseidonLayout(), nil
	case "all_solidity":
		return getAllSolidityLayout(), nil
	case "":
		return getPlainLayout(), nil
	default:
		return Layout{}, fmt.Errorf("Layout %s not found", layout)
	}
}
