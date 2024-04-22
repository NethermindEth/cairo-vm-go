package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
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
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					hinter.InitializeScopeManager(ctx, map[string]any{})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newDefaultDictNewHint(ctx.operanders["default_value"])
				},
				check: func(t *testing.T, ctx *hintTestContext) {
					dictionaryManagerValue, err := ctx.runnerContext.ScopeManager.GetVariableValue("__dict_manager")
					if err != nil {
						t.Fatalf("__dict_manager missing")
					}

					dictionaryManager := dictionaryManagerValue.(hinter.DictionaryManager)
					apAddr := ctx.vm.Context.AddressAp()
					dictAddr, err := ctx.vm.Memory.ReadFromAddressAsAddress(&apAddr)
					if err != nil {
						t.Fatalf("error reading dictionary address from ap")
					}

					key := new(fp.Element).SetUint64(100)
					value, err := dictionaryManager.At(&dictAddr, key)
					if err != nil {
						t.Fatalf("error fetching value from dictionary")
					}
					valueFelt, err := value.FieldElement()
					if err != nil {
						t.Fatalf("mv: %s cannot be converted to felt", value)
					}
					expectedValueFelt := new(fp.Element).SetUint64(12345)

					if !valueFelt.Equal(expectedValueFelt) {
						t.Fatalf("at key: %v expected: %s actual: %s", key, expectedValueFelt, valueFelt)
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
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					hinter.InitializeDictionaryManager(ctx)
					hinter.InitializeScopeManager(ctx, map[string]any{
						"__dict_manager": ctx.DictionaryManager,
					})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					defaultValueMv := memory.MemoryValueFromInt(12345)
					ctx.runnerContext.DictionaryManager.NewDefaultDictionary(ctx.vm, &defaultValueMv)
					return newDictReadHint(ctx.operanders["dict_ptr"], ctx.operanders["key"], ctx.operanders["value"])
				},
				check: varValueEquals("value", feltUint64(12345)),
			},
		},
	})
}
