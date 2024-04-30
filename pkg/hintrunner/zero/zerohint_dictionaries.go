package zero

import (
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func newSquashDictInnerNextKeyHint() hinter.Hinter {
	return &GenericZeroHinter{
		Name: "SquashDictInnerNextKey",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {

			keys_, err := ctx.ScopeManager.GetVariableValue("keys")
			if err != nil {
				return err
			}
			keys := keys_.([]f.Element)

			if len(keys) == 0 {
				return fmt.Errorf("len(keys) == 0` No keys left for processing")
			}

			return nil
		},
	}
}

func createSquashDictInnerNextKeyHinter() (hinter.Hinter, error) {
	return newSquashDictInnerNextKeyHint(), nil
}
