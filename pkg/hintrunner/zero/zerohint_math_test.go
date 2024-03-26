package zero

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/parsers/starknet"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
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

		"Assert250bits": {
			{
				operanders: []*hintOperander{
					{Name: "low", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 0)},
					{Name: "high", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 1)},
					{Name: "value", Kind: apRelative, Value: feltInt64(3042)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newAssert250bitsHint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["value"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"low":  feltInt64(3042),
					"high": feltInt64(0),
				}),
			},

			{
				operanders: []*hintOperander{
					{Name: "low", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 0)},
					{Name: "high", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 1)},
					{Name: "value", Kind: fpRelative, Value: feltInt64(4938538853994)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newAssert250bitsHint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["value"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"low":  feltInt64(4938538853994),
					"high": feltInt64(0),
				}),
			},

			{
				operanders: []*hintOperander{
					{Name: "low", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 0)},
					{Name: "high", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 1)},
					{Name: "value", Kind: apRelative, Value: feltString("348329493943842849393993999999231222222222")},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newAssert250bitsHint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["value"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"low":  feltString("220632583722801270961776596532341902734"),
					"high": feltInt64(1023),
				}),
			},

			{
				operanders: []*hintOperander{
					{Name: "low", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 0)},
					{Name: "high", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 1)},
					{Name: "value", Kind: apRelative, Value: feltString("348329493943842849393124453993999999231222222222")},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newAssert250bitsHint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["value"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"low":  feltString("302658603189151847763334509038790380942"),
					"high": feltInt64(1023648380),
				}),
			},

			{
				operanders: []*hintOperander{
					{Name: "low", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 0)},
					{Name: "high", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 1)},
					{Name: "value", Kind: apRelative, Value: feltInt64(-233)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newAssert250bitsHint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["value"])
				},
				errCheck: errorTextContains("outside of the range [0, 2**250)"),
			},
		},

		"SplitFelt": {
			{
				operanders: []*hintOperander{
					{Name: "low", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 0)},
					{Name: "high", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 1)},
					{Name: "value", Kind: apRelative, Value: feltString("100000000000000000000000000000000000000")},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSplitFeltHint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["value"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"low":  feltString("100000000000000000000000000000000000000"),
					"high": feltInt64(0),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "low", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 0)},
					{Name: "high", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 1)},
					{Name: "value", Kind: apRelative, Value: &utils.FeltMax128},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSplitFeltHint(ctx.operanders["low"], ctx.operanders["high"], ctx.operanders["value"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"low":  feltInt64(0),
					"high": feltInt64(1),
				}),
			},
		},
		"SignedDivRem": {
			{
				operanders: []*hintOperander{
					{Name: "value", Kind: apRelative, Value: feltString("0")},
					{Name: "div", Kind: apRelative, Value: &utils.FeltMax128},
					{Name: "bound", Kind: apRelative, Value: &utils.Felt127},
					{Name: "r", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 0)},
					{Name: "biased_q", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSignedDivRemHint(ctx.operanders["value"], ctx.operanders["div"], ctx.operanders["bound"], ctx.operanders["r"], ctx.operanders["biased_q"])
				},
				errCheck: errorTextContains(fmt.Sprintf("div=%v is out of the valid range.", &utils.FeltMax128)),
			},
			{
				operanders: []*hintOperander{
					{Name: "value", Kind: apRelative, Value: feltString("0")},
					{Name: "div", Kind: apRelative, Value: feltString("1")},
					{Name: "bound", Kind: apRelative, Value: new(fp.Element).SetBigInt(new(big.Int).Lsh(big.NewInt(1), 130))},
					{Name: "r", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 0)},
					{Name: "biased_q", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSignedDivRemHint(ctx.operanders["value"], ctx.operanders["div"], ctx.operanders["bound"], ctx.operanders["r"], ctx.operanders["biased_q"])
				},
				errCheck: errorTextContains(fmt.Sprintf("bound=%v is out of the valid range", new(fp.Element).SetBigInt(new(big.Int).Lsh(big.NewInt(1), 130)))),
			},
			{
				operanders: []*hintOperander{
					{Name: "value", Kind: apRelative, Value: feltString("4")},
					{Name: "div", Kind: apRelative, Value: feltString("2")},
					{Name: "bound", Kind: apRelative, Value: feltString("1")},
					{Name: "r", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 0)},
					{Name: "biased_q", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSignedDivRemHint(ctx.operanders["value"], ctx.operanders["div"], ctx.operanders["bound"], ctx.operanders["r"], ctx.operanders["biased_q"])
				},
				errCheck: errorTextContains("is out of the range"),
			},
			{
				operanders: []*hintOperander{
					{Name: "value", Kind: apRelative, Value: feltString("5")},
					{Name: "div", Kind: apRelative, Value: feltString("2")},
					{Name: "bound", Kind: apRelative, Value: &utils.Felt127},
					{Name: "r", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 0)},
					{Name: "biased_q", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSignedDivRemHint(ctx.operanders["value"], ctx.operanders["div"], ctx.operanders["bound"], ctx.operanders["r"], ctx.operanders["biased_q"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"r":        feltString("1"),
					"biased_q": new(fp.Element).Add(feltString("2"), &utils.Felt127),
				}),
			},
		},
		"SqrtHint": {
			{
				operanders: []*hintOperander{
					{Name: "root", Kind: uninitialized},
					{Name: "value", Kind: fpRelative, Value: feltInt64(25)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSqrtHint(ctx.operanders["root"], ctx.operanders["value"])
				},
				check: varValueEquals("root", feltInt64(5)),
			},
			{
				operanders: []*hintOperander{
					{Name: "root", Kind: uninitialized},
					{Name: "value", Kind: fpRelative, Value: feltInt64(0)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSqrtHint(ctx.operanders["root"], ctx.operanders["value"])
				},
				check: varValueEquals("root", feltInt64(0)),
			},
			{
				operanders: []*hintOperander{
					{Name: "root", Kind: uninitialized},
					{Name: "value", Kind: fpRelative, Value: feltInt64(50)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSqrtHint(ctx.operanders["root"], ctx.operanders["value"])
				},
				check: varValueEquals("root", feltInt64(7)),
			},
			{
				operanders: []*hintOperander{
					{Name: "root", Kind: uninitialized},
					{Name: "value", Kind: fpRelative, Value: feltInt64(-128)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSqrtHint(ctx.operanders["root"], ctx.operanders["value"])
				},
				errCheck: errorTextContains("outside of the range [0, 2**250)"),
			},
		},

		"UnsignedDivRem": {
			{
				operanders: []*hintOperander{
					{Name: "value", Kind: fpRelative, Value: feltUint64(100)},
					{Name: "div", Kind: fpRelative, Value: feltUint64(6)},
					{Name: "r", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 0)},
					{Name: "q", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUnsignedDivRemHinter(ctx.operanders["value"], ctx.operanders["div"], ctx.operanders["q"], ctx.operanders["r"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"q": feltInt64(16),
					"r": feltInt64(4),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "value", Kind: fpRelative, Value: feltUint64(450326666)},
					{Name: "div", Kind: fpRelative, Value: feltUint64(136310839)},
					{Name: "r", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 0)},
					{Name: "q", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUnsignedDivRemHinter(ctx.operanders["value"], ctx.operanders["div"], ctx.operanders["q"], ctx.operanders["r"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"q": feltInt64(3),
					"r": feltInt64(41394149),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "value", Kind: fpRelative, Value: feltUint64(0)},
					{Name: "div", Kind: fpRelative, Value: feltUint64(10)},
					{Name: "r", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 0)},
					{Name: "q", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUnsignedDivRemHinter(ctx.operanders["value"], ctx.operanders["div"], ctx.operanders["q"], ctx.operanders["r"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"q": feltInt64(0),
					"r": feltInt64(0),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "value", Kind: fpRelative, Value: feltUint64(10)},
					{Name: "div", Kind: fpRelative, Value: feltUint64(0)},
					{Name: "r", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 0)},
					{Name: "q", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUnsignedDivRemHinter(ctx.operanders["value"], ctx.operanders["div"], ctx.operanders["q"], ctx.operanders["r"])
				},
				errCheck: errorTextContains("div=0x0 is out of the valid range."),
			},
			{
				operanders: []*hintOperander{
					{Name: "value", Kind: fpRelative, Value: feltUint64(10)},
					{Name: "div", Kind: fpRelative, Value: feltString("10633823966279327296825105735305134079")},
					{Name: "r", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 0)},
					{Name: "q", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUnsignedDivRemHinter(ctx.operanders["value"], ctx.operanders["div"], ctx.operanders["q"], ctx.operanders["r"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"q": feltInt64(0),
					"r": feltInt64(10),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "value", Kind: fpRelative, Value: feltUint64(10)},
					{Name: "div", Kind: fpRelative, Value: feltString("10633823966279327296825105735305134080")},
					{Name: "r", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 0)},
					{Name: "q", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUnsignedDivRemHinter(ctx.operanders["value"], ctx.operanders["div"], ctx.operanders["q"], ctx.operanders["r"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"q": feltInt64(0),
					"r": feltInt64(10),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "value", Kind: fpRelative, Value: feltUint64(10)},
					{Name: "div", Kind: fpRelative, Value: feltString("10633823966279327296825105735305134081")},
					{Name: "r", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 0)},
					{Name: "q", Kind: reference, Value: addrBuiltin(starknet.RangeCheck, 1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUnsignedDivRemHinter(ctx.operanders["value"], ctx.operanders["div"], ctx.operanders["q"], ctx.operanders["r"])
				},
				errCheck: errorTextContains("div=0x8000000000000110000000000000001 is out of the valid range."),
			},
		},
	})
}
