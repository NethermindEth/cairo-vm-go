package zero

import (
	"math/big"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

// GetUint256AsFelts helper function reads two consecutive memory cells,
// the first one containing the low part of the `uint256` variable and
// the second one containing the high part of the `uint256` variable
//
// The low and high parts previously extracted from memory are then
// converted to field elements and returned
func GetUint256AsFelts(vm *VM.VirtualMachine, ref hinter.ResOperander) (*fp.Element, *fp.Element, error) {
	lowRefAddr, err := ref.GetAddress(vm)
	if err != nil {
		return nil, nil, err
	}

	lowPart, err := vm.Memory.ReadFromAddress(&lowRefAddr)
	if err != nil {
		return nil, nil, err
	}

	highRefAddr, err := lowRefAddr.AddOffset(1)
	if err != nil {
		return nil, nil, err
	}

	highPart, err := vm.Memory.ReadFromAddress(&highRefAddr)
	if err != nil {
		return nil, nil, err
	}

	low, err := lowPart.FieldElement()
	if err != nil {
		return nil, nil, err
	}

	high, err := highPart.FieldElement()
	if err != nil {
		return nil, nil, err
	}

	return low, high, nil
}

func GetUint256ExpandAsFelts(vm *VM.VirtualMachine, ref hinter.ResOperander) ([]*fp.Element, error) {
	//> struct Uint256_expand {
	//> 	B0: felt,
	//> 	b01: felt,
	//> 	b12: felt,
	//> 	b23: felt,
	//> 	b3: felt,
	//> }
	refAddr, err := ref.GetAddress(vm)
	if err != nil {
		return nil, err
	}
	uint256Expanded := make([]*fp.Element, 6)
	for i := 0; i < 5; i++ {
		refValMV, err := vm.Memory.ReadFromAddress(&refAddr)
		if err != nil {
			return nil, err
		}
		uint256Expanded[i], err = refValMV.FieldElement()
		if err != nil {
			return nil, err
		}
		refAddr, err = refAddr.AddOffset(1)
		if err != nil {
			return nil, err
		}
	}
	return uint256Expanded, nil

func GetUint512AsFelts(vm *VM.VirtualMachine, ref hinter.ResOperander) (*fp.Element, *fp.Element, *fp.Element, *fp.Element, error) {
	lowRefAddr, err := ref.GetAddress(vm)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	lowPart, err := vm.Memory.ReadFromAddress(&lowRefAddr)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	highRefAddr, err := lowRefAddr.AddOffset(1)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	highPart, err := vm.Memory.ReadFromAddress(&highRefAddr)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	highLowRefAddr, err := highRefAddr.AddOffset(1)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	highLowPart, err := vm.Memory.ReadFromAddress(&highLowRefAddr)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	highHighRefAddr, err := highLowRefAddr.AddOffset(1)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	highHighPart, err := vm.Memory.ReadFromAddress(&highHighRefAddr)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	lowLow, err := lowPart.FieldElement()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	lowHigh, err := highPart.FieldElement()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	highLow, err := highLowPart.FieldElement()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	highHigh, err := highHighPart.FieldElement()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return lowLow, lowHigh, highLow, highHigh, nil
}

// This helper function is used in FastEcAddAssignNewY and
// EcDoubleAssignNewYV1 hints to compute the y-coordinate of
// a point on an elliptic curve
//
// ComputeYCoordinate returns `valueBig` which is the result of
// the computation: (slope * (x - new_x) - y) % SECP_P
func ComputeYCoordinate(slopeBig *big.Int, xBig *big.Int, new_xBig *big.Int, yBig *big.Int, secPBig *big.Int) *big.Int {
	new_yBig := new(big.Int)
	new_yBig.Sub(xBig, new_xBig)
	new_yBig.Mul(new_yBig, slopeBig)
	new_yBig.Sub(new_yBig, yBig)
	new_yBig.Mod(new_yBig, secPBig)

	valueBig := new(big.Int)
	valueBig.Set(new_yBig)

	return valueBig
}
