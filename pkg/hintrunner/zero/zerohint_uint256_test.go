package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func TestZeroHintUint256(t *testing.T) {
	runHinterTests(t, map[string][]hintTestCase{
		"Uint128Add": {
			{
				operanders: []*hintOperander{
					{Name: "a", Kind: fpRelative, Value: &utils.Felt127},
					{Name: "b", Kind: apRelative, Value: &utils.Felt127},
					{Name: "carry", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint128AddHint(ctx.operanders["a"], ctx.operanders["b"], ctx.operanders["carry"])
				},
				check: varValueEquals("carry", feltUint64(1)),
			},
			{
				operanders: []*hintOperander{
					{Name: "a", Kind: fpRelative, Value: feltUint64(0)},
					{Name: "b", Kind: apRelative, Value: feltUint64(0)},
					{Name: "carry", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint128AddHint(ctx.operanders["a"], ctx.operanders["b"], ctx.operanders["carry"])
				},
				check: varValueEquals("carry", feltUint64(0)),
			},
			{
				operanders: []*hintOperander{
					{Name: "a", Kind: fpRelative, Value: &utils.Felt127},
					{Name: "b", Kind: apRelative, Value: feltUint64(0)},
					{Name: "carry", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint128AddHint(ctx.operanders["a"], ctx.operanders["b"], ctx.operanders["carry"])
				},
				check: varValueEquals("carry", feltUint64(0)),
			},
			{
				operanders: []*hintOperander{
					{Name: "a", Kind: fpRelative, Value: feltUint64(0)},
					{Name: "b", Kind: apRelative, Value: &utils.Felt127},
					{Name: "carry", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint128AddHint(ctx.operanders["a"], ctx.operanders["b"], ctx.operanders["carry"])
				},
				check: varValueEquals("carry", feltUint64(0)),
			},
		},
		"Uint256Add": {
			{
				operanders: []*hintOperander{
					{Name: "a.low", Kind: fpRelative, Value: &utils.Felt127},
					{Name: "a.high", Kind: fpRelative, Value: feltUint64(0)},
					{Name: "b.low", Kind: apRelative, Value: &utils.Felt127},
					{Name: "b.high", Kind: apRelative, Value: feltUint64(0)},
					{Name: "carry_low", Kind: uninitialized},
					{Name: "carry_high", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint256AddHint(ctx.operanders["a.low"], ctx.operanders["b.low"], ctx.operanders["carry_low"], ctx.operanders["carry_high"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"carry_low":  feltUint64(1),
					"carry_high": feltUint64(0),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "a.low", Kind: fpRelative, Value: &utils.Felt127},
					{Name: "a.high", Kind: fpRelative, Value: &utils.Felt127},
					{Name: "b.low", Kind: apRelative, Value: &utils.Felt127},
					{Name: "b.high", Kind: apRelative, Value: &utils.Felt127},
					{Name: "carry_low", Kind: uninitialized},
					{Name: "carry_high", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint256AddHint(ctx.operanders["a.low"], ctx.operanders["b.low"], ctx.operanders["carry_low"], ctx.operanders["carry_high"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"carry_low":  feltUint64(1),
					"carry_high": feltUint64(1),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "a.low", Kind: fpRelative, Value: feltUint64(0)},
					{Name: "a.high", Kind: fpRelative, Value: &utils.Felt127},
					{Name: "b.low", Kind: apRelative, Value: feltUint64(0)},
					{Name: "b.high", Kind: apRelative, Value: &utils.Felt127},
					{Name: "carry_low", Kind: uninitialized},
					{Name: "carry_high", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint256AddHint(ctx.operanders["a.low"], ctx.operanders["b.low"], ctx.operanders["carry_low"], ctx.operanders["carry_high"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"carry_low":  feltUint64(0),
					"carry_high": feltUint64(1),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "a.low", Kind: fpRelative, Value: &utils.Felt127},
					{Name: "a.high", Kind: fpRelative, Value: feltUint64(0)},
					{Name: "b.low", Kind: apRelative, Value: feltUint64(0)},
					{Name: "b.high", Kind: apRelative, Value: &utils.Felt127},
					{Name: "carry_low", Kind: uninitialized},
					{Name: "carry_high", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint256AddHint(ctx.operanders["a.low"], ctx.operanders["b.low"], ctx.operanders["carry_low"], ctx.operanders["carry_high"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"carry_low":  feltUint64(0),
					"carry_high": feltUint64(0),
				}),
			},
		},
		"Split64": {
			// `high` is zero
			{
				operanders: []*hintOperander{
					{Name: "a", Kind: fpRelative, Value: feltUint64(8746)},
					{Name: "low", Kind: uninitialized},
					{Name: "high", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSplit64Hint(ctx.operanders["a"], ctx.operanders["low"], ctx.operanders["high"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"low":  feltUint64(8746),
					"high": feltUint64(0),
				}),
			},
			// `low` is zero
			{
				operanders: []*hintOperander{
					{Name: "a", Kind: fpRelative, Value: &utils.Felt127},
					{Name: "low", Kind: uninitialized},
					{Name: "high", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSplit64Hint(ctx.operanders["a"], ctx.operanders["low"], ctx.operanders["high"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"low":  feltUint64(0),
					"high": feltUint64(1 << 63),
				}),
			},
			// `high` is a felt that doesn't fit in uint64
			{
				operanders: []*hintOperander{
					{Name: "a", Kind: fpRelative, Value: feltInt64(-5)},
					{Name: "low", Kind: uninitialized},
					{Name: "high", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSplit64Hint(ctx.operanders["a"], ctx.operanders["low"], ctx.operanders["high"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"low":  feltUint64(18446744073709551612),                                        // felt(-5) & ((1 << 64) - 1)
					"high": feltString("196159429230833779654668657131193454380566933979560673279"), // felt(-5) >> 64
				}),
			},
		},
		"Uint256Sqrt": {
			{
				operanders: []*hintOperander{
					{Name: "n.low", Kind: apRelative, Value: feltInt64(0)},
					{Name: "n.high", Kind: apRelative, Value: feltInt64(0)},
					{Name: "root.low", Kind: uninitialized},
					{Name: "root.high", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint256SqrtHint(ctx.operanders["n.low"], ctx.operanders["root.low"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"root.low":  feltUint64(0),
					"root.high": feltUint64(0),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "n.low", Kind: apRelative, Value: feltInt64(5)},
					{Name: "n.high", Kind: apRelative, Value: feltInt64(0)},
					{Name: "root.low", Kind: uninitialized},
					{Name: "root.high", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint256SqrtHint(ctx.operanders["n.low"], ctx.operanders["root.low"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"root.low":  feltUint64(2),
					"root.high": feltUint64(0),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "n.low", Kind: fpRelative, Value: feltInt64(65536)},
					{Name: "n.high", Kind: fpRelative, Value: feltInt64(0)},
					{Name: "root.low", Kind: uninitialized},
					{Name: "root.high", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint256SqrtHint(ctx.operanders["n.low"], ctx.operanders["root.low"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"root.low":  feltUint64(256),
					"root.high": feltUint64(0),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "n.low", Kind: fpRelative, Value: &utils.Felt127},
					{Name: "n.high", Kind: fpRelative, Value: &utils.Felt127},
					{Name: "root.low", Kind: uninitialized},
					{Name: "root.high", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint256SqrtHint(ctx.operanders["n.low"], ctx.operanders["root.low"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"root.low":  feltString("240615969168004511545033772477625056927"),
					"root.high": feltUint64(0),
				}),
			},
		},
		"Uint256SignedNN": {
			{
				operanders: []*hintOperander{
					{Name: "a.low", Kind: fpRelative, Value: feltUint64(0)},
					{Name: "a.high", Kind: fpRelative, Value: &utils.Felt127},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint256SignedNNHint(ctx.operanders["a.low"])
				},
				check: apValueEquals(feltUint64(0)),
			},
			{
				operanders: []*hintOperander{
					{Name: "a.low", Kind: fpRelative, Value: feltUint64(0)},
					{Name: "a.high", Kind: fpRelative, Value: feltInt64(1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint256SignedNNHint(ctx.operanders["a.low"])
				},
				check: apValueEquals(feltUint64(1)),
			},
		},
		"Uint256UnsignedDivRem": {
			{
				operanders: []*hintOperander{
					{Name: "a.low", Kind: fpRelative, Value: feltUint64(6)},
					{Name: "a.high", Kind: fpRelative, Value: feltUint64(0)},
					{Name: "div.low", Kind: fpRelative, Value: feltUint64(2)},
					{Name: "div.high", Kind: fpRelative, Value: feltUint64(0)},
					{Name: "quotient.low", Kind: uninitialized},
					{Name: "quotient.high", Kind: uninitialized},
					{Name: "remainder.low", Kind: uninitialized},
					{Name: "remainder.high", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint256UnsignedDivRemHint(ctx.operanders["a.low"], ctx.operanders["div.low"], ctx.operanders["quotient.low"], ctx.operanders["remainder.low"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"quotient.low":   feltUint64(3),
					"quotient.high":  feltUint64(0),
					"remainder.low":  feltUint64(0),
					"remainder.high": feltUint64(0),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "a.low", Kind: fpRelative, Value: &utils.Felt127},
					{Name: "a.high", Kind: fpRelative, Value: feltUint64(0)},
					{Name: "div.low", Kind: fpRelative, Value: &utils.Felt127},
					{Name: "div.high", Kind: fpRelative, Value: feltUint64(0)},
					{Name: "quotient.low", Kind: uninitialized},
					{Name: "quotient.high", Kind: uninitialized},
					{Name: "remainder.low", Kind: uninitialized},
					{Name: "remainder.high", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint256UnsignedDivRemHint(ctx.operanders["a.low"], ctx.operanders["div.low"], ctx.operanders["quotient.low"], ctx.operanders["remainder.low"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"quotient.low":   feltUint64(1),
					"quotient.high":  feltUint64(0),
					"remainder.low":  feltUint64(0),
					"remainder.high": feltUint64(0),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "a.low", Kind: fpRelative, Value: feltUint64(5)},
					{Name: "a.high", Kind: fpRelative, Value: &utils.Felt127},
					{Name: "div.low", Kind: fpRelative, Value: feltUint64(0)},
					{Name: "div.high", Kind: fpRelative, Value: &utils.Felt127},
					{Name: "quotient.low", Kind: uninitialized},
					{Name: "quotient.high", Kind: uninitialized},
					{Name: "remainder.low", Kind: uninitialized},
					{Name: "remainder.high", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint256UnsignedDivRemHint(ctx.operanders["a.low"], ctx.operanders["div.low"], ctx.operanders["quotient.low"], ctx.operanders["remainder.low"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"quotient.low":   feltUint64(1),
					"quotient.high":  feltUint64(0),
					"remainder.low":  feltUint64(5),
					"remainder.high": feltUint64(0),
				}),
			},
		},
		"Uint256MulDivMod": {
			{
				operanders: []*hintOperander{
					{Name: "a.low", Kind: apRelative, Value: feltUint64(6)},
					{Name: "a.high", Kind: apRelative, Value: feltUint64(0)},
					{Name: "b.low", Kind: apRelative, Value: feltUint64(6)},
					{Name: "b.high", Kind: apRelative, Value: feltUint64(0)},
					{Name: "div.low", Kind: apRelative, Value: feltUint64(2)},
					{Name: "div.high", Kind: apRelative, Value: feltUint64(0)},
					{Name: "quotient_low.low", Kind: uninitialized},
					{Name: "quotient_low.high", Kind: uninitialized},
					{Name: "quotient_high.low", Kind: uninitialized},
					{Name: "quotient_high.high", Kind: uninitialized},
					{Name: "remainder.low", Kind: uninitialized},
					{Name: "remainder.high", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint256MulDivModHint(ctx.operanders["a.low"], ctx.operanders["b.low"], ctx.operanders["div.low"], ctx.operanders["quotient_low.low"], ctx.operanders["quotient_high.low"], ctx.operanders["remainder.low"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"quotient_low.low":   feltUint64(18),
					"quotient_low.high":  feltUint64(0),
					"quotient_high.low":  feltUint64(0),
					"quotient_high.high": feltUint64(0),
					"remainder.low":      feltUint64(0),
					"remainder.high":     feltUint64(0),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "a.low", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "a.high", Kind: apRelative, Value: feltString("2")},
					{Name: "b.low", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "b.high", Kind: apRelative, Value: feltString("3")},
					{Name: "div.low", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "div.high", Kind: apRelative, Value: feltString("2")},
					{Name: "quotient_low.low", Kind: uninitialized},
					{Name: "quotient_low.high", Kind: uninitialized},
					{Name: "quotient_high.low", Kind: uninitialized},
					{Name: "quotient_high.high", Kind: uninitialized},
					{Name: "remainder.low", Kind: uninitialized},
					{Name: "remainder.high", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint256MulDivModHint(ctx.operanders["a.low"], ctx.operanders["b.low"], ctx.operanders["div.low"], ctx.operanders["quotient_low.low"], ctx.operanders["quotient_high.low"], ctx.operanders["remainder.low"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"quotient_low.low":   &utils.FeltZero,
					"quotient_low.high":  feltUint64(3),
					"quotient_high.low":  &utils.FeltZero,
					"quotient_high.high": &utils.FeltZero,
					"remainder.low":      &utils.FeltZero,
					"remainder.high":     &utils.FeltZero,
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "a.low", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "a.high", Kind: apRelative, Value: &utils.Felt127},
					{Name: "b.low", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "b.high", Kind: apRelative, Value: &utils.Felt127},
					{Name: "div.low", Kind: apRelative, Value: feltString("85070591730234615865843651857942052864")},
					{Name: "div.high", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "quotient_low.low", Kind: uninitialized},
					{Name: "quotient_low.high", Kind: uninitialized},
					{Name: "quotient_high.low", Kind: uninitialized},
					{Name: "quotient_high.high", Kind: uninitialized},
					{Name: "remainder.low", Kind: uninitialized},
					{Name: "remainder.high", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint256MulDivModHint(ctx.operanders["a.low"], ctx.operanders["b.low"], ctx.operanders["div.low"], ctx.operanders["quotient_low.low"], ctx.operanders["quotient_high.low"], ctx.operanders["remainder.low"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"quotient_low.low":   &utils.FeltZero,
					"quotient_low.high":  &utils.FeltZero,
					"quotient_high.low":  &utils.FeltZero,
					"quotient_high.high": feltInt64(1),
					"remainder.low":      &utils.FeltZero,
					"remainder.high":     &utils.FeltZero,
				}),
			},
		},
		"Uint256Sub": {
			{
				// equal values
				operanders: []*hintOperander{
					{Name: "a.low", Kind: apRelative, Value: feltUint64(11)},
					{Name: "a.high", Kind: apRelative, Value: feltUint64(11)},
					{Name: "b.low", Kind: apRelative, Value: feltUint64(11)},
					{Name: "b.high", Kind: apRelative, Value: feltUint64(11)},
					{Name: "res.low", Kind: uninitialized},
					{Name: "res.high", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint256SubHint(ctx.operanders["a.low"], ctx.operanders["b.low"], ctx.operanders["res.low"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"res.low":  feltUint64(0),
					"res.high": feltUint64(0),
				}),
			},
			{
				// random values
				operanders: []*hintOperander{
					{Name: "a.low", Kind: apRelative, Value: feltUint64(12345)},
					{Name: "a.high", Kind: apRelative, Value: feltUint64(67890)},
					{Name: "b.low", Kind: apRelative, Value: feltUint64(11111)},
					{Name: "b.high", Kind: apRelative, Value: feltUint64(22222)},
					{Name: "res.low", Kind: uninitialized},
					{Name: "res.high", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint256SubHint(ctx.operanders["a.low"], ctx.operanders["b.low"], ctx.operanders["res.low"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"res.low":  feltUint64(1234),
					"res.high": feltUint64(45668),
				}),
			},
			{
				// smaller value - bigger value
				operanders: []*hintOperander{
					{Name: "a.low", Kind: apRelative, Value: feltUint64(0)},
					{Name: "a.high", Kind: apRelative, Value: feltUint64(1)},
					{Name: "b.low", Kind: apRelative, Value: feltUint64(1)},
					{Name: "b.high", Kind: apRelative, Value: feltUint64(0)},
					{Name: "res.low", Kind: uninitialized},
					{Name: "res.high", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint256SubHint(ctx.operanders["a.low"], ctx.operanders["b.low"], ctx.operanders["res.low"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"res.low":  feltString("340282366920938463463374607431768211455"),
					"res.high": feltUint64(0),
				}),
			},
		},
		"SplitXX": {
			{
				operanders: []*hintOperander{
					{Name: "x.low", Kind: uninitialized},
					{Name: "x.high", Kind: uninitialized},
					{Name: "xx.low", Kind: apRelative, Value: feltString("19681161376707505956807079304988542015446066515923890162744021073123829784752")},
					{Name: "xx.high", Kind: apRelative, Value: feltUint64(1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSplitXXHint(ctx.operanders["x.low"], ctx.operanders["xx.low"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"x.low":  feltString("31967510603135814718935067919817634170"),
					"x.high": feltString("136383360823982924942210781858056431457"),
				}),
			},
		},
	})
}
