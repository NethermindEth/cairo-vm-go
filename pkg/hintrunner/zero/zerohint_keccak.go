package zero

import (
	"fmt"
	"math"
	"math/big"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/builtins"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/holiman/uint256"
	"golang.org/x/crypto/sha3"
)

// CairoKeccakFinalize writes the result of F1600 Keccak permutation padded by __keccak_state_size_felts__ zeros to consecutive memory cells, __block_size__ times.
//
// `CairoKeccakFinalize` takes 1 operander as argument
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

// UnsafeKeccak computes keccak hash of the data in memory without validity enforcement and writes the result in the `low` and `high` memory cells
//
// `newUnsafeKeccakHint` takes 4 operanders as arguments
//   - `data` is the address in memory where the base of the data array to be hashed is stored. Each word in the array is 16 bytes long, except the last one, which could vary
//   - `length` is the length of the data to hash
//   - `low` is the low part of the produced hash
//   - `high` is the high part of the produced hash
func newUnsafeKeccakHint(data, length, high, low hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "UnsafeKeccak",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//>	from eth_hash.auto import keccak
			//>	data, length = ids.data, ids.length
			//>	if '__keccak_max_size' in globals():
			//>		assert length <= __keccak_max_size, \
			//>			f'unsafe_keccak() can only be used with length<={__keccak_max_size}. ' \
			//>			f'Got: length={length}.'
			//>	keccak_input = bytearray()
			//>	for word_i, byte_i in enumerate(range(0, length, 16)):
			//>		word = memory[data + word_i]
			//>		n_bytes = min(16, length - byte_i)
			//>		assert 0 <= word < 2 ** (8 * n_bytes)
			//>		keccak_input += word.to_bytes(n_bytes, 'big')
			//>	hashed = keccak(keccak_input)
			//>	ids.high = int.from_bytes(hashed[:16], 'big')
			//>	ids.low = int.from_bytes(hashed[16:32], 'big')

			//>	data, length = ids.data, ids.length
			lengthVal, err := hinter.ResolveAsUint64(vm, length)
			if err != nil {
				return err
			}

			//>	if '__keccak_max_size' in globals():
			//>		assert length <= __keccak_max_size, \
			//>			f'unsafe_keccak() can only be used with length<={__keccak_max_size}. ' \
			//>			f'Got: length={length}.'
			keccakMaxSize := uint64(1 << 20)
			if lengthVal > keccakMaxSize {
				return fmt.Errorf("unsafe_keccak() can only be used with length<=%d.\n Got: length=%d", keccakMaxSize, lengthVal)
			}

			dataPtr, err := hinter.ResolveAsAddress(vm, data)
			if err != nil {
				return err
			}

			//>	keccak_input = bytearray()
			keccakInput := make([]byte, 0)

			//>	for word_i, byte_i in enumerate(range(0, length, 16)):
			for i := uint64(0); i < lengthVal; i += 16 {
				wordFelt, err := vm.Memory.ReadAsElement(dataPtr.SegmentIndex, dataPtr.Offset)
				if err != nil {
					return err
				}
				//>		word = memory[data + word_i]
				word := uint256.Int(wordFelt.Bits())

				//>		n_bytes = min(16, length - byte_i)
				nBytes := utils.Min(lengthVal-i, 16)

				//>		assert 0 <= word < 2 ** (8 * n_bytes)
				if uint64(word.BitLen()) >= 8*nBytes {
					return fmt.Errorf("word %v is out range 0 <= word < 2 ** %d", &word, 8*nBytes)
				}

				//>		keccak_input += word.to_bytes(n_bytes, 'big')
				wordBytes := word.Bytes20()
				keccakInput = append(keccakInput, wordBytes[20-int(nBytes):]...)
				*dataPtr, err = dataPtr.AddOffset(1)
				if err != nil {
					return err
				}
			}
			hash := sha3.NewLegacyKeccak256()
			hash.Write(keccakInput)
			//>	hashed = keccak(keccak_input)
			hashedBytes := hash.Sum(nil)
			hashedHigh := new(fp.Element).SetBytes(hashedBytes[:16])
			hashedLow := new(fp.Element).SetBytes(hashedBytes[16:32])
			highAddr, err := high.GetAddress(vm)
			if err != nil {
				return err
			}
			hashedHighMV := memory.MemoryValueFromFieldElement(hashedHigh)
			//>	ids.high = int.from_bytes(hashed[:16], 'big')
			err = vm.Memory.WriteToAddress(&highAddr, &hashedHighMV)
			if err != nil {
				return err
			}
			lowAddr, err := low.GetAddress(vm)
			if err != nil {
				return err
			}
			hashedLowMV := memory.MemoryValueFromFieldElement(hashedLow)
			//>	ids.low = int.from_bytes(hashed[16:32], 'big')
			return vm.Memory.WriteToAddress(&lowAddr, &hashedLowMV)
		},
	}
}

func createUnsafeKeccakHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	data, err := resolver.GetResOperander("data")
	if err != nil {
		return nil, err
	}
	length, err := resolver.GetResOperander("length")
	if err != nil {
		return nil, err
	}
	high, err := resolver.GetResOperander("high")
	if err != nil {
		return nil, err
	}
	low, err := resolver.GetResOperander("low")
	if err != nil {
		return nil, err
	}
	return newUnsafeKeccakHint(data, length, high, low), nil
}

// UnsafeKeccakFinalize computes keccak hash of the data in memory without validity enforcement and writes the result in the `low` and `high` memory cells. It gets the data pointers from the `keccakState` memory cell, computing the hash of the data in the range [start_ptr, end_ptr).
//
// `newUnsafeKeccakFinalizeHint` takes 3 operanders as arguments
//   - `keccakState` is the address in memory where KeccakState struct containing 2 fields start_ptr and end_ptr is stored
//   - `low` is the address in memory where the low part of the produced hash should be written to
//   - `high` is the address in memory where the high part of the produced hash should be written to
//
// This hint utilises a struct called `KeccakState`:
//
//	struct KeccakState {
//	    start_ptr: felt*,
//	    end_ptr: felt*,
//	}
func newUnsafeKeccakFinalizeHint(keccakState, high, low hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "UnsafeKeccakFinalize",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> from eth_hash.auto import keccak
			//> keccak_input = bytearray()
			//> n_elms = ids.keccak_state.end_ptr - ids.keccak_state.start_ptr
			//> for word in memory.get_range(ids.keccak_state.start_ptr, n_elms):
			//>     keccak_input += word.to_bytes(16, 'big')
			//> hashed = keccak(keccak_input)
			//> ids.high = int.from_bytes(hashed[:16], 'big')
			//> ids.low = int.from_bytes(hashed[16:32], 'big')
			keccakStateAddr, err := keccakState.GetAddress(vm)
			if err != nil {
				return err
			}

			keccakStateMemoryValues, err := vm.Memory.GetConsecutiveMemoryValues(keccakStateAddr, 2)
			if err != nil {
				return err
			}

			startPtr, err := keccakStateMemoryValues[0].MemoryAddress()
			if err != nil {
				return err
			}

			endPtr, err := keccakStateMemoryValues[1].MemoryAddress()
			if err != nil {
				return err
			}

			//> n_elms = ids.keccak_state.end_ptr - ids.keccak_state.start_ptr
			nElems, err := endPtr.SubAddress(startPtr)
			if err != nil {
				return err
			}

			//> keccak_input = bytearray()
			keccakInput := make([]byte, 0)
			memoryValuesInRange, err := vm.Memory.GetConsecutiveMemoryValues(*startPtr, int16(nElems))
			if err != nil {
				return err
			}

			//> for word in memory.get_range(ids.keccak_state.start_ptr, n_elms):
			//>     keccak_input += word.to_bytes(16, 'big')
			for _, mv := range memoryValuesInRange {
				wordFelt, err := mv.FieldElement()
				if err != nil {
					return err
				}
				if wordFelt.Cmp(&utils.FeltMax128) > -1 {
					return fmt.Errorf("word %v is out range 0 <= word < 2 ** 128", wordFelt)
				}
				word := uint256.Int(wordFelt.Bits())
				wordBytes := word.Bytes20()
				keccakInput = append(keccakInput, wordBytes[4:]...)
			}

			//> hashed = keccak(keccak_input)
			hash := sha3.NewLegacyKeccak256()
			hash.Write(keccakInput)
			hashedBytes := hash.Sum(nil)
			hashedHigh := new(fp.Element).SetBytes(hashedBytes[:16])
			hashedLow := new(fp.Element).SetBytes(hashedBytes[16:32])
			highAddr, err := high.GetAddress(vm)
			if err != nil {
				return err
			}
			//> ids.high = int.from_bytes(hashed[:16], 'big')
			hashedHighMV := memory.MemoryValueFromFieldElement(hashedHigh)
			err = vm.Memory.WriteToAddress(&highAddr, &hashedHighMV)
			if err != nil {
				return err
			}
			lowAddr, err := low.GetAddress(vm)
			if err != nil {
				return err
			}

			//> ids.low = int.from_bytes(hashed[16:32], 'big')
			hashedLowMV := memory.MemoryValueFromFieldElement(hashedLow)
			return vm.Memory.WriteToAddress(&lowAddr, &hashedLowMV)
		},
	}
}

func createUnsafeKeccakFinalizeHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	keccak_state, err := resolver.GetResOperander("keccak_state")
	if err != nil {
		return nil, err
	}
	high, err := resolver.GetResOperander("high")
	if err != nil {
		return nil, err
	}
	low, err := resolver.GetResOperander("low")
	if err != nil {
		return nil, err
	}
	return newUnsafeKeccakFinalizeHint(keccak_state, high, low), nil
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
			mvLowLow := memory.MemoryValueFromFieldElement(&lowResultFeltLow)

			lowResultUint256High := lowUint256
			lowResultUint256High.Rsh(&lowResultUint256High, 64)
			lowResultUint256High.And(&lowResultUint256High, &maxUint64)
			lowResulBytes32High := lowResultUint256High.Bytes32()
			lowResultFeltHigh, _ := fp.BigEndian.Element(&lowResulBytes32High)
			mvLowHigh := memory.MemoryValueFromFieldElement(&lowResultFeltHigh)

			highResultUint256Low := highUint256
			highResultUint256Low.And(&maxUint64, &highResultUint256Low)
			highResulBytes32Low := highResultUint256Low.Bytes32()
			highResultFeltLow, _ := fp.BigEndian.Element(&highResulBytes32Low)
			mvHighLow := memory.MemoryValueFromFieldElement(&highResultFeltLow)

			highResultUint256High := highUint256
			highResultUint256High.Rsh(&highResultUint256High, 64)
			highResultUint256High.And(&maxUint64, &highResultUint256High)
			highResulBytes32High := highResultUint256High.Bytes32()
			highResultFeltHigh, _ := fp.BigEndian.Element(&highResulBytes32High)
			mvHighHigh := memory.MemoryValueFromFieldElement(&highResultFeltHigh)

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

// CompareKeccakFullRateInBytes hint compares a value to KECCAK_FULL_RATE_IN_BYTES constant, i.e., 136
//
// `newKeccakWriteArgsHint` takes 1 operander as argument
//   - `nBytes` is the value to be compared with KECCAK_FULL_RATE_IN_BYTES
//
// `newKeccakWriteArgsHint` writes 1 or 0 to `ap` memory address depending on whether
// `n_bytes` is greater or equal to KECCAK_FULL_RATE_IN_BYTES or not
func newCompareKeccakFullRateInBytesHint(nBytes hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "CompareKeccakFullRateInBytes",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> python hint: ids.n_bytes >= ids.KECCAK_FULL_RATE_IN_BYTES
			//> JSON file hint: memory[ap] = to_felt_or_relocatable(ids.n_bytes >= ids.KECCAK_FULL_RATE_IN_BYTES)

			// n_bytes should fit into a uint64
			// we cannot 100% exclude the possibility that it doesn't
			nBytesVal, err := hinter.ResolveAsUint64(vm, nBytes)
			if err != nil {
				return err
			}

			apAddr := vm.Context.AddressAp()
			var resultMv memory.MemoryValue
			if nBytesVal >= uint64(utils.KECCAK_FULL_RATE_IN_BYTES) {
				resultMv = memory.MemoryValueFromFieldElement(&utils.FeltOne)
			} else {
				resultMv = memory.MemoryValueFromFieldElement(&utils.FeltZero)
			}

			return vm.Memory.WriteToAddress(&apAddr, &resultMv)
		},
	}
}

func createCompareKeccakFullRateInBytesNondetHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	nBytes, err := resolver.GetResOperander("n_bytes")
	if err != nil {
		return nil, err
	}

	return newCompareKeccakFullRateInBytesHint(nBytes), nil
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

// CompareBytesInWord hint compares a value to BYTES_IN_WORD constant, i.e., 8
//
// `newCompareBytesInWordHint` takes 1 operander as argument
//   - `nBytes` is the value to be compared with BYTES_IN_WORD
//
// `newCompareBytesInWordHint` writes 1 or 0 to `ap` memory address depending on whether
// `n_bytes` is lower than BYTES_IN_WORD or not
func newCompareBytesInWordHint(nBytes hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "CompareBytesInWord",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> python hint: ids.n_bytes < ids.BYTES_IN_WORD
			//> JSON file hint: memory[ap] = to_felt_or_relocatable(ids.n_bytes < ids.BYTES_IN_WORD)

			// n_bytes should fit into a uint64
			// we cannot 100% exclude the possibility that it doesn't
			nBytesVal, err := hinter.ResolveAsUint64(vm, nBytes)
			if err != nil {
				return err
			}

			bytesInWord := uint64(8)
			apAddr := vm.Context.AddressAp()
			var resultMv memory.MemoryValue
			if nBytesVal < bytesInWord {
				resultMv = memory.MemoryValueFromFieldElement(&utils.FeltOne)
			} else {
				resultMv = memory.MemoryValueFromFieldElement(&utils.FeltZero)
			}

			return vm.Memory.WriteToAddress(&apAddr, &resultMv)
		},
	}
}

func createCompareBytesInWordNondetHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	nBytes, err := resolver.GetResOperander("n_bytes")
	if err != nil {
		return nil, err
	}

	return newCompareBytesInWordHint(nBytes), nil
}

// SplitInput3 hint writes at address `ids.high3` and `ids.low3` in memory
// the quotient and remainder of the division of the value at memory address
// `ids.inputs + 3` by 256
//
// `newSplitInput3Hint` takes 3 operanders as arguments
//   - `high3` is the address in memory where to store the quotient of the division
//   - `low3` is the address in memory where to store the remainder of the division
//   - `inputs` is the address in memory to which we add an offset of 3 and read that value
func newSplitInput3Hint(high3, low3, inputs hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "SplitInput3",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> ids.high3, ids.low3 = divmod(memory[ids.inputs + 3], 256)

			high3Addr, err := hinter.ResolveAsAddress(vm, high3)
			if err != nil {
				return err
			}

			low3Addr, err := hinter.ResolveAsAddress(vm, low3)
			if err != nil {
				return err
			}

			inputsAddr, err := hinter.ResolveAsAddress(vm, inputs)
			if err != nil {
				return err
			}

			*inputsAddr, err = inputsAddr.AddOffset(3)
			if err != nil {
				return err
			}

			inputValue, err := vm.Memory.ReadFromAddress(inputsAddr)
			if err != nil {
				return err
			}

			var inputBigInt big.Int
			inputValue.Felt.BigInt(&inputBigInt)

			divisor := big.NewInt(256)

			high3BigInt := new(big.Int)
			low3BigInt := new(big.Int)

			high3BigInt.DivMod(&inputBigInt, divisor, low3BigInt)

			var high3Felt fp.Element
			high3Felt.SetBigInt(high3BigInt)
			high3Mv := memory.MemoryValueFromFieldElement(&high3Felt)

			var low3Felt fp.Element
			high3Felt.SetBigInt(low3BigInt)
			low3Mv := memory.MemoryValueFromFieldElement(&low3Felt)

			err = vm.Memory.WriteToAddress(low3Addr, &high3Mv)
			if err != nil {
				return err
			}

			return vm.Memory.WriteToAddress(high3Addr, &low3Mv)
		},
	}
}

func createSplitInput3Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	high3, err := resolver.GetResOperander("high3")
	if err != nil {
		return nil, err
	}

	low3, err := resolver.GetResOperander("low3")
	if err != nil {
		return nil, err
	}

	inputs, err := resolver.GetResOperander("inputs")
	if err != nil {
		return nil, err
	}

	return newSplitInput3Hint(high3, low3, inputs), nil
}
