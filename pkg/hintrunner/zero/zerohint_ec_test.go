package zero

import (
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
				check: varValueInScopeEquals("value", "83121579216557378445487899878180864668798711284981320763518679672151497189239"),
			},
			{
				operanders: []*hintOperander{
					{Name: "x.d0", Kind: apRelative, Value: feltString("0")},
					{Name: "x.d1", Kind: apRelative, Value: feltString("0")},
					{Name: "x.d2", Kind: apRelative, Value: feltString("0")},
					{Name: "y.d0", Kind: apRelative, Value: feltString("0")},
					{Name: "y.d1", Kind: apRelative, Value: feltString("0")},
					{Name: "y.d2", Kind: apRelative, Value: feltString("0")},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcNegateHint(ctx.operanders["x.d0"])
				},
				check: varValueInScopeEquals("value", "0"),
			},
			{
				operanders: []*hintOperander{
					{Name: "x.d0", Kind: apRelative, Value: feltString("3618502788666127798953978732740734578953660990361066340291730267696802036752")},
					{Name: "x.d1", Kind: apRelative, Value: feltString("0")},
					{Name: "x.d2", Kind: apRelative, Value: feltString("0")},
					{Name: "y.d0", Kind: apRelative, Value: feltString("3618502788666127798953978732740734578953660990361066340291730267696802036752")},
					{Name: "y.d1", Kind: apRelative, Value: feltString("0")},
					{Name: "y.d2", Kind: apRelative, Value: feltString("0")},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcNegateHint(ctx.operanders["x.d0"])
				},
				check: varValueInScopeEquals("value", "3414743344050354335526669446224970530359681361788439069983729"),
			},
			{
				operanders: []*hintOperander{
					{Name: "x.d0", Kind: apRelative, Value: feltString("0")},
					{Name: "x.d1", Kind: apRelative, Value: feltString("3618502788666127798953978732740734578953660990361066340291730267696802036752")},
					{Name: "x.d2", Kind: apRelative, Value: feltString("0")},
					{Name: "y.d0", Kind: apRelative, Value: feltString("0")},
					{Name: "y.d1", Kind: apRelative, Value: feltString("3618502788666127798953978732740734578953660990361066340291730267696802036752")},
					{Name: "y.d2", Kind: apRelative, Value: feltString("0")},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcNegateHint(ctx.operanders["x.d0"])
				},
				check: varValueInScopeEquals("value", "332307077013822705460080369276551168"),
			},
			{
				operanders: []*hintOperander{
					{Name: "x.d0", Kind: apRelative, Value: feltString("0")},
					{Name: "x.d1", Kind: apRelative, Value: feltString("0")},
					{Name: "x.d2", Kind: apRelative, Value: feltString("0")},
					{Name: "y.d0", Kind: apRelative, Value: feltString("0")},
					{Name: "y.d1", Kind: apRelative, Value: feltString("0")},
					{Name: "y.d2", Kind: apRelative, Value: feltString("3618502788666127798953978732740734578953660990361066340291730267696802036752")},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcNegateHint(ctx.operanders["x.d0"])
				},
				check: varValueInScopeEquals("value", "25711014748331348032841660844170547741622139443892033895268352"),
			},
			{
				operanders: []*hintOperander{
					{Name: "x.d0", Kind: apRelative, Value: feltString("0")},
					{Name: "x.d1", Kind: apRelative, Value: feltString("0")},
					{Name: "x.d2", Kind: apRelative, Value: feltString("0")},
					{Name: "y.d0", Kind: apRelative, Value: feltString("3618502788666127798953978732740734578953660990361066340291730267696802036752")},
					{Name: "y.d1", Kind: apRelative, Value: feltString("3618502788666127798953978732740734578953660990361066340291730267696802036752")},
					{Name: "y.d2", Kind: apRelative, Value: feltString("0")},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcNegateHint(ctx.operanders["x.d0"])
				},
				check: varValueInScopeEquals("value", "3414743344050354335526669778532047544182386821868808346534897"),
			},
			{
				operanders: []*hintOperander{
					{Name: "x.d0", Kind: apRelative, Value: feltString("0")},
					{Name: "x.d1", Kind: apRelative, Value: feltString("0")},
					{Name: "x.d2", Kind: apRelative, Value: feltString("0")},
					{Name: "y.d0", Kind: apRelative, Value: feltString("0")},
					{Name: "y.d1", Kind: apRelative, Value: feltString("3618502788666127798953978732740734578953660990361066340291730267696802036752")},
					{Name: "y.d2", Kind: apRelative, Value: feltString("3618502788666127798953978732740734578953660990361066340291730267696802036752")},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcNegateHint(ctx.operanders["x.d0"])
				},
				check: varValueInScopeEquals("value", "25711014748331348032841661176477624755444844903972403171819520"),
			},
			{
				operanders: []*hintOperander{
					{Name: "x.d0", Kind: apRelative, Value: feltString("0")},
					{Name: "x.d1", Kind: apRelative, Value: feltString("0")},
					{Name: "x.d2", Kind: apRelative, Value: feltString("0")},
					{Name: "y.d0", Kind: apRelative, Value: feltString("3618502788666127798953978732740734578953660990361066340291730267696802036752")},
					{Name: "y.d1", Kind: apRelative, Value: feltString("0")},
					{Name: "y.d2", Kind: apRelative, Value: feltString("3618502788666127798953978732740734578953660990361066340291730267696802036752")},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newEcNegateHint(ctx.operanders["x.d0"])
				},
				check: varValueInScopeEquals("value", "29125758092381702368368330290395518271981820805680472965252081"),
			},
			{
				operanders: []*hintOperander{
					{Name: "x.d0", Kind: apRelative, Value: feltString("0")},
					{Name: "x.d1", Kind: apRelative, Value: feltString("0")},
					{Name: "x.d2", Kind: apRelative, Value: feltString("0")},
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
				check: varValueInScopeEquals("value", "29125758092381702368368330622702595285804526265760842241803249"),
			},
			{
				operanders: []*hintOperander{
					{Name: "x.d0", Kind: apRelative, Value: feltString("0")},
					{Name: "x.d1", Kind: apRelative, Value: feltString("0")},
					{Name: "x.d2", Kind: apRelative, Value: feltString("0")},
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
				check: varValueInScopeEquals("value", "115792089237316195423511115915312127562362008772591693155831694873530722155557"),
			},
		},
	})
}
