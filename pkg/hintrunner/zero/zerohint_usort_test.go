package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/require"
)

func TestZeroHintUsort(t *testing.T) {
	runHinterTests(t, map[string][]hintTestCase{
		"UsortEnterScope": {
			{
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{
						"__usort_max_size": feltUint64(1),
					})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUsortEnterScopeHinter()
				},
				check: varValueInScopeEquals("__usort_max_size", feltUint64(1)),
			},
		},
		"UsortVerifyMultiplicityAssert": {
			{
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("positions", []uint64{1})
					if err != nil {
						panic(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUsortVerifyMultiplicityAssertHinter()
				},
				errCheck: errorTextContains("assertion `len(positions) == 0` failed"),
			},
			{
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("positions", []uint64{})
					if err != nil {
						panic(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUsortVerifyMultiplicityAssertHinter()
				},
				errCheck: errorIsNil,
			},
		},
		"UsortVerify": {
			{
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("positions_dict", map[fp.Element][]uint64{
						*feltUint64(0): {1, 2, 3},
					})
					if err != nil {
						panic(err)
					}
				},
				operanders: []*hintOperander{
					{Name: "value", Kind: fpRelative, Value: feltUint64(0)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUsortVerifyHinter(ctx.operanders["value"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					positions, err := ctx.runnerContext.ScopeManager.GetVariableValue("positions")
					require.NoError(t, err)

					require.Equal(t, []uint64{3, 2, 1}, positions)

					lastPos, err := ctx.runnerContext.ScopeManager.GetVariableValue("last_pos")
					require.NoError(t, err)

					require.Equal(t, 0, lastPos)
				},
			},
		},
		"UsortVerifyMultiplicityBody": {
			// Tests when no variables (positions, last_pos) are in the scope.
			{
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUsortVerifyMultiplicityBodyHint(ctx.operanders["next_item_index"])
				},
				errCheck: func(t *testing.T, ctx *hintTestContext, err error) {
					require.NotNil(t, err)
				},
			},
			// Tests when we can calculate new memory and variable values.
			{
				operanders: []*hintOperander{
					{Name: "next_item_index", Kind: uninitialized},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{
						"positions":   []int64{8, 6, 4},
						"current_pos": int64(2),
						"last_pos":    int64(1),
					})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUsortVerifyMultiplicityBodyHint(ctx.operanders["next_item_index"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					allVarValueInScopeEquals(map[string]any{
						"current_pos": int64(4),
						"last_pos":    int64(3),
					})(t, ctx)

					varValueEquals("next_item_index", feltInt64(1))(t, ctx)
				},
			},
		},
	})
}
