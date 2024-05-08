package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func TestZeroHintKeccak(t *testing.T) {
	runHinterTests(t, map[string][]hintTestCase{
		"KeccakWriteArgs": {
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
					{Name: "low", Kind: fpRelative, Value: feltString("18446744073709551618")},
					{Name: "high", Kind: fpRelative, Value: feltString("18446744073709551618")},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newKeccakWriteArgsHint(ctx.operanders["inputs"], ctx.operanders["low"], ctx.operanders["high"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"inputs",
					[]*fp.Element{
						feltUint64(2),
						feltUint64(1),
						feltUint64(2),
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
			{
				operanders: []*hintOperander{
					{Name: "inputs", Kind: apRelative, Value: addr(7)},
					{Name: "low", Kind: fpRelative, Value: feltString("340282366920938463463374607431768211455")},
					{Name: "high", Kind: fpRelative, Value: feltString("18446744073709551626")},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newKeccakWriteArgsHint(ctx.operanders["inputs"], ctx.operanders["low"], ctx.operanders["high"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"inputs",
					[]*fp.Element{
						feltUint64(18446744073709551615),
						feltUint64(18446744073709551615),
						feltUint64(10),
						feltUint64(1),
					}),
			},
			{
				operanders: []*hintOperander{
					{Name: "inputs", Kind: apRelative, Value: addr(7)},
					{Name: "low", Kind: fpRelative, Value: feltString("368934881474191032340")},
					{Name: "high", Kind: fpRelative, Value: feltString("184467440737095516170")},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newKeccakWriteArgsHint(ctx.operanders["inputs"], ctx.operanders["low"], ctx.operanders["high"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"inputs",
					[]*fp.Element{
						feltUint64(20),
						feltUint64(20),
						feltUint64(10),
						feltUint64(10),
					}),
			},
		},
		"BlockPermutation": {
			{
				operanders: []*hintOperander{
					{Name: "KECCAK_STATE_SIZE_FELTS", Kind: apRelative, Value: feltUint64(25)},
					{Name: "keccak_ptr", Kind: fpRelative, Value: addr(31)},
					{Name: "data.0", Kind: apRelative, Value: feltUint64(1)},
					{Name: "data.1", Kind: apRelative, Value: feltUint64(2)},
					{Name: "data.2", Kind: apRelative, Value: feltUint64(3)},
					{Name: "data.3", Kind: apRelative, Value: feltUint64(4)},
					{Name: "data.4", Kind: apRelative, Value: feltUint64(5)},
					{Name: "data.5", Kind: apRelative, Value: feltUint64(6)},
					{Name: "data.6", Kind: apRelative, Value: feltUint64(7)},
					{Name: "data.7", Kind: apRelative, Value: feltUint64(8)},
					{Name: "data.8", Kind: apRelative, Value: feltUint64(9)},
					{Name: "data.9", Kind: apRelative, Value: feltUint64(10)},
					{Name: "data.10", Kind: apRelative, Value: feltUint64(11)},
					{Name: "data.11", Kind: apRelative, Value: feltUint64(12)},
					{Name: "data.12", Kind: apRelative, Value: feltUint64(13)},
					{Name: "data.13", Kind: apRelative, Value: feltUint64(14)},
					{Name: "data.14", Kind: apRelative, Value: feltUint64(15)},
					{Name: "data.15", Kind: apRelative, Value: feltUint64(16)},
					{Name: "data.16", Kind: apRelative, Value: feltUint64(17)},
					{Name: "data.17", Kind: apRelative, Value: feltUint64(18)},
					{Name: "data.18", Kind: apRelative, Value: feltUint64(19)},
					{Name: "data.19", Kind: apRelative, Value: feltUint64(20)},
					{Name: "data.20", Kind: apRelative, Value: feltUint64(21)},
					{Name: "data.21", Kind: apRelative, Value: feltUint64(22)},
					{Name: "data.22", Kind: apRelative, Value: feltUint64(23)},
					{Name: "data.23", Kind: apRelative, Value: feltUint64(24)},
					{Name: "data.24", Kind: apRelative, Value: feltUint64(25)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlockPermutationHint(ctx.operanders["KECCAK_STATE_SIZE_FELTS"], ctx.operanders["keccak_ptr"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {},
			},
		},
	})
}
