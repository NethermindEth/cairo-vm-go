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
				errCheck: errorTextContains("num != 0"),
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
		"FastEcAddAssignNewY": {
			{
				operanders: []*hintOperander{},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})

					slopeBig := big.NewInt(100)
					x0Big := big.NewInt(20)
					new_xBig := big.NewInt(10)
					y0Big := big.NewInt(10)

					err := ctx.ScopeManager.AssignVariable("slope", slopeBig)
					if err != nil {
						t.Errorf("Error assigning variable value in scope")
					}
					err = ctx.ScopeManager.AssignVariable("x0", x0Big)
					if err != nil {
						t.Errorf("Error assigning variable value in scope")
					}
					err = ctx.ScopeManager.AssignVariable("new_x", new_xBig)
					if err != nil {
						t.Errorf("Error assigning variable value in scope")
					}
					err = ctx.ScopeManager.AssignVariable("y0", y0Big)
					if err != nil {
						t.Errorf("Error assigning variable value in scope")
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newFastEcAddAssignNewYHint()
				},
				check: allVarValueInScopeEquals(map[string]any{
					"new_y": big.NewInt(990),
					"value": big.NewInt(990),
				}),
			},
			{
				operanders: []*hintOperander{},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})

					slopeBig := big.NewInt(0)
					x0Big := big.NewInt(20)
					new_xBig := big.NewInt(10)
					y0Big := big.NewInt(10)

					err := ctx.ScopeManager.AssignVariable("slope", slopeBig)
					if err != nil {
						t.Errorf("Error assigning variable value in scope")
					}
					err = ctx.ScopeManager.AssignVariable("x0", x0Big)
					if err != nil {
						t.Errorf("Error assigning variable value in scope")
					}
					err = ctx.ScopeManager.AssignVariable("new_x", new_xBig)
					if err != nil {
						t.Errorf("Error assigning variable value in scope")
					}
					err = ctx.ScopeManager.AssignVariable("y0", y0Big)
					if err != nil {
						t.Errorf("Error assigning variable value in scope")
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newFastEcAddAssignNewYHint()
				},
				check: allVarValueInScopeEquals(map[string]any{
					"new_y": bigIntString("115792089237316195423570985008687907853269984665640564039457584007908834671653", 10),
					"value": bigIntString("115792089237316195423570985008687907853269984665640564039457584007908834671653", 10),
				}),
			},
			{
				operanders: []*hintOperander{},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})

					slopeBig := bigIntString("115792089237316195423570985008687907853269984665640564039457584007908834671683", 10)
					x0Big := big.NewInt(200)
					new_xBig := big.NewInt(199)
					y0Big := big.NewInt(20)

					err := ctx.ScopeManager.AssignVariable("slope", slopeBig)
					if err != nil {
						t.Errorf("Error assigning variable value in scope")
					}
					err = ctx.ScopeManager.AssignVariable("x0", x0Big)
					if err != nil {
						t.Errorf("Error assigning variable value in scope")
					}
					err = ctx.ScopeManager.AssignVariable("new_x", new_xBig)
					if err != nil {
						t.Errorf("Error assigning variable value in scope")
					}
					err = ctx.ScopeManager.AssignVariable("y0", y0Big)
					if err != nil {
						t.Errorf("Error assigning variable value in scope")
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newFastEcAddAssignNewYHint()
				},
				check: allVarValueInScopeEquals(map[string]any{
					"new_y": big.NewInt(0),
					"value": big.NewInt(0),
				}),
			},
		},
		"FastEcAddAssignNewX": {
			{
				operanders: []*hintOperander{
					{Name: "slope.d0", Kind: apRelative, Value: feltString("64081873649130491683833713")},
					{Name: "slope.d1", Kind: apRelative, Value: feltString("34843994309543177837008178")},
					{Name: "slope.d2", Kind: apRelative, Value: feltString("16548672716077616016846383")},
					{Name: "point0.x.d0", Kind: apRelative, Value: feltString("51215")},
					{Name: "point0.x.d1", Kind: apRelative, Value: feltString("36848548548458")},
					{Name: "point0.x.d2", Kind: apRelative, Value: feltString("634734734")},
					{Name: "point0.y.d0", Kind: apRelative, Value: feltString("26362")},
					{Name: "point0.y.d1", Kind: apRelative, Value: feltString("263724839599")},
					{Name: "point0.y.d2", Kind: apRelative, Value: feltString("901297012")},
					{Name: "point1.x.d0", Kind: apRelative, Value: feltString("45789")},
					{Name: "point1.x.d1", Kind: apRelative, Value: feltString("45612238789798")},
					{Name: "point1.x.d2", Kind: apRelative, Value: feltString("214455666")},
					{Name: "point1.y.d0", Kind: apRelative, Value: feltString("12457")},
					{Name: "point1.y.d1", Kind: apRelative, Value: feltString("895645646464")},
					{Name: "point1.y.d2", Kind: apRelative, Value: feltString("211245645")},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newFastEcAddAssignNewXHint(ctx.operanders["slope.d0"], ctx.operanders["point0.x.d0"], ctx.operanders["point1.x.d0"])
				},
				check: allVarValueInScopeEquals(map[string]any{
					"slope": bigIntString("99065496658741969395000079476826955370154683653966841736214499259699304892273", 10),
					"x0":    bigIntString("3799719333936312867907730225219317480871818784521830610814991", 10),
					"x1":    bigIntString("1283798249446970358602040710287144628881017552091260500619997", 10),
					"y0":    bigIntString("5395443952678709065478416501711989224759665054189740766553850", 10),
					"value": bigIntString("53863685200989733811273896838983614723181733288322685009664997422229669431265", 10),
					"new_x": bigIntString("53863685200989733811273896838983614723181733288322685009664997422229669431265", 10),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "slope.d0", Kind: apRelative, Value: feltString("0")},
					{Name: "slope.d1", Kind: apRelative, Value: feltString("0")},
					{Name: "slope.d2", Kind: apRelative, Value: feltString("0")},
					{Name: "point0.x.d0", Kind: apRelative, Value: feltString("0")},
					{Name: "point0.x.d1", Kind: apRelative, Value: feltString("0")},
					{Name: "point0.x.d2", Kind: apRelative, Value: feltString("0")},
					{Name: "point0.y.d0", Kind: apRelative, Value: feltString("0")},
					{Name: "point0.y.d1", Kind: apRelative, Value: feltString("0")},
					{Name: "point0.y.d2", Kind: apRelative, Value: feltString("0")},
					{Name: "point1.x.d0", Kind: apRelative, Value: feltString("0")},
					{Name: "point1.x.d1", Kind: apRelative, Value: feltString("0")},
					{Name: "point1.x.d2", Kind: apRelative, Value: feltString("0")},
					{Name: "point1.y.d0", Kind: apRelative, Value: feltString("0")},
					{Name: "point1.y.d1", Kind: apRelative, Value: feltString("0")},
					{Name: "point1.y.d2", Kind: apRelative, Value: feltString("0")},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newFastEcAddAssignNewXHint(ctx.operanders["slope.d0"], ctx.operanders["point0.x.d0"], ctx.operanders["point1.x.d0"])
				},
				check: allVarValueInScopeEquals(map[string]any{
					"slope": bigIntString("0", 10),
					"x0":    bigIntString("0", 10),
					"x1":    bigIntString("0", 10),
					"y0":    bigIntString("0", 10),
					"value": bigIntString("0", 10),
					"new_x": bigIntString("0", 10),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "slope.d0", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007908834671663")},
					{Name: "slope.d1", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007908834671663")},
					{Name: "slope.d2", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007908834671663")},
					{Name: "point0.x.d0", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007908834671663")},
					{Name: "point0.x.d1", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007908834671663")},
					{Name: "point0.x.d2", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007908834671663")},
					{Name: "point0.y.d0", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007908834671663")},
					{Name: "point0.y.d1", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007908834671663")},
					{Name: "point0.y.d2", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007908834671663")},
					{Name: "point1.x.d0", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007908834671663")},
					{Name: "point1.x.d1", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007908834671663")},
					{Name: "point1.x.d2", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007908834671663")},
					{Name: "point1.y.d0", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007908834671663")},
					{Name: "point1.y.d1", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007908834671663")},
					{Name: "point1.y.d2", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007908834671663")},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newFastEcAddAssignNewXHint(ctx.operanders["slope.d0"], ctx.operanders["point0.x.d0"], ctx.operanders["point1.x.d0"])
				},
				check: allVarValueInScopeEquals(map[string]any{
					"slope": bigIntString("-20441714640463444415550039378657358828977094550744864608392924301285287608509921726516187492362679433566942659569", 10),
					"x0":    bigIntString("-20441714640463444415550039378657358828977094550744864608392924301285287608509921726516187492362679433566942659569", 10),
					"x1":    bigIntString("-20441714640463444415550039378657358828977094550744864608392924301285287608509921726516187492362679433566942659569", 10),
					"y0":    bigIntString("-20441714640463444415550039378657358828977094550744864608392924301285287608509921726516187492362679433566942659569", 10),
					"value": bigIntString("30230181511926491618309110200401529297651013854327841200453332701540948849717", 10),
					"new_x": bigIntString("30230181511926491618309110200401529297651013854327841200453332701540948849717", 10),
				}),
			},
		},
		"EcDoubleSlopeV1": {
			{
				operanders: []*hintOperander{
					{Name: "point.x.d0", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007908834671663")},
					{Name: "point.x.d1", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007908834671663")},
					{Name: "point.x.d2", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007908834671663")},
					{Name: "point.y.d0", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007908834671663")},
					{Name: "point.y.d1", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007908834671663")},
					{Name: "point.y.d2", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007908834671663")},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcDoubleSlopeV1Hint(ctx.operanders["point.x.d0"])
				},
				check: allVarValueInScopeEquals(map[string]any{
					"x":     bigIntString("-20441714640463444415550039378657358828977094550744864608392924301285287608509921726516187492362679433566942659569", 10),
					"y":     bigIntString("-20441714640463444415550039378657358828977094550744864608392924301285287608509921726516187492362679433566942659569", 10),
					"value": bigIntString("57896044618658054023148353931790401374139058278927353312939393362691054630958", 10),
					"slope": bigIntString("57896044618658054023148353931790401374139058278927353312939393362691054630958", 10),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "point.x.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point.x.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point.x.d2", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point.y.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point.y.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point.y.d2", Kind: apRelative, Value: &utils.FeltZero},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcDoubleSlopeV1Hint(ctx.operanders["point.x.d0"])
				},
				errCheck: errorTextContains("point[1] % p == 0"),
			},
			{
				operanders: []*hintOperander{
					{Name: "point.x.d0", Kind: apRelative, Value: feltString("51215")},
					{Name: "point.x.d1", Kind: apRelative, Value: feltString("368485485484584")},
					{Name: "point.x.d2", Kind: apRelative, Value: feltString("4564564687987")},
					{Name: "point.y.d0", Kind: apRelative, Value: feltString("26362")},
					{Name: "point.y.d1", Kind: apRelative, Value: feltString("263724839599")},
					{Name: "point.y.d2", Kind: apRelative, Value: feltString("1321654896123789784652346")},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcDoubleSlopeV1Hint(ctx.operanders["point.x.d0"])
				},
				check: allVarValueInScopeEquals(map[string]any{
					"x":     bigIntString("27324902462242089002202715756360945650502697953428275540292323343", 10),
					"y":     bigIntString("7911836854973739773537612350570845963794165335703809150610926758199350552314", 10),
					"value": bigIntString("8532480558268366897328020348259450788170980412191993744326748439943456131995", 10),
					"slope": bigIntString("8532480558268366897328020348259450788170980412191993744326748439943456131995", 10),
				}),
			},
			{
				operanders: []*hintOperander{
					// 2**256 - 10
					{Name: "point.x.d0", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007913129639926")},
					// 2**256 - 10000
					{Name: "point.x.d1", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007913129629936")},
					// 2**256 - 100000
					{Name: "point.x.d2", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007913129539936")},
					// 2**256 - 68950
					{Name: "point.y.d0", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007913129570986")},
					// 2**256 - 4545646
					{Name: "point.y.d1", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007913125094290")},
					// 2**256 - 123124564
					{Name: "point.y.d2", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007913006515372")},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcDoubleSlopeV1Hint(ctx.operanders["point.x.d0"])
				},
				check: allVarValueInScopeEquals(map[string]any{
					"x":     bigIntString("-20441714640463444415550039378657358828977094550744838897976998602531359063761497108760225884244327111888086237226", 10),
					"y":     bigIntString("-20441714640463444415550039378657358828977094550744839634440263239133543719985711427903485175980925446193542139254", 10),
					"value": bigIntString("12124034331835072072286217393751993577217461936002714368452167712954852501083", 10),
					"slope": bigIntString("12124034331835072072286217393751993577217461936002714368452167712954852501083", 10),
				}),
			},
			{
				operanders: []*hintOperander{
					// 2**80
					{Name: "point.x.d0", Kind: apRelative, Value: feltString("1208925819614629174706176")},
					{Name: "point.x.d1", Kind: apRelative, Value: feltString("1208925819614629174706176")},
					{Name: "point.x.d2", Kind: apRelative, Value: feltString("1208925819614629174706176")},
					// 2**40
					{Name: "point.y.d0", Kind: apRelative, Value: feltString("1099511627776")},
					{Name: "point.y.d1", Kind: apRelative, Value: feltString("1099511627776")},
					{Name: "point.y.d2", Kind: apRelative, Value: feltString("1099511627776")},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcDoubleSlopeV1Hint(ctx.operanders["point.x.d0"])
				},
				check: allVarValueInScopeEquals(map[string]any{
					"x":     bigIntString("7237005577332262213973186656579099030007160806638365755234031873103428059136", 10),
					"y":     bigIntString("6582018229284824168619876815299993750165559328377972410848116736", 10),
					"value": bigIntString("154266052248863066452028362858593603519505739480817180031844352", 10),
					"slope": bigIntString("154266052248863066452028362858593603519505739480817180031844352", 10),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "point.x.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point.x.d1", Kind: apRelative, Value: &utils.FeltOne},
					{Name: "point.x.d2", Kind: apRelative, Value: feltUint64(2)},
					{Name: "point.y.d0", Kind: apRelative, Value: feltUint64(3)},
					{Name: "point.y.d1", Kind: apRelative, Value: feltUint64(4)},
					{Name: "point.y.d2", Kind: apRelative, Value: feltUint64(5)},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcDoubleSlopeV1Hint(ctx.operanders["point.x.d0"])
				},
				check: allVarValueInScopeEquals(map[string]any{
					"x":     bigIntString("11972621413014756705924586226983042952357666573254656", 10),
					"y":     bigIntString("29931553532536891764811465683514486063898567204929539", 10),
					"value": bigIntString("35023503208535022533116513151423452638642669107476233313413226008091253006355", 10),
					"slope": bigIntString("35023503208535022533116513151423452638642669107476233313413226008091253006355", 10),
				}),
			},
		},
	})
}
