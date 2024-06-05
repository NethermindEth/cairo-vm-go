package zero

import (
	"fmt"
	"math/big"

	"github.com/holiman/uint256"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/core"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	math_utils "github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/utils"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

// IsLeFelt hint determines if one value is less than or
// equal to another within a finite field
//
// `newIsLeFeltHint` takes 2 operanders as arguments
//   - `a` and `b` are the values that will be compared
//
// `newIsLeFeltHint` writes to `[ap]` the result of the comparison
// i.e, 1 if `a % PRIME <= b % PRIME`, 0 otherwise
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

// AssertLtFelt hint asserts that one value is strictly
// less than another within a finite field
//
// `newAssertLtFeltHint` takes 2 operanders as arguments
//   - `a` and `b` are the values that will be compared
//
// `newAssertLtFeltHint` returns an error if `a` is not
// strictly less than `b` within a finite field
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

// AssertNotZero hint asserts that a value is not zero
//
// `newAssertNotZeroHint` takes 1 operander as argument
//   - `value` is the value that will be compared to 0
//
// `newAssertNotZeroHint` returns an error if `value` is zero
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

// AssertNN hint asserts that a given value is non-negative
// within the range `0  <= value % PRIME < range-check builtin bound`
//
// `newAssertNNHint` takes 1 operander as argument
//   - `a` is the value that will be evaluated
//
// `newAssertNNHint` returns an error if `value` is negative
// within a specified range
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

// AssertNotEqual hint asserts that two given values are not equal
//
// `newAssertNotEqualHint` takes 2 operanders as arguments
//   - `a` and `b` are the values that will be compared
//
// `newAssertNotEqualHint` returns an error if `a` and `b` are not equal
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

// Assert250bits hint asserts that a value is within the range of 250 bits
//
// `newAssert250bitsHint` takes 3 operanders as arguments
//   - `value` is the value that will be evaluated
//   - `low` and `high` are the variables that will store the quotient and
//     remainder of the modular division of `value` by 2**128
//
// `newAssert250bitsHint` writes the quotient and the remainder of the modular
// division of `value` by 2**128 at `low` and `high` addresses in memory, respectively
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

// AsserLeFelt hint assert that one value is less than or equal to another
// within a finite field
//
// `newAssertLeFeltHint` takes 3 operanders as arguments
//   - `a` and `b` is the values that will be evaluated
//   - `rangeCheckPtr` is a pointer to the range-check builtin
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

// AssertLeFeltExcluded0 hint is a custom assertion related to `AssertLeFelt` hint
func createAssertLeFeltExcluded0Hinter() (hinter.Hinter, error) {
	return &core.AssertLeIsFirstArcExcluded{SkipExcludeAFlag: hinter.ApCellRef(0)}, nil
}

// AssertLeFeltExcluded1 hint is a custom assertion related to `AssertLeFelt` hint
func createAssertLeFeltExcluded1Hinter() (hinter.Hinter, error) {
	return &core.AssertLeIsSecondArcExcluded{SkipExcludeBMinusA: hinter.ApCellRef(0)}, nil
}

// AssertLeFeltExcluded2 hint is a custom assertion related to `AssertLeFelt` hint
func createAssertLeFeltExcluded2Hinter() (hinter.Hinter, error) {
	// This hint is Cairo0-specific
	// It only does a python-scoped variable named "excluded" assert
	// We store that variable inside a hinter context

	h := &GenericZeroHinter{
		Name: "AssertLeFeltExcluded2",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			excluded, err := ctx.ScopeManager.GetVariableValue("excluded")
			if err != nil {
				return err
			}

			if excluded != 2 {
				return fmt.Errorf("assertion `excluded == 2` failed")
			}

			return nil
		},
	}

	return h, nil
}

// IsNN hint checks if a value is non-negative within the range
// `0  <= value % PRIME < range-check builtin bound`
//
// `newIsNNHint` takes 1 operander as argument
//   - `a` is the value that will be evaluated
//
// `newIsNNHint` writes 0 or 1 to memory at `[ap]` address, depending on
// whether `a` is positive or negative, respectively
func newIsNNHint(a hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "IsNN",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> memory[ap] = 0 if 0 <= (ids.a % PRIME) < range_check_builtin.bound else 1

			apAddr := vm.Context.AddressAp()

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

// IsNNOutOfRange hint checks if the negation of a value minus one
// is non-negative within a specific range
//
// `newIsNNOutOfRangeHint` takes 1 operander as argument
//   - `a` is the value that will be evaluated
//
// `newIsNNOutOfRangeHint` writes 0 or 1 to memory at `[ap]` address, depending on
// whether the negation of `a` minus one is positive or negative, respectively
func newIsNNOutOfRangeHint(a hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "IsNNOutOfRange",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> memory[ap] = 0 if 0 <= ((-ids.a - 1) % PRIME) < range_check_builtin.bound else 1

			apAddr := vm.Context.AddressAp()

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

// IsPositive hint checks if a value is positive within
// a specific range and prime context
//
// `newIsPositiveHint` takes 2 operanders as arguments
//   - `value` is the value that will be evaluated
//   - `dst` is the address where to write the result in memmory
//
// `newIsPositiveHint` writes 1 or 0 to `dest` address, depending on
// whether `value` is positive or negative in the context, respectively
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

// SplitIntAssertRange hint asserts that the value to split in `SplitInt`
// function is zero once there are no more limbs to create
//
// `newSplitIntAssertRangeHint` takes 1 operander as argument
//   - `value` is the value that will be evaluated
//
// `newSplitIntAssertRangeHint` returns an error if `value` is not zero
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

// SplitIntHint hint splits an integer in limbs based on a base
// and asserts that each limb is within a certain range
//
// `newSplitIntHint` takes 4 operanders as arguments
//   - `output` is the location in memory where to write the result
//   - `value` is the value to split
//   - `base` is the base used to split the value in limbs
//   - `bound` represents the maximum value for each limb
//
// `newSplitIntHint` is recursively called when splitting an integer
// and writes to the `output` memory address the calculated limb
func newSplitIntHint(output, value, base, bound hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "SplitInt",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> memory[ids.output] = res = (int(ids.value) % PRIME) % ids.base
			//> assert res < ids.bound, f'split_int(): Limb {res} is out of range.'

			outputPtr, err := hinter.ResolveAsAddress(vm, output)
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

			return vm.Memory.WriteToAddress(outputPtr, &v)
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

// Pow hint calculates the least significant bit of the exponent
// of a number within a prime field
//
// `newPowHint` takes 2 operanders as arguments
//   - `locs` is the variable that will store the result
//   - `prevLocs` is the the variable used to calculate the exponent
//
// `newPowHint` writes to the memory address of `locs` variable the value of the least
// significant bit of the exponent of `prevLocs` variable module a prime field
func newPowHint(locs, prevLocs hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "Pow",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> ids.locs.bit = (ids.prev_locs.exp % PRIME) & 1
			/*> struct LoopLocals {
				bit: felt,
				temp0: felt,

				res: felt,
				base: felt,
				exp: felt,
			} */

			const expStructOffset = 4
			locsBitAddress, err := locs.GetAddress(vm)
			if err != nil {
				return err
			}

			prevLocsBitAddress, err := prevLocs.GetAddress(vm)
			if err != nil {
				return err
			}

			prevLocsExpAddr, err := vm.Memory.Read(prevLocsBitAddress.SegmentIndex, prevLocsBitAddress.Offset+expStructOffset)
			if err != nil {
				return err
			}

			prevLocsExp, err := prevLocsExpAddr.FieldElement()
			if err != nil {
				return err
			}

			var prevLocsExpBig big.Int
			prevLocsExp.BigInt(&prevLocsExpBig)
			locsBitBig := new(big.Int).And(&prevLocsExpBig, big.NewInt(1))
			v := memory.MemoryValueFromFieldElement(new(fp.Element).SetBigInt(locsBitBig))

			return vm.Memory.WriteToAddress(&locsBitAddress, &v)
		},
	}
}

func createPowHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	locs, err := resolver.GetCellRefer("locs")
	if err != nil {
		return nil, err
	}

	prev_locs, err := resolver.GetCellRefer("prev_locs")
	if err != nil {
		return nil, err
	}

	locsRes := hinter.Deref{Deref: locs}
	prevLocsRes := hinter.Deref{Deref: prev_locs}

	return newPowHint(locsRes, prevLocsRes), nil
}

// SplitFelt hint splits a finite field element into high and low components
//
// `newSplitFeltHint` takes 3 operanders as arguments
//   - `low` and `high` are the variables that will store the low and high components
//   - `value` is the variable to split
//
// `newSplitFeltHint` writes the low and high components in the `low` and `high`
// memory address, respectively
func newSplitFeltHint(low, high, value hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "SplitFelt",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.math_utils import assert_integer
			// assert ids.MAX_HIGH < 2**128 and ids.MAX_LOW < 2**128
			// assert PRIME - 1 == ids.MAX_HIGH * 2**128 + ids.MAX_LOW
			// assert_integer(ids.value)
			// ids.low = ids.value & ((1 << 128) - 1)
			// ids.high = ids.value >> 128

			//> assert ids.MAX_HIGH < 2**128 and ids.MAX_LOW < 2**128
			maxHigh := new(fp.Element).Div(new(fp.Element).SetInt64(-1), &utils.FeltMax128)
			maxLow := &utils.FeltZero

			//> assert PRIME - 1 == ids.MAX_HIGH * 2**128 + ids.MAX_LOW
			leftHandSide := new(fp.Element).SetInt64(-1)
			rightHandSide := new(fp.Element).Add(new(fp.Element).Mul(maxHigh, &utils.FeltMax128), maxLow)
			if leftHandSide.Cmp(rightHandSide) != 0 {
				return fmt.Errorf("assertion `split_felt(): The sum of MAX_HIGH and MAX_LOW does not equal to PRIME - 1` failed")
			}

			//> assert_integer(ids.value)
			value, err := hinter.ResolveAsFelt(vm, value)
			if err != nil {
				return err
			}

			var valueBigInt big.Int
			value.BigInt(&valueBigInt)
			lowAddr, err := low.GetAddress(vm)
			if err != nil {
				return err
			}

			highAddr, err := high.GetAddress(vm)
			if err != nil {
				return err
			}

			//> ids.low = ids.value & ((1 << 128) - 1)
			felt128 := new(big.Int).Lsh(big.NewInt(1), 128)
			felt128 = new(big.Int).Sub(felt128, big.NewInt(1))
			lowBigInt := new(big.Int).And(&valueBigInt, felt128)
			lowValue := memory.MemoryValueFromFieldElement(new(fp.Element).SetBigInt(lowBigInt))

			err = vm.Memory.WriteToAddress(&lowAddr, &lowValue)
			if err != nil {
				return err
			}

			//> ids.high = ids.value >> 128
			highBigInt := new(big.Int).Rsh(&valueBigInt, 128)
			highValue := memory.MemoryValueFromFieldElement(new(fp.Element).SetBigInt(highBigInt))

			return vm.Memory.WriteToAddress(&highAddr, &highValue)

		},
	}
}

func createSplitFeltHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
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

	return newSplitFeltHint(low, high, value), nil
}

// SignedDivRem hint computes a signed division and modulus operation on a given value,
// ensuring the quotient is within a specified range
//
// `newSignedDivRemHint` takes 5 operanders as arguments
//   - `value` is the dividende of the operation
//   - `div` is the divisor of the operation
//   - `bound` is the upper limit of the specified range
//   - `r` is the variable that will store the remainder of the modular division
//   - `biased_q` is the variable that will store the quotient of the modular division
//
// `newSignedDivRemHint` writes `r` and `biased_q` values to their respective memory address
func newSignedDivRemHint(value, div, bound, r, biased_q hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "SignedDivRem",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.math_utils import as_int, assert_integer
			//> assert_integer(ids.div)
			//> assert 0 < ids.div <= PRIME // range_check_builtin.bound, f'div={hex(ids.div)} is out of the valid range.'
			//> assert_integer(ids.bound)
			//> assert ids.bound <= range_check_builtin.bound // 2, f'bound={hex(ids.bound)} is out of the valid range.'
			//> int_value = as_int(ids.value, PRIME)
			//> q, ids.r = divmod(int_value, ids.div)
			//> assert -ids.bound <= q < ids.bound, f'{int_value} / {ids.div} = {q} is out of the range [{-ids.bound}, {ids.bound}).'
			//> ids.biased_q = q + ids.bound

			//> assert_integer(ids.div)
			divFelt, err := hinter.ResolveAsFelt(vm, div)
			if err != nil {
				return err
			}

			//> assert 0 < ids.div <= PRIME // range_check_builtin.bound, f'div={hex(ids.div)} is out of the valid range.'
			if divFelt.IsZero() || !utils.FeltLe(divFelt, &utils.PrimeHigh) {
				return fmt.Errorf("div=%v is out of the valid range", divFelt)
			}

			//> assert_integer(ids.bound)
			boundFelt, err := hinter.ResolveAsFelt(vm, bound)
			if err != nil {
				return err
			}

			//> assert ids.bound <= range_check_builtin.bound // 2, f'bound={hex(ids.bound)} is out of the valid range.'
			if !utils.FeltLe(boundFelt, &utils.Felt127) {
				return fmt.Errorf("bound=%v is out of the valid range", boundFelt)
			}

			//> int_value = as_int(ids.value, PRIME)
			valueFelt, err := hinter.ResolveAsFelt(vm, value)
			if err != nil {
				return err
			}

			intValueBig := math_utils.AsInt(valueFelt)

			//> q, ids.r = divmod(int_value, ids.div)
			var divBig, boundBig big.Int
			divFelt.BigInt(&divBig)
			boundFelt.BigInt(&boundBig)
			qBig, rBig := new(big.Int).DivMod(&intValueBig, &divBig, new(big.Int))
			rFelt := new(fp.Element).SetBigInt(rBig)
			rAddr, err := r.GetAddress(vm)
			if err != nil {
				return err
			}
			rValue := memory.MemoryValueFromFieldElement(rFelt)
			err = vm.Memory.WriteToAddress(&rAddr, &rValue)
			if err != nil {
				return err
			}

			//> assert -ids.bound <= q < ids.bound, f'{int_value} / {ids.div} = {q} is out of the range [{-ids.bound}, {ids.bound}).'
			if !(qBig.Cmp(new(big.Int).Neg(&boundBig)) >= 0 && qBig.Cmp(&boundBig) == -1) {
				return fmt.Errorf("%v / %v = %v is out of the range [-%v, %v]", valueFelt, divFelt, qBig, boundFelt, boundFelt)
			}

			//> ids.biased_q = q + ids.bound
			biasedQBig := new(big.Int).Add(qBig, &boundBig)
			biasedQ := new(fp.Element).SetBigInt(biasedQBig)
			biasedQAddr, err := biased_q.GetAddress(vm)
			if err != nil {
				return err
			}
			biasedQValue := memory.MemoryValueFromFieldElement(biasedQ)
			return vm.Memory.WriteToAddress(&biasedQAddr, &biasedQValue)
		},
	}
}

func createSignedDivRemHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	value, err := resolver.GetResOperander("value")
	if err != nil {
		return nil, err
	}

	div, err := resolver.GetResOperander("div")
	if err != nil {
		return nil, err
	}

	bound, err := resolver.GetResOperander("bound")
	if err != nil {
		return nil, err
	}

	r, err := resolver.GetResOperander("r")
	if err != nil {
		return nil, err
	}

	biased_q, err := resolver.GetResOperander("biased_q")
	if err != nil {
		return nil, err
	}

	return newSignedDivRemHint(value, div, bound, r, biased_q), nil
}

// Sqrt hint computes the square root of a given value,
// ensuring it falls within a specific range
//
// `newSqrtHint` takes 2 operanders as arguments
//   - `root` is the variable that will store the result
//   - `value` is the variable to operate on
//
// `newSqrtHint` writes the result of the hint at `root` address in memory
func newSqrtHint(root, value hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "Sqrt",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> from starkware.python.math_utils import isqrt
			// value = ids.value % PRIME
			// assert value < 2 ** 250, f"value={value} is outside of the range [0, 2**250)."
			// assert 2 ** 250 < PRIME
			// ids.root = isqrt(value)

			rootAddr, err := root.GetAddress(vm)
			if err != nil {
				return err
			}

			value, err := hinter.ResolveAsFelt(vm, value)
			if err != nil {
				return err
			}

			if !utils.FeltLt(value, &utils.FeltUpperBound) {
				return fmt.Errorf("assertion failed: %v is outside of the range [0, 2**250)", value)
			}

			// Conversion needed to handle non-square values
			valueU256 := uint256.Int(value.Bits())
			valueU256.Sqrt(&valueU256)

			result := fp.Element{}
			result.SetBytes(valueU256.Bytes())

			v := memory.MemoryValueFromFieldElement(&result)

			return vm.Memory.WriteToAddress(&rootAddr, &v)
		},
	}
}

func createSqrtHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	root, err := resolver.GetResOperander("root")
	if err != nil {
		return nil, err
	}

	value, err := resolver.GetResOperander("value")
	if err != nil {
		return nil, err
	}

	return newSqrtHint(root, value), nil
}

// UnsignedDivRem hint computes an unsigned division and modulus
// operation on a given dividend and divisor
//
// `newUnsignedDivRemHinter` takes 4 operanders as arguments
//   - `value` is the dividend of the operation
//   - `div` is the divisor of the operation
//   - `q` is the variable that will store the quotient of the modular division
//   - `r` is the variable that will store the remainder of the modular division
//
// `newUnsignedDivRemHinter` writes `q` and `r` values to their respective memory address

func newUnsignedDivRemHint(value, div, q, r hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "UnsignedDivRem",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.math_utils import assert_integer
			//> assert_integer(ids.div)
			//> assert 0 < ids.div <= PRIME // range_check_builtin.bound, \
			//>     f'div={hex(ids.div)} is out of the valid range.'
			//> ids.q, ids.r = divmod(ids.value, ids.div)

			value, err := hinter.ResolveAsFelt(vm, value)
			if err != nil {
				return err
			}

			div, err := hinter.ResolveAsFelt(vm, div)
			if err != nil {
				return err
			}

			qAddr, err := q.GetAddress(vm)
			if err != nil {
				return err
			}

			rAddr, err := r.GetAddress(vm)
			if err != nil {
				return err
			}

			// (PRIME // range_check_builtin.bound)
			// 800000000000011000000000000000000000000000000000000000000000001 // 2**128
			var divUpperBound big.Int
			utils.PrimeHigh.BigInt(&divUpperBound)

			var divBig big.Int
			div.BigInt(&divBig)

			if div.IsZero() || divBig.Cmp(&divUpperBound) == 1 {
				return fmt.Errorf("div=0x%v is out of the valid range", divBig.Text(16))
			}

			q, r := utils.FeltDivRem(value, div)

			qValue := memory.MemoryValueFromFieldElement(&q)
			if err := vm.Memory.WriteToAddress(&qAddr, &qValue); err != nil {
				return err
			}

			rValue := memory.MemoryValueFromFieldElement(&r)

			return vm.Memory.WriteToAddress(&rAddr, &rValue)
		},
	}
}

func createUnsignedDivRemHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	value, err := resolver.GetResOperander("value")
	if err != nil {
		return nil, err
	}

	div, err := resolver.GetResOperander("div")
	if err != nil {
		return nil, err
	}

	q, err := resolver.GetResOperander("q")
	if err != nil {
		return nil, err
	}

	r, err := resolver.GetResOperander("r")
	if err != nil {
		return nil, err
	}

	return newUnsignedDivRemHint(value, div, q, r), nil
}

func newIsQuadResidueHint(x, y hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "IsQuadResidue",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> from starkware.crypto.signature.signature import FIELD_PRIME
			//>	from starkware.python.math_utils import div_mod, is_quad_residue, sqrt
			//>	x = ids.x
			//>	if is_quad_residue(x, FIELD_PRIME):
			//>		ids.y = sqrt(x, FIELD_PRIME)
			//>	else:
			//>		ids.y = sqrt(div_mod(x, 3, FIELD_PRIME), FIELD_PRIME)

			x, err := hinter.ResolveAsFelt(vm, x)
			if err != nil {
				return err
			}

			yAddr, err := y.GetAddress(vm)
			if err != nil {
				return err
			}

			xBigInt := math_utils.AsInt(x)

			var value = memory.MemoryValue{}

			if x.IsZero() || x.IsOne() {
				value = memory.MemoryValueFromFieldElement(x)

			} else {
				var result *fp.Element = new(fp.Element)

				if x.Legendre() == 1 {
					// result = x.Sqrt(x)

					const primeString = "3618502788666131213697322783095070105623107215331596699973092056135872020481"
					primeBigInt, ok := new(big.Int).SetString(primeString, 10)
					if !ok {
						panic("failed to convert prime string to big.Int")
					}

					// divide primeBigInt by 2
					halfPrimeBigInt := new(big.Int).Rsh(primeBigInt, 1)

					tempResult := new(big.Int).ModSqrt(&xBigInt, primeBigInt)

					// ensures that tempResult is the smaller of the two possible square roots in the prime field.
					if tempResult.Cmp(halfPrimeBigInt) > 0 {
						tempResult.Sub(primeBigInt, tempResult)
					}

					result.SetBigInt(tempResult)

				} else {
					result = x.Sqrt(new(fp.Element).Div(x, new(fp.Element).SetUint64(3)))
				}

				value = memory.MemoryValueFromFieldElement(result)
			}

			return vm.Memory.WriteToAddress(&yAddr, &value)
		},
	}
}

func createIsQuadResidueHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	x, err := resolver.GetResOperander("x")
	if err != nil {
		return nil, err
	}
	y, err := resolver.GetResOperander("y")
	if err != nil {
		return nil, err
	}

	return newIsQuadResidueHint(x, y), nil
}
