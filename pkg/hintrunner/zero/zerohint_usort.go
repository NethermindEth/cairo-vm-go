package zero

import (
	"fmt"
	"sort"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	usortUtils "github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/utils"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func createUsortBodyHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	input, err := resolver.GetResOperander("input")
	if err != nil {
		return nil, err
	}
	input_len, err := resolver.GetResOperander("input_len")
	if err != nil {
		return nil, err
	}
	output, err := resolver.GetResOperander("output")
	if err != nil {
		return nil, err
	}
	output_len, err := resolver.GetResOperander("output_len")
	if err != nil {
		return nil, err
	}
	multiplicities, err := resolver.GetResOperander("multiplicities")
	if err != nil {
		return nil, err
	}
	return newUsortBodyHint(input, input_len, output, output_len, multiplicities), nil
}

func newUsortBodyHint(input, input_len, output, output_len, multiplicities hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "AssertLtFelt",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> 	from collections import defaultdict
			//>
			//> 		input_ptr = ids.input
			//> 		input_len = int(ids.input_len)
			//> 		if __usort_max_size is not None:
			//> 			assert input_len <= __usort_max_size, (
			//> 				f"usort() can only be used with input_len<={__usort_max_size}. "
			//> 				f"Got: input_len={input_len}."
			//> 			)
			//>
			//> 		positions_dict = defaultdict(list)
			//> 		for i in range(input_len):
			//> 			val = memory[input_ptr + i]
			//> 			positions_dict[val].append(i)
			//>
			//> 		output = sorted(positions_dict.keys())
			//> 		ids.output_len = len(output)
			//> 		ids.output = segments.gen_arg(output)
			//> 		ids.multiplicities = segments.gen_arg([len(positions_dict[k]) for k in output])
			//>
			//> 		input_ptr = ids.input
			inputBasePtr, err := hinter.ResolveAsAddress(vm, input)
			if err != nil {
				return err
			}
			inputLen, err := hinter.ResolveAsUint64(vm, input_len)
			if err != nil {
				return err
			}
			usortMaxSizeInterface, err := ctx.ScopeManager.GetVariableValue("__usort_max_size")
			if err != nil {
				return err
			}
			usortMaxSize := usortMaxSizeInterface.(uint64)
			if inputLen > usortMaxSize {
				return fmt.Errorf("usort() can only be used with input_len<=%d.\n Got: input_len=%d", usortMaxSize, inputLen)
			}
			positionsDict := make(map[fp.Element][]uint64, inputLen)
			inputBasePtrCopy := *inputBasePtr
			for i := uint64(0); i < inputLen; i++ {
				val, err := vm.Memory.ReadFromAddressAsElement(&inputBasePtrCopy)
				if err != nil {
					return err
				}
				positionsDict[val] = append(positionsDict[val], uint64(i))
				inputBasePtrCopy, err = inputBasePtrCopy.AddOffset(1)
				if err != nil {
					return err
				}
			}

			outputArray := make([]fp.Element, len(positionsDict))
			iterator := 0
			for key := range positionsDict {
				outputArray[iterator] = key
				iterator++
			}
			sort.Sort(usortUtils.SortFelt(outputArray))

			outputLenAddr, err := output_len.GetAddress(vm)
			if err != nil {
				return err
			}
			outputLenMV := memory.MemoryValueFromFieldElement(new(fp.Element).SetUint64(uint64(len(outputArray))))
			err = vm.Memory.WriteToAddress(&outputLenAddr, &outputLenMV)
			if err != nil {
				return err
			}
			outputSegmentBaseAddr := vm.Memory.AllocateEmptySegment()
			outputAddr, err := output.GetAddress(vm)
			if err != nil {
				return err
			}
			outputSegmentBaseAddrMV := memory.MemoryValueFromMemoryAddress(&outputSegmentBaseAddr)
			err = vm.Memory.WriteToAddress(&outputAddr, &outputSegmentBaseAddrMV)
			for i, v := range outputArray {
				outputSegmentWriteArgsPtr, err := outputSegmentBaseAddr.AddOffset(int16(i))
				if err != nil {
					return err
				}
				outputElementMV := memory.MemoryValueFromFieldElement(&v)
				err = vm.Memory.WriteToAddress(&outputSegmentWriteArgsPtr, &outputElementMV)
				if err != nil {
					return err
				}
			}
			if err != nil {
				return err
			}
			multiplicitiesArray := make([]*fp.Element, len(outputArray))
			for i, v := range outputArray {
				multiplicitiesArray[i] = new(fp.Element).SetUint64(uint64(len(positionsDict[v])))
			}
			multiplicitesSegmentBaseAddr := vm.Memory.AllocateEmptySegment()
			multiplicitiesAddr, err := multiplicities.GetAddress(vm)
			if err != nil {
				return err
			}
			multiplicitesSegmentBaseAddrMV := memory.MemoryValueFromMemoryAddress(&multiplicitesSegmentBaseAddr)
			err = vm.Memory.WriteToAddress(&multiplicitiesAddr, &multiplicitesSegmentBaseAddrMV)
			if err != nil {
				return err
			}
			for i, v := range multiplicitiesArray {
				multiplicitesSegmentWriteArgsPtr, err := multiplicitesSegmentBaseAddr.AddOffset(int16(i))
				if err != nil {
					return err
				}
				multiplicitiesElementMV := memory.MemoryValueFromFieldElement(v)
				err = vm.Memory.WriteToAddress(&multiplicitesSegmentWriteArgsPtr, &multiplicitiesElementMV)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func newUsortEnterScopeHinter() hinter.Hinter {
	return &GenericZeroHinter{
		Name: "UsortEnterScope",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> vm_enter_scope(dict(__usort_max_size = globals().get('__usort_max_size')))
			usortMaxSize, err := ctx.ScopeManager.GetVariableValue("__usort_max_size")
			if err != nil {
				return err
			}

			ctx.ScopeManager.EnterScope(map[string]any{
				"__usort_max_size": usortMaxSize,
			})

			return nil
		},
	}
}

func createUsortEnterScopeHinter() (hinter.Hinter, error) {
	return newUsortEnterScopeHinter(), nil
}

func newUsortVerifyMultiplicityAssertHinter() hinter.Hinter {
	return &GenericZeroHinter{
		Name: "UsortVerifyMultiplicityAssert",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> assert len(positions) == 0
			positionsInterface, err := ctx.ScopeManager.GetVariableValue("positions")

			if err != nil {
				return err
			}

			positions, ok := positionsInterface.([]uint64)
			if !ok {
				return fmt.Errorf("casting positions into an array failed")
			}

			if len(positions) != 0 {
				return fmt.Errorf("assertion `len(positions) == 0` failed")
			}

			return nil
		},
	}
}

func createUsortVerifyMultiplicityAssertHinter() (hinter.Hinter, error) {
	return newUsortEnterScopeHinter(), nil
}

func newUsortVerifyHinter(value hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "UsortVerify",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> last_pos = 0
			//> positions = positions_dict[ids.value][::-1]

			positionsDictInterface, err := ctx.ScopeManager.GetVariableValue("positions_dict")

			if err != nil {
				return err
			}

			positionsDict, ok := positionsDictInterface.(map[fp.Element][]uint64)

			if !ok {
				return fmt.Errorf("casting positions_dict into an dictionary failed")
			}

			value, err := hinter.ResolveAsFelt(vm, value)

			if err != nil {
				return err
			}

			positions := positionsDict[*value]
			utils.Reverse(positions)

			return ctx.ScopeManager.AssignVariables(map[string]any{
				"last_pos":  0,
				"positions": positions,
			})
		},
	}
}

func createUsortVerifyHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	value, err := resolver.GetResOperander("value")

	if err != nil {
		return nil, err
	}

	return newUsortVerifyHinter(value), nil
}

func newUsortVerifyMultiplicityBodyHint(nextItemIndex hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "UsortVerifyMultiplicityBody",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> current_pos = positions.pop()
			//> ids.next_item_index = current_pos - last_pos
			//> last_pos = current_pos + 1

			positionsInterface, err := ctx.ScopeManager.GetVariableValue("positions")
			if err != nil {
				return err
			}

			positions, ok := positionsInterface.([]int64)
			if !ok {
				return fmt.Errorf("cannot cast positionsInterface to []int64")
			}

			newCurrentPos, err := utils.Pop(&positions)
			if err != nil {
				return err
			}

			currentPos, err := ctx.ScopeManager.GetVariableValue("current_pos")
			if err != nil {
				return err
			}

			currentPosInt, ok := currentPos.(int64)
			if !ok {
				return fmt.Errorf("cannot cast current_pos to int64")
			}

			lastPos, err := ctx.ScopeManager.GetVariableValue("last_pos")
			if err != nil {
				return err
			}

			lastPosInt, ok := lastPos.(int64)
			if !ok {
				return fmt.Errorf("cannot cast last_pos to int64")
			}

			// Calculate `next_item_index` memory value
			newNextItemIndexValue := currentPosInt - lastPosInt
			newNextItemIndexMemoryValue := memory.MemoryValueFromInt(newNextItemIndexValue)

			// Save `next_item_index` value in address
			addrNextItemIndex, err := nextItemIndex.GetAddress(vm)
			if err != nil {
				return err
			}

			err = vm.Memory.WriteToAddress(&addrNextItemIndex, &newNextItemIndexMemoryValue)
			if err != nil {
				return err
			}

			// Save `current_pos` and `last_pos` values in scope variables
			return ctx.ScopeManager.AssignVariables(map[string]any{
				"current_pos": newCurrentPos,
				"last_pos":    int64(currentPosInt + 1),
			})
		},
	}
}

func createUsortVerifyMultiplicityBodyHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	nextItemIndex, err := resolver.GetResOperander("next_item_index")
	if err != nil {
		return nil, err
	}

	return newUsortVerifyMultiplicityBodyHint(nextItemIndex), nil
}
