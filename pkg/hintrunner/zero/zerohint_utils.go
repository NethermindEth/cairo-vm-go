package zero

import (
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func GetUint256AsFelts(vm *VM.VirtualMachine, ref hinter.ResOperander) (*fp.Element, *fp.Element, error) {
	refAddr, err := ref.GetAddress(vm)
	if err != nil {
		return nil, nil, err
	}

	values, err := hinter.GetConsecutiveValues(vm, refAddr, int16(2))
	if err != nil {
		return nil, nil, err
	}

	low, err := values[0].FieldElement()
	if err != nil {
		return nil, nil, err
	}

	high, err := values[1].FieldElement()
	if err != nil {
		return nil, nil, err
	}

	return low, high, nil
}
