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
					err := ctx.ScopeManager.AssignVariable("value", value)
					if err != nil {
						t.Fatal(err)
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
					// GetSecPBig() % fp.Modulus()
					value := bigIntString("3618502788666127798953978732740734578953660990361066340291730267696802036752", 10)
					err := ctx.ScopeManager.AssignVariable("value", value)
					if err != nil {
						t.Fatal(err)
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
					err := ctx.ScopeManager.AssignVariable("value", big.NewInt(123456))
					if err != nil {
						t.Fatal(err)
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
					err := ctx.ScopeManager.AssignVariable("value", big.NewInt(-123456))
					if err != nil {
						t.Fatal(err)
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
					// 2**86 - 1
					value := bigIntString("77371252455336267181195263", 10)
					err := ctx.ScopeManager.AssignVariable("value", value)
					if err != nil {
						t.Fatal(err)
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
					// 2**86
					value := bigIntString("77371252455336267181195264", 10)
					err := ctx.ScopeManager.AssignVariable("value", value)
					if err != nil {
						t.Fatal(err)
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
					// 2**86 + 1
					value := bigIntString("77371252455336267181195265", 10)
					err := ctx.ScopeManager.AssignVariable("value", value)
					if err != nil {
						t.Fatal(err)
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
					err := ctx.ScopeManager.AssignVariable("value", big.NewInt(0))
					if err != nil {
						t.Fatal(err)
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
					// (2**86 - 1) * 2
					value := bigIntString("154742504910672534362390526", 10)
					err := ctx.ScopeManager.AssignVariable("value", value)
					if err != nil {
						t.Fatal(err)
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
					// 2**86 * 2
					value := bigIntString("154742504910672534362390528", 10)
					err := ctx.ScopeManager.AssignVariable("value", value)
					if err != nil {
						t.Fatal(err)
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
					// (2**86 + 1) * 2
					value := bigIntString("154742504910672534362390530", 10)
					err := ctx.ScopeManager.AssignVariable("value", value)
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newNondetBigint3V1Hint(ctx.operanders["res"])
				},
				check: consecutiveVarValueEquals("res", []*fp.Element{feltString("2"), feltString("2"), &utils.FeltZero}),
			},
		},
		"FastEcAddAssignNewX": {
			{
				operanders: []*hintOperander{
					// random values
					{Name: "slope.d0", Kind: apRelative, Value: feltString("64081873649130491683833713")},
					{Name: "slope.d1", Kind: apRelative, Value: feltString("34843994309543177837008178")},
					{Name: "slope.d2", Kind: apRelative, Value: feltString("16548672716077616016846383")},
					{Name: "point0.x.d0", Kind: apRelative, Value: feltUint64(51215)},
					{Name: "point0.x.d1", Kind: apRelative, Value: feltUint64(36848548548458)},
					{Name: "point0.x.d2", Kind: apRelative, Value: feltUint64(634734734)},
					{Name: "point0.y.d0", Kind: apRelative, Value: feltUint64(26362)},
					{Name: "point0.y.d1", Kind: apRelative, Value: feltUint64(263724839599)},
					{Name: "point0.y.d2", Kind: apRelative, Value: feltUint64(901297012)},
					{Name: "point1.x.d0", Kind: apRelative, Value: feltUint64(45789)},
					{Name: "point1.x.d1", Kind: apRelative, Value: feltUint64(45612238789798)},
					{Name: "point1.x.d2", Kind: apRelative, Value: feltUint64(214455666)},
					{Name: "point1.y.d0", Kind: apRelative, Value: feltUint64(12457)},
					{Name: "point1.y.d1", Kind: apRelative, Value: feltUint64(895645646464)},
					{Name: "point1.y.d2", Kind: apRelative, Value: feltUint64(211245645)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {

					secPBig, ok := secp_utils.GetSecPBig()
					if !ok {
						return nil
					}
					return newFastEcAddAssignNewXHint(ctx.operanders["slope.d0"], ctx.operanders["point0.x.d0"], ctx.operanders["point1.x.d0"], secPBig)
				},
				check: allVarValueInScopeEquals(map[string]any{
					"slope": bigIntString("99065496658741969395000079476826955370154683653966841736214499259699304892273", 10),
					"x0":    bigIntString("3799719333936312867907730225219317480871818784521830610814991", 10),
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
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					secPBig, ok := secp_utils.GetSecPBig()
					if !ok {
						return nil
					}
					return newFastEcAddAssignNewXHint(ctx.operanders["slope.d0"], ctx.operanders["point0.x.d0"], ctx.operanders["point1.x.d0"], secPBig)
				},
				check: allVarValueInScopeEquals(map[string]any{
					"slope": bigIntString("0", 10),
					"x0":    bigIntString("0", 10),
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
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					secPBig, ok := secp_utils.GetSecPBig()
					if !ok {
						return nil
					}
					return newFastEcAddAssignNewXHint(ctx.operanders["slope.d0"], ctx.operanders["point0.x.d0"], ctx.operanders["point1.x.d0"], secPBig)
				},
				check: allVarValueInScopeEquals(map[string]any{
					"slope": bigIntString("-20441714640463444415550039378657358828977094550744864608392924301285287608509921726516187492362679433566942659569", 10),
					"x0":    bigIntString("-20441714640463444415550039378657358828977094550744864608392924301285287608509921726516187492362679433566942659569", 10),
					"y0":    bigIntString("-20441714640463444415550039378657358828977094550744864608392924301285287608509921726516187492362679433566942659569", 10),
					"value": bigIntString("30230181511926491618309110200401529297651013854327841200453332701540948849717", 10),
					"new_x": bigIntString("30230181511926491618309110200401529297651013854327841200453332701540948849717", 10),
				}),
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

					err := ctx.ScopeManager.AssignVariables(map[string]any{"slope": slopeBig, "x0": x0Big, "new_x": new_xBig, "y0": y0Big, "SECP_P": &secPBig})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newFastEcAddAssignNewYHint()
				},
				check: allVarValueInScopeEquals(map[string]any{
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

					err := ctx.ScopeManager.AssignVariables(map[string]any{"slope": slopeBig, "x0": x0Big, "new_x": new_xBig, "y0": y0Big, "SECP_P": &secPBig})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newFastEcAddAssignNewYHint()
				},
				check: allVarValueInScopeEquals(map[string]any{
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

					err := ctx.ScopeManager.AssignVariables(map[string]any{"slope": slopeBig, "x0": x0Big, "new_x": new_xBig, "y0": y0Big, "SECP_P": &secPBig})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newFastEcAddAssignNewYHint()
				},
				check: allVarValueInScopeEquals(map[string]any{
					"value": big.NewInt(0),
				}),
			},
		},
		"EcDoubleSlopeV1": {
			{
				operanders: []*hintOperander{
					{Name: "point.x.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point.x.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point.x.d2", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point.y.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point.y.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point.y.d2", Kind: apRelative, Value: &utils.FeltZero},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcDoubleSlopeV1Hint(ctx.operanders["point.x.d0"])
				},
				errCheck: errorTextContains("point[1] % p == 0"),
			},
			{
				operanders: []*hintOperander{
					// values are random
					{Name: "point.x.d0", Kind: apRelative, Value: feltUint64(51215)},
					{Name: "point.x.d1", Kind: apRelative, Value: feltUint64(368485485484584)},
					{Name: "point.x.d2", Kind: apRelative, Value: feltUint64(4564564687987)},
					{Name: "point.y.d0", Kind: apRelative, Value: feltUint64(26362)},
					{Name: "point.y.d1", Kind: apRelative, Value: feltUint64(263724839599)},
					{Name: "point.y.d2", Kind: apRelative, Value: feltString("1321654896123789784652346")},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcDoubleSlopeV1Hint(ctx.operanders["point.x.d0"])
				},
				check: allVarValueInScopeEquals(map[string]any{
					"value": bigIntString("8532480558268366897328020348259450788170980412191993744326748439943456131995", 10),
				}),
			},
			{
				operanders: []*hintOperander{
					// 2**80
					{Name: "point.x.d0", Kind: apRelative, Value: feltString("1208925819614629174706176")},
					{Name: "point.x.d1", Kind: apRelative, Value: feltString("1208925819614629174706176")},
					{Name: "point.x.d2", Kind: apRelative, Value: feltString("1208925819614629174706176")},
					// 2**40
					{Name: "point.y.d0", Kind: apRelative, Value: feltUint64(1099511627776)},
					{Name: "point.y.d1", Kind: apRelative, Value: feltUint64(1099511627776)},
					{Name: "point.y.d2", Kind: apRelative, Value: feltUint64(1099511627776)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcDoubleSlopeV1Hint(ctx.operanders["point.x.d0"])
				},
				check: allVarValueInScopeEquals(map[string]any{
					"value": bigIntString("154266052248863066452028362858593603519505739480817180031844352", 10),
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
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcDoubleSlopeV1Hint(ctx.operanders["point.x.d0"])
				},
				check: allVarValueInScopeEquals(map[string]any{
					"value": bigIntString("35023503208535022533116513151423452638642669107476233313413226008091253006355", 10),
				}),
			},
		},
		"EcDoubleSlopeV3": {
			{
				operanders: []*hintOperander{
					// values are random
					{Name: "pt.x.d0", Kind: apRelative, Value: feltUint64(51215)},
					{Name: "pt.x.d1", Kind: apRelative, Value: feltUint64(368485485484584)},
					{Name: "pt.x.d2", Kind: apRelative, Value: feltUint64(4564564687987)},
					{Name: "pt.y.d0", Kind: apRelative, Value: feltUint64(26362)},
					{Name: "pt.y.d1", Kind: apRelative, Value: feltUint64(263724839599)},
					{Name: "pt.y.d2", Kind: apRelative, Value: feltString("1321654896123789784652346")},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcDoubleSlopeV3Hint(ctx.operanders["pt.x.d0"])
				},
				check: allVarValueInScopeEquals(map[string]any{
					"value": bigIntString("8532480558268366897328020348259450788170980412191993744326748439943456131995", 10),
				}),
			},
			{
				operanders: []*hintOperander{
					// 2**80
					{Name: "pt.x.d0", Kind: apRelative, Value: feltString("1208925819614629174706176")},
					{Name: "pt.x.d1", Kind: apRelative, Value: feltString("1208925819614629174706176")},
					{Name: "pt.x.d2", Kind: apRelative, Value: feltString("1208925819614629174706176")},
					// 2**40
					{Name: "pt.y.d0", Kind: apRelative, Value: feltUint64(1099511627776)},
					{Name: "pt.y.d1", Kind: apRelative, Value: feltUint64(1099511627776)},
					{Name: "pt.y.d2", Kind: apRelative, Value: feltUint64(1099511627776)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcDoubleSlopeV3Hint(ctx.operanders["pt.x.d0"])
				},
				check: allVarValueInScopeEquals(map[string]any{
					"value": bigIntString("154266052248863066452028362858593603519505739480817180031844352", 10),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "pt.x.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "pt.x.d1", Kind: apRelative, Value: &utils.FeltOne},
					{Name: "pt.x.d2", Kind: apRelative, Value: feltUint64(2)},
					{Name: "pt.y.d0", Kind: apRelative, Value: feltUint64(3)},
					{Name: "pt.y.d1", Kind: apRelative, Value: feltUint64(4)},
					{Name: "pt.y.d2", Kind: apRelative, Value: feltUint64(5)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcDoubleSlopeV1Hint(ctx.operanders["pt.x.d0"])
				},
				check: allVarValueInScopeEquals(map[string]any{
					"value": bigIntString("35023503208535022533116513151423452638642669107476233313413226008091253006355", 10),
				}),
			},
		},
		"EcDoubleAssignNewX": {
			{
				operanders: []*hintOperander{
					{Name: "slope.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "slope.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "slope.d2", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point.x.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point.x.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point.x.d2", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point.y.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point.y.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "point.y.d2", Kind: apRelative, Value: &utils.FeltZero},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcDoubleAssignNewXHint(ctx.operanders["slope.d0"], ctx.operanders["point.x.d0"])
				},
				check: allVarValueInScopeEquals(map[string]any{
					"slope": bigIntString("0", 10),
					"x":     bigIntString("0", 10),
					"y":     bigIntString("0", 10),
					"value": bigIntString("0", 10),
					"new_x": bigIntString("0", 10),
				}),
			},
			{
				operanders: []*hintOperander{
					// random values
					{Name: "slope.d0", Kind: apRelative, Value: feltString("75893937474639987141425142")},
					{Name: "slope.d1", Kind: apRelative, Value: feltString("99484727364721283590428239")},
					{Name: "slope.d2", Kind: apRelative, Value: feltString("89273748821013318302045802")},
					{Name: "point.x.d0", Kind: apRelative, Value: feltUint64(84737)},
					{Name: "point.x.d1", Kind: apRelative, Value: feltUint64(823758498371235)},
					{Name: "point.x.d2", Kind: apRelative, Value: feltUint64(7684847382874)},
					{Name: "point.y.d0", Kind: apRelative, Value: feltUint64(3244612)},
					{Name: "point.y.d1", Kind: apRelative, Value: feltUint64(83478234123)},
					{Name: "point.y.d2", Kind: apRelative, Value: feltString("6837128718738732781737")},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcDoubleAssignNewXHint(ctx.operanders["slope.d0"], ctx.operanders["point.x.d0"])
				},
				check: allVarValueInScopeEquals(map[string]any{
					"x":     bigIntString("46003884165973832456933262296354598115596485770084020998681742081", 10),
					"y":     bigIntString("40929176850754749976490215751286880883172677177113744707710685171812500036", 10),
					"value": bigIntString("112687466468745171568302397569403892765553175022602416609657454443897975462107", 10),
					"slope": bigIntString("534420398377282472759697724574779677488837473675614769635757753630086524221430", 10),
					"new_x": bigIntString("112687466468745171568302397569403892765553175022602416609657454443897975462107", 10),
				}),
			},
			{
				operanders: []*hintOperander{
					// GetSecPBig()
					{Name: "slope.d0", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007908834671663")},
					{Name: "slope.d1", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007908834671663")},
					{Name: "slope.d2", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007908834671663")},
					{Name: "point.x.d0", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007908834671663")},
					{Name: "point.x.d1", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007908834671663")},
					{Name: "point.x.d2", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007908834671663")},
					{Name: "point.y.d0", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007908834671663")},
					{Name: "point.y.d1", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007908834671663")},
					{Name: "point.y.d2", Kind: apRelative, Value: feltString("115792089237316195423570985008687907853269984665640564039457584007908834671663")},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcDoubleAssignNewXHint(ctx.operanders["slope.d0"], ctx.operanders["point.x.d0"])
				},
				check: allVarValueInScopeEquals(map[string]any{
					"x":     bigIntString("-20441714640463444415550039378657358828977094550744864608392924301285287608509921726516187492362679433566942659569", 10),
					"y":     bigIntString("-20441714640463444415550039378657358828977094550744864608392924301285287608509921726516187492362679433566942659569", 10),
					"value": bigIntString("30230181511926491618309110200401529297651013854327841200453332701540948849717", 10),
					"slope": bigIntString("-20441714640463444415550039378657358828977094550744864608392924301285287608509921726516187492362679433566942659569", 10),
					"new_x": bigIntString("30230181511926491618309110200401529297651013854327841200453332701540948849717", 10),
				}),
			},
		},
		"EcDoubleAssignNewYV1": {
			{
				operanders: []*hintOperander{},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					slopeBig := big.NewInt(100)
					xBig := big.NewInt(20)
					new_xBig := big.NewInt(10)
					yBig := big.NewInt(10)
					secPBig, _ := secp_utils.GetSecPBig()

					err := ctx.ScopeManager.AssignVariables(map[string]any{"slope": slopeBig, "x": xBig, "new_x": new_xBig, "y": yBig, "SECP_P": &secPBig})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcDoubleAssignNewYV1Hint()
				},
				check: allVarValueInScopeEquals(map[string]any{
					"value": big.NewInt(990),
				}),
			},
			{
				operanders: []*hintOperander{},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					slopeBig := big.NewInt(0)
					xBig := big.NewInt(20)
					new_xBig := big.NewInt(10)
					yBig := big.NewInt(10)
					secPBig, _ := secp_utils.GetSecPBig()

					err := ctx.ScopeManager.AssignVariables(map[string]any{"slope": slopeBig, "x": xBig, "new_x": new_xBig, "y": yBig, "SECP_P": &secPBig})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcDoubleAssignNewYV1Hint()
				},
				check: allVarValueInScopeEquals(map[string]any{
					"value": bigIntString("115792089237316195423570985008687907853269984665640564039457584007908834671653", 10),
				}),
			},
			{
				operanders: []*hintOperander{},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					// GetSecPBig() + 20
					slopeBig := bigIntString("115792089237316195423570985008687907853269984665640564039457584007908834671683", 10)
					xBig := big.NewInt(200)
					new_xBig := big.NewInt(199)
					yBig := big.NewInt(20)
					secPBig, _ := secp_utils.GetSecPBig()

					err := ctx.ScopeManager.AssignVariables(map[string]any{"slope": slopeBig, "x": xBig, "new_x": new_xBig, "y": yBig, "SECP_P": &secPBig})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcDoubleAssignNewYV1Hint()
				},
				check: allVarValueInScopeEquals(map[string]any{
					"value": big.NewInt(0),
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
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newComputeSlopeV1Hint(ctx.operanders["point0.x.d0"], ctx.operanders["point1.x.d0"])
				},
				errCheck: errorTextContains("the slope of the line is invalid"),
			},
			{
				operanders: []*hintOperander{
					// random values
					{Name: "point0.x.d0", Kind: apRelative, Value: feltInt64(134)},
					{Name: "point0.x.d1", Kind: apRelative, Value: feltInt64(5123)},
					{Name: "point0.x.d2", Kind: apRelative, Value: feltInt64(140)},
					{Name: "point0.y.d0", Kind: apRelative, Value: feltInt64(1232)},
					{Name: "point0.y.d1", Kind: apRelative, Value: feltInt64(4652)},
					{Name: "point0.y.d2", Kind: apRelative, Value: feltInt64(720)},
					{Name: "point1.x.d0", Kind: apRelative, Value: feltInt64(156)},
					{Name: "point1.x.d1", Kind: apRelative, Value: feltInt64(6545)},
					{Name: "point1.x.d2", Kind: apRelative, Value: feltInt64(100010)},
					{Name: "point1.y.d0", Kind: apRelative, Value: feltInt64(1123)},
					{Name: "point1.y.d1", Kind: apRelative, Value: feltInt64(1325)},
					{Name: "point1.y.d2", Kind: apRelative, Value: feltInt64(910)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {

					return newComputeSlopeV1Hint(ctx.operanders["point0.x.d0"], ctx.operanders["point1.x.d0"])
				},
				check: allVarValueInScopeEquals(map[string]any{
					"value": bigIntString("41419765295989780131385135514529906223027172305400087935755859001910844026631", 10),
				}),
			},
		},
		"ComputeSlopeV3": {
			{
				operanders: []*hintOperander{
					// random values
					{Name: "pt0.x.d0", Kind: apRelative, Value: feltInt64(134)},
					{Name: "pt0.x.d1", Kind: apRelative, Value: feltInt64(5123)},
					{Name: "pt0.x.d2", Kind: apRelative, Value: feltInt64(140)},
					{Name: "pt0.y.d0", Kind: apRelative, Value: feltInt64(1232)},
					{Name: "pt0.y.d1", Kind: apRelative, Value: feltInt64(4652)},
					{Name: "pt0.y.d2", Kind: apRelative, Value: feltInt64(720)},
					{Name: "pt1.x.d0", Kind: apRelative, Value: feltInt64(156)},
					{Name: "pt1.x.d1", Kind: apRelative, Value: feltInt64(6545)},
					{Name: "pt1.x.d2", Kind: apRelative, Value: feltInt64(100010)},
					{Name: "pt1.y.d0", Kind: apRelative, Value: feltInt64(1123)},
					{Name: "pt1.y.d1", Kind: apRelative, Value: feltInt64(1325)},
					{Name: "pt1.y.d2", Kind: apRelative, Value: feltInt64(910)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {

					return newComputeSlopeV3Hint(ctx.operanders["pt0.x.d0"], ctx.operanders["pt1.x.d0"])
				},
				check: allVarValueInScopeEquals(map[string]any{
					"value": bigIntString("41419765295989780131385135514529906223027172305400087935755859001910844026631", 10),
				}),
			},
		},
		"Reduce": {
			{
				operanders: []*hintOperander{
					{Name: "x.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "x.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "x.d2", Kind: apRelative, Value: &utils.FeltZero},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					hinter.InitializeScopeManager(ctx, map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newReduceHint(ctx.operanders["x.d0"])
				},
				check: varValueInScopeEquals("value", bigIntString("0", 10)),
			},
			{
				operanders: []*hintOperander{
					{Name: "x.d0", Kind: apRelative, Value: feltString("3618502788666131213697322783095070105623107215331596699973092056135872020482")},
					{Name: "x.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "x.d2", Kind: apRelative, Value: &utils.FeltZero},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					hinter.InitializeScopeManager(ctx, map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newReduceHint(ctx.operanders["x.d0"])
				},
				check: varValueInScopeEquals("value", bigIntString("1", 10)),
			},
			{
				operanders: []*hintOperander{
					{Name: "x.d0", Kind: apRelative, Value: feltString("10")},
					{Name: "x.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "x.d2", Kind: apRelative, Value: &utils.FeltZero},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					hinter.InitializeScopeManager(ctx, map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newReduceHint(ctx.operanders["x.d0"])
				},
				check: varValueInScopeEquals("value", bigIntString("10", 10)),
			},
			{
				operanders: []*hintOperander{
					{Name: "x.d0", Kind: apRelative, Value: feltString("1")},
					{Name: "x.d1", Kind: apRelative, Value: feltString("2")},
					{Name: "x.d2", Kind: apRelative, Value: feltString("3")},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					hinter.InitializeScopeManager(ctx, map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newReduceHint(ctx.operanders["x.d0"])
				},
				check: varValueInScopeEquals("value", bigIntString("17958932119522135058886879379160190656204633450479617", 10)),
			},
			{
				operanders: []*hintOperander{
					{Name: "x.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "x.d1", Kind: apRelative, Value: feltString("2")},
					{Name: "x.d2", Kind: apRelative, Value: feltString("3")},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					hinter.InitializeScopeManager(ctx, map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newReduceHint(ctx.operanders["x.d0"])
				},
				check: varValueInScopeEquals("value", bigIntString("17958932119522135058886879379160190656204633450479616", 10)),
			},
		},
		"EcMulInner": {
			{
				operanders: []*hintOperander{
					{Name: "scalar", Kind: apRelative, Value: feltUint64(10)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcMulInnerHint(ctx.operanders["scalar"])
				},
				check: apValueEquals(&utils.FeltZero),
			},
			{
				operanders: []*hintOperander{
					{Name: "scalar", Kind: apRelative, Value: feltUint64(19)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcMulInnerHint(ctx.operanders["scalar"])
				},
				check: apValueEquals(&utils.FeltOne),
			},
		},
		"IsZeroNondet": {
			{
				operanders: []*hintOperander{},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("x", bigIntString("0", 10))
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newIsZeroNondetHint()
				},
				check: apValueEquals(feltUint64(1)),
			},
			{
				operanders: []*hintOperander{},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("x", bigIntString("42", 10))
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newIsZeroNondetHint()
				},
				check: apValueEquals(feltUint64(0)),
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
		"RecoverY": {
			{
				operanders: []*hintOperander{
					{Name: "x", Kind: apRelative, Value: feltString("2497468900767850684421727063357792717599762502387246235265616708902555305129")},
					{Name: "p.x", Kind: uninitialized},
					{Name: "p.y", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newRecoverYHint(ctx.operanders["x"], ctx.operanders["p.x"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"p.x": feltString("2497468900767850684421727063357792717599762502387246235265616708902555305129"),
					"p.y": feltString("205857351767627712295703269674687767888261140702556021834663354704341414042"),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "x", Kind: apRelative, Value: feltString("205857351767627712295703269674687767888261140702556021834663354704341414042")},
					{Name: "p.x", Kind: uninitialized},
					{Name: "p.y", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newRecoverYHint(ctx.operanders["x"], ctx.operanders["p.x"])
				},
				errCheck: errorTextContains("does not represent the x coordinate of a point on the curve"),
			},
			{
				operanders: []*hintOperander{
					{Name: "x", Kind: apRelative, Value: feltString("3004956058830981475544150447242655232275382685012344776588097793621230049020")},
					{Name: "p.x", Kind: uninitialized},
					{Name: "p.y", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newRecoverYHint(ctx.operanders["x"], ctx.operanders["p.x"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"p.x": feltString("3004956058830981475544150447242655232275382685012344776588097793621230049020"),
					"p.y": feltString("386236054595386575795345623791920124827519018828430310912260655089307618738"),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "x", Kind: apRelative, Value: feltString("138597138396302485058562442936200017709939129389766076747102238692717075504")},
					{Name: "p.x", Kind: uninitialized},
					{Name: "p.y", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newRecoverYHint(ctx.operanders["x"], ctx.operanders["p.x"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"p.x": feltString("138597138396302485058562442936200017709939129389766076747102238692717075504"),
					"p.y": feltString("1116947097676727397390632683964789044871379304271794004325353078455954290524"),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "x", Kind: apRelative, Value: feltString("71635783675677659163985681365816684268526846280467284682674852685628658265882465826464572245")},
					{Name: "p.x", Kind: uninitialized},
					{Name: "p.y", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newRecoverYHint(ctx.operanders["x"], ctx.operanders["p.x"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"p.x": feltString("71635783675677659163985681365816684268526846280467284682674852685628658265882465826464572245"),
					"p.y": feltString("903372048565605391120071143811887302063650776015287438589675702929494830362"),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "x", Kind: apRelative, Value: feltString("42424242424242424242")},
					{Name: "p.x", Kind: uninitialized},
					{Name: "p.y", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newRecoverYHint(ctx.operanders["x"], ctx.operanders["p.x"])
				},
				errCheck: errorTextContains("does not represent the x coordinate of a point on the curve"),
			},
		},
		"RandomEcPoint": {
			{
				operanders: []*hintOperander{
					{Name: "p.x", Kind: apRelative, Value: feltString("3004956058830981475544150447242655232275382685012344776588097793621230049020")},
					{Name: "p.y", Kind: apRelative, Value: feltString("3232266734070744637901977159303149980795588196503166389060831401046564401743")},
					{Name: "m", Kind: apRelative, Value: feltUint64(34)},
					{Name: "q.x", Kind: apRelative, Value: feltString("2864041794633455918387139831609347757720597354645583729611044800117714995244")},
					{Name: "q.y", Kind: apRelative, Value: feltString("2252415379535459416893084165764951913426528160630388985542241241048300343256")},
					{Name: "s.x", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newRandomEcPointHint(
						ctx.operanders["p.x"],
						ctx.operanders["m"],
						ctx.operanders["q.x"],
						ctx.operanders["s.x"],
					)
				},
				check: consecutiveVarValueEquals("s.x", []*fp.Element{
					feltString("96578541406087262240552119423829615463800550101008760434566010168435227837635"),
					feltString("3412645436898503501401619513420382337734846074629040678138428701431530606439"),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "p.x", Kind: apRelative, Value: feltUint64(12345)},
					{Name: "p.y", Kind: apRelative, Value: feltUint64(6789)},
					{Name: "m", Kind: apRelative, Value: feltUint64(101)},
					{Name: "q.x", Kind: apRelative, Value: feltUint64(98765)},
					{Name: "q.y", Kind: apRelative, Value: feltUint64(4321)},
					{Name: "s.x", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newRandomEcPointHint(
						ctx.operanders["p.x"],
						ctx.operanders["m"],
						ctx.operanders["q.x"],
						ctx.operanders["s.x"],
					)
				},
				check: consecutiveVarValueEquals("s.x", []*fp.Element{
					feltString("39190969885360777615413526676655883809466222002423777590585892821354159079496"),
					feltString("533983185449702770508526175744869430974740140562200547506631069957329272485"),
				}),
			},
		},
		"ChainedEcOp": {
			{
				operanders: []*hintOperander{
					{Name: "len", Kind: apRelative, Value: feltUint64(3)},
					{Name: "p.x", Kind: apRelative, Value: feltString("3004956058830981475544150447242655232275382685012344776588097793621230049020")},
					{Name: "p.y", Kind: apRelative, Value: feltString("3232266734070744637901977159303149980795588196503166389060831401046564401743")},
					{Name: "m", Kind: apRelative, Value: addr(8)},
					{Name: "m_value", Kind: apRelative, Value: feltUint64(34)},
					{Name: "m_value", Kind: apRelative, Value: feltUint64(34)},
					{Name: "m_value", Kind: apRelative, Value: feltUint64(34)},
					{Name: "q.q1.x", Kind: apRelative, Value: feltString("2864041794633455918387139831609347757720597354645583729611044800117714995244")},
					{Name: "q.q1.y", Kind: apRelative, Value: feltString("2252415379535459416893084165764951913426528160630388985542241241048300343256")},
					{Name: "q.q2.x", Kind: apRelative, Value: feltString("2864041794633455918387139831609347757720597354645583729611044800117714995244")},
					{Name: "q.q2.y", Kind: apRelative, Value: feltString("2252415379535459416893084165764951913426528160630388985542241241048300343256")},
					{Name: "q.q3.x", Kind: apRelative, Value: feltString("2864041794633455918387139831609347757720597354645583729611044800117714995244")},
					{Name: "q.q3.y", Kind: apRelative, Value: feltString("2252415379535459416893084165764951913426528160630388985542241241048300343256")},
					{Name: "q", Kind: apRelative, Value: addrWithSegment(1, 11)},
					{Name: "s", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newChainedEcOpHint(
						ctx.operanders["len"],
						ctx.operanders["p.x"],
						ctx.operanders["m"],
						ctx.operanders["q"],
						ctx.operanders["s"],
					)
				},
				check: consecutiveVarValueEquals("s", []*fp.Element{
					feltString("1354562415074475070179359167082942891834423311678180448592849484844152837347"),
					feltString("907662328694455187848008017177970257426839229889571025406355869359245158736"),
				}),
			},
		},
	},
	)
}
