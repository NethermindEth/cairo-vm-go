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
					{Name: "output", Kind: uninitialized},
					{Name: "output_length", Kind: uninitialized},
					{Name: "multiplicities", Kind: uninitialized},
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
					{Name: "input", Kind: apRelative, Value: addr(5)},
					{Name: "input.el0", Kind: apRelative, Value: feltUint64(3)},
					{Name: "input.el2", Kind: apRelative, Value: feltUint64(2)},
					{Name: "input.el3", Kind: apRelative, Value: feltUint64(1)},
					{Name: "input_length", Kind: apRelative, Value: feltUint64(3)},
					{Name: "output", Kind: uninitialized},
					{Name: "output_length", Kind: uninitialized},
					{Name: "multiplicities", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUsortBodyHint(ctx.operanders["input"], ctx.operanders["input_length"], ctx.operanders["output"], ctx.operanders["output_length"], ctx.operanders["multiplicities"])
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					hinter.InitializeScopeManager(ctx, map[string]any{})
					ctx.ScopeManager.AssignVariable("__usort_max_size", new(big.Int).SetUint64(100))
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
				operanders: []*hintOperander{
					{Name: "input", Kind: apRelative, Value: addr(5)},
					{Name: "input.el0", Kind: apRelative, Value: feltUint64(3)},
					{Name: "input.el2", Kind: apRelative, Value: feltUint64(2)},
					{Name: "input.el3", Kind: apRelative, Value: feltUint64(1)},
					{Name: "input.el3", Kind: apRelative, Value: feltUint64(1)},
					{Name: "input.el3", Kind: apRelative, Value: feltUint64(1)},
					{Name: "input_length", Kind: apRelative, Value: feltUint64(5)},
					{Name: "output", Kind: uninitialized},
					{Name: "output_length", Kind: uninitialized},
					{Name: "multiplicities", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUsortBodyHint(ctx.operanders["input"], ctx.operanders["input_length"], ctx.operanders["output"], ctx.operanders["output_length"], ctx.operanders["multiplicities"])
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					hinter.InitializeScopeManager(ctx, map[string]any{})
					ctx.ScopeManager.AssignVariable("__usort_max_size", new(big.Int).SetUint64(100))
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
		},
	})
}
