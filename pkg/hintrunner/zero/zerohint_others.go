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

func newMemcpyContinueCopyingHint(continueCopying hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "MemcpyContinueCopying",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> n -= 1
			//> ids.continue_copying = 1 if n > 0 else 0

			//> n-=1
			n, err := ctx.ScopeManager.GetVariableValue("n")
			if err != nil {
				return err
			}

			newN := new(fp.Element)
			newN = newN.Sub(n.(*fp.Element), &utils.FeltOne)

			if err := ctx.ScopeManager.AssignVariable("n", newN); err != nil {
				return err
			}

			//> ids.continue_copying = 1 if n > 0 else 0
			continueCopyingAddr, err := continueCopying.GetAddress(vm)
			if err != nil {
				return err
			}

			var continueCopyingMv memory.MemoryValue
			if utils.FeltLt(&utils.FeltZero, newN) {
				continueCopyingMv = memory.MemoryValueFromFieldElement(&utils.FeltOne)
			} else {
				continueCopyingMv = memory.MemoryValueFromFieldElement(&utils.FeltZero)
			}

			return vm.Memory.WriteToAddress(&continueCopyingAddr, &continueCopyingMv)
		},
	}
}

func createMemcpyContinueCopyingHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	continueCopying, err := resolver.GetResOperander("continue_copying")
	if err != nil {
		return nil, err
	}

	return newMemcpyContinueCopyingHint(continueCopying), nil
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

// MemcpyEnterScope hint enters a new scope for the memory copy operation with a specified length
//
// `newMemcpyEnterScopeHint` takes 1 operander as argument
//   - `len` is the length value that is added in the new scope
func newMemcpyEnterScopeHint(len hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "MemcpyEnterScope",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//>  vm_enter_scope({'n': ids.len})

			len, err := hinter.ResolveAsFelt(vm, len)
			if err != nil {
				return err
			}

			ctx.ScopeManager.EnterScope(map[string]any{"n": *len})
			return nil
		},
	}
}

func createMemcpyEnterScopeHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	len, err := resolver.GetResOperander("len")
	if err != nil {
		return nil, err
	}
	return newMemcpyEnterScopeHint(len), nil
}

// SearchSortedLower hint searches for the first element in a sorted array
// that is greater than or equal to a given key and returns its index
//
// `newSearchSortedLowerHint` takes 5 operanders as arguments
//   - `arrayPtr` represents the offset in the execution segment of memory where starts the sorted array
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
			arrayPtr, err := hinter.ResolveAsUint64(vm, arrayPtr)
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
			nElms, err := hinter.ResolveAsFelt(vm, nElms)
			if err != nil {
				return err
			}

			//> if '__find_element_max_size' in globals():
			//> 	assert n_elms <= __find_element_max_size, \
			//> 		f'find_element() can only be used with n_elms<={__find_element_max_size}. ' \
			//> 		f'Got: n_elms={n_elms}.'
			elementMaxSize, err := ctx.ScopeManager.GetVariableValue("__find_element_max_size")
			if err == nil {
				elementMaxSizeFelt, ok := elementMaxSize.(fp.Element)
				if !ok {
					return fmt.Errorf("failed obtaining the variable: __find_element_max_size")
				}

				if !utils.FeltLe(nElms, &elementMaxSizeFelt) {
					return fmt.Errorf("find_element() can only be used with n_elms<=%v. Got: n_elms=%v", elementMaxSizeFelt, nElms)
				}
			}

			key, err := hinter.ResolveAsFelt(vm, key)
			if err != nil {
				return err
			}

			indexAddr, err := index.GetAddress(vm)
			if err != nil {
				return err
			}

			nElemsRange := nElms.Uint64()

			//> for i in range(n_elms):
			for i := uint64(0); i < nElemsRange; i++ {
				//> 	if memory[array_ptr + elm_size * i] >= ids.key:
				index := arrayPtr + elmSize*i

				value, err := vm.Memory.ReadAsElement(uint64(1), index)
				if err != nil {
					return err
				}

				if utils.FeltLe(key, &value) {
					//> 		ids.index = i
					//> 		break
					indexValue := memory.MemoryValueFromFieldElement((new(fp.Element).SetUint64(i)))
					return vm.Memory.WriteToAddress(&indexAddr, &indexValue)
				}
			}

			//> else:
			//> 	ids.index = n_elms
			indexValue := memory.MemoryValueFromFieldElement(nElms)

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
