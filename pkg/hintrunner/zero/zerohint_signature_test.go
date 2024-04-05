package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
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
				errCheck: errorTextContains("verify_zero: Invalid input (1, 1, 1)."),
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
	})
}
