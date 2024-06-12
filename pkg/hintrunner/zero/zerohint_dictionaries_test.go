package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func TestZeroHintDictionaries(t *testing.T) {
	runHinterTests(t, map[string][]hintTestCase{
		"DictNew": {
			{
				operanders: []*hintOperander{},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					value1 := memory.MemoryValueFromUint(uint(1000))
					value2 := memory.MemoryValueFromUint(uint(2000))
					value3 := memory.MemoryValueFromUint(uint(3000))
					err := ctx.ScopeManager.AssignVariable("initial_dict", map[fp.Element]memory.MemoryValue{
						*feltUint64(10): value1,
						*feltUint64(20): value2,
						*feltUint64(30): value3,
					})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newDictNewHint()
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					_, err := ctx.runnerContext.ScopeManager.GetVariableValue("initial_dict")
					if err.Error() != "variable initial_dict not found in current scope" {
						t.Fatalf("initial_dict not deleted")
					}

					apAddr := ctx.vm.Context.AddressAp()
					dictAddr, err := ctx.vm.Memory.ReadFromAddressAsAddress(&apAddr)
					if err != nil {
						t.Fatalf("error reading address from ap")
					}
					if dictAddr.String() != "2:0" {
						t.Fatalf("incorrect apValue: %s expected %s", dictAddr.String(), "2:0")
					}

					dictionaryManagerValue, err := ctx.runnerContext.ScopeManager.GetVariableValue("__dict_manager")
					if err != nil {
						t.Fatalf("__dict_manager missing")
					}
					dictionaryManager := dictionaryManagerValue.(hinter.ZeroDictionaryManager)
					dictionary, err := dictionaryManager.GetDictionary(dictAddr)
					if err != nil {
						t.Fatalf("error fetching dictionary from address at ap")
					}

					for _, key := range []fp.Element{*feltUint64(10), *feltUint64(20), *feltUint64(30)} {
						value, err := dictionary.At(key)
						if err != nil {
							t.Fatalf("error fetching value for key: %v", key)
						}
						valueFelt, err := value.FieldElement()
						if err != nil {
							t.Fatalf("mv: %s cannot be converted to felt", value)
						}
						expectedValueFelt := new(fp.Element).Mul(&key, feltUint64(100))

						if !valueFelt.Equal(expectedValueFelt) {
							t.Fatalf("at key: %v expected: %s actual: %s", key, expectedValueFelt, valueFelt)
						}
					}
				},
			},
		},
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
				check: func(t *testing.T, ctx *hintTestContext) {
					varValueEquals("value", feltUint64(12345))(t, ctx)

					dictPtr := addrWithSegment(2, 0)
					expectedData := map[fp.Element]memory.MemoryValue{}
					expectedDefaultValue := memory.MemoryValueFromInt(12345)
					expectedFreeOffset := uint64(3)
					zeroDictInScopeEquals(*dictPtr, expectedData, expectedDefaultValue, expectedFreeOffset)(t, ctx)
				},
			},
		},
		"DictWrite": {
			{
				operanders: []*hintOperander{
					{Name: "key", Kind: apRelative, Value: feltUint64(100)},
					{Name: "new_value", Kind: apRelative, Value: feltUint64(9999)},
					{Name: "dict_ptr", Kind: apRelative, Value: addrWithSegment(2, 0)},
					{Name: "dict_ptr.prev_value", Kind: apRelative, Value: addrWithSegment(2, 1)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					dictionaryManager := hinter.NewZeroDictionaryManager()
					err := ctx.runnerContext.ScopeManager.AssignVariable("__dict_manager", dictionaryManager)
					if err != nil {
						t.Fatal(err)
					}
					defaultValueMv := memory.MemoryValueFromInt(12345)
					dictionaryManager.NewDefaultDictionary(ctx.vm, defaultValueMv)
					return newDictWriteHint(ctx.operanders["dict_ptr"], ctx.operanders["key"], ctx.operanders["new_value"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					consecutiveVarAddrResolvedValueEquals(
						"dict_ptr.prev_value",
						[]*fp.Element{
							feltString("12345"),
						})(t, ctx)

					dictPtr := addrWithSegment(2, 0)
					expectedData := map[fp.Element]memory.MemoryValue{*feltUint64(100): memory.MemoryValueFromInt(9999)}
					expectedDefaultValue := memory.MemoryValueFromInt(12345)
					expectedFreeOffset := uint64(3)
					zeroDictInScopeEquals(*dictPtr, expectedData, expectedDefaultValue, expectedFreeOffset)(t, ctx)
				},
			},
		},
		"DictUpdate": {
			{
				operanders: []*hintOperander{
					{Name: "key", Kind: apRelative, Value: feltUint64(100)},
					{Name: "new_value", Kind: apRelative, Value: feltUint64(4)},
					{Name: "prev_value", Kind: apRelative, Value: feltUint64(2)},
					{Name: "dict_ptr", Kind: apRelative, Value: addrWithSegment(2, 0)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					dictionaryManager := hinter.NewZeroDictionaryManager()
					err := ctx.runnerContext.ScopeManager.AssignVariable("__dict_manager", dictionaryManager)
					if err != nil {
						t.Fatal(err)
					}
					defaultValueMv := memory.MemoryValueFromInt(1)
					dictionaryManager.NewDefaultDictionary(ctx.vm, defaultValueMv)
					return newDictUpdateHint(ctx.operanders["dict_ptr"], ctx.operanders["key"], ctx.operanders["new_value"], ctx.operanders["prev_value"])
				},
				errCheck: errorTextContains("Wrong previous value in dict. Got 2, expected 1."),
			},
			{
				operanders: []*hintOperander{
					{Name: "key", Kind: apRelative, Value: feltUint64(100)},
					{Name: "new_value", Kind: apRelative, Value: feltUint64(4)},
					{Name: "prev_value", Kind: apRelative, Value: feltUint64(1)},
					{Name: "dict_ptr", Kind: apRelative, Value: addrWithSegment(2, 0)},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					dictionaryManager := hinter.NewZeroDictionaryManager()
					err := ctx.runnerContext.ScopeManager.AssignVariable("__dict_manager", dictionaryManager)
					if err != nil {
						t.Fatal(err)
					}
					defaultValueMv := memory.MemoryValueFromInt(1)
					dictionaryManager.NewDefaultDictionary(ctx.vm, defaultValueMv)
					return newDictUpdateHint(ctx.operanders["dict_ptr"], ctx.operanders["key"], ctx.operanders["new_value"], ctx.operanders["prev_value"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					dictPtr := addrWithSegment(2, 0)
					expectedData := map[fp.Element]memory.MemoryValue{*feltUint64(100): memory.MemoryValueFromInt(4)}
					expectedDefaultValue := memory.MemoryValueFromInt(1)
					expectedFreeOffset := uint64(3)
					zeroDictInScopeEquals(*dictPtr, expectedData, expectedDefaultValue, expectedFreeOffset)(t, ctx)
				},
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
		"SquashDictInnerFirstIteration": {
			{
				operanders: []*hintOperander{
					{Name: "range_check_ptr", Kind: fpRelative, Value: addr(6)},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariables(map[string]any{"access_indices": map[fp.Element][]fp.Element{*feltUint64(0): {*feltUint64(2), *feltUint64(1), *feltUint64(3)}}, "key": *feltUint64(0)})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSquashDictInnerFirstIterationHint(ctx.operanders["range_check_ptr"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					valueAtAddressEquals(*addr(6), feltUint64(1))(t, ctx)
					allVarValueInScopeEquals(map[string]any{"current_access_indices": []fp.Element{*feltUint64(3), *feltUint64(2)}, "current_access_index": *feltUint64(1), "access_indices": map[fp.Element][]fp.Element{*feltUint64(0): {*feltUint64(2), *feltUint64(1), *feltUint64(3)}}})(t, ctx)
				},
			},
			{
				operanders: []*hintOperander{
					{Name: "range_check_ptr", Kind: fpRelative, Value: addr(6)},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariables(map[string]any{"access_indices": map[fp.Element][]fp.Element{*feltUint64(0): {}, *feltUint64(1): {*feltUint64(22), *feltUint64(76), *feltUint64(94), *feltUint64(55), *feltUint64(18), *feltUint64(92)}}, "key": *feltUint64(1)})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSquashDictInnerFirstIterationHint(ctx.operanders["range_check_ptr"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					valueAtAddressEquals(*addr(6), feltUint64(18))(t, ctx)
					allVarValueInScopeEquals(map[string]any{"current_access_indices": []fp.Element{*feltUint64(94), *feltUint64(92), *feltUint64(76), *feltUint64(55), *feltUint64(22)}, "current_access_index": *feltUint64(18), "access_indices": map[fp.Element][]fp.Element{*feltUint64(0): {}, *feltUint64(1): {*feltUint64(22), *feltUint64(76), *feltUint64(94), *feltUint64(55), *feltUint64(18), *feltUint64(92)}}, "key": *feltUint64(1)})(t, ctx)
				},
			},
			{
				operanders: []*hintOperander{
					{Name: "range_check_ptr", Kind: fpRelative, Value: addr(6)},
				},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					err := ctx.ScopeManager.AssignVariables(map[string]any{"access_indices": map[fp.Element][]fp.Element{*feltUint64(0): {*feltUint64(22)}, *feltUint64(1): {*feltUint64(5), *feltUint64(28)}, *feltUint64(2): {*feltUint64(543), *feltUint64(323), *feltUint64(324), *feltUint64(999), *feltUint64(888), *feltUint64(777)}}, "key": *feltUint64(2)})
					if err != nil {
						t.Fatal(err)
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSquashDictInnerFirstIterationHint(ctx.operanders["range_check_ptr"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					valueAtAddressEquals(*addr(6), feltUint64(323))(t, ctx)
					varValueInScopeEquals("current_access_indices", []fp.Element{*feltUint64(999), *feltUint64(888), *feltUint64(777), *feltUint64(543), *feltUint64(324)})(t, ctx)
					varValueInScopeEquals("current_access_index", *feltUint64(323))(t, ctx)
				},
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
	})
}
