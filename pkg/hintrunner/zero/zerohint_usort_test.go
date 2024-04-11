package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
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
		"UsortVerifyMultiplicityAssert": {
			{
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{
						"positions": []*fp.Element{
							feltUint64(1),
						},
					})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUsortVerifyMultiplicityAssertHinter()
				},
				errCheck: errorTextContains("assertion `len(positions) == 0` failed"),
			},
			{
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{
						"positions": []*fp.Element{},
					})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUsortVerifyMultiplicityAssertHinter()
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					print("AAAAAA")
					positions_interface, err := ctx.runnerContext.ScopeManager.GetVariableValue("positions")
					if err != nil {
						t.Fatal(err)
					}

					positions, ok := positions_interface.([]*fp.Element)
					if !ok {
						t.Fatal("casting positions into an array failed")
					}

					if len(positions) != 0 {
						t.Fatal("assertion `len(positions) == 0` failed")
					}

				},
			},
		},
	})
}
