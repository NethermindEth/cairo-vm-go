package zero

import (
	"fmt"
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/require"
)

func TestZeroHintUsort(t *testing.T) {
	runHinterTests(t, map[string][]hintTestCase{
		"UsortEnterScope": {
			{
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{
						"__usort_max_size": uint64(1),
					})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUsortEnterScopeHinter()
				},
				check: varValueInScopeEquals("__usort_max_size", uint64(1)),
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
					return newUsortVerifyMultiplicityAssertHinter()
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
						t.Fatal(err)
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
		"UsortBody": {
			{
				// input length greater then allowed size
				operanders: []*hintOperander{
					{Name: "input", Kind: apRelative, Value: addr(7)},
					{Name: "input_length", Kind: apRelative, Value: feltUint64(20)},
					{Name: "output", Kind: uninitialized},
					{Name: "output_length", Kind: uninitialized},
					{Name: "multiplicities", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUsortBodyHint(ctx.operanders["input"], ctx.operanders["input_length"], ctx.operanders["output"], ctx.operanders["output_length"], ctx.operanders["multiplicites"])
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{
						"__usort_max_size": uint64(10),
					})
				},
				errCheck: errorTextContains(fmt.Sprintf("usort() can only be used with input_len<=%d.\n Got: input_len=%d", 10, 20)),
			},
			{
				// sort items with multiplicity of 1
				operanders: []*hintOperander{
					{Name: "input", Kind: apRelative, Value: addr(5)},
					{Name: "input.el0", Kind: apRelative, Value: feltUint64(2)},
					{Name: "input.el1", Kind: apRelative, Value: feltUint64(3)},
					{Name: "input.el2", Kind: apRelative, Value: feltUint64(1)},
					{Name: "input_length", Kind: apRelative, Value: feltUint64(3)},
					{Name: "output", Kind: uninitialized},
					{Name: "output_length", Kind: uninitialized},
					{Name: "multiplicities", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUsortBodyHint(ctx.operanders["input"], ctx.operanders["input_length"], ctx.operanders["output"], ctx.operanders["output_length"], ctx.operanders["multiplicities"])
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{
						"__usort_max_size": uint64(100),
					})
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					varAddrResolvedValueEquals("output_length", feltUint64(3))(t, ctx)
					consecutiveVarAddrResolvedValueEquals("output", []*fp.Element{
						feltUint64(1),
						feltUint64(2),
						feltUint64(3),
					})(t, ctx)
					consecutiveVarAddrResolvedValueEquals("multiplicities", []*fp.Element{
						feltUint64(1),
						feltUint64(1),
						feltUint64(1),
					})(t, ctx)
				},
			},
			{
				// sort items with one of them having multiplicity > 1
				operanders: []*hintOperander{
					{Name: "input", Kind: apRelative, Value: addr(5)},
					{Name: "input.el0", Kind: apRelative, Value: feltUint64(3)},
					{Name: "input.el1", Kind: apRelative, Value: feltUint64(2)},
					{Name: "input.el2", Kind: apRelative, Value: feltUint64(1)},
					{Name: "input.el3", Kind: apRelative, Value: feltUint64(1)},
					{Name: "input.el4", Kind: apRelative, Value: feltUint64(1)},
					{Name: "input_length", Kind: apRelative, Value: feltUint64(5)},
					{Name: "output", Kind: uninitialized},
					{Name: "output_length", Kind: uninitialized},
					{Name: "multiplicities", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUsortBodyHint(ctx.operanders["input"], ctx.operanders["input_length"], ctx.operanders["output"], ctx.operanders["output_length"], ctx.operanders["multiplicities"])
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{
						"__usort_max_size": uint64(100),
					})
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					varAddrResolvedValueEquals("output_length", feltUint64(3))(t, ctx)
					consecutiveVarAddrResolvedValueEquals("output", []*fp.Element{
						feltUint64(1),
						feltUint64(2),
						feltUint64(3),
					})(t, ctx)
					consecutiveVarAddrResolvedValueEquals("multiplicities", []*fp.Element{
						feltUint64(3),
						feltUint64(1),
						feltUint64(1),
					})(t, ctx)
				},
			},
			{
				// sort random items
				operanders: []*hintOperander{
					{Name: "input", Kind: apRelative, Value: addr(5)},
					{Name: "input.el0", Kind: apRelative, Value: &utils.Felt127},
					{Name: "input.el1", Kind: apRelative, Value: feltUint64(100)},
					{Name: "input.el2", Kind: apRelative, Value: feltUint64(1001)},
					{Name: "input.el3", Kind: apRelative, Value: &utils.Felt127},
					{Name: "input.el4", Kind: apRelative, Value: feltUint64(1)},
					{Name: "input.el5", Kind: apRelative, Value: feltUint64(100)},
					{Name: "input.el6", Kind: apRelative, Value: feltUint64(987654321)},
					{Name: "input.el7", Kind: apRelative, Value: feltUint64(1001)},
					{Name: "input.el8", Kind: apRelative, Value: &utils.Felt127},
					{Name: "input.el9", Kind: apRelative, Value: feltUint64(2)},
					{Name: "input_length", Kind: apRelative, Value: feltUint64(10)},
					{Name: "output", Kind: uninitialized},
					{Name: "output_length", Kind: uninitialized},
					{Name: "multiplicities", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUsortBodyHint(ctx.operanders["input"], ctx.operanders["input_length"], ctx.operanders["output"], ctx.operanders["output_length"], ctx.operanders["multiplicities"])
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{
						"__usort_max_size": uint64(100),
					})
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					varAddrResolvedValueEquals("output_length", feltUint64(6))(t, ctx)
					consecutiveVarAddrResolvedValueEquals("output", []*fp.Element{
						feltUint64(1),
						feltUint64(2),
						feltUint64(100),
						feltUint64(1001),
						feltUint64(987654321),
						&utils.Felt127,
					})(t, ctx)
					consecutiveVarAddrResolvedValueEquals("multiplicities", []*fp.Element{
						feltUint64(1),
						feltUint64(1),
						feltUint64(2),
						feltUint64(2),
						feltUint64(1),
						feltUint64(3),
					})(t, ctx)
				},
			},
			{
				// sort empty
				operanders: []*hintOperander{
					{Name: "input", Kind: apRelative, Value: addr(5)},
					{Name: "input_length", Kind: apRelative, Value: feltUint64(0)},
					{Name: "output", Kind: uninitialized},
					{Name: "output_length", Kind: uninitialized},
					{Name: "multiplicities", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUsortBodyHint(ctx.operanders["input"], ctx.operanders["input_length"], ctx.operanders["output"], ctx.operanders["output_length"], ctx.operanders["multiplicities"])
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{
						"__usort_max_size": uint64(100),
					})
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					varAddrResolvedValueEquals("output_length", feltUint64(0))(t, ctx)
					consecutiveVarAddrResolvedValueEquals("output", []*fp.Element{})(t, ctx)
					consecutiveVarAddrResolvedValueEquals("multiplicities", []*fp.Element{})(t, ctx)
				},
			},
		},
	})
}
