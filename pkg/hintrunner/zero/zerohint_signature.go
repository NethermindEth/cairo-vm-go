package zero

import (
	"fmt"
	"math/big"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	secp_utils "github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/utils"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func newDivModNPackedDivmodV1Hint(a, b hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "DivModNPackedDivmodV1",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_secp.secp_utils import N, pack
			//> from starkware.python.math_utils import div_mod, safe_div
			//> a = pack(ids.a, PRIME)
			//> b = pack(ids.b, PRIME)
			//> value = res = div_mod(a, b, N)

			aAddr, err := a.GetAddress(vm)
			if err != nil {
				return err
			}
			aMemoryValues, err := hinter.GetConsecutiveValues(vm, aAddr, int16(3))
			if err != nil {
				return err
			}

			bAddr, err := b.GetAddress(vm)
			if err != nil {
				return err
			}
			bMemoryValues, err := hinter.GetConsecutiveValues(vm, bAddr, int16(3))
			if err != nil {
				return err
			}

			var aValues [3]*fp.Element
			var bValues [3]*fp.Element

			for i := 0; i < 3; i++ {
				aValue, err := aMemoryValues[i].FieldElement()
				if err != nil {
					return err
				}
				aValues[i] = aValue

				bValue, err := bMemoryValues[i].FieldElement()
				if err != nil {
					return err
				}
				bValues[i] = bValue
			}

			//> a = pack(ids.a, PRIME)
			aPackedBig, err := secp_utils.SecPPacked(aValues)
			if err != nil {
				return err
			}

			//> b = pack(ids.b, PRIME)
			bPackedBig, err := secp_utils.SecPPacked(bValues)
			if err != nil {
				return err
			}

			nBig, ok := secp_utils.GetN()
			if !ok {
				return fmt.Errorf("GetN failed")
			}

			//> value = res = div_mod(a, b, N)
			resBig, err := secp_utils.Div_mod(aPackedBig, bPackedBig, nBig)
			if err != nil {
				return err
			}
			valueBig := new(big.Int).Set(resBig)

			return ctx.ScopeManager.AssignVariables(map[string]any{"res": resBig, "value": valueBig})
		},
	}
}

func createDivModNPackedDivmodV1Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	a, err := resolver.GetResOperander("a")
	if err != nil {
		return nil, err
	}
	b, err := resolver.GetResOperander("b")
	if err != nil {
		return nil, err
	}

	return newDivModNPackedDivmodV1Hint(a, b), nil
}
