package zero

import (
	"fmt"
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
)

func TestZeroHintMemcpy(t *testing.T) {

	runHinterTests(t, map[string][]hintTestCase{
		"MemcpyContinueCopying": {
			{
				operanders: []*hintOperander{
					{Name: "continue_copying", Kind: uninitialized},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("n", *feltInt64(1))
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newMemContinueHint(ctx.operanders["continue_copying"], false)
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					varValueInScopeEquals("n", *feltInt64(0))(t, ctx)
					varValueEquals("continue_copying", feltInt64(0))(t, ctx)
				},
			},
			{
				operanders: []*hintOperander{
					{Name: "continue_copying", Kind: uninitialized},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("n", *feltInt64(5))
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newMemContinueHint(ctx.operanders["continue_copying"], false)
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					varValueInScopeEquals("n", *feltInt64(4))(t, ctx)
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
					return newMemEnterScopeHint(ctx.operanders["len"], false)
				},
				check: varValueInScopeEquals("n", *feltUint64(1)),
			},
		},
		"FindElement": {
			{
				operanders: []*hintOperander{
					{Name: "array_ptr", Kind: apRelative, Value: addr(6)},
					{Name: "elm_size", Kind: apRelative, Value: feltUint64(0)},
					{Name: "key", Kind: apRelative, Value: feltUint64(1)},
					{Name: "index", Kind: uninitialized},
					{Name: "n_elms", Kind: apRelative, Value: feltUint64(0)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newFindElementHint(ctx.operanders["array_ptr"], ctx.operanders["elm_size"], ctx.operanders["key"], ctx.operanders["index"], ctx.operanders["n_elms"])
				},
				errCheck: errorTextContains(fmt.Sprintf("Invalid value for elm_size. Got: %v", 0)),
			},
			{
				operanders: []*hintOperander{
					{Name: "array_ptr", Kind: apRelative, Value: addr(9)},
					{Name: "elm_size", Kind: apRelative, Value: feltUint64(1)},
					{Name: "key", Kind: apRelative, Value: feltUint64(999)},
					{Name: "index", Kind: uninitialized},
					{Name: "n_elms", Kind: apRelative, Value: feltUint64(0)},
					{Name: "array.0", Kind: apRelative, Value: feltUint64(100)},
					{Name: "array.1", Kind: apRelative, Value: feltUint64(200)},
					{Name: "array.2", Kind: apRelative, Value: feltUint64(300)},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("__find_element_index", uint64(0))
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newFindElementHint(ctx.operanders["array_ptr"], ctx.operanders["elm_size"], ctx.operanders["key"], ctx.operanders["index"], ctx.operanders["n_elms"])
				},
				errCheck: errorTextContains(fmt.Sprintf("Invalid index found in __find_element_index. index: %v, expected key %v, found key: %v", feltUint64(0), feltUint64(999), feltUint64(100))),
			},
			{
				operanders: []*hintOperander{
					{Name: "array_ptr", Kind: apRelative, Value: addr(9)},
					{Name: "elm_size", Kind: apRelative, Value: feltUint64(1)},
					{Name: "key", Kind: apRelative, Value: feltUint64(300)},
					{Name: "index", Kind: uninitialized},
					{Name: "n_elms", Kind: apRelative, Value: feltUint64(0)},
					{Name: "array.0", Kind: apRelative, Value: feltUint64(100)},
					{Name: "array.1", Kind: apRelative, Value: feltUint64(200)},
					{Name: "array.2", Kind: apRelative, Value: feltUint64(300)},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("__find_element_index", uint64(2))
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
					{Name: "index", Kind: uninitialized},
					{Name: "n_elms", Kind: apRelative, Value: feltUint64(2)},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("__find_element_max_size", uint64(1))
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newFindElementHint(ctx.operanders["array_ptr"], ctx.operanders["elm_size"], ctx.operanders["key"], ctx.operanders["index"], ctx.operanders["n_elms"])
				},
				errCheck: errorTextContains(fmt.Sprintf("find_element() can only be used with n_elms<=%v. Got: n_elms=%v", feltUint64(1), feltUint64(2))),
			},
			{
				operanders: []*hintOperander{
					{Name: "array_ptr", Kind: apRelative, Value: addr(9)},
					{Name: "elm_size", Kind: apRelative, Value: feltUint64(1)},
					{Name: "key", Kind: apRelative, Value: feltUint64(300)},
					{Name: "index", Kind: uninitialized},
					{Name: "n_elms", Kind: apRelative, Value: feltUint64(3)},
					{Name: "array.0", Kind: apRelative, Value: feltUint64(100)},
					{Name: "array.1", Kind: apRelative, Value: feltUint64(200)},
					{Name: "array.2", Kind: apRelative, Value: feltUint64(300)},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("__find_element_max_size", uint64(100))
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
					{Name: "array_ptr", Kind: apRelative, Value: addr(9)},
					{Name: "elm_size", Kind: apRelative, Value: feltUint64(1)},
					{Name: "key", Kind: apRelative, Value: feltUint64(999)},
					{Name: "index", Kind: uninitialized},
					{Name: "n_elms", Kind: apRelative, Value: feltUint64(0)},
					{Name: "array.0", Kind: apRelative, Value: feltUint64(100)},
					{Name: "array.1", Kind: apRelative, Value: feltUint64(200)},
					{Name: "array.2", Kind: apRelative, Value: feltUint64(300)},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("__find_element_max_size", uint64(100))
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newFindElementHint(ctx.operanders["array_ptr"], ctx.operanders["elm_size"], ctx.operanders["key"], ctx.operanders["index"], ctx.operanders["n_elms"])
				},
				errCheck: errorTextContains(fmt.Sprintf("Key %v was not found", feltUint64(999))),
			},
		},
	})
}
