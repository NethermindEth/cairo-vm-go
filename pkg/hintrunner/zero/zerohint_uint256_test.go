package zero

import (
	"math/big"
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	// "github.com/NethermindEth/cairo-vm-go/pkg/utils"
)

func TestZeroHintUint256(t *testing.T) {
	// 1 << 127
	felt127 := new(fp.Element).SetBigInt(new(big.Int).Lsh(big.NewInt(1), 127))
	runHinterTests(t, map[string][]hintTestCase{
		"Uint256Add": {
			{
				operanders: []*hintOperander{
					{Name: "a", Kind: fpRelative, Value: []*fp.Element{felt127, feltUint64(0)}},
					{Name: "b", Kind: apRelative, Value: []*fp.Element{felt127, feltUint64(0)}},
					{Name: "carry_low", Kind: uninitialized},
					{Name: "carry_high", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint256AddHint(ctx.operanders["a"], ctx.operanders["b"], ctx.operanders["carry_low"], ctx.operanders["carry_high"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"carry_low":  feltUint64(1),
					"carry_high": feltUint64(0),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "a", Kind: fpRelative, Value: []*fp.Element{felt127, felt127}},
					{Name: "b", Kind: apRelative, Value: []*fp.Element{felt127, felt127}},
					{Name: "carry_low", Kind: uninitialized},
					{Name: "carry_high", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint256AddHint(ctx.operanders["a"], ctx.operanders["b"], ctx.operanders["carry_low"], ctx.operanders["carry_high"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"carry_low":  feltUint64(1),
					"carry_high": feltUint64(1),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "a", Kind: fpRelative, Value: []*fp.Element{feltUint64(0), felt127}},
					{Name: "b", Kind: apRelative, Value: []*fp.Element{feltUint64(0), felt127}},
					{Name: "carry_low", Kind: uninitialized},
					{Name: "carry_high", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint256AddHint(ctx.operanders["a"], ctx.operanders["b"], ctx.operanders["carry_low"], ctx.operanders["carry_high"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"carry_low":  feltUint64(0),
					"carry_high": feltUint64(1),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "a", Kind: fpRelative, Value: []*fp.Element{felt127, feltUint64(0)}},
					{Name: "b", Kind: apRelative, Value: []*fp.Element{feltUint64(0), felt127}},
					{Name: "carry_low", Kind: uninitialized},
					{Name: "carry_high", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUint256AddHint(ctx.operanders["a"], ctx.operanders["b"], ctx.operanders["carry_low"], ctx.operanders["carry_high"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"carry_low":  feltUint64(0),
					"carry_high": feltUint64(0),
				}),
			},
		},
	})
}
