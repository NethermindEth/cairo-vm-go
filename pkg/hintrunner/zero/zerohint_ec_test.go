package zero

import (
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"math/big"
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
)

func TestZeroHintEc(t *testing.T) {
	runHinterTests(t, map[string][]hintTestCase{
		"EcNegate": {
			{
				operanders: []*hintOperander{
					{Name: "x.d0", Kind: apRelative, Value: feltString("0xe28d959f2815b16f81798")},
					{Name: "x.d1", Kind: apRelative, Value: feltString("0xa573a1c2c1c0a6ff36cb7")},
					{Name: "x.d2", Kind: apRelative, Value: feltString("0x79be667ef9dcbbac55a06")},
					{Name: "y.d0", Kind: apRelative, Value: feltString("0x554199c47d08ffb10d4b8")},
					{Name: "y.d1", Kind: apRelative, Value: feltString("0x2ff0384422a3f45ed1229a")},
					{Name: "y.d2", Kind: apRelative, Value: feltString("0x483ada7726a3c4655da4f")},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcNegateHint(ctx.operanders["x.d0"])
				},
				check: varValueInScopeEquals("value", bigIntString("83121579216557378445487899878180864668798711284981320763518679672151497189239", 10)),
			},
			{
				operanders: []*hintOperander{
					{Name: "x.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "x.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "x.d2", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "y.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "y.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "y.d2", Kind: apRelative, Value: &utils.FeltZero},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcNegateHint(ctx.operanders["x.d0"])
				},
				check: varValueInScopeEquals("value", bigIntString("0", 10)),
			},
			{
				operanders: []*hintOperander{
					{Name: "x.d0", Kind: apRelative, Value: feltString("3618502788666127798953978732740734578953660990361066340291730267696802036752")},
					{Name: "x.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "x.d2", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "y.d0", Kind: apRelative, Value: feltString("3618502788666127798953978732740734578953660990361066340291730267696802036752")},
					{Name: "y.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "y.d2", Kind: apRelative, Value: &utils.FeltZero},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcNegateHint(ctx.operanders["x.d0"])
				},
				check: varValueInScopeEquals("value", bigIntString("3414743344050354335526669446224970530359681361788439069983729", 10)),
			},
			{
				operanders: []*hintOperander{
					{Name: "x.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "x.d1", Kind: apRelative, Value: feltString("3618502788666127798953978732740734578953660990361066340291730267696802036752")},
					{Name: "x.d2", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "y.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "y.d1", Kind: apRelative, Value: feltString("3618502788666127798953978732740734578953660990361066340291730267696802036752")},
					{Name: "y.d2", Kind: apRelative, Value: &utils.FeltZero},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcNegateHint(ctx.operanders["x.d0"])
				},
				check: varValueInScopeEquals("value", bigIntString("332307077013822705460080369276551168", 10)),
			},
			{
				operanders: []*hintOperander{
					{Name: "x.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "x.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "x.d2", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "y.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "y.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "y.d2", Kind: apRelative, Value: feltString("3618502788666127798953978732740734578953660990361066340291730267696802036752")},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcNegateHint(ctx.operanders["x.d0"])
				},
				check: varValueInScopeEquals("value", bigIntString("25711014748331348032841660844170547741622139443892033895268352", 10)),
			},
			{
				operanders: []*hintOperander{
					{Name: "x.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "x.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "x.d2", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "y.d0", Kind: apRelative, Value: feltString("3618502788666127798953978732740734578953660990361066340291730267696802036752")},
					{Name: "y.d1", Kind: apRelative, Value: feltString("3618502788666127798953978732740734578953660990361066340291730267696802036752")},
					{Name: "y.d2", Kind: apRelative, Value: &utils.FeltZero},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcNegateHint(ctx.operanders["x.d0"])
				},
				check: varValueInScopeEquals("value", bigIntString("3414743344050354335526669778532047544182386821868808346534897", 10)),
			},
			{
				operanders: []*hintOperander{
					{Name: "x.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "x.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "x.d2", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "y.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "y.d1", Kind: apRelative, Value: feltString("3618502788666127798953978732740734578953660990361066340291730267696802036752")},
					{Name: "y.d2", Kind: apRelative, Value: feltString("3618502788666127798953978732740734578953660990361066340291730267696802036752")},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcNegateHint(ctx.operanders["x.d0"])
				},
				check: varValueInScopeEquals("value", bigIntString("25711014748331348032841661176477624755444844903972403171819520", 10)),
			},
			{
				operanders: []*hintOperander{
					{Name: "x.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "x.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "x.d2", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "y.d0", Kind: apRelative, Value: feltString("3618502788666127798953978732740734578953660990361066340291730267696802036752")},
					{Name: "y.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "y.d2", Kind: apRelative, Value: feltString("3618502788666127798953978732740734578953660990361066340291730267696802036752")},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcNegateHint(ctx.operanders["x.d0"])
				},
				check: varValueInScopeEquals("value", bigIntString("29125758092381702368368330290395518271981820805680472965252081", 10)),
			},
			{
				operanders: []*hintOperander{
					{Name: "x.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "x.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "x.d2", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "y.d0", Kind: apRelative, Value: feltString("3618502788666127798953978732740734578953660990361066340291730267696802036752")},
					{Name: "y.d1", Kind: apRelative, Value: feltString("3618502788666127798953978732740734578953660990361066340291730267696802036752")},
					{Name: "y.d2", Kind: apRelative, Value: feltString("3618502788666127798953978732740734578953660990361066340291730267696802036752")},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcNegateHint(ctx.operanders["x.d0"])
				},
				check: varValueInScopeEquals("value", bigIntString("29125758092381702368368330622702595285804526265760842241803249", 10)),
			},
			{
				operanders: []*hintOperander{
					{Name: "x.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "x.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "x.d2", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "y.d0", Kind: apRelative, Value: feltString("10")},
					{Name: "y.d1", Kind: apRelative, Value: feltString("100")},
					{Name: "y.d2", Kind: apRelative, Value: feltString("10001")},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcNegateHint(ctx.operanders["x.d0"])
				},
				check: varValueInScopeEquals("value", bigIntString("115792089237316195423511115915312127562362008772591693155831694873530722155557", 10)),
			},
		},
		"NondetBigint3V1": {
			{
				operanders: []*hintOperander{
					{Name: "res", Kind: uninitialized},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})

					value, ok := new(big.Int).SetString("7618502788666127798953978732740734578953660990361066340291730267696802036752", 10)
					if !ok {
						t.Errorf("Error creating big.Int")
					}

					err := ctx.ScopeManager.AssignVariable("value", value)
					if err != nil {
						t.Errorf("Error assigning variable value in scope")
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newNondetBigint3V1Hint(ctx.operanders["res"])
				},
				check: consecutiveVarValueEquals("res", []*fp.Element{feltString("72082201994522260246887440"), feltString("9036023809832564006525928"), feltString("1272654087330362946288670")}),
			},
			{
				operanders: []*hintOperander{
					{Name: "res", Kind: uninitialized},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})

					value, ok := new(big.Int).SetString("3618502788666127798953978732740734578953660990361066340291730267696802036752", 10)
					if !ok {
						t.Errorf("Error creating big.Int")
					}

					err := ctx.ScopeManager.AssignVariable("value", value)
					if err != nil {
						t.Errorf("Error assigning variable value in scope")
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newNondetBigint3V1Hint(ctx.operanders["res"])
				},
				check: consecutiveVarValueEquals("res", []*fp.Element{feltString("77371252455336262886226960"), feltString("77371252455336267181195263"), feltString("604462909807314034753535")}),
			},
			{
				operanders: []*hintOperander{
					{Name: "res", Kind: uninitialized},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})

					value := big.NewInt(123456)

					err := ctx.ScopeManager.AssignVariable("value", value)
					if err != nil {
						t.Errorf("Error assigning variable value in scope")
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newNondetBigint3V1Hint(ctx.operanders["res"])
				},
				check: consecutiveVarValueEquals("res", []*fp.Element{feltString("123456"), &utils.FeltZero, &utils.FeltZero}),
			},
			{
				operanders: []*hintOperander{
					{Name: "res", Kind: uninitialized},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})

					value := big.NewInt(-123456)

					err := ctx.ScopeManager.AssignVariable("value", value)
					if err != nil {
						t.Errorf("Error assigning variable value in scope")
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newNondetBigint3V1Hint(ctx.operanders["res"])
				},
				errCheck: errorTextContains("value != 0"),
			},
			{
				operanders: []*hintOperander{
					{Name: "res", Kind: uninitialized},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})

					value, ok := new(big.Int).SetString("77371252455336267181195263", 10)
					if !ok {
						t.Errorf("Error creating big.Int")
					}

					err := ctx.ScopeManager.AssignVariable("value", value)
					if err != nil {
						t.Errorf("Error assigning variable value in scope")
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newNondetBigint3V1Hint(ctx.operanders["res"])
				},
				check: consecutiveVarValueEquals("res", []*fp.Element{feltString("77371252455336267181195263"), &utils.FeltZero, &utils.FeltZero}),
			},
			{
				operanders: []*hintOperander{
					{Name: "res", Kind: uninitialized},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})

					value, ok := new(big.Int).SetString("77371252455336267181195264", 10)
					if !ok {
						t.Errorf("Error creating big.Int")
					}

					err := ctx.ScopeManager.AssignVariable("value", value)
					if err != nil {
						t.Errorf("Error assigning variable value in scope")
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newNondetBigint3V1Hint(ctx.operanders["res"])
				},
				check: consecutiveVarValueEquals("res", []*fp.Element{&utils.FeltZero, &utils.FeltOne, &utils.FeltZero}),
			},
			{
				operanders: []*hintOperander{
					{Name: "res", Kind: uninitialized},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})

					value, ok := new(big.Int).SetString("77371252455336267181195265", 10)
					if !ok {
						t.Errorf("Error creating big.Int")
					}

					err := ctx.ScopeManager.AssignVariable("value", value)
					if err != nil {
						t.Errorf("Error assigning variable value in scope")
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newNondetBigint3V1Hint(ctx.operanders["res"])
				},
				check: consecutiveVarValueEquals("res", []*fp.Element{&utils.FeltOne, &utils.FeltOne, &utils.FeltZero}),
			},
			{
				operanders: []*hintOperander{
					{Name: "res", Kind: uninitialized},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})

					value := big.NewInt(0)

					err := ctx.ScopeManager.AssignVariable("value", value)
					if err != nil {
						t.Errorf("Error assigning variable value in scope")
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newNondetBigint3V1Hint(ctx.operanders["res"])
				},
				check: consecutiveVarValueEquals("res", []*fp.Element{&utils.FeltZero, &utils.FeltZero, &utils.FeltZero}),
			},
			{
				operanders: []*hintOperander{
					{Name: "res", Kind: uninitialized},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})

					value, ok := new(big.Int).SetString("154742504910672534362390526", 10)
					if !ok {
						t.Errorf("Error creating big.Int")
					}

					err := ctx.ScopeManager.AssignVariable("value", value)
					if err != nil {
						t.Errorf("Error assigning variable value in scope")
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newNondetBigint3V1Hint(ctx.operanders["res"])
				},
				check: consecutiveVarValueEquals("res", []*fp.Element{feltString("77371252455336267181195262"), &utils.FeltOne, &utils.FeltZero}),
			},
			{
				operanders: []*hintOperander{
					{Name: "res", Kind: uninitialized},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})

					value, ok := new(big.Int).SetString("154742504910672534362390528", 10)
					if !ok {
						t.Errorf("Error creating big.Int")
					}

					err := ctx.ScopeManager.AssignVariable("value", value)
					if err != nil {
						t.Errorf("Error assigning variable value in scope")
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newNondetBigint3V1Hint(ctx.operanders["res"])
				},
				check: consecutiveVarValueEquals("res", []*fp.Element{&utils.FeltZero, feltString("2"), &utils.FeltZero}),
			},
			{
				operanders: []*hintOperander{
					{Name: "res", Kind: uninitialized},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})

					value, ok := new(big.Int).SetString("154742504910672534362390530", 10)
					if !ok {
						t.Errorf("Error creating big.Int")
					}

					err := ctx.ScopeManager.AssignVariable("value", value)
					if err != nil {
						t.Errorf("Error assigning variable value in scope")
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newNondetBigint3V1Hint(ctx.operanders["res"])
				},
				check: consecutiveVarValueEquals("res", []*fp.Element{feltString("2"), feltString("2"), &utils.FeltZero}),
			},
		},
	})
}
