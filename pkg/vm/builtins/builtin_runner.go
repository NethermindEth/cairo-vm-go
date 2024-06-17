package builtins

import (
	starknetParser "github.com/NethermindEth/cairo-vm-go/pkg/parsers/starknet"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
)

func Runner(name starknetParser.Builtin) memory.BuiltinRunner {
	switch name {
	case starknetParser.Output:
		return &Output{}
	case starknetParser.RangeCheck:
		return &RangeCheck{}
	case starknetParser.Pedersen:
		return &Pedersen{}
	case starknetParser.ECDSA:
		return &ECDSA{}
	case starknetParser.Keccak:
		return &Keccak{}
	case starknetParser.Bitwise:
		return &Bitwise{}
	case starknetParser.ECOP:
		return &EcOp{}
	case starknetParser.Poseidon:
		return &Poseidon{}
	case starknetParser.SegmentArena:
		panic("Not implemented")
	default:
		panic("Unknown builtin")
	}
}
