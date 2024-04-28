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
						"n": uint64(1),
					})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newMemcpyContinueCopyingHint(ctx.operanders["continue_copying"])
				},
				check: varValueEquals("continue_copying", &utils.FeltZero),
			},
			{
				operanders: []*hintOperander{
					{Name: "continue_copying", Kind: uninitialized},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
					err := ctx.ScopeManager.AssignVariables(map[string]any{
						"n": uint64(5),
					})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newMemcpyContinueCopyingHint(ctx.operanders["continue_copying"])
				},
				check: varValueEquals("continue_copying", &utils.FeltOne),
			},
		},
	})
}
