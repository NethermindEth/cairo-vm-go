package zero

import(
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	// "github.com/holiman/uint256"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
)

func GetUint256AsFelts(vm *VM.VirtualMachine, ref hinter.ResOperander) (fp.Element, fp.Element, error) {
	values, err := hinter.GetConsecutiveValues(vm, ref, int16(2))
	if err != nil {
		return fp.Element{}, fp.Element{}, err
	}

	low, err := values[0].FieldElement()
	if err != nil {
		return fp.Element{}, fp.Element{}, err
	}

	high, err :=  values[1].FieldElement()
	if err != nil {
		return fp.Element{}, fp.Element{}, err
	}

	return *low, *high, nil
}