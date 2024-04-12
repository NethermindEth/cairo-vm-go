package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func TestZeroHintDictionaries(t *testing.T) {
	runHinterTests(t, map[string][]hintTestCase{
		"DictNew": {
			{
				operanders: []*hintOperander{},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					value1 := mem.MemoryValueFromUint(uint(1000))
					value2 := mem.MemoryValueFromUint(uint(2000))
					value3 := mem.MemoryValueFromUint(uint(3000))
					hinter.InitializeScopeManager(ctx, map[string]any{
						"initial_dict": map[fp.Element]*mem.MemoryValue{
							*new(fp.Element).SetUint64(10): &value1,
							*new(fp.Element).SetUint64(20): &value2,
							*new(fp.Element).SetUint64(30): &value3,
						},
					})
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
					dictionaryManager := dictionaryManagerValue.(*hinter.DictionaryManager)
					dictionary, err := dictionaryManager.GetDictionary(&dictAddr)
					if err != nil {
						t.Fatalf("error fetching dictionary from address at ap")
					}

					for _, key := range []fp.Element{*new(fp.Element).SetUint64(10), *new(fp.Element).SetUint64(20), *new(fp.Element).SetUint64(30)} {
						value, err := dictionary.At(&key)
						if err != nil {
							t.Fatalf("error fetching value for key: %v", key)
						}
						valueFelt, err := value.FieldElement()
						if err != nil {
							t.Fatalf("mv: %s cannot be converted to felt", value)
						}
						expectedValueFelt := new(fp.Element).Mul(&key, new(fp.Element).SetUint64(100))

						if !valueFelt.Equal(expectedValueFelt) {
							t.Fatalf("at key: %v expected: %s actual: %s", key, expectedValueFelt, valueFelt)
						}
					}
				},
			},
		},
	})
}
