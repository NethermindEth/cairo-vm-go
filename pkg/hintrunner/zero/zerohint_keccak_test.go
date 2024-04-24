package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func TestZeroHintKeccak(t *testing.T) {
	runHinterTests(t, map[string][]hintTestCase{
		"newKeccakWriteArgs": {
			{
				operanders: []*hintOperander{
					{Name: "inputs", Kind: apRelative, Value: addr(7)},
					{Name: "low", Kind: fpRelative, Value: feltUint64(1)},
					{Name: "high", Kind: fpRelative, Value: feltUint64(1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newKeccakWriteArgsHint(ctx.operanders["inputs"], ctx.operanders["low"], ctx.operanders["high"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"inputs",
					[]*fp.Element{
						feltString("1"),
						feltString("0"),
						feltString("1"),
						feltString("0"),
					}),
			},
		},
	})
}
