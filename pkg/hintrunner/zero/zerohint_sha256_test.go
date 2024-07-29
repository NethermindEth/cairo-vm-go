package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func TestZeroHintSha256(t *testing.T) {
	runHinterTests(t, map[string][]hintTestCase{
		"PackedSha256": {
			{
				operanders: []*hintOperander{
					{Name: "sha256_start", Kind: apRelative, Value: addr(6)},
					{Name: "output", Kind: apRelative, Value: addr(22)},
					{Name: "buffer", Kind: apRelative, Value: feltUint64(0)},
					{Name: "buffer", Kind: apRelative, Value: feltUint64(0)},
					{Name: "buffer", Kind: apRelative, Value: feltUint64(0)},
					{Name: "buffer", Kind: apRelative, Value: feltUint64(0)},
					{Name: "buffer", Kind: apRelative, Value: feltUint64(0)},
					{Name: "buffer", Kind: apRelative, Value: feltUint64(0)},
					{Name: "buffer", Kind: apRelative, Value: feltUint64(0)},
					{Name: "buffer", Kind: apRelative, Value: feltUint64(0)},
					{Name: "buffer", Kind: apRelative, Value: feltUint64(0)},
					{Name: "buffer", Kind: apRelative, Value: feltUint64(0)},
					{Name: "buffer", Kind: apRelative, Value: feltUint64(0)},
					{Name: "buffer", Kind: apRelative, Value: feltUint64(0)},
					{Name: "buffer", Kind: apRelative, Value: feltUint64(0)},
					{Name: "buffer", Kind: apRelative, Value: feltUint64(0)},
					{Name: "buffer", Kind: apRelative, Value: feltUint64(0)},
					{Name: "buffer", Kind: apRelative, Value: feltUint64(0)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newPackedSha256Hint(ctx.operanders["sha256_start"], ctx.operanders["output"])
				},
				check: consecutiveVarAddrResolvedValueEquals(
					"output",
					[]*fp.Element{
						feltString("3663108286"),
						feltString("398046313"),
						feltString("1647531929"),
						feltString("2006957770"),
						feltString("2363872401"),
						feltString("3235013187"),
						feltString("3137272298"),
						feltString("406301144"),
					}),
			},
		},
	})
}
