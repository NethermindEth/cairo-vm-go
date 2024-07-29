package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func TestInvModPUint512(t *testing.T) {
	runHinterTests(t, map[string][]hintTestCase{
		"InvModPUint512": {
			{
				operanders: []*hintOperander{
					{Name: "x.d0", Kind: apRelative, Value: feltUint64(101)},
					{Name: "x.d1", Kind: apRelative, Value: feltUint64(2)},
					{Name: "x.d2", Kind: apRelative, Value: feltUint64(15)},
					{Name: "x.d3", Kind: apRelative, Value: feltUint64(61)},
					{Name: "x_inverse_mod_p.low", Kind: uninitialized},
					{Name: "x_inverse_mod_p.high", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newInvModPUint512Hint(ctx.operanders["x.d0"], ctx.operanders["x_inverse_mod_p.low"])
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"x_inverse_mod_p.low":  feltString("80275402838848031859800366538378848249"),
					"x_inverse_mod_p.high": feltString("5810892639608724280512701676461676039"),
				}),
			},
		},
	})
}
