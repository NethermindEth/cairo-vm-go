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
					return newFastEcAddAssignNewXHint(ctx.operanders["slope.d0"], ctx.operanders["point0.x.d0"], ctx.operanders["point1.x.d0"])
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
					return newFastEcAddAssignNewXHint(ctx.operanders["slope.d0"], ctx.operanders["point0.x.d0"], ctx.operanders["point1.x.d0"])
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
					return newFastEcAddAssignNewXHint(ctx.operanders["slope.d0"], ctx.operanders["point0.x.d0"], ctx.operanders["point1.x.d0"])
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
		"EcDoubleAssignNewXV1": {
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
					return newEcDoubleAssignNewXV1Hint(ctx.operanders["slope.d0"], ctx.operanders["point.x.d0"])
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
					return newEcDoubleAssignNewXV1Hint(ctx.operanders["slope.d0"], ctx.operanders["point.x.d0"])
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
					return newEcDoubleAssignNewXV1Hint(ctx.operanders["slope.d0"], ctx.operanders["point.x.d0"])
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
		"ReduceV1": {
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
					return newReduceV1Hint(ctx.operanders["x.d0"])
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
					return newReduceV1Hint(ctx.operanders["x.d0"])
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
					return newReduceV1Hint(ctx.operanders["x.d0"])
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
					return newReduceV1Hint(ctx.operanders["x.d0"])
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
					return newReduceV1Hint(ctx.operanders["x.d0"])
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
			{
				operanders: []*hintOperander{},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("x", bigIntString("101", 10))
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
	},
	)
}
