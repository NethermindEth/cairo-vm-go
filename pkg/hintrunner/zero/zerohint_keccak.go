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
//
// There are 2 versions of this hint, depending on whether `_block_size` should be lower than 10 or 1000
// Corresponding hintcodes are cairoKeccakFinalizeCode and cairoKeccakFinalizeBlockSize1000Code
func newCairoKeccakFinalizeHint(keccakPtrEnd hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "CairoKeccakFinalize",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> _keccak_state_size_felts = int(ids.KECCAK_STATE_SIZE_FELTS)
			//> _block_size = int(ids.BLOCK_SIZE)
			//> assert 0 <= _keccak_state_size_felts < 100
			//> assert 0 <= _block_size < 10 (cairoKeccakFinalize)  //> assert 0 <= _block_size < 1000 (cairoKeccakFinalizeBlockSize1000)
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
	return &GenericZeroHinter{
		Name: "KeccakWriteArgs",
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
	if err != nil {
		return nil, err
	}

	low, err := resolver.GetResOperander("low")
	if err != nil {
		return nil, err
	}

	high, err := resolver.GetResOperander("high")
	if err != nil {
		return nil, err
	}

	return newKeccakWriteArgsHint(inputs, low, high), nil
}

// CompareKeccakFullRateInBytes hint compares a value to KECCAK_FULL_RATE_IN_BYTES constant, i.e., 136
//
// `newCompareKeccakFullRateInBytesHint` takes 1 operander as argument
//   - `nBytes` is the value to be compared with KECCAK_FULL_RATE_IN_BYTES
//
// `newCompareKeccakFullRateInBytesHint` writes 1 or 0 to `ap` memory address depending on whether
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
	return &GenericZeroHinter{
		Name: "BlockPermutation",
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
	if err != nil {
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
		Name: "CompareBytesInWordHint",
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

// SplitInput12 hint assigns to `ids.high12` and `ids.low12` variables
// the quotient and remainder of the division of the value at memory address
// `ids.inputs + 12` by 256 ** 4
//
// `newSplitInput12Hint` takes 3 operanders as arguments
//   - `high12` is the variable that will store the quotient of the division
//   - `low12` is the variable that will store the remainder of the division
//   - `inputs` is the address in memory to which we add an offset of 12 and read that value
func newSplitInput12Hint(high12, low12, inputs hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "SplitInput12",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> ids.high12, ids.low12 = divmod(memory[ids.inputs + 12], 256 ** 4)

			high12Addr, err := high12.GetAddress(vm)
			if err != nil {
				return err
			}

			low12Addr, err := low12.GetAddress(vm)
			if err != nil {
				return err
			}

			inputsAddr, err := hinter.ResolveAsAddress(vm, inputs)
			if err != nil {
				return err
			}

			*inputsAddr, err = inputsAddr.AddOffset(12)
			if err != nil {
				return err
			}

			inputValue, err := vm.Memory.ReadFromAddress(inputsAddr)
			if err != nil {
				return err
			}

			var inputBigInt big.Int
			inputValue.Felt.BigInt(&inputBigInt)

			// 256 ** 4
			divisor := big.NewInt(4294967296)

			high12BigInt := new(big.Int)
			low12BigInt := new(big.Int)

			high12BigInt.DivMod(&inputBigInt, divisor, low12BigInt)

			var high12Felt fp.Element
			high12Felt.SetBigInt(high12BigInt)
			high12Mv := memory.MemoryValueFromFieldElement(&high12Felt)

			var low12Felt fp.Element
			low12Felt.SetBigInt(low12BigInt)
			low12Mv := memory.MemoryValueFromFieldElement(&low12Felt)

			err = vm.Memory.WriteToAddress(&low12Addr, &low12Mv)
			if err != nil {
				return err
			}

			return vm.Memory.WriteToAddress(&high12Addr, &high12Mv)
		},
	}
}

func createSplitInput12Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	high12, err := resolver.GetResOperander("high12")
	if err != nil {
		return nil, err
	}

	low12, err := resolver.GetResOperander("low12")
	if err != nil {
		return nil, err
	}

	inputs, err := resolver.GetResOperander("inputs")
	if err != nil {
		return nil, err
	}

	return newSplitInput12Hint(high12, low12, inputs), nil
}

// SplitInput15 hint assigns to `ids.high15` and `ids.low15` variables
// the quotient and remainder of the division of the value at memory address
// `ids.inputs + 15` by 256 ** 5
//
// `newSplitInput9Hint` takes 3 operanders as arguments
//   - `high15` is the variable that will store the quotient of the division
//   - `low15` is the variable that will store the remainder of the division
//   - `inputs` is the address in memory to which we add an offset of 15 and read that value
func newSplitInput15Hint(high15, low15, inputs hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "SplitInput15",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> ids.high15, ids.low15 = divmod(memory[ids.inputs + 15], 256 ** 5)

			high15Addr, err := high15.GetAddress(vm)
			if err != nil {
				return err
			}

			low15Addr, err := low15.GetAddress(vm)
			if err != nil {
				return err
			}

			inputsAddr, err := hinter.ResolveAsAddress(vm, inputs)
			if err != nil {
				return err
			}

			*inputsAddr, err = inputsAddr.AddOffset(15)
			if err != nil {
				return err
			}

			inputValue, err := vm.Memory.ReadFromAddress(inputsAddr)
			if err != nil {
				return err
			}

			var inputBigInt big.Int
			inputValue.Felt.BigInt(&inputBigInt)

			// 256 ** 5
			divisor := big.NewInt(1099511627776)

			high15BigInt := new(big.Int)
			low15BigInt := new(big.Int)

			high15BigInt.DivMod(&inputBigInt, divisor, low15BigInt)

			var high15Felt fp.Element
			high15Felt.SetBigInt(high15BigInt)
			high15Mv := memory.MemoryValueFromFieldElement(&high15Felt)

			var low15Felt fp.Element
			low15Felt.SetBigInt(low15BigInt)
			low15Mv := memory.MemoryValueFromFieldElement(&low15Felt)

			err = vm.Memory.WriteToAddress(&low15Addr, &low15Mv)
			if err != nil {
				return err
			}

			return vm.Memory.WriteToAddress(&high15Addr, &high15Mv)
		},
	}
}

func createSplitInput15Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	high15, err := resolver.GetResOperander("high15")
	if err != nil {
		return nil, err
	}

	low15, err := resolver.GetResOperander("low15")
	if err != nil {
		return nil, err
	}

	inputs, err := resolver.GetResOperander("inputs")
	if err != nil {
		return nil, err
	}

	return newSplitInput15Hint(high15, low15, inputs), nil
}

// SplitOutputMidLowHigh hint assigns to `ids.output1_low` the remainder of the division
// of `ids.output1` variable by 256 ** 7 and uses its quotient as a variable which is
// divided by 2 ** 128, the quotient and remainder of which are then assigned to `ids.output1_high`
// and `ids.output1_mid` respectively.
//
// `newSplitOutputMidLowHighHint` takes 4 operanders as arguments
//   - `output1Low` is the variable that will store the remainder of the first division
//   - `output1Mid` is the variable that will store the remainder of the second division
//   - `output1High` is the variable that will store the quotient of the second division
//   - `output1` is the variable that will be divided in the first division
func newSplitOutputMidLowHighHint(output1, output1Low, output1Mid, output1High hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "SplitOutputMidLowHigh",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> tmp, ids.output1_low = divmod(ids.output1, 256 ** 7)
			//> ids.output1_high, ids.output1_mid = divmod(tmp, 2 ** 128)

			output1LowAddr, err := output1Low.GetAddress(vm)
			if err != nil {
				return err
			}

			output1MidAddr, err := output1Mid.GetAddress(vm)
			if err != nil {
				return err
			}

			output1HighAddr, err := output1High.GetAddress(vm)
			if err != nil {
				return err
			}

			output1Felt, err := hinter.ResolveAsFelt(vm, output1)
			if err != nil {
				return err
			}

			output1BigInt := new(big.Int)
			output1Felt.BigInt(output1BigInt)

			tmpBigInt := new(big.Int)
			output1LowBigInt := new(big.Int)
			output1MidBigInt := new(big.Int)
			output1HighBigInt := new(big.Int)

			divisorOne := new(big.Int).Exp(big.NewInt(256), big.NewInt(7), nil)
			divisorTwo := new(big.Int).Exp(big.NewInt(2), big.NewInt(128), nil)

			tmpBigInt.DivMod(output1BigInt, divisorOne, output1LowBigInt)
			output1HighBigInt.DivMod(tmpBigInt, divisorTwo, output1MidBigInt)

			var output1LowFelt fp.Element
			output1LowFelt.SetBigInt(output1LowBigInt)
			output1LowMv := memory.MemoryValueFromFieldElement(&output1LowFelt)

			var output1MidFelt fp.Element
			output1MidFelt.SetBigInt(output1MidBigInt)
			output1MidMv := memory.MemoryValueFromFieldElement(&output1MidFelt)

			var output1HighFelt fp.Element
			output1HighFelt.SetBigInt(output1HighBigInt)
			output1HighMv := memory.MemoryValueFromFieldElement(&output1HighFelt)

			err = vm.Memory.WriteToAddress(&output1LowAddr, &output1LowMv)
			if err != nil {
				return err
			}

			err = vm.Memory.WriteToAddress(&output1MidAddr, &output1MidMv)
			if err != nil {
				return err
			}
			return vm.Memory.WriteToAddress(&output1HighAddr, &output1HighMv)
		},
	}
}

func createSplitOutputMidLowHighHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	output1, err := resolver.GetResOperander("output1")
	if err != nil {
		return nil, err
	}

	output1Low, err := resolver.GetResOperander("output1_low")
	if err != nil {
		return nil, err
	}

	output1Mid, err := resolver.GetResOperander("output1_mid")
	if err != nil {
		return nil, err
	}

	output1High, err := resolver.GetResOperander("output1_high")
	if err != nil {
		return nil, err
	}

	return newSplitOutputMidLowHighHint(output1, output1Low, output1Mid, output1High), nil
}

// SplitOutput0 hint splits `output0` into `output0_low` (16 bytes) and `output0_high` (9 bytes)
//
// `newSplitOutput0Hint` takes 3 operanders as arguments
//   - `output0_low` is the variable that will store the low part of `output0`
//   - `output0_high` is the variable that will store the high part of `output0`
//   - `output0` is the value to split
func newSplitOutput0Hint(output0Low, output0High, output0 hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "SplitOutput0",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> ids.output0_low = ids.output0 & ((1 << 128) - 1)
			//> ids.output0_high = ids.output0 >> 128

			output0LowAddr, err := output0Low.GetAddress(vm)
			if err != nil {
				return err
			}

			output0HighAddr, err := output0High.GetAddress(vm)
			if err != nil {
				return err
			}

			output0, err := hinter.ResolveAsFelt(vm, output0)
			if err != nil {
				return err
			}

			output0Uint := uint256.Int(output0.Bits())

			var output0Low uint256.Int
			mask := new(uint256.Int).Lsh(uint256.NewInt(1), 128)
			mask.Sub(mask, uint256.NewInt(1))
			output0Low.And(&output0Uint, mask)
			output0LowBytes := output0Low.Bytes()
			output0LowFelt := fp.Element{}
			output0LowFelt.SetBytes(output0LowBytes)
			output0LowMv := memory.MemoryValueFromFieldElement(&output0LowFelt)

			var output0High uint256.Int
			output0High.Rsh(&output0Uint, 128)
			output0HighBytes := output0High.Bytes()
			output0HighFelt := fp.Element{}
			output0HighFelt.SetBytes(output0HighBytes)
			output0HighMv := memory.MemoryValueFromFieldElement(&output0HighFelt)

			err = vm.Memory.WriteToAddress(&output0LowAddr, &output0LowMv)
			if err != nil {
				return err
			}

			return vm.Memory.WriteToAddress(&output0HighAddr, &output0HighMv)
		},
	}
}

func createSplitOutput0Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	output0Low, err := resolver.GetResOperander("output0_low")
	if err != nil {
		return nil, err
	}

	output0High, err := resolver.GetResOperander("output0_high")
	if err != nil {
		return nil, err
	}

	output0, err := resolver.GetResOperander("output0")
	if err != nil {
		return nil, err
	}

	return newSplitOutput0Hint(output0Low, output0High, output0), nil
}

// SplitNBytes hint assigns to `ids.n_words_to_copy` and `ids.n_bytes_left` variables
// the quotient and remainder of the division of `ids.n_bytes` variable by the
// variable `ids.BYTES_IN_WORD`
//
// `newSplitNBytesHint` takes 3 operanders as arguments
//   - `nWordsToCopy` is the variable that will store the quotient of the division
//   - `nBytesLeft` is the variable that will store the remainder of the division
//   - `nBytes` is the variable that will be divided
func newSplitNBytesHint(nBytes, nWordsToCopy, nBytesLeft hinter.ResOperander) hinter.Hinter {
	name := "SplitNBytes"
	return &GenericZeroHinter{
		Name: name,
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> ids.n_words_to_copy, ids.n_bytes_left = divmod(ids.n_bytes, ids.BYTES_IN_WORD)

			nWordsToCopyAddr, err := nWordsToCopy.GetAddress(vm)
			if err != nil {
				return err
			}

			nBytesLeftAddr, err := nBytesLeft.GetAddress(vm)
			if err != nil {
				return err
			}

			nBytesFelt, err := hinter.ResolveAsFelt(vm, nBytes)
			if err != nil {
				return err
			}

			nBytesBigInt := new(big.Int)
			nBytesFelt.BigInt(nBytesBigInt)

			bytesInWord := big.NewInt(8)

			nWordsToCopyBigInt := new(big.Int)
			nBytesLeftBigInt := new(big.Int)

			nWordsToCopyBigInt.DivMod(nBytesBigInt, bytesInWord, nBytesLeftBigInt)

			var nWordsToCopyFelt fp.Element
			nWordsToCopyFelt.SetBigInt(nWordsToCopyBigInt)
			nWordsToCopyMv := memory.MemoryValueFromFieldElement(&nWordsToCopyFelt)

			var nBytesLeftFelt fp.Element
			nBytesLeftFelt.SetBigInt(nBytesLeftBigInt)
			nBytesLeftMv := memory.MemoryValueFromFieldElement(&nBytesLeftFelt)

			err = vm.Memory.WriteToAddress(&nBytesLeftAddr, &nBytesLeftMv)
			if err != nil {
				return err
			}

			return vm.Memory.WriteToAddress(&nWordsToCopyAddr, &nWordsToCopyMv)
		},
	}
}

func createSplitNBytesHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	nBytes, err := resolver.GetResOperander("n_bytes")
	if err != nil {
		return nil, err
	}

	nWordsToCopy, err := resolver.GetResOperander("n_words_to_copy")
	if err != nil {
		return nil, err
	}

	nBytesLeft, err := resolver.GetResOperander("n_bytes_left")
	if err != nil {
		return nil, err
	}

	return newSplitNBytesHint(nBytes, nWordsToCopy, nBytesLeft), nil
}

// SplitInput3 hint assigns to `ids.high3` and `ids.low3` variables
// the quotient and remainder of the division of the value at memory address
// `ids.inputs + 3` by 256
//
// `newSplitInput3Hint` takes 3 operanders as arguments
//   - `high3` is the variable that will store the quotient of the division
//   - `low3` is the variable that will store the remainder of the division
//   - `inputs` is the address in memory to which we add an offset of 3 and read that value
func newSplitInput3Hint(high3, low3, inputs hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "SplitInput3",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> ids.high3, ids.low3 = divmod(memory[ids.inputs + 3], 256)

			high3Addr, err := high3.GetAddress(vm)
			if err != nil {
				return err
			}

			low3Addr, err := low3.GetAddress(vm)
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
			low3Felt.SetBigInt(low3BigInt)
			low3Mv := memory.MemoryValueFromFieldElement(&low3Felt)

			err = vm.Memory.WriteToAddress(&low3Addr, &low3Mv)
			if err != nil {
				return err
			}

			return vm.Memory.WriteToAddress(&high3Addr, &high3Mv)
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

// SplitInput6 hint assigns to `ids.high6` and `ids.low6` variables
// the quotient and remainder of the division of the value at memory address
// `ids.inputs + 6` by 256 ** 2
//
// `newSplitInput6Hint` takes 3 operanders as arguments
//   - `high6` is the variable that will store the quotient of the division
//   - `low6` is the variable that will store the remainder of the division
//   - `inputs` is the address in memory to which we add an offset of 6 and read that value
func newSplitInput6Hint(high6, low6, inputs hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "SplitInput6",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> ids.high6, ids.low6 = divmod(memory[ids.inputs + 6], 256 ** 2)

			high6Addr, err := high6.GetAddress(vm)
			if err != nil {
				return err
			}

			low6Addr, err := low6.GetAddress(vm)
			if err != nil {
				return err
			}

			inputsAddr, err := hinter.ResolveAsAddress(vm, inputs)
			if err != nil {
				return err
			}

			*inputsAddr, err = inputsAddr.AddOffset(6)
			if err != nil {
				return err
			}

			inputValue, err := vm.Memory.ReadFromAddress(inputsAddr)
			if err != nil {
				return err
			}

			var inputBigInt big.Int
			inputValue.Felt.BigInt(&inputBigInt)

			// 256 ** 2
			divisor := big.NewInt(65536)

			high6BigInt := new(big.Int)
			low6BigInt := new(big.Int)

			high6BigInt.DivMod(&inputBigInt, divisor, low6BigInt)

			var high6Felt fp.Element
			high6Felt.SetBigInt(high6BigInt)
			high6Mv := memory.MemoryValueFromFieldElement(&high6Felt)

			var low6Felt fp.Element
			low6Felt.SetBigInt(low6BigInt)
			low6Mv := memory.MemoryValueFromFieldElement(&low6Felt)

			err = vm.Memory.WriteToAddress(&low6Addr, &low6Mv)
			if err != nil {
				return err
			}

			return vm.Memory.WriteToAddress(&high6Addr, &high6Mv)
		},
	}
}

func createSplitInput6Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	high6, err := resolver.GetResOperander("high6")
	if err != nil {
		return nil, err
	}

	low6, err := resolver.GetResOperander("low6")
	if err != nil {
		return nil, err
	}

	inputs, err := resolver.GetResOperander("inputs")
	if err != nil {
		return nil, err
	}

	return newSplitInput6Hint(high6, low6, inputs), nil
}

// SplitInput9 hint assigns to `ids.high9` and `ids.low9` variables
// the quotient and remainder of the division of the value at memory address
// `ids.inputs + 9` by 256 ** 3
//
// `newSplitInput9Hint` takes 3 operanders as arguments
//   - `high9` is the variable that will store the quotient of the division
//   - `low9` is the variable that will store the remainder of the division
//   - `inputs` is the address in memory to which we add an offset of 9 and read that value
func newSplitInput9Hint(high9, low9, inputs hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "SplitInput9",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> ids.high9, ids.low9 = divmod(memory[ids.inputs + 9], 256 ** 3)

			high9Addr, err := high9.GetAddress(vm)
			if err != nil {
				return err
			}

			low9Addr, err := low9.GetAddress(vm)
			if err != nil {
				return err
			}

			inputsAddr, err := hinter.ResolveAsAddress(vm, inputs)
			if err != nil {
				return err
			}

			*inputsAddr, err = inputsAddr.AddOffset(9)
			if err != nil {
				return err
			}

			inputValue, err := vm.Memory.ReadFromAddress(inputsAddr)
			if err != nil {
				return err
			}

			var inputBigInt big.Int
			inputValue.Felt.BigInt(&inputBigInt)

			// 256 ** 3
			divisor := big.NewInt(16777216)

			high9BigInt := new(big.Int)
			low9BigInt := new(big.Int)

			high9BigInt.DivMod(&inputBigInt, divisor, low9BigInt)

			var high9Felt fp.Element
			high9Felt.SetBigInt(high9BigInt)
			high9Mv := memory.MemoryValueFromFieldElement(&high9Felt)

			var low9Felt fp.Element
			low9Felt.SetBigInt(low9BigInt)
			low9Mv := memory.MemoryValueFromFieldElement(&low9Felt)

			err = vm.Memory.WriteToAddress(&low9Addr, &low9Mv)
			if err != nil {
				return err
			}

			return vm.Memory.WriteToAddress(&high9Addr, &high9Mv)
		},
	}
}

func createSplitInput9Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	high9, err := resolver.GetResOperander("high9")
	if err != nil {
		return nil, err
	}

	low9, err := resolver.GetResOperander("low9")
	if err != nil {
		return nil, err
	}

	inputs, err := resolver.GetResOperander("inputs")
	if err != nil {
		return nil, err
	}

	return newSplitInput9Hint(high9, low9, inputs), nil
}
