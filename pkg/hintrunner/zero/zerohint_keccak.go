package zero

import (
	"fmt"
	"math"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/builtins"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/holiman/uint256"
	"golang.org/x/crypto/sha3"
)

// CairoKeccakFinalize writes the result of F1600 Keccak permutation padded by __keccak_state_size_felts__ zeros to consecutive memory cells, __block_size__ times.
//
// `CairoKeccakFinalize` takes 3 operanders as arguments
//   - `keccakStateSizeFelts` is the number of felts in the Keccak state size
//   - `blockSize` is the number of blocks to write
//   - `keccakPtrEnd` is the address in memory where to start writing the result
//
// The `low` and `high` parts are splitted in 64-bit integers
// Ultimately, the result is written into 4 memory cells
func newCairoKeccakFinalizeHint(keccakStateSizeFelts, blockSize, keccakPtrEnd hinter.ResOperander) hinter.Hinter {
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

			keccakStateSizeFeltsVal, err := hinter.ResolveAsUint64(vm, keccakStateSizeFelts)
			if err != nil {
				return err
			}
			if keccakStateSizeFeltsVal >= 100 {
				return fmt.Errorf("assertion failed: 0 <= keccak_state_size_felts < 100.")
			}
			blockSizeVal, err := hinter.ResolveAsUint64(vm, blockSize)
			if err != nil {
				return err
			}
			if blockSizeVal >= 10 {
				return fmt.Errorf("assertion failed: 0 <= _block_size < 10.")
			}

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
	keccakStateSizeFelts, err := resolver.GetResOperander("KECCAK_STATE_SIZE_FELTS")
	if err != nil {
		return nil, err
	}
	blockSize, err := resolver.GetResOperander("BLOCK_SIZE")
	if err != nil {
		return nil, err
	}
	keccakPtrEnd, err := resolver.GetResOperander("keccak_ptr_end")
	if err != nil {
		return nil, err
	}
	return newCairoKeccakFinalizeHint(keccakStateSizeFelts, blockSize, keccakPtrEnd), nil
}

// UnsafeKeccak computes keccak hash of the data in memory without validity enforcement and writes the result in the `low` and `high` memory cells
//
// `newUnsafeKeccakHint` takes 4 operanders as arguments
//   - `data` is the address in memory to the base of the data array to hash is stored. Each word in the array is 16 bytes long, except the last one
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

			lengthVal, err := hinter.ResolveAsUint64(vm, length)
			if err != nil {
				return err
			}
			keccakMaxSize, err := ctx.ScopeManager.GetVariableValueAsUint64("__keccak_max_size")
			if err != nil {
				return err
			}
			if lengthVal > keccakMaxSize {
				return fmt.Errorf("unsafe_keccak() can only be used with length<=%d.\n Got: length=%d.", keccakMaxSize, lengthVal)
			}
			dataPtr, err := hinter.ResolveAsAddress(vm, data)
			if err != nil {
				return err
			}

			keccakInput := make([]byte, 0)
			for i := uint64(0); i < lengthVal; i += 16 {
				wordFelt, err := vm.Memory.ReadAsElement(dataPtr.SegmentIndex, dataPtr.Offset)
				if err != nil {
					return err
				}
				word := uint256.Int(wordFelt.Bits())
				nBytes := lengthVal - i
				if nBytes > 16 {
					nBytes = 16
				}
				if uint64(word.BitLen()) >= 8*nBytes {
					return fmt.Errorf("word %v is out range 0 <= word < 2 ** %d", &word, 8*nBytes)
				}
				wordBytes := word.Bytes20()
				keccakInput = append(keccakInput, wordBytes[20-int(nBytes):]...)
				*dataPtr, err = dataPtr.AddOffset(1)
				if err != nil {
					return err
				}
			}
			hash := sha3.NewLegacyKeccak256()
			hash.Write(keccakInput)
			hashedBytes := hash.Sum(nil)
			hashedHigh := new(fp.Element).SetBytes(hashedBytes[:16])
			hashedLow := new(fp.Element).SetBytes(hashedBytes[16:32])
			highAddr, err := high.GetAddress(vm)
			if err != nil {
				return err
			}
			hashedHighMV := memory.MemoryValueFromFieldElement(hashedHigh)
			err = vm.Memory.WriteToAddress(&highAddr, &hashedHighMV)
			if err != nil {
				return err
			}
			lowAddr, err := low.GetAddress(vm)
			if err != nil {
				return err
			}
			hashedLowMV := memory.MemoryValueFromFieldElement(hashedLow)
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
			// segments.write_arg(ids.inputs, [ids.low % 2 ** 64, ids.low // 2 ** 64])
			// segments.write_arg(ids.inputs + 2, [ids.high % 2 ** 64, ids.high // 2 ** 64])

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
