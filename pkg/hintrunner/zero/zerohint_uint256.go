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

// Uint256Add hint computes the sum of the `low` and `high` parts of
// two `uint256` variables and checks for overflow
//
// `newUint256AddHint` takes 4 operanders as arguments
//   - `a` and `b` are the two `uint256` variables that will be added
//   - `carryLow` and `carryHigh` represent the potential extra bit that needs to be carried
//     if the sum of the `low` or `high` parts exceeds 2**128 - 1
func newUint256AddHint(a, b, carryLow, carryHigh hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "Uint256Add",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> sum_low = ids.a.low + ids.b.low
			//> ids.carry_low = 1 if sum_low >= ids.SHIFT else 0
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

			return nil
		},
	}
}

func createUint256AddHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
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

	carryHigh, err := resolver.GetResOperander("carry_high")
	if err != nil {
		return nil, err
	}

	return newUint256AddHint(a, b, carryLow, carryHigh), nil
}

// Split64 hint splits a field element in the range [0, 2^192) to its low 64-bit and high 128-bit parts
//
// `newSplit64Hint` takes 3 operanders as arguments
//   - `a` is the `felt` variable in range [0, 2^192) that will be splitted
//   - `low` and `high` represent the `low` 64 bits and the `high` 128 bits of the `felt` variable
func newSplit64Hint(a, low, high hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "Split64",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> ids.low = ids.a & ((1<<64) - 1)
			//> ids.high = ids.a >> 64

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

// Uint256Sqrt hint computes the square root of a given value, ensuring
// it falls within a specific range, i.e., `0 <= result < 2 ** 128`
//
// `newUint256SqrtHint` takes 2 operanders as arguments
//   - `n` represents the `uint256` variable for which we will calculate the square root
//   - `root` is the variable that will store the result of the hint in memory
func newUint256SqrtHint(n, root hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "Uint256Sqrt",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> from starkware.python.math_utils import isqrt
			//> n = (ids.n.high << 128) + ids.n.low
			//> root = isqrt(n)
			//> assert 0 <= root < 2 ** 128
			//> ids.root.low = root
			//> ids.root.high = 0

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
			return vm.Memory.WriteUint256ToAddress(rootAddr, &calculatedFeltRoot, &utils.FeltZero)
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

// Uint256SignedNN hint checks if a `uint256` variable is non-negative
// when considered as a signed number
//
// `newUint256SignedNNHint` takes 1 operander as argument
//   - `a` represents the `uint256` variable that will be checked
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

// Uint256UnsignedDivRem hint computes the division and modulus operations
// on `uint256` variables, combining the `high` and `low` parts of the dividend and divisor
//
// `newUint256UnsignedDivRemHint` takes 4 operanders as arguments
//   - `a` is the `uint256` variable that will be divided
//   - `div` is the `uint256` variable that will divide `a`
//   - `quotient` is the quotient of the Euclidean division of `a` by `div`
//   - `remainder` is the remainder of the Euclidean division of `a` by `div`
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

			err = vm.Memory.WriteUint256ToAddress(quotientAddr, lowQuot, highQuot)
			if err != nil {
				return err
			}

			remainderAddr, err := remainder.GetAddress(vm)
			if err != nil {
				return err
			}

			return vm.Memory.WriteUint256ToAddress(remainderAddr, lowRem, highRem)
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

// Uint256MulDivMod hint multiplies two `uint256` variables, divides the result
// by another `uint256` variable, and computes the quotient and the remainder
//
// `newUint256MulDivModHint` takes 6 operanders as arguments
//   - `a` and `b` are the `uint256` variables that will be multiplied
//   - `div` is the `uint256` variable that will divide the result of `a * b`
//   - `quotient` is the quotient of the Euclidean division of `a * b` by `div`
//   - `remainder` is the remainder of the Euclidean division of `a * b` by `div`
func newUint256MulDivModHint(a, b, div, quotientLow, quotientHigh, remainder hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "Uint256MulDivMod",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> a = (ids.a.high << 128) + ids.a.low
			//> b = (ids.b.high << 128) + ids.b.low
			//> div = (ids.div.high << 128) + ids.div.low
			//> quotient, remainder = divmod(a * b, div)
			//
			//> ids.quotient_low.low = quotient & ((1 << 128) - 1)
			//> ids.quotient_low.high = (quotient >> 128) & ((1 << 128) - 1)
			//> ids.quotient_high.low = (quotient >> 256) & ((1 << 128) - 1)
			//> ids.quotient_high.high = quotient >> 384
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

			err = vm.Memory.WriteUint256ToAddress(quotientLowAddr, lowQuotLow, lowQuotHigh)
			if err != nil {
				return err
			}

			quotientHighAddr, err := quotientHigh.GetAddress(vm)
			if err != nil {
				return err
			}

			err = vm.Memory.WriteUint256ToAddress(quotientHighAddr, highQuotLow, highQuotHigh)
			if err != nil {
				return err
			}

			remainderAddr, err := remainder.GetAddress(vm)
			if err != nil {
				return err
			}

			return vm.Memory.WriteUint256ToAddress(remainderAddr, lowRem, highRem)
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
