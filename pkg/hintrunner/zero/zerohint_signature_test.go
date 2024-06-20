package zero

import (
	"math/big"
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/parsers/starknet"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"github.com/stretchr/testify/require"
)

func TestVerifyZeroHint(t *testing.T) {
	runHinterTests(t, map[string][]hintTestCase{
		"VerifyZero": {
			{
				operanders: []*hintOperander{
					{Name: "val.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "val.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "val.d2", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "q", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newVerifyZeroHint(ctx.operanders["val.d0"], ctx.operanders["q"])
				},
				check: varValueEquals("q", feltInt64(0)),
			},
			{
				operanders: []*hintOperander{
					{Name: "val.d0", Kind: apRelative, Value: &utils.FeltOne},
					{Name: "val.d1", Kind: apRelative, Value: &utils.FeltOne},
					{Name: "val.d2", Kind: apRelative, Value: &utils.FeltOne},
					{Name: "q", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newVerifyZeroHint(ctx.operanders["val.d0"], ctx.operanders["q"])
				},
				errCheck: errorTextContains("verify_zero: Invalid input (1, 1, 1)"),
			},
			{
				operanders: []*hintOperander{
					// values are the 3 results of split(SEC_P)
					{Name: "val.d0", Kind: apRelative, Value: feltString("77371252455336262886226991")},
					{Name: "val.d1", Kind: apRelative, Value: feltString("77371252455336267181195263")},
					{Name: "val.d2", Kind: apRelative, Value: feltString("19342813113834066795298815")},
					{Name: "q", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newVerifyZeroHint(ctx.operanders["val.d0"], ctx.operanders["q"])
				},
				check: varValueEquals("q", feltInt64(1)),
			},
		},
		"VerifyECDSASignature": {
			{
				operanders: []*hintOperander{
					{Name: "ecdsaPtr", Kind: reference, Value: addrBuiltin(starknet.ECDSA, 0)},
					{Name: "signature_r", Kind: apRelative, Value: feltString("3086480810278599376317923499561306189851900463386393948998357832163236918254")},
					{Name: "signature_s", Kind: apRelative, Value: feltString("598673427589502599949712887611119751108407514580626464031881322743364689811")},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					ecdsaPtr := ctx.operanders["ecdsaPtr"].(*hinter.DoubleDeref).Deref
					return newVerifyECDSASignatureHint(ecdsaPtr, ctx.operanders["signature_r"], ctx.operanders["signature_s"])
				},
				errCheck: func(t *testing.T, ctx *hintTestContext, err error) {
					require.NoError(t, err)
				},
			},
		},
		"GetPointFromX": {
			{
				//> if v % 2 == y % 2
				operanders: []*hintOperander{
					{Name: "xCube.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "xCube.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "xCube.d2", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "v", Kind: apRelative, Value: &utils.FeltZero},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newGetPointFromXHint(ctx.operanders["xCube.d0"], ctx.operanders["v"])
				},
				check: varValueInScopeEquals("value", bigIntString("64828261740814840065360381756190772627110652128289340260788836867053167272156", 10)),
			},
			// if v % 2 != y % 2:
			{
				operanders: []*hintOperander{
					{Name: "xCube.d0", Kind: apRelative, Value: &utils.FeltOne},
					{Name: "xCube.d1", Kind: apRelative, Value: &utils.FeltOne},
					{Name: "xCube.d2", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "v", Kind: apRelative, Value: &utils.FeltZero},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newGetPointFromXHint(ctx.operanders["xCube.d0"], ctx.operanders["v"])
				},
				check: varValueInScopeEquals("value", bigIntString("3754707778961574900176639079436749683878498834289427635045629810524611907876", 10)),
			},
			// values are 2**86 BASE
			{
				operanders: []*hintOperander{
					{Name: "xCube.d0", Kind: apRelative, Value: feltString("77371252455336267181195264")},
					{Name: "xCube.d1", Kind: apRelative, Value: feltString("77371252455336267181195264")},
					{Name: "xCube.d2", Kind: apRelative, Value: feltString("77371252455336267181195264")},
					{Name: "v", Kind: apRelative, Value: feltString("77371252455336267181195264")},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newGetPointFromXHint(ctx.operanders["xCube.d0"], ctx.operanders["v"])
				},
				check: varValueInScopeEquals("value", bigIntString("64330220386510520462271671435567806262107470356169873352512014089172394266548", 10)),
			},
		},
		"ImportSecp256R1P": {
			{
				operanders: []*hintOperander{},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newImportSecp256R1PHint()
				},
				check: varValueInScopeEquals("SECP_P", bigIntString("115792089210356248762697446949407573530086143415290314195533631308867097853951", 10)),
			},
		},
		"DivModNSafeDiv": {
			{
				// zero quotient
				operanders: []*hintOperander{},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariables(map[string]any{
						"res": bigIntString("0", 10),
						"a":   bigIntString("0", 10),
						"b":   bigIntString("0", 10),
						"N":   bigIntString("1", 10),
					})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newDivModSafeDivHint()
				},
				check: varListInScopeEquals(map[string]any{
					"value": bigIntString("0", 10),
				}),
			},
			{
				// negative quotient
				operanders: []*hintOperander{},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariables(map[string]any{
						"res": bigIntString("1", 10),
						"a":   bigIntString("2", 10),
						"b":   bigIntString("1", 10),
						"N":   bigIntString("1", 10),
					})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newDivModSafeDivHint()
				},
				check: varListInScopeEquals(map[string]any{
					"value": bigIntString("-1", 10),
				}),
			},
			{
				// positive quotient
				operanders: []*hintOperander{},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariables(map[string]any{
						"res": bigIntString("10", 10),
						"a":   bigIntString("20", 10),
						"b":   bigIntString("30", 10),
						"N":   bigIntString("2", 10),
					})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newDivModSafeDivHint()
				},
				check: varListInScopeEquals(map[string]any{
					"value": bigIntString("140", 10),
				}),
			},
		},
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
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newDivModNPackedDivmodV1Hint(ctx.operanders["a.d0"], ctx.operanders["b.d0"])
				},
				check: allVarValueInScopeEquals(map[string]any{
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
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newDivModNPackedDivmodV1Hint(ctx.operanders["a.d0"], ctx.operanders["b.d0"])
				},
				check: allVarValueInScopeEquals(map[string]any{
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
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newDivModNPackedDivmodV1Hint(ctx.operanders["a.d0"], ctx.operanders["b.d0"])
				},
				check: allVarValueInScopeEquals(map[string]any{
					"value": bigIntString("62733347149736974538461843763852691885676254208529184638286052021917647089374", 10),
				}),
			},
		},
		"IsZeroPack": {
			{
				operanders: []*hintOperander{
					{Name: "x.d0", Kind: apRelative, Value: feltString("42")},
					{Name: "x.d1", Kind: apRelative, Value: feltString("0")},
					{Name: "x.d2", Kind: apRelative, Value: feltString("0")},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newIsZeroPackHint(ctx.operanders["x.d0"])
				},
				check: varValueInScopeEquals("x", big.NewInt(42)),
			},
			{
				operanders: []*hintOperander{
					{Name: "x.d0", Kind: apRelative, Value: feltString("100")},
					{Name: "x.d1", Kind: apRelative, Value: feltString("99")},
					{Name: "x.d2", Kind: apRelative, Value: feltString("88")},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newIsZeroPackHint(ctx.operanders["x.d0"])
				},
				check: varValueInScopeEquals("x", bigIntString("526795342172649295060681798242672774947232024188944484", 10)),
			},
			{
				operanders: []*hintOperander{
					{Name: "x.d0", Kind: apRelative, Value: feltString("77371252455336262886226991")},
					{Name: "x.d1", Kind: apRelative, Value: feltString("77371252455336267181195263")},
					{Name: "x.d2", Kind: apRelative, Value: feltString("19342813113834066795298815")},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newIsZeroPackHint(ctx.operanders["x.d0"])
				},
				check: varValueInScopeEquals("x", big.NewInt(0)),
			},
			{
				operanders: []*hintOperander{
					{Name: "x.d0", Kind: apRelative, Value: feltString("3618502788666131213697322783095070105623107215331596699973092056135872020481")},
					{Name: "x.d1", Kind: apRelative, Value: feltString("0")},
					{Name: "x.d2", Kind: apRelative, Value: feltString("0")},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newIsZeroPackHint(ctx.operanders["x.d0"])
				},
				check: varValueInScopeEquals("x", big.NewInt(0)),
			},
		},
		"IsZeroDivMod": {
			{
				operanders: []*hintOperander{},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("x", bigIntString("1", 10))
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newIsZeroDivModHint()
				},
				check: varValueInScopeEquals("value", bigIntString("1", 10)),
			},
			{
				operanders: []*hintOperander{},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("x", bigIntString("115792089237316195423570985008687907853269984665640564039457584007908834671664", 10))
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newIsZeroDivModHint()
				},
				check: varValueInScopeEquals("value", bigIntString("1", 10)),
			},
			{
				operanders: []*hintOperander{},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("x", bigIntString("57662894568246526582652685623", 10))
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newIsZeroDivModHint()
				},
				check: varValueInScopeEquals("value", bigIntString("77726902514058095204421112730928006705863972015508190238152451720695936255632", 10)),
			},
			{
				operanders: []*hintOperander{},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("x", bigIntString("28948022309329048855892746252171976963317496166410141009864396001977208667916", 10))
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newIsZeroDivModHint()
				},
				check: varValueInScopeEquals("value", bigIntString("4", 10)),
			},
		},
	})
}
