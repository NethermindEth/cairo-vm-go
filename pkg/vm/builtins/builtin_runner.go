package builtins

import (
	"fmt"
	"math"
	"strconv"

	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
)

type Builtin uint8

const (
	OutputEnum Builtin = iota + 1
	RangeCheckEnum
	PedersenEnum
	ECDSAEnum
	KeccakEnum
	BitwiseEnum
	ECOPEnum
	PoseidonEnum
	SegmentArenaEnum
	RangeCheck96Enum
)

func Runner(name Builtin) memory.BuiltinRunner {
	switch name {
	case OutputEnum:
		return &Output{}
	case RangeCheckEnum:
		return &RangeCheck{0, 8}
	case RangeCheck96Enum:
		return &RangeCheck{0, 6}
	case PedersenEnum:
		return &Pedersen{}
	case ECDSAEnum:
		return &ECDSA{}
	case KeccakEnum:
		return &Keccak{}
	case BitwiseEnum:
		return &Bitwise{}
	case ECOPEnum:
		return &EcOp{}
	case PoseidonEnum:
		return &Poseidon{}
	case SegmentArenaEnum:
		panic("Not implemented")
	default:
		panic("Unknown builtin")
	}
}

// GetBuiltinAllocatedInstances calculates the number of instances of given builtin for current step and builtin ratio.
// Ratio parameter defines the ratio between the number of steps to the number of builtin instances. It means that this builtin is expected to be used once every 'ratio' steps of the execution.
// cellsPerInstance defines the number of cells that one instance of the builtin occupies.
// segmentUsedSize defines the real number of cells used in the segment.
// instancesPerComponent defines the number of instances per component (segment or a part of the Cairo program that makes use of builtins).
func GetBuiltinAllocatedInstances(ratio uint64, cellsPerInstance uint64, segmentUsedSize uint64, instancesPerComponent uint64, vmCurrentStep uint64) (uint64, error) {
	if ratio == 0 {
		instances := math.Ceil(float64(segmentUsedSize) / float64(cellsPerInstance))
		neededComponents := math.Ceil(instances / float64(instancesPerComponent))
		components := uint64(0)
		if neededComponents > 0 {
			components = utils.NextPowerOfTwo(uint64(neededComponents))
		}
		return components * instancesPerComponent, nil
	}
	minSteps := ratio * instancesPerComponent
	if vmCurrentStep < minSteps {
		return 0, fmt.Errorf("number of steps must be at least %d. Current step: %d", minSteps, vmCurrentStep)
	}
	return vmCurrentStep / ratio, nil
}

func getBuiltinAllocatedSize(segmentUsedSize uint64, vmCurrentStep uint64, ratio uint64, inputCellsPerInstance uint64, instancesPerComponent uint64, cellsPerInstance uint64) (uint64, error) {
	allocatedInstances, err := GetBuiltinAllocatedInstances(ratio, inputCellsPerInstance, segmentUsedSize, instancesPerComponent, vmCurrentStep)
	if err != nil {
		return 0, err
	}
	return allocatedInstances * cellsPerInstance, nil
}


func (b Builtin) MarshalJSON() ([]byte, error) {
	switch b {
	case OutputEnum:
		return []byte(OutputName), nil
	case RangeCheckEnum:
		return []byte(RangeCheckName), nil
	case RangeCheck96Enum:
		return []byte(RangeCheck96Name), nil
	case PedersenEnum:
		return []byte(PedersenName), nil
	case ECDSAEnum:
		return []byte(ECDSAName), nil
	case KeccakEnum:
		return []byte(KeccakName), nil
	case BitwiseEnum:
		return []byte(BitwiseName), nil
	case ECOPEnum:
		return []byte(EcOpName), nil
	case PoseidonEnum:
		return []byte(PoseidonName), nil
	case SegmentArenaEnum:
		return []byte(SegmentArenaName), nil

	}
	return nil, fmt.Errorf("marshal unknown builtin: %d", uint8(b))
}

func (b *Builtin) UnmarshalJSON(data []byte) error {
	builtinName, err := strconv.Unquote(string(data))
	if err != nil {
		return fmt.Errorf("unmarshal builtin: %w", err)
	}

	switch builtinName {
	case OutputName:
		*b = OutputEnum
	case RangeCheckName:
		*b = RangeCheckEnum
	case RangeCheck96Name:
		*b = RangeCheck96Enum
	case PedersenName:
		*b = PedersenEnum
	case ECDSAName:
		*b = ECDSAEnum
	case KeccakName:
		*b = KeccakEnum
	case BitwiseName:
		*b = BitwiseEnum
	case EcOpName:
		*b = ECOPEnum
	case PoseidonName:
		*b = PoseidonEnum
	case SegmentArenaName:
		*b = SegmentArenaEnum
	default:
		return fmt.Errorf("unmarshal unknown builtin: %s", builtinName)
	}
	return nil
}

