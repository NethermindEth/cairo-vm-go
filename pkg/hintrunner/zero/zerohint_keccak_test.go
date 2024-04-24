package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
)

func TestZeroHintKeccak(t *testing.T) {

	runHinterTests(t, map[string][]hintTestCase{
		"newKeccakWriteArgs": {
			{

				operanders: []*hintOperander{
					{Name: "inputs", Kind: apRelative, Value: addr(0)},
					{Name: "low", Kind: apRelative, Value: feltString("0")},
					{Name: "high", Kind: apRelative, Value: feltString("0")},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newKeccakWriteArgsHint(ctx.operanders["inputs"], ctx.operanders["low"], ctx.operanders["high"])
				},
				check: valueAtAddressEquals("inputs", feltString("0"))},
		},
	})
}
