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
		panic("Not implemented")
	case starknetParser.Keccak:
		panic("Not implemented")
	case starknetParser.Bitwise:
		return &Bitwise{}
	case starknetParser.ECOP:
		panic("Not implemented")
	case starknetParser.Poseidon:
		panic("Not implemented")
	case starknetParser.SegmentArena:
		panic("Not implemented")
	default:
		panic("Unknown builtin")
	}
}
