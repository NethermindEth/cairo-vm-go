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
	Builtin BuiltinType
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
		{Runner: &Output{}, Builtin: OutputType},
		{Runner: &Pedersen{ratio: 8}, Builtin: PedersenType},
		{Runner: &RangeCheck{ratio: 8, RangeCheckNParts: 8}, Builtin: RangeCheckType},
		{Runner: &ECDSA{ratio: 512}, Builtin: ECDSAType},
	}}
}

func getDexLayout() Layout {
	return Layout{Name: "dex", RcUnits: 4, Builtins: []LayoutBuiltin{
		{Runner: &Output{}, Builtin: OutputType},
		{Runner: &Pedersen{ratio: 8}, Builtin: PedersenType},
		{Runner: &RangeCheck{ratio: 8, RangeCheckNParts: 8}, Builtin: RangeCheckType},
		{Runner: &ECDSA{ratio: 512}, Builtin: ECDSAType},
	}}
}

func getRecursiveLayout() Layout {
	return Layout{Name: "recursive", RcUnits: 4, Builtins: []LayoutBuiltin{
		{Runner: &Output{}, Builtin: OutputType},
		{Runner: &Pedersen{ratio: 128}, Builtin: PedersenType},
		{Runner: &RangeCheck{ratio: 8, RangeCheckNParts: 8}, Builtin: RangeCheckType},
		{Runner: &Bitwise{ratio: 8}, Builtin: BitwiseType},
	}}
}

func getStarknetLayout() Layout {
	return Layout{Name: "starknet", RcUnits: 4, Builtins: []LayoutBuiltin{
		{Runner: &Output{}, Builtin: OutputType},
		{Runner: &Pedersen{ratio: 32}, Builtin: PedersenType},
		{Runner: &RangeCheck{ratio: 16, RangeCheckNParts: 8}, Builtin: RangeCheckType},
		{Runner: &ECDSA{ratio: 2048}, Builtin: ECDSAType},
		{Runner: &Bitwise{ratio: 64}, Builtin: BitwiseType},
		{Runner: &EcOp{ratio: 1024, cache: make(map[uint64]fp.Element)}, Builtin: ECOPType},
		{Runner: &Poseidon{ratio: 32, cache: make(map[uint64]fp.Element)}, Builtin: PoseidonType},
	}}
}

func getStarknetWithKeccakLayout() Layout {
	return Layout{Name: "starknet_with_keccak", RcUnits: 4, Builtins: []LayoutBuiltin{
		{Runner: &Output{}, Builtin: OutputType},
		{Runner: &Pedersen{ratio: 32}, Builtin: PedersenType},
		{Runner: &RangeCheck{ratio: 16, RangeCheckNParts: 8}, Builtin: RangeCheckType},
		{Runner: &ECDSA{ratio: 2048}, Builtin: ECDSAType},
		{Runner: &Bitwise{ratio: 64}, Builtin: BitwiseType},
		{Runner: &EcOp{ratio: 1024, cache: make(map[uint64]fp.Element)}, Builtin: ECOPType},
		{Runner: &Keccak{ratio: 2048, cache: make(map[uint64]fp.Element)}, Builtin: KeccakType},
		{Runner: &Poseidon{ratio: 32, cache: make(map[uint64]fp.Element)}, Builtin: PoseidonType},
	}}
}

func getRecursiveLargeOutputLayout() Layout {
	return Layout{Name: "recursive_large_output", RcUnits: 4, Builtins: []LayoutBuiltin{
		{Runner: &Output{}, Builtin: OutputType},
		{Runner: &Pedersen{ratio: 128}, Builtin: PedersenType},
		{Runner: &RangeCheck{ratio: 8, RangeCheckNParts: 8}, Builtin: RangeCheckType},
		{Runner: &Bitwise{ratio: 8}, Builtin: BitwiseType},
		{Runner: &Poseidon{ratio: 8, cache: make(map[uint64]fp.Element)}, Builtin: PoseidonType},
	}}
}

func getRecursiveWithPoseidonLayout() Layout {
	return Layout{Name: "recursive_with_poseidon", RcUnits: 4, Builtins: []LayoutBuiltin{
		{Runner: &Output{}, Builtin: OutputType},
		{Runner: &Pedersen{ratio: 256}, Builtin: PedersenType},
		{Runner: &RangeCheck{ratio: 16, RangeCheckNParts: 8}, Builtin: RangeCheckType},
		{Runner: &Bitwise{ratio: 16}, Builtin: BitwiseType},
		{Runner: &Poseidon{ratio: 64, cache: make(map[uint64]fp.Element)}, Builtin: PoseidonType},
	}}
}

func getAllSolidityLayout() Layout {
	return Layout{Name: "recursive_with_poseidon", RcUnits: 8, Builtins: []LayoutBuiltin{
		{Runner: &Output{}, Builtin: OutputType},
		{Runner: &Pedersen{ratio: 8}, Builtin: PedersenType},
		{Runner: &RangeCheck{ratio: 8, RangeCheckNParts: 8}, Builtin: RangeCheckType},
		{Runner: &ECDSA{ratio: 512}, Builtin: ECDSAType},
		{Runner: &Bitwise{ratio: 256}, Builtin: BitwiseType},
		{Runner: &EcOp{ratio: 256, cache: make(map[uint64]fp.Element)}, Builtin: ECOPType},
	}}
}

func getAllCairoLayout() Layout {
	return Layout{Name: "all_cairo", RcUnits: 8, Builtins: []LayoutBuiltin{
		{Runner: &Output{}, Builtin: OutputType},
		{Runner: &Pedersen{ratio: 256}, Builtin: PedersenType},
		{Runner: &RangeCheck{ratio: 8, RangeCheckNParts: 8}, Builtin: RangeCheckType},
		{Runner: &ECDSA{ratio: 2048}, Builtin: ECDSAType},
		{Runner: &Bitwise{ratio: 16}, Builtin: BitwiseType},
		{Runner: &EcOp{ratio: 1024, cache: make(map[uint64]fp.Element)}, Builtin: ECOPType},
		{Runner: &Keccak{ratio: 2048, cache: make(map[uint64]fp.Element)}, Builtin: KeccakType},
		{Runner: &Poseidon{ratio: 256, cache: make(map[uint64]fp.Element)}, Builtin: PoseidonType},
		{Runner: &RangeCheck{ratio: 8, RangeCheckNParts: 6}, Builtin: RangeCheck96Type},
		{Runner: &ModBuiltin{ratio: 128, wordBitLen: 1, batchSize: 96, modBuiltinType: Add}, Builtin: AddModeType},
		{Runner: &ModBuiltin{ratio: 256, wordBitLen: 1, batchSize: 96, modBuiltinType: Mul}, Builtin: MulModType},
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
	case "all_cairo":
		return getAllCairoLayout(), nil
	case "":
		return getPlainLayout(), nil
	default:
		return Layout{}, fmt.Errorf("Layout %s not found", layout)
	}
}
