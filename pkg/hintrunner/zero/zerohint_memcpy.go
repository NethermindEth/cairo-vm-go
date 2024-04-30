package zero

import (
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/core"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
)

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

func newMemcpyEnterScopeHint(len hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "MemcpyEnterScope",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//>  vm_enter_scope({'n': ids.len})
			len, err := hinter.ResolveAsFelt(vm, len)
			if err != nil {
				return err
			}
			ctx.ScopeManager.EnterScope(map[string]any{"n": len})
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

func createVMExitScopeHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	return &GenericZeroHinter{
		Name: "VMExitScope",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			return ctx.ScopeManager.ExitScope()
		},
	}, nil
}
