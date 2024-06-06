package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
)

func TestZeroHintMemcpy(t *testing.T) {

	runHinterTests(t, map[string][]hintTestCase{
		"MemcpyContinueCopying": {
			{
				operanders: []*hintOperander{
					{Name: "continue_copying", Kind: uninitialized},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("n", *feltInt64(1))
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newMemContinueHint(ctx.operanders["continue_copying"], false)
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					varValueInScopeEquals("n", *feltInt64(0))(t, ctx)
					varValueEquals("continue_copying", feltInt64(0))(t, ctx)
				},
			},
			{
				operanders: []*hintOperander{
					{Name: "continue_copying", Kind: uninitialized},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("n", *feltInt64(5))
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newMemContinueHint(ctx.operanders["continue_copying"], false)
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					varValueInScopeEquals("n", *feltInt64(4))(t, ctx)
					varValueEquals("continue_copying", feltInt64(1))(t, ctx)
				},
			},
		},
		"MemcpyEnterScope": {
			{
				operanders: []*hintOperander{
					{Name: "len", Kind: apRelative, Value: feltUint64(1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newMemEnterScopeHint(ctx.operanders["len"], false)
				},
				check: varValueInScopeEquals("n", *feltUint64(1)),
			},
		},
	})
}
