package hinter

import (
	"fmt"

	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
)

type Hinter interface {
	fmt.Stringer

	Execute(vm *VM.VirtualMachine, ctx *HintRunnerContext) error
}

// Global context to keep track of different results across different
// hints execution.
type HintRunnerContext struct {
	DictionaryManager         DictionaryManager
	SquashedDictionaryManager SquashedDictionaryManager
	ScopeManager              ScopeManager
	// points towards free memory of a segment
	ConstantSizeSegment mem.MemoryAddress
}

func InitializeDefaultContext() *HintRunnerContext {
	return &HintRunnerContext{
		DictionaryManager:         DictionaryManager{},
		SquashedDictionaryManager: SquashedDictionaryManager{},
		ScopeManager:              *DefaultNewScopeManager(),
		ConstantSizeSegment:       mem.UnknownAddress,
	}
}

func SetContextWithScope(scope map[string]any) *HintRunnerContext {
	ctx := HintRunnerContext{
		ScopeManager: *NewScopeManager(scope),
	}
	return &ctx
}
