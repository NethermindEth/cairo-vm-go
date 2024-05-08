package zero

import (
	"fmt"
	"reflect"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

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
			elmList, err := hinter.GetConsecutiveValues(vm, *elmPtr, int16(elmSizeInt))
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
				memoryElmList, err := hinter.GetConsecutiveValues(vm, *setPtr, int16(elmSizeInt))
				if err != nil {
					return err
				}
				*setPtr, err = setPtr.AddOffset(int16(elmSizeInt))
				if err != nil {
					return err
				}
				if reflect.DeepEqual(memoryElmList, elmList) {
					indexFelt := fp.NewElement(i / elmSizeInt)
					indexMv := mem.MemoryValueFromFieldElement(&indexFelt)
					err := vm.Memory.WriteToAddress(&indexAddr, &indexMv)
					if err != nil {
						return err
					}
					isElmInSetFelt = utils.FeltOne
					break
				}
			}

			mv := mem.MemoryValueFromFieldElement(&isElmInSetFelt)
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
