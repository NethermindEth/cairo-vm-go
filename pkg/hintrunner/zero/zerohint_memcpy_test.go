package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
)

func TestMemcPyHint(t *testing.T) {
	runHinterTests(t, map[string][]hintTestCase{
		"MemcPyContinueCopying": {
			{
				operanders: []*hintOperander{
					{Name: "continue_copying", Kind: uninitialized},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
					err := ctx.ScopeManager.AssignVariables(map[string]any{
						"n": &utils.FeltOne,
					})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newMemcpyContinueCopyingHint(ctx.operanders["continue_copying"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					allVarValueInScopeEquals(map[string]any{"n": feltInt64(0)})(t, ctx)
					varValueEquals("continue_copying", feltInt64(0))(t, ctx)
				},
			},
			{
				operanders: []*hintOperander{
					{Name: "continue_copying", Kind: uninitialized},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
					err := ctx.ScopeManager.AssignVariables(map[string]any{
						"n": feltString("5"),
					})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newMemcpyContinueCopyingHint(ctx.operanders["continue_copying"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					allVarValueInScopeEquals(map[string]any{"n": feltInt64(4)})(t, ctx)
					varValueEquals("continue_copying", feltInt64(1))(t, ctx)
				},
			},
		},
	})
}
