package zero

import (
	"math/big"
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	secp_utils "github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/utils"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func TestZeroHintEc(t *testing.T) {
	runHinterTests(t, map[string][]hintTestCase{
		"EcNegate": {
			{
				operanders: []*hintOperander{
					// random values
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
					// GetSecPBig() % fp.Modulus()
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
					// GetSecPBig() % fp.Modulus()
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
					// GetSecPBig() % fp.Modulus()
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
					// GetSecPBig() % fp.Modulus()
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
					// GetSecPBig() % fp.Modulus()
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
					// GetSecPBig() % fp.Modulus()
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
					// GetSecPBig() % fp.Modulus()
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
					// GetSecPBig() % fp.Modulus() but with first digit 3 replaced with 7
					value := bigIntString("7618502788666127798953978732740734578953660990361066340291730267696802036752", 10)
					ctx.ScopeManager.EnterScope(map[string]any{"value": value})
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
					// GetSecPBig() % fp.Modulus()
					value := bigIntString("3618502788666127798953978732740734578953660990361066340291730267696802036752", 10)
					ctx.ScopeManager.EnterScope(map[string]any{"value": value})
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
					ctx.ScopeManager.EnterScope(map[string]any{"value": big.NewInt(123456)})
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
					ctx.ScopeManager.EnterScope(map[string]any{"value": big.NewInt(-123456)})
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
					// 2**86 - 1
					value := bigIntString("77371252455336267181195263", 10)
					ctx.ScopeManager.EnterScope(map[string]any{"value": value})
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
					// 2**86
					value := bigIntString("77371252455336267181195264", 10)
					ctx.ScopeManager.EnterScope(map[string]any{"value": value})
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
					// 2**86 + 1
					value := bigIntString("77371252455336267181195265", 10)
					ctx.ScopeManager.EnterScope(map[string]any{"value": value})
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
					ctx.ScopeManager.EnterScope(map[string]any{"value": big.NewInt(0)})
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
					// (2**86 - 1) * 2
					value := bigIntString("154742504910672534362390526", 10)
					ctx.ScopeManager.EnterScope(map[string]any{"value": value})
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
					// 2**86 * 2
					value := bigIntString("154742504910672534362390528", 10)
					ctx.ScopeManager.EnterScope(map[string]any{"value": value})
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
					// (2**86 + 1) * 2
					value := bigIntString("154742504910672534362390530", 10)
					ctx.ScopeManager.EnterScope(map[string]any{"value": value})
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
					slopeBig := big.NewInt(100)
					x0Big := big.NewInt(20)
					new_xBig := big.NewInt(10)
					y0Big := big.NewInt(10)
					secPBig, _ := secp_utils.GetSecPBig()

					ctx.ScopeManager.EnterScope(map[string]any{"slope": slopeBig, "x0": x0Big, "new_x": new_xBig, "y0": y0Big, "SECP_P": secPBig})
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
					slopeBig := big.NewInt(0)
					x0Big := big.NewInt(20)
					new_xBig := big.NewInt(10)
					y0Big := big.NewInt(10)
					secPBig, _ := secp_utils.GetSecPBig()

					ctx.ScopeManager.EnterScope(map[string]any{"slope": slopeBig, "x0": x0Big, "new_x": new_xBig, "y0": y0Big, "SECP_P": secPBig})
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
					// GetSecPBig() + 20
					slopeBig := bigIntString("115792089237316195423570985008687907853269984665640564039457584007908834671683", 10)
					x0Big := big.NewInt(200)
					new_xBig := big.NewInt(199)
					y0Big := big.NewInt(20)
					secPBig, _ := secp_utils.GetSecPBig()

					ctx.ScopeManager.EnterScope(map[string]any{"slope": slopeBig, "x0": x0Big, "new_x": new_xBig, "y0": y0Big, "SECP_P": secPBig})
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
					// random values
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
					{Name: "slope.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "slope.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "slope.d2", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point0.x.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point0.x.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point0.x.d2", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point0.y.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point0.y.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point0.y.d2", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point1.x.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point1.x.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point1.x.d2", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point1.y.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point1.y.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point1.y.d2", Kind: apRelative, Value: &utils.FeltZero},
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
					// GetSecPBig()
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
		"ComputeSlopeV1": {
			{
				operanders: []*hintOperander{
					{Name: "point0.x.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point0.x.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point0.x.d2", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point0.y.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point0.y.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point0.y.d2", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point1.x.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point1.x.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point1.x.d2", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point1.y.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point1.y.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point1.y.d2", Kind: apRelative, Value: &utils.FeltZero},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newComputeSlopeV1Hint(ctx.operanders["point0.x.d0"], ctx.operanders["point1.x.d0"])
				},
				errCheck: errorTextContains("0 is multiple of cairo Prime"),
			},
			{
				operanders: []*hintOperander{
					// random values
					{Name: "point0.x.d0", Kind: apRelative, Value: feltString("134")},
					{Name: "point0.x.d1", Kind: apRelative, Value: feltString("5123")},
					{Name: "point0.x.d2", Kind: apRelative, Value: feltString("140")},
					{Name: "point0.y.d0", Kind: apRelative, Value: feltString("1232")},
					{Name: "point0.y.d1", Kind: apRelative, Value: feltString("4652")},
					{Name: "point0.y.d2", Kind: apRelative, Value: feltString("720")},
					{Name: "point1.x.d0", Kind: apRelative, Value: feltString("156")},
					{Name: "point1.x.d1", Kind: apRelative, Value: feltString("6545")},
					{Name: "point1.x.d2", Kind: apRelative, Value: feltString("100010")},
					{Name: "point1.y.d0", Kind: apRelative, Value: feltString("1123")},
					{Name: "point1.y.d1", Kind: apRelative, Value: feltString("1325")},
					{Name: "point1.y.d2", Kind: apRelative, Value: feltString("910")},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},

				makeHinter: func(ctx *hintTestContext) hinter.Hinter {

					return newComputeSlopeV1Hint(ctx.operanders["point0.x.d0"], ctx.operanders["point1.x.d0"])
				},
				// Printing the value of point0.x.d0

				check: allVarValueInScopeEquals(map[string]any{
					"slope": bigIntString("41419765295989780131385135514529906223027172305400087935755859001910844026631", 10),
					"value": bigIntString("41419765295989780131385135514529906223027172305400087935755859001910844026631", 10),
				}),
				//fmt.Printf("Name: %s, Age: %d\n", slope, value)
			},
		},
	})
}
