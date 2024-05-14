package zero

import (
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

// DefaultDictNew hint creates a new dictionary with a default value
//
// `newDefaultDictNewHint` takes 1 operander as argument
//   - `default_value` variable will be the default value
//     returned for keys not present in the dictionary
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
			defaultValueMv := memory.MemoryValueFromFieldElement(defaultValue)
			newDefaultDictionaryAddr := dictionaryManager.NewDefaultDictionary(vm, &defaultValueMv)
			newDefaultDictionaryAddrMv := memory.MemoryValueFromMemoryAddress(&newDefaultDictionaryAddr)
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

// DictWrite hint updates the value of given key for given dictionary
// while also writing the prev_value of the key to memory
//
// `newDictWriteHint` takes 3 operander as argument
//   - `dictPtr` variable will be pointer to the dictionary to update
//   - `key` variable will be the key whose value is updated in the dictionary
//   - `newValue` variable will be the new value for given `key` in the dictionary
func newDictWriteHint(dictPtr, key, newValue hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "DictWrite",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> dict_tracker = __dict_manager.get_tracker(ids.dict_ptr)
			//> dict_tracker.current_ptr += ids.DictAccess.SIZE
			//> ids.dict_ptr.prev_value = dict_tracker.data[ids.key]
			//> dict_tracker.data[ids.key] = ids.new_value

			//> dict_tracker = __dict_manager.get_tracker(ids.dict_ptr)
			dictPtrAddr, err := hinter.ResolveAsAddress(vm, dictPtr)
			if err != nil {
				return err
			}
			dictionaryManager, ok := ctx.ScopeManager.GetDictionaryManager()
			if !ok {
				return fmt.Errorf("__dict_manager not in scope")
			}
			dictionary, err := dictionaryManager.GetDictionary(dictPtrAddr)
			if err != nil {
				return err
			}

			//> dict_tracker.current_ptr += ids.DictAccess.SIZE
			dictionary.IncrementFreeOffset(3)

			key, err := hinter.ResolveAsFelt(vm, key)
			if err != nil {
				return err
			}

			//> ids.dict_ptr.prev_value = dict_tracker.data[ids.key]
			//> # dict_ptr points to a DictAccess
			//> struct DictAccess {
			//> 	key: felt,
			//> 	prev_value: felt,
			//> 	new_value: felt,
			//> }
			prevKeyValue, err := dictionary.At(key)
			if err != nil {
				return err
			}
			err = hinter.WriteToNthStructField(vm, *dictPtrAddr, *prevKeyValue, 1)
			if err != nil {
				return err
			}

			//> dict_tracker.data[ids.key] = ids.new_value
			newValue, err := hinter.ResolveAsFelt(vm, newValue)
			if err != nil {
				return err
			}
			newValueMv := memory.MemoryValueFromFieldElement(newValue)
			dictionary.Set(key, &newValueMv)

			return nil
		},
	}
}

func createDictWriteHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	dictPtr, err := resolver.GetResOperander("dict_ptr")
	if err != nil {
		return nil, err
	}
	key, err := resolver.GetResOperander("key")
	if err != nil {
		return nil, err
	}
	newValue, err := resolver.GetResOperander("new_value")
	if err != nil {
		return nil, err
	}
	return newDictWriteHint(dictPtr, key, newValue), nil
}

// DictUpdate hint updates the value of given key for given dictionary
// while also asserting the previous value of the key
//
// `newDictUpdateHint` takes 4 operander as argument
//   - `dictPtr` variable will be pointer to the dictionary to update
//   - `key` variable will be the key whose value is updated in the dictionary
//   - `newValue` variable will be the new value for given `key` in the dictionary
//   - `prevValue` variable will be the old value for given `key` in the dictionary
//     which will be asserted before the update
func newDictUpdateHint(dictPtr, key, newValue, prevValue hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "DictUpdate",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> # Verify dict pointer and prev value.
			//> dict_tracker = __dict_manager.get_tracker(ids.dict_ptr)
			//> current_value = dict_tracker.data[ids.key]
			//> assert current_value == ids.prev_value, \
			//>     f'Wrong previous value in dict. Got {ids.prev_value}, expected {current_value}.'
			//>
			//> # Update value.
			//> dict_tracker.data[ids.key] = ids.new_value
			//> dict_tracker.current_ptr += ids.DictAccess.SIZE

			//> dict_tracker = __dict_manager.get_tracker(ids.dict_ptr)
			dictPtrAddr, err := hinter.ResolveAsAddress(vm, dictPtr)
			if err != nil {
				return err
			}
			dictionaryManager, ok := ctx.ScopeManager.GetDictionaryManager()
			if !ok {
				return fmt.Errorf("__dict_manager not in scope")
			}
			dictionary, err := dictionaryManager.GetDictionary(dictPtrAddr)
			if err != nil {
				return err
			}

			key, err := hinter.ResolveAsFelt(vm, key)
			if err != nil {
				return err
			}

			//> current_value = dict_tracker.data[ids.key]
			currentValueMv, err := dictionary.At(key)
			if err != nil {
				return err
			}
			currentValue, err := currentValueMv.FieldElement()
			if err != nil {
				return err
			}

			//> assert current_value == ids.prev_value, \
			//>     f'Wrong previous value in dict. Got {ids.prev_value}, expected {current_value}.'
			prevValue, err := hinter.ResolveAsFelt(vm, prevValue)
			if err != nil {
				return err
			}
			if !currentValue.Equal(prevValue) {
				return fmt.Errorf("Wrong previous value in dict. Got %s, expected %s.", prevValue, currentValue)
			}

			//> # Update value.
			//> dict_tracker.data[ids.key] = ids.new_value
			newValue, err := hinter.ResolveAsFelt(vm, newValue)
			if err != nil {
				return err
			}
			newValueMv := memory.MemoryValueFromFieldElement(newValue)
			dictionary.Set(key, &newValueMv)

			//> dict_tracker.current_ptr += ids.DictAccess.SIZE
			dictionary.IncrementFreeOffset(3)

			return nil
		},
	}
}

func createDictUpdateHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	dictPtr, err := resolver.GetResOperander("dict_ptr")
	if err != nil {
		return nil, err
	}
	key, err := resolver.GetResOperander("key")
	if err != nil {
		return nil, err
	}
	newValue, err := resolver.GetResOperander("new_value")
	if err != nil {
		return nil, err
	}
	prevValue, err := resolver.GetResOperander("prev_value")
	if err != nil {
		return nil, err
	}
	return newDictUpdateHint(dictPtr, key, newValue, prevValue), nil
}

// SquashDictInnerAssertLenKeys hint asserts that the length
// of the `keys` descending list is zero during the squashing process
//
// `newSquashDictInnerAssertLenKeysHint` doesn't take any operander as argument
func newSquashDictInnerAssertLenKeysHint() hinter.Hinter {
	return &GenericZeroHinter{
		Name: "SquashDictInnerAssertLenKeys",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> assert len(keys) == 0
			keys_, err := ctx.ScopeManager.GetVariableValue("keys")
			if err != nil {
				return err
			}
			keys := keys_.([]f.Element)
			if len(keys) != 0 {
				return fmt.Errorf("assertion `len(keys) == 0` failed")
			}
			return nil
		},
	}
}

func createSquashDictInnerAssertLenKeysHinter() (hinter.Hinter, error) {
	return newSquashDictInnerAssertLenKeysHint(), nil
}

// SquashDictInnerNextKey hint retrieves the next key for processing during
// dictionary squashing after checking that the array of keys is not empty
//
// `newSquashDictInnerNextKeyHint` takes 1 operander as argument
//   - `next_key` variable will be assigned to the next key in `keys`
func newSquashDictInnerNextKeyHint(nextKey hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "SquashDictInnerNextKey",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> assert len(keys) > 0, 'No keys left but remaining_accesses > 0.'
			//> ids.next_key = key = keys.pop()

			keys_, err := ctx.ScopeManager.GetVariableValue("keys")
			if err != nil {
				return err
			}

			keys := keys_.([]f.Element)
			if len(keys) == 0 {
				return fmt.Errorf("no keys left but remaining_accesses > 0")
			}

			newKey, err := utils.Pop(&keys)
			if err != nil {
				return err
			}

			err = ctx.ScopeManager.AssignVariable("keys", keys)
			if err != nil {
				return err
			}

			err = ctx.ScopeManager.AssignVariable("key", newKey)
			if err != nil {
				return err
			}

			newKeyMemoryValue := memory.MemoryValueFromFieldElement(&newKey)

			addrNextKey, err := nextKey.GetAddress(vm)
			if err != nil {
				return err
			}

			return vm.Memory.WriteToAddress(&addrNextKey, &newKeyMemoryValue)
		},
	}
}

func createSquashDictInnerNextKeyHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	nextKey, err := resolver.GetResOperander("next_key")
	if err != nil {
		return nil, err
	}

	return newSquashDictInnerNextKeyHint(nextKey), nil
}
