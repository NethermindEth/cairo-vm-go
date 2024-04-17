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
			secPUint256 ok := secp_utils.GetSecPUint256()
			if !ok {
				return fmt.Errorf("GetSecPUint256 failed")
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
			qUint256, rUint256 := uint256.NewInt(0), uint256.NewInt(0)
			qUint256.DivMod(packedValue, secPUint256, rUint256)

			//> assert r == 0, f"verify_zero: Invalid input {ids.val.d0, ids.val.d1, ids.val.d2}."
			if rUint256.Cmp(uint256.NewInt(0)) != 0 {
				return fmt.Errorf("verify_zero: Invalid input (%v, %v, %v).", valValues[0], valValues[1], valValues[2])
			}

			//> ids.q = q % PRIME
			fpModulusUint256 := uint256.FromBig(fp.Modulus())
			qUint256.Mod(qUint256, fpModulusUint256)
			qBig := uint256.ToBig(qUint256)
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
