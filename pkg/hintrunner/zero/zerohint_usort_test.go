package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func TestZeroHintUsort(t *testing.T) {
	runHinterTests(t, map[string][]hintTestCase{
		"UsortEnterScope": {
			{
				operanders: []*hintOperander{
					{Name: "__usort_max_size", Kind: apRelative, Value: feltUint64(1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newUsortEnterScopeHinter()
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"__usort_max_size": feltUint64(1),
				}),
			},
		},
	})
}
