package zero

import (
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/core"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func createAllocSegmentHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	return &core.AllocSegment{Dst: hinter.ApCellRef(0)}, nil
}

func createVMEnterScopeHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	return &GenericZeroHinter{
		Name: "VMEnterScope",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			ctx.ScopeManager.EnterScope(make(map[string]any))
			return nil
		},
	}, nil
}

func newMemcpyEnterScopeHint(len hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "MemcpyEnterScope",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//>  vm_enter_scope({'n': ids.len})
			len, err := hinter.ResolveAsFelt(vm, len)
			if err != nil {
				return err
			}
			ctx.ScopeManager.EnterScope(map[string]any{"n": len})
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

func createVMExitScopeHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	return &GenericZeroHinter{
		Name: "VMExitScope",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			return ctx.ScopeManager.ExitScope()
		},
	}, nil
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

func newSearchSortedLowerHint(arrayPtr, elmSize, nElms, key, index hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "SearchSortedLower",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {

			//> array_ptr = ids.array_ptr
			arrayPtr, err := hinter.ResolveAsAddress(vm, arrayPtr)
			if err != nil {
				return err
			}

			//> elm_size = ids.elm_size
			elmSize, err := hinter.ResolveAsFelt(vm, elmSize)
			if err != nil {
				return err
			}
			//> assert isinstance(elm_size, int) and elm_size > 0, \
			//> 	f'Invalid value for elm_size. Got: {elm_size}.'
			if !elmSize.IsZero() {
				return fmt.Errorf("Invalid value for elm_size. Got: %v.", elmSize)
			}

			//> n_elms = ids.n_elms
			//> assert isinstance(n_elms, int) and n_elms >= 0, \
			//> 	f'Invalid value for n_elms. Got: {n_elms}.'
			nElms, err := hinter.ResolveAsFelt(vm, nElms)

			if err != nil {
				return err
			}
			if !nElms.IsZero() {
				return fmt.Errorf("Invalid value for n_elms. Got: %v.", nElms)
			}

			//> if '__find_element_max_size' in globals():
			//> 	assert n_elms <= __find_element_max_size, \
			//> 		f'find_element() can only be used with n_elms<={__find_element_max_size}. ' \
			//> 		f'Got: n_elms={n_elms}.'
			elementMaxSize, err := ctx.ScopeManager.GetVariableValue("__find_element_max_size")
			if err != nil {
				return err
			}

			elementMaxSizeFelt, ok := elementMaxSize.(*fp.Element)
			if !ok {
				return fmt.Errorf("failed obtaining the variable: __find_element_max_size")
			}

			if !utils.FeltLe(nElms, elementMaxSizeFelt) {
				return fmt.Errorf("find_element() can only be used with n_elms<=%v. Got: n_elms=%v.", elementMaxSizeFelt, nElms)
			}

			key, err := hinter.ResolveAsFelt(vm, key)
			if err != nil {
				return err
			}

			indexAddr, err := index.GetAddress(vm)
			if err != nil {
				return err
			}

			// nElms should not be bigger than Uint64 due to felt logic
			nElemsRange := nElms.Uint64()
			elmSizeInt16 := int16(elmSize.Uint64())

			//> for i in range(n_elms):
			for i := uint64(0); i < nElemsRange; i++ {
				arrayPtrValue := memory.MemoryValueFromMemoryAddress(arrayPtr)
				arrayValue, err := arrayPtrValue.FieldElement()
				if err != nil {
					return err
				}

				//> 	if memory[array_ptr + elm_size * i] >= ids.key:
				if utils.FeltLe(key, arrayValue) {
					//> 		ids.index = i
					//> 		break
					indexValue := memory.MemoryValueFromFieldElement((new(fp.Element).SetUint64(i)))
					return vm.Memory.WriteToAddress(&indexAddr, &indexValue)
				}
				// This adds elmsSize
				arrayPtr.AddOffset(elmSizeInt16)
			}

			//> else:
			//> 	ids.index = n_elms
			indexValue := memory.MemoryValueFromFieldElement(nElms)

			return vm.Memory.WriteToAddress(&indexAddr, &indexValue)
		},
	}
}
