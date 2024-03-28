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

			//> def pack(z, prime):
			//>     """
			//>     Takes an UnreducedBigInt3 struct which represents a triple of limbs (d0, d1, d2) of field
			//>     elements and reconstructs the corresponding 256-bit integer (see split()).
			//>     Note that the limbs do not have to be in the range [0, BASE).
			//>     prime should be the Cairo field, and it is used to handle negative values of the limbs.
			//>     """
			//>     limbs = z.d0, z.d1, z.d2
			//>     return sum(as_int(limb, prime) * (BASE**i) for i, limb in enumerate(limbs))

			valueBig := big.NewInt(0)
			primeBig := fp.Modulus()

			for idx, yD := range []*fp.Element{yD0, yD1, yD2} {
				var yDBig big.Int
				yD.BigInt(&yDBig)

				//> as_int(limb, prime)
				if yDBig.Cmp(new(big.Int).Div(primeBig, big.NewInt(2))) != -1 {
					yDBig.Sub(&yDBig, primeBig)
				}

				valueToAddBig := new(big.Int).Exp(baseBig, big.NewInt(int64(idx)), nil)
				valueToAddBig.Mul(valueToAddBig, &yDBig)
				valueBig.Add(valueBig, valueToAddBig)
			}

			//> value = (-y) % SECP_P
			valueBig.Neg(valueBig)
			valueBig.Mod(valueBig, secPBig)

			err = ctx.ScopeManager.AssignVariable("value", valueBig)
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

			//> def split(num: int) -> List[int]:
			//>     """
			//?     Takes a 256-bit integer and returns its canonical representation as:
			//>         d0 + BASE * d1 + BASE**2 * d2,
			//>     where BASE = 2**86.
			//>     """
			//?     a = []
			//>     for _ in range(3):
			//>         num, residue = divmod(num, BASE)
			//>         a.append(residue)
			//>     assert num == 0
			//>     return a
			//>
			//> segments.write_arg(ids.res.address_, split(value))

			var residue big.Int
			for i := 0; i < 3; i++ {
				//> num, residue = divmod(num, BASE)
				valueBig.DivMod(valueBig, baseBig, &residue)

				residueAddr, err := address.AddOffset(int16(i))
				if err != nil {
					return err
				}

				residueFelt := new(fp.Element).SetBigInt(&residue)
				residueMv := mem.MemoryValueFromFieldElement(residueFelt)

				err = vm.Memory.WriteToAddress(&residueAddr, &residueMv)
				if err != nil {
					return err
				}
			}

			//> assert num == 0
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
