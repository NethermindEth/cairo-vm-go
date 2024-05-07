package zero

import (
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

// SquashDictInnerAssertLenKeysHint hint asserts that the length
// of the `keys` descending list is zero during the squashing process
//
// `newSquashDictInnerAssertLenKeysHint` doesn't take any operander as argument
func newSquashDictInnerAssertLenKeysHint() hinter.Hinter {
	return &GenericZeroHinter{
		Name: "SquashDictInnerAssertLenKeys",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> assert len(keys) == 0
			keys_, err := ctx.ScopeManager.GetVariableValue("keys")
			if err != nil {
				return err
			}
			keys := keys_.([]f.Element)
			if len(keys) != 0 {
				return fmt.Errorf("assertion `len(keys) == 0` failed")
			}
			return nil
		},
	}
}

func createSquashDictInnerAssertLenKeysHinter() (hinter.Hinter, error) {
	return newSquashDictInnerAssertLenKeysHint(), nil
}
