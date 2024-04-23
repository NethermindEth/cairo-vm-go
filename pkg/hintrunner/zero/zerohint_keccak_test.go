package zero

import (
	"fmt"
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func TestZeroHintKeccak(t *testing.T) {

	runHinterTests(t, map[string][]hintTestCase{
		"CairoKeccakFinalize": {
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
		"UnsafeKeccak": {
			{
				operanders: []*hintOperander{
					{Name: "data", Kind: uninitialized},
					{Name: "length", Kind: apRelative, Value: feltUint64(101)},
					{Name: "high", Kind: uninitialized},
					{Name: "low", Kind: uninitialized},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					hinter.InitializeScopeManager(ctx, map[string]any{
						"__keccak_max_size": uint64(100),
					})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUnsafeKeccakHint(ctx.operanders["data"], ctx.operanders["length"], ctx.operanders["high"], ctx.operanders["low"])
				},
				errCheck: errorTextContains(fmt.Sprintf("unsafe_keccak() can only be used with length<=%d.\n Got: length=%d.", 100, 101)),
			},
			{
				operanders: []*hintOperander{
					{Name: "data", Kind: apRelative, Value: addr(5)},
					{Name: "data.0", Kind: apRelative, Value: feltUint64(65537)},
					{Name: "length", Kind: apRelative, Value: feltUint64(1)},
					{Name: "high", Kind: uninitialized},
					{Name: "low", Kind: uninitialized},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					hinter.InitializeScopeManager(ctx, map[string]any{
						"__keccak_max_size": uint64(100),
					})
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
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					hinter.InitializeScopeManager(ctx, map[string]any{
						"__keccak_max_size": uint64(100),
					})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUnsafeKeccakHint(ctx.operanders["data"], ctx.operanders["length"], ctx.operanders["high"], ctx.operanders["low"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					varValueEquals("high", feltString("108955721224378455455648573289483395612"))(t, ctx)
					varValueEquals("low", feltString("253531040214470063354971884479696309631"))(t, ctx)
				},
			},
		},
		"UnsafeKeccakFinalize": {
			{
				// random values
				operanders: []*hintOperander{
					{Name: "keccak_state.start_ptr", Kind: apRelative, Value: addr(6)},
					{Name: "keccak_state.end_ptr", Kind: apRelative, Value: addr(10)},
					{Name: "keccak_state.0", Kind: apRelative, Value: feltUint64(1)},
					{Name: "keccak_state.1", Kind: apRelative, Value: feltUint64(2)},
					{Name: "keccak_state.2", Kind: apRelative, Value: feltUint64(3)},
					{Name: "keccak_state.3", Kind: apRelative, Value: feltUint64(4)},
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
	})
}
