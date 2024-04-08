package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
)

func TestZeroHintKeccak(t *testing.T) {

	runHinterTests(t, map[string][]hintTestCase{
		"CairoKeccakFinalize": {
			{
				operanders: []*hintOperander{
					{Name: "a", Kind: apRelative, Value: feltUint64(0)},
					{Name: "b", Kind: apRelative, Value: feltUint64(0)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newCairoKeccakFinalizeHint(ctx.operanders["a"], ctx.operanders["b"])
				},
				check: apValueEquals(feltUint64(0)),
			},
		},
	})
}
