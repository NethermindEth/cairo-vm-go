package builtins

import (
	"errors"
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
)

const RangeCheckName = "range_check"
const cellsPerRangeCheck = 1

// TODO: Move to JSON
const ratioRangeCheck = 8
const instancesPerComponentRangeCheck = 1

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

func (r *RangeCheck) GetAllocatedSize(segmentUsedSize uint64, vmCurrentStep uint64) (uint64, error) {
	allocatedInstances, err := GetAllocatedInstances(ratioRangeCheck, cellsPerRangeCheck, segmentUsedSize, instancesPerComponentRangeCheck, vmCurrentStep)
	if err != nil {
		return 0, err
	}
	return allocatedInstances * cellsPerRangeCheck, nil
}
