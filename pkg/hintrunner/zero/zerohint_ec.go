package zero

import (
	"fmt"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
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

			// 2**256 - 2**32 - 2**9 - 2**8 - 2**7 - 2**6 - 2**4 - 1
			var secPBig big.Int
			secPBig.SetString("115792089237316195423570985008687907853269984665640564039457584007908834671663", 10)

			// 2**86
			var baseBig big.Int
			baseBig.SetString("77371252455336267181195264", 10)

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
				idxBig := new(big.Int).SetInt64(int64(idx))
				valueToAdd := new(big.Int).Exp(&baseBig, idxBig, nil)
				valueToAdd.Mul(valueToAdd, &yDBig)
				packedSum.Add(&packedSum, valueToAdd)
			}

			value := new(big.Int).Neg(&packedSum)
			value.Mod(value, &secPBig)

			ctx.ScopeManager.EnterScope(map[string]any{})
			err = ctx.ScopeManager.AssignVariable("value", value.String())
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

func newNondetBigint3V1Hint(res hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "NondetBigint3V1",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_secp.secp_utils import split
			//> segments.write_arg(ids.res.address_, split(value))

			address, err := res.GetAddress(vm)
			if err != nil {
				return err
			}

			value, err := ctx.ScopeManager.GetVariableValue("value")
			if err != nil {
				return err
			}
			valueBig, ok := value.(*big.Int)
			if !ok {
				return fmt.Errorf("value: %s is not a *big.Int", value)
			}

			baseBig, ok := utils.GetEcBaseBig()
			if !ok {
				return fmt.Errorf("GetEcBaseBig failed")
			}

			var splitValueBig big.Int
			for i := 0; i < 3; i++ {
				valueBig.DivMod(valueBig, baseBig, &splitValueBig)

				splitValueAddr, err := address.AddOffset(int16(i))
				if err != nil {
					return err
				}

				splitValueFelt := new(fp.Element).SetBigInt(&splitValueBig)
				splitValueMv := mem.MemoryValueFromFieldElement(splitValueFelt)

				err = vm.Memory.WriteToAddress(&splitValueAddr, &splitValueMv)
				if err != nil {
					return err
				}
			}

			if valueBig.BitLen() != 0 {
				return fmt.Errorf("value != 0")
			}

			return nil
		},
	}
}

func createNondetBigint3V1Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	res, err := resolver.GetResOperander("res")
	if err != nil {
		return nil, err
	}

	return newNondetBigint3V1Hint(res), nil
}
