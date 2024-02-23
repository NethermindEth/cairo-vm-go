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
					{Name: "b", Kind: apRelative, Value: feltInt64(0)},
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
					{Name: "a", Kind: apRelative, Value: feltInt64(0)},
					{Name: "b", Kind: apRelative, Value: feltInt64(0)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newAssertLtFeltHint(ctx.operanders["a"], ctx.operanders["b"])
				},
				errCheck: errorTextContains("a = 0 is not less than b = 0"),
			},

			{
				operanders: []*hintOperander{
					{Name: "a", Kind: immediate, Value: feltInt64(1)},
					{Name: "b", Kind: apRelative, Value: feltInt64(0)},
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
					{Name: "b", Kind: immediate, Value: feltInt64(14)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newAssertLtFeltHint(ctx.operanders["a"], ctx.operanders["b"])
				},
				errCheck: errorTextContains("a = -10 is not less than b = 14"),
			},

			{
				operanders: []*hintOperander{
					{Name: "a", Kind: fpRelative, Value: feltInt64(1)},
					{Name: "b", Kind: fpRelative, Value: feltInt64(10)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newAssertLtFeltHint(ctx.operanders["a"], ctx.operanders["b"])
				},
				errCheck: errorIsNil,
			},
		},

		"IsNN": {
			// is_nn would return 1 for non-negative values, but the
			// hint itself writes 0 in this case (it's used as a jump condition).

			{
				operanders: []*hintOperander{
					{Name: "a", Kind: apRelative, Value: feltInt64(2421)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newIsNNHint(ctx.operanders["a"])
				},
				check: apValueEquals(feltUint64(0)),
			},

			{
				operanders: []*hintOperander{
					{Name: "a", Kind: apRelative, Value: feltInt64(0)},
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
					// (-a - 1) => (1 - 1) => 0
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
	})
}
