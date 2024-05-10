package zero

import (
	"fmt"
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
)

func TestZeroHintMemcpy(t *testing.T) {

	runHinterTests(t, map[string][]hintTestCase{
		"MemcpyEnterScope": {
			{
				operanders: []*hintOperander{
					{Name: "len", Kind: apRelative, Value: feltUint64(1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newMemcpyEnterScopeHint(ctx.operanders["len"])
				},
				check: varValueInScopeEquals("n", feltUint64(1)),
			},
		},
		"FindElement": {
			{
				operanders: []*hintOperander{
					{Name: "array_ptr", Kind: apRelative, Value: addr(7)},
					{Name: "elm_size", Kind: apRelative, Value: feltUint64(0)},
					{Name: "key", Kind: apRelative, Value: feltUint64(1)},
					{Name: "index", Kind: apRelative, Value: uninitialized},
					{Name: "n_elms", Kind: apRelative, Value: feltUint64(0)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newFindElementHint(ctx.operanders["array_ptr"], ctx.operanders["elm_size"], ctx.operanders["key"], ctx.operanders["index"], ctx.operanders["n_elms"])
				},
				errCheck: errorTextContains(fmt.Sprintf("Invalid value for elm_size. Got: %v", 0)),
			},
			{
				operanders: []*hintOperander{
					{Name: "array_ptr", Kind: apRelative, Value: addr(7)},
					{Name: "array.0", Kind: apRelative, Value: feltUint64(1)},
					{Name: "array.1", Kind: apRelative, Value: feltUint64(2)},
					{Name: "array.2", Kind: apRelative, Value: feltUint64(3)},
					{Name: "elm_size", Kind: apRelative, Value: feltUint64(1)},
					{Name: "key", Kind: apRelative, Value: feltUint64(2)},
					{Name: "index", Kind: apRelative, Value: uninitialized},
					{Name: "n_elms", Kind: apRelative, Value: feltUint64(0)},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("__find_element_index", feltUint64(0))
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newFindElementHint(ctx.operanders["array_ptr"], ctx.operanders["elm_size"], ctx.operanders["key"], ctx.operanders["index"], ctx.operanders["n_elms"])
				},
				errCheck: errorTextContains(fmt.Sprintf("Invalid index found in __find_element_index. index: %d, expected key %d, found key: %d", feltUint64(3), feltUint64(2), feltUint64(3))),
			},
			{
				operanders: []*hintOperander{
					{Name: "array_ptr", Kind: apRelative, Value: addr(7)},
					{Name: "array.0", Kind: apRelative, Value: feltUint64(1)},
					{Name: "array.1", Kind: apRelative, Value: feltUint64(2)},
					{Name: "array.2", Kind: apRelative, Value: feltUint64(3)},
					{Name: "elm_size", Kind: apRelative, Value: feltUint64(1)},
					{Name: "key", Kind: apRelative, Value: feltUint64(3)},
					{Name: "index", Kind: apRelative, Value: uninitialized},
					{Name: "n_elms", Kind: apRelative, Value: feltUint64(0)},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("__find_element_index", feltUint64(2))
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newFindElementHint(ctx.operanders["array_ptr"], ctx.operanders["elm_size"], ctx.operanders["key"], ctx.operanders["index"], ctx.operanders["n_elms"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					varValueEquals("index", feltUint64(2))(t, ctx)
					varValueNotInScope("__find_element_index")(t, ctx)
				},
			},
			{
				operanders: []*hintOperander{
					{Name: "array_ptr", Kind: apRelative, Value: addr(7)},
					{Name: "elm_size", Kind: apRelative, Value: feltUint64(1)},
					{Name: "key", Kind: apRelative, Value: feltUint64(3)},
					{Name: "index", Kind: apRelative, Value: uninitialized},
					{Name: "n_elms", Kind: apRelative, Value: feltUint64(2)},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("__find_element_max_size", feltUint64(1))
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newFindElementHint(ctx.operanders["array_ptr"], ctx.operanders["elm_size"], ctx.operanders["key"], ctx.operanders["index"], ctx.operanders["n_elms"])
				},
				errCheck: errorTextContains(fmt.Sprintf("find_element() can only be used with n_elms<=%d. Got: n_elms=%d", feltUint64(1), feltUint64(2))),
			},
			{
				operanders: []*hintOperander{
					{Name: "array_ptr", Kind: apRelative, Value: addr(7)},
					{Name: "array.0", Kind: apRelative, Value: feltUint64(1)},
					{Name: "array.1", Kind: apRelative, Value: feltUint64(2)},
					{Name: "array.2", Kind: apRelative, Value: feltUint64(3)},
					{Name: "elm_size", Kind: apRelative, Value: feltUint64(1)},
					{Name: "key", Kind: apRelative, Value: feltUint64(3)},
					{Name: "index", Kind: apRelative, Value: uninitialized},
					{Name: "n_elms", Kind: apRelative, Value: feltUint64(0)},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("__find_element_max_size", feltUint64(100))
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newFindElementHint(ctx.operanders["array_ptr"], ctx.operanders["elm_size"], ctx.operanders["key"], ctx.operanders["index"], ctx.operanders["n_elms"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					varValueEquals("index", feltUint64(2))(t, ctx)
					varValueNotInScope("__find_element_index")(t, ctx)
				},
			},
			{
				operanders: []*hintOperander{
					{Name: "array_ptr", Kind: apRelative, Value: addr(7)},
					{Name: "array.0", Kind: apRelative, Value: feltUint64(1)},
					{Name: "array.1", Kind: apRelative, Value: feltUint64(2)},
					{Name: "array.2", Kind: apRelative, Value: feltUint64(3)},
					{Name: "elm_size", Kind: apRelative, Value: feltUint64(1)},
					{Name: "key", Kind: apRelative, Value: feltUint64(5)},
					{Name: "index", Kind: apRelative, Value: uninitialized},
					{Name: "n_elms", Kind: apRelative, Value: feltUint64(0)},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("__find_element_max_size", feltUint64(100))
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newFindElementHint(ctx.operanders["array_ptr"], ctx.operanders["elm_size"], ctx.operanders["key"], ctx.operanders["index"], ctx.operanders["n_elms"])
				},
				errCheck: errorTextContains(fmt.Sprintf("Key %v was not found", feltUint64(5))),
			},
		},
	})
}
