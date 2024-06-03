package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
)

func TestZeroHintMemcpy(t *testing.T) {

	runHinterTests(t, map[string][]hintTestCase{
		"MemcPyContinueCopying": {
			{
				operanders: []*hintOperander{
					{Name: "continue_copying", Kind: uninitialized},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{
						"n": &utils.FeltOne,
					})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newMemcpyContinueCopyingHint(ctx.operanders["continue_copying"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					allVarValueInScopeEquals(map[string]any{"n": feltInt64(0)})(t, ctx)
					varValueEquals("continue_copying", feltInt64(0))(t, ctx)
				},
			},
			{
				operanders: []*hintOperander{
					{Name: "continue_copying", Kind: uninitialized},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{
						"n": feltString("5"),
					})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newMemcpyContinueCopyingHint(ctx.operanders["continue_copying"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					allVarValueInScopeEquals(map[string]any{"n": feltInt64(4)})(t, ctx)
					varValueEquals("continue_copying", feltInt64(1))(t, ctx)
				},
			},
		},
		"MemcpyEnterScope": {
			{
				operanders: []*hintOperander{
					{Name: "len", Kind: apRelative, Value: feltUint64(1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newMemcpyEnterScopeHint(ctx.operanders["len"])
				},
				check: varValueInScopeEquals("n", *feltUint64(1)),
			},
		},
		"SearchSortedLower": {
			{
				operanders: []*hintOperander{
					{Name: "arrayPtr", Kind: fpRelative, Value: feltUint64(10)},
					{Name: "elmSize", Kind: fpRelative, Value: feltInt64(0)},
					{Name: "nElms", Kind: uninitialized},
					{Name: "index", Kind: uninitialized},
					{Name: "key", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSearchSortedLowerHint(
						ctx.operanders["arrayPtr"],
						ctx.operanders["elmSize"],
						ctx.operanders["nElms"],
						ctx.operanders["key"],
						ctx.operanders["index"],
					)
				},
				errCheck: errorTextContains("invalid value for elm_size. Got: 0"),
			},
			{
				operanders: []*hintOperander{
					{Name: "arrayPtr", Kind: fpRelative, Value: feltInt64(10)},
					{Name: "elmSize", Kind: fpRelative, Value: feltInt64(1)},
					{Name: "nElms", Kind: fpRelative, Value: feltInt64(0)},
					{Name: "index", Kind: uninitialized},
					{Name: "key", Kind: fpRelative, Value: feltInt64(1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSearchSortedLowerHint(
						ctx.operanders["arrayPtr"],
						ctx.operanders["elmSize"],
						ctx.operanders["nElms"],
						ctx.operanders["key"],
						ctx.operanders["index"],
					)
				},
				errCheck: errorIsNil,
			},
			{
				operanders: []*hintOperander{
					// We needs these buffer here because we read from offset 10 in segment 1
					// and test allocs until addr(6)
					{Name: "buffer", Kind: apRelative, Value: feltInt64(0)},
					{Name: "buffer", Kind: apRelative, Value: feltInt64(0)},
					{Name: "buffer", Kind: apRelative, Value: feltInt64(0)},
					{Name: "buffer", Kind: apRelative, Value: feltInt64(0)},

					{Name: "arrayPtr", Kind: fpRelative, Value: feltInt64(10)},
					{Name: "elmSize", Kind: fpRelative, Value: feltInt64(1)},
					{Name: "nElms", Kind: fpRelative, Value: feltInt64(10)},
					{Name: "index", Kind: uninitialized},
					{Name: "key", Kind: fpRelative, Value: feltInt64(1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSearchSortedLowerHint(
						ctx.operanders["arrayPtr"],
						ctx.operanders["elmSize"],
						ctx.operanders["nElms"],
						ctx.operanders["key"],
						ctx.operanders["index"],
					)
				},
				errCheck: errorIsNil,
			},
		},
	})
}
