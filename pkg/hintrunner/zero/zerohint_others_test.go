package zero

import (
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
	})
}
