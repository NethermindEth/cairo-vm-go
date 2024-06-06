package zero

import (
	"math"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/builtins"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/holiman/uint256"
)

// CairoKeccakFinalize writes the result of F1600 Keccak permutation padded by __keccak_state_size_felts__ zeros to consecutive memory cells, __block_size__ times.
//
// `CairoKeccakFinalize` takes 3 operanders as arguments
//   - `keccakPtrEnd` is the address in memory where to start writing the result
func newCairoKeccakFinalizeHint(keccakPtrEnd hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "CairoKeccakFinalize",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> _keccak_state_size_felts = int(ids.KECCAK_STATE_SIZE_FELTS)
			//> _block_size = int(ids.BLOCK_SIZE)
			//> assert 0 <= _keccak_state_size_felts < 100
			//> assert 0 <= _block_size < 10
			//> inp = [0] * _keccak_state_size_felts
			//> padding = (inp + keccak_func(inp)) * _block_size
			//> segments.write_arg(ids.keccak_ptr_end, padding)

			keccakStateSizeFeltsVal := uint64(25)
			blockSizeVal := uint64(3)

			var input [25]uint64
			builtins.KeccakF1600(&input)
			padding := make([]uint64, keccakStateSizeFeltsVal)
			padding = append(padding, input[:]...)
			result := make([]uint64, 0, keccakStateSizeFeltsVal*blockSizeVal)
			for i := uint64(0); i < blockSizeVal; i++ {
				result = append(result, padding...)
			}
			keccakPtrEnd, err := hinter.ResolveAsAddress(vm, keccakPtrEnd)
			if err != nil {
				return err
			}
			for i := 0; i < len(result); i++ {
				resultMV := memory.MemoryValueFromUint(result[i])
				err = vm.Memory.WriteToAddress(keccakPtrEnd, &resultMV)
				if err != nil {
					return err
				}
				keccakPtrEndIncremented, err := keccakPtrEnd.AddOffset(1)
				if err != nil {
					return err
				}
				keccakPtrEnd = &keccakPtrEndIncremented
			}
			return nil
		},
	}
}

func createCairoKeccakFinalizeHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	keccakPtrEnd, err := resolver.GetResOperander("keccak_ptr_end")
	if err != nil {
		return nil, err
	}
	return newCairoKeccakFinalizeHint(keccakPtrEnd), nil
}

// KeccakWriteArgs hint writes Keccak function arguments in memory
//
// `newKeccakWriteArgsHint` takes 3 operanders as arguments
//   - `inputs` is the address in memory where to write Keccak arguments
//   - `low` is the low part of the `uint256` argument for the Keccac function
//   - `high` is the high part of the `uint256` argument for the Keccac function
func newKeccakWriteArgsHint(inputs, low, high hinter.ResOperander) hinter.Hinter {
	name := "KeccakWriteArgs"
	return &GenericZeroHinter{
		Name: name,
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> segments.write_arg(ids.inputs, [ids.low % 2 ** 64, ids.low // 2 ** 64])
			//> segments.write_arg(ids.inputs + 2, [ids.high % 2 ** 64, ids.high // 2 ** 64])

			low, err := hinter.ResolveAsFelt(vm, low)
			if err != nil {
				return err
			}

			high, err := hinter.ResolveAsFelt(vm, high)
			if err != nil {
				return err
			}

			inputsPtr, err := hinter.ResolveAsAddress(vm, inputs)
			if err != nil {
				return err
			}

			var lowUint256 uint256.Int = uint256.Int(low.Bits())
			var highUint256 uint256.Int = uint256.Int(high.Bits())

			var maxUint64 uint256.Int = *uint256.NewInt(math.MaxUint64)

			lowResultUint256Low := lowUint256
			lowResultUint256Low.And(&maxUint64, &lowResultUint256Low)
			lowResulBytes32Low := lowResultUint256Low.Bytes32()
			lowResultFeltLow, _ := fp.BigEndian.Element(&lowResulBytes32Low)
			mvLowLow := mem.MemoryValueFromFieldElement(&lowResultFeltLow)

			lowResultUint256High := lowUint256
			lowResultUint256High.Rsh(&lowResultUint256High, 64)
			lowResultUint256High.And(&lowResultUint256High, &maxUint64)
			lowResulBytes32High := lowResultUint256High.Bytes32()
			lowResultFeltHigh, _ := fp.BigEndian.Element(&lowResulBytes32High)
			mvLowHigh := mem.MemoryValueFromFieldElement(&lowResultFeltHigh)

			highResultUint256Low := highUint256
			highResultUint256Low.And(&maxUint64, &highResultUint256Low)
			highResulBytes32Low := highResultUint256Low.Bytes32()
			highResultFeltLow, _ := fp.BigEndian.Element(&highResulBytes32Low)
			mvHighLow := mem.MemoryValueFromFieldElement(&highResultFeltLow)

			highResultUint256High := highUint256
			highResultUint256High.Rsh(&highResultUint256High, 64)
			highResultUint256High.And(&maxUint64, &highResultUint256High)
			highResulBytes32High := highResultUint256High.Bytes32()
			highResultFeltHigh, _ := fp.BigEndian.Element(&highResulBytes32High)
			mvHighHigh := mem.MemoryValueFromFieldElement(&highResultFeltHigh)

			err = vm.Memory.Write(inputsPtr.SegmentIndex, inputsPtr.Offset, &mvLowLow)
			if err != nil {
				return err
			}

			err = vm.Memory.Write(inputsPtr.SegmentIndex, inputsPtr.Offset+1, &mvLowHigh)
			if err != nil {
				return err
			}

			err = vm.Memory.Write(inputsPtr.SegmentIndex, inputsPtr.Offset+2, &mvHighLow)
			if err != nil {
				return err
			}

			return vm.Memory.Write(inputsPtr.SegmentIndex, inputsPtr.Offset+3, &mvHighHigh)
		},
	}
}

func createKeccakWriteArgsHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	inputs, err := resolver.GetResOperander("inputs")
	if inputs != nil {
		return nil, err
	}

	low, err := resolver.GetResOperander("low")
	if low != nil {
		return nil, err
	}

	high, err := resolver.GetResOperander("high")
	if high != nil {
		return nil, err
	}

	return newKeccakWriteArgsHint(inputs, low, high), nil
}

// BlockPermutation hint executes the Keccak block permutation function to a segment of memory
//
// `newBlockPermutationHint` takes 1 operander as argument
//   - `keccakPtr` is a pointer to the address in memory where to write the result of the permutation
//
// `KECCAK_STATE_SIZE_FELTS` is an operander in the Python VM but it is constant that we decided to hardcode
// `newBlockPermutationHint` reads 25 memory cells starting from `keccakPtr -  25`, and writes
// the result of the Keccak block permutation in the next 25 memory cells, starting from `keccakPtr`
func newBlockPermutationHint(keccakPtr hinter.ResOperander) hinter.Hinter {
	name := "BlockPermutation"
	return &GenericZeroHinter{
		Name: name,
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.keccak_utils.keccak_utils import keccak_func
			//> _keccak_state_size_felts = int(ids.KECCAK_STATE_SIZE_FELTS)
			//> assert 0 <= _keccak_state_size_felts < 100
			//
			//> output_values = keccak_func(memory.get_range(
			//> 	ids.keccak_ptr - _keccak_state_size_felts, _keccak_state_size_felts))
			//> segments.write_arg(ids.keccak_ptr, output_values)

			keccakWritePtr, err := hinter.ResolveAsAddress(vm, keccakPtr)
			if err != nil {
				return err
			}

			keccakStateSize := uint64(25)

			var readAddr = *keccakWritePtr
			var offset int16 = int16(keccakStateSize)
			var negOffset int16 = -offset

			readAddr, err = readAddr.AddOffset(negOffset)
			if err != nil {
				return err
			}

			inputValuesInRange, err := vm.Memory.GetConsecutiveMemoryValues(readAddr, offset)
			if err != nil {
				return err
			}

			var keccakInput [25]uint64

			for i, valueMemoryValue := range inputValuesInRange {
				valueUint64, err := valueMemoryValue.Uint64()
				if err != nil {
					return err
				}

				keccakInput[i] = valueUint64
			}

			builtins.KeccakF1600(&keccakInput)

			for i := 0; i < 25; i++ {
				inputValue := memory.MemoryValueFromUint(keccakInput[i])
				memoryOffset := uint64(i)

				err = vm.Memory.Write(keccakWritePtr.SegmentIndex, keccakWritePtr.Offset+memoryOffset, &inputValue)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func createBlockPermutationHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	keccakPtr, err := resolver.GetResOperander("keccak_ptr")
	if keccakPtr != nil {
		return nil, err
	}

	return newBlockPermutationHint(keccakPtr), nil
}
