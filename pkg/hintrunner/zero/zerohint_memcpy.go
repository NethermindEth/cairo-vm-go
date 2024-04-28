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

func createVMExitScopeHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	return &GenericZeroHinter{
		Name: "VMExitScope",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			return ctx.ScopeManager.ExitScope()
		},
	}, nil
}

func createSearchSortedLowerHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	array_ptr, err := resolver.GetResOperander("array_ptr")
	if err != nil {
		return nil, err
	}
	elm_size, err := resolver.GetResOperander("elm_size")
	if err != nil {
		return nil, err
	}
	n_elms, err := resolver.GetResOperander("n_elms")
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
	return newSearchSortedLowerHint(array_ptr, elm_size, n_elms, key, index), nil
}

func newSearchSortedLowerHint(array_ptr, elm_size, n_elms, key, index hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "SearchSortedLower",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {

			//> array_ptr = ids.array_ptr
			array_ptr, err := hinter.ResolveAsFelt(vm, array_ptr)
			if err != nil {
				return err
			}

			//> elm_size = ids.elm_size
			elm_size, err := hinter.ResolveAsFelt(vm, elm_size)

			//> assert isinstance(elm_size, int) and elm_size > 0, \
			//> 	f'Invalid value for elm_size. Got: {elm_size}.'
			if err != nil && utils.FeltLt(&utils.FeltZero, elm_size) {
				return fmt.Errorf("Invalid value for elm_size. Got: %v.", elm_size)
			}

			//> n_elms = ids.n_elms
			//> assert isinstance(n_elms, int) and n_elms >= 0, \
			//> 	f'Invalid value for n_elms. Got: {n_elms}.'
			n_elms, err := hinter.ResolveAsFelt(vm, n_elms)
			if err != nil && utils.FeltLt(&utils.FeltZero, n_elms) {
				return fmt.Errorf("Invalid value for n_elms. Got: %v.", n_elms)
			}

			var __find_element_max_size *fp.Element
			// TODO: How to do this --> if __find_element_max_size is a variable?
			//> if '__find_element_max_size' in globals():
			//> 	assert n_elms <= __find_element_max_size, \
			//> 		f'find_element() can only be used with n_elms<={__find_element_max_size}. ' \
			//> 		f'Got: n_elms={n_elms}.'

			if !utils.FeltLe(n_elms, __find_element_max_size) {
				return fmt.Errorf("find_element() can only be used with n_elms<=%v. Got: n_elms=%v.", __find_element_max_size, n_elms)
			}

			key, err := hinter.ResolveAsFelt(vm, key)
			if err != nil {
				return err
			}

			indexAddr, err := index.GetAddress(vm)
			if err != nil {
				return err
			}
			//> for i in range(n_elms):
			//> 	if memory[array_ptr + elm_size * i] >= ids.key:
			//> 		ids.index = i
			//> 		break
			for i := uint64(0); i < n_elms.Uint64(); i++ {
				v := memory.MemoryValueFromFieldElement(utils.Felt127.Add(array_ptr, elm_size))
				e, _ := v.FieldElement()

				if utils.FeltLt(key, e) {
					indexValue := memory.MemoryValueFromFieldElement((new(fp.Element).SetInt64(int64(i))))
					return vm.Memory.WriteToAddress(&indexAddr, &indexValue)
				}
			}

			//> else:
			//> 	ids.index = n_elms
			indexValue := memory.MemoryValueFromFieldElement(n_elms)

			return vm.Memory.WriteToAddress(&indexAddr, &indexValue)
		},
	}
}
