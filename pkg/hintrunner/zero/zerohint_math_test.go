package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
)

func TestZeroHintMath(t *testing.T) {
	runHinterTests(t, map[string][]hintTestCase{
		"IsLeFelt": {
			{
				operanders: []*hintOperander{
					{Name: "a", Kind: apRelative, Value: feltUint64(0)},
					{Name: "b", Kind: apRelative, Value: feltUint64(0)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newIsLeFeltHint(ctx.operanders["a"], ctx.operanders["b"])
				},
				check: apValueEquals(feltUint64(0)),
			},

			{
				operanders: []*hintOperander{
					{Name: "a", Kind: apRelative, Value: feltUint64(1)},
					{Name: "b", Kind: immediate, Value: feltUint64(0)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newIsLeFeltHint(ctx.operanders["a"], ctx.operanders["b"])
				},
				check: apValueEquals(feltUint64(1)),
			},

			{
				operanders: []*hintOperander{
					{Name: "a", Kind: apRelative, Value: feltUint64(0)},
					{Name: "b", Kind: apRelative, Value: feltUint64(1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newIsLeFeltHint(ctx.operanders["a"], ctx.operanders["b"])
				},
				check: apValueEquals(feltUint64(0)),
			},
		},

		"AssertLtFelt": {
			{
				operanders: []*hintOperander{
					{Name: "a", Kind: apRelative, Value: feltUint64(0)},
					{Name: "b", Kind: apRelative, Value: feltUint64(0)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newAssertLtFeltHint(ctx.operanders["a"], ctx.operanders["b"])
				},
				errCheck: errorTextContains("a = 0 is not less than b = 0"),
			},

			{
				operanders: []*hintOperander{
					{Name: "a", Kind: immediate, Value: feltUint64(1)},
					{Name: "b", Kind: apRelative, Value: feltUint64(0)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newAssertLtFeltHint(ctx.operanders["a"], ctx.operanders["b"])
				},
				errCheck: errorTextContains("a = 1 is not less than b = 0"),
			},

			{
				// -10 felt is 3618502788666131213697322783095070105623107215331596699973092056135872020467
				// and it will not be less than 14 in Cairo as well.
				operanders: []*hintOperander{
					{Name: "a", Kind: apRelative, Value: feltInt64(-10)},
					{Name: "b", Kind: immediate, Value: feltUint64(14)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newAssertLtFeltHint(ctx.operanders["a"], ctx.operanders["b"])
				},
				errCheck: errorTextContains("a = -10 is not less than b = 14"),
			},

			{
				operanders: []*hintOperander{
					{Name: "a", Kind: fpRelative, Value: feltUint64(1)},
					{Name: "b", Kind: fpRelative, Value: feltUint64(10)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newAssertLtFeltHint(ctx.operanders["a"], ctx.operanders["b"])
				},
				errCheck: errorIsNil,
			},
		},

		"AssertNotZero": {
			{
				operanders: []*hintOperander{
					{Name: "value", Kind: apRelative, Value: feltInt64(0)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newAssertNotZeroHint(ctx.operanders["value"])
				},
				errCheck: errorTextContains("assertion failed: value is zero"),
			},

			{
				operanders: []*hintOperander{
					{Name: "value", Kind: fpRelative, Value: feltInt64(1353)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newAssertNotZeroHint(ctx.operanders["value"])
				},
				errCheck: errorIsNil,
			},
		},

		"AssertNN": {
			// Like IsNN, but does an assertion instead.

			{
				operanders: []*hintOperander{
					{Name: "a", Kind: apRelative, Value: feltInt64(2421)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newAssertNNHint(ctx.operanders["a"])
				},
				errCheck: errorIsNil,
			},

			{
				operanders: []*hintOperander{
					{Name: "a", Kind: apRelative, Value: feltInt64(0)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newAssertNNHint(ctx.operanders["a"])
				},
				errCheck: errorIsNil,
			},

			{
				operanders: []*hintOperander{
					{Name: "a", Kind: apRelative, Value: feltInt64(-2)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newAssertNNHint(ctx.operanders["a"])
				},
				errCheck: errorTextContains("assertion failed: a = -2 is out of range"),
			},
		},

		"AssertNotEqual": {
			// Different address values.
			{
				operanders: []*hintOperander{
					{Name: "a", Kind: apRelative, Value: addr(0)},
					{Name: "b", Kind: apRelative, Value: addr(1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newAssertNotEqualHint(ctx.operanders["a"], ctx.operanders["b"])
				},
				errCheck: errorIsNil,
			},
			{
				operanders: []*hintOperander{
					{Name: "a", Kind: apRelative, Value: addrWithSegment(0, 1)},
					{Name: "b", Kind: apRelative, Value: addrWithSegment(1, 1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newAssertNotEqualHint(ctx.operanders["a"], ctx.operanders["b"])
				},
				errCheck: errorIsNil,
			},

			// Different felt values.
			{
				operanders: []*hintOperander{
					{Name: "a", Kind: apRelative, Value: feltInt64(1)},
					{Name: "b", Kind: apRelative, Value: feltInt64(-1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newAssertNotEqualHint(ctx.operanders["a"], ctx.operanders["b"])
				},
				errCheck: errorIsNil,
			},

			// Mismatching types.
			{
				operanders: []*hintOperander{
					{Name: "a", Kind: apRelative, Value: addr(0)},
					{Name: "b", Kind: apRelative, Value: feltInt64(0)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newAssertNotEqualHint(ctx.operanders["a"], ctx.operanders["b"])
				},
				errCheck: errorTextContains("assertion failed: non-comparable values: 1:0, 0"),
			},
			{
				operanders: []*hintOperander{
					{Name: "a", Kind: apRelative, Value: feltInt64(1)},
					{Name: "b", Kind: apRelative, Value: addr(1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newAssertNotEqualHint(ctx.operanders["a"], ctx.operanders["b"])
				},
				errCheck: errorTextContains("assertion failed: non-comparable values: 1, 1:1"),
			},

			// Same value addresses.
			{
				operanders: []*hintOperander{
					{Name: "a", Kind: apRelative, Value: addr(0)},
					{Name: "b", Kind: apRelative, Value: addr(0)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newAssertNotEqualHint(ctx.operanders["a"], ctx.operanders["b"])
				},
				errCheck: errorTextContains("assertion failed: 1:0 = 1:0"),
			},

			// Same value felts.
			{
				operanders: []*hintOperander{
					{Name: "a", Kind: apRelative, Value: feltInt64(-1)},
					{Name: "b", Kind: apRelative, Value: feltInt64(-1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newAssertNotEqualHint(ctx.operanders["a"], ctx.operanders["b"])
				},
				errCheck: errorTextContains("assertion failed: -1 = -1"),
			},
		},

		"IsNN": {
			// is_nn would return 1 for non-negative values, but the
			// hint itself writes 0 in this case (it's used as a jump condition).

			{
				operanders: []*hintOperander{
					{Name: "a", Kind: apRelative, Value: feltUint64(2421)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newIsNNHint(ctx.operanders["a"])
				},
				check: apValueEquals(feltUint64(0)),
			},

			{
				operanders: []*hintOperander{
					{Name: "a", Kind: apRelative, Value: feltUint64(0)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newIsNNHint(ctx.operanders["a"])
				},
				check: apValueEquals(feltUint64(0)),
			},

			{
				operanders: []*hintOperander{
					{Name: "a", Kind: apRelative, Value: feltInt64(-2)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newIsNNHint(ctx.operanders["a"])
				},
				check: apValueEquals(feltUint64(1)),
			},
		},

		"IsNNOutOfRange": {
			// Note that "a" is (-a - 1).

			{
				operanders: []*hintOperander{
					// (-a - 1) => (-1 - 1) => -2
					{Name: "a", Kind: apRelative, Value: feltInt64(1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newIsNNOutOfRangeHint(ctx.operanders["a"])
				},
				check: apValueEquals(feltUint64(1)),
			},

			{
				operanders: []*hintOperander{
					// (-a - 1) => (0 - 1) => -1
					{Name: "a", Kind: apRelative, Value: feltInt64(0)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newIsNNOutOfRangeHint(ctx.operanders["a"])
				},
				check: apValueEquals(feltUint64(1)),
			},

			{
				operanders: []*hintOperander{
					{Name: "a", Kind: apRelative, Value: feltInt64(-1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newIsNNOutOfRangeHint(ctx.operanders["a"])
				},
				check: apValueEquals(feltUint64(0)),
			},

			{
				operanders: []*hintOperander{
					{Name: "a", Kind: apRelative, Value: feltAdd(&utils.FeltMax128, feltInt64(1))},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newIsNNOutOfRangeHint(ctx.operanders["a"])
				},
				check: apValueEquals(feltUint64(1)),
			},

			{
				operanders: []*hintOperander{
					{Name: "a", Kind: apRelative, Value: feltAdd(&utils.FeltMax128, feltInt64(2))},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newIsNNOutOfRangeHint(ctx.operanders["a"])
				},
				check: apValueEquals(feltUint64(1)),
			},

			{
				operanders: []*hintOperander{
					{Name: "a", Kind: apRelative, Value: &utils.FeltMax128},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newIsNNOutOfRangeHint(ctx.operanders["a"])
				},
				check: apValueEquals(feltUint64(1)),
			},
		},

		"IsPositive": {
			{
				operanders: []*hintOperander{
					{Name: "value", Kind: apRelative, Value: feltInt64(10)},
					{Name: "is_positive", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newIsPositiveHint(ctx.operanders["value"], ctx.operanders["is_positive"])
				},
				check: varValueEquals("is_positive", feltInt64(1)),
			},

			{
				operanders: []*hintOperander{
					{Name: "value", Kind: apRelative, Value: feltInt64(0)},
					{Name: "is_positive", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newIsPositiveHint(ctx.operanders["value"], ctx.operanders["is_positive"])
				},
				check: varValueEquals("is_positive", feltInt64(1)),
			},

			{
				operanders: []*hintOperander{
					{Name: "value", Kind: apRelative, Value: feltInt64(-1)},
					{Name: "is_positive", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newIsPositiveHint(ctx.operanders["value"], ctx.operanders["is_positive"])
				},
				check: varValueEquals("is_positive", feltInt64(0)),
			},

			{
				operanders: []*hintOperander{
					{Name: "value", Kind: apRelative, Value: feltAdd(&utils.FeltMax128, feltInt64(-1))},
					{Name: "is_positive", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newIsPositiveHint(ctx.operanders["value"], ctx.operanders["is_positive"])
				},
				check: varValueEquals("is_positive", feltInt64(1)),
			},

			{
				operanders: []*hintOperander{
					{Name: "value", Kind: apRelative, Value: &utils.FeltMax128},
					{Name: "is_positive", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newIsPositiveHint(ctx.operanders["value"], ctx.operanders["is_positive"])
				},
				check: varValueEquals("is_positive", feltInt64(0)),
			},

			{
				operanders: []*hintOperander{
					{Name: "value", Kind: apRelative, Value: feltAdd(&utils.FeltMax128, feltInt64(1))},
					{Name: "is_positive", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newIsPositiveHint(ctx.operanders["value"], ctx.operanders["is_positive"])
				},
				check: varValueEquals("is_positive", feltInt64(0)),
			},
		},

		"SplitIntAssertRange": {
			// Zero value - assertion failed, any other - nothing.

			{
				operanders: []*hintOperander{
					{Name: "value", Kind: apRelative, Value: feltInt64(1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSplitIntAssertRangeHint(ctx.operanders["value"])
				},
				errCheck: errorTextContains("split_int(): value is out of range"),
			},

			{
				operanders: []*hintOperander{
					{Name: "value", Kind: fpRelative, Value: feltInt64(0)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSplitIntAssertRangeHint(ctx.operanders["value"])
				},
				errCheck: errorIsNil,
			},
		},

		"SplitIntHint": {
			// Performs value%base and stores that to output.
			// Asserts output<bound.

			{
				operanders: []*hintOperander{
					{Name: "output", Kind: uninitialized},
					{Name: "value", Kind: fpRelative, Value: feltInt64(15)},
					{Name: "base", Kind: fpRelative, Value: feltInt64(2)},
					{Name: "bound", Kind: fpRelative, Value: feltInt64(5)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSplitIntHint(ctx.operanders["output"], ctx.operanders["value"], ctx.operanders["base"], ctx.operanders["bound"])
				},
				check: varValueEquals("output", feltInt64(1)),
			},

			{
				operanders: []*hintOperander{
					{Name: "output", Kind: uninitialized},
					{Name: "value", Kind: fpRelative, Value: feltInt64(100)},
					{Name: "base", Kind: fpRelative, Value: feltInt64(10)},
					{Name: "bound", Kind: fpRelative, Value: feltInt64(1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSplitIntHint(ctx.operanders["output"], ctx.operanders["value"], ctx.operanders["base"], ctx.operanders["bound"])
				},
				check: varValueEquals("output", feltInt64(0)),
			},

			{
				operanders: []*hintOperander{
					{Name: "output", Kind: uninitialized},
					{Name: "value", Kind: fpRelative, Value: feltInt64(100)},
					{Name: "base", Kind: fpRelative, Value: feltInt64(6)},
					{Name: "bound", Kind: fpRelative, Value: feltInt64(7)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSplitIntHint(ctx.operanders["output"], ctx.operanders["value"], ctx.operanders["base"], ctx.operanders["bound"])
				},
				check: varValueEquals("output", feltInt64(4)),
			},

			{
				operanders: []*hintOperander{
					{Name: "output", Kind: uninitialized},
					{Name: "value", Kind: fpRelative, Value: feltInt64(100)},
					{Name: "base", Kind: fpRelative, Value: feltInt64(6)},
					{Name: "bound", Kind: fpRelative, Value: feltInt64(3)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSplitIntHint(ctx.operanders["output"], ctx.operanders["value"], ctx.operanders["base"], ctx.operanders["bound"])
				},
				errCheck: errorTextContains("assertion `split_int(): Limb 4 is out of range` failed"),
			},
		},
	})
}
