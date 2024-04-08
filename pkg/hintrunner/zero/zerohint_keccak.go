package zero

import (
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
)

func newCairoKeccakFinalizeHint(KECCAK_STATE_SIZE_FELTS, BLOCK_SIZE hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "IsLeFelt",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> memory[ap] = 0 if (ids.a % PRIME) <= (ids.b % PRIME) else 1
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
	return newIsLeFeltHint(keccakStateSizeFelts, blockSize), nil
}
