package zero

import (
	"fmt"

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

<<<<<<< HEAD
			a, err := resolveFieldElement(vm, a)
			if err != nil {
				return err
			}
			b, err := resolveFieldElement(vm, b)
=======
<<<<<<< Updated upstream
			a, err := a.Resolve(vm)
			if err != nil {
				return err
			}
			aFelt, err := a.FieldElement()
			if err != nil {
				return err
			}
			b, err := b.Resolve(vm)
			if err != nil {
				return err
			}
			bFelt, err := b.FieldElement()
=======
			a, err := hinter.ResolveAsFelt(vm, a)
			if err != nil {
				return err
			}
			b, err := hinter.ResolveAsFelt(vm, b)
>>>>>>> Stashed changes
>>>>>>> 6f9e68d (replace resolveAsFieldElement as hinter.ResolveAsFelt)
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
<<<<<<< HEAD
			a, err := resolveFieldElement(vm, a)
			if err != nil {
				return err
			}
			b, err := resolveFieldElement(vm, b)
=======
<<<<<<< Updated upstream
			a, err := a.Resolve(vm)
			if err != nil {
				return err
			}
			aFelt, err := a.FieldElement()
			if err != nil {
				return err
			}
			b, err := b.Resolve(vm)
			if err != nil {
				return err
			}
			bFelt, err := b.FieldElement()
=======
			a, err := hinter.ResolveAsFelt(vm, a)
			if err != nil {
				return err
			}
			b, err := hinter.ResolveAsFelt(vm, b)
>>>>>>> Stashed changes
>>>>>>> 6f9e68d (replace resolveAsFieldElement as hinter.ResolveAsFelt)
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
<<<<<<< HEAD

			// a is already modulo PRIME, no need to adjust it.
			a, err := resolveFieldElement(vm, a)
=======
<<<<<<< Updated upstream
			a, err := a.Resolve(vm)
			if err != nil {
				return err
			}
			// aFelt is already modulo PRIME, no need to adjust it.
			aFelt, err := a.FieldElement()
=======

			// a is already modulo PRIME, no need to adjust it.
			a, err := hinter.ResolveAsFelt(vm, a)
>>>>>>> Stashed changes
>>>>>>> 6f9e68d (replace resolveAsFieldElement as hinter.ResolveAsFelt)
			if err != nil {
				return err
			}
			// range_check_builtin.bound is utils.FeltMax128 (1 << 128).
			var v memory.MemoryValue
			if utils.FeltLt(a, &utils.FeltMax128) {
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
<<<<<<< HEAD
			a, err := resolveFieldElement(vm, a)
=======
<<<<<<< Updated upstream
			a, err := a.Resolve(vm)
			if err != nil {
				return err
			}
			aFelt, err := a.FieldElement()
=======
			a, err := hinter.ResolveAsFelt(vm, a)
>>>>>>> Stashed changes
>>>>>>> 6f9e68d (replace resolveAsFieldElement as hinter.ResolveAsFelt)
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

<<<<<<< HEAD
			value, err := resolveFieldElement(vm, value)
=======
<<<<<<< Updated upstream
			value, err := value.Resolve(vm)
			if err != nil {
				return err
			}
			valueFelt, err := value.FieldElement()
=======
			value, err := hinter.ResolveAsFelt(vm, value)
>>>>>>> Stashed changes
>>>>>>> 6f9e68d (replace resolveAsFieldElement as hinter.ResolveAsFelt)
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

<<<<<<< HEAD
			value, err := resolveFieldElement(vm, value)
=======
<<<<<<< Updated upstream
			value, err := value.Resolve(vm)
=======
			value, err := hinter.ResolveAsFelt(vm, value)
>>>>>>> Stashed changes
>>>>>>> 6f9e68d (replace resolveAsFieldElement as hinter.ResolveAsFelt)
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

<<<<<<< HEAD
			base, err := resolveFieldElement(vm, base)
=======
<<<<<<< Updated upstream
			base, err := base.Resolve(vm)
			if err != nil {
				return err
			}
			baseFelt, err := base.FieldElement()
=======
			base, err := hinter.ResolveAsFelt(vm, base)
>>>>>>> Stashed changes
>>>>>>> 6f9e68d (replace resolveAsFieldElement as hinter.ResolveAsFelt)
			if err != nil {
				return err
			}

<<<<<<< HEAD
			value, err := resolveFieldElement(vm, value)
=======
<<<<<<< Updated upstream
			value, err := value.Resolve(vm)
			if err != nil {
				return err
			}
			valueFelt, err := value.FieldElement()
=======
			value, err := hinter.ResolveAsFelt(vm, value)
>>>>>>> Stashed changes
>>>>>>> 6f9e68d (replace resolveAsFieldElement as hinter.ResolveAsFelt)
			if err != nil {
				return err
			}

<<<<<<< HEAD
			bound, err := resolveFieldElement(vm, bound)
=======
<<<<<<< Updated upstream
			bound, err := bound.Resolve(vm)
			if err != nil {
				return err
			}
			boundFelt, err := bound.FieldElement()
=======
			bound, err := hinter.ResolveAsFelt(vm, bound)
>>>>>>> Stashed changes
>>>>>>> 6f9e68d (replace resolveAsFieldElement as hinter.ResolveAsFelt)
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
