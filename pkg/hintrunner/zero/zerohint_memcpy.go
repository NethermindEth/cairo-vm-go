package zero

import (
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/core"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
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
	//> array_ptr = ids.array_ptr
	//> elm_size = ids.elm_size
	//> assert isinstance(elm_size, int) and elm_size > 0, \
	//> 	f'Invalid value for elm_size. Got: {elm_size}.'

	//> n_elms = ids.n_elms
	//> assert isinstance(n_elms, int) and n_elms >= 0, \
	//> 	f'Invalid value for n_elms. Got: {n_elms}.'
	//> if '__find_element_max_size' in globals():
	//> 	assert n_elms <= __find_element_max_size, \
	//> 		f'find_element() can only be used with n_elms<={__find_element_max_size}. ' \
	//> 		f'Got: n_elms={n_elms}.'

	//> for i in range(n_elms):
	//> 	if memory[array_ptr + elm_size * i] >= ids.key:
	//> 		ids.index = i
	//> 		break
	//> else:
	//> 	ids.index = n_elms

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
			//> elm_size = ids.elm_size
			//> assert isinstance(elm_size, int) and elm_size > 0, \
			//> 	f'Invalid value for elm_size. Got: {elm_size}.'

			//> n_elms = ids.n_elms
			//> assert isinstance(n_elms, int) and n_elms >= 0, \
			//> 	f'Invalid value for n_elms. Got: {n_elms}.'
			//> if '__find_element_max_size' in globals():
			//> 	assert n_elms <= __find_element_max_size, \
			//> 		f'find_element() can only be used with n_elms<={__find_element_max_size}. ' \
			//> 		f'Got: n_elms={n_elms}.'

			//> for i in range(n_elms):
			//> 	if memory[array_ptr + elm_size * i] >= ids.key:
			//> 		ids.index = i
			//> 		break
			//> else:
			//> 	ids.index = n_elms

			array_ptr, err := hinter.ResolveAsAddress(vm, array_ptr)
			if err != nil {
				return err
			}
			elm_size, err := hinter.ResolveAsFelt(vm, elm_size)
			if err != nil {
				return err
			}
			n_elms, err := hinter.ResolveAsFelt(vm, n_elms)
			if err != nil {
				return err
			}
			key, err := hinter.ResolveAsFelt(vm, key)
			if err != nil {
				return err
			}
			index, err := hinter.ResolveAsFelt(vm, index)
			if err != nil {
				return err
			}

			fmt.Print(array_ptr, elm_size, n_elms, key, index)

			return nil
		},
	}
}
