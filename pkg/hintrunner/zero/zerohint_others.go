package zero

import (
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/core"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"

	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func newMemContinueHint(continueTarget hinter.ResOperander, memset bool) hinter.Hinter {
	var name string
	if memset {
		name = "MemsetContinueLoop"
	} else {
		name = "MemcpyContinueCopying"
	}
	return &GenericZeroHinter{
		Name: name,
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			// MemsetContinueLoop
			//> n -= 1
			//> ids.continue_loop = 1 if n > 0 else 0

			// MemcpyContinueCopying
			//> n -= 1
			//> ids.continue_copying = 1 if n > 0 else 0

			//> n-=1
			n, err := ctx.ScopeManager.GetVariableValue("n")
			if err != nil {
				return err
			}

			newN := new(f.Element)
			newN = newN.Sub(n.(*f.Element), &utils.FeltOne)

			if err := ctx.ScopeManager.AssignVariable("n", newN); err != nil {
				return err
			}

			//> ids.continue_loop/continue_copying = 1 if n > 0 else 0
			continueTargetAddr, err := continueTarget.GetAddress(vm)
			if err != nil {
				return err
			}

			var continueTargetMv memory.MemoryValue
			if utils.FeltLt(&utils.FeltZero, newN) {
				continueTargetMv = memory.MemoryValueFromFieldElement(&utils.FeltOne)
			} else {
				continueTargetMv = memory.MemoryValueFromFieldElement(&utils.FeltZero)
			}

			return vm.Memory.WriteToAddress(&continueTargetAddr, &continueTargetMv)
		},
	}
}

func createMemContinueHinter(resolver hintReferenceResolver, memset bool) (hinter.Hinter, error) {
	var continueTarget hinter.ResOperander
	var err error
	if memset {
		continueTarget, err = resolver.GetResOperander("continue_loop")
	} else {
		continueTarget, err = resolver.GetResOperander("continue_copying")
	}
	if err != nil {
		return nil, err
	}
	return newMemContinueHint(continueTarget, memset), nil
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

// MemEnterScope hint enters a new scope for the memory copy/set operation with a specified value
//
// `newMemEnterScopeHint` takes 2 operanders as arguments
//   - `value` is the value that is added in the new scope
//   - `memset` specifies whether it's a memset or memcpy operation
func newMemEnterScopeHint(value hinter.ResOperander, memset bool) hinter.Hinter {
	var name string
	if memset {
		name = "MemsetEnterScope"
	} else {
		name = "MemcpyEnterScope"
	}
	return &GenericZeroHinter{
		Name: name,
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			// MemsetEnterScope
			//> vm_enter_scope({'n': ids.n})

			// MemcpyEnterScope
			//> vm_enter_scope({'n': ids.len})

			value, err := hinter.ResolveAsFelt(vm, value)
			if err != nil {
				return err
			}

			ctx.ScopeManager.EnterScope(map[string]any{"n": *value})
			return nil
		},
	}
}

func createMemEnterScopeHinter(resolver hintReferenceResolver, memset bool) (hinter.Hinter, error) {
	var value hinter.ResOperander
	var err error
	if memset {
		value, err = resolver.GetResOperander("n")
	} else {
		value, err = resolver.GetResOperander("len")
	}
	if err != nil {
		return nil, err
	}
	return newMemEnterScopeHint(value, memset), nil
}
