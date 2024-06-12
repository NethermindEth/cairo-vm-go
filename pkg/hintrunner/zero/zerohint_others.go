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
					indexValue := memory.MemoryValueFromFieldElement((new(fp.Element).SetUint64(i)))
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
