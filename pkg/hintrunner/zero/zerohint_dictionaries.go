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

// SquashDictInnerUsedAccessesAssert hint checks that `n_used_accesses` Cairo local variable
// is equal to the the number of used accesses for a key during dictionary squashing
//
// `newSquashDictInnerUsedAccessesAssertHint` takes one operander as argument
//   - `nUsedAccesses` represents the number of used accesses for a given key
func newSquashDictInnerUsedAccessesAssertHint(nUsedAccesses hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "SquashDictInnerUsedAccessesAssert",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> assert ids.n_used_accesses == len(access_indices[key])
			access_indices_, err := ctx.ScopeManager.GetVariableValue("access_indices")
			if err != nil {
				return err
			}

			access_indices := access_indices_.(map[fp.Element][]fp.Element)

			key_, err := ctx.ScopeManager.GetVariableValue("key")
			if err != nil {
				return err
			}

			key := key_.(fp.Element)

			accessIndicesAtKey := uint64(len(access_indices[key]))

			nUsedAccessesField, err := hinter.ResolveAsUint64(vm, nUsedAccesses)
			if err != nil {
				return err
			}

			if accessIndicesAtKey != nUsedAccessesField {
				return fmt.Errorf("assertion ids.n_used_accesses == len(access_indices[key] failed")
			}

			return nil
		},
	}
}

func createSquashDictInnerUsedAccessesAssertHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	nUsedAccesses, err := resolver.GetResOperander("n_used_accesses")
	if err != nil {
		return nil, err
	}

	return newSquashDictInnerUsedAccessesAssertHint(nUsedAccesses), nil
}
