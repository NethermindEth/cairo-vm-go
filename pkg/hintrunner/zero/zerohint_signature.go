package zero

import (
	"fmt"
	"math/big"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	secp_utils "github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/utils"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func newVerifyZeroHint(val, q hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "VerifyZero",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack
			//> q, r = divmod(pack(ids.val, PRIME), SECP_P)
			//> assert r == 0, f"verify_zero: Invalid input {ids.val.d0, ids.val.d1, ids.val.d2}."
			//> ids.q = q % PRIME

			//> from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack
			secPBig, ok := secp_utils.GetSecPBig()
			if !ok {
				return fmt.Errorf("GetSecPBig failed")
			}

			valAddr, err := val.GetAddress(vm)
			if err != nil {
				return err
			}

			valMemoryValues, err := hinter.GetConsecutiveValues(vm, valAddr, int16(3))
			if err != nil {
				return err
			}

			// [d0, d1, d2]
			var valValues [3]*fp.Element

			for i := 0; i < 3; i++ {
				valValue, err := valMemoryValues[i].FieldElement()
				if err != nil {
					return err
				}
				valValues[i] = valValue
			}

			//> q, r = divmod(pack(ids.val, PRIME), SECP_P)
			packedValue, err := secp_utils.SecPPacked(valValues)
			if err != nil {
				return err
			}
			qBig, rBig := new(big.Int), new(big.Int)
			qBig.DivMod(packedValue, secPBig, rBig)

			//> assert r == 0, f"verify_zero: Invalid input {ids.val.d0, ids.val.d1, ids.val.d2}."
			if rBig.Cmp(big.NewInt(0)) != 0 {
				return fmt.Errorf("verify_zero: Invalid input (%v, %v, %v).", valValues[0], valValues[1], valValues[2])
			}

			//> ids.q = q % PRIME
			qBig.Mod(qBig, fp.Modulus())
			qFelt := new(fp.Element).SetBigInt(qBig)
			qAddr, err := q.GetAddress(vm)
			if err != nil {
				return err
			}
			qMv := memory.MemoryValueFromFieldElement(qFelt)
			return vm.Memory.WriteToAddress(&qAddr, &qMv)
		},
	}
}

func createVerifyZeroHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	val, err := resolver.GetResOperander("val")
	if err != nil {
		return nil, err
	}
	q, err := resolver.GetResOperander("q")
	if err != nil {
		return nil, err
	}

	return newVerifyZeroHint(val, q), nil
}
