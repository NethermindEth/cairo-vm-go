package builtins

import (
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

type RangeCheck struct{}

// 1 << 128
var max128 = fp.Element{18446744073700081665, 17407, 18446744073709551584, 576460752142434320}

func (r *RangeCheck) CheckWrite(segment *memory.Segment, offset uint64, value *memory.MemoryValue) error {
	felt, err := value.FieldElement()
	if err != nil {
		return err
	}

	// felt >= (2^128)
	if felt.Cmp(&max128) != -1 {
		return fmt.Errorf("range check builtin failed for offset: %d value %s", offset, value)
	}
	return nil
}

func (r *RangeCheck) InferValue(segment *memory.Segment, offset uint64) error {
	segment.Data[offset] = memory.EmptyMemoryValueAsFelt()
	return nil
}
