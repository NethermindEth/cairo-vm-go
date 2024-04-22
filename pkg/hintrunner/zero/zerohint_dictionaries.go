package zero

import (
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
)

func newDefaultDictNewHint(defaultValue hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "DefaultDictNew",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> if '__dict_manager' not in globals():
			//> 	from starkware.cairo.common.dict import DictManager
			//> 	__dict_manager = DictManager()
			//>
			//> memory[ap] = __dict_manager.new_default_dict(segments, ids.default_value)

			//> if '__dict_manager' not in globals():
			//> 	from starkware.cairo.common.dict import DictManager
			//> 	__dict_manager = DictManager()
			dictionaryManager, ok := ctx.ScopeManager.GetDictionaryManager()
			if !ok {
				hinter.InitializeDictionaryManager(ctx)
				dictionaryManager = ctx.DictionaryManager
				err := ctx.ScopeManager.AssignVariable("__dict_manager", dictionaryManager)
				if err != nil {
					return err
				}
			}

			//> memory[ap] = __dict_manager.new_default_dict(segments, ids.default_value)
			defaultValue, err := hinter.ResolveAsFelt(vm, defaultValue)
			if err != nil {
				return err
			}
			defaultValueMv := mem.MemoryValueFromFieldElement(defaultValue)
			newDefaultDictionaryAddr := dictionaryManager.NewDefaultDictionary(vm, &defaultValueMv)
			newDefaultDictionaryAddrMv := mem.MemoryValueFromMemoryAddress(&newDefaultDictionaryAddr)
			apAddr := vm.Context.AddressAp()
			return vm.Memory.WriteToAddress(&apAddr, &newDefaultDictionaryAddrMv)
		},
	}
}

func createDefaultDictNewHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	defaultValue, err := resolver.GetResOperander("default_value")
	if err != nil {
		return nil, err
	}
	return newDefaultDictNewHint(defaultValue), nil
}
