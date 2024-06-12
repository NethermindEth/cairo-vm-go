package zero

import (
	"fmt"
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func TestZeroHintOthers(t *testing.T) {
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
		"SearchSortedLower": {
			{
				operanders: []*hintOperander{
					{Name: "array_ptr", Kind: fpRelative, Value: addr(7)},
					{Name: "elm_size", Kind: fpRelative, Value: feltInt64(0)},
					{Name: "n_elms", Kind: uninitialized},
					{Name: "key", Kind: uninitialized},
					{Name: "index", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSearchSortedLowerHint(
						ctx.operanders["array_ptr"],
						ctx.operanders["elm_size"],
						ctx.operanders["n_elms"],
						ctx.operanders["key"],
						ctx.operanders["index"],
					)
				},
				errCheck: errorTextContains("invalid value for elm_size. Got: 0"),
			},
			{
				operanders: []*hintOperander{
					{Name: "array_ptr", Kind: fpRelative, Value: addr(8)},
					{Name: "elm_size", Kind: fpRelative, Value: feltInt64(1)},
					{Name: "n_elms", Kind: fpRelative, Value: feltInt64(1)},
					{Name: "key", Kind: fpRelative, Value: feltInt64(1)},
					{Name: "firstElement", Kind: apRelative, Value: feltInt64(0)},
					{Name: "index", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSearchSortedLowerHint(
						ctx.operanders["array_ptr"],
						ctx.operanders["elm_size"],
						ctx.operanders["n_elms"],
						ctx.operanders["key"],
						ctx.operanders["index"],
					)
				},
				errCheck: errorIsNil,
			},
			{
				operanders: []*hintOperander{
					{Name: "array_ptr", Kind: fpRelative, Value: addr(8)},
					{Name: "elm_size", Kind: fpRelative, Value: feltInt64(1)},
					{Name: "n_elms", Kind: fpRelative, Value: feltInt64(1)},
					{Name: "key", Kind: fpRelative, Value: feltInt64(2)},
					{Name: "firstElement", Kind: apRelative, Value: feltInt64(1)},
					{Name: "index", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSearchSortedLowerHint(
						ctx.operanders["array_ptr"],
						ctx.operanders["elm_size"],
						ctx.operanders["n_elms"],
						ctx.operanders["key"],
						ctx.operanders["index"],
					)
				},
				check: varValueEquals("index", feltInt64(1)),
			},
			{
				operanders: []*hintOperander{
					{Name: "array_ptr", Kind: fpRelative, Value: addr(8)},
					{Name: "elm_size", Kind: fpRelative, Value: feltInt64(1)},
					{Name: "n_elms", Kind: fpRelative, Value: feltInt64(1)},
					{Name: "key", Kind: fpRelative, Value: feltInt64(1)},
					{Name: "firstElement", Kind: apRelative, Value: feltInt64(1)},
					{Name: "index", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSearchSortedLowerHint(
						ctx.operanders["array_ptr"],
						ctx.operanders["elm_size"],
						ctx.operanders["n_elms"],
						ctx.operanders["key"],
						ctx.operanders["index"],
					)
				},
				check: varValueEquals("index", feltInt64(0)),
			},
			{
				operanders: []*hintOperander{
					{Name: "array_ptr", Kind: fpRelative, Value: addr(8)},
					{Name: "elm_size", Kind: fpRelative, Value: feltInt64(1)},
					{Name: "n_elms", Kind: fpRelative, Value: feltInt64(4)},
					{Name: "key", Kind: fpRelative, Value: feltInt64(2)},
					{Name: "firstElement", Kind: apRelative, Value: feltInt64(0)},
					{Name: "secondElement", Kind: apRelative, Value: feltInt64(1)},
					{Name: "thirdElement", Kind: apRelative, Value: feltInt64(2)},
					{Name: "fourthElement", Kind: apRelative, Value: feltInt64(3)},
					{Name: "index", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSearchSortedLowerHint(
						ctx.operanders["array_ptr"],
						ctx.operanders["elm_size"],
						ctx.operanders["n_elms"],
						ctx.operanders["key"],
						ctx.operanders["index"],
					)
				},
				check: varValueEquals("index", feltInt64(2)),
			},
			{
				operanders: []*hintOperander{
					{Name: "array_ptr", Kind: fpRelative, Value: addr(8)},
					{Name: "elm_size", Kind: fpRelative, Value: feltInt64(1)},
					{Name: "n_elms", Kind: fpRelative, Value: feltInt64(6)},
					{Name: "key", Kind: fpRelative, Value: feltInt64(47)},
					{Name: "firstElement", Kind: apRelative, Value: feltInt64(11)},
					{Name: "secondElement", Kind: apRelative, Value: feltInt64(22)},
					{Name: "thirdElement", Kind: apRelative, Value: feltInt64(33)},
					{Name: "fourthElement", Kind: apRelative, Value: feltInt64(44)},
					{Name: "fifthElement", Kind: apRelative, Value: feltInt64(55)},
					{Name: "sixthElement", Kind: apRelative, Value: feltInt64(66)},
					{Name: "index", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSearchSortedLowerHint(
						ctx.operanders["array_ptr"],
						ctx.operanders["elm_size"],
						ctx.operanders["n_elms"],
						ctx.operanders["key"],
						ctx.operanders["index"],
					)
				},
				check: varValueEquals("index", feltInt64(4)),
			},
			{
				operanders: []*hintOperander{
					{Name: "array_ptr", Kind: fpRelative, Value: addr(8)},
					{Name: "elm_size", Kind: fpRelative, Value: feltInt64(1)},
					{Name: "n_elms", Kind: fpRelative, Value: feltInt64(6)},
					{Name: "key", Kind: fpRelative, Value: feltInt64(67)},
					{Name: "firstElement", Kind: apRelative, Value: feltInt64(11)},
					{Name: "secondElement", Kind: apRelative, Value: feltInt64(22)},
					{Name: "thirdElement", Kind: apRelative, Value: feltInt64(33)},
					{Name: "fourthElement", Kind: apRelative, Value: feltInt64(44)},
					{Name: "fifthElement", Kind: apRelative, Value: feltInt64(55)},
					{Name: "sixthElement", Kind: apRelative, Value: feltInt64(66)},
					{Name: "index", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSearchSortedLowerHint(
						ctx.operanders["array_ptr"],
						ctx.operanders["elm_size"],
						ctx.operanders["n_elms"],
						ctx.operanders["key"],
						ctx.operanders["index"],
					)
				},
				check: varValueEquals("index", feltInt64(6)),
			},
			{
				operanders: []*hintOperander{
					{Name: "array_ptr", Kind: fpRelative, Value: addr(8)},
					{Name: "elm_size", Kind: fpRelative, Value: feltInt64(2)},
					{Name: "n_elms", Kind: fpRelative, Value: feltInt64(3)},
					{Name: "key", Kind: fpRelative, Value: feltInt64(43)},
					{Name: "firstElement", Kind: apRelative, Value: feltInt64(11)},
					{Name: "secondElement", Kind: apRelative, Value: feltInt64(22)},
					{Name: "thirdElement", Kind: apRelative, Value: feltInt64(33)},
					{Name: "fourthElement", Kind: apRelative, Value: feltInt64(44)},
					{Name: "fifthElement", Kind: apRelative, Value: feltInt64(55)},
					{Name: "sixthElement", Kind: apRelative, Value: feltInt64(66)},
					{Name: "index", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSearchSortedLowerHint(
						ctx.operanders["array_ptr"],
						ctx.operanders["elm_size"],
						ctx.operanders["n_elms"],
						ctx.operanders["key"],
						ctx.operanders["index"],
					)
				},
				check: varValueEquals("index", feltInt64(2)),
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
		"SetAdd": {
			{
				operanders: []*hintOperander{
					{Name: "elm_size", Kind: apRelative, Value: feltUint64(0)},
					{Name: "elm_ptr", Kind: apRelative, Value: addrWithSegment(1, 0)},
					{Name: "set_ptr", Kind: apRelative, Value: addrWithSegment(1, 0)},
					{Name: "set_end_ptr", Kind: apRelative, Value: addrWithSegment(1, 0)},
					{Name: "index", Kind: uninitialized},
					{Name: "is_elm_in_set", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSetAddHint(
						ctx.operanders["elm_size"],
						ctx.operanders["elm_ptr"],
						ctx.operanders["set_ptr"],
						ctx.operanders["set_end_ptr"],
						ctx.operanders["index"],
						ctx.operanders["is_elm_in_set"],
					)
				},
				errCheck: errorTextContains("assert ids.elm_size > 0 failed"),
			},
			{
				operanders: []*hintOperander{
					{Name: "elm_size", Kind: apRelative, Value: feltUint64(1)},
					{Name: "elm_ptr", Kind: apRelative, Value: addrWithSegment(1, 0)},
					{Name: "set_ptr", Kind: apRelative, Value: addrWithSegment(1, 1)},
					{Name: "set_end_ptr", Kind: apRelative, Value: addrWithSegment(1, 0)},
					{Name: "index", Kind: uninitialized},
					{Name: "is_elm_in_set", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSetAddHint(
						ctx.operanders["elm_size"],
						ctx.operanders["elm_ptr"],
						ctx.operanders["set_ptr"],
						ctx.operanders["set_end_ptr"],
						ctx.operanders["index"],
						ctx.operanders["is_elm_in_set"],
					)
				},
				errCheck: errorTextContains("assert ids.set_ptr <= ids.set_end_ptr failed"),
			},
			{
				operanders: []*hintOperander{
					{Name: "elm.1", Kind: apRelative, Value: feltUint64(1)},
					{Name: "elm.2", Kind: apRelative, Value: feltUint64(2)},
					{Name: "elm.3", Kind: apRelative, Value: feltUint64(3)},
					{Name: "elm.4", Kind: apRelative, Value: feltUint64(4)},
					{Name: "set.1", Kind: apRelative, Value: feltUint64(5)},
					{Name: "set.2", Kind: apRelative, Value: feltUint64(6)},
					{Name: "set.3", Kind: apRelative, Value: feltUint64(7)},
					{Name: "set.4", Kind: apRelative, Value: feltUint64(8)},
					{Name: "set.5", Kind: apRelative, Value: feltUint64(9)},
					{Name: "set.6", Kind: apRelative, Value: feltUint64(10)},
					{Name: "set.7", Kind: apRelative, Value: feltUint64(11)},
					{Name: "set.8", Kind: apRelative, Value: feltUint64(12)},
					{Name: "set.9", Kind: apRelative, Value: feltUint64(1)},
					{Name: "set.10", Kind: apRelative, Value: feltUint64(2)},
					{Name: "set.11", Kind: apRelative, Value: feltUint64(3)},
					{Name: "set.12", Kind: apRelative, Value: feltUint64(4)},
					{Name: "set.13", Kind: apRelative, Value: feltUint64(13)},
					{Name: "set.14", Kind: apRelative, Value: feltUint64(14)},
					{Name: "set.15", Kind: apRelative, Value: feltUint64(15)},
					{Name: "set.16", Kind: apRelative, Value: feltUint64(16)},
					{Name: "elm_size", Kind: apRelative, Value: feltUint64(4)},
					{Name: "elm_ptr", Kind: apRelative, Value: addrWithSegment(1, 4)},
					{Name: "set_ptr", Kind: apRelative, Value: addrWithSegment(1, 8)},
					{Name: "set_end_ptr", Kind: apRelative, Value: addrWithSegment(1, 24)},
					{Name: "index", Kind: uninitialized},
					{Name: "is_elm_in_set", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSetAddHint(
						ctx.operanders["elm_size"],
						ctx.operanders["elm_ptr"],
						ctx.operanders["set_ptr"],
						ctx.operanders["set_end_ptr"],
						ctx.operanders["index"],
						ctx.operanders["is_elm_in_set"],
					)
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"index":         feltUint64(2),
					"is_elm_in_set": feltUint64(1),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "elm.1", Kind: apRelative, Value: feltUint64(1)},
					{Name: "elm.2", Kind: apRelative, Value: feltUint64(2)},
					{Name: "elm.3", Kind: apRelative, Value: feltUint64(3)},
					{Name: "elm.4", Kind: apRelative, Value: feltUint64(4)},
					{Name: "set.1", Kind: apRelative, Value: feltUint64(5)},
					{Name: "set.2", Kind: apRelative, Value: feltUint64(6)},
					{Name: "set.3", Kind: apRelative, Value: feltUint64(7)},
					{Name: "set.4", Kind: apRelative, Value: feltUint64(8)},
					{Name: "set.5", Kind: apRelative, Value: feltUint64(9)},
					{Name: "set.6", Kind: apRelative, Value: feltUint64(10)},
					{Name: "set.7", Kind: apRelative, Value: feltUint64(11)},
					{Name: "set.8", Kind: apRelative, Value: feltUint64(12)},
					{Name: "set.9", Kind: apRelative, Value: feltUint64(13)},
					{Name: "set.10", Kind: apRelative, Value: feltUint64(14)},
					{Name: "set.11", Kind: apRelative, Value: feltUint64(15)},
					{Name: "set.12", Kind: apRelative, Value: feltUint64(16)},
					{Name: "set.13", Kind: apRelative, Value: feltUint64(17)},
					{Name: "set.14", Kind: apRelative, Value: feltUint64(18)},
					{Name: "set.15", Kind: apRelative, Value: feltUint64(19)},
					{Name: "set.16", Kind: apRelative, Value: feltUint64(20)},
					{Name: "elm_size", Kind: apRelative, Value: feltUint64(4)},
					{Name: "elm_ptr", Kind: apRelative, Value: addrWithSegment(1, 4)},
					{Name: "set_ptr", Kind: apRelative, Value: addrWithSegment(1, 8)},
					{Name: "set_end_ptr", Kind: apRelative, Value: addrWithSegment(1, 24)},
					{Name: "index", Kind: uninitialized},
					{Name: "is_elm_in_set", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSetAddHint(
						ctx.operanders["elm_size"],
						ctx.operanders["elm_ptr"],
						ctx.operanders["set_ptr"],
						ctx.operanders["set_end_ptr"],
						ctx.operanders["index"],
						ctx.operanders["is_elm_in_set"],
					)
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"is_elm_in_set": feltUint64(0),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "elm.1", Kind: apRelative, Value: feltUint64(1)},
					{Name: "elm.2", Kind: apRelative, Value: feltUint64(2)},
					{Name: "elm.3", Kind: apRelative, Value: feltUint64(3)},
					{Name: "elm.4", Kind: apRelative, Value: feltUint64(4)},
					{Name: "elm.5", Kind: apRelative, Value: feltUint64(5)},
					{Name: "set.1", Kind: apRelative, Value: feltUint64(1)},
					{Name: "set.2", Kind: apRelative, Value: feltUint64(2)},
					{Name: "set.3", Kind: apRelative, Value: feltUint64(3)},
					{Name: "set.4", Kind: apRelative, Value: feltUint64(4)},
					{Name: "set.5", Kind: apRelative, Value: feltUint64(5)},
					{Name: "elm_size", Kind: apRelative, Value: feltUint64(5)},
					{Name: "elm_ptr", Kind: apRelative, Value: addrWithSegment(1, 4)},
					{Name: "set_ptr", Kind: apRelative, Value: addrWithSegment(1, 9)},
					{Name: "set_end_ptr", Kind: apRelative, Value: addrWithSegment(1, 14)},
					{Name: "index", Kind: uninitialized},
					{Name: "is_elm_in_set", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSetAddHint(
						ctx.operanders["elm_size"],
						ctx.operanders["elm_ptr"],
						ctx.operanders["set_ptr"],
						ctx.operanders["set_end_ptr"],
						ctx.operanders["index"],
						ctx.operanders["is_elm_in_set"],
					)
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"index":         feltUint64(0),
					"is_elm_in_set": feltUint64(1),
				}),
			},
		},
	})
}
