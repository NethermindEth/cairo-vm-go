package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func TestZeroHintKeccak(t *testing.T) {
	runHinterTests(t, map[string][]hintTestCase{
		"newCairoKeccakFinalize": {
			{
				operanders: []*hintOperander{
					{Name: "KECCAK_STATE_SIZE_FELTS", Kind: apRelative, Value: feltUint64(100)},
					{Name: "BLOCK_SIZE", Kind: uninitialized},
					{Name: "keccak_ptr_end", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newCairoKeccakFinalizeHint(ctx.operanders["KECCAK_STATE_SIZE_FELTS"], ctx.operanders["BLOCK_SIZE"], ctx.operanders["keccak_ptr_end"])
				},
				errCheck: errorTextContains("assert 0 <= _keccak_state_size_felts < 100."),
			},
			{
				operanders: []*hintOperander{
					{Name: "KECCAK_STATE_SIZE_FELTS", Kind: apRelative, Value: feltUint64(10)},
					{Name: "BLOCK_SIZE", Kind: apRelative, Value: feltUint64(11)},
					{Name: "keccak_ptr_end", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newCairoKeccakFinalizeHint(ctx.operanders["KECCAK_STATE_SIZE_FELTS"], ctx.operanders["BLOCK_SIZE"], ctx.operanders["keccak_ptr_end"])
				},
				errCheck: errorTextContains("assert 0 <= _block_size < 10."),
			},
			{
				operanders: []*hintOperander{
					{Name: "KECCAK_STATE_SIZE_FELTS", Kind: apRelative, Value: feltUint64(30)},
					{Name: "BLOCK_SIZE", Kind: apRelative, Value: feltUint64(2)},
					{Name: "keccak_ptr_end", Kind: apRelative, Value: addr(10)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newCairoKeccakFinalizeHint(ctx.operanders["KECCAK_STATE_SIZE_FELTS"], ctx.operanders["BLOCK_SIZE"], ctx.operanders["keccak_ptr_end"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					testValues := []uint64{
						0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 17376452488221285863, 9571781953733019530, 15391093639620504046, 13624874521033984333, 10027350355371872343, 18417369716475457492, 10448040663659726788, 10113917136857017974, 12479658147685402012, 3500241080921619556, 16959053435453822517, 12224711289652453635, 9342009439668884831, 4879704952849025062, 140226327413610143, 424854978622500449, 7259519967065370866, 7004910057750291985, 13293599522548616907, 10105770293752443592, 10668034807192757780, 1747952066141424100, 1654286879329379778, 8500057116360352059, 16929593379567477321, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 17376452488221285863, 9571781953733019530, 15391093639620504046, 13624874521033984333, 10027350355371872343, 18417369716475457492, 10448040663659726788, 10113917136857017974, 12479658147685402012, 3500241080921619556, 16959053435453822517, 12224711289652453635, 9342009439668884831, 4879704952849025062, 140226327413610143, 424854978622500449, 7259519967065370866, 7004910057750291985, 13293599522548616907, 10105770293752443592, 10668034807192757780, 1747952066141424100, 1654286879329379778, 8500057116360352059, 16929593379567477321,
					}
					testValuesFelt := make([]*fp.Element, len(testValues))
					for i, v := range testValues {
						testValuesFelt[i] = feltUint64(v)
					}
					consecutiveVarAddrResolvedValueEquals("keccak_ptr_end", testValuesFelt)(t, ctx)
				},
			},
			{
				operanders: []*hintOperander{
					{Name: "KECCAK_STATE_SIZE_FELTS", Kind: apRelative, Value: feltUint64(25)},
					{Name: "BLOCK_SIZE", Kind: apRelative, Value: feltUint64(1)},
					{Name: "keccak_ptr_end", Kind: apRelative, Value: addr(10)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newCairoKeccakFinalizeHint(ctx.operanders["KECCAK_STATE_SIZE_FELTS"], ctx.operanders["BLOCK_SIZE"], ctx.operanders["keccak_ptr_end"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					testValues := []uint64{
						0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 17376452488221285863, 9571781953733019530, 15391093639620504046, 13624874521033984333, 10027350355371872343, 18417369716475457492, 10448040663659726788, 10113917136857017974, 12479658147685402012, 3500241080921619556, 16959053435453822517, 12224711289652453635, 9342009439668884831, 4879704952849025062, 140226327413610143, 424854978622500449, 7259519967065370866, 7004910057750291985, 13293599522548616907, 10105770293752443592, 10668034807192757780, 1747952066141424100, 1654286879329379778, 8500057116360352059, 16929593379567477321,
					}
					testValuesFelt := make([]*fp.Element, len(testValues))
					for i, v := range testValues {
						testValuesFelt[i] = feltUint64(v)
					}
					consecutiveVarAddrResolvedValueEquals("keccak_ptr_end", testValuesFelt)(t, ctx)
				},
			},
		},
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
	})
}
