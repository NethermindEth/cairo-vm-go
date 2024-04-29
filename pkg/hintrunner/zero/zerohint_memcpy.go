package zero

import (
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/core"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"

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

			n = n.(uint64) - 1
			if err := ctx.ScopeManager.AssignVariable("n", n); err != nil {
				return err
			}

			//> ids.continue_copying = 1 if n > 0 else 0
			continueCopyingAddr, err := continueCopying.GetAddress(vm)
			if err != nil {
				return err
			}

			var continueCopyingMv memory.MemoryValue
			if n.(uint64) > 0 {
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

func createAllocSegmentHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	return &core.AllocSegment{Dst: hinter.ApCellRef(0)}, nil
}

func createVMEnterScopeHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	return &GenericZeroHinter{
		Name: "VMEnterScope",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			ctx.ScopeManager.EnterScope(make(map[string]any))
			return nil
		},
	}, nil
}

func createVMExitScopeHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	return &GenericZeroHinter{
		Name: "VMExitScope",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			return ctx.ScopeManager.ExitScope()
		},
	}, nil
}

