package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
)

func TestZeroHintMemcpy(t *testing.T) {

	runHinterTests(t, map[string][]hintTestCase{
		"MemcpyEnterScope": {
			{
				operanders: []*hintOperander{
					{Name: "len", Kind: apRelative, Value: feltUint64(1)},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					hinter.InitializeScopeManager(ctx, map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newMemcpyEnterScopeHinter(ctx.operanders["len"])
				},
				check: varValueInScopeEquals("n", feltUint64(1)),
			},
		},
	})
}
