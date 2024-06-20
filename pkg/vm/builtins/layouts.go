package builtins

import (
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/parsers/starknet"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
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

func getSmallLayout() Layout {
	return Layout{Name: "small", RcUnits: 16, Builtins: []LayoutBuiltin{
		{Runner: &Output{}, Builtin: starknet.Output},
		{Runner: &Pedersen{ratio: 8}, Builtin: starknet.Pedersen},
		{Runner: &RangeCheck{ratio: 8, RangeCheckNParts: 8, InnerRCBound: 2 << 16}, Builtin: starknet.RangeCheck},
		{Runner: &ECDSA{ratio: 512}, Builtin: starknet.ECDSA},
	}}
}

func getPlainLayout() Layout {
	return Layout{Name: "plain", RcUnits: 16, Builtins: []LayoutBuiltin{}}
}

func getStarknetWithKeccakLayout() Layout {
	return Layout{Name: "starknet_with_keccak", RcUnits: 16, Builtins: []LayoutBuiltin{
		{Runner: &Output{}, Builtin: starknet.Output},
		{Runner: &Pedersen{ratio: 32}, Builtin: starknet.Pedersen},
		{Runner: &RangeCheck{ratio: 16, RangeCheckNParts: 8, InnerRCBound: 2 << 16}, Builtin: starknet.RangeCheck},
		{Runner: &ECDSA{ratio: 2048}, Builtin: starknet.ECDSA},
		{Runner: &Bitwise{ratio: 64}, Builtin: starknet.Bitwise},
		{Runner: &EcOp{ratio: 1024}, Builtin: starknet.ECOP},
		{Runner: &Keccak{ratio: 2048}, Builtin: starknet.Keccak},
		{Runner: &Poseidon{ratio: 32}, Builtin: starknet.Poseidon},
	}}
}

func GetLayout(layout string) (Layout, error) {
	switch layout {
	case "small":
		return getSmallLayout(), nil
	case "plain":
		return getPlainLayout(), nil
	case "starknet_with_keccak":
		return getStarknetWithKeccakLayout(), nil
	case "":
		return getPlainLayout(), nil
	default:
		return Layout{}, fmt.Errorf("Layout %s not found", layout)
	}
}
