package builtins

import (
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

type LayoutBuiltin struct {
	// Runner for the builtin
	Runner memory.BuiltinRunner
	// Builtin id from starknet parser
	Builtin Builtin
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
		{Runner: &Output{}, Builtin: OutputEnum},
		{Runner: &Pedersen{ratio: 8}, Builtin: PedersenEnum},
		{Runner: &RangeCheck{ratio: 8, RangeCheckNParts: 8}, Builtin: RangeCheckEnum},
		{Runner: &ECDSA{ratio: 512}, Builtin: ECDSAEnum},
	}}
}

func getDexLayout() Layout {
	return Layout{Name: "dex", RcUnits: 4, Builtins: []LayoutBuiltin{
		{Runner: &Output{}, Builtin: OutputEnum},
		{Runner: &Pedersen{ratio: 8}, Builtin: PedersenEnum},
		{Runner: &RangeCheck{ratio: 8, RangeCheckNParts: 8}, Builtin: RangeCheckEnum},
		{Runner: &ECDSA{ratio: 512}, Builtin: ECDSAEnum},
	}}
}

func getRecursiveLayout() Layout {
	return Layout{Name: "recursive", RcUnits: 4, Builtins: []LayoutBuiltin{
		{Runner: &Output{}, Builtin: OutputEnum},
		{Runner: &Pedersen{ratio: 128}, Builtin: PedersenEnum},
		{Runner: &RangeCheck{ratio: 8, RangeCheckNParts: 8}, Builtin: RangeCheckEnum},
		{Runner: &Bitwise{ratio: 8}, Builtin: BitwiseEnum},
	}}
}

func getStarknetLayout() Layout {
	return Layout{Name: "starknet", RcUnits: 4, Builtins: []LayoutBuiltin{
		{Runner: &Output{}, Builtin: OutputEnum},
		{Runner: &Pedersen{ratio: 32}, Builtin: PedersenEnum},
		{Runner: &RangeCheck{ratio: 16, RangeCheckNParts: 8}, Builtin: RangeCheckEnum},
		{Runner: &ECDSA{ratio: 2048}, Builtin: ECDSAEnum},
		{Runner: &Bitwise{ratio: 64}, Builtin: BitwiseEnum},
		{Runner: &EcOp{ratio: 1024, cache: make(map[uint64]fp.Element)}, Builtin: ECOPEnum},
		{Runner: &Poseidon{ratio: 32, cache: make(map[uint64]fp.Element)}, Builtin: PoseidonEnum},
	}}
}

func getStarknetWithKeccakLayout() Layout {
	return Layout{Name: "starknet_with_keccak", RcUnits: 4, Builtins: []LayoutBuiltin{
		{Runner: &Output{}, Builtin: OutputEnum},
		{Runner: &Pedersen{ratio: 32}, Builtin: PedersenEnum},
		{Runner: &RangeCheck{ratio: 16, RangeCheckNParts: 8}, Builtin: RangeCheckEnum},
		{Runner: &ECDSA{ratio: 2048}, Builtin: ECDSAEnum},
		{Runner: &Bitwise{ratio: 64}, Builtin: BitwiseEnum},
		{Runner: &EcOp{ratio: 1024, cache: make(map[uint64]fp.Element)}, Builtin: ECOPEnum},
		{Runner: &Keccak{ratio: 2048, cache: make(map[uint64]fp.Element)}, Builtin: KeccakEnum},
		{Runner: &Poseidon{ratio: 32, cache: make(map[uint64]fp.Element)}, Builtin: PoseidonEnum},
	}}
}

func getRecursiveLargeOutputLayout() Layout {
	return Layout{Name: "recursive_large_output", RcUnits: 4, Builtins: []LayoutBuiltin{
		{Runner: &Output{}, Builtin: OutputEnum},
		{Runner: &Pedersen{ratio: 128}, Builtin: PedersenEnum},
		{Runner: &RangeCheck{ratio: 8, RangeCheckNParts: 8}, Builtin: RangeCheckEnum},
		{Runner: &Bitwise{ratio: 8}, Builtin: BitwiseEnum},
		{Runner: &Poseidon{ratio: 8, cache: make(map[uint64]fp.Element)}, Builtin: PoseidonEnum},
	}}
}

func getRecursiveWithPoseidonLayout() Layout {
	return Layout{Name: "recursive_with_poseidon", RcUnits: 4, Builtins: []LayoutBuiltin{
		{Runner: &Output{}, Builtin: OutputEnum},
		{Runner: &Pedersen{ratio: 256}, Builtin: PedersenEnum},
		{Runner: &RangeCheck{ratio: 16, RangeCheckNParts: 8}, Builtin: RangeCheckEnum},
		{Runner: &Bitwise{ratio: 16}, Builtin: BitwiseEnum},
		{Runner: &Poseidon{ratio: 64, cache: make(map[uint64]fp.Element)}, Builtin: PoseidonEnum},
	}}
}

func getAllSolidityLayout() Layout {
	return Layout{Name: "recursive_with_poseidon", RcUnits: 8, Builtins: []LayoutBuiltin{
		{Runner: &Output{}, Builtin: OutputEnum},
		{Runner: &Pedersen{ratio: 8}, Builtin: PedersenEnum},
		{Runner: &RangeCheck{ratio: 8, RangeCheckNParts: 8}, Builtin: RangeCheckEnum},
		{Runner: &ECDSA{ratio: 512}, Builtin: ECDSAEnum},
		{Runner: &Bitwise{ratio: 256}, Builtin: BitwiseEnum},
		{Runner: &EcOp{ratio: 256, cache: make(map[uint64]fp.Element)}, Builtin: ECOPEnum},
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
