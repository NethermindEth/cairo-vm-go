package zero

import (
	"fmt"
	"math/big"
	"reflect"

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
			newN, err := hinter.GetVariableAs[uint64](&ctx.ScopeManager, "n")
			if err != nil {
				return err
			}

			newN -= 1

			if err := ctx.ScopeManager.AssignVariable("n", newN); err != nil {
				return err
			}

			//> ids.continue_loop/continue_copying = 1 if n > 0 else 0
			continueTargetAddr, err := continueTarget.GetAddress(vm)
			if err != nil {
				return err
			}

			var continueTargetMv memory.MemoryValue
			if newN > 0 {
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

			value, err := hinter.ResolveAsUint64(vm, value)
			if err != nil {
				return err
			}

			ctx.ScopeManager.EnterScope(map[string]any{"n": value})
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

// GetFeltBitLength hint assigns to `bit_length` the bit length of `x` variable
//
// `newGetFeltBitLengthHint` takes 2 operanders as arguments
//   - `x` is a felt variable
//   - `bit_length` is the variable that will store the bit length of x
func newGetFeltBitLengthHint(x, bitLength hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "GetFeltBitLength",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> x = ids.x
			//> ids.bit_length = x.bit_length()

			bitLengthAddr, err := bitLength.GetAddress(vm)
			if err != nil {
				return err
			}

			xVal, err := hinter.ResolveAsFelt(vm, x)
			if err != nil {
				return err
			}

			var xBig big.Int
			xVal.BigInt(&xBig)

			bitLen := xBig.BitLen()
			bitLenMv := memory.MemoryValueFromInt(bitLen)

			return vm.Memory.WriteToAddress(&bitLengthAddr, &bitLenMv)
		},
	}
}

func createGetFeltBitLengthHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	x, err := resolver.GetResOperander("x")
	if err != nil {
		return nil, err
	}

	bitLength, err := resolver.GetResOperander("bit_length")
	if err != nil {
		return nil, err
	}

	return newGetFeltBitLengthHint(x, bitLength), nil
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

			//> array_ptr = ids.array_ptr
			arrayPtrAddr, err := hinter.ResolveAsAddress(vm, arrayPtr)
			if err != nil {
				return err
			}

			//> elm_size = ids.elm_size
			elmSizeVal, err := hinter.ResolveAsUint64(vm, elmSize)
			if err != nil {
				return err
			}

			//> assert isinstance(elm_size, int) and elm_size > 0, \
			//>		f'Invalid value for elm_size. Got: {elm_size}.'
			if elmSizeVal == 0 {
				return fmt.Errorf("invalid value for elm_size. Got: %v", elmSizeVal)
			}

			//> key = ids.key
			keyVal, err := hinter.ResolveAsFelt(vm, key)
			if err != nil {
				return err
			}

			//> if '__find_element_index' in globals():
			//>		ids.index = __find_element_index
			findElementIndex, err := hinter.GetVariableAs[uint64](&ctx.ScopeManager, "__find_element_index")
			if err == nil {
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
				//>		found_key = memory[array_ptr + elm_size * __find_element_index]
				foundKey, err := vm.Memory.ReadFromAddressAsElement(arrayPtrAddr)
				if err != nil {
					return err
				}

				//>		assert found_key == key, \
				//>			f'Invalid index found in __find_element_index. index: {__find_element_index}, ' \
				//>			f'expected key {key}, found key: {found_key}.'
				if foundKey.Cmp(keyVal) != 0 {
					return fmt.Errorf("invalid index found in __find_element_index. index: %v, expected key %v, found key: %v", findElementIndex, keyVal, &foundKey)
				}

				//>		# Delete __find_element_index to make sure it's not used for the next calls.
				//>		del __find_element_index
				err = ctx.ScopeManager.DeleteVariable("__find_element_index")
				if err != nil {
					return err
				}

			} else {
				//>		assert isinstance(n_elms, int) and n_elms >= 0, \
				//>			f'Invalid value for n_elms. Got: {n_elms}.'
				//>		n_elms = ids.n_elms
				nElms, err := hinter.ResolveAsUint64(vm, nElms)
				if err != nil {
					return err
				}

				//>		if '__find_element_max_size' in globals():
				//>			assert n_elms <= __find_element_max_size, \
				//>				f'find_element() can only be used with n_elms<={__find_element_max_size}. ' \
				//>				f'Got: n_elms={n_elms}.'

				findElementMaxSize, err := hinter.GetVariableAs[uint64](&ctx.ScopeManager, "__find_element_max_size")
				if err == nil {
					if nElms > findElementMaxSize {
						return fmt.Errorf("find_element() can only be used with n_elms<=%v. Got: n_elms=%v", findElementMaxSize, nElms)
					}
				}

				//>		for i in range(n_elms):
				//>			if memory[array_ptr + elm_size * i] == key:
				//>				ids.index = i
				//>				break
				found := false
				for i := uint64(0); i < nElms; i++ {
					val, err := vm.Memory.ReadFromAddressAsElement(arrayPtrAddr)
					if err != nil {
						return err
					}
					if val.Cmp(keyVal) == 0 {
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
					*arrayPtrAddr, err = arrayPtrAddr.AddOffset(int16(elmSizeVal))
					if err != nil {
						return err
					}
				}

				//>			raise ValueError(f'Key {key} was not found.')
				if !found {
					return fmt.Errorf("key %v was not found", keyVal)
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

func newSetAddHint(elmSize, elmPtr, setPtr, setEndPtr, index, isElmInSet hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "SetAdd",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> assert ids.elm_size > 0
			//> assert ids.set_ptr <= ids.set_end_ptr
			//> elm_list = memory.get_range(ids.elm_ptr, ids.elm_size)
			//> for i in range(0, ids.set_end_ptr - ids.set_ptr, ids.elm_size):
			//>     if memory.get_range(ids.set_ptr + i, ids.elm_size) == elm_list:
			//>         ids.index = i // ids.elm_size
			//>         ids.is_elm_in_set = 1
			//>         break
			//>     else:
			//>         ids.is_elm_in_set = 0

			elmSize, err := hinter.ResolveAsUint64(vm, elmSize)
			if err != nil {
				return err
			}
			elmPtr, err := hinter.ResolveAsAddress(vm, elmPtr)
			if err != nil {
				return err
			}
			setPtr, err := hinter.ResolveAsAddress(vm, setPtr)
			if err != nil {
				return err
			}
			setEndPtr, err := hinter.ResolveAsAddress(vm, setEndPtr)
			if err != nil {
				return err
			}
			indexAddr, err := index.GetAddress(vm)
			if err != nil {
				return err
			}
			isElmInSetAddr, err := isElmInSet.GetAddress(vm)
			if err != nil {
				return err
			}

			//> assert ids.elm_size > 0
			if elmSize == 0 {
				return fmt.Errorf("assert ids.elm_size > 0 failed")
			}

			//> assert ids.set_ptr <= ids.set_end_ptr
			if setPtr.Offset > setEndPtr.Offset {
				return fmt.Errorf("assert ids.set_ptr <= ids.set_end_ptr failed")
			}

			//> elm_list = memory.get_range(ids.elm_ptr, ids.elm_size)
			elmList, err := vm.Memory.GetConsecutiveMemoryValues(*elmPtr, elmSize)
			if err != nil {
				return err
			}

			//> for i in range(0, ids.set_end_ptr - ids.set_ptr, ids.elm_size):
			//>     if memory.get_range(ids.set_ptr + i, ids.elm_size) == elm_list:
			//>         ids.index = i // ids.elm_size
			//>         ids.is_elm_in_set = 1
			//>         break
			//>     else:
			//>         ids.is_elm_in_set = 0
			isElmInSetFelt := utils.FeltZero
			totalSetLength := setEndPtr.Offset - setPtr.Offset
			for i := uint64(0); i < totalSetLength; i += elmSize {
				memoryElmList, err := vm.Memory.GetConsecutiveMemoryValues(*setPtr, elmSize)
				if err != nil {
					return err
				}
				*setPtr, err = setPtr.AddOffset(int16(elmSize))
				if err != nil {
					return err
				}
				if reflect.DeepEqual(memoryElmList, elmList) {
					indexFelt := fp.NewElement(i / elmSize)
					indexMv := memory.MemoryValueFromFieldElement(&indexFelt)
					err := vm.Memory.WriteToAddress(&indexAddr, &indexMv)
					if err != nil {
						return err
					}
					isElmInSetFelt = utils.FeltOne
					break
				}
			}

			mv := memory.MemoryValueFromFieldElement(&isElmInSetFelt)
			return vm.Memory.WriteToAddress(&isElmInSetAddr, &mv)
		},
	}
}

func createSetAddHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	elmSize, err := resolver.GetResOperander("elm_size")
	if err != nil {
		return nil, err
	}
	elmPtr, err := resolver.GetResOperander("elm_ptr")
	if err != nil {
		return nil, err
	}
	setPtr, err := resolver.GetResOperander("set_ptr")
	if err != nil {
		return nil, err
	}
	setEndPtr, err := resolver.GetResOperander("set_end_ptr")
	if err != nil {
		return nil, err
	}
	index, err := resolver.GetResOperander("index")
	if err != nil {
		return nil, err
	}
	isElmInSet, err := resolver.GetResOperander("is_elm_in_set")
	if err != nil {
		return nil, err
	}

	return newSetAddHint(elmSize, elmPtr, setPtr, setEndPtr, index, isElmInSet), nil
}

// SearchSortedLower hint searches for the first element in a sorted array
// that is greater than or equal to a given key and returns its index
//
// `newSearchSortedLowerHint` takes 5 operanders as arguments
//   - `arrayPtr` represents the address in memory where starts the sorted array
//   - `elmSize` is the size in terms of memory cells per element in the array
//   - `nElms` is the number of elements in the array
//   - `key` is the given key that acts a threshold
//   - `index` is the result, i.e., the index of the first element greater or equal to the given key
func newSearchSortedLowerHint(arrayPtr, elmSize, nElms, key, index hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "SearchSortedLower",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> array_ptr = ids.array_ptr
			//> elm_size = ids.elm_size
			//> assert isinstance(elm_size, int) and elm_size > 0, \
			//> 	f'Invalid value for elm_size. Got: {elm_size}.'
			//>
			//> n_elms = ids.n_elms
			//> assert isinstance(n_elms, int) and n_elms >= 0, \
			//> 	f'Invalid value for n_elms. Got: {n_elms}.'
			//> if '__find_element_max_size' in globals():
			//> 	assert n_elms <= __find_element_max_size, \
			//> 		f'find_element() can only be used with n_elms<={__find_element_max_size}. ' \
			//> 		f'Got: n_elms={n_elms}.'
			//>
			//> for i in range(n_elms):
			//> 	if memory[array_ptr + elm_size * i] >= ids.key:
			//> 		ids.index = i
			//> 		break
			//> else:
			//> 	ids.index = n_elms

			//> array_ptr = ids.array_ptr
			arrayPtr, err := hinter.ResolveAsAddress(vm, arrayPtr)
			if err != nil {
				return err
			}

			//> elm_size = ids.elm_size
			elmSize, err := hinter.ResolveAsUint64(vm, elmSize)
			if err != nil {
				return err
			}

			//> assert isinstance(elm_size, int) and elm_size > 0, \
			//> 	f'Invalid value for elm_size. Got: {elm_size}.'
			if elmSize == 0 {
				return fmt.Errorf("invalid value for elm_size. Got: %v", elmSize)
			}

			//> n_elms = ids.n_elms
			//> assert isinstance(n_elms, int) and n_elms >= 0, \
			//> 	f'Invalid value for n_elms. Got: {n_elms}.'
			nElms, err := hinter.ResolveAsUint64(vm, nElms)
			if err != nil {
				return err
			}

			//> if '__find_element_max_size' in globals():
			//> 	assert n_elms <= __find_element_max_size, \
			//> 		f'find_element() can only be used with n_elms<={__find_element_max_size}. ' \
			//> 		f'Got: n_elms={n_elms}.'
			elementMaxSize := uint64(1 << 20)
			if nElms > elementMaxSize {
				return fmt.Errorf("find_element() can only be used with n_elms<=%d.\n Got: length=%d", elementMaxSize, nElms)
			}

			key, err := hinter.ResolveAsFelt(vm, key)
			if err != nil {
				return err
			}

			indexAddr, err := index.GetAddress(vm)
			if err != nil {
				return err
			}

			index := arrayPtr

			//> for i in range(n_elms):
			for i := uint64(0); i < nElms; i++ {
				//> 	if memory[array_ptr + elm_size * i] >= ids.key:
				value, err := vm.Memory.ReadFromAddressAsElement(index)
				if err != nil {
					return err
				}

				if utils.FeltLe(key, &value) {
					//> 		ids.index = i
					//> 		break
					indexValue := memory.MemoryValueFromUint(i)
					return vm.Memory.WriteToAddress(&indexAddr, &indexValue)
				}

				*index, err = index.AddOffset(int16(elmSize))
				if err != nil {
					return err
				}
			}

			//> else:
			//> 	ids.index = n_elms
			indexValue := memory.MemoryValueFromInt(nElms)

			return vm.Memory.WriteToAddress(&indexAddr, &indexValue)
		},
	}
}

func createSearchSortedLowerHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	arrayPtr, err := resolver.GetResOperander("array_ptr")
	if err != nil {
		return nil, err
	}

	elmSize, err := resolver.GetResOperander("elm_size")
	if err != nil {
		return nil, err
	}

	nElms, err := resolver.GetResOperander("n_elms")
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

	return newSearchSortedLowerHint(arrayPtr, elmSize, nElms, key, index), nil
}

// NondetElementsOverX hint compares the offset difference between two memory address and
// writes 1 or 0 at `ap` memory address, depending on whether the difference is greater or
// equal to x or not
//
// `newNondetElementsOverXHint` takes 3 arguments
//   - `elementsEnd` represents the address in memory right after the last element of the array
//   - `elements` represents the address in memory of the first element of the array
//   - `x` represents the offset difference used to decide the result
func newNondetElementsOverXHint(elementsEnd, elements hinter.ResOperander, x uint64) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "NondetElementsOverX",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> python hint in cairo file: "ids.elements_end - ids.elements >= x"
			//> python hint in whitelist: "memory[ap] = to_felt_or_relocatable(ids.elements_end - ids.elements >= x)"

			elementsEndAddr, err := hinter.ResolveAsAddress(vm, elementsEnd)
			if err != nil {
				return err
			}
			elementsAddr, err := hinter.ResolveAsAddress(vm, elements)
			if err != nil {
				return err
			}

			apAddr := vm.Context.AddressAp()
			var resultMv memory.MemoryValue
			offsetDiff := elementsEndAddr.Offset - elementsAddr.Offset
			if offsetDiff >= x {
				resultMv = memory.MemoryValueFromFieldElement(&utils.FeltOne)
			} else {
				resultMv = memory.MemoryValueFromFieldElement(&utils.FeltZero)
			}

			return vm.Memory.WriteToAddress(&apAddr, &resultMv)
		},
	}
}

func createNondetElementsOverXHinter(resolver hintReferenceResolver, x uint64) (hinter.Hinter, error) {
	elementsEnd, err := resolver.GetResOperander("elements_end")
	if err != nil {
		return nil, err
	}
	elements, err := resolver.GetResOperander("elements")
	if err != nil {
		return nil, err
	}

	return newNondetElementsOverXHint(elementsEnd, elements, x), nil
}

// NormalizeAddress hint checks if given addr value is less than ADDR_BOUND or not
// and writes the result as 1 or 0 to is_small
//
// `newNormalizeAddressHint` takes 2 arguments
//   - `isSmall` represents the address where the result of the comparison is stored
//   - `addr` represents the address whose value is checked against ADDR_BOUND
func newNormalizeAddressHint(isSmall, addr hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "NormalizeAddress",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> # Verify the assumptions on the relationship between 2**250, ADDR_BOUND and PRIME.
			//> ADDR_BOUND = ids.ADDR_BOUND % PRIME
			//> assert (2**250 < ADDR_BOUND <= 2**251) and (2 * 2**250 < PRIME) and (
			//>         ADDR_BOUND * 2 > PRIME), \
			//>    'normalize_address() cannot be used with the current constants.'
			//> ids.is_small = 1 if ids.addr < ADDR_BOUND else 0

			// ADDR_BOUND and PRIME are constants: https://github.com/starkware-libs/cairo-lang/blob/0e4dab8a6065d80d1c726394f5d9d23cb451706a/src/starkware/starknet/common/storage.cairo#L6
			// so the assert check is skipped as the constants satisfy them

			// 2 ** 251 - 256
			addrBoundFelt := fp.Element{18446743986131443745, 160989183, 18446744073709255680, 576459263475590224}

			addrFelt, err := hinter.ResolveAsFelt(vm, addr)
			if err != nil {
				return err
			}
			isSmallAddr, err := isSmall.GetAddress(vm)
			if err != nil {
				return err
			}

			//> ids.is_small = 1 if ids.addr < ADDR_BOUND else 0
			var resultMv memory.MemoryValue
			if addrFelt.Cmp(&addrBoundFelt) < 0 {
				resultMv = memory.MemoryValueFromFieldElement(&utils.FeltOne)
			} else {
				resultMv = memory.MemoryValueFromFieldElement(&utils.FeltZero)
			}
			return vm.Memory.WriteToAddress(&isSmallAddr, &resultMv)
		},
	}
}

func createNormalizeAddressHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	isSmall, err := resolver.GetResOperander("is_small")
	if err != nil {
		return nil, err
	}
	addr, err := resolver.GetResOperander("addr")
	if err != nil {
		return nil, err
	}

	return newNormalizeAddressHint(isSmall, addr), nil
}

func newSha256AndBlake2sInputHint(low, high, value hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "Sha256AndBlake2sInput",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {

		},
	}
}

func createSha256AndBlake2sInputHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	return newSha256AndBlake2sInputHint()
}
