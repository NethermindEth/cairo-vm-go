package builtins

import (
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/parsers/starknet"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
)

type Builtin struct {
	Runner  memory.BuiltinRunner
	Builtin starknet.Builtin
}

type Layout struct {
	Name     string
	RcUnits  uint64
	Builtins []Builtin
}

func getSmallLayout() Layout {
	return Layout{Name: "small", RcUnits: 16, Builtins: []Builtin{
		{Runner: &Output{}, Builtin: starknet.Output},
		{Runner: &Pedersen{ratioPedersen: 8, instancesPerComponentPedersen: 1}, Builtin: starknet.Pedersen},
		{Runner: &RangeCheck{ratioRangeCheck: 8, instancesPerComponentRangeCheck: 1, RangeCheckNParts: 8, InnerRCBound: 2 << 16}, Builtin: starknet.RangeCheck},
		{Runner: &ECDSA{ratioECDSA: 512, instancesPerComponentECDSA: 1}, Builtin: starknet.ECDSA},
	}}
}

func getPlainLayout() Layout {
	return Layout{Name: "plain", RcUnits: 16, Builtins: []Builtin{}}
}

func getAllCairoLayout() Layout {
	return Layout{Name: "all_cairo", RcUnits: 16, Builtins: []Builtin{
		{Runner: &Output{}, Builtin: starknet.Output},
		{Runner: &Pedersen{ratioPedersen: 256, instancesPerComponentPedersen: 1}, Builtin: starknet.Pedersen},
		{Runner: &RangeCheck{ratioRangeCheck: 8, instancesPerComponentRangeCheck: 1, RangeCheckNParts: 8, InnerRCBound: 2 << 16}, Builtin: starknet.RangeCheck},
		{Runner: &ECDSA{ratioECDSA: 2048, instancesPerComponentECDSA: 1}, Builtin: starknet.ECDSA},
		{Runner: &Bitwise{ratioBitwise: 16, instancesPerComponentBitwise: 1}, Builtin: starknet.Bitwise},
		{Runner: &EcOp{ratioEcOp: 1024, instancesPerComponentEcOp: 1}, Builtin: starknet.ECOP},
		{Runner: &Keccak{ratioKeccak: 2048, instancesPerComponentKeccak: 16}, Builtin: starknet.Keccak},
	}}

}

func GetLayout(layout string) (Layout, error) {
	switch layout {
	case "small":
		return getSmallLayout(), nil
	case "plain":
		return getPlainLayout(), nil
	case "all_cairo":
		return getAllCairoLayout(), nil
	case "":
		return getPlainLayout(), nil
	default:
		return Layout{}, fmt.Errorf("Layout %s not found", layout)
	}
}
