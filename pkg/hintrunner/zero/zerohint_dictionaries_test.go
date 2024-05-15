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
		"SquashDict": {
			{
				operanders: []*hintOperander{
					{Name: "dict_accesses.1.key", Kind: apRelative, Value: feltUint64(10)},
					{Name: "dict_accesses.1.prev_value", Kind: apRelative, Value: feltUint64(10)},
					{Name: "dict_accesses.1.new_value", Kind: apRelative, Value: feltUint64(10)},
					{Name: "dict_accesses.2.key", Kind: apRelative, Value: feltUint64(10)},
					{Name: "dict_accesses.2.prev_value", Kind: apRelative, Value: feltUint64(10)},
					{Name: "dict_accesses.2.new_value", Kind: apRelative, Value: feltUint64(10)},
					{Name: "ptr_diff", Kind: apRelative, Value: feltUint64(7)},
					{Name: "n_accesses", Kind: apRelative, Value: feltUint64(4)},
					{Name: "big_keys", Kind: uninitialized},
					{Name: "first_key", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSquashDictHint(
						ctx.operanders["dict_accesses.1.key"],
						ctx.operanders["ptr_diff"],
						ctx.operanders["n_accesses"],
						ctx.operanders["big_keys"],
						ctx.operanders["first_key"],
					)
				},
				errCheck: errorTextContains("Accesses array size must be divisible by DictAccess.SIZE"),
			},
			{
				operanders: []*hintOperander{
					{Name: "dict_accesses.1.key", Kind: apRelative, Value: feltUint64(10)},
					{Name: "dict_accesses.1.prev_value", Kind: apRelative, Value: feltUint64(10)},
					{Name: "dict_accesses.1.new_value", Kind: apRelative, Value: feltUint64(10)},
					{Name: "dict_accesses.2.key", Kind: apRelative, Value: feltUint64(10)},
					{Name: "dict_accesses.2.prev_value", Kind: apRelative, Value: feltUint64(10)},
					{Name: "dict_accesses.2.new_value", Kind: apRelative, Value: feltUint64(10)},
					{Name: "ptr_diff", Kind: apRelative, Value: feltUint64(6)},
					{Name: "n_accesses", Kind: apRelative, Value: feltUint64(1048577)},
					{Name: "big_keys", Kind: uninitialized},
					{Name: "first_key", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSquashDictHint(
						ctx.operanders["dict_accesses.1.key"],
						ctx.operanders["ptr_diff"],
						ctx.operanders["n_accesses"],
						ctx.operanders["big_keys"],
						ctx.operanders["first_key"],
					)
				},
				errCheck: errorTextContains("squash_dict() can only be used with n_accesses<={1048576}. Got: n_accesses={1048577}."),
			},
			{
				operanders: []*hintOperander{
					{Name: "dict_accesses.1.key", Kind: apRelative, Value: feltUint64(8)},
					{Name: "dict_accesses.1.prev_value", Kind: apRelative, Value: feltUint64(0)},
					{Name: "dict_accesses.1.new_value", Kind: apRelative, Value: feltUint64(0)},
					{Name: "dict_accesses.2.key", Kind: apRelative, Value: feltUint64(1)},
					{Name: "dict_accesses.2.prev_value", Kind: apRelative, Value: feltUint64(0)},
					{Name: "dict_accesses.2.new_value", Kind: apRelative, Value: feltUint64(0)},
					{Name: "dict_accesses.3.key", Kind: apRelative, Value: feltUint64(21)},
					{Name: "dict_accesses.3.prev_value", Kind: apRelative, Value: feltUint64(0)},
					{Name: "dict_accesses.3.new_value", Kind: apRelative, Value: feltUint64(0)},
					{Name: "dict_accesses.4.key", Kind: apRelative, Value: feltUint64(22)},
					{Name: "dict_accesses.4.prev_value", Kind: apRelative, Value: feltUint64(0)},
					{Name: "dict_accesses.4.new_value", Kind: apRelative, Value: feltUint64(0)},
					{Name: "dict_accesses.5.key", Kind: apRelative, Value: feltUint64(6)},
					{Name: "dict_accesses.5.prev_value", Kind: apRelative, Value: feltUint64(0)},
					{Name: "dict_accesses.5.new_value", Kind: apRelative, Value: feltUint64(0)},
					{Name: "ptr_diff", Kind: apRelative, Value: feltUint64(15)},
					{Name: "n_accesses", Kind: apRelative, Value: feltUint64(5)},
					{Name: "big_keys", Kind: uninitialized},
					{Name: "first_key", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSquashDictHint(
						ctx.operanders["dict_accesses.1.key"],
						ctx.operanders["ptr_diff"],
						ctx.operanders["n_accesses"],
						ctx.operanders["big_keys"],
						ctx.operanders["first_key"],
					)
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					allVarValueEquals(map[string]*fp.Element{
						"big_keys":  feltInt64(0),
						"first_key": feltInt64(1),
					})(t, ctx)
					allVarValueInScopeEquals(map[string]any{
						"access_indices": map[fp.Element][]uint64{
							*feltUint64(8):  {0},
							*feltUint64(1):  {1},
							*feltUint64(21): {2},
							*feltUint64(22): {3},
							*feltUint64(6):  {4},
						},
						"keys": []fp.Element{*feltUint64(22), *feltUint64(21), *feltUint64(8), *feltUint64(6)},
						"key":  *feltUint64(1),
					})(t, ctx)
				},
			},
		},
	})
}
