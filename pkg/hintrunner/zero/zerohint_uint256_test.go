package zero

import (
	"math/big"
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func TestZeroHintUint256(t *testing.T) {
	// 1 << 127
	felt127 := new(fp.Element).SetBigInt(new(big.Int).Lsh(big.NewInt(1), 127))
	runHinterTests(t, map[string][]hintTestCase{
		"Uint256Add": {
			{
				operanders: []*hintOperander{
					{Name: "a", Kind: fpRelative, Value: felt127},
					{Name: "a", Kind: fpRelative, Value: feltUint64(0)},
					{Name: "b", Kind: apRelative, Value: felt127},
					{Name: "b", Kind: apRelative, Value: feltUint64(0)},
					{Name: "carry_low", Kind: uninitialized},
					{Name: "carry_high", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint256AddHint(ctx.operanders["a"], ctx.operanders["b"], ctx.operanders["carry_low"], ctx.operanders["carry_high"], false)
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
					{Name: "b.low", Kind: apRelative, Value: felt127},
					{Name: "b.high", Kind: apRelative, Value: feltUint64(0)},
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
	})
}
