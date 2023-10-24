package hintrunner

import (
	"fmt"

	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
)

type DictionaryManager struct {
}

// It creates a new dictionary
func (dm DictionaryManager) NewDictionary(vm *VM.VirtualMachine) mem.MemoryAddress {
	return mem.MemoryAddress{
		SegmentIndex: uint64(vm.Memory.AllocateEmptySegment()),
		Offset:       0,
	}
}

type HintRunnerContext struct {
	DictionaryManager DictionaryManager
}

// todo: Can two or more hints be assigned to a specific PC?
type HintRunner struct {
	// Execution context required by certain hints such as dictionaires
	context HintRunnerContext
	// A mapping from program counter to hint implementation
	hints map[uint64]Hinter
}

func NewHintRunner(hints map[uint64]Hinter) HintRunner {
	return HintRunner{
		context: HintRunnerContext{
			DictionaryManager{},
		},
		hints: hints,
	}
}

func (hr *HintRunner) RunHint(vm *VM.VirtualMachine) error {
	hint := hr.hints[vm.Context.Pc.Offset]
	if hint == nil {
		return nil
	}

	err := hint.Execute(vm, &hr.context)
	if err != nil {
		return fmt.Errorf("execute hint %s: %v", hint, err)
	}
	return nil
}
