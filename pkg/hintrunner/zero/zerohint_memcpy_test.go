package zero

import (
	"math"
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
)

func TestZeroHintOthers(t *testing.T) {
	runHinterTests(t, map[string][]hintTestCase{
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
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSearchSortedLowerHint(
						ctx.operanders["arrayPtr"],
						ctx.operanders["elmSize"],
						ctx.operanders["nElms"],
						ctx.operanders["key"],
						ctx.operanders["index"],
					)
				},
				errCheck: errorTextContains("find_element() can only be used with n_elms<="),
			},

			{
				operanders: []*hintOperander{
					{Name: "array.el0", Kind: apRelative, Value: feltInt64(0)},
					{Name: "array.el1", Kind: apRelative, Value: feltInt64(1)},
					{Name: "array.el2", Kind: apRelative, Value: feltInt64(2)},
					{Name: "array.el3", Kind: apRelative, Value: feltInt64(3)},
					{Name: "arrayPtr", Kind: apRelative, Value: addr(0)},
					{Name: "elmSize", Kind: fpRelative, Value: feltInt64(4)},
					{Name: "nElms", Kind: fpRelative, Value: feltInt64(9)},
					{Name: "index", Kind: fpRelative, Value: feltInt64(3)},
					{Name: "key", Kind: fpRelative, Value: feltInt64(0)},
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
