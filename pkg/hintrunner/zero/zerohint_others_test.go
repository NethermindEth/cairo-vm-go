package zero

import (
	"math"
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
					{Name: "arrayPtr", Kind: apRelative, Value: addr(5)},
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
				errCheck: errorTextContains("Invalid value for elm_size"),
			},
			{
				operanders: []*hintOperander{
					{Name: "arrayPtr", Kind: apRelative, Value: addr(5)},
					{Name: "elmSize", Kind: fpRelative, Value: feltInt64(4)},
					{Name: "nElms", Kind: fpRelative, Value: feltInt64(0)},
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
				errCheck: errorTextContains("Invalid value for n_elms"),
			},
			{
				operanders: []*hintOperander{
					{Name: "arrayPtr", Kind: apRelative, Value: addr(0)},
					{Name: "elmSize", Kind: fpRelative, Value: feltInt64(4)},
					{Name: "nElms", Kind: fpRelative, Value: feltInt64(4)},
					{Name: "index", Kind: fpRelative, Value: feltInt64(2)},
					{Name: "key", Kind: fpRelative, Value: feltInt64(2)},
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
				errCheck: errorTextContains("failed obtaining the variable: __find_element_max_size"),
			},
			{
				operanders: []*hintOperander{
					{Name: "array.el0", Kind: apRelative, Value: feltInt64(0)},
					{Name: "array.el1", Kind: apRelative, Value: feltInt64(1)},
					{Name: "array.el2", Kind: apRelative, Value: feltInt64(2)},
					{Name: "array.el3", Kind: apRelative, Value: feltInt64(3)},
					{Name: "arrayPtr", Kind: apRelative, Value: addr(0)},
					{Name: "elmSize", Kind: fpRelative, Value: feltInt64(4)},
					{Name: "nElms", Kind: fpRelative, Value: feltInt64(8)},
					{Name: "index", Kind: fpRelative, Value: feltInt64(2)},
					{Name: "key", Kind: fpRelative, Value: feltInt64(2)},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					// make __find_element_max_size exist in scope
					ctx.ScopeManager.EnterScope(map[string]any{"__find_element_max_size": math.Pow(2, 20)})
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
					{Name: "index", Kind: uninitialized},
					{Name: "arrayPtr", Kind: apRelative, Value: addr(0)},
					{Name: "key", Kind: fpRelative, Value: feltInt64(4)},
					{Name: "nElms", Kind: fpRelative, Value: feltInt64(4)},
					{Name: "elmSize", Kind: fpRelative, Value: feltInt64(4)},
					{Name: "array.el0", Kind: apRelative, Value: feltInt64(0)},
					{Name: "array.el1", Kind: apRelative, Value: feltInt64(1)},
					{Name: "array.el2", Kind: apRelative, Value: feltInt64(2)},
					{Name: "array.el3", Kind: apRelative, Value: feltInt64(3)},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					// make __find_element_max_size exist in scope
					ctx.ScopeManager.EnterScope(map[string]any{"__find_element_max_size": math.Pow(2, 20)})
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
					{Name: "index", Kind: uninitialized},
					{Name: "arrayPtr", Kind: apRelative, Value: addr(0)},
					{Name: "key", Kind: fpRelative, Value: feltInt64(5)},
					{Name: "nElms", Kind: fpRelative, Value: feltInt64(4)},
					{Name: "elmSize", Kind: fpRelative, Value: feltInt64(4)},
					{Name: "array.el0", Kind: apRelative, Value: feltInt64(0)},
					{Name: "array.el1", Kind: apRelative, Value: feltInt64(1)},
					{Name: "array.el2", Kind: apRelative, Value: feltInt64(2)},
					{Name: "array.el3", Kind: apRelative, Value: feltInt64(3)},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					// make __find_element_max_size exist in scope
					ctx.ScopeManager.EnterScope(map[string]any{"__find_element_max_size": math.Pow(2, 20)})
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
