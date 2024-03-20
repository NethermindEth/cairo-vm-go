package zero

import (
	"fmt"
	"math/big"

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
	return &GenericZeroHinter{
		Name: "Uint256Sqrt",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> from starkware.python.math_utils import isqrt
			//n = (ids.n.high << 128) + ids.n.low
			//root = isqrt(n)
			//assert 0 <= root < 2 ** 128
			//ids.root.low = root
			//ids.root.high = 0

			nLow, nHigh, err := GetUint256AsFelts(vm, n)
			if err != nil {
				return err
			}
			//> n = (ids.n.high << 128) + ids.n.low
			valueLowU256 := uint256.Int(nLow.Bits())
			value := uint256.Int(nHigh.Bits())
			value.Lsh(&value, 128)
			value.Add(&value, &valueLowU256)

			//> root = isqrt(n)
			calculatedUint256Root := uint256.Int{}
			calculatedUint256Root.Sqrt(&value)
			calculatedUint256RootBytes := calculatedUint256Root.Bytes()
			calculatedFeltRoot := fp.Element{}
			calculatedFeltRoot.SetBytes(calculatedUint256RootBytes)
			//> assert 0 <= root < 2 ** 128
			if !utils.FeltIsPositive(&calculatedFeltRoot) {
				return fmt.Errorf("assertion failed: a = %v is out of range", calculatedUint256Root)
			}
			rootAddr, err := root.GetAddress(vm)
			if err != nil {
				return err
			}
			//> ids.root.low = root
			//> ids.root.high = 0
			return hinter.WriteUint256ToAddress(vm, rootAddr, &calculatedFeltRoot, &utils.FeltZero)
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

func newUint256SignedNNHint(a hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "Uint256SignedNN",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> memory[ap] = 1 if 0 <= (ids.a.high % PRIME) < 2 ** 127 else 0
			apAddr := vm.Context.AddressAp()

			_, aHigh, err := GetUint256AsFelts(vm, a)
			if err != nil {
				return err
			}
			var v memory.MemoryValue

			if utils.FeltLt(aHigh, &utils.Felt127) {
				v = memory.MemoryValueFromFieldElement(&utils.FeltOne)
			} else {
				v = memory.MemoryValueFromFieldElement(&utils.FeltZero)
			}
			return vm.Memory.WriteToAddress(&apAddr, &v)
		},
	}

}

func createUint256SignedNNHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	a, err := resolver.GetResOperander("a")
	if err != nil {
		return nil, err
	}
	return newUint256SignedNNHint(a), nil
}

func newUint256UnsignedDivRemHint(a, div, quotient, remainder hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "Uint256UnsignedDivRem",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> a = (ids.a.high << 128) + ids.a.low
			//> div = (ids.div.high << 128) + ids.div.low
			//> quotient, remainder = divmod(a, div)
			//> ids.quotient.low = quotient & ((1 << 128) - 1)
			//> ids.quotient.high = quotient >> 128
			//> ids.remainder.low = remainder & ((1 << 128) - 1)
			//> ids.remainder.high = remainder >> 128
			aLow, aHigh, err := GetUint256AsFelts(vm, a)
			if err != nil {
				return err
			}
			var aLowBig big.Int
			aLow.BigInt(&aLowBig)
			var aHighBig big.Int
			aHigh.BigInt(&aHighBig)

			divLow, divHigh, err := GetUint256AsFelts(vm, div)
			if err != nil {
				return err
			}
			var divLowBig big.Int
			divLow.BigInt(&divLowBig)
			var divHighBig big.Int
			divHigh.BigInt(&divHighBig)

			aBig := new(big.Int).Add(new(big.Int).Lsh(&aHighBig, 128), &aLowBig)
			divBig := new(big.Int).Add(new(big.Int).Lsh(&divHighBig, 128), &divLowBig)
			quotBig := new(big.Int).Div(aBig, divBig)
			remBig := new(big.Int).Mod(aBig, divBig)

			mask := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))

			lowQuot := new(fp.Element).SetBigInt(new(big.Int).And(quotBig, mask))
			highQuot := new(fp.Element).SetBigInt(new(big.Int).Rsh(quotBig, 128))

			lowRem := new(fp.Element).SetBigInt(new(big.Int).And(remBig, mask))
			highRem := new(fp.Element).SetBigInt(new(big.Int).Rsh(remBig, 128))

			quotientAddr, err := quotient.GetAddress(vm)
			if err != nil {
				return err
			}
			err = hinter.WriteUint256ToAddress(vm, quotientAddr, lowQuot, highQuot)
			if err != nil {
				return err
			}
			remainderAddr, err := remainder.GetAddress(vm)
			if err != nil {
				return err
			}
			return hinter.WriteUint256ToAddress(vm, remainderAddr, lowRem, highRem)
		},
	}

}

func createUint256UnsignedDivRemHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	a, err := resolver.GetResOperander("a")
	if err != nil {
		return nil, err
	}
	div, err := resolver.GetResOperander("div")
	if err != nil {
		return nil, err
	}
	quotient, err := resolver.GetResOperander("quotient")
	if err != nil {
		return nil, err
	}
	remainder, err := resolver.GetResOperander("remainder")
	if err != nil {
		return nil, err
	}
	return newUint256UnsignedDivRemHint(a, div, quotient, remainder), nil
}

func newUint256MulDivModHint(a, b, div, quotientLow, quotientHigh, remainder hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "Uint256MulDivMod",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {

			//> a = (ids.a.high << 128) + ids.a.low
			// b = (ids.b.high << 128) + ids.b.low
			// div = (ids.div.high << 128) + ids.div.low
			// quotient, remainder = divmod(a * b, div)

			// ids.quotient_low.low = quotient & ((1 << 128) - 1)
			// ids.quotient_low.high = (quotient >> 128) & ((1 << 128) - 1)
			// ids.quotient_high.low = (quotient >> 256) & ((1 << 128) - 1)
			// ids.quotient_high.high = quotient >> 384
			// ids.remainder.low = remainder & ((1 << 128) - 1)
			// ids.remainder.high = remainder >> 128
			aLow, aHigh, err := GetUint256AsFelts(vm, a)
			if err != nil {
				return err
			}
			var aLowBig big.Int
			aLow.BigInt(&aLowBig)
			var aHighBig big.Int
			aHigh.BigInt(&aHighBig)
			bLow, bHigh, err := GetUint256AsFelts(vm, b)
			if err != nil {
				return err
			}
			var bLowBig big.Int
			bLow.BigInt(&bLowBig)
			var bHighBig big.Int
			bHigh.BigInt(&bHighBig)
			divLow, divHigh, err := GetUint256AsFelts(vm, div)
			if err != nil {
				return err
			}
			var divLowBig big.Int
			divLow.BigInt(&divLowBig)
			var divHighBig big.Int
			divHigh.BigInt(&divHighBig)
			a := new(big.Int).Add(new(big.Int).Lsh(&aHighBig, 128), &aLowBig)
			b := new(big.Int).Add(new(big.Int).Lsh(&bHighBig, 128), &bLowBig)
			div := new(big.Int).Add(new(big.Int).Lsh(&divHighBig, 128), &divLowBig)
			quot := new(big.Int).Div(new(big.Int).Mul(a, b), div)
			rem := new(big.Int).Mod(new(big.Int).Mul(a, b), div)
			mask := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))
			lowQuotLow := new(fp.Element).SetBigInt(new(big.Int).And(quot, mask))
			lowQuotHigh := new(fp.Element).SetBigInt(new(big.Int).And(new(big.Int).Rsh(quot, 128), mask))
			highQuotLow := new(fp.Element).SetBigInt(new(big.Int).And(new(big.Int).Rsh(quot, 256), mask))
			highQuotHigh := new(fp.Element).SetBigInt(new(big.Int).Rsh(quot, 384))
			lowRem := new(fp.Element).SetBigInt(new(big.Int).And(rem, mask))
			highRem := new(fp.Element).SetBigInt(new(big.Int).Rsh(rem, 128))
			quotientLowAddr, err := quotientLow.GetAddress(vm)
			if err != nil {
				return err
			}
			err = hinter.WriteUint256ToAddress(vm, quotientLowAddr, lowQuotLow, lowQuotHigh)
			if err != nil {
				return err
			}
			quotientHighAddr, err := quotientHigh.GetAddress(vm)
			if err != nil {
				return err
			}
			err = hinter.WriteUint256ToAddress(vm, quotientHighAddr, highQuotLow, highQuotHigh)
			if err != nil {
				return err
			}
			remainderAddr, err := remainder.GetAddress(vm)
			if err != nil {
				return err
			}
			return hinter.WriteUint256ToAddress(vm, remainderAddr, lowRem, highRem)
		},
	}

}

func createUint256MulDivModHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	a, err := resolver.GetResOperander("a")
	if err != nil {
		return nil, err
	}
	b, err := resolver.GetResOperander("b")
	if err != nil {
		return nil, err
	}
	div, err := resolver.GetResOperander("div")
	if err != nil {
		return nil, err
	}
	quotientLow, err := resolver.GetResOperander("quotient_low")
	if err != nil {
		return nil, err
	}
	quotientHigh, err := resolver.GetResOperander("quotient_high")
	if err != nil {
		return nil, err
	}
	remainder, err := resolver.GetResOperander("remainder")
	if err != nil {
		return nil, err
	}
	return newUint256MulDivModHint(a, b, div, quotientLow, quotientHigh, remainder), nil
}
