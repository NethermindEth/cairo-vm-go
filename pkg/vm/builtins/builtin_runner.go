package builtins

import (
	"fmt"
	"math"
	"strconv"

	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
)

type BuiltinType uint8

const (
	OutputType BuiltinType = iota + 1
	RangeCheckType
	PedersenType
	ECDSAType
	KeccakType
	BitwiseType
	ECOPType
	PoseidonType
	SegmentArenaType
	RangeCheck96Type
	AddModeType
	MulModType
	GasBuiltinType
)

func Runner(name BuiltinType) memory.BuiltinRunner {
	switch name {
	case OutputType:
		return &Output{}
	case RangeCheckType:
		return &RangeCheck{0, 8}
	case RangeCheck96Type:
		return &RangeCheck{0, 6}
	case PedersenType:
		return &Pedersen{}
	case ECDSAType:
		return &ECDSA{}
	case KeccakType:
		return &Keccak{}
	case BitwiseType:
		return &Bitwise{}
	case ECOPType:
		return &EcOp{}
	case PoseidonType:
		return &Poseidon{}
	case AddModeType:
		return &ModBuiltin{modBuiltinType: Add}
	case MulModType:
		return &ModBuiltin{modBuiltinType: Mul}
	case SegmentArenaType:
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

func (b BuiltinType) MarshalJSON() ([]byte, error) {
	switch b {
	case OutputType:
		return []byte(OutputName), nil
	case RangeCheckType:
		return []byte(RangeCheckName), nil
	case RangeCheck96Type:
		return []byte(RangeCheck96Name), nil
	case PedersenType:
		return []byte(PedersenName), nil
	case ECDSAType:
		return []byte(ECDSAName), nil
	case KeccakType:
		return []byte(KeccakName), nil
	case BitwiseType:
		return []byte(BitwiseName), nil
	case ECOPType:
		return []byte(EcOpName), nil
	case PoseidonType:
		return []byte(PoseidonName), nil
	case AddModeType:
		return []byte("Add" + ModuloName), nil
	case MulModType:
		return []byte("Mul" + ModuloName), nil
	case SegmentArenaType:
		return []byte(SegmentArenaName), nil
	case GasBuiltinType:
		return []byte(GasBuiltinName), nil
	}
	return nil, fmt.Errorf("marshal unknown builtin: %d", uint8(b))
}

func (b *BuiltinType) UnmarshalJSON(data []byte) error {
	builtinName, err := strconv.Unquote(string(data))
	if err != nil {
		return fmt.Errorf("unmarshal builtin: %w", err)
	}

	switch builtinName {
	case OutputName:
		*b = OutputType
	case RangeCheckName:
		*b = RangeCheckType
	case RangeCheck96Name:
		*b = RangeCheck96Type
	case PedersenName:
		*b = PedersenType
	case ECDSAName:
		*b = ECDSAType
	case KeccakName:
		*b = KeccakType
	case BitwiseName:
		*b = BitwiseType
	case EcOpName:
		*b = ECOPType
	case PoseidonName:
		*b = PoseidonType
	case "Add" + ModuloName:
		*b = AddModeType
	case "Mul" + ModuloName:
		*b = MulModType
	case SegmentArenaName:
		*b = SegmentArenaType
	case GasBuiltinName:
		*b = GasBuiltinType
	default:
		return fmt.Errorf("unmarshal unknown builtin: %s", builtinName)
	}
	return nil
}
