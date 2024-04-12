package zero

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func TestZeroHintUsort(t *testing.T) {

	runHinterTests(t, map[string][]hintTestCase{
		"UsortBody": {
			{
				operanders: []*hintOperander{
					{Name: "input", Kind: apRelative, Value: addr(7)},
					{Name: "input_length", Kind: apRelative, Value: feltUint64(20)},
					{Name: "output", Kind: apRelative, Value: uninitialized},
					{Name: "output_length", Kind: apRelative, Value: uninitialized},
					{Name: "multiplicities", Kind: apRelative, Value: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUsortBodyHint(ctx.operanders["input"], ctx.operanders["input_length"], ctx.operanders["output"], ctx.operanders["output_length"], ctx.operanders["multiplicites"])
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					hinter.InitializeScopeManager(ctx, map[string]any{})
					ctx.ScopeManager.AssignVariable("__usort_max_size", new(big.Int).SetUint64(10))
				},
				errCheck: errorTextContains(fmt.Sprintf("usort() can only be used with input_len<=%d.\n Got: input_len=%d", 10, 20)),
			},
			{
				operanders: []*hintOperander{
					{Name: "input", Kind: apRelative, Value: addr(7)},
					{Name: "input_length", Kind: apRelative, Value: feltUint64(20)},
					{Name: "output", Kind: apRelative, Value: addr(7)},
					{Name: "output_length", Kind: apRelative, Value: feltUint64(0)},
					{Name: "multiplicities", Kind: apRelative, Value: feltUint64(0)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUsortBodyHint(ctx.operanders["input"], ctx.operanders["input_length"], ctx.operanders["output"], ctx.operanders["output_length"], ctx.operanders["multiplicites"])
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					hinter.InitializeScopeManager(ctx, map[string]any{})
					ctx.ScopeManager.AssignVariable("__usort_max_size", new(big.Int).SetUint64(10))
				},
				check: consecutiveVarAddrResolvedValueEquals("output", []*fp.Element{
					feltUint64(0),
					feltUint64(1),
					feltUint64(2),
					feltUint64(3),
				}),
			},
		},
	})
}
