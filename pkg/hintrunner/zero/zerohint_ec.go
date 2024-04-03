package zero

import (
	"fmt"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	secp_utils "github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/utils"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func newEcNegateHint(point hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "EcNegate",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack
			//> y = pack(ids.point.y, PRIME) % SECP_P
			//> # The modulo operation in python always returns a nonnegative number.
			//> value = (-y) % SECP_P

			secPBig, ok := secp_utils.GetSecPBig()
			if !ok {
				return fmt.Errorf("GetSecPBig failed")
			}

			pointAddr, err := point.GetAddress(vm)
			if err != nil {
				return err
			}

			pointValues, err := hinter.GetConsecutiveValues(vm, pointAddr, int16(6))
			if err != nil {
				return err
			}

			// [y.d0, y.d1, y.d2]
			var yValues [3]*fp.Element
			for i := 0; i < 3; i++ {
				yValue, err := pointValues[i+3].FieldElement()
				if err != nil {
					return err
				}
				yValues[i] = yValue
			}

			//> y = pack(ids.point.y, PRIME) % SECP_P
			yBig, err := secp_utils.SecPPacked(yValues)
			if err != nil {
				return err
			}
			yBig.Mod(yBig, secPBig)

			//> value = (-y) % SECP_P
			yBig.Neg(yBig)
			yBig.Mod(yBig, secPBig)

			return ctx.ScopeManager.AssignVariable("value", yBig)
		},
	}
}

func createEcNegateHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	point, err := resolver.GetResOperander("point")
	if err != nil {
		return nil, err
	}

	return newEcNegateHint(point), nil
}
