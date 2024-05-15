package zero

import (
	"fmt"
	"sort"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"golang.org/x/exp/maps"
)

func newSquashDictHint(dictAccesses, ptrDiff, nAccesses, bigKeys, firstKey hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "SquashDict",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> dict_access_size = ids.DictAccess.SIZE
			//> address = ids.dict_accesses.address_
			//> assert ids.ptr_diff % dict_access_size == 0, \
			//>     'Accesses array size must be divisible by DictAccess.SIZE'
			//> n_accesses = ids.n_accesses
			//> if '__squash_dict_max_size' in globals():
			//>     assert n_accesses <= __squash_dict_max_size, \
			//>         f'squash_dict() can only be used with n_accesses<={__squash_dict_max_size}. ' \
			//>         f'Got: n_accesses={n_accesses}.'
			//> # A map from key to the list of indices accessing it.
			//> access_indices = {}
			//> for i in range(n_accesses):
			//>     key = memory[address + dict_access_size * i]
			//>     access_indices.setdefault(key, []).append(i)
			//> # Descending list of keys.
			//> keys = sorted(access_indices.keys(), reverse=True)
			//> # Are the keys used bigger than range_check bound.
			//> ids.big_keys = 1 if keys[0] >= range_check_builtin.bound else 0
			//> ids.first_key = key = keys.pop()

			//> dict_access_size = ids.DictAccess.SIZE
			dictAccessSize := uint64(3)

			//> address = ids.dict_accesses.address_
			address, err := dictAccesses.GetAddress(vm)
			if err != nil {
				return err
			}

			//> assert ids.ptr_diff % dict_access_size == 0, \
			//>     'Accesses array size must be divisible by DictAccess.SIZE'
			ptrDiffValue, err := hinter.ResolveAsUint64(vm, ptrDiff)
			if err != nil {
				return err
			}
			if ptrDiffValue%dictAccessSize != 0 {
				return fmt.Errorf("Accesses array size must be divisible by DictAccess.SIZE")
			}

			//> n_accesses = ids.n_accesses
			nAccessesValue, err := hinter.ResolveAsUint64(vm, nAccesses)
			if err != nil {
				return err
			}

			//> if '__squash_dict_max_size' in globals():
			//>     assert n_accesses <= __squash_dict_max_size, \
			//>         f'squash_dict() can only be used with n_accesses<={__squash_dict_max_size}. ' \
			//>         f'Got: n_accesses={n_accesses}.'
			//  __squash_dict_max_size is always in scope and has a value of 2**20,
			squashDictMaxSize := uint64(1048576)
			if nAccessesValue > squashDictMaxSize {
				return fmt.Errorf("squash_dict() can only be used with n_accesses<={%d}. Got: n_accesses={%d}.", squashDictMaxSize, nAccessesValue)
			}

			//> # A map from key to the list of indices accessing it.
			//> access_indices = {}
			//> for i in range(n_accesses):
			//>     key = memory[address + dict_access_size * i]
			//>     access_indices.setdefault(key, []).append(i)
			accessIndices := make(map[f.Element][]uint64)
			for i := uint64(0); i < nAccessesValue; i++ {
				memoryAddress, err := address.AddOffset(int16(dictAccessSize * i))
				if err != nil {
					return err
				}
				key, err := vm.Memory.ReadFromAddressAsElement(&memoryAddress)
				if err != nil {
					return err
				}
				accessIndices[key] = append(accessIndices[key], i)
			}

			//> # Descending list of keys.
			//> keys = sorted(access_indices.keys(), reverse=True)
			keys := maps.Keys(accessIndices)
			if len(keys) == 0 {
				return fmt.Errorf("empty keys array")
			}
			sort.Slice(keys, func(i, j int) bool {
				return keys[i].Cmp(&keys[j]) > 0
			})

			//> ids.big_keys = 1 if keys[0] >= range_check_builtin.bound else 0
			bigKeysAddr, err := bigKeys.GetAddress(vm)
			if err != nil {
				return err
			}
			var bigKeysMv memory.MemoryValue
			if utils.FeltIsPositive(&keys[0]) {
				bigKeysMv = memory.MemoryValueFromFieldElement(&utils.FeltZero)
			} else {
				bigKeysMv = memory.MemoryValueFromFieldElement(&utils.FeltOne)
			}
			err = vm.Memory.WriteToAddress(&bigKeysAddr, &bigKeysMv)
			if err != nil {
				return err
			}

			//> ids.first_key = key = keys.pop()
			firstKeyAddr, err := firstKey.GetAddress(vm)
			if err != nil {
				return err
			}
			firstKeyValue := keys[len(keys)-1]
			firstKeyMv := memory.MemoryValueFromFieldElement(&firstKeyValue)
			keys = keys[:len(keys)-1]
			err = vm.Memory.WriteToAddress(&firstKeyAddr, &firstKeyMv)
			if err != nil {
				return err
			}

			err = ctx.ScopeManager.AssignVariable("access_indices", accessIndices)
			if err != nil {
				return err
			}
			err = ctx.ScopeManager.AssignVariable("keys", keys)
			if err != nil {
				return err
			}
			return ctx.ScopeManager.AssignVariable("key", firstKeyValue)
		},
	}
}

func createSquashDictHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	dictAccesses, err := resolver.GetResOperander("dict_accesses")
	if err != nil {
		return nil, err
	}
	ptrDiff, err := resolver.GetResOperander("ptr_diff")
	if err != nil {
		return nil, err
	}
	nAccesses, err := resolver.GetResOperander("n_accesses")
	if err != nil {
		return nil, err
	}
	bigKeys, err := resolver.GetResOperander("big_keys")
	if err != nil {
		return nil, err
	}
	firstKey, err := resolver.GetResOperander("first_key")
	if err != nil {
		return nil, err
	}

	return newSquashDictHint(dictAccesses, ptrDiff, nAccesses, bigKeys, firstKey), nil
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
