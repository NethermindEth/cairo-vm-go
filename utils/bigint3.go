package utils

import (
    "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"

)

func ResolveAsBigInt3(vm *VM.VirtualMachine, valAddr mem.MemoryAddress) ([]*fp.Element, error) {
	valMemoryValues, err := hinter.GetConsecutiveValues(vm, valAddr, int16(3))
	if err != nil {
		return nil, err
	}

	var valValues [3]*fp.Element
	for i := 0; i < 3; i++ {
		valValue, err := valMemoryValues[i].FieldElement()
		if err != nil {
			return nil, err
		}
		valValues[i] = valValue
	}

	return valValues[:], nil
}
