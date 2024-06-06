package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func TestZeroHintDictionaries(t *testing.T) {
	runHinterTests(t, map[string][]hintTestCase{
		"DefaultDictNew": {
			{
				operanders: []*hintOperander{
					{Name: "default_value", Kind: apRelative, Value: feltUint64(12345)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newDefaultDictNewHint(ctx.operanders["default_value"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					dictionaryManagerValue, err := ctx.runnerContext.ScopeManager.GetVariableValue("__dict_manager")
					if err != nil {
						t.Fatalf("__dict_manager missing")
					}

					dictionaryManager := dictionaryManagerValue.(hinter.ZeroDictionaryManager)
					apAddr := ctx.vm.Context.AddressAp()
					dictAddr, err := ctx.vm.Memory.ReadFromAddressAsAddress(&apAddr)
					if err != nil {
						t.Fatalf("error reading dictionary address from ap")
					}

					key := fp.NewElement(100)
					value, err := dictionaryManager.At(dictAddr, key)
					if err != nil {
						t.Fatalf("error fetching value from dictionary")
					}
					valueFelt, err := value.FieldElement()
					if err != nil {
						t.Fatalf("mv: %s cannot be converted to felt", value)
					}
					expectedValueFelt := fp.NewElement(12345)

					if !valueFelt.Equal(&expectedValueFelt) {
						t.Fatalf("at key: %v expected: %s actual: %s", key, &expectedValueFelt, valueFelt)
					}
				},
			},
		},
		"DictRead": {
			{
				operanders: []*hintOperander{
					{Name: "dict_ptr", Kind: apRelative, Value: addrWithSegment(2, 0)},
					{Name: "key", Kind: apRelative, Value: feltUint64(100)},
					{Name: "value", Kind: uninitialized},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					dictionaryManager := hinter.NewZeroDictionaryManager()
					err := ctx.runnerContext.ScopeManager.AssignVariable("__dict_manager", dictionaryManager)
					if err != nil {
						t.Fatal(err)
					}
					defaultValueMv := memory.MemoryValueFromInt(12345)
					dictionaryManager.NewDefaultDictionary(ctx.vm, defaultValueMv)
					return newDictReadHint(ctx.operanders["dict_ptr"], ctx.operanders["key"], ctx.operanders["value"])
				},
				check: varValueEquals("value", feltUint64(12345)),
			},
		},
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
					return newSquashDictInnerContinueLoopHint(ctx.operanders["loop_temps.index_delta_minus1"])
				},
				check: varValueEquals("loop_temps.should_continue", feltInt64(1)),
			},
			{
				operanders: []*hintOperander{
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
					return newSquashDictInnerContinueLoopHint(ctx.operanders["loop_temps.index_delta_minus1"])
				},
				check: varValueEquals("loop_temps.should_continue", feltInt64(0)),
			},
		},
		"SquashDictInnerSkipLoop": {
			{
				operanders: []*hintOperander{
					{Name: "should_skip_loop", Kind: uninitialized},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("current_access_indices", []fp.Element{*feltUint64(1), *feltUint64(2), *feltUint64(3)})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSquashDictInnerSkipLoopHint(ctx.operanders["should_skip_loop"])
				},
				check: varValueEquals("should_skip_loop", feltInt64(0)),
			},
			{
				operanders: []*hintOperander{
					{Name: "should_skip_loop", Kind: uninitialized},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariable("current_access_indices", []fp.Element{})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSquashDictInnerSkipLoopHint(ctx.operanders["should_skip_loop"])
				},
				check: varValueEquals("should_skip_loop", feltInt64(1)),
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
		"SquashDictInnerUsedAccessesAssert": {
			{
				operanders: []*hintOperander{
					{Name: "n_used_accesses", Kind: apRelative, Value: feltInt64(0)},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariables(map[string]any{"access_indices": map[fp.Element][]fp.Element{*feltUint64(0): {}, *feltUint64(1): {*feltUint64(1), *feltUint64(2), *feltUint64(3)}}, "key": *feltUint64(0)})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSquashDictInnerUsedAccessesAssertHint(ctx.operanders["n_used_accesses"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {},
			},
			{
				operanders: []*hintOperander{
					{Name: "n_used_accesses", Kind: apRelative, Value: feltInt64(0)},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariables(map[string]any{"access_indices": map[fp.Element][]fp.Element{*feltUint64(0): {}, *feltUint64(1): {*feltUint64(1), *feltUint64(2), *feltUint64(3)}}, "key": *feltUint64(1)})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSquashDictInnerUsedAccessesAssertHint(ctx.operanders["n_used_accesses"])
				},
				errCheck: errorTextContains("assertion ids.n_used_accesses == len(access_indices[key]) failed"),
			},
			{
				operanders: []*hintOperander{
					{Name: "n_used_accesses", Kind: apRelative, Value: feltInt64(3)},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariables(map[string]any{"access_indices": map[fp.Element][]fp.Element{*feltUint64(0): {}, *feltUint64(1): {*feltUint64(1), *feltUint64(2), *feltUint64(3)}}, "key": *feltUint64(1)})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSquashDictInnerUsedAccessesAssertHint(ctx.operanders["n_used_accesses"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {},
			},
			{
				operanders: []*hintOperander{
					{Name: "n_used_accesses", Kind: apRelative, Value: feltInt64(3)},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariables(map[string]any{"access_indices": map[fp.Element][]fp.Element{*feltUint64(0): {}, *feltUint64(1): {*feltUint64(1), *feltUint64(2), *feltUint64(3)}}, "key": *feltUint64(0)})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSquashDictInnerUsedAccessesAssertHint(ctx.operanders["n_used_accesses"])
				},
				errCheck: errorTextContains("assertion ids.n_used_accesses == len(access_indices[key]) failed"),
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
					// checking if it catches ptr_diff % 3 != 0
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
					// checking if it catches n_accesses > 1048576
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
					// random correct values
					{Name: "dict_accesses.1.key", Kind: apRelative, Value: feltUint64(8)},
					{Name: "dict_accesses.1.prev_value", Kind: apRelative, Value: feltUint64(0)},
					{Name: "dict_accesses.1.new_value", Kind: apRelative, Value: feltUint64(0)},
					{Name: "dict_accesses.2.key", Kind: apRelative, Value: feltUint64(1)},
					{Name: "dict_accesses.2.prev_value", Kind: apRelative, Value: feltUint64(0)},
					{Name: "dict_accesses.2.new_value", Kind: apRelative, Value: feltUint64(0)},
					{Name: "dict_accesses.3.key", Kind: apRelative, Value: feltUint64(21)},
					{Name: "dict_accesses.3.prev_value", Kind: apRelative, Value: feltUint64(0)},
					{Name: "dict_accesses.3.new_value", Kind: apRelative, Value: feltUint64(0)},
					// largest key within range_check_builtin.bound
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
			{
				operanders: []*hintOperander{
					// random correct values
					{Name: "dict_accesses.1.key", Kind: apRelative, Value: feltUint64(8)},
					{Name: "dict_accesses.1.prev_value", Kind: apRelative, Value: feltUint64(0)},
					{Name: "dict_accesses.1.new_value", Kind: apRelative, Value: feltUint64(0)},
					{Name: "dict_accesses.2.key", Kind: apRelative, Value: feltUint64(1)},
					{Name: "dict_accesses.2.prev_value", Kind: apRelative, Value: feltUint64(0)},
					{Name: "dict_accesses.2.new_value", Kind: apRelative, Value: feltUint64(0)},
					{Name: "dict_accesses.3.key", Kind: apRelative, Value: feltUint64(21)},
					{Name: "dict_accesses.3.prev_value", Kind: apRelative, Value: feltUint64(0)},
					{Name: "dict_accesses.3.new_value", Kind: apRelative, Value: feltUint64(0)},
					// largest key bigger than range_check_builtin.bound
					{Name: "dict_accesses.4.key", Kind: apRelative, Value: &utils.FeltMax128},
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
						"big_keys":  feltInt64(1),
						"first_key": feltInt64(1),
					})(t, ctx)
					allVarValueInScopeEquals(map[string]any{
						"access_indices": map[fp.Element][]uint64{
							*feltUint64(8):   {0},
							*feltUint64(1):   {1},
							*feltUint64(21):  {2},
							utils.FeltMax128: {3},
							*feltUint64(6):   {4},
						},
						"keys": []fp.Element{utils.FeltMax128, *feltUint64(21), *feltUint64(8), *feltUint64(6)},
						"key":  *feltUint64(1),
					})(t, ctx)
				},
			},
			{
				operanders: []*hintOperander{
					// random correct values where keys are repeated
					{Name: "dict_accesses.1.key", Kind: apRelative, Value: feltUint64(80)},
					{Name: "dict_accesses.1.prev_value", Kind: apRelative, Value: feltUint64(0)},
					{Name: "dict_accesses.1.new_value", Kind: apRelative, Value: feltUint64(0)},
					{Name: "dict_accesses.2.key", Kind: apRelative, Value: feltUint64(29)},
					{Name: "dict_accesses.2.prev_value", Kind: apRelative, Value: feltUint64(0)},
					{Name: "dict_accesses.2.new_value", Kind: apRelative, Value: feltUint64(0)},
					{Name: "dict_accesses.3.key", Kind: apRelative, Value: feltUint64(210)},
					{Name: "dict_accesses.3.prev_value", Kind: apRelative, Value: feltUint64(0)},
					{Name: "dict_accesses.3.new_value", Kind: apRelative, Value: feltUint64(0)},
					// largest key bigger than range_check_builtin.bound
					{Name: "dict_accesses.4.key", Kind: apRelative, Value: &utils.FeltUpperBound},
					{Name: "dict_accesses.4.prev_value", Kind: apRelative, Value: feltUint64(0)},
					{Name: "dict_accesses.4.new_value", Kind: apRelative, Value: feltUint64(0)},
					{Name: "dict_accesses.5.key", Kind: apRelative, Value: feltUint64(29)},
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
						"big_keys":  feltInt64(1),
						"first_key": feltInt64(29),
					})(t, ctx)
					allVarValueInScopeEquals(map[string]any{
						"access_indices": map[fp.Element][]uint64{
							*feltUint64(80):      {0},
							*feltUint64(29):      {1, 4},
							*feltUint64(210):     {2},
							utils.FeltUpperBound: {3},
						},
						"keys": []fp.Element{utils.FeltUpperBound, *feltUint64(210), *feltUint64(80)},
						"key":  *feltUint64(29),
					})(t, ctx)
				},
			},
		},
	})
}
