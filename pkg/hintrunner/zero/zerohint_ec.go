package zero

import (
	"fmt"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	secp_utils "github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/utils"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"math/big"
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

			pointValues, err := hinter.GetConsecutiveValues(vm, point, int16(6))
			if err != nil {
				return err
			}

			// [y.d0, y.d1, y.d2]
			var pointValuesBig [3]*big.Int
			for i := 0; i < 3; i++ {
				pointValue, err := pointValues[i+3].FieldElement()
				if err != nil {
					return err
				}
				pointValueBig := pointValue.BigInt(new(big.Int))
				pointValuesBig[i] = pointValueBig
			}

			primeBig := fp.Modulus()

			//> y = pack(ids.point.y, PRIME) % SECP_P
			yBig, err := secp_utils.SecPPacked(pointValuesBig[0], pointValuesBig[1], pointValuesBig[2], primeBig)
			if err != nil {
				return err
			}
			yBig.Mod(yBig, secPBig)

			//> value = (-y) % SECP_P
			yBig.Neg(yBig)
			yBig.Mod(yBig, secPBig)

			err = ctx.ScopeManager.AssignVariable("value", yBig)
			if err != nil {
				return err
			}

			return nil
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
