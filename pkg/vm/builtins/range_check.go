package builtins

import (
	"errors"
	"fmt"
	"math"

	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
)

const RangeCheckName = "range_check"
const inputCellsPerRangeCheck = 1
const cellsPerRangeCheck = 1
const instancesPerComponentRangeCheck = 1

// Each range check instance consists of RANGE_CHECK_N_PARTS 16-bit parts. INNER_RC_BOUND_SHIFT and INNER_RC_BOUND_MASK are used to extract 16-bit parts from the field elements stored in the range check segment.
const INNER_RC_BOUND_SHIFT = 16
const INNER_RC_BOUND_MASK = (1 << 16) - 1
const RANGE_CHECK_N_PARTS = 8

type RangeCheck struct {
	ratio uint64
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
	return getBuiltinAllocatedSize(segmentUsedSize, vmCurrentStep, r.ratio, inputCellsPerRangeCheck, instancesPerComponentRangeCheck, cellsPerRangeCheck)
}

// GetRangeCheckUsage returns the min and max values used in the range check segment. Since each range check instance consists of 16-bit parts, the min and max values are calculated by iterating over the segment data and extracting the 16-bit parts from each field element.
func (r *RangeCheck) GetRangeCheckUsage(rangeCheckSegment *memory.Segment) (uint16, uint16) {
	var minVal, maxVal uint16 = math.MaxUint16, 0
	for _, value := range rangeCheckSegment.Data {
		valueFelt, err := value.FieldElement()
		if err != nil {
			continue
		}
		feltDigits := valueFelt.Bits()
		for _, digit := range feltDigits {
			for i := 0; i < 4; i++ {
				part := uint16((digit >> (i * INNER_RC_BOUND_SHIFT)) & INNER_RC_BOUND_MASK)
				if part < minVal {
					minVal = part
				}
				if part > maxVal {
					maxVal = part
				}
			}
		}
	}
	return minVal, maxVal
}
