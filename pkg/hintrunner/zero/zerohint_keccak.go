package zero

import (
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/builtins"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
)

func newCairoKeccakFinalizeHint(keccakStateSizeFeltsResOperander, blockSizeResOperander, keccakPtrEndResOperander hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "IsLeFelt",
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
			blockSize, err := hinter.ResolveAsUint64(vm, blockSizeResOperander)
			if err != nil {
				return err
			}
			keccakPtrEnd, err := hinter.ResolveAsAddress(vm, keccakPtrEndResOperander)
			if err != nil {
				return err
			}
			if keccakStateSizeFelts < 0 || keccakStateSizeFelts >= 100 {
				return fmt.Errorf("assert 0 <= _keccak_state_size_felts < 100.")
			}
			if blockSize < 0 || blockSize >= 10 {
				return fmt.Errorf("assert 0 <= _block_size < 10.")
			}

			var input [25]uint64
			builtins.KeccakF1600(&input)
			padding := make([]uint64, keccakStateSizeFelts)
			padding = append(padding, input[:]...)
			result := make([]uint64, keccakStateSizeFelts*blockSize)
			for i := uint64(0); i < blockSize; i++ {
				result = append(result, padding...)
			}
			keccakPtrEndCopy := *keccakPtrEnd
			for i := 0; i < len(result); i++ {
				multiplicitesSegmentWriteArgsPtr, err := keccakPtrEndCopy.AddOffset(int16(i))
				if err != nil {
					return err
				}
				resultMV := memory.MemoryValueFromUint(result[i])
				err = vm.Memory.WriteToAddress(&multiplicitesSegmentWriteArgsPtr, &resultMV)
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
