package hintrunner

import (
	"fmt"

	h "github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
)

type HintRunner struct {
	// Execution context required by certain hints such as dictionaires
	context h.HintRunnerContext
	// A mapping from program counter to hint implementation
	hints map[uint64][]h.Hinter
}

func NewHintRunner(hints map[uint64][]h.Hinter) HintRunner {
	return HintRunner{
		// Context for certain hints that require it. Each manager is
		// initialized only when required by the hint
		context: *h.InitializeDefaultContext(),
		hints: hints,
	}
}

func (hr *HintRunner) RunHint(vm *VM.VirtualMachine) error {
	hints := hr.hints[vm.Context.Pc.Offset]
	if len(hints) == 0 {
		return nil
	}

	for _, hint := range hints {
		err := hint.Execute(vm, &hr.context)
		if err != nil {
			return fmt.Errorf("execute hint %s: %v", hint, err)
		}
	}

	return nil
}
