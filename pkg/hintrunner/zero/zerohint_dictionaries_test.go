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
				check: varValueEquals("value", feltUint64(12345)),
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
				check: consecutiveVarAddrResolvedValueEquals(
					"dict_ptr.prev_value",
					[]*fp.Element{
						feltString("12345"),
					}),
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
					{Name: "dict_ptr1", Kind: apRelative, Value: addrWithSegment(2, 1)},
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
				check: func(t *testing.T, ctx *hintTestContext) {},
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
