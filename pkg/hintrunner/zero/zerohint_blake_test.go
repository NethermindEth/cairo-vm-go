package zero

import (
	"fmt"
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
					{Name: "N_PACKED_INSTANCES", Kind: fpRelative, Value: feltUint64(7)},
					{Name: "INPUT_BLOCK_FELTS", Kind: apRelative, Value: feltUint64(16)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sFinalizeHint(ctx.operanders["blake2s_ptr_end"], ctx.operanders["N_PACKED_INSTANCES"], ctx.operanders["INPUT_BLOCK_FELTS"])
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
		"Blake2sFinalize check assert N_PACKED_INSTANCES": {
			{
				operanders: []*hintOperander{
					{Name: "blake2s_ptr_end", Kind: apRelative, Value: addrWithSegment(1, 7)},
					{Name: "N_PACKED_INSTANCES", Kind: fpRelative, Value: feltUint64(20)},
					{Name: "INPUT_BLOCK_FELTS", Kind: apRelative, Value: feltUint64(16)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sFinalizeHint(ctx.operanders["blake2s_ptr_end"], ctx.operanders["N_PACKED_INSTANCES"], ctx.operanders["INPUT_BLOCK_FELTS"])
				},
				errCheck: errorTextContains(fmt.Sprintf("in range %s, got %s", "[0, 20)", "20")),
			},
		},
		"Blake2sFinalize check assert INPUT_BLOCK_FELTS": {
			{
				operanders: []*hintOperander{
					{Name: "blake2s_ptr_end", Kind: apRelative, Value: addrWithSegment(1, 7)},
					{Name: "N_PACKED_INSTANCES", Kind: fpRelative, Value: feltUint64(7)},
					{Name: "INPUT_BLOCK_FELTS", Kind: apRelative, Value: feltUint64(200)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlake2sFinalizeHint(ctx.operanders["blake2s_ptr_end"], ctx.operanders["N_PACKED_INSTANCES"], ctx.operanders["INPUT_BLOCK_FELTS"])
				},
				errCheck: errorTextContains(fmt.Sprintf("in range %s, got %s", "[0, 100)", "200")),
			},
		},
	})
}
