package builtins

import (
	starknetParser "github.com/NethermindEth/cairo-vm-go/pkg/parsers/starknet"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
)

func AddBuiltin(name starknetParser.Builtin, segment *memory.Segment) *memory.Segment {
	switch name {
	case starknetParser.Output:
		panic("Not implemented")
	case starknetParser.RangeCheck:
		return segment.WithBuiltinRunner(&RangeCheck{})
	case starknetParser.Pedersen:
		panic("Not implemented")
	case starknetParser.ECDSA:
		panic("Not implemented")
	case starknetParser.Keccak:
		panic("Not implemented")
	case starknetParser.Bitwise:
		panic("Not implemented")
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
