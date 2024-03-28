package zero

import (
	"fmt"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
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

			secPBig, ok := utils.GetSecPBig()
			if !ok {
				return fmt.Errorf("GetSecPBig failed")
			}

			baseBig, ok := utils.GetEcBaseBig()
			if !ok {
				return fmt.Errorf("GetEcBaseBig failed")
			}

			pointValues, err := hinter.GetConsecutiveValues(vm, point, int16(6))
			if err != nil {
				return err
			}

			yD0, err := pointValues[3].FieldElement()
			if err != nil {
				return err
			}
			yD1, err := pointValues[4].FieldElement()
			if err != nil {
				return err
			}
			yD2, err := pointValues[5].FieldElement()
			if err != nil {
				return err
			}

			var packedSum big.Int
			packedSum.SetInt64(0)
			primeBig := fp.Modulus()

			for idx, yD := range []*fp.Element{yD0, yD1, yD2} {
				var yDBig big.Int
				yD.BigInt(&yDBig)
				if yDBig.Cmp(new(big.Int).Div(primeBig, new(big.Int).SetUint64(2))) != -1 {
					yDBig.Sub(&yDBig, primeBig)
				}
				idxBig := big.NewInt(int64(idx))
				valueToAdd := new(big.Int).Exp(baseBig, idxBig, nil)
				valueToAdd.Mul(valueToAdd, &yDBig)
				packedSum.Add(&packedSum, valueToAdd)
			}

			value := new(big.Int).Neg(&packedSum)
			value.Mod(value, secPBig)

			err = ctx.ScopeManager.AssignVariable("value", value)
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
