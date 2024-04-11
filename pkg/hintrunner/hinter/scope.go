package hinter

import (
	"fmt"
	"math/big"
)

// ScopeManager handles all operations regarding scopes:
// - Creating a new scope
// - Exiting current scope
// - Variable declaration and assignment inside a certain scope
// - Accessing variable values
type ScopeManager struct {
	scopes []map[string]any
}

func InitializeScopeManager(ctx *HintRunnerContext, scope map[string]any) {
	if ctx.ScopeManager.scopes == nil {
		ctx.ScopeManager = *NewScopeManager(scope)
	}
}

func NewScopeManager(globals map[string]any) *ScopeManager {
	return &ScopeManager{
		scopes: []map[string]any{
			// One scope needed (current execution scope)
			globals,
		},
	}
}

func DefaultNewScopeManager() *ScopeManager {
	return NewScopeManager(make(map[string]any))
}

func (sm *ScopeManager) EnterScope(newScope map[string]any) {
	sm.scopes = append(sm.scopes, newScope)
}

func (sm *ScopeManager) ExitScope() error {
	if len(sm.scopes) < 2 {
		return fmt.Errorf("expected at least one existing scope")
	}
	sm.scopes = sm.scopes[:(len(sm.scopes) - 1)]
	return nil
}

func (sm *ScopeManager) AssignVariable(name string, value any) error {
	scope, err := sm.getCurrentScope()
	if err != nil {
		return err
	}

	(*scope)[name] = value
	return nil
}

func (sm *ScopeManager) AssignVariables(values map[string]any) error {
	for name, value := range values {
		err := sm.AssignVariable(name, value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sm *ScopeManager) DeleteVariable(name string) error {
	scope, err := sm.getCurrentScope()
	if err != nil {
		return err
	}

	delete(*scope, name)
	return nil
}

func (sm *ScopeManager) GetVariableValue(name string) (any, error) {
	scope, err := sm.getCurrentScope()
	if err != nil {
		return nil, err
	}

	if value, ok := (*scope)[name]; ok {
		return value, nil
	}

	return nil, fmt.Errorf("variable %s not found in current scope", name)
}

func (sm *ScopeManager) GetVariableValueAsBigInt(name string) (*big.Int, error) {
	value, err := sm.GetVariableValue(name)
	if err != nil {
		return nil, err
	}

	valueBig, ok := value.(*big.Int)
	if !ok {
		return nil, fmt.Errorf("value: %s is not a *big.Int", value)
	}

	return valueBig, nil
}

func (sm *ScopeManager) getCurrentScope() (*map[string]any, error) {
	if len(sm.scopes) == 0 {
		return nil, fmt.Errorf("expected at least one existing scope")
	}
	return &sm.scopes[len(sm.scopes)-1], nil
}
