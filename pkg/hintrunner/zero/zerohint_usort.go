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

// UsortEnterScope hint enters a new scope with `__usort_max_size` value
//
// `newUsortEnterScopeHint` doesn't take any operander as argument
//
// `newUsortEnterScopeHint` gets `__usort_max_size` value from the current
// scope and enters a new scope with this same value
func newUsortEnterScopeHint() hinter.Hinter {
	return &GenericZeroHinter{
		Name: "UsortEnterScope",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> vm_enter_scope(dict(__usort_max_size = globals().get('__usort_max_size')))
			usortMaxSize := uint64(1 << 20)

			ctx.ScopeManager.EnterScope(map[string]any{
				"__usort_max_size": usortMaxSize,
			})

			return nil
		},
	}
}

func createUsortEnterScopeHinter() (hinter.Hinter, error) {
	return newUsortEnterScopeHint(), nil
}

// UsortBody hint sorts the input array of field elements. The sorting results in generation of output array without duplicates and multiplicities array, where each element represents the number of times the corresponding element in the output array appears in the input array. The output and multiplicities arrays are written to the new, separate segments in memory.
//
// `newSplit64Hint` takes 5 operanders as arguments
//   - `input` is the pointer to the base of input array of field elements
//   - `inputLen` is the length of the input array
//   - `output` is the pointer to the base of the output array of field elements
//   - `outputLen` is the length of the output array
//   - `multiplicities` is the pointer to the base of the multiplicities array of field elements
func newUsortBodyHint(input, inputLen, output, outputLen, multiplicities hinter.Reference) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "UsortBody",
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

			inputLenValue, err := hinter.ResolveAsUint64(vm, inputLen)
			if err != nil {
				return err
			}

			usortMaxSize, err := hinter.GetVariableAs[uint64](&ctx.ScopeManager, "__usort_max_size")
			if err != nil {
				return err
			}

			if inputLenValue > usortMaxSize {
				return fmt.Errorf("usort() can only be used with input_len<=%d.\n Got: input_len=%d", usortMaxSize, inputLenValue)
			}

			positionsDict := make(map[fp.Element][]uint64, inputLenValue)
			for i := uint64(0); i < inputLenValue; i++ {
				val, err := vm.Memory.ReadFromAddressAsElement(inputBasePtr)
				if err != nil {
					return err
				}

				positionsDict[val] = append(positionsDict[val], i)
				*inputBasePtr, err = inputBasePtr.AddOffset(1)
				if err != nil {
					return err
				}
			}

			err = ctx.ScopeManager.AssignVariable("positions_dict", positionsDict)
			if err != nil {
				return err
			}

			outputArray := make([]fp.Element, len(positionsDict))
			iterator := 0
			for key := range positionsDict {
				outputArray[iterator] = key
				iterator++
			}

			sort.Sort(usortUtils.SortFelt(outputArray))

			outputLenAddr, err := outputLen.Get(vm)
			if err != nil {
				return err
			}

			outputLenMV := memory.MemoryValueFromFieldElement(new(fp.Element).SetUint64(uint64(len(outputArray))))
			err = vm.Memory.WriteToAddress(&outputLenAddr, &outputLenMV)
			if err != nil {
				return err
			}

			outputSegmentBaseAddr := vm.Memory.AllocateEmptySegment()
			outputAddr, err := output.Get(vm)
			if err != nil {
				return err
			}

			outputSegmentBaseAddrMV := memory.MemoryValueFromMemoryAddress(&outputSegmentBaseAddr)
			err = vm.Memory.WriteToAddress(&outputAddr, &outputSegmentBaseAddrMV)
			if err != nil {
				return err
			}

			for _, v := range outputArray {
				outputElementMV := memory.MemoryValueFromFieldElement(&v)
				err = vm.Memory.WriteToAddress(&outputSegmentBaseAddr, &outputElementMV)
				if err != nil {
					return err
				}

				outputSegmentBaseAddr, err = outputSegmentBaseAddr.AddOffset(1)
				if err != nil {
					return err
				}
			}

			multiplicitiesArray := make([]*fp.Element, len(outputArray))
			for i, v := range outputArray {
				multiplicitiesArray[i] = new(fp.Element).SetUint64(uint64(len(positionsDict[v])))
			}

			multiplicitesSegmentBaseAddr := vm.Memory.AllocateEmptySegment()
			multiplicitiesAddr, err := multiplicities.Get(vm)
			if err != nil {
				return err
			}

			multiplicitesSegmentBaseAddrMV := memory.MemoryValueFromMemoryAddress(&multiplicitesSegmentBaseAddr)
			err = vm.Memory.WriteToAddress(&multiplicitiesAddr, &multiplicitesSegmentBaseAddrMV)
			if err != nil {
				return err
			}

			for _, v := range multiplicitiesArray {
				multiplicitiesElementMV := memory.MemoryValueFromFieldElement(v)
				err = vm.Memory.WriteToAddress(&multiplicitesSegmentBaseAddr, &multiplicitiesElementMV)
				if err != nil {
					return err
				}

				multiplicitesSegmentBaseAddr, err = multiplicitesSegmentBaseAddr.AddOffset(1)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func createUsortBodyHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	input, err := resolver.GetReference("input")
	if err != nil {
		return nil, err
	}

	input_len, err := resolver.GetReference("input_len")
	if err != nil {
		return nil, err
	}

	output, err := resolver.GetReference("output")
	if err != nil {
		return nil, err
	}

	output_len, err := resolver.GetReference("output_len")
	if err != nil {
		return nil, err
	}

	multiplicities, err := resolver.GetReference("multiplicities")
	if err != nil {
		return nil, err
	}

	return newUsortBodyHint(input, input_len, output, output_len, multiplicities), nil
}

// UsortVerify hint prepares for verifying the presence of duplicates of
// a specific value in the sorted output (array of fields)
//
// `newUsortVerifyHint` takes one operander as argument
//   - `value` is the value at the given position in the output
//
// `last_pos` is set to zero
// `positions` is set to the reversed order list associated with `ids.value`
// key in `positions_dict`
// `newUsortVerifyHint` assigns `last_pos` and `positions` in the current scope
func newUsortVerifyHint(value hinter.Reference) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "UsortVerify",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> last_pos = 0
			//> positions = positions_dict[ids.value][::-1]

			positionsDict, err := hinter.GetVariableAs[map[fp.Element][]uint64](&ctx.ScopeManager, "positions_dict")
			if err != nil {
				return err
			}

			value, err := hinter.ResolveAsFelt(vm, value)
			if err != nil {
				return err
			}

			positionsToCopy := positionsDict[*value]

			positions := make([]uint64, len(positionsToCopy))
			copy(positions, positionsToCopy)

			utils.Reverse(positions)

			return ctx.ScopeManager.AssignVariables(map[string]any{
				"last_pos":  uint64(0),
				"positions": positions,
			})
		},
	}
}

func createUsortVerifyHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	value, err := resolver.GetReference("value")

	if err != nil {
		return nil, err
	}

	return newUsortVerifyHint(value), nil
}

// UsortVerifyMultiplicityBody hint extracts a specific value
// of the sorted output with `pop`, updating indices for the verification
// of the next value
//
// `newUsortVerifyMultiplicityBodyHint` takes one operander as argument
//   - `nextItemIndex` is the index of the next item
//
// `next_item_index` is set to `current_pos - last_pos` for the next iteration
// `newUsortVerifyMultiplicityBodyHint` assigns `last_pos` in the current scope
func newUsortVerifyMultiplicityBodyHint(nextItemIndex hinter.Reference) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "UsortVerifyMultiplicityBody",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> current_pos = positions.pop()
			//> ids.next_item_index = current_pos - last_pos
			//> last_pos = current_pos + 1

			positions, err := hinter.GetVariableAs[[]uint64](&ctx.ScopeManager, "positions")
			if err != nil {
				return err
			}

			currentPos, err := utils.Pop(&positions)
			if err != nil {
				return err
			}

			err = ctx.ScopeManager.AssignVariable("positions", positions)
			if err != nil {
				return err
			}

			lastPos, err := hinter.GetVariableAs[uint64](&ctx.ScopeManager, "last_pos")
			if err != nil {
				return err
			}

			// Calculate `next_item_index` memory value
			newNextItemIndexValue := currentPos - lastPos
			newNextItemIndexMemoryValue := memory.MemoryValueFromUint(newNextItemIndexValue)

			// Save `next_item_index` value in address
			addrNextItemIndex, err := nextItemIndex.Get(vm)
			if err != nil {
				return err
			}

			err = ctx.ScopeManager.AssignVariable("last_pos", currentPos+1)
			if err != nil {
				return err
			}

			return vm.Memory.WriteToAddress(&addrNextItemIndex, &newNextItemIndexMemoryValue)
		},
	}
}

func createUsortVerifyMultiplicityBodyHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	nextItemIndex, err := resolver.GetReference("next_item_index")
	if err != nil {
		return nil, err
	}

	return newUsortVerifyMultiplicityBodyHint(nextItemIndex), nil
}

// UsortVerifyMultiplicityAssert hint checks that the `positions` variable in scope
// doesn't contain any value
//
// `newUsortVerifyMultiplicityAssertHint` doesn't take any operander as argument
//
// This hint is used when sorting an array of field elements while removing duplicates
// in `usort` Cairo function
func newUsortVerifyMultiplicityAssertHint() hinter.Hinter {
	return &GenericZeroHinter{
		Name: "UsortVerifyMultiplicityAssert",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> assert len(positions) == 0

			positions, err := hinter.GetVariableAs[[]uint64](&ctx.ScopeManager, "positions")
			if err != nil {
				return err
			}

			if len(positions) != 0 {
				return fmt.Errorf("assertion `len(positions) == 0` failed")
			}

			return nil
		},
	}
}

func createUsortVerifyMultiplicityAssertHinter() (hinter.Hinter, error) {
	return newUsortVerifyMultiplicityAssertHint(), nil
}
