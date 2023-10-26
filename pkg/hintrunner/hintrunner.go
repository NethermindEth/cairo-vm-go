package hintrunner

import (
	"fmt"

	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

type DictionaryManager struct {
	// Each Dictionary belongs to its own segment, so given a memory address
	// to a dictionary we can select the right one given the address index

	// dict segment index -> dict key -> dict value
	dictionaries map[uint64]map[f.Element]*mem.MemoryValue
}

// It creates a new segment which will hold dictionary values. It returns the memory
// address of that segment.
func (dm *DictionaryManager) NewDictionary(vm *VM.VirtualMachine) mem.MemoryAddress {
	newDictAddr := vm.Memory.AllocateEmptySegment()
	dm.dictionaries[newDictAddr.SegmentIndex] = make(map[f.Element]*mem.MemoryValue)
	return newDictAddr
}

func (dm *DictionaryManager) At(dictAddr *mem.MemoryAddress, key *f.Element) (*mem.MemoryValue, error) {
	if dict, ok := dm.dictionaries[dictAddr.SegmentIndex]; ok {
		return dict[*key], nil
	}
	return nil, fmt.Errorf("no dictionary at address %s", dictAddr)
}

func (dm *DictionaryManager) Set(dictAddr *mem.MemoryAddress, key *f.Element, value *mem.MemoryValue) error {
	if dict, ok := dm.dictionaries[dictAddr.SegmentIndex]; ok {
		dict[*key] = value
		return nil
	}
	return fmt.Errorf("no dictionary at address %s", dictAddr)
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
