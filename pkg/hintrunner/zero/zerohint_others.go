package zero

import (
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/core"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"

	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
)

func newMemcpyContinueCopyingHint(continueCopying hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "MemcpyContinueCopying",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> n -= 1
			//> ids.continue_copying = 1 if n > 0 else 0

			//> n-=1
			n, err := ctx.ScopeManager.GetVariableValue("n")
			if err != nil {
				return err
			}

			newN := new(fp.Element)
			newN = newN.Sub(n.(*fp.Element), &utils.FeltOne)

			if err := ctx.ScopeManager.AssignVariable("n", newN); err != nil {
				return err
			}

			//> ids.continue_copying = 1 if n > 0 else 0
			continueCopyingAddr, err := continueCopying.GetAddress(vm)
			if err != nil {
				return err
			}

			var continueCopyingMv memory.MemoryValue
			if utils.FeltLt(&utils.FeltZero, newN) {
				continueCopyingMv = memory.MemoryValueFromFieldElement(&utils.FeltOne)
			} else {
				continueCopyingMv = memory.MemoryValueFromFieldElement(&utils.FeltZero)
			}

			return vm.Memory.WriteToAddress(&continueCopyingAddr, &continueCopyingMv)
		},
	}
}

func createMemcpyContinueCopyingHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	continueCopying, err := resolver.GetResOperander("continue_copying")
	if err != nil {
		return nil, err
	}
	return newMemcpyContinueCopyingHint(continueCopying), nil
}

// AllocSegment hint adds a new segment to the Cairo VM memory
func createAllocSegmentHinter() (hinter.Hinter, error) {
	return &core.AllocSegment{Dst: hinter.ApCellRef(0)}, nil
}

// VMEnterScope hint enters a new scope in the Cairo VM
func createVMEnterScopeHinter() (hinter.Hinter, error) {
	return &GenericZeroHinter{
		Name: "VMEnterScope",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			ctx.ScopeManager.EnterScope(make(map[string]any))
			return nil
		},
	}, nil
}

// VMExitScop hint exits the current scope in the Cairo VM
func createVMExitScopeHinter() (hinter.Hinter, error) {
	return &GenericZeroHinter{
		Name: "VMExitScope",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			return ctx.ScopeManager.ExitScope()
		},
	}, nil
}

// MemcpyEnterScope hint enters a new scope for the memory copy operation with a specified length
//
// `newMemcpyEnterScopeHint` takes 1 operander as argument
//   - `len` is the length value that is added in the new scope
func newMemcpyEnterScopeHint(len hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "MemcpyEnterScope",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//>  vm_enter_scope({'n': ids.len})
			len, err := hinter.ResolveAsFelt(vm, len)
			if err != nil {
				return err
			}

			ctx.ScopeManager.EnterScope(map[string]any{"n": *len})
			return nil
		},
	}
}

func createMemcpyEnterScopeHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	len, err := resolver.GetResOperander("len")
	if err != nil {
		return nil, err
	}
	return newMemcpyEnterScopeHint(len), nil
}
