package zero

import (
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

// DictNew hint creates a new dictionary
//
// `newDictNewHint` takes no operander as argument
func newDictNewHint() hinter.Hinter {
	return &GenericZeroHinter{
		Name: "DictNew",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> if '__dict_manager' not in globals():
			//>   from starkware.cairo.common.dict import DictManager
			//>   __dict_manager = DictManager()
			//>
			//> memory[ap] = __dict_manager.new_dict(segments, initial_dict)
			//> del initial_dict

			//> if '__dict_manager' not in globals():
			//>   from starkware.cairo.common.dict import DictManager
			//>   __dict_manager = DictManager()
			dictionaryManager, ok := ctx.ScopeManager.GetDictionaryManager()
			if !ok {
				newDictionaryManager := hinter.NewDictionaryManager()
				dictionaryManager = newDictionaryManager
				err := ctx.ScopeManager.AssignVariable("__dict_manager", dictionaryManager)
				if err != nil {
					return err
				}
			}

			initialDictValue, err := ctx.ScopeManager.GetVariableValue("initial_dict")
			if err != nil {
				return err
			}
			initialDict, ok := initialDictValue.(map[fp.Element]*memory.MemoryValue)
			if !ok {
				return fmt.Errorf("value: %s is not a *map[f.Element]*mem.MemoryValue", initialDictValue)
			}

			//> memory[ap] = __dict_manager.new_dict(segments, initial_dict)
			newDictAddr := dictionaryManager.NewDictionaryWithData(vm, initialDict)
			newDictAddrMv := memory.MemoryValueFromMemoryAddress(&newDictAddr)
			apAddr := vm.Context.AddressAp()
			err = vm.Memory.WriteToAddress(&apAddr, &newDictAddrMv)
			if err != nil {
				return err
			}

			//> del initial_dict
			return ctx.ScopeManager.DeleteVariable("initial_dict")
		},
	}
}

func createDictNewHinter() (hinter.Hinter, error) {
	return newDictNewHint(), nil
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
