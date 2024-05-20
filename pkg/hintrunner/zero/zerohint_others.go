package zero

import (
	"fmt"
	"reflect"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/core"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"

	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func newMemcpyContinueCopyingHint(continueCopying hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "MemcpyContinueCopying",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> n -= 1
			//> ids.continue_copying = 1 if n > 0 else 0

			//> n-=1
			n_, err := ctx.ScopeManager.GetVariableValue("n")
			if err != nil {
				return err
			}
			n, ok := n_.(f.Element)
			if !ok {
				return fmt.Errorf("casting n_ into a felt failed")
			}

			newN := new(f.Element)
			newN = newN.Sub(&n, &utils.FeltOne)

			if err := ctx.ScopeManager.AssignVariable("n", *newN); err != nil {
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

			elmSize, err := hinter.ResolveAsFelt(vm, elmSize)
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

			elmSizeInt := elmSize.Uint64()

			//> assert ids.elm_size > 0
			if elmSize.IsZero() {
				return fmt.Errorf("assert ids.elm_size > 0 failed")
			}

			//> assert ids.set_ptr <= ids.set_end_ptr
			if setPtr.Offset > setEndPtr.Offset {
				return fmt.Errorf("assert ids.set_ptr <= ids.set_end_ptr failed")
			}

			//> elm_list = memory.get_range(ids.elm_ptr, ids.elm_size)
			elmList, err := vm.Memory.GetConsecutiveMemoryValues(*elmPtr, int16(elmSizeInt))
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
			for i := uint64(0); i < totalSetLength; i += elmSizeInt {
				memoryElmList, err := vm.Memory.GetConsecutiveMemoryValues(*setPtr, int16(elmSizeInt))
				if err != nil {
					return err
				}
				*setPtr, err = setPtr.AddOffset(int16(elmSizeInt))
				if err != nil {
					return err
				}
				if reflect.DeepEqual(memoryElmList, elmList) {
					indexFelt := f.NewElement(i / elmSizeInt)
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
