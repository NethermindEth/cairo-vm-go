package hintrunner

import (
	"fmt"

	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
)

// todo: Can two or more hints be assigned to a specific PC?
type HintRunner struct {
	// A mapping from program counter to hint implementation
	hints map[uint64]Hinter
}

func NewHintRunner(hints map[uint64]Hinter) HintRunner {
	return HintRunner{hints}
}

func (hr HintRunner) RunHint(vm *VM.VirtualMachine) error {
	hint := hr.hints[vm.Context.Pc.Offset]
	if hint == nil {
		return nil
	}

	err := hint.Execute(vm)
	if err != nil {
		return fmt.Errorf("execute hint %s: %v", hint, err)
	}
	return nil
}
