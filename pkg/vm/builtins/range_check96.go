package builtins

import (
	"errors"
	"fmt"
	"math"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

const RangeCheck96Name = "range_check96"
const inputCellsPerRangeCheck96 = 1
const cellsPerRangeCheck96 = 1
const instancesPerComponentRangeCheck96 = 1

// Each range_check96 instance consists of RANGE_CHECK_96_N_PARTS 16-bit parts. INNER_RC_96_BOUND_SHIFT and INNER_RC_96_BOUND_MASK are used to extract 16-bit parts from the field elements stored in the range check segment.
const INNER_RC_96_BOUND_SHIFT = 16
const INNER_RC_96_BOUND_MASK = (1 << 16) - 1
const RANGE_CHECK_96_N_PARTS = 6

type RangeCheck96 struct {
	ratio uint64
}

func (r *RangeCheck96) CheckWrite(segment *memory.Segment, offset uint64, value *memory.MemoryValue) error {
	felt, err := value.FieldElement()
	if err != nil {
		return fmt.Errorf("check write: %w", err)
	}

	// 2**96
	BOUND_96, err := new(fp.Element).SetString("79228162514264337593543950336")
	if err != nil {
		return fmt.Errorf("check write: %w", err)
	}

	// felt >= (2^96)
	if felt.Cmp(BOUND_96) != -1 {
		return fmt.Errorf("check write: 2**96 < %s", value)
	}
	return nil
}

func (r *RangeCheck96) InferValue(segment *memory.Segment, offset uint64) error {
	return errors.New("cannot infer value")
}

func (r *RangeCheck96) String() string {
	return RangeCheckName
}

func (r *RangeCheck96) GetAllocatedSize(segmentUsedSize uint64, vmCurrentStep uint64) (uint64, error) {
	return getBuiltinAllocatedSize(segmentUsedSize, vmCurrentStep, r.ratio, inputCellsPerRangeCheck96, instancesPerComponentRangeCheck96, cellsPerRangeCheck96)
}

// GetRangeCheck96Usage returns the min and max values used in the range_check96 segment. Since each range_check96 instance consists of 16-bit parts, the min and max values are calculated by iterating over the segment data and extracting the 16-bit parts from each field element.
func (r *RangeCheck96) GetRangeCheck96Usage(rangeCheck96Segment *memory.Segment) (uint16, uint16) {
	var minVal, maxVal uint16 = math.MaxUint16, 0
	for _, value := range rangeCheck96Segment.Data {
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
