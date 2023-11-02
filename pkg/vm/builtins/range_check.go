package builtins

import (
	"errors"
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
)

const RangeCheckName = "range_check"

type RangeCheck struct{}

func (r *RangeCheck) CheckWrite(segment *memory.Segment, offset uint64, value *memory.MemoryValue) error {
	felt, err := value.FieldElement()
	if err != nil {
		return fmt.Errorf("check write: %w", err)
	}

	// felt >= (2^128)
	if felt.Cmp(&utils.FeltMax128) != -1 {
		return fmt.Errorf("check write: 2**128 < %s", value)
	}
	return nil
}

func (r *RangeCheck) InferValue(segment *memory.Segment, offset uint64) error {
	return errors.New("cannot infer value")
}

func (r *RangeCheck) String() string {
	return RangeCheckName
}
