package hinter

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScope(t *testing.T) {
	sm := DefaultNewScopeManager()

	// Assing variable n and get its current value
	err := sm.AssignVariable("n", *BigIntScopeValue(big.NewInt(3)))
	require.NoError(t, err)

	n, err := sm.GetVariableValue("n")
	require.NoError(t, err)
	require.Equal(t, 3, int(n.BigIntValue.Int64()))

	// Creating new scope with another value for variable n
	// This variable should shadow the one in the previous scope
	sm.EnterScope(ScopeMap{"n": *BigIntScopeValue(big.NewInt(5))})
	n, err = sm.GetVariableValue("n")
	require.NoError(t, err)
	require.Equal(t, 5, int(n.BigIntValue.Int64()))

	// Try to get the value of a variable that has not been defined
	_, err = sm.GetVariableValue("x")
	require.ErrorContains(t, err, "variable x not found in current scope")

	// Exit current scope and check for the value of n again
	err = sm.ExitScope()
	require.NoError(t, err)
	n, err = sm.GetVariableValue("n")
	require.NoError(t, err)
	require.Equal(t, 3, int(n.BigIntValue.Int64()))

	// Try exiting main scope should error out
	err = sm.ExitScope()
	require.ErrorContains(t, err, "expected at least one existing scope")
}
