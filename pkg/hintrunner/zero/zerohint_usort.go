package zero

import (
	"fmt"
	"math/big"
	"sort"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/utils"
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

			//> 		input_ptr = ids.input
			//> 		input_len = int(ids.input_len)
			//> 		if __usort_max_size is not None:
			//> 			assert input_len <= __usort_max_size, (
			//> 				f"usort() can only be used with input_len<={__usort_max_size}. "
			//> 				f"Got: input_len={input_len}."
			//> 			)

			//> 		positions_dict = defaultdict(list)
			//> 		for i in range(input_len):
			//> 			val = memory[input_ptr + i]
			//> 			positions_dict[val].append(i)

			//> 		output = sorted(positions_dict.keys())
			//> 		ids.output_len = len(output)
			//> 		ids.output = segments.gen_arg(output)
			//> 		ids.multiplicities = segments.gen_arg([len(positions_dict[k]) for k in output])

			//> 		input_ptr = ids.input
			inputPtr, err := hinter.ResolveAsAddress(vm, input)
			if err != nil {
				return err
			}

			inputLen, err := hinter.ResolveAsUint64(vm, input_len)
			if err != nil {
				return err
			}

			inputLenBig := new(big.Int).SetUint64(inputLen)
			usortMaxSize, err := ctx.ScopeManager.GetVariableValueAsBigInt("__usort_max_size")
			if err == nil {
				if inputLenBig.Cmp(usortMaxSize) > 0 {
					return fmt.Errorf("usort() can only be used with input_len<=%d.\n Got: input_len=%d", usortMaxSize, inputLenBig)
				}
			}
			positionsDict := make(map[*fp.Element][]uint64, inputLen)

			var val fp.Element
			for i := int16(0); i < int16(inputLen); i++ {
				inputPtr, err := inputPtr.AddOffset(i)
				if err != nil {
					return err
				}
				val, err = vm.Memory.ReadFromAddressAsElement(&inputPtr)
				if err != nil {
					return err
				}

				positionsDict[&val] = append(positionsDict[&val], uint64(i))
			}

			outputArray := make([]*fp.Element, 0, len(positionsDict))
			for key := range positionsDict {
				outputArray = append(outputArray, key)
			}
			sort.Sort(utils.SortFelt(outputArray))

			outputLenAddr, err := hinter.ResolveAsAddress(vm, output_len)
			if err != nil {
				return err
			}
			mv := memory.MemoryValueFromFieldElement(new(fp.Element).SetUint64(uint64(len(outputArray))))
			err = vm.Memory.WriteToAddress(outputLenAddr, &mv)
			if err != nil {
				return err
			}
			outputSegmentBaseAddr, err := vm.Memory.AllocateSegment(outputArray)
			if err != nil {
				return err
			}
			outputAddr, err := hinter.ResolveAsAddress(vm, output)
			if err != nil {
				return err
			}
			valueMv := memory.MemoryValueFromMemoryAddress(&outputSegmentBaseAddr)
			err = vm.Memory.WriteToAddress(outputAddr, &valueMv)
			if err != nil {
				return err
			}
			multiplicitiesArray := make([]*fp.Element, len(outputArray))
			for _, v := range positionsDict {
				multiplicitiesArray = append(multiplicitiesArray, new(fp.Element).SetUint64(uint64(len(v))))
			}
			multiplicitesSegmentBaseAddr, err := vm.Memory.AllocateSegment(multiplicitiesArray)
			if err != nil {
				return err
			}
			multiplicitiesAddr, err := hinter.ResolveAsAddress(vm, multiplicities)
			if err != nil {
				return err
			}
			valueMv = memory.MemoryValueFromMemoryAddress(&multiplicitesSegmentBaseAddr)
			err = vm.Memory.WriteToAddress(multiplicitiesAddr, &valueMv)
			if err != nil {
				return err
			}
			return err
		},
	}
}
