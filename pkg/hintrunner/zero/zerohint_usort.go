package zero

import (
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
)

func newUsortEnterScopeHinter() hinter.Hinter {
	return &GenericZeroHinter{
		Name: "UsortEnterScope",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			__usort_max_size, err := ctx.ScopeManager.GetVariableValue("__usort_max_size")
			if err != nil {
				return err
			}

			ctx.ScopeManager.EnterScope(map[string]any{
				"__usort_max_size": __usort_max_size,
			})

			return nil
		},
	}
}

func createUsortEnterScopeHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	return newUsortEnterScopeHinter(), nil
}
