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

// SquashDictInnerContinueLoop hint determines if the loop should continue
// based on remaining access indices
//
// `newSquashDictInnerContinueLoopHint` takes 1 operander as argument
//   - `loopTemps` variable is a struct containing a `should_continue` field
//
// `newSquashDictInnerContinueLoopHint`writes 0 or 1 in the `should_continue` field
// depending on whether the `current_access_indices` array contains items or not
func newSquashDictInnerContinueLoopHint(loopTemps hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "SquashDictInnerContinueLoop",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> ids.loop_temps.should_continue = 1 if current_access_indices else 0

			currentAccessIndices_, err := ctx.ScopeManager.GetVariableValue("current_access_indices")
			if err != nil {
				return err
			}

			currentAccessIndices, ok := currentAccessIndices_.([]fp.Element)
			if !ok {
				return fmt.Errorf("casting currentAccessIndices_ into an array of felts failed")
			}

			loopTempsAddr, err := loopTemps.GetAddress(vm)
			if err != nil {
				return err
			}

			resultAddr, err := loopTempsAddr.AddOffset(3)
			if err != nil {
				return err
			}

			if len(currentAccessIndices) == 0 {
				resultMemZero := memory.MemoryValueFromFieldElement(&utils.FeltZero)
				return vm.Memory.WriteToAddress(&resultAddr, &resultMemZero)

			} else {
				resultMemOne := memory.MemoryValueFromFieldElement(&utils.FeltOne)
				return vm.Memory.WriteToAddress(&resultAddr, &resultMemOne)
			}
		},
	}
}

func createSquashDictInnerContinueLoopHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	loopTemps, err := resolver.GetResOperander("loop_temps")
	if err != nil {
		return nil, err
	}

	return newSquashDictInnerContinueLoopHint(loopTemps), nil
}

// SquashDictInnerAssertLenKeys hint asserts the length of the current
// access indices for a given key is zero
// `current_access_indices` is a reversed order list of access indices
// for a given key, i.e., `sorted(access_indices[key])[::-1]`
//
// `newSquashDictInnerAssertLenKeysHint` doesn't take any operander as argument
// and retrieves `current_access_indices` value from the current scope
func newSquashDictInnerLenAssertHint() hinter.Hinter {
	return &GenericZeroHinter{
		Name: "SquashDictInnerLenAssert",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> assert len(current_access_indices) == 0

			currentAccessIndices_, err := ctx.ScopeManager.GetVariableValue("current_access_indices")
			if err != nil {
				return err
			}

			currentAccessIndices := currentAccessIndices_.([]f.Element)
			if len(currentAccessIndices) != 0 {
				return fmt.Errorf("assertion `len(current_access_indices) == 0` failed")
			}

			return nil
		},
	}
}

func createSquashDictInnerLenAssertHinter() (hinter.Hinter, error) {
	return newSquashDictInnerLenAssertHint(), nil
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
