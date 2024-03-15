package zero

import (
	"fmt"
	"math/big"
	"slices"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/holiman/uint256"
)

func GetUint256AsFelts(vm *VM.VirtualMachine, ref hinter.ResOperander) (*fp.Element, *fp.Element, error) {
	values, err := hinter.GetConsecutiveValues(vm, ref, int16(2))
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

func newUint256AddHint(a, b, carryLow, carryHigh hinter.ResOperander, lowOnly bool) hinter.Hinter {
	name := "Uint256Add"
	if lowOnly {
		name += "Low"
	}
	return &GenericZeroHinter{
		Name: name,
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> sum_low = ids.a.low + ids.b.low
			//> ids.carry_low = 1 if sum_low >= ids.SHIFT else 0

			// Uint256AddLow does not implement this part
			//> sum_high = ids.a.high + ids.b.high + ids.carry_low
			//> ids.carry_high = 1 if sum_high >= ids.SHIFT else 0

			aLow, aHigh, err := GetUint256AsFelts(vm, a)
			if err != nil {
				return err
			}
			bLow, bHigh, err := GetUint256AsFelts(vm, b)
			if err != nil {
				return err
			}

			// Calculate `carry_low` memory value
			sumLow := new(fp.Element).Add(aLow, bLow)
			var cLow *fp.Element
			if utils.FeltLe(&utils.FeltMax128, sumLow) {
				cLow = &utils.FeltOne
			} else {
				cLow = &utils.FeltZero
			}
			cLowValue := memory.MemoryValueFromFieldElement(cLow)

			// Save `carry_low` value in address
			addrCarryLow, err := carryLow.GetAddress(vm)
			if err != nil {
				return err
			}
			err = vm.Memory.WriteToAddress(&addrCarryLow, &cLowValue)
			if err != nil {
				return err
			}

			if !lowOnly {
				// Calculate `carry_high` memory value
				sumHigh := new(fp.Element).Add(aHigh, bHigh)
				sumHigh.Add(sumHigh, cLow)
				var cHigh *fp.Element
				if utils.FeltLe(&utils.FeltMax128, sumHigh) {
					cHigh = &utils.FeltOne
				} else {
					cHigh = &utils.FeltZero
				}
				cHighValue := memory.MemoryValueFromFieldElement(cHigh)

				// Save `carry_high` value in address
				addrCarryHigh, err := carryHigh.GetAddress(vm)
				if err != nil {
					return err
				}
				err = vm.Memory.WriteToAddress(&addrCarryHigh, &cHighValue)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func createUint256AddHinter(resolver hintReferenceResolver, low bool) (hinter.Hinter, error) {
	a, err := resolver.GetResOperander("a")
	if err != nil {
		return nil, err
	}

	b, err := resolver.GetResOperander("b")
	if err != nil {
		return nil, err
	}

	carryLow, err := resolver.GetResOperander("carry_low")
	if err != nil {
		return nil, err
	}

	var carryHigh hinter.ResOperander
	if !low {
		carryHigh, err = resolver.GetResOperander("carry_high")
		if err != nil {
			return nil, err
		}
	}

	return newUint256AddHint(a, b, carryLow, carryHigh, low), nil
}

func newSplit64Hint(a, low, high hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "Split64",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> 	ids.low = ids.a & ((1<<64) - 1)
			//> 	ids.high = ids.a >> 64

			a, err := hinter.ResolveAsFelt(vm, a)
			if err != nil {
				return err
			}
			var aBig big.Int
			a.BigInt(&aBig)

			// Calculate low value
			mask := new(big.Int).SetUint64(^uint64(0))
			lowBig := new(big.Int).And(&aBig, mask)
			low64 := lowBig.Uint64()
			lowValue := memory.MemoryValueFromUint(low64)

			lowAddr, err := low.GetAddress(vm)
			if err != nil {
				return err
			}

			err = vm.Memory.WriteToAddress(&lowAddr, &lowValue)
			if err != nil {
				return err
			}

			// Calculate high value
			highBig := new(big.Int).Rsh(&aBig, 64)
			highValue := memory.MemoryValueFromFieldElement(new(fp.Element).SetBigInt(highBig))

			highAddr, err := high.GetAddress(vm)
			if err != nil {
				return err
			}

			return vm.Memory.WriteToAddress(&highAddr, &highValue)
		},
	}
}

func createSplit64Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	a, err := resolver.GetResOperander("a")
	if err != nil {
		return nil, err
	}

	low, err := resolver.GetResOperander("low")
	if err != nil {
		return nil, err
	}

	high, err := resolver.GetResOperander("high")
	if err != nil {
		return nil, err
	}

	return newSplit64Hint(a, low, high), nil
}

func newUint256SqrtHint(n hinter.ResOperander, root hinter.ResOperander) hinter.Hinter {
	// from starkware.python.math_utils import isqrt\nn = (ids.n.high << 128) + ids.n.low\nroot = isqrt(n)\nassert 0 <= root < 2 ** 128\nids.root.low = root\nids.root.high = 0
	return &GenericZeroHinter{
		Name: "Uint256Sqrt",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> from starkware.python.math_utils import isqrt
			//n = (ids.n.high << 128) + ids.n.low
			//root = isqrt(n)
			//assert 0 <= root < 2 ** 128
			//ids.root.low = root
			//ids.root.high = 0

			nHigh, nLow, err := GetUint256AsFelts(vm, n)
			if err != nil {
				return err
			}
			//> n = (ids.n.high << 128) + ids.n.low
			valueLowU256 := uint256.Int(nLow.Bits())
			value := uint256.Int(nHigh.Bits())
			value.Lsh(&value, 128)
			value.Add(&value, &valueLowU256)

			//> root = isqrt(n)
			calculatedUint256Root := new(uint256.Int).Sqrt(&value)
			calculatedUint256RootBytes := calculatedUint256Root.Bytes()
			slices.Reverse(calculatedUint256RootBytes)
			calculatedFeltRoot := new(fp.Element).SetBytes(calculatedUint256RootBytes)

			//> assert 0 <= root < 2 ** 128
			if !utils.FeltIsPositive(calculatedFeltRoot) {
				return fmt.Errorf("assertion failed: a = %v is out of range", calculatedUint256Root)
			}
			rootAddr, err := root.GetAddress(vm)
			if err != nil {
				return err
			}
			//> ids.root.low = root
			rootLowValue := memory.MemoryValueFromFieldElement(calculatedFeltRoot)
			err = vm.Memory.WriteToAddress(&rootAddr, &rootLowValue)
			if err != nil {
				return err
			}
			//> ids.root.high = 0
			rootHighValue := memory.MemoryValueFromFieldElement(new(fp.Element).SetInt64(0))
			return hinter.WriteToNthStructField(vm, root, rootHighValue, 1)
		},
	}
}

func createUint256SqrtHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	n, err := resolver.GetResOperander("n")
	if err != nil {
		return nil, err
	}
	root, err := resolver.GetResOperander("root")
	if err != nil {
		return nil, err
	}
	return newUint256SqrtHint(n, root), nil
}
