package builtinrunner

import (
	starknetParser "github.com/NethermindEth/cairo-vm-go/pkg/parsers/starknet"
)

// Desing ideas each builtin menionted in compile file gets its own segment that is then
// mapped to the proper struct
type BuiltinRunner struct {
	builtins []Builtin
}

func (b *BuiltinRunner) AddBuiltin(name starknetParser.Builtin, segmentIndex uint64) {
	switch name {
	case starknetParser.Output:
		b.builtins = append(b.builtins, NewOutput(segmentIndex))
	case starknetParser.RangeCheck:
		b.builtins = append(b.builtins, NewRangeCheck(segmentIndex))
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

func (b BuiltinRunner) Builtins() []Builtin {
	return b.builtins
}

type Builtin interface {
	Run()
	Segment() uint64
	Name() string
}
