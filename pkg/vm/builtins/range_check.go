package builtins

import (
	"errors"
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
)

const RangeCheckName = "range_check"
const cellsPerRangeCheck = 1
const INNER_RC_BOUND_SHIFT = 16
const INNER_RC_BOUND_MASK = (1 << 16) - 1

type RangeCheck struct {
	ratioRangeCheck                 uint64
	instancesPerComponentRangeCheck uint64
	RangeCheckNParts                uint64
	InnerRCBound                    uint64
}

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
	allocatedInstances, err := GetAllocatedInstances(r.ratioRangeCheck, cellsPerRangeCheck, segmentUsedSize, r.instancesPerComponentRangeCheck, vmCurrentStep)
	if err != nil {
		return 0, err
	}
	return allocatedInstances * cellsPerRangeCheck, nil
}

func (r *RangeCheck) GetRangeCheckUsage(rangeCheckSegment *memory.Segment) (uint64, uint64) {
	minVal, maxVal := ^uint64(0), uint64(0)
	for _, value := range rangeCheckSegment.Data {
		valueFelt, err := value.FieldElement()
		if err != nil {
			continue
		}
		feltDigits := valueFelt.Bits()
		for _, digit := range feltDigits {
			for i := 3; i >= 0; i-- {
				part := (digit >> (i * INNER_RC_BOUND_SHIFT)) & INNER_RC_BOUND_MASK
				if part < minVal {
					minVal = part
				}
				if part > maxVal {
					maxVal = part
				}
			}
		}
	}
	return uint64(minVal), uint64(maxVal)
}
