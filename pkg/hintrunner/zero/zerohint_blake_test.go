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
					return newBlake2sAddUint256Hint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"], true)
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
					return newBlake2sAddUint256Hint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"], true)
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
					return newBlake2sAddUint256Hint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"], true)
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
					return newBlake2sAddUint256Hint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"], true)
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
					return newBlake2sAddUint256Hint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"], true)
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
					return newBlake2sAddUint256Hint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"], true)
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
					return newBlake2sAddUint256Hint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"], true)
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
					return newBlake2sAddUint256Hint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"], true)
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
					return newBlake2sAddUint256Hint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"], true)
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
		"Blake2sAddUint256": {
			{
				// 2**256 - 1
				operanders: []*hintOperander{
					{Name: "high", Kind: fpRelative, Value: feltString("10633823966279317261796329637309054975")},
					{Name: "low", Kind: fpRelative, Value: feltString("340282366920938463463374607431768211424")},
					{Name: "data", Kind: apRelative, Value: addr(7)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256Hint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"], false)
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
					{Name: "data", Kind: apRelative, Value: addr(7)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256Hint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"], false)
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
					{Name: "data", Kind: apRelative, Value: addr(7)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256Hint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"], false)
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
					{Name: "data", Kind: apRelative, Value: addr(7)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256Hint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"], false)
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
					{Name: "data", Kind: apRelative, Value: addr(7)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256Hint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"], false)
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
					{Name: "data", Kind: apRelative, Value: addr(7)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256Hint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"], false)
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
					{Name: "data", Kind: apRelative, Value: addr(7)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256Hint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"], false)
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
					return newBlake2sAddUint256Hint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"], false)
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
					{Name: "data", Kind: apRelative, Value: addr(7)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sAddUint256Hint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["data"], false)
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
		"Blake2sFinalize": {
			{
				operanders: []*hintOperander{
					{Name: "blake2s_ptr_end", Kind: apRelative, Value: addrWithSegment(1, 7)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sFinalizeHint(ctx.operanders["blake2s_ptr_end"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"blake2s_ptr_end",
					[]*fp.Element{
						feltUint64(1795745351), feltUint64(3144134277), feltUint64(1013904242), feltUint64(2773480762),
						feltUint64(1359893119), feltUint64(2600822924), feltUint64(528734635), feltUint64(1541459225),
						feltUint64(0), feltUint64(0), feltUint64(0), feltUint64(0),
						feltUint64(0), feltUint64(0), feltUint64(0), feltUint64(0),
						feltUint64(0), feltUint64(0), feltUint64(0), feltUint64(0),
						feltUint64(0), feltUint64(0), feltUint64(0), feltUint64(0),
						feltUint64(0), feltUint64(4294967295), feltUint64(813310313), feltUint64(2491453561),
						feltUint64(3491828193), feltUint64(2085238082), feltUint64(1219908895), feltUint64(514171180),
						feltUint64(4245497115), feltUint64(4193177630), feltUint64(1795745351), feltUint64(3144134277),
						feltUint64(1013904242), feltUint64(2773480762), feltUint64(1359893119), feltUint64(2600822924),
						feltUint64(528734635), feltUint64(1541459225), feltUint64(0), feltUint64(0),
						feltUint64(0), feltUint64(0), feltUint64(0), feltUint64(0),
						feltUint64(0), feltUint64(0), feltUint64(0), feltUint64(0),
						feltUint64(0), feltUint64(0), feltUint64(0), feltUint64(0),
						feltUint64(0), feltUint64(0), feltUint64(0), feltUint64(4294967295),
						feltUint64(813310313), feltUint64(2491453561), feltUint64(3491828193), feltUint64(2085238082),
						feltUint64(1219908895), feltUint64(514171180), feltUint64(4245497115), feltUint64(4193177630),
						feltUint64(1795745351), feltUint64(3144134277), feltUint64(1013904242), feltUint64(2773480762),
						feltUint64(1359893119), feltUint64(2600822924), feltUint64(528734635), feltUint64(1541459225),
						feltUint64(0), feltUint64(0), feltUint64(0), feltUint64(0),
						feltUint64(0), feltUint64(0), feltUint64(0), feltUint64(0),
						feltUint64(0), feltUint64(0), feltUint64(0), feltUint64(0),
						feltUint64(0), feltUint64(0), feltUint64(0), feltUint64(0),
						feltUint64(0), feltUint64(4294967295), feltUint64(813310313), feltUint64(2491453561),
						feltUint64(3491828193), feltUint64(2085238082), feltUint64(1219908895), feltUint64(514171180),
						feltUint64(4245497115), feltUint64(4193177630), feltUint64(1795745351), feltUint64(3144134277),
						feltUint64(1013904242), feltUint64(2773480762), feltUint64(1359893119), feltUint64(2600822924),
						feltUint64(528734635), feltUint64(1541459225), feltUint64(0), feltUint64(0),
						feltUint64(0), feltUint64(0), feltUint64(0), feltUint64(0),
						feltUint64(0), feltUint64(0), feltUint64(0), feltUint64(0),
						feltUint64(0), feltUint64(0), feltUint64(0), feltUint64(0),
						feltUint64(0), feltUint64(0), feltUint64(0), feltUint64(4294967295),
						feltUint64(813310313), feltUint64(2491453561), feltUint64(3491828193), feltUint64(2085238082),
						feltUint64(1219908895), feltUint64(514171180), feltUint64(4245497115), feltUint64(4193177630),
						feltUint64(1795745351), feltUint64(3144134277), feltUint64(1013904242), feltUint64(2773480762),
						feltUint64(1359893119), feltUint64(2600822924), feltUint64(528734635), feltUint64(1541459225),
						feltUint64(0), feltUint64(0), feltUint64(0), feltUint64(0),
						feltUint64(0), feltUint64(0), feltUint64(0), feltUint64(0),
						feltUint64(0), feltUint64(0), feltUint64(0), feltUint64(0),
						feltUint64(0), feltUint64(0), feltUint64(0), feltUint64(0),
						feltUint64(0), feltUint64(4294967295), feltUint64(813310313), feltUint64(2491453561),
						feltUint64(3491828193), feltUint64(2085238082), feltUint64(1219908895), feltUint64(514171180),
						feltUint64(4245497115), feltUint64(4193177630), feltUint64(1795745351), feltUint64(3144134277),
						feltUint64(1013904242), feltUint64(2773480762), feltUint64(1359893119), feltUint64(2600822924),
						feltUint64(528734635), feltUint64(1541459225), feltUint64(0), feltUint64(0),
						feltUint64(0), feltUint64(0), feltUint64(0), feltUint64(0),
						feltUint64(0), feltUint64(0), feltUint64(0), feltUint64(0),
						feltUint64(0), feltUint64(0), feltUint64(0), feltUint64(0),
						feltUint64(0), feltUint64(0), feltUint64(0), feltUint64(4294967295),
						feltUint64(813310313), feltUint64(2491453561), feltUint64(3491828193), feltUint64(2085238082),
						feltUint64(1219908895), feltUint64(514171180), feltUint64(4245497115), feltUint64(4193177630),
					}),
			},
		},
		"Blake2sCompute": {
			{
				operanders: []*hintOperander{
					{Name: "output", Kind: apRelative, Value: addrWithSegment(1, 31)},
					{Name: "h.1", Kind: apRelative, Value: feltUint64(1)},
					{Name: "h.2", Kind: apRelative, Value: feltUint64(2)},
					{Name: "h.3", Kind: apRelative, Value: feltUint64(3)},
					{Name: "h.4", Kind: apRelative, Value: feltUint64(4)},
					{Name: "h.5", Kind: apRelative, Value: feltUint64(5)},
					{Name: "h.6", Kind: apRelative, Value: feltUint64(6)},
					{Name: "h.7", Kind: apRelative, Value: feltUint64(7)},
					{Name: "h.8", Kind: apRelative, Value: feltUint64(8)},
					{Name: "message.1", Kind: apRelative, Value: feltUint64(9)},
					{Name: "message.2", Kind: apRelative, Value: feltUint64(10)},
					{Name: "message.3", Kind: apRelative, Value: feltUint64(11)},
					{Name: "message.4", Kind: apRelative, Value: feltUint64(12)},
					{Name: "message.5", Kind: apRelative, Value: feltUint64(13)},
					{Name: "message.6", Kind: apRelative, Value: feltUint64(14)},
					{Name: "message.7", Kind: apRelative, Value: feltUint64(15)},
					{Name: "message.8", Kind: apRelative, Value: feltUint64(16)},
					{Name: "message.9", Kind: apRelative, Value: feltUint64(17)},
					{Name: "message.10", Kind: apRelative, Value: feltUint64(18)},
					{Name: "message.11", Kind: apRelative, Value: feltUint64(19)},
					{Name: "message.12", Kind: apRelative, Value: feltUint64(20)},
					{Name: "message.13", Kind: apRelative, Value: feltUint64(21)},
					{Name: "message.14", Kind: apRelative, Value: feltUint64(22)},
					{Name: "message.15", Kind: apRelative, Value: feltUint64(23)},
					{Name: "message.16", Kind: apRelative, Value: feltUint64(24)},
					{Name: "t", Kind: apRelative, Value: feltUint64(25)},
					{Name: "f", Kind: apRelative, Value: feltUint64(26)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sComputeHint(ctx.operanders["output"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"output",
					[]*fp.Element{
						feltUint64(1503208424),
						feltUint64(3786571270),
						feltUint64(625865791),
						feltUint64(657700341),
						feltUint64(3174522044),
						feltUint64(3976146666),
						feltUint64(3581823059),
						feltUint64(2049603206),
					}),
			},
		},
		"Blake2sCompress": {
			{
				operanders: []*hintOperander{
					{Name: "output", Kind: apRelative, Value: addrWithSegment(3, 25)},
					{Name: "n_bytes", Kind: apRelative, Value: addr(7)},
					{Name: "blake2s_start", Kind: apRelative, Value: addr(7)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sCompressHint(ctx.operanders["n_bytes"], ctx.operanders["output"], ctx.operanders["blake2s_start"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"output",
					[]*fp.Element{
						feltUint64(1503208424),
						feltUint64(3786571270),
						feltUint64(625865791),
						feltUint64(657700341),
						feltUint64(3174522044),
						feltUint64(3976146666),
						feltUint64(3581823059),
						feltUint64(2049603206),
					}),
			},
		},
	})
}
