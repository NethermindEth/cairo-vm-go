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
			if err != nil {
				return err
			}

			var v memory.MemoryValue
			if utils.FeltLe(aFelt, bFelt) {
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
			if err != nil {
				return err
			}

			if !utils.FeltLt(aFelt, bFelt) {
				return fmt.Errorf("a = %v is not less than b = %v", aFelt, bFelt)
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
			a, err := a.Resolve(vm)
			if err != nil {
				return err
			}
			// aFelt is already modulo PRIME, no need to adjust it.
			aFelt, err := a.FieldElement()
			if err != nil {
				return err
			}
			// range_check_builtin.bound is utils.FeltMax128 (1 << 128).
			var v memory.MemoryValue
			if utils.FeltLt(aFelt, &utils.FeltMax128) {
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
			a, err := a.Resolve(vm)
			if err != nil {
				return err
			}
			aFelt, err := a.FieldElement()
			if err != nil {
				return err
			}
			var lhs fp.Element
			lhs.Sub(&utils.FeltZero, aFelt) //> -ids.a
			lhs.Sub(&lhs, &utils.FeltOne)
			var v memory.MemoryValue
			if utils.FeltLt(aFelt, &utils.FeltMax128) {
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
