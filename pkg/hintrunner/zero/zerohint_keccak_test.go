package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func TestZeroHintKeccak(t *testing.T) {
	runHinterTests(t, map[string][]hintTestCase{
		"newKeccakWriteArgs": {
			{
				operanders: []*hintOperander{
					{Name: "inputs", Kind: apRelative, Value: addr(7)},
					{Name: "low", Kind: fpRelative, Value: feltUint64(0)},
					{Name: "high", Kind: fpRelative, Value: feltUint64(0)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newKeccakWriteArgsHint(ctx.operanders["inputs"], ctx.operanders["low"], ctx.operanders["high"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"inputs",
					[]*fp.Element{
						feltString("0"),
						feltString("0"),
						feltString("0"),
						feltString("0"),
					}),
			},
			{
				operanders: []*hintOperander{
					{Name: "inputs", Kind: apRelative, Value: addr(7)},
					{Name: "low", Kind: fpRelative, Value: feltUint64(1)},
					{Name: "high", Kind: fpRelative, Value: feltUint64(1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newKeccakWriteArgsHint(ctx.operanders["inputs"], ctx.operanders["low"], ctx.operanders["high"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"inputs",
					[]*fp.Element{
						feltString("1"),
						feltString("0"),
						feltString("1"),
						feltString("0"),
					}),
			},
			{
				operanders: []*hintOperander{
					{Name: "inputs", Kind: apRelative, Value: addr(7)},
					{Name: "low", Kind: fpRelative, Value: feltUint64(1)},
					{Name: "high", Kind: fpRelative, Value: feltUint64(0)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newKeccakWriteArgsHint(ctx.operanders["inputs"], ctx.operanders["low"], ctx.operanders["high"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"inputs",
					[]*fp.Element{
						feltString("1"),
						feltString("0"),
						feltString("0"),
						feltString("0"),
					}),
			},
			{
				operanders: []*hintOperander{
					{Name: "inputs", Kind: apRelative, Value: addr(7)},
					{Name: "low", Kind: fpRelative, Value: feltUint64(0)},
					{Name: "high", Kind: fpRelative, Value: feltUint64(1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newKeccakWriteArgsHint(ctx.operanders["inputs"], ctx.operanders["low"], ctx.operanders["high"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"inputs",
					[]*fp.Element{
						feltString("0"),
						feltString("0"),
						feltString("1"),
						feltString("0"),
					}),
			},
			{
				operanders: []*hintOperander{
					{Name: "inputs", Kind: apRelative, Value: addr(7)},
					{Name: "low", Kind: fpRelative, Value: feltUint64(10)},
					{Name: "high", Kind: fpRelative, Value: feltUint64(10)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newKeccakWriteArgsHint(ctx.operanders["inputs"], ctx.operanders["low"], ctx.operanders["high"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"inputs",
					[]*fp.Element{
						feltString("10"),
						feltString("0"),
						feltString("10"),
						feltString("0"),
					}),
			},
			{
				operanders: []*hintOperander{
					{Name: "inputs", Kind: apRelative, Value: addr(7)},
					{Name: "low", Kind: fpRelative, Value: feltString("18446744073709551615")},
					{Name: "high", Kind: fpRelative, Value: feltString("18446744073709551615")},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newKeccakWriteArgsHint(ctx.operanders["inputs"], ctx.operanders["low"], ctx.operanders["high"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"inputs",
					[]*fp.Element{
						feltUint64(18446744073709551615),
						feltUint64(0),
						feltUint64(18446744073709551615),
						feltUint64(0),
					}),
			},
			{
				operanders: []*hintOperander{
					{Name: "inputs", Kind: apRelative, Value: addr(7)},
					{Name: "low", Kind: fpRelative, Value: feltString("18446744073709551616")},
					{Name: "high", Kind: fpRelative, Value: feltString("18446744073709551616")},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newKeccakWriteArgsHint(ctx.operanders["inputs"], ctx.operanders["low"], ctx.operanders["high"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"inputs",
					[]*fp.Element{
						feltUint64(0),
						feltUint64(1),
						feltUint64(0),
						feltUint64(1),
					}),
			},
			{
				operanders: []*hintOperander{
					{Name: "inputs", Kind: apRelative, Value: addr(7)},
					{Name: "low", Kind: fpRelative, Value: feltString("18446744073709551617")},
					{Name: "high", Kind: fpRelative, Value: feltString("18446744073709551617")},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newKeccakWriteArgsHint(ctx.operanders["inputs"], ctx.operanders["low"], ctx.operanders["high"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"inputs",
					[]*fp.Element{
						feltUint64(1),
						feltUint64(1),
						feltUint64(1),
						feltUint64(1),
					}),
			},
			{
				operanders: []*hintOperander{
					{Name: "inputs", Kind: apRelative, Value: addr(7)},
					{Name: "low", Kind: fpRelative, Value: feltString("18446744073709551617")},
					{Name: "high", Kind: fpRelative, Value: feltString("18446744073709551617")},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newKeccakWriteArgsHint(ctx.operanders["inputs"], ctx.operanders["low"], ctx.operanders["high"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"inputs",
					[]*fp.Element{
						feltUint64(1),
						feltUint64(1),
						feltUint64(1),
						feltUint64(1),
					}),
			},
			{
				operanders: []*hintOperander{
					{Name: "inputs", Kind: apRelative, Value: addr(7)},
					{Name: "low", Kind: fpRelative, Value: feltString("340282366920938463463374607431768211455")},
					{Name: "high", Kind: fpRelative, Value: feltString("340282366920938463463374607431768211455")},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newKeccakWriteArgsHint(ctx.operanders["inputs"], ctx.operanders["low"], ctx.operanders["high"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"inputs",
					[]*fp.Element{
						feltUint64(18446744073709551615),
						feltUint64(18446744073709551615),
						feltUint64(18446744073709551615),
						feltUint64(18446744073709551615),
					}),
			},
		},
	})
}
