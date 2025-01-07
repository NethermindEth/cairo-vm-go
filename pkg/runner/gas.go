package runner

import (
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
)

type TokenGasCost uint8

const (
	ConstToken TokenGasCost = iota + 1
	HoleToken
	RangeCheckToken
	RangeCheck96Token
	PedersenToken
	PoseidonToken
	BitwiseToken
	EcOpToken
	AddModToken
	MulModToken
)

func getTokenGasCost(token TokenGasCost) (uint64, error) {
	switch token {
	case ConstToken:
		return 1, nil
	case PedersenToken:
		return 4050, nil
	case PoseidonToken:
		return 491, nil
	case BitwiseToken:
		return 583, nil
	case EcOpToken:
		return 4085, nil
	case AddModToken:
		return 230, nil
	case MulModToken:
		return 604, nil
	default:
		return 0, fmt.Errorf("token has no cost")
	}
}

func gasInitialization(memory *memory.Memory) error {
	builtinsCostSegmentAddress := memory.AllocateEmptySegment()
	mv := mem.MemoryValueFromMemoryAddress(&builtinsCostSegmentAddress)
	programSegment := memory.Segments[vm.ProgramSegment]
	err := memory.Write(0, programSegment.Len(), &mv)
	if err != nil {
		return err
	}
	preCostTokenTypes := []TokenGasCost{PedersenToken, PoseidonToken, BitwiseToken, EcOpToken, AddModToken, MulModToken}
	for _, token := range preCostTokenTypes {
		cost, err := getTokenGasCost(token)
		if err != nil {
			return err
		}
		mv := mem.MemoryValueFromUint(cost)
		err = memory.WriteToAddress(&builtinsCostSegmentAddress, &mv)
		if err != nil {
			return err
		}
		builtinsCostSegmentAddress, err = builtinsCostSegmentAddress.AddOffset(1)
		if err != nil {
			return err
		}
	}
	return nil
}
