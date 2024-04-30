package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
)

func TestZeroHintOthers(t *testing.T) {
	runHinterTests(t, map[string][]hintTestCase{
		"SearchSortedLower": {
			{
				operanders: []*hintOperander{
					{Name: "arrayPtr", Kind: apRelative, Value: addr(0)},
					{Name: "elmSize", Kind: fpRelative, Value: feltInt64(0)},
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
					{Name: "arrayPtr", Kind: apRelative, Value: addr(0)},
					{Name: "elmSize", Kind: fpRelative, Value: feltInt64(4)},
					{Name: "nElms", Kind: fpRelative, Value: feltInt64(0)},
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
					{Name: "arrayPtr", Kind: apRelative, Value: addr(0)},
					{Name: "elmSize", Kind: fpRelative, Value: feltInt64(4)},
					{Name: "nElms", Kind: fpRelative, Value: feltInt64(8)},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					// make __find_element_max_size exist in scope
					maxElementSize := feltInt64(2)
					ctx.ScopeManager.EnterScope(map[string]any{"__find_element_max_size": maxElementSize})
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
					{Name: "arrayPtr", Kind: apRelative, Value: addr(0)},
					{Name: "elmSize", Kind: fpRelative, Value: feltInt64(4)},
					{Name: "nElms", Kind: fpRelative, Value: feltInt64(4)},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					// make __find_element_max_size exist in scope
					maxElementSize := feltInt64(8)
					ctx.ScopeManager.EnterScope(map[string]any{"__find_element_max_size": maxElementSize})
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
