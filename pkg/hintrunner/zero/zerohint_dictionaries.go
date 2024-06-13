package zero

import (
	"fmt"
	"sort"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

//	struct DictAccess {
//	  key: felt,
//		 prev_value: felt,
//		 new_value: felt,
//	}
//
// The size of DictAccess is 3
const DictAccessSize = 3

// DictNew hint creates a new dictionary with its initial content seeded from a scope value "initial_dict"
// Querying the value of a key which doesn't exist in the dictionary returns an error
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
			dictionaryManager, ok := ctx.ScopeManager.GetZeroDictionaryManager()
			if !ok {
				dictionaryManager = hinter.NewZeroDictionaryManager()
				err := ctx.ScopeManager.AssignVariable("__dict_manager", dictionaryManager)
				if err != nil {
					return err
				}
			}

			initialDictValue, err := ctx.ScopeManager.GetVariableValue("initial_dict")
			if err != nil {
				return err
			}
			initialDict, ok := initialDictValue.(map[f.Element]memory.MemoryValue)
			if !ok {
				return fmt.Errorf("value: %s is not a map[f.Element]mem.MemoryValue", initialDictValue)
			}

			//> memory[ap] = __dict_manager.new_dict(segments, initial_dict)
			newDictAddr := dictionaryManager.NewDictionary(vm, initialDict)
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

// DefaultDictNew hint creates a new dictionary with a default value
// Querying the value of a key which doesn't exist in the dictionary returns the default value
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
			dictionaryManager, ok := ctx.ScopeManager.GetZeroDictionaryManager()
			if !ok {
				dictionaryManager = hinter.NewZeroDictionaryManager()
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
			newDefaultDictionaryAddr := dictionaryManager.NewDefaultDictionary(vm, defaultValueMv)
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

// DictRead hint accesses the value of a dictionary for a given key
// and writes it to a variable
//
// `newDictReadHint` takes 3 operanders as argument
//   - `dictPtr` variable will be pointer to the dictionary to read from
//   - `key` variable represents the key we are accessing the dictionary with
//   - `value` variable will hold the value of the dictionary stored at key after the hint is run
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
			dictionaryManager, ok := ctx.ScopeManager.GetZeroDictionaryManager()
			if !ok {
				return fmt.Errorf("__dict_manager not in scope")
			}

			//> dict_tracker.current_ptr += ids.DictAccess.SIZE
			err = dictionaryManager.IncrementFreeOffset(*dictPtr, DictAccessSize)
			if err != nil {
				return err
			}

			//> ids.value = dict_tracker.data[ids.key]
			key, err := hinter.ResolveAsFelt(vm, key)
			if err != nil {
				return err
			}
			keyValue, err := dictionaryManager.At(*dictPtr, *key)
			if err != nil {
				return err
			}
			valueAddr, err := value.GetAddress(vm)
			if err != nil {
				return err
			}
			return vm.Memory.WriteToAddress(&valueAddr, &keyValue)
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

// DictWrite hint writes a value for a given key in a dictionary
// and writes to memory the previous value for the key in the dictionary
//
// `newDictWriteHint` takes 3 operanders as argument
//   - `dictPtr` variable will be pointer to the dictionary to update
//   - `key` variable will be the key whose value is updated in the dictionary
//   - `newValue` variable will be the new value for given key in the dictionary
func newDictWriteHint(dictPtr, key, newValue hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "DictWrite",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> dict_tracker = __dict_manager.get_tracker(ids.dict_ptr)
			//> dict_tracker.current_ptr += ids.DictAccess.SIZE
			//> ids.dict_ptr.prev_value = dict_tracker.data[ids.key]
			//> dict_tracker.data[ids.key] = ids.new_value

			//> dict_tracker = __dict_manager.get_tracker(ids.dict_ptr)
			dictPtr, err := hinter.ResolveAsAddress(vm, dictPtr)
			if err != nil {
				return err
			}
			dictionaryManager, ok := ctx.ScopeManager.GetZeroDictionaryManager()
			if !ok {
				return fmt.Errorf("__dict_manager not in scope")
			}

			//> dict_tracker.current_ptr += ids.DictAccess.SIZE
			err = dictionaryManager.IncrementFreeOffset(*dictPtr, DictAccessSize)
			if err != nil {
				return err
			}

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
			prevKeyValue, err := dictionaryManager.At(*dictPtr, *key)
			if err != nil {
				return err
			}
			err = vm.Memory.WriteToNthStructField(*dictPtr, prevKeyValue, 1)
			if err != nil {
				return err
			}

			//> dict_tracker.data[ids.key] = ids.new_value
			newValue, err := hinter.ResolveAsFelt(vm, newValue)
			if err != nil {
				return err
			}
			newValueMv := memory.MemoryValueFromFieldElement(newValue)
			return dictionaryManager.Set(*dictPtr, *key, newValueMv)
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

// DictUpdate hint updates the value of given key in a dictionary
// and asserts the previous value of the key in the dictionary before the update
//
// `newDictUpdateHint` takes 4 operanders as argument
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
			dictPtr, err := hinter.ResolveAsAddress(vm, dictPtr)
			if err != nil {
				return err
			}
			dictionaryManager, ok := ctx.ScopeManager.GetZeroDictionaryManager()
			if !ok {
				return fmt.Errorf("__dict_manager not in scope")
			}

			key, err := hinter.ResolveAsFelt(vm, key)
			if err != nil {
				return err
			}

			//> current_value = dict_tracker.data[ids.key]
			currentValueMv, err := dictionaryManager.At(*dictPtr, *key)
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
			err = dictionaryManager.Set(*dictPtr, *key, newValueMv)
			if err != nil {
				return err
			}

			//> dict_tracker.current_ptr += ids.DictAccess.SIZE
			return dictionaryManager.IncrementFreeOffset(*dictPtr, DictAccessSize)
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

			keys := keys_.([]fp.Element)
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

// SquashDictInnerCheckAccessIndex hint updates the access index
// during the dictionary squashing process
//
// `newSquashDictInnerCheckAccessIndexHint` takes 1 operander as argument
//
//   - `loopTemps` variable is a struct containing `index_delta_minus1` as
//     a first field
//
//     struct LoopTemps {
//     index_delta_minus1: felt,
//     index_delta: felt,
//     ptr_delta: felt,
//     should_continue: felt,
//     }
//
// `newSquashDictInnerCheckAccessIndexHint` writes to the first field `index_delta_minus1`
// of the `loop_temps` struct the result of `new_access_index - current_access_index - 1`
// and assigns `new_access_index` to `current_access_index` in the scope
func newSquashDictInnerCheckAccessIndexHint(loopTemps hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "SquashDictInnerCheckAccessIndex",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> new_access_index = current_access_indices.pop()
			//> ids.loop_temps.index_delta_minus1 = new_access_index - current_access_index - 1
			//> current_access_index = new_access_index

			currentAccessIndices_, err := ctx.ScopeManager.GetVariableValue("current_access_indices")
			if err != nil {
				return err
			}

			currentAccessIndices, ok := currentAccessIndices_.([]fp.Element)
			if !ok {
				return fmt.Errorf("casting currentAccessIndices_ into an array of felts failed")
			}

			newAccessIndex, err := utils.Pop(&currentAccessIndices)
			if err != nil {
				return err
			}

			err = ctx.ScopeManager.AssignVariable("current_access_indices", currentAccessIndices)
			if err != nil {
				return err
			}

			currentAccessIndex_, err := ctx.ScopeManager.GetVariableValue("current_access_index")
			if err != nil {
				return err
			}

			currentAccessIndex, ok := currentAccessIndex_.(fp.Element)
			if !ok {
				return fmt.Errorf("casting currentAccessIndex_ into a felt failed")
			}

			err = ctx.ScopeManager.AssignVariable("current_access_index", newAccessIndex)
			if err != nil {
				return err
			}

			loopTempsAddr, err := loopTemps.GetAddress(vm)
			if err != nil {
				return err
			}

			var result fp.Element
			result.Sub(&newAccessIndex, &currentAccessIndex)
			result.Sub(&result, &utils.FeltOne)

			resultMem := memory.MemoryValueFromFieldElement(&result)

			// We use `WriteToAddress` function as we write to the first field of the `loop_temps` struct
			return vm.Memory.WriteToAddress(&loopTempsAddr, &resultMem)
		},
	}
}

func createSquashDictInnerCheckAccessIndexHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	loopTemps, err := resolver.GetResOperander("loop_temps")
	if err != nil {
		return nil, err
	}

	return newSquashDictInnerCheckAccessIndexHint(loopTemps), nil
}

// SquashDictInnerContinueLoop hint determines if the loop should continue
// based on remaining access indices
//
// `newSquashDictInnerContinueLoopHint` takes 1 operander as argument
//
//   - `loopTemps` variable is a struct containing a `should_continue` field
//
//     struct LoopTemps {
//     index_delta_minus1: felt,
//     index_delta: felt,
//     ptr_delta: felt,
//     should_continue: felt,
//     }
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

			if len(currentAccessIndices) == 0 {
				resultMemZero := memory.MemoryValueFromFieldElement(&utils.FeltZero)
				return vm.Memory.WriteToNthStructField(loopTempsAddr, resultMemZero, int16(3))

			} else {
				resultMemOne := memory.MemoryValueFromFieldElement(&utils.FeltOne)
				return vm.Memory.WriteToNthStructField(loopTempsAddr, resultMemOne, int16(3))
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

// SquashDictInnerFirstIteration hint sets up the first iteration
// of a loop for dictionary squashing, extracting `current_access_index`
// from the `current_access_indices` descending list
//
// `newSquashDictInnerFirstIterationHint` takes 1 operander as argument
//   - `rangeCheckPtr` is the address in memory where to write `current_access_index`
//
// `newSquashDictInnerFirstIterationHint`writes `current_access_index` at `rangeCheckPtr`
// offset in the execution segment of memory
func newSquashDictInnerFirstIterationHint(rangeCheckPtr hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "SquashDictInnerFirstIteration",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> current_access_indices = sorted(access_indices[key])[::-1]
			//> current_access_index = current_access_indices.pop()
			//> memory[ids.range_check_ptr] = current_access_index

			key_, err := ctx.ScopeManager.GetVariableValue("key")
			if err != nil {
				return err
			}

			accessIndices_, err := ctx.ScopeManager.GetVariableValue("access_indices")
			if err != nil {
				return err
			}

			accessIndices, ok := accessIndices_.(map[fp.Element][]fp.Element)
			if !ok {
				return fmt.Errorf("cannot cast access_indices_ to a mapping of felts")
			}

			key, ok := key_.(fp.Element)
			if !ok {
				return fmt.Errorf("cannot cast key_ to felt")
			}

			accessIndicesAtKey := accessIndices[key]

			accessIndicesAtKeyCopy := make([]fp.Element, len(accessIndicesAtKey))
			copy(accessIndicesAtKeyCopy, accessIndicesAtKey)

			sort.Slice(accessIndicesAtKeyCopy, func(i, j int) bool {
				return accessIndicesAtKeyCopy[i].Cmp(&accessIndicesAtKeyCopy[j]) > 0
			})

			currentAccessIndex, err := utils.Pop(&accessIndicesAtKeyCopy)
			if err != nil {
				return err
			}

			currentAccessIndexMv := memory.MemoryValueFromFieldElement(&currentAccessIndex)

			err = ctx.ScopeManager.AssignVariable("current_access_indices", accessIndicesAtKeyCopy)
			if err != nil {
				return err
			}

			err = ctx.ScopeManager.AssignVariable("current_access_index", currentAccessIndex)
			if err != nil {
				return err
			}

			rangeCheckPtrAddr, err := hinter.ResolveAsAddress(vm, rangeCheckPtr)
			if err != nil {
				return err
			}

			return vm.Memory.WriteToAddress(rangeCheckPtrAddr, &currentAccessIndexMv)
		},
	}
}

func createSquashDictInnerFirstIterationHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	rangeCheckPtr, err := resolver.GetResOperander("range_check_ptr")
	if err != nil {
		return nil, err
	}

	return newSquashDictInnerFirstIterationHint(rangeCheckPtr), nil
}

// SquashDictInnerSkipLoop hint determines if the loop should be skipped
// based on remaining access indices
//
// `newSquashDictInnerSkipLoopHint` takes 1 operander as argument
//   - `should_skip_loop` variable will be set to 0 or 1
//
// `newSquashDictInnerSkipLoopHint` writes 0 or 1 in the `should_skip_loop`variable
// depending on whether the `current_access_indices` array contains items or not
func newSquashDictInnerSkipLoopHint(shouldSkipLoop hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "SquashDictInnerSkipLoop",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> ids.should_skip_loop = 0 if current_access_indices else 1

			currentAccessIndices_, err := ctx.ScopeManager.GetVariableValue("current_access_indices")
			if err != nil {
				return err
			}

			currentAccessIndices, ok := currentAccessIndices_.([]fp.Element)
			if !ok {
				return fmt.Errorf("casting currentAccessIndices_ into an array of felts failed")
			}

			shouldSkipLoopAddr, err := shouldSkipLoop.GetAddress(vm)
			if err != nil {
				return err
			}

			if len(currentAccessIndices) == 0 {
				resultMemOne := memory.MemoryValueFromFieldElement(&utils.FeltOne)
				return vm.Memory.WriteToAddress(&shouldSkipLoopAddr, &resultMemOne)

			} else {
				resultMemZero := memory.MemoryValueFromFieldElement(&utils.FeltZero)
				return vm.Memory.WriteToAddress(&shouldSkipLoopAddr, &resultMemZero)
			}
		},
	}
}

func createSquashDictInnerSkipLoopHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	shouldSkipLoop, err := resolver.GetResOperander("should_skip_loop")
	if err != nil {
		return nil, err
	}

	return newSquashDictInnerSkipLoopHint(shouldSkipLoop), nil
}

// SquashDictInnerAssertLenKeys hint asserts the length of the current
// access indices for a given key is zero
// `current_access_indices` is a reversed order list of access indices
// for a given key, i.e., `sorted(access_indices[key])[::-1]`
//
// `newSquashDictInnerLenAssertHint` doesn't take any operander as argument
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

			currentAccessIndices := currentAccessIndices_.([]fp.Element)
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

			keys := keys_.([]fp.Element)
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
			accessIndices_, err := ctx.ScopeManager.GetVariableValue("access_indices")
			if err != nil {
				return err
			}

			accessIndices, ok := accessIndices_.(map[fp.Element][]fp.Element)
			if !ok {
				return fmt.Errorf("cannot cast access_indices_ to a mapping of felts")
			}

			key_, err := ctx.ScopeManager.GetVariableValue("key")
			if err != nil {
				return err
			}

			key, ok := key_.(fp.Element)
			if !ok {
				return fmt.Errorf("cannot cast key_ to felt")
			}

			accessIndicesAtKeyLen := uint64(len(accessIndices[key]))

			nUsedAccesses, err := hinter.ResolveAsUint64(vm, nUsedAccesses)
			if err != nil {
				return err
			}

			if accessIndicesAtKeyLen != nUsedAccesses {
				return fmt.Errorf("assertion ids.n_used_accesses == len(access_indices[key]) failed")
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
