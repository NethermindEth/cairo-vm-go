package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func TestZeroHintBlake(t *testing.T) {
	runHinterTests(t, map[string][]hintTestCase{
		"Blake2sAddUint256Bigend": {
			{
				// 2**256 - 1
				operanders: []*hintOperander{
					{Name: "high", Kind: fpRelative, Value: feltString("10633823966279317261796329637309054975")},
					{Name: "low", Kind: fpRelative, Value: feltString("340282366920938463463374607431768211424")},
					{Name: "data", Kind: apRelative, Value: addr(7)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256BigendHint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"data",
					[]*fp.Element{
						feltString("134217727"),
						feltString("4294966768"),
						feltString("4294967295"),
						feltString("4294967295"),
						feltString("4294967295"),
						feltString("4294967295"),
						feltString("4294967295"),
						feltString("4294967264"),
					}),
			},
			{
				// 2**256
				operanders: []*hintOperander{
					{Name: "high", Kind: fpRelative, Value: feltString("10633823966279317261796329637309054975")},
					{Name: "low", Kind: fpRelative, Value: feltString("340282366920938463463374607431768211425")},
					{Name: "data", Kind: apRelative, Value: addr(7)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256BigendHint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"data",
					[]*fp.Element{
						feltString("134217727"),
						feltString("4294966768"),
						feltString("4294967295"),
						feltString("4294967295"),
						feltString("4294967295"),
						feltString("4294967295"),
						feltString("4294967295"),
						feltString("4294967265"),
					}),
			},
			{
				// 2**400
				operanders: []*hintOperander{
					{Name: "high", Kind: fpRelative, Value: feltString("10633629342298111006479807194589036544")},
					{Name: "low", Kind: fpRelative, Value: feltString("21044980667851464052662337537")},
					{Name: "data", Kind: apRelative, Value: addr(7)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256BigendHint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"data",
					[]*fp.Element{
						feltString("134215271"),
						feltString("2147483664"),
						feltString("4294967295"),
						feltString("4292870144"),
						feltString("0"),
						feltString("1140850687"),
						feltString("4294967151"),
						feltString("2147483649"),
					}),
			},
			{
				// 689
				operanders: []*hintOperander{
					{Name: "high", Kind: fpRelative, Value: feltString("0")},
					{Name: "low", Kind: fpRelative, Value: feltString("689")},
					{Name: "data", Kind: apRelative, Value: addr(7)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256BigendHint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"data",
					[]*fp.Element{
						feltString("0"),
						feltString("0"),
						feltString("0"),
						feltString("0"),
						feltString("0"),
						feltString("0"),
						feltString("0"),
						feltString("689"),
					}),
			},
			{
				// 2**128 - 1
				operanders: []*hintOperander{
					{Name: "high", Kind: fpRelative, Value: feltString("0")},
					{Name: "low", Kind: fpRelative, Value: feltString("340282366920938463463374607431768211455")},
					{Name: "data", Kind: apRelative, Value: addr(7)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256BigendHint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"data",
					[]*fp.Element{
						feltString("0"),
						feltString("0"),
						feltString("0"),
						feltString("0"),
						feltString("4294967295"),
						feltString("4294967295"),
						feltString("4294967295"),
						feltString("4294967295"),
					}),
			},
			{
				// 2**128
				operanders: []*hintOperander{
					{Name: "high", Kind: fpRelative, Value: feltString("1")},
					{Name: "low", Kind: fpRelative, Value: feltString("0")},
					{Name: "data", Kind: apRelative, Value: addr(7)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256BigendHint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"data",
					[]*fp.Element{
						feltString("0"),
						feltString("0"),
						feltString("0"),
						feltString("1"),
						feltString("0"),
						feltString("0"),
						feltString("0"),
						feltString("0"),
					}),
			},
			{
				// 0 or modulus()
				operanders: []*hintOperander{
					{Name: "high", Kind: fpRelative, Value: feltString("0")},
					{Name: "low", Kind: fpRelative, Value: feltString("0")},
					{Name: "data", Kind: apRelative, Value: addr(7)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256BigendHint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"data",
					[]*fp.Element{
						feltString("0"),
						feltString("0"),
						feltString("0"),
						feltString("0"),
						feltString("0"),
						feltString("0"),
						feltString("0"),
						feltString("0"),
					}),
			},
			{
				// modulus() - 1
				operanders: []*hintOperander{
					{Name: "high", Kind: fpRelative, Value: feltString("10633823966279327296825105735305134080")},
					{Name: "low", Kind: fpRelative, Value: feltString("0")},
					{Name: "data", Kind: apRelative, Value: addr(7)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256BigendHint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"data",
					[]*fp.Element{
						feltString("134217728"),
						feltString("17"),
						feltString("0"),
						feltString("0"),
						feltString("0"),
						feltString("0"),
						feltString("0"),
						feltString("0"),
					}),
			},
			{
				// modulus() + 1
				operanders: []*hintOperander{
					{Name: "high", Kind: fpRelative, Value: feltString("0")},
					{Name: "low", Kind: fpRelative, Value: feltString("1")},
					{Name: "data", Kind: apRelative, Value: addr(7)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256BigendHint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"data",
					[]*fp.Element{
						feltString("0"),
						feltString("0"),
						feltString("0"),
						feltString("0"),
						feltString("0"),
						feltString("0"),
						feltString("0"),
						feltString("1"),
					}),
			},
		},
	})
}
