package zero

import (
	"fmt"
	"math/big"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/core"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func newIsLeFeltHint(a, b hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "IsLeFelt",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> memory[ap] = 0 if (ids.a % PRIME) <= (ids.b % PRIME) else 1
			apAddr := vm.Context.AddressAp()

			a, err := hinter.ResolveAsFelt(vm, a)
			if err != nil {
				return err
			}
			b, err := hinter.ResolveAsFelt(vm, b)
			if err != nil {
				return err
			}

			var v memory.MemoryValue
			if utils.FeltLe(a, b) {
				v = memory.MemoryValueFromFieldElement(&utils.FeltZero)
			} else {
				v = memory.MemoryValueFromFieldElement(&utils.FeltOne)
			}
			return vm.Memory.WriteToAddress(&apAddr, &v)
		},
	}
}

func createIsLeFeltHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	a, err := resolver.GetResOperander("a")
	if err != nil {
		return nil, err
	}
	b, err := resolver.GetResOperander("b")
	if err != nil {
		return nil, err
	}
	return newIsLeFeltHint(a, b), nil
}

func newAssertLtFeltHint(a, b hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "AssertLtFelt",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.math_utils import assert_integer
			//> assert_integer(ids.a)
			//> assert_integer(ids.b)
			//> assert (ids.a % PRIME) < (ids.b % PRIME),
			//>        f'a = {ids.a % PRIME} is not less than b = {ids.b % PRIME}.'
			a, err := hinter.ResolveAsFelt(vm, a)
			if err != nil {
				return err
			}
			b, err := hinter.ResolveAsFelt(vm, b)
			if err != nil {
				return err
			}

			if !utils.FeltLt(a, b) {
				return fmt.Errorf("a = %v is not less than b = %v", a, b)
			}
			return nil
		},
	}
}

func createAssertLtFeltHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	a, err := resolver.GetResOperander("a")
	if err != nil {
		return nil, err
	}
	b, err := resolver.GetResOperander("b")
	if err != nil {
		return nil, err
	}
	return newAssertLtFeltHint(a, b), nil
}

func newAssertNotZeroHint(value hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "AssertNotZero",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.math_utils import assert_integer
			//> assert_integer(ids.value)
			//> assert ids.value % PRIME != 0, f'assert_not_zero failed: {ids.value} = 0.'
			value, err := hinter.ResolveAsFelt(vm, value)
			if err != nil {
				return err
			}

			if value.IsZero() {
				return fmt.Errorf("assertion failed: value is zero")
			}
			return nil
		},
	}
}

func createAssertNotZeroHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	value, err := resolver.GetResOperander("value")
	if err != nil {
		return nil, err
	}
	return newAssertNotZeroHint(value), nil
}

func newAssertNNHint(a hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "AssertNN",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.math_utils import assert_integer
			//> assert_integer(ids.a)
			//> assert 0 <= ids.a % PRIME < range_check_builtin.bound, f'a = {ids.a} is out of range.'
			a, err := hinter.ResolveAsFelt(vm, a)
			if err != nil {
				return err
			}

			if !utils.FeltIsPositive(a) {
				return fmt.Errorf("assertion failed: a = %v is out of range", a)
			}
			return nil
		},
	}
}

func createAssertNNHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	a, err := resolver.GetResOperander("a")
	if err != nil {
		return nil, err
	}
	return newAssertNNHint(a), nil
}

func newAssertNotEqualHint(a, b hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "AssertNotEqual",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> from starkware.cairo.lang.vm.relocatable import RelocatableValue
			//> both_ints = isinstance(ids.a, int) and isinstance(ids.b, int)
			//> both_relocatable = (
			//>    isinstance(ids.a, RelocatableValue) and isinstance(ids.b, RelocatableValue) and
			//>    ids.a.segment_index == ids.b.segment_index)
			//> assert both_ints or both_relocatable,
			//>    f'assert_not_equal failed: non-comparable values: {ids.a}, {ids.b}.'
			//> assert (ids.a - ids.b) % PRIME != 0, f'assert_not_equal failed: {ids.a} = {ids.b}.'

			a, err := a.Resolve(vm)
			if err != nil {
				return err
			}
			b, err := b.Resolve(vm)
			if err != nil {
				return err
			}

			// Since IsFelt result can be treated as enum value for the type (there are only 2 types possible),
			// comparing it is enough to satisfy the "same type" constraint.
			if a.IsFelt() != b.IsFelt() {
				return fmt.Errorf("assertion failed: non-comparable values: %v, %v", a, b)
			}

			if a.Equal(&b) {
				return fmt.Errorf("assertion failed: %v = %v", a, b)
			}
			return nil
		},
	}
}

func createAssertNotEqualHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	a, err := resolver.GetResOperander("a")
	if err != nil {
		return nil, err
	}
	b, err := resolver.GetResOperander("b")
	if err != nil {
		return nil, err
	}
	return newAssertNotEqualHint(a, b), nil
}

func newAssert250bitsHint(low, high, value hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "Assert250bits",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> const UPPER_BOUND = 2 ** 250;
			//> const SHIFT = 2 ** 128;
			//
			//> from starkware.cairo.common.math_utils import as_int
			//> # Correctness check.
			//> value = as_int(ids.value, PRIME) % PRIME
			//> assert value < ids.UPPER_BOUND, f'{value} is outside of the range [0, 2**250).'
			//> # Calculation for the assertion.
			//> ids.high, ids.low = divmod(ids.value, ids.SHIFT)

			value, err := hinter.ResolveAsFelt(vm, value)
			if err != nil {
				return err
			}
			if !utils.FeltLt(value, &utils.FeltUpperBound) {
				return fmt.Errorf("assertion failed: %v is outside of the range [0, 2**250)", value)
			}

			lowAddr, err := low.GetAddress(vm)
			if err != nil {
				return err
			}
			highAddr, err := high.GetAddress(vm)
			if err != nil {
				return err
			}

			div, rem := utils.FeltDivRem(value, &utils.FeltMax128)

			// div goes to high, rem goes to low.
			divValue := memory.MemoryValueFromFieldElement(&div)
			if err := vm.Memory.WriteToAddress(&highAddr, &divValue); err != nil {
				return err
			}
			remValue := memory.MemoryValueFromFieldElement(&rem)
			if err := vm.Memory.WriteToAddress(&lowAddr, &remValue); err != nil {
				return err
			}

			return nil
		},
	}
}

func createAssert250bitsHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	// low and high are expected to be references to a range_check_ptr builtin.
	// Like that:
	//> let low = [range_check_ptr];
	//> let high = [range_check_ptr + 1];

	low, err := resolver.GetResOperander("low")
	if err != nil {
		return nil, err
	}
	high, err := resolver.GetResOperander("high")
	if err != nil {
		return nil, err
	}
	value, err := resolver.GetResOperander("value")
	if err != nil {
		return nil, err
	}
	return newAssert250bitsHint(low, high, value), nil
}

func newAssertLeFeltHint(a, b, rangeCheckPtr hinter.ResOperander) hinter.Hinter {
	return &core.AssertLeFindSmallArc{
		A:             a,
		B:             b,
		RangeCheckPtr: rangeCheckPtr,
	}
}

func createAssertLeFeltHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	a, err := resolver.GetResOperander("a")
	if err != nil {
		return nil, err
	}
	b, err := resolver.GetResOperander("b")
	if err != nil {
		return nil, err
	}
	rangeCheckPtr, err := resolver.GetResOperander("range_check_ptr")
	if err != nil {
		return nil, err
	}
	return newAssertLeFeltHint(a, b, rangeCheckPtr), nil
}

func createAssertLeFeltExcluded0Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	return &core.AssertLeIsFirstArcExcluded{SkipExcludeAFlag: hinter.ApCellRef(0)}, nil
}

func createAssertLeFeltExcluded1Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	return &core.AssertLeIsSecondArcExcluded{SkipExcludeBMinusA: hinter.ApCellRef(0)}, nil
}

func createAssertLeFeltExcluded2Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	// This hint is Cairo0-specific.
	// It only does a python-scoped variable named "excluded" assert.
	// We store that variable inside a hinter context.
	h := &GenericZeroHinter{
		Name: "AssertLeFeltExcluded2",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			if ctx.ExcludedArc != 2 {
				return fmt.Errorf("assertion `excluded == 2` failed")
			}
			return nil
		},
	}
	return h, nil
}

func newIsNNHint(a hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "IsNN",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			apAddr := vm.Context.AddressAp()
			//> memory[ap] = 0 if 0 <= (ids.a % PRIME) < range_check_builtin.bound else 1

			// a is already modulo PRIME, no need to adjust it.
			a, err := hinter.ResolveAsFelt(vm, a)
			if err != nil {
				return err
			}
			// range_check_builtin.bound is utils.FeltMax128 (1 << 128).
			var v memory.MemoryValue
			if utils.FeltIsPositive(a) {
				v = memory.MemoryValueFromFieldElement(&utils.FeltZero)
			} else {
				v = memory.MemoryValueFromFieldElement(&utils.FeltOne)
			}
			return vm.Memory.WriteToAddress(&apAddr, &v)
		},
	}
}

func createIsNNHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	a, err := resolver.GetResOperander("a")
	if err != nil {
		return nil, err
	}
	return newIsNNHint(a), nil
}

func newIsNNOutOfRangeHint(a hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "IsNNOutOfRange",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			apAddr := vm.Context.AddressAp()
			//> memory[ap] = 0 if 0 <= ((-ids.a - 1) % PRIME) < range_check_builtin.bound else 1
			a, err := hinter.ResolveAsFelt(vm, a)
			if err != nil {
				return err
			}
			var lhs fp.Element
			lhs.Sub(&utils.FeltZero, a) //> -ids.a
			lhs.Sub(&lhs, &utils.FeltOne)
			var v memory.MemoryValue
			if utils.FeltLt(&lhs, &utils.FeltMax128) {
				v = memory.MemoryValueFromFieldElement(&utils.FeltZero)
			} else {
				v = memory.MemoryValueFromFieldElement(&utils.FeltOne)
			}
			return vm.Memory.WriteToAddress(&apAddr, &v)
		},
	}
}

func createIsNNOutOfRangeHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	// This hint is executed for the negative values.
	// If the value was non-negative, it's usually handled by the IsNN hint.

	a, err := resolver.GetResOperander("a")
	if err != nil {
		return nil, err
	}
	return newIsNNOutOfRangeHint(a), nil
}

func newIsPositiveHint(value, dst hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "IsPositive",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.math_utils import is_positive
			//> ids.is_positive = 1 if is_positive(
			//>     value=ids.value, prime=PRIME, rc_bound=range_check_builtin.bound) else 0

			isPositiveAddr, err := dst.GetAddress(vm)
			if err != nil {
				return err
			}

			value, err := hinter.ResolveAsFelt(vm, value)
			if err != nil {
				return err
			}

			var v memory.MemoryValue
			if utils.FeltIsPositive(value) {
				v = memory.MemoryValueFromFieldElement(&utils.FeltOne)
			} else {
				v = memory.MemoryValueFromFieldElement(&utils.FeltZero)
			}
			return vm.Memory.WriteToAddress(&isPositiveAddr, &v)
		},
	}
}

func createIsPositiveHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	value, err := resolver.GetResOperander("value")
	if err != nil {
		return nil, err
	}
	output, err := resolver.GetResOperander("output")
	if err != nil {
		return nil, err
	}
	return newIsPositiveHint(value, output), nil
}

func newSplitIntAssertRangeHint(value hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "SplitIntAssertRange",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> assert ids.value == 0, 'split_int(): value is out of range.'

			value, err := hinter.ResolveAsFelt(vm, value)
			if err != nil {
				return err
			}
			if !value.IsZero() {
				return fmt.Errorf("assertion `split_int(): value is out of range` failed")
			}

			return nil
		},
	}
}

func createSplitIntAssertRangeHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	value, err := resolver.GetResOperander("value")
	if err != nil {
		return nil, err
	}
	return newSplitIntAssertRangeHint(value), nil
}

func newSplitIntHint(output, value, base, bound hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "SplitIntHint",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> memory[ids.output] = res = (int(ids.value) % PRIME) % ids.base
			//> assert res < ids.bound, f'split_int(): Limb {res} is out of range.'

			outputAddr, err := output.GetAddress(vm)
			if err != nil {
				return err
			}

			base, err := hinter.ResolveAsFelt(vm, base)
			if err != nil {
				return err
			}

			value, err := hinter.ResolveAsFelt(vm, value)
			if err != nil {
				return err
			}

			bound, err := hinter.ResolveAsFelt(vm, bound)
			if err != nil {
				return err
			}

			result := utils.FeltMod(value, base)
			if !utils.FeltLt(&result, bound) {
				return fmt.Errorf("assertion `split_int(): Limb %v is out of range` failed", &result)
			}

			v := memory.MemoryValueFromFieldElement(&result)
			return vm.Memory.WriteToAddress(&outputAddr, &v)
		},
	}
}

func createSplitIntHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	output, err := resolver.GetResOperander("output")
	if err != nil {
		return nil, err
	}
	value, err := resolver.GetResOperander("value")
	if err != nil {
		return nil, err
	}
	base, err := resolver.GetResOperander("base")
	if err != nil {
		return nil, err
	}
	bound, err := resolver.GetResOperander("bound")
	if err != nil {
		return nil, err
	}
	return newSplitIntHint(output, value, base, bound), nil
}

func newSplitFeltHint(maxHigh, maxLow, low, high, value hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "SplitFelt",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.math_utils import assert_integer\nassert ids.MAX_HIGH < 2**128 and ids.MAX_LOW < 2**128\nassert PRIME - 1 == ids.MAX_HIGH * 2**128 + ids.MAX_LOW\nassert_integer(ids.value)\nids.low = ids.value & ((1 << 128) - 1)\nids.high = ids.value >> 128

			// assert ids.MAX_HIGH < 2**128 and ids.MAX_LOW < 2**128
			maxHigh, err := hinter.ResolveAsFelt(vm, maxHigh)
			if err != nil {
				return err
			}
			if !utils.FeltLt(maxHigh, &utils.FeltMax128) {
				return fmt.Errorf("assertion `split_felt(): MAX_HIGH %v is out of range` failed", maxHigh)
			}
			maxLow, err := hinter.ResolveAsFelt(vm, maxLow)
			if err != nil {
				return err
			}
			if !utils.FeltLt(maxLow, &utils.FeltMax128) {
				return fmt.Errorf("assertion `split_felt(): MAX_LOW %v is out of range` failed", maxLow)
			}
			// assert PRIME - 1 == ids.MAX_HIGH * 2**128 + ids.MAX_LOW
			leftHandSide := new(fp.Element).SetInt64(-1)
			rightHandSide := new(fp.Element).Add(new(fp.Element).Mul(maxHigh, &utils.FeltMax128), maxLow)
			if leftHandSide.Cmp(rightHandSide) != 0 {
				return fmt.Errorf("assertion `split_felt(): The sum of MAX_HIGH and MAX_LOW does not equal to PRIME - 1` failed")
			}
			// assert_integer(ids.value)
			value, err := hinter.ResolveAsFelt(vm, value)
			if err != nil {
				return err
			}
			var valueBigInt big.Int
			value.BigInt(&valueBigInt)
			lowFelt, err := hinter.ResolveAsFelt(vm, low)
			if err != nil {
				return err
			}
			var lowBigInt big.Int
			lowFelt.BigInt(&lowBigInt)
			highFelt, err := hinter.ResolveAsFelt(vm, high)
			if err != nil {
				return err
			}
			var highBigInt big.Int
			highFelt.BigInt(&highBigInt)

			// ids.low = ids.value & ((1 << 128)
			var felt128 big.Int
			utils.FeltMax128.BigInt(&felt128)
			lowBigInt.And(&valueBigInt, &felt128)
			lowValue := memory.MemoryValueFromFieldElement(new(fp.Element).SetBigInt(&lowBigInt))

			lowAddr, err := low.GetAddress(vm)
			if err != nil {
				return err
			}

			err = vm.Memory.WriteToAddress(&lowAddr, &lowValue)
			if err != nil {
				return err
			}
			// ids.high = ids.value >> 128
			highBigInt.Rsh(&valueBigInt, 128)
			highValue := memory.MemoryValueFromFieldElement(new(fp.Element).SetBigInt(&highBigInt))

			highAddr, err := high.GetAddress(vm)
			if err != nil {
				return err
			}

			return vm.Memory.WriteToAddress(&highAddr, &highValue)

		},
	}
}

func createSplitFeltHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	maxHigh, err := resolver.GetResOperander("MAX_HIGH")
	if err != nil {
		return nil, err
	}

	maxLow, err := resolver.GetResOperander("MAX_LOW")
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

	value, err := resolver.GetResOperander("value")
	if err != nil {
		return nil, err
	}

	return newSplitFeltHint(maxHigh, maxLow, low, high, value), nil
}
