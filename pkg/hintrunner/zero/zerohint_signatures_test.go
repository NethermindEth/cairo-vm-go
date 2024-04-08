package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/parsers/starknet"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"github.com/stretchr/testify/require"
)

func TestSignatures(t *testing.T) {
	runHinterTests(t, map[string][]hintTestCase{
		"VerifyECDSASignature": {
			{
				operanders: []*hintOperander{
					{Name: "ecdsaPtr", Kind: reference, Value: addrBuiltin(starknet.ECDSA, 0)},
					{Name: "signature_r", Kind: apRelative, Value: feltString("3086480810278599376317923499561306189851900463386393948998357832163236918254")},
					{Name: "signature_s", Kind: apRelative, Value: feltString("598673427589502599949712887611119751108407514580626464031881322743364689811")},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					ecdsaPtr := ctx.operanders["ecdsaPtr"].(*hinter.DoubleDeref).Deref
					return newVerifyECDSASignatureHinter(ecdsaPtr, ctx.operanders["signature_r"], ctx.operanders["signature_s"])
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
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newGetPointFromXHinter(ctx.operanders["xCube.d0"], ctx.operanders["v"])
				},
				check: varValueInScopeEquals("value", bigIntString("64828261740814840065360381756190772627110652128289340260788836867053167272156")),
			},
			// if v % 2 != y % 2:
			{
				operanders: []*hintOperander{
					{Name: "xCube.d0", Kind: apRelative, Value: &utils.FeltOne},
					{Name: "xCube.d1", Kind: apRelative, Value: &utils.FeltOne},
					{Name: "xCube.d2", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "v", Kind: apRelative, Value: &utils.FeltZero},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newGetPointFromXHinter(ctx.operanders["xCube.d0"], ctx.operanders["v"])
				},
				check: varValueInScopeEquals("value", bigIntString("3754707778961574900176639079436749683878498834289427635045629810524611907876")),
			},
			// values are 2**86 BASE
			{
				operanders: []*hintOperander{
					{Name: "xCube.d0", Kind: apRelative, Value: feltString("77371252455336267181195264")},
					{Name: "xCube.d1", Kind: apRelative, Value: feltString("77371252455336267181195264")},
					{Name: "xCube.d2", Kind: apRelative, Value: feltString("77371252455336267181195264")},
					{Name: "v", Kind: apRelative, Value: feltString("77371252455336267181195264")},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newGetPointFromXHinter(ctx.operanders["xCube.d0"], ctx.operanders["v"])
				},
				check: varValueInScopeEquals("value", bigIntString("64330220386510520462271671435567806262107470356169873352512014089172394266548")),
			},
		},
		"DivModNSafeDivHint": {
			{
				// zero quotient
				operanders: []*hintOperander{},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
					ctx.ScopeManager.AssignVariable("res", bigIntString("0"))
					ctx.ScopeManager.AssignVariable("a", bigIntString("0"))
					ctx.ScopeManager.AssignVariable("b", bigIntString("0"))
					ctx.ScopeManager.AssignVariable("N", bigIntString("1"))
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newDivModSafeDivHinter()
				},
				check: varListInScopeEquals(map[string]any{
					"value": bigIntString("0"),
					"k":     bigIntString("0"),
				}),
			},
			{
				// negative quotient
				operanders: []*hintOperander{},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
					err := ctx.ScopeManager.AssignVariable("res", bigIntString("1"))
					if err != nil {
						t.Fatal(err)
					}
					err = ctx.ScopeManager.AssignVariable("a", bigIntString("2"))
					if err != nil {
						t.Fatal(err)
					}
					err = ctx.ScopeManager.AssignVariable("b", bigIntString("1"))
					if err != nil {
						t.Fatal(err)
					}
					err = ctx.ScopeManager.AssignVariable("N", bigIntString("1"))
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newDivModSafeDivHinter()
				},
				check: varListInScopeEquals(map[string]any{
					"value": bigIntString("-1"),
					"k":     bigIntString("-1"),
				})},
			{
				// positive quotient
				operanders: []*hintOperander{},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
					ctx.ScopeManager.AssignVariable("res", bigIntString("10"))
					ctx.ScopeManager.AssignVariable("a", bigIntString("20"))
					ctx.ScopeManager.AssignVariable("b", bigIntString("30"))
					ctx.ScopeManager.AssignVariable("N", bigIntString("2"))
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newDivModSafeDivHinter()
				},
				check: varListInScopeEquals(map[string]any{
					"value": bigIntString("140"),
					"k":     bigIntString("140"),
				}),
			},
		},
	})
}
