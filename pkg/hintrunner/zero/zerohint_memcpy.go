package zero

import (
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/core"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"

	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
)

func newMemcpyContinueCopyingHinter(output hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "MemcpyContinueCopying",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> n -= 1
			//> ids.continue_copying = 1 if n > 0 else 0
			var lhs fp.Element

			//> n-=1
			n, err := ctx.ScopeManager.GetVariableValue("n")
			if err != nil {
				return err
			}

			ctx.ScopeManager.AssignVariable("n", lhs.Sub(n.(*fp.Element), &utils.FeltOne))

			//> ids.continue_copying = 1 if n > 0 else 0
			continueCopyingAddr, err := output.GetAddress(vm)
			if err != nil {
				return err
			}

			var v memory.MemoryValue
			if !utils.FeltLt(&utils.FeltZero, n) {
				v = memory.MemoryValueFromFieldElement(&utils.FeltZero)
			} else {
				v = memory.MemoryValueFromFieldElement(&utils.FeltOne)
			}
			return vm.Memory.WriteToAddress(&continueCopyingAddr, &v)
		},
	}
}

func createMemcpyContinueCopyingHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	output, err := resolver.GetResOperander("continue_copying")
	if err != nil {
		return nil, err
	}
	return newMemcpyContinueCopyingHinter(output), nil
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

