package zero

import (
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/builtins"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/holiman/uint256"
	"golang.org/x/crypto/sha3"
)

func newCairoKeccakFinalizeHint(keccakStateSizeFeltsResOperander, blockSizeResOperander, keccakPtrEndResOperander hinter.ResOperander) hinter.Hinter {
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

			keccakStateSizeFelts, err := hinter.ResolveAsUint64(vm, keccakStateSizeFeltsResOperander)
			if err != nil {
				return err
			}

			if keccakStateSizeFelts >= 100 {
				return fmt.Errorf("assert 0 <= _keccak_state_size_felts < 100.")
			}
			blockSize, err := hinter.ResolveAsUint64(vm, blockSizeResOperander)
			if err != nil {
				return err
			}
			if blockSize >= 10 {
				return fmt.Errorf("assert 0 <= _block_size < 10.")
			}

			var input [25]uint64
			builtins.KeccakF1600(&input)
			padding := make([]uint64, keccakStateSizeFelts)
			padding = append(padding, input[:]...)
			result := make([]uint64, 0, keccakStateSizeFelts*blockSize)
			for i := uint64(0); i < blockSize; i++ {
				result = append(result, padding...)
			}
			keccakPtrEnd, err := hinter.ResolveAsAddress(vm, keccakPtrEndResOperander)
			if err != nil {
				return err
			}
			keccakPtrEndCopy := *keccakPtrEnd
			for i := 0; i < len(result); i++ {
				resultMV := memory.MemoryValueFromUint(result[i])
				err = vm.Memory.WriteToAddress(&keccakPtrEndCopy, &resultMV)
				if err != nil {
					return err
				}
				keccakPtrEndCopy, err = keccakPtrEndCopy.AddOffset(1)
				if err != nil {
					return err
				}
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

func newUnsafeKeccakHint(data, length, high, low hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "CairoKeccakFinalize",
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
			if err == nil {
				if lengthVal > keccakMaxSize {
					return fmt.Errorf("unsafe_keccak() can only be used with length<=%d.\n Got: length=%d.", keccakMaxSize, lengthVal)
				}
			}
			dataPtr, err := hinter.ResolveAsAddress(vm, data)
			if err != nil {
				return err
			}
			dataPtrCopy := *dataPtr
			keccakInput := make([]byte, 0)
			for i := uint64(0); i < lengthVal; i += 16 {
				wordFelt, err := vm.Memory.ReadAsElement(dataPtrCopy.SegmentIndex, dataPtrCopy.Offset)
				if err != nil {
					return err
				}
				word := uint256.Int(wordFelt.Bits())
				nBytes := lengthVal - i
				if lengthVal-i > 16 {
					nBytes = 16
				}
				if uint64(word.BitLen()) >= 8*nBytes {
					return fmt.Errorf("word %v is out range 0 <= word < 2 ** %d", &word, 8*nBytes)
				}
				wordBytes := word.Bytes20()
				keccakInput = append(keccakInput, wordBytes[20-int(nBytes):]...)
				dataPtrCopy, err = dataPtrCopy.AddOffset(1)
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
			keccakStateMemoryValues, err := hinter.GetConsecutiveValues(vm, keccakStateAddr, 2)
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
			nElems, err := endPtr.SubAddress(startPtr)
			if err != nil {
				return err
			}
			keccakInput := make([]byte, 0)
			memoryValuesInRange, err := hinter.GetConsecutiveValues(vm, *startPtr, int16(nElems))
			if err != nil {
				return err
			}
			for _, mv := range memoryValuesInRange {
				wordFelt, err := mv.FieldElement()
				if err != nil {
					return err
				}
				word := uint256.Int(wordFelt.Bits())
				wordBytes := word.Bytes20()
				keccakInput = append(keccakInput, wordBytes[4:]...)
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
