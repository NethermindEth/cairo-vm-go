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

func GetLayout(layout string) (Layout, error) {
	switch layout {
	case "small":
		return getSmallLayout(), nil
	case "plain":
		return getPlainLayout(), nil
	case "":
		return getPlainLayout(), nil
	default:
		return Layout{}, fmt.Errorf("Layout %s not found", layout)
	}
}
