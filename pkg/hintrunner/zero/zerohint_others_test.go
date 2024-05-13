package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func TestZeroHintOthers(t *testing.T) {
	runHinterTests(t, map[string][]hintTestCase{
		"MemcpyEnterScope": {
			{
				operanders: []*hintOperander{
					{Name: "len", Kind: apRelative, Value: feltUint64(1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newMemcpyEnterScopeHint(ctx.operanders["len"])
				},
				check: varValueInScopeEquals("n", feltUint64(1)),
			},
		},
		"SetAdd": {
			{
				operanders: []*hintOperander{
					{Name: "elm.1", Kind: apRelative, Value: feltUint64(1)},
					{Name: "elm.2", Kind: apRelative, Value: feltUint64(2)},
					{Name: "elm.3", Kind: apRelative, Value: feltUint64(3)},
					{Name: "elm.4", Kind: apRelative, Value: feltUint64(4)},
					{Name: "set.1", Kind: apRelative, Value: feltUint64(5)},
					{Name: "set.2", Kind: apRelative, Value: feltUint64(6)},
					{Name: "set.3", Kind: apRelative, Value: feltUint64(7)},
					{Name: "set.4", Kind: apRelative, Value: feltUint64(8)},
					{Name: "set.5", Kind: apRelative, Value: feltUint64(9)},
					{Name: "set.6", Kind: apRelative, Value: feltUint64(10)},
					{Name: "set.7", Kind: apRelative, Value: feltUint64(11)},
					{Name: "set.8", Kind: apRelative, Value: feltUint64(12)},
					{Name: "set.9", Kind: apRelative, Value: feltUint64(1)},
					{Name: "set.10", Kind: apRelative, Value: feltUint64(2)},
					{Name: "set.11", Kind: apRelative, Value: feltUint64(3)},
					{Name: "set.12", Kind: apRelative, Value: feltUint64(4)},
					{Name: "set.9", Kind: apRelative, Value: feltUint64(13)},
					{Name: "set.10", Kind: apRelative, Value: feltUint64(14)},
					{Name: "set.11", Kind: apRelative, Value: feltUint64(15)},
					{Name: "set.12", Kind: apRelative, Value: feltUint64(16)},
					{Name: "elm_size", Kind: apRelative, Value: feltUint64(4)},
					{Name: "elm_ptr", Kind: apRelative, Value: addrWithSegment(1, 4)},
					{Name: "set_ptr", Kind: apRelative, Value: addrWithSegment(1, 8)},
					{Name: "set_end_ptr", Kind: apRelative, Value: addrWithSegment(1, 24)},
					{Name: "index", Kind: uninitialized},
					{Name: "is_elm_in_set", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSetAddHint(
						ctx.operanders["elm_size"],
						ctx.operanders["elm_ptr"],
						ctx.operanders["set_ptr"],
						ctx.operanders["set_end_ptr"],
						ctx.operanders["index"],
						ctx.operanders["is_elm_in_set"],
					)
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"index":         feltUint64(2),
					"is_elm_in_set": feltUint64(1),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "elm.1", Kind: apRelative, Value: feltUint64(1)},
					{Name: "elm.2", Kind: apRelative, Value: feltUint64(2)},
					{Name: "elm.3", Kind: apRelative, Value: feltUint64(3)},
					{Name: "elm.4", Kind: apRelative, Value: feltUint64(4)},
					{Name: "set.1", Kind: apRelative, Value: feltUint64(5)},
					{Name: "set.2", Kind: apRelative, Value: feltUint64(6)},
					{Name: "set.3", Kind: apRelative, Value: feltUint64(7)},
					{Name: "set.4", Kind: apRelative, Value: feltUint64(8)},
					{Name: "set.5", Kind: apRelative, Value: feltUint64(9)},
					{Name: "set.6", Kind: apRelative, Value: feltUint64(10)},
					{Name: "set.7", Kind: apRelative, Value: feltUint64(11)},
					{Name: "set.8", Kind: apRelative, Value: feltUint64(12)},
					{Name: "set.9", Kind: apRelative, Value: feltUint64(13)},
					{Name: "set.10", Kind: apRelative, Value: feltUint64(14)},
					{Name: "set.11", Kind: apRelative, Value: feltUint64(15)},
					{Name: "set.12", Kind: apRelative, Value: feltUint64(16)},
					{Name: "set.9", Kind: apRelative, Value: feltUint64(17)},
					{Name: "set.10", Kind: apRelative, Value: feltUint64(18)},
					{Name: "set.11", Kind: apRelative, Value: feltUint64(19)},
					{Name: "set.12", Kind: apRelative, Value: feltUint64(20)},
					{Name: "elm_size", Kind: apRelative, Value: feltUint64(4)},
					{Name: "elm_ptr", Kind: apRelative, Value: addrWithSegment(1, 4)},
					{Name: "set_ptr", Kind: apRelative, Value: addrWithSegment(1, 8)},
					{Name: "set_end_ptr", Kind: apRelative, Value: addrWithSegment(1, 24)},
					{Name: "index", Kind: uninitialized},
					{Name: "is_elm_in_set", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSetAddHint(
						ctx.operanders["elm_size"],
						ctx.operanders["elm_ptr"],
						ctx.operanders["set_ptr"],
						ctx.operanders["set_end_ptr"],
						ctx.operanders["index"],
						ctx.operanders["is_elm_in_set"],
					)
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"is_elm_in_set": feltUint64(0),
				}),
			},
			{
				operanders: []*hintOperander{
					{Name: "elm.1", Kind: apRelative, Value: feltUint64(1)},
					{Name: "elm.2", Kind: apRelative, Value: feltUint64(2)},
					{Name: "elm.3", Kind: apRelative, Value: feltUint64(3)},
					{Name: "elm.4", Kind: apRelative, Value: feltUint64(4)},
					{Name: "elm.5", Kind: apRelative, Value: feltUint64(5)},
					{Name: "set.1", Kind: apRelative, Value: feltUint64(1)},
					{Name: "set.2", Kind: apRelative, Value: feltUint64(2)},
					{Name: "set.3", Kind: apRelative, Value: feltUint64(3)},
					{Name: "set.4", Kind: apRelative, Value: feltUint64(4)},
					{Name: "set.5", Kind: apRelative, Value: feltUint64(5)},
					{Name: "elm_size", Kind: apRelative, Value: feltUint64(5)},
					{Name: "elm_ptr", Kind: apRelative, Value: addrWithSegment(1, 4)},
					{Name: "set_ptr", Kind: apRelative, Value: addrWithSegment(1, 9)},
					{Name: "set_end_ptr", Kind: apRelative, Value: addrWithSegment(1, 14)},
					{Name: "index", Kind: uninitialized},
					{Name: "is_elm_in_set", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSetAddHint(
						ctx.operanders["elm_size"],
						ctx.operanders["elm_ptr"],
						ctx.operanders["set_ptr"],
						ctx.operanders["set_end_ptr"],
						ctx.operanders["index"],
						ctx.operanders["is_elm_in_set"],
					)
				},
				check: allVarValueEquals(map[string]*fp.Element{
					"index":         feltUint64(0),
					"is_elm_in_set": feltUint64(1),
				}),
			},
		},
	})
}
