package hinter

import (
	"fmt"

	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
)

// Global context to keep track of different results across different
// hints execution.
type HintRunnerContext struct {
	DictionaryManager         DictionaryManager
	SquashedDictionaryManager SquashedDictionaryManager
	ScopeManager              ScopeManager
	// points towards free memory of a segment
	ConstantSizeSegment mem.MemoryAddress
}

// ScopeManager handles all operations regarding scopes:
// - Creating a new scope
// - Exiting current scope
// - Variable declaration and assignment inside a certain scope
// - Accessing variable values
type ScopeManager struct {
	Scopes []map[string]any
}

func InitializeScopeManager() *ScopeManager {
	return &ScopeManager{
		Scopes: []map[string]any {
			// One scope needed (current execution scope)
			make(map[string]any),
		},
	}
}

func (sm *ScopeManager) EnterScope(newScope map[string]any) {
	sm.Scopes = append(sm.Scopes, newScope)
}

func (sm *ScopeManager) ExitScope() error {
	if len(sm.Scopes) < 2 {
		return fmt.Errorf("expected at least one existing scope")
	}
	sm.Scopes = sm.Scopes[:(len(sm.Scopes) - 1)]
	return nil
}

func (sm *ScopeManager) AssignVariable(name string, value any) error {
	scope, err := sm.GetCurrentScope()
	if err != nil {
		return err
	}

	(*scope)[name] = value
	return nil
}

func (sm *ScopeManager) GetVariableValue(name string) (any, error) {
	scope, err := sm.GetCurrentScope()
	if err != nil {
		return nil, err
	}

	if value, ok := (*scope)[name]; ok {
		return value, nil
	}

	return nil, fmt.Errorf("variable %s not found in current scope", name)
}

func (sm *ScopeManager) GetCurrentScope() (*map[string]any, error) {
	if len(sm.Scopes) == 0 {
		return nil, fmt.Errorf("expected at least one existing scope")
	}
	return &sm.Scopes[len(sm.Scopes) - 1], nil
}