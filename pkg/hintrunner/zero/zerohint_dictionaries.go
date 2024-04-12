package zero

import (
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

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
			initialDict, ok := initialDictValue.(map[fp.Element]*mem.MemoryValue)
			if !ok {
				return fmt.Errorf("value: %s is not a *map[f.Element]*mem.MemoryValue", initialDictValue)
			}

			//> memory[ap] = __dict_manager.new_dict(segments, initial_dict)
			newDictAddr := dictionaryManager.NewDictionaryWithData(vm, &initialDict)
			newDictAddrMv := mem.MemoryValueFromMemoryAddress(&newDictAddr)
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
