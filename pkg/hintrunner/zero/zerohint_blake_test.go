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
					{Name: "data1", Kind: apRelative, Value: addr(100)},
					{Name: "data2", Kind: apRelative, Value: addr(101)},
					{Name: "data3", Kind: apRelative, Value: addr(102)},
					{Name: "data4", Kind: apRelative, Value: addr(103)},
					{Name: "data5", Kind: apRelative, Value: addr(104)},
					{Name: "data6", Kind: apRelative, Value: addr(105)},
					{Name: "data7", Kind: apRelative, Value: addr(106)},
					{Name: "data8", Kind: apRelative, Value: addr(107)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256BigendHint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data1"])
				},
				check: allVarAddrResolvedValueEquals(map[string]*fp.Element{
					"data1": feltString("134217727"),
					"data2": feltString("4294966768"),
					"data3": feltString("4294967295"),
					"data4": feltString("4294967295"),
					"data5": feltString("4294967295"),
					"data6": feltString("4294967295"),
					"data7": feltString("4294967295"),
					"data8": feltString("4294967264"),
				}),
			},
			{
				// 2**256
				operanders: []*hintOperander{
					{Name: "high", Kind: fpRelative, Value: feltString("10633823966279317261796329637309054975")},
					{Name: "low", Kind: fpRelative, Value: feltString("340282366920938463463374607431768211425")},
					{Name: "data1", Kind: apRelative, Value: addr(100)},
					{Name: "data2", Kind: apRelative, Value: addr(101)},
					{Name: "data3", Kind: apRelative, Value: addr(102)},
					{Name: "data4", Kind: apRelative, Value: addr(103)},
					{Name: "data5", Kind: apRelative, Value: addr(104)},
					{Name: "data6", Kind: apRelative, Value: addr(105)},
					{Name: "data7", Kind: apRelative, Value: addr(106)},
					{Name: "data8", Kind: apRelative, Value: addr(107)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256BigendHint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data1"])
				},
				check: allVarAddrResolvedValueEquals(map[string]*fp.Element{
					"data1": feltString("134217727"),
					"data2": feltString("4294966768"),
					"data3": feltString("4294967295"),
					"data4": feltString("4294967295"),
					"data5": feltString("4294967295"),
					"data6": feltString("4294967295"),
					"data7": feltString("4294967295"),
					"data8": feltString("4294967265"),
				}),
			},
			{
				// 2**400
				operanders: []*hintOperander{
					{Name: "high", Kind: fpRelative, Value: feltString("10633629342298111006479807194589036544")},
					{Name: "low", Kind: fpRelative, Value: feltString("21044980667851464052662337537")},
					{Name: "data1", Kind: apRelative, Value: addr(100)},
					{Name: "data2", Kind: apRelative, Value: addr(101)},
					{Name: "data3", Kind: apRelative, Value: addr(102)},
					{Name: "data4", Kind: apRelative, Value: addr(103)},
					{Name: "data5", Kind: apRelative, Value: addr(104)},
					{Name: "data6", Kind: apRelative, Value: addr(105)},
					{Name: "data7", Kind: apRelative, Value: addr(106)},
					{Name: "data8", Kind: apRelative, Value: addr(107)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256BigendHint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data1"])
				},
				check: allVarAddrResolvedValueEquals(map[string]*fp.Element{
					"data1": feltString("134215271"),
					"data2": feltString("2147483664"),
					"data3": feltString("4294967295"),
					"data4": feltString("4292870144"),
					"data5": feltString("0"),
					"data6": feltString("1140850687"),
					"data7": feltString("4294967151"),
					"data8": feltString("2147483649"),
				}),
			},
			{
				// 689
				operanders: []*hintOperander{
					{Name: "high", Kind: fpRelative, Value: feltString("0")},
					{Name: "low", Kind: fpRelative, Value: feltString("689")},
					{Name: "data1", Kind: apRelative, Value: addr(100)},
					{Name: "data2", Kind: apRelative, Value: addr(101)},
					{Name: "data3", Kind: apRelative, Value: addr(102)},
					{Name: "data4", Kind: apRelative, Value: addr(103)},
					{Name: "data5", Kind: apRelative, Value: addr(104)},
					{Name: "data6", Kind: apRelative, Value: addr(105)},
					{Name: "data7", Kind: apRelative, Value: addr(106)},
					{Name: "data8", Kind: apRelative, Value: addr(107)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256BigendHint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data1"])
				},
				check: allVarAddrResolvedValueEquals(map[string]*fp.Element{
					"data1": feltString("0"),
					"data2": feltString("0"),
					"data3": feltString("0"),
					"data4": feltString("0"),
					"data5": feltString("0"),
					"data6": feltString("0"),
					"data7": feltString("0"),
					"data8": feltString("689"),
				}),
			},
			{
				// 2**128 - 1
				operanders: []*hintOperander{
					{Name: "high", Kind: fpRelative, Value: feltString("0")},
					{Name: "low", Kind: fpRelative, Value: feltString("340282366920938463463374607431768211455")},
					{Name: "data1", Kind: apRelative, Value: addr(100)},
					{Name: "data2", Kind: apRelative, Value: addr(101)},
					{Name: "data3", Kind: apRelative, Value: addr(102)},
					{Name: "data4", Kind: apRelative, Value: addr(103)},
					{Name: "data5", Kind: apRelative, Value: addr(104)},
					{Name: "data6", Kind: apRelative, Value: addr(105)},
					{Name: "data7", Kind: apRelative, Value: addr(106)},
					{Name: "data8", Kind: apRelative, Value: addr(107)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256BigendHint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data1"])
				},
				check: allVarAddrResolvedValueEquals(map[string]*fp.Element{
					"data1": feltString("0"),
					"data2": feltString("0"),
					"data3": feltString("0"),
					"data4": feltString("0"),
					"data5": feltString("4294967295"),
					"data6": feltString("4294967295"),
					"data7": feltString("4294967295"),
					"data8": feltString("4294967295"),
				}),
			},
			{
				// 2**128
				operanders: []*hintOperander{
					{Name: "high", Kind: fpRelative, Value: feltString("1")},
					{Name: "low", Kind: fpRelative, Value: feltString("0")},
					{Name: "data1", Kind: apRelative, Value: addr(100)},
					{Name: "data2", Kind: apRelative, Value: addr(101)},
					{Name: "data3", Kind: apRelative, Value: addr(102)},
					{Name: "data4", Kind: apRelative, Value: addr(103)},
					{Name: "data5", Kind: apRelative, Value: addr(104)},
					{Name: "data6", Kind: apRelative, Value: addr(105)},
					{Name: "data7", Kind: apRelative, Value: addr(106)},
					{Name: "data8", Kind: apRelative, Value: addr(107)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256BigendHint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data1"])
				},
				check: allVarAddrResolvedValueEquals(map[string]*fp.Element{
					"data1": feltString("0"),
					"data2": feltString("0"),
					"data3": feltString("0"),
					"data4": feltString("1"),
					"data5": feltString("0"),
					"data6": feltString("0"),
					"data7": feltString("0"),
					"data8": feltString("0"),
				}),
			},
			{
				// 0 or modulus()
				operanders: []*hintOperander{
					{Name: "high", Kind: fpRelative, Value: feltString("0")},
					{Name: "low", Kind: fpRelative, Value: feltString("0")},
					{Name: "data1", Kind: apRelative, Value: addr(100)},
					{Name: "data2", Kind: apRelative, Value: addr(101)},
					{Name: "data3", Kind: apRelative, Value: addr(102)},
					{Name: "data4", Kind: apRelative, Value: addr(103)},
					{Name: "data5", Kind: apRelative, Value: addr(104)},
					{Name: "data6", Kind: apRelative, Value: addr(105)},
					{Name: "data7", Kind: apRelative, Value: addr(106)},
					{Name: "data8", Kind: apRelative, Value: addr(107)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256BigendHint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data1"])
				},
				check: allVarAddrResolvedValueEquals(map[string]*fp.Element{
					"data1": feltString("0"),
					"data2": feltString("0"),
					"data3": feltString("0"),
					"data4": feltString("0"),
					"data5": feltString("0"),
					"data6": feltString("0"),
					"data7": feltString("0"),
					"data8": feltString("0"),
				}),
			},
			{
				// modulus() - 1
				operanders: []*hintOperander{
					{Name: "high", Kind: fpRelative, Value: feltString("10633823966279327296825105735305134080")},
					{Name: "low", Kind: fpRelative, Value: feltString("0")},
					{Name: "data1", Kind: apRelative, Value: addr(100)},
					{Name: "data2", Kind: apRelative, Value: addr(101)},
					{Name: "data3", Kind: apRelative, Value: addr(102)},
					{Name: "data4", Kind: apRelative, Value: addr(103)},
					{Name: "data5", Kind: apRelative, Value: addr(104)},
					{Name: "data6", Kind: apRelative, Value: addr(105)},
					{Name: "data7", Kind: apRelative, Value: addr(106)},
					{Name: "data8", Kind: apRelative, Value: addr(107)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256BigendHint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data1"])
				},
				check: allVarAddrResolvedValueEquals(map[string]*fp.Element{
					"data1": feltString("134217728"),
					"data2": feltString("17"),
					"data3": feltString("0"),
					"data4": feltString("0"),
					"data5": feltString("0"),
					"data6": feltString("0"),
					"data7": feltString("0"),
					"data8": feltString("0"),
				}),
			},
			{
				// modulus() + 1
				operanders: []*hintOperander{
					{Name: "high", Kind: fpRelative, Value: feltString("0")},
					{Name: "low", Kind: fpRelative, Value: feltString("1")},
					{Name: "data1", Kind: apRelative, Value: addr(100)},
					{Name: "data2", Kind: apRelative, Value: addr(101)},
					{Name: "data3", Kind: apRelative, Value: addr(102)},
					{Name: "data4", Kind: apRelative, Value: addr(103)},
					{Name: "data5", Kind: apRelative, Value: addr(104)},
					{Name: "data6", Kind: apRelative, Value: addr(105)},
					{Name: "data7", Kind: apRelative, Value: addr(106)},
					{Name: "data8", Kind: apRelative, Value: addr(107)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256BigendHint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data1"])
				},
				check: allVarAddrResolvedValueEquals(map[string]*fp.Element{
					"data1": feltString("0"),
					"data2": feltString("0"),
					"data3": feltString("0"),
					"data4": feltString("0"),
					"data5": feltString("0"),
					"data6": feltString("0"),
					"data7": feltString("0"),
					"data8": feltString("1"),
				}),
			},
		},

		"Blake2sAddUint256": {
			{
				// 2**256 - 1
				operanders: []*hintOperander{
					{Name: "high", Kind: fpRelative, Value: feltString("10633823966279317261796329637309054975")},
					{Name: "low", Kind: fpRelative, Value: feltString("340282366920938463463374607431768211424")},
					{Name: "data", Kind: apRelative, Value: addr(50)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256Hint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"data",
					[]*fp.Element{
						feltString("4294967264"),
						feltString("4294967295"),
						feltString("4294967295"),
						feltString("4294967295"),
						feltString("4294967295"),
						feltString("4294967295"),
						feltString("4294966768"),
						feltString("134217727"),
					}),
			},
			{
				// 2**256
				operanders: []*hintOperander{
					{Name: "high", Kind: fpRelative, Value: feltString("10633823966279317261796329637309054975")},
					{Name: "low", Kind: fpRelative, Value: feltString("340282366920938463463374607431768211425")},
					{Name: "data", Kind: apRelative, Value: addr(50)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256Hint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"data",
					[]*fp.Element{
						feltString("4294967265"),
						feltString("4294967295"),
						feltString("4294967295"),
						feltString("4294967295"),
						feltString("4294967295"),
						feltString("4294967295"),
						feltString("4294966768"),
						feltString("134217727"),
					}),
			},
			{
				// 2**400
				operanders: []*hintOperander{
					{Name: "high", Kind: fpRelative, Value: feltString("10633629342298111006479807194589036544")},
					{Name: "low", Kind: fpRelative, Value: feltString("21044980667851464052662337537")},
					{Name: "data", Kind: apRelative, Value: addr(50)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256Hint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"data",
					[]*fp.Element{
						feltString("2147483649"),
						feltString("4294967151"),
						feltString("1140850687"),
						feltString("0"),
						feltString("4292870144"),
						feltString("4294967295"),
						feltString("2147483664"),
						feltString("134215271"),
					}),
			},
			{
				// 689
				operanders: []*hintOperander{
					{Name: "high", Kind: fpRelative, Value: feltString("0")},
					{Name: "low", Kind: fpRelative, Value: feltString("689")},
					{Name: "data", Kind: apRelative, Value: addr(50)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256Hint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"data",
					[]*fp.Element{
						feltString("689"),
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
				// 2**128 - 1
				operanders: []*hintOperander{
					{Name: "high", Kind: fpRelative, Value: feltString("0")},
					{Name: "low", Kind: fpRelative, Value: feltString("340282366920938463463374607431768211455")},
					{Name: "data", Kind: apRelative, Value: addr(50)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256Hint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"data",
					[]*fp.Element{
						feltString("4294967295"),
						feltString("4294967295"),
						feltString("4294967295"),
						feltString("4294967295"),
						feltString("0"),
						feltString("0"),
						feltString("0"),
						feltString("0"),
					}),
			},
			{
				// 2**128
				operanders: []*hintOperander{
					{Name: "high", Kind: fpRelative, Value: feltString("1")},
					{Name: "low", Kind: fpRelative, Value: feltString("0")},
					{Name: "data", Kind: apRelative, Value: addr(50)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256Hint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"data",
					[]*fp.Element{
						feltString("0"),
						feltString("0"),
						feltString("0"),
						feltString("0"),
						feltString("1"),
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
					{Name: "data", Kind: apRelative, Value: addr(50)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256Hint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"])
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
					{Name: "data", Kind: apRelative, Value: addr(50)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256Hint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"])
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
						feltString("17"),
						feltString("134217728"),
					}),
			},
			{
				// modulus() + 1
				operanders: []*hintOperander{
					{Name: "high", Kind: fpRelative, Value: feltString("0")},
					{Name: "low", Kind: fpRelative, Value: feltString("1")},
					{Name: "data", Kind: apRelative, Value: addr(50)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256Hint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"data",
					[]*fp.Element{
						feltString("1"),
						feltString("0"),
						feltString("0"),
						feltString("0"),
						feltString("0"),
						feltString("0"),
						feltString("0"),
						feltString("0"),
					}),
			},
		},
	})
}
