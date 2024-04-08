package zero

import (
	"math/big"
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
)

func TestVerifyZeroHint(t *testing.T) {
	runHinterTests(t, map[string][]hintTestCase{
		"DivModNPackedDivmodV1": {
			{
				operanders: []*hintOperander{
					{Name: "a.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "a.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "a.d2", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "b.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "b.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "b.d2", Kind: apRelative, Value: &utils.FeltZero},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newDivModNPackedDivmodV1Hint(ctx.operanders["a.d0"], ctx.operanders["b.d0"])
				},
				errCheck: errorTextContains("no solution exists (gcd(m, p) != 1)"),
			},
			{
				operanders: []*hintOperander{
					{Name: "a.d0", Kind: apRelative, Value: &utils.FeltOne},
					{Name: "a.d1", Kind: apRelative, Value: &utils.FeltOne},
					{Name: "a.d2", Kind: apRelative, Value: &utils.FeltOne},
					{Name: "b.d0", Kind: apRelative, Value: &utils.FeltOne},
					{Name: "b.d1", Kind: apRelative, Value: &utils.FeltOne},
					{Name: "b.d2", Kind: apRelative, Value: &utils.FeltOne},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newDivModNPackedDivmodV1Hint(ctx.operanders["a.d0"], ctx.operanders["b.d0"])
				},
				check: allVarValueInScopeEquals(map[string]any{
					"res":   big.NewInt(1),
					"value": big.NewInt(1),
				}),
			},
			{
				operanders: []*hintOperander{
					// values are the 3 results of split(SEC_P)
					{Name: "a.d0", Kind: apRelative, Value: feltString("77371252455336262886226991")},
					{Name: "a.d1", Kind: apRelative, Value: feltString("77371252455336267181195263")},
					{Name: "a.d2", Kind: apRelative, Value: feltString("19342813113834066795298815")},
					{Name: "b.d0", Kind: apRelative, Value: feltString("77371252455336262886226991")},
					{Name: "b.d1", Kind: apRelative, Value: feltString("77371252455336267181195263")},
					{Name: "b.d2", Kind: apRelative, Value: feltString("19342813113834066795298815")},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newDivModNPackedDivmodV1Hint(ctx.operanders["a.d0"], ctx.operanders["b.d0"])
				},
				check: allVarValueInScopeEquals(map[string]any{
					"res":   big.NewInt(1),
					"value": big.NewInt(1),
				}),
			},
			{
				operanders: []*hintOperander{
					// random values
					{Name: "a.d0", Kind: apRelative, Value: feltString("124")},
					{Name: "a.d1", Kind: apRelative, Value: feltString("668")},
					{Name: "a.d2", Kind: apRelative, Value: feltString("979")},
					{Name: "b.d0", Kind: apRelative, Value: feltString("741")},
					{Name: "b.d1", Kind: apRelative, Value: feltString("17")},
					{Name: "b.d2", Kind: apRelative, Value: feltString("670")},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newDivModNPackedDivmodV1Hint(ctx.operanders["a.d0"], ctx.operanders["b.d0"])
				},
				check: allVarValueInScopeEquals(map[string]any{
					"res":   bigIntString("62733347149736974538461843763852691885676254208529184638286052021917647089374", 10),
					"value": bigIntString("62733347149736974538461843763852691885676254208529184638286052021917647089374", 10),
				}),
			},
			{
				operanders: []*hintOperander{
					// SECP N
					{Name: "a.d0", Kind: apRelative, Value: feltString("115792089237316195423570985008687907852837564279074904382605163141518161494337")},
					{Name: "a.d1", Kind: apRelative, Value: feltString("115792089237316195423570985008687907852837564279074904382605163141518161494337")},
					{Name: "a.d2", Kind: apRelative, Value: feltString("115792089237316195423570985008687907852837564279074904382605163141518161494337")},
					// random values
					{Name: "b.d0", Kind: apRelative, Value: feltString("2")},
					{Name: "b.d1", Kind: apRelative, Value: feltString("3")},
					{Name: "b.d2", Kind: apRelative, Value: feltString("4")},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newDivModNPackedDivmodV1Hint(ctx.operanders["a.d0"], ctx.operanders["b.d0"])
				},
				check: allVarValueInScopeEquals(map[string]any{
					"res":   bigIntString("46265455505090914726589047408527334283412885095607834804602783449640584795651", 10),
					"value": bigIntString("46265455505090914726589047408527334283412885095607834804602783449640584795651", 10),
				}),
			},
		},
	})
}
