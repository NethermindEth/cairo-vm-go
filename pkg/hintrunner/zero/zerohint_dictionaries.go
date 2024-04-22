package zero

import (
	"fmt"

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

func newDictReadHint(dictPtr, key, value hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "DictRead",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> dict_tracker = __dict_manager.get_tracker(ids.dict_ptr)
			//> dict_tracker.current_ptr += ids.DictAccess.SIZE
			//> ids.value = dict_tracker.data[ids.key]

			//> dict_tracker = __dict_manager.get_tracker(ids.dict_ptr)
			dictPtr, err := hinter.ResolveAsAddress(vm, dictPtr)
			if err != nil {
				return err
			}
			dictionaryManager, ok := ctx.ScopeManager.GetDictionaryManager()
			if !ok {
				return fmt.Errorf("__dict_manager not in scope")
			}
			dictionary, err := dictionaryManager.GetDictionary(dictPtr)
			if err != nil {
				return err
			}

			//> dict_tracker.current_ptr += ids.DictAccess.SIZE
			dictionary.IncrementFreeOffset(3)

			//> ids.value = dict_tracker.data[ids.key]
			key, err := hinter.ResolveAsFelt(vm, key)
			if err != nil {
				return err
			}
			keyValue, err := dictionary.At(key)
			if err != nil {
				return err
			}
			valueAddr, err := value.GetAddress(vm)
			if err != nil {
				return err
			}
			return vm.Memory.WriteToAddress(&valueAddr, keyValue)
		},
	}
}

func createDictReadHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	dictPtr, err := resolver.GetResOperander("dict_ptr")
	if err != nil {
		return nil, err
	}
	key, err := resolver.GetResOperander("key")
	if err != nil {
		return nil, err
	}
	value, err := resolver.GetResOperander("value")
	if err != nil {
		return nil, err
	}
	return newDictReadHint(dictPtr, key, value), nil
}
