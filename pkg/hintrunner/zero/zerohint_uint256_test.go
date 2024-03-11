package zero

import (
	"math/big"
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func TestZeroHintUint256(t *testing.T) {
	// Values used in the test cases
	// 1 << 127
	felt127 := new(fp.Element).SetBigInt(new(big.Int).Lsh(big.NewInt(1), 127))

	runHinterTests(t, map[string][]hintTestCase{
		"Uint256Add": {
			{
				operanders: []*hintOperander{
					{Name: "a.low", Kind: fpRelative, Value: felt127},
					{Name: "a.high", Kind: fpRelative, Value: feltUint64(0)},
					{Name: "b.low", Kind: apRelative, Value: felt127},
					{Name: "b.high", Kind: apRelative, Value: feltUint64(0)},
					{Name: "carry_low", Kind: uninitialized},
					{Name: "carry_high", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint256AddHint(ctx.operanders["a.low"], ctx.operanders["b.low"], ctx.operanders["carry_low"], ctx.operanders["carry_high"], false)
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"carry_low":  feltUint64(1),
					"carry_high": feltUint64(0),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "a.low", Kind: fpRelative, Value: felt127},
					{Name: "a.high", Kind: fpRelative, Value: felt127},
					{Name: "b.low", Kind: apRelative, Value: felt127},
					{Name: "b.high", Kind: apRelative, Value: felt127},
					{Name: "carry_low", Kind: uninitialized},
					{Name: "carry_high", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint256AddHint(ctx.operanders["a.low"], ctx.operanders["b.low"], ctx.operanders["carry_low"], ctx.operanders["carry_high"], false)
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"carry_low":  feltUint64(1),
					"carry_high": feltUint64(1),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "a.low", Kind: fpRelative, Value: feltUint64(0)},
					{Name: "a.high", Kind: fpRelative, Value: felt127},
					{Name: "b.low", Kind: apRelative, Value: feltUint64(0)},
					{Name: "b.high", Kind: apRelative, Value: felt127},
					{Name: "carry_low", Kind: uninitialized},
					{Name: "carry_high", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint256AddHint(ctx.operanders["a.low"], ctx.operanders["b.low"], ctx.operanders["carry_low"], ctx.operanders["carry_high"], false)
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"carry_low":  feltUint64(0),
					"carry_high": feltUint64(1),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "a.low", Kind: fpRelative, Value: felt127},
					{Name: "a.high", Kind: fpRelative, Value: feltUint64(0)},
					{Name: "b.low", Kind: apRelative, Value: feltUint64(0)},
					{Name: "b.high", Kind: apRelative, Value: felt127},
					{Name: "carry_low", Kind: uninitialized},
					{Name: "carry_high", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint256AddHint(ctx.operanders["a.low"], ctx.operanders["b.low"], ctx.operanders["carry_low"], ctx.operanders["carry_high"], false)
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
					{Name: "a", Kind: fpRelative, Value: felt127},
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
					{Name: "n", Kind: fpRelative, Value: felt127},
					{Name: "root", Kind: apRelative, Value: feltUint64(0)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint256SqrtHint(ctx.operanders["n"], ctx.operanders["root"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"root": feltUint64(1),
				}),
			},
		},
	})
}
