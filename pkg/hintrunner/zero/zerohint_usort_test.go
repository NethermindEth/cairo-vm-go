package zero

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func TestZeroHintUsort(t *testing.T) {

	runHinterTests(t, map[string][]hintTestCase{
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
					hinter.InitializeScopeManager(ctx, map[string]any{})
					ctx.ScopeManager.AssignVariable("__usort_max_size", new(big.Int).SetUint64(10))
				},
				errCheck: errorTextContains(fmt.Sprintf("usort() can only be used with input_len<=%d.\n Got: input_len=%d", 10, 20)),
			},
			{
				// sort items with multiplicity of 1
				operanders: []*hintOperander{
					{Name: "input", Kind: apRelative, Value: addr(5)},
					{Name: "input.el0", Kind: apRelative, Value: feltUint64(2)},
					{Name: "input.el2", Kind: apRelative, Value: feltUint64(3)},
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
				// sort items with one of them having multiplicity > 1
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
					hinter.InitializeScopeManager(ctx, map[string]any{})
					ctx.ScopeManager.AssignVariable("__usort_max_size", new(big.Int).SetUint64(100))
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
					hinter.InitializeScopeManager(ctx, map[string]any{})
					ctx.ScopeManager.AssignVariable("__usort_max_size", new(big.Int).SetUint64(100))
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
