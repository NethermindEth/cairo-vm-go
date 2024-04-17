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
					return newUsortEnterScopeHint()
				},
				check: varValueInScopeEquals("__usort_max_size", feltUint64(1)),
			},
		},
		"UsortVerify": {
			{
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					hinter.InitializeScopeManager(ctx, map[string]any{
						"positions_dict": map[uint64][]*fp.Element{
							0: {feltUint64(1), feltUint64(2), feltUint64(3)},
						},
					})
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

					require.Equal(t, []*fp.Element{feltUint64(3), feltUint64(2), feltUint64(1)}, positions)

					lastPos, err := ctx.runnerContext.ScopeManager.GetVariableValue("last_pos")
					require.NoError(t, err)

					require.Equal(t, 0, lastPos)
				},
				// check: allVarValueInScopeEquals(map[string]any{
				// 	"last_pos":  0,
				// 	"positions": []*fp.Element{feltUint64(3), feltUint64(2), feltUint64(1)},
				// }),
				// check: allVarValueInScopeEquals(map[string]any{
				// 	"last_pos":  0,
				// 	"positions": [...]*fp.Element{feltUint64(3), feltUint64(2), feltUint64(1)},
				// }),
			},
		},
	})
}
