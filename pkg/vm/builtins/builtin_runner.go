package builtins

import (
	"fmt"
	"math"

	starknetParser "github.com/NethermindEth/cairo-vm-go/pkg/parsers/starknet"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
)

func Runner(name starknetParser.Builtin) memory.BuiltinRunner {
	switch name {
	case starknetParser.Output:
		return &Output{}
	case starknetParser.RangeCheck:
		return &RangeCheck{}
	case starknetParser.Pedersen:
		return &Pedersen{}
	case starknetParser.ECDSA:
		return &ECDSA{}
	case starknetParser.Keccak:
		return &Keccak{}
	case starknetParser.Bitwise:
		return &Bitwise{}
	case starknetParser.ECOP:
		return &EcOp{}
	case starknetParser.Poseidon:
		return &Poseidon{}
	case starknetParser.SegmentArena:
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
		return 0, fmt.Errorf("Number of steps must be at least %d. Current step: %d", minSteps, vmCurrentStep)
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
