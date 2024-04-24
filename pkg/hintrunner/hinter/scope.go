package hinter

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

// ScopeManager handles all operations regarding scopes:
// - Creating a new scope
// - Exiting current scope
// - Variable declaration and assignment inside a certain scope
// - Accessing variable values

type ScopeValue struct {
	Type        int
	FeltValue   *fp.Element
	BigIntValue *big.Int
}

type ScopeMap map[string]ScopeValue

type ScopeValueType int

const (
	TypeNone = iota
	TypeFeltValue
	TypeBigInt
)

type ScopeManager struct {
	scopes []ScopeMap
}

func InitializeScopeManager(ctx *HintRunnerContext, scope ScopeMap) {
	if ctx.ScopeManager.scopes == nil {
		ctx.ScopeManager = *NewScopeManager(scope)
	}
}

func NewScopeManager(globals ScopeMap) *ScopeManager {

	return &ScopeManager{
		scopes: []ScopeMap{
			// One scope needed (current execution scope)
			globals,
		},
	}
}

func DefaultNewScopeManager() *ScopeManager {
	return NewScopeManager(make(ScopeMap))
}

func (sm *ScopeManager) EnterScope(newScope ScopeMap) {

	sm.scopes = append(sm.scopes, newScope)
}

func (sm *ScopeManager) ExitScope() error {
	if len(sm.scopes) < 2 {
		return fmt.Errorf("expected at least one existing scope")
	}
	sm.scopes = sm.scopes[:(len(sm.scopes) - 1)]
	return nil
}

func (sm *ScopeManager) AssignVariable(name string, value ScopeValue) error {
	scope, err := sm.getCurrentScope()
	if err != nil {
		return err
	}

	(*scope)[name] = value
	return nil
}

func (sm *ScopeManager) AssignVariables(values ScopeMap) error {
	for name, value := range values {
		err := sm.AssignVariable(name, value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sm *ScopeManager) GetVariableValue(name string) (*ScopeValue, error) {
	scope, err := sm.getCurrentScope()
	if err != nil {
		return nil, err
	}

	if value, ok := (*scope)[name]; ok {
		return &value, nil
	}

	return nil, fmt.Errorf("variable %s not found in current scope", name)
}

func (sm *ScopeManager) GetVariableValueAsBigInt(name string) (*big.Int, error) {
	value, err := sm.GetVariableValue(name)
	if err != nil {
		return nil, err
	}

	// valueBig, ok := value.(*big.Int)
	// if !ok {
	// 	return nil, fmt.Errorf("value: %s is not a *big.Int", value)
	// }

	return value.BigIntValue, nil
}

func (sm *ScopeManager) getCurrentScope() (*ScopeMap, error) {
	if len(sm.scopes) == 0 {
		return nil, fmt.Errorf("expected at least one existing scope")
	}
	return &sm.scopes[len(sm.scopes)-1], nil
}

func BigIntScopeValue(value *big.Int) *ScopeValue {
	return &ScopeValue{Type: TypeBigInt, BigIntValue: value}
}

func FeltValueScopeValue(value *fp.Element) *ScopeValue {
	return &ScopeValue{Type: TypeFeltValue, FeltValue: value}
}
