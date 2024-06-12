package zero

import (
	"fmt"
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
			elmList, err := vm.Memory.GetConsecutiveMemoryValues(*elmPtr, int16(elmSize))
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
				memoryElmList, err := vm.Memory.GetConsecutiveMemoryValues(*setPtr, int16(elmSize))
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
