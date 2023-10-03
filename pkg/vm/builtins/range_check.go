package builtins

import (
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

type RangeCheck struct{}

// 1 << 128
var max, _ = new(fp.Element).SetString("0x100000000000000000000000000000000")

func (r *RangeCheck) CheckWrite(segment *memory.Segment, offset uint64, value *memory.MemoryValue) error {
	felt, err := value.ToFieldElement()
	if err != nil {
		return err
	}

	// felt >= (2^128)
	if felt.Cmp(max) != -1 {
		return fmt.Errorf("range check builtin failed for offset: %d value %s", offset, value)
	}
	return nil
}

func (r *RangeCheck) DeduceValue(segment *memory.Segment, offset uint64) error {
	segment.Data[offset] = memory.EmptyMemoryValueAsFelt()
	return nil
}
