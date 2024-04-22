package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
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
	})
}
