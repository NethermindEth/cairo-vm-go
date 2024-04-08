package zero

import (
	"fmt"
	"math/big"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
)

func newCairoKeccakFinalizeHint(KECCAK_STATE_SIZE_FELTS, BLOCK_SIZE hinter.ResOperander) hinter.Hinter {
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

			keccakStateSizeFelts, err := hinter.ResolveAsFelt(vm, KECCAK_STATE_SIZE_FELTS)
			if err != nil {
				return err
			}
			var keccakStateSizeFeltsBig *big.Int
			keccakStateSizeFelts.BigInt(keccakStateSizeFeltsBig)
			blockSize, err := hinter.ResolveAsFelt(vm, BLOCK_SIZE)
			if err != nil {
				return err
			}
			var blockSizeBig *big.Int
			blockSize.BigInt(blockSizeBig)
			if keccakStateSizeFeltsBig.Cmp(big.NewInt(0)) < 0 || keccakStateSizeFeltsBig.Cmp(big.NewInt(100)) >= 0 {
				return fmt.Errorf("assert 0 <= _keccak_state_size_felts < 100.")
			}
			if blockSizeBig.Cmp(big.NewInt(0)) < 0 || blockSizeBig.Cmp(big.NewInt(10)) >= 0 {
				return fmt.Errorf("assert 0 <= _block_size < 10.")
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
	return newCairoKeccakFinalizeHint(keccakStateSizeFelts, blockSize), nil
}
