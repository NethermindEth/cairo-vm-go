package zero

import (
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

// GetUint256AsFelts helper function reads two consecutive memory cells,
// the first one containing the low part of the `uint256` and the second one
// containing the high part of the `uint256`
//
// The low and high parts previously extracted from memory are then
// converted to field elements and returned
func GetUint256AsFelts(vm *VM.VirtualMachine, ref hinter.ResOperander) (*fp.Element, *fp.Element, error) {
	lowRefAddr, err := ref.GetAddress(vm)
	if err != nil {
		return nil, nil, err
	}

	values := make([]mem.MemoryValue, 2)

	lowPart, err := vm.Memory.ReadFromAddress(&lowRefAddr)
	if err != nil {
		return nil, nil, err
	}

	values[0] = lowPart

	highRefAddr, err := lowRefAddr.AddOffset(1)
	if err != nil {
		return nil, nil, err
	}

	highPart, err := vm.Memory.ReadFromAddress(&highRefAddr)
	if err != nil {
		return nil, nil, err
	}

	values[1] = highPart

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
