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
					err := ctx.ScopeManager.AssignVariable("__usort_max_size", feltUint64(1))
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUsortEnterScopeHint()
				},
				check: varValueInScopeEquals("__usort_max_size", feltUint64(1)),
			},
		},
		"UsortVerifyMultiplicityAssert": {
			{
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("positions", []uint64{1})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUsortVerifyMultiplicityAssertHint()
				},
				errCheck: errorTextContains("assertion `len(positions) == 0` failed"),
			},
			{
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("positions", []uint64{})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUsortVerifyMultiplicityAssertHint()
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
						t.Fatal(err)
					}
				},
				operanders: []*hintOperander{
					{Name: "value", Kind: fpRelative, Value: feltUint64(0)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUsortVerifyHint(ctx.operanders["value"])
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
						"positions": []fp.Element{*feltInt64(8), *feltInt64(6), *feltInt64(4)},
						"last_pos":  *feltInt64(2),
					})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUsortVerifyMultiplicityBodyHint(ctx.operanders["next_item_index"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					allVarValueInScopeEquals(map[string]any{
						"last_pos": *feltInt64(5), "positions": []fp.Element{*feltInt64(8), *feltInt64(6)},
					})(t, ctx)

					varValueEquals("next_item_index", feltInt64(2))(t, ctx)
				},
			},
			{
				operanders: []*hintOperander{
					{Name: "next_item_index", Kind: uninitialized},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{
						"positions": []fp.Element{*feltInt64(90), *feltInt64(80), *feltInt64(70), *feltInt64(60), *feltInt64(50)},
						"last_pos":  *feltInt64(0),
					})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUsortVerifyMultiplicityBodyHint(ctx.operanders["next_item_index"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					allVarValueInScopeEquals(map[string]any{
						"last_pos": *feltInt64(51), "positions": []fp.Element{*feltInt64(90), *feltInt64(80), *feltInt64(70), *feltInt64(60)},
					})(t, ctx)

					varValueEquals("next_item_index", feltInt64(50))(t, ctx)
				},
			},
			{
				operanders: []*hintOperander{
					{Name: "next_item_index", Kind: uninitialized},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{
						"positions": []fp.Element{*feltInt64(87), *feltInt64(51), *feltInt64(43), *feltInt64(37)},
						"last_pos":  *feltInt64(29),
					})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUsortVerifyMultiplicityBodyHint(ctx.operanders["next_item_index"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					allVarValueInScopeEquals(map[string]any{
						"last_pos": *feltInt64(38), "positions": []fp.Element{*feltInt64(87), *feltInt64(51), *feltInt64(43)},
					})(t, ctx)

					varValueEquals("next_item_index", feltInt64(8))(t, ctx)
				},
			},
		},
	})
}
