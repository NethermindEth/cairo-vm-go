package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func TestZeroHintDictionaries(t *testing.T) {
	runHinterTests(t, map[string][]hintTestCase{
		"SquashDictInnerAssertLenKeys": {
			{
				operanders: []*hintOperander{},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("keys", []fp.Element{})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSquashDictInnerAssertLenKeysHint()
				},
				check: func(t *testing.T, ctx *hintTestContext) {},
			},
			{
				operanders: []*hintOperander{},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("keys", []fp.Element{*feltUint64(1), *feltUint64(2)})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSquashDictInnerAssertLenKeysHint()
				},
				errCheck: errorTextContains("assertion `len(keys) == 0` failed"),
			},
		},
		"SquashDictInnerContinueLoop": {
			{
				operanders: []*hintOperander{
					{Name: "loop_temps", Kind: fpRelative, Value: addr(7)},
					{Name: "loop_temps.index_delta_minus1", Kind: apRelative, Value: feltInt64(0)},
					{Name: "loop_temps.index_delta", Kind: apRelative, Value: feltInt64(0)},
					{Name: "loop_temps.ptr_delta", Kind: apRelative, Value: feltInt64(0)},
					{Name: "loop_temps.should_continue", Kind: uninitialized},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("current_access_indices", []fp.Element{*feltUint64(1), *feltUint64(2), *feltUint64(3)})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSquashDictInnerContinueLoopHint(ctx.operanders["loop_temps"])
				},
				check: varValueEquals("loop_temps.should_continue", feltInt64(1)),
			},
			{
				operanders: []*hintOperander{
					{Name: "loop_temps", Kind: fpRelative, Value: addr(7)},
					{Name: "loop_temps.index_delta_minus1", Kind: apRelative, Value: feltInt64(0)},
					{Name: "loop_temps.index_delta", Kind: apRelative, Value: feltInt64(0)},
					{Name: "loop_temps.ptr_delta", Kind: apRelative, Value: feltInt64(0)},
					{Name: "loop_temps.should_continue", Kind: uninitialized},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("current_access_indices", []fp.Element{})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSquashDictInnerContinueLoopHint(ctx.operanders["loop_temps"])
				},
				check: varValueEquals("loop_temps.should_continue", feltInt64(0)),
			},
		},
		"SquashDictInnerLenAssert": {
			{
				operanders: []*hintOperander{},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("current_access_indices", []fp.Element{})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSquashDictInnerLenAssertHint()
				},
				check: func(t *testing.T, ctx *hintTestContext) {},
			},
			{
				operanders: []*hintOperander{},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("current_access_indices", []fp.Element{*feltUint64(1), *feltUint64(2)})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSquashDictInnerLenAssertHint()
				},
				errCheck: errorTextContains("assertion `len(current_access_indices) == 0` failed"),
			},
		},
		"SquashDictInnerNextKey": {
			{
				operanders: []*hintOperander{
					{Name: "next_key", Kind: uninitialized},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("keys", []fp.Element{})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSquashDictInnerNextKeyHint(ctx.operanders["next_key"])
				},
				errCheck: errorTextContains("no keys left but remaining_accesses > 0"),
			},
			{
				operanders: []*hintOperander{
					{Name: "next_key", Kind: uninitialized},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("keys", []fp.Element{*feltUint64(3), *feltUint64(2), *feltUint64(1)})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSquashDictInnerNextKeyHint(ctx.operanders["next_key"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					allVarValueInScopeEquals(map[string]any{"keys": []fp.Element{*feltUint64(3), *feltUint64(2)}, "key": *feltUint64((1))})(t, ctx)
					varValueEquals("next_key", feltUint64(1))(t, ctx)
				},
			},
			{
				operanders: []*hintOperander{
					{Name: "next_key", Kind: uninitialized},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("keys", []fp.Element{*feltUint64(15), *feltUint64(12), *feltUint64(9), *feltUint64(7), *feltUint64(6), *feltUint64(4)})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSquashDictInnerNextKeyHint(ctx.operanders["next_key"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					allVarValueInScopeEquals(map[string]any{"keys": []fp.Element{*feltUint64(15), *feltUint64(12), *feltUint64(9), *feltUint64(7), *feltUint64(6)}, "key": *feltUint64((4))})(t, ctx)
					varValueEquals("next_key", feltUint64(4))(t, ctx)
				},
			},
		},
	})
}
