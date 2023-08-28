package hintrunner

import (
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"golang.org/x/exp/constraints"
)

// todo: Can two or more hints be assigned to a specific PC?
type HintRunner struct {
	// A mapping from program counter to hint implementation
	hints map[uint64]Hinter
}

func CreateHintRunner() *HintRunner {
	return nil
}

func (hr HintRunner) RunHint(vm *VM.VirtualMachine) *HintRunnerError {
	hint := hr.hints[vm.Context.Pc]
	if hint == nil {
		return nil
	}

	err := hint.Execute(vm)
	if err != nil {
		return NewHintRunnerError(err)
	}
	return nil
}

// move this function to a possible util packages
func Abs[T constraints.Signed](num T) T {
	if num >= 0 {
		return num
	}
	return num * -1
}
