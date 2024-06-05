package zero

import (
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/core"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"

	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func newMemContinueHint(continueTarget hinter.ResOperander, memset bool) hinter.Hinter {
	var name string
	if memset {
		name = "MemsetContinueLoop"
	} else {
		name = "MemcpyContinueCopying"
	}
	return &GenericZeroHinter{
		Name: name,
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			// MemsetContinueLoop
			//> n -= 1
			//> ids.continue_loop = 1 if n > 0 else 0

			// MemcpyContinueCopying
			//> n -= 1
			//> ids.continue_copying = 1 if n > 0 else 0

			//> n-=1
			n, err := ctx.ScopeManager.GetVariableValue("n")
			if err != nil {
				return err
			}

			newN, ok := n.(fp.Element)
			if !ok {
				return fmt.Errorf("casting n into a felt failed")
			}

			newN.Sub(&newN, &utils.FeltOne)

			if err := ctx.ScopeManager.AssignVariable("n", newN); err != nil {
				return err
			}

			//> ids.continue_loop/continue_copying = 1 if n > 0 else 0
			continueTargetAddr, err := continueTarget.GetAddress(vm)
			if err != nil {
				return err
			}

			var continueTargetMv memory.MemoryValue
			if utils.FeltLt(&utils.FeltZero, &newN) {
				continueTargetMv = memory.MemoryValueFromFieldElement(&utils.FeltOne)
			} else {
				continueTargetMv = memory.MemoryValueFromFieldElement(&utils.FeltZero)
			}

			return vm.Memory.WriteToAddress(&continueTargetAddr, &continueTargetMv)
		},
	}
}

func createMemContinueHinter(resolver hintReferenceResolver, memset bool) (hinter.Hinter, error) {
	var continueTarget hinter.ResOperander
	var err error
	if memset {
		continueTarget, err = resolver.GetResOperander("continue_loop")
	} else {
		continueTarget, err = resolver.GetResOperander("continue_copying")
	}
	if err != nil {
		return nil, err
	}
	return newMemContinueHint(continueTarget, memset), nil
}

// AllocSegment hint adds a new segment to the Cairo VM memory
func createAllocSegmentHinter() (hinter.Hinter, error) {
	return &core.AllocSegment{Dst: hinter.ApCellRef(0)}, nil
}

// VMEnterScope hint enters a new scope in the Cairo VM
func createVMEnterScopeHinter() (hinter.Hinter, error) {
	return &GenericZeroHinter{
		Name: "VMEnterScope",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			ctx.ScopeManager.EnterScope(make(map[string]any))
			return nil
		},
	}, nil
}

// VMExitScop hint exits the current scope in the Cairo VM
func createVMExitScopeHinter() (hinter.Hinter, error) {
	return &GenericZeroHinter{
		Name: "VMExitScope",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			return ctx.ScopeManager.ExitScope()
		},
	}, nil
}

// MemEnterScope hint enters a new scope for the memory copy/set operation with a specified value
//
// `newMemEnterScopeHint` takes 2 operanders as arguments
//   - `value` is the value that is added in the new scope
//   - `memset` specifies whether it's a memset or memcpy operation
func newMemEnterScopeHint(value hinter.ResOperander, memset bool) hinter.Hinter {
	var name string
	if memset {
		name = "MemsetEnterScope"
	} else {
		name = "MemcpyEnterScope"
	}
	return &GenericZeroHinter{
		Name: name,
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			// MemsetEnterScope
			//> vm_enter_scope({'n': ids.n})

			// MemcpyEnterScope
			//> vm_enter_scope({'n': ids.len})

			value, err := hinter.ResolveAsFelt(vm, value)
			if err != nil {
				return err
			}

			ctx.ScopeManager.EnterScope(map[string]any{"n": *value})
			return nil
		},
	}
}

func createMemEnterScopeHinter(resolver hintReferenceResolver, memset bool) (hinter.Hinter, error) {
	var value hinter.ResOperander
	var err error
	if memset {
		value, err = resolver.GetResOperander("n")
	} else {
		value, err = resolver.GetResOperander("len")
	}
	if err != nil {
		return nil, err
	}
	return newMemEnterScopeHint(value, memset), nil
}

// FindElement hint finds element in the array by given key. It either returns element at index provided by __find_element_index or searches for the key in the array, returning error if key wasn't found.
//
// `newFindElementHint` takes 5 operanders as arguments
//   - `arrayPtr` is the pointer to the base of the array in memory
//   - `elmSize` is the size of the element in the array (the number of memory cells that the element occupies)
//   - `key` is the felt key to search for in the array
//   - `index` is the address in memory where to write the index of the found element in the array
//   - `nElms` is the number of elements in the array
func newFindElementHint(arrayPtr, elmSize, key, index, nElms hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "FindElement",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> array_ptr = ids.array_ptr
			//> elm_size = ids.elm_size
			//> assert isinstance(elm_size, int) and elm_size > 0, \
			//>		f'Invalid value for elm_size. Got: {elm_size}.'
			//> key = ids.key
			//>
			//> if '__find_element_index' in globals():
			//>		ids.index = __find_element_index
			//>		found_key = memory[array_ptr + elm_size * __find_element_index]
			//>		assert found_key == key, \
			//>			f'Invalid index found in __find_element_index. index: {__find_element_index}, ' \
			//>			f'expected key {key}, found key: {found_key}.'
			//>		# Delete __find_element_index to make sure it's not used for the next calls.
			//>		del __find_element_index
			//> else:
			//>		n_elms = ids.n_elms
			//>		assert isinstance(n_elms, int) and n_elms >= 0, \
			//>			f'Invalid value for n_elms. Got: {n_elms}.'
			//>		if '__find_element_max_size' in globals():
			//>			assert n_elms <= __find_element_max_size, \
			//>				f'find_element() can only be used with n_elms<={__find_element_max_size}. ' \
			//>				f'Got: n_elms={n_elms}.'
			//>
			//>		for i in range(n_elms):
			//>			if memory[array_ptr + elm_size * i] == key:
			//>				ids.index = i
			//>				break
			//>		else:
			//>			raise ValueError(f'Key {key} was not found.')
			arrayPtrAddr, err := hinter.ResolveAsAddress(vm, arrayPtr)
			if err != nil {
				return err
			}
			elmSizeVal, err := hinter.ResolveAsUint64(vm, elmSize)
			if err != nil {
				return err
			}
			if elmSizeVal == 0 {
				return fmt.Errorf("Invalid value for elm_size. Got: %v", elmSizeVal)
			}
			keyVal, err := hinter.ResolveAsFelt(vm, key)
			if err != nil {
				return err
			}
			findElementIndex, err := ctx.ScopeManager.GetVariableValue("__find_element_index")
			if err == nil {
				findElementIndex := findElementIndex.(uint64)
				findElementIndexFelt := new(fp.Element).SetUint64(findElementIndex)
				findElementIndexMemoryValue := memory.MemoryValueFromFieldElement(findElementIndexFelt)
				indexValAddr, err := index.GetAddress(vm)
				if err != nil {
					return err
				}
				err = vm.Memory.WriteToAddress(&indexValAddr, &findElementIndexMemoryValue)
				if err != nil {
					return err
				}
				arrayPtrAddr.Offset = arrayPtrAddr.Offset + elmSizeVal*findElementIndex
				foundKey, err := vm.Memory.ReadFromAddress(arrayPtrAddr)
				if err != nil {
					return err
				}
				foundKeyVal, err := foundKey.FieldElement()
				if err != nil {
					return err
				}
				if foundKeyVal.Cmp(keyVal) != 0 {
					return fmt.Errorf("Invalid index found in __find_element_index. index: %v, expected key %v, found key: %v", findElementIndex, keyVal, &foundKey)
				}
				err = ctx.ScopeManager.DeleteVariable("__find_element_index")
				if err != nil {
					return err
				}
			} else {
				nElms, err := hinter.ResolveAsUint64(vm, nElms)
				if err != nil {
					return err
				}
				findElementMaxSize, err := ctx.ScopeManager.GetVariableValue("__find_element_max_size")
				if err == nil {
					findElementMaxSize := findElementMaxSize.(uint64)
					if nElms > findElementMaxSize {
						return fmt.Errorf("find_element() can only be used with n_elms<=%v. Got: n_elms=%v", findElementMaxSize, nElms)
					}
				}
				found := false
				for i := uint64(0); i < nElms; i++ {
					val, err := vm.Memory.ReadFromAddress(arrayPtrAddr)
					if err != nil {
						return err
					}
					valFelt, err := val.FieldElement()
					if err != nil {
						return err
					}
					if valFelt.Cmp(keyVal) == 0 {
						indexValAddr, err := index.GetAddress(vm)
						if err != nil {
							return err
						}
						iFelt := new(fp.Element).SetUint64(i)
						iFeltMemoryValue := memory.MemoryValueFromFieldElement(iFelt)
						err = vm.Memory.WriteToAddress(&indexValAddr, &iFeltMemoryValue)
						if err != nil {
							return err
						}
						found = true
						break
					}
					// TODO: Check if this overflows using integration tests
					arrayPtrAddr.Offset = arrayPtrAddr.Offset + elmSizeVal
				}
				if !found {
					return fmt.Errorf("Key %v was not found", keyVal)
				}
			}
			return nil
		},
	}
}

func createFindElementHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	arrayPtr, err := resolver.GetResOperander("array_ptr")
	if err != nil {
		return nil, err
	}
	elmSize, err := resolver.GetResOperander("elm_size")
	if err != nil {
		return nil, err
	}
	key, err := resolver.GetResOperander("key")
	if err != nil {
		return nil, err
	}
	index, err := resolver.GetResOperander("index")
	if err != nil {
		return nil, err
	}
	nElms, err := resolver.GetResOperander("n_elms")
	if err != nil {
		return nil, err
	}
	return newFindElementHint(arrayPtr, elmSize, key, index, nElms), nil
}
