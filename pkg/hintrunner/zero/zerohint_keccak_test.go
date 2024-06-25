package zero

import (
	"fmt"
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func TestZeroHintKeccak(t *testing.T) {
	runHinterTests(t, map[string][]hintTestCase{
		"CairoKeccakFinalize": {
			{
				operanders: []*hintOperander{
					{Name: "keccak_ptr_end", Kind: apRelative, Value: addr(10)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newCairoKeccakFinalizeHint(ctx.operanders["keccak_ptr_end"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					testValues := []uint64{
						0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 17376452488221285863, 9571781953733019530, 15391093639620504046, 13624874521033984333, 10027350355371872343, 18417369716475457492, 10448040663659726788, 10113917136857017974, 12479658147685402012, 3500241080921619556, 16959053435453822517, 12224711289652453635, 9342009439668884831, 4879704952849025062, 140226327413610143, 424854978622500449, 7259519967065370866, 7004910057750291985, 13293599522548616907, 10105770293752443592, 10668034807192757780, 1747952066141424100, 1654286879329379778, 8500057116360352059, 16929593379567477321, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 17376452488221285863, 9571781953733019530, 15391093639620504046, 13624874521033984333, 10027350355371872343, 18417369716475457492, 10448040663659726788, 10113917136857017974, 12479658147685402012, 3500241080921619556, 16959053435453822517, 12224711289652453635, 9342009439668884831, 4879704952849025062, 140226327413610143, 424854978622500449, 7259519967065370866, 7004910057750291985, 13293599522548616907, 10105770293752443592, 10668034807192757780, 1747952066141424100, 1654286879329379778, 8500057116360352059, 16929593379567477321, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 17376452488221285863, 9571781953733019530, 15391093639620504046, 13624874521033984333, 10027350355371872343, 18417369716475457492, 10448040663659726788, 10113917136857017974, 12479658147685402012, 3500241080921619556, 16959053435453822517, 12224711289652453635, 9342009439668884831, 4879704952849025062, 140226327413610143, 424854978622500449, 7259519967065370866, 7004910057750291985, 13293599522548616907, 10105770293752443592, 10668034807192757780, 1747952066141424100, 1654286879329379778, 8500057116360352059, 16929593379567477321}
					testValuesFelt := make([]*fp.Element, len(testValues))
					for i, v := range testValues {
						testValuesFelt[i] = feltUint64(v)
					}
					consecutiveVarAddrResolvedValueEquals("keccak_ptr_end", testValuesFelt)(t, ctx)
				},
			},
			{
				operanders: []*hintOperander{
					{Name: "keccak_ptr_end", Kind: apRelative, Value: addr(10)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newCairoKeccakFinalizeHint(ctx.operanders["keccak_ptr_end"])
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
		"UnsafeKeccak": {
			{
				operanders: []*hintOperander{
					{Name: "data", Kind: uninitialized},
					{Name: "length", Kind: apRelative, Value: feltUint64((1 << 20) + 1)},
					{Name: "high", Kind: uninitialized},
					{Name: "low", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUnsafeKeccakHint(ctx.operanders["data"], ctx.operanders["length"], ctx.operanders["high"], ctx.operanders["low"])
				},
				errCheck: errorTextContains(fmt.Sprintf("unsafe_keccak() can only be used with length<=%d.\n Got: length=%d.", 1<<20, (1<<20)+1)),
			},
			{
				operanders: []*hintOperander{
					{Name: "data", Kind: apRelative, Value: addr(5)},
					{Name: "data.0", Kind: apRelative, Value: feltUint64(65537)},
					{Name: "length", Kind: apRelative, Value: feltUint64(1)},
					{Name: "high", Kind: uninitialized},
					{Name: "low", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUnsafeKeccakHint(ctx.operanders["data"], ctx.operanders["length"], ctx.operanders["high"], ctx.operanders["low"])
				},
				errCheck: errorTextContains(fmt.Sprintf("word %v is out range 0 <= word < 2 ** %d", feltUint64(65537), 8)),
			},
			{
				operanders: []*hintOperander{
					{Name: "data", Kind: apRelative, Value: addr(5)},
					{Name: "data.0", Kind: apRelative, Value: feltUint64(1)},
					{Name: "data.1", Kind: apRelative, Value: feltUint64(2)},
					{Name: "data.2", Kind: apRelative, Value: feltUint64(3)},
					{Name: "data.3", Kind: apRelative, Value: feltUint64(4)},
					{Name: "length", Kind: apRelative, Value: feltUint64(4)},
					{Name: "high", Kind: uninitialized},
					{Name: "low", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUnsafeKeccakHint(ctx.operanders["data"], ctx.operanders["length"], ctx.operanders["high"], ctx.operanders["low"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					varValueEquals("high", feltString("108955721224378455455648573289483395612"))(t, ctx)
					varValueEquals("low", feltString("253531040214470063354971884479696309631"))(t, ctx)
				},
			},
			{
				operanders: []*hintOperander{
					{Name: "data", Kind: apRelative, Value: addr(5)},
					{Name: "data.0", Kind: apRelative, Value: feltUint64(1)},
					{Name: "data.1", Kind: apRelative, Value: feltUint64(2)},
					{Name: "data.2", Kind: apRelative, Value: feltUint64(3)},
					{Name: "data.3", Kind: apRelative, Value: feltUint64(4)},
					{Name: "data.4", Kind: apRelative, Value: feltUint64(1)},
					{Name: "data.5", Kind: apRelative, Value: feltUint64(2)},
					{Name: "data.6", Kind: apRelative, Value: feltUint64(3)},
					{Name: "data.7", Kind: apRelative, Value: feltUint64(4)},
					{Name: "data.8", Kind: apRelative, Value: feltUint64(1)},
					{Name: "data.9", Kind: apRelative, Value: feltUint64(2)},
					{Name: "data.10", Kind: apRelative, Value: feltUint64(3)},
					{Name: "data.11", Kind: apRelative, Value: feltUint64(4)},
					{Name: "data.12", Kind: apRelative, Value: feltUint64(1)},
					{Name: "data.13", Kind: apRelative, Value: feltUint64(2)},
					{Name: "data.14", Kind: apRelative, Value: feltUint64(3)},
					{Name: "data.15", Kind: apRelative, Value: feltUint64(4)},
					{Name: "data.16", Kind: apRelative, Value: feltUint64(4)},
					{Name: "data.17", Kind: apRelative, Value: feltUint64(1)},
					{Name: "data.18", Kind: apRelative, Value: feltUint64(2)},
					{Name: "data.19", Kind: apRelative, Value: feltUint64(3)},
					{Name: "data.20", Kind: apRelative, Value: feltUint64(4)},
					{Name: "data.21", Kind: apRelative, Value: feltUint64(1)},
					{Name: "data.22", Kind: apRelative, Value: feltUint64(2)},
					{Name: "data.23", Kind: apRelative, Value: feltUint64(3)},
					{Name: "data.24", Kind: apRelative, Value: feltUint64(4)},
					{Name: "data.25", Kind: apRelative, Value: feltUint64(1)},
					{Name: "data.26", Kind: apRelative, Value: feltUint64(2)},
					{Name: "data.27", Kind: apRelative, Value: feltUint64(3)},
					{Name: "data.28", Kind: apRelative, Value: feltUint64(4)},
					{Name: "data.29", Kind: apRelative, Value: feltUint64(1)},
					{Name: "data.30", Kind: apRelative, Value: feltUint64(2)},
					{Name: "data.31", Kind: apRelative, Value: feltUint64(3)},
					{Name: "data.32", Kind: apRelative, Value: feltUint64(4)},
					{Name: "data.33", Kind: apRelative, Value: feltUint64(4)},
					{Name: "length", Kind: apRelative, Value: feltUint64(34)},
					{Name: "high", Kind: uninitialized},
					{Name: "low", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUnsafeKeccakHint(ctx.operanders["data"], ctx.operanders["length"], ctx.operanders["high"], ctx.operanders["low"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					varValueEquals("high", feltString("43087684015060895958086736298363333858"))(t, ctx)
					varValueEquals("low", feltString("115090685687501856751902560332884088627"))(t, ctx)
				},
			},
		},
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
		"UnsafeKeccakFinalize": {
			{
				// random values
				operanders: []*hintOperander{
					{Name: "keccak_state.start_ptr", Kind: apRelative, Value: addr(6)},
					{Name: "keccak_state.end_ptr", Kind: apRelative, Value: addr(10)},
					{Name: "input.0", Kind: apRelative, Value: feltUint64(1)},
					{Name: "input.1", Kind: apRelative, Value: feltUint64(2)},
					{Name: "input.2", Kind: apRelative, Value: feltUint64(3)},
					{Name: "input.3", Kind: apRelative, Value: feltUint64(4)},
					{Name: "high", Kind: uninitialized},
					{Name: "low", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUnsafeKeccakFinalizeHint(ctx.operanders["keccak_state.start_ptr"], ctx.operanders["high"], ctx.operanders["low"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					varValueEquals("high", feltString("83237039305559461909708573162245910191"))(t, ctx)
					varValueEquals("low", feltString("136371816236323410535574079355476245456"))(t, ctx)
				},
			},
			{
				// random values
				operanders: []*hintOperander{
					{Name: "keccak_state.start_ptr", Kind: apRelative, Value: addr(6)},
					{Name: "keccak_state.end_ptr", Kind: apRelative, Value: addr(22)},
					{Name: "input.0", Kind: apRelative, Value: feltUint64(16)},
					{Name: "input.1", Kind: apRelative, Value: feltUint64(15)},
					{Name: "input.2", Kind: apRelative, Value: feltUint64(14)},
					{Name: "input.3", Kind: apRelative, Value: feltUint64(13)},
					{Name: "input.4", Kind: apRelative, Value: feltUint64(12)},
					{Name: "input.5", Kind: apRelative, Value: feltUint64(11)},
					{Name: "input.6", Kind: apRelative, Value: feltUint64(10)},
					{Name: "input.7", Kind: apRelative, Value: feltUint64(9)},
					{Name: "input.8", Kind: apRelative, Value: feltUint64(8)},
					{Name: "input.9", Kind: apRelative, Value: feltUint64(7)},
					{Name: "input.10", Kind: apRelative, Value: feltUint64(6)},
					{Name: "input.11", Kind: apRelative, Value: feltUint64(5)},
					{Name: "input.12", Kind: apRelative, Value: feltUint64(4)},
					{Name: "input.13", Kind: apRelative, Value: feltUint64(3)},
					{Name: "input.14", Kind: apRelative, Value: feltUint64(2)},
					{Name: "input.15", Kind: apRelative, Value: feltUint64(1)},
					{Name: "high", Kind: uninitialized},
					{Name: "low", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUnsafeKeccakFinalizeHint(ctx.operanders["keccak_state.start_ptr"], ctx.operanders["high"], ctx.operanders["low"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					varValueEquals("high", feltString("138808390298063253082378171006150548014"))(t, ctx)
					varValueEquals("low", feltString("105025107803124291085217464954131521934"))(t, ctx)
				},
			},
			{
				// random big values
				operanders: []*hintOperander{
					{Name: "keccak_state.start_ptr", Kind: apRelative, Value: addr(6)},
					{Name: "keccak_state.end_ptr", Kind: apRelative, Value: addr(9)},
					{Name: "input.0", Kind: apRelative, Value: new(fp.Element).Sub(&utils.FeltMax128, &utils.FeltOne)},
					{Name: "input.1", Kind: apRelative, Value: new(fp.Element).Sub(&utils.FeltMax128, &utils.FeltOne)},
					{Name: "input.2", Kind: apRelative, Value: new(fp.Element).Sub(&utils.FeltMax128, &utils.FeltOne)},
					{Name: "high", Kind: uninitialized},
					{Name: "low", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUnsafeKeccakFinalizeHint(ctx.operanders["keccak_state.start_ptr"], ctx.operanders["high"], ctx.operanders["low"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					varValueEquals("high", feltString("223857737769798933062329368034598208286"))(t, ctx)
					varValueEquals("low", feltString("310600734091368901761065530388449236858"))(t, ctx)
				},
			},
			{
				// random values exceeding 2<<128
				operanders: []*hintOperander{
					{Name: "keccak_state.start_ptr", Kind: apRelative, Value: addr(6)},
					{Name: "keccak_state.end_ptr", Kind: apRelative, Value: addr(9)},
					{Name: "input.0", Kind: apRelative, Value: &utils.Felt127},
					{Name: "input.1", Kind: apRelative, Value: new(fp.Element).Sub(&utils.FeltMax128, &utils.FeltOne)},
					{Name: "input.2", Kind: apRelative, Value: &utils.FeltMax128},
					{Name: "high", Kind: uninitialized},
					{Name: "low", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUnsafeKeccakFinalizeHint(ctx.operanders["keccak_state.start_ptr"], ctx.operanders["high"], ctx.operanders["low"])
				},
				errCheck: errorTextContains(fmt.Sprintf("word %v is out range 0 <= word < 2 ** 128", &utils.FeltMax128)),
			},
			{
				// end - start is 0
				operanders: []*hintOperander{
					{Name: "keccak_state.start_ptr", Kind: apRelative, Value: addr(6)},
					{Name: "keccak_state.end_ptr", Kind: apRelative, Value: addr(6)},
					{Name: "high", Kind: uninitialized},
					{Name: "low", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUnsafeKeccakFinalizeHint(ctx.operanders["keccak_state.start_ptr"], ctx.operanders["high"], ctx.operanders["low"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					varValueEquals("high", feltString("262949717399590921288928019264691438528"))(t, ctx)
					varValueEquals("low", feltString("304396909071904405792975023732328604784"))(t, ctx)
				},
			},
		},
		"BlockPermutation": {
			{
				operanders: []*hintOperander{
					{Name: "keccak_ptr", Kind: fpRelative, Value: addr(30)},
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
					return newBlockPermutationHint(ctx.operanders["keccak_ptr"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"keccak_ptr",
					[]*fp.Element{
						feltUint64(12483095336943515612),
						feltUint64(15677359730926197488),
						feltUint64(7487778311628531317),
						feltUint64(1821627048823728482),
						feltUint64(11485992932336471799),
						feltUint64(16469217220755308995),
						feltUint64(3029672297743876521),
						feltUint64(4787226438136518340),
						feltUint64(17694526120416454034),
						feltUint64(17551465471496379789),
						feltUint64(9299325703581808762),
						feltUint64(8817815188065733198),
						feltUint64(8697009915081020406),
						feltUint64(8906369854620102227),
						feltUint64(14321543399670582665),
						feltUint64(6384661976273651103),
						feltUint64(11524950614921587710),
						feltUint64(10736292889273693277),
						feltUint64(9487666051186580327),
						feltUint64(12129519010572669737),
						feltUint64(13749616481815304298),
						feltUint64(11956376265587856622),
						feltUint64(7332632521547853118),
						feltUint64(3137160411931496300),
						feltUint64(4751701212705667336),
					}),
			},
			{
				operanders: []*hintOperander{
					{Name: "keccak_ptr", Kind: fpRelative, Value: addr(30)},
					{Name: "data.0", Kind: apRelative, Value: feltUint64(12)},
					{Name: "data.1", Kind: apRelative, Value: feltUint64(12)},
					{Name: "data.2", Kind: apRelative, Value: feltUint64(12)},
					{Name: "data.3", Kind: apRelative, Value: feltUint64(4)},
					{Name: "data.4", Kind: apRelative, Value: feltUint64(5)},
					{Name: "data.5", Kind: apRelative, Value: feltUint64(5)},
					{Name: "data.6", Kind: apRelative, Value: feltUint64(5)},
					{Name: "data.7", Kind: apRelative, Value: feltUint64(5)},
					{Name: "data.8", Kind: apRelative, Value: feltUint64(5)},
					{Name: "data.9", Kind: apRelative, Value: feltUint64(5)},
					{Name: "data.10", Kind: apRelative, Value: feltUint64(5)},
					{Name: "data.11", Kind: apRelative, Value: feltUint64(5)},
					{Name: "data.12", Kind: apRelative, Value: feltUint64(5)},
					{Name: "data.13", Kind: apRelative, Value: feltUint64(5)},
					{Name: "data.14", Kind: apRelative, Value: feltUint64(5)},
					{Name: "data.15", Kind: apRelative, Value: feltUint64(5)},
					{Name: "data.16", Kind: apRelative, Value: feltUint64(9223372036854775813)},
					{Name: "data.17", Kind: apRelative, Value: feltUint64(5)},
					{Name: "data.18", Kind: apRelative, Value: feltUint64(5)},
					{Name: "data.19", Kind: apRelative, Value: feltUint64(5)},
					{Name: "data.20", Kind: apRelative, Value: feltUint64(5)},
					{Name: "data.21", Kind: apRelative, Value: feltUint64(5)},
					{Name: "data.22", Kind: apRelative, Value: feltUint64(5)},
					{Name: "data.23", Kind: apRelative, Value: feltUint64(5)},
					{Name: "data.24", Kind: apRelative, Value: feltUint64(5)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newBlockPermutationHint(ctx.operanders["keccak_ptr"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"keccak_ptr",
					[]*fp.Element{
						feltUint64(4383730378237241658),
						feltUint64(12444326698015973637),
						feltUint64(12104859138858588073),
						feltUint64(16800419876497365534),
						feltUint64(1687670000686051907),
						feltUint64(9357203110893952832),
						feltUint64(15543449195353005377),
						feltUint64(10018439473843331598),
						feltUint64(4252067722003831863),
						feltUint64(15043026706643958401),
						feltUint64(3385897973286642985),
						feltUint64(14187183346503360277),
						feltUint64(14371620541932978658),
						feltUint64(13148092299950144338),
						feltUint64(1566125651106358880),
						feltUint64(1657163393784433257),
						feltUint64(17402730062666126860),
						feltUint64(10877209625975612823),
						feltUint64(10316109692548342913),
						feltUint64(14165981879239017861),
						feltUint64(5012909831161588138),
						feltUint64(12731032323068111667),
						feltUint64(6467007756398199334),
						feltUint64(17322593527878179950),
						feltUint64(14146902728521851886),
					}),
			},
		},
		"CompareBytesInWordHint": {
			{
				operanders: []*hintOperander{
					{Name: "n_bytes", Kind: fpRelative, Value: feltUint64(5)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newCompareBytesInWordHint(ctx.operanders["n_bytes"])
				},
				check: apValueEquals(feltUint64(1)),
			},
			{
				operanders: []*hintOperander{
					{Name: "n_bytes", Kind: fpRelative, Value: feltUint64(10)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newCompareBytesInWordHint(ctx.operanders["n_bytes"])
				},
				check: apValueEquals(feltUint64(0)),
			},
		},
	})
}
