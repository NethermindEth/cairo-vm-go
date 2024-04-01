package zero

import (
	"fmt"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	hintrunnerUtils "github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/utils"
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

func newFastEcAddAssignNewYHint() hinter.Hinter {
	return &GenericZeroHinter{
		Name: "FastEcAddAssignNewY",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> value = new_y = (slope * (x0 - new_x) - y0) % SECP_P

			slope, err := ctx.ScopeManager.GetVariableValue("slope")
			if err != nil {
				return err
			}
			slopeBig, ok := slope.(*big.Int)
			if !ok {
				return fmt.Errorf("value: %s is not a *big.Int", slope)
			}

			x0, err := ctx.ScopeManager.GetVariableValue("x0")
			if err != nil {
				return err
			}
			x0Big, ok := x0.(*big.Int)
			if !ok {
				return fmt.Errorf("value: %s is not a *big.Int", x0)
			}

			new_x, err := ctx.ScopeManager.GetVariableValue("new_x")
			if err != nil {
				return err
			}
			new_xBig, ok := new_x.(*big.Int)
			if !ok {
				return fmt.Errorf("value: %s is not a *big.Int", new_x)
			}

			y0, err := ctx.ScopeManager.GetVariableValue("y0")
			if err != nil {
				return err
			}
			y0Big, ok := y0.(*big.Int)
			if !ok {
				return fmt.Errorf("value: %s is not a *big.Int", y0)
			}

			secPBig, ok := utils.GetSecPBig()
			if !ok {
				return fmt.Errorf("GetSecPBig failed")
			}

			new_yBig := new(big.Int)
			new_yBig.Sub(x0Big, new_xBig)
			new_yBig.Mul(new_yBig, slopeBig)
			new_yBig.Sub(new_yBig, y0Big)
			new_yBig.Mod(new_yBig, secPBig)

			valueBig := new(big.Int)
			valueBig.Set(new_yBig)

			err = ctx.ScopeManager.AssignVariable("new_y", new_yBig)
			if err != nil {
				return err
			}

			err = ctx.ScopeManager.AssignVariable("value", valueBig)
			if err != nil {
				return err
			}

			return nil
		},
	}
}

func createFastEcAddAssignNewYHinter() (hinter.Hinter, error) {
	return newFastEcAddAssignNewYHint(), nil
}

func newFastEcAddAssignNewXHint(slope, point0, point1 hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "FastEcAddAssignNewX",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack
			//
			//> slope = pack(ids.slope, PRIME)
			//> x0 = pack(ids.point0.x, PRIME)
			//> x1 = pack(ids.point1.x, PRIME)
			//> y0 = pack(ids.point0.y, PRIME)
			//> value = new_x = (pow(slope, 2, SECP_P) - x0 - x1) % SECP_P

			primeBig := fp.Modulus()

			slopeValues, err := hinter.GetConsecutiveValues(vm, slope, int16(3))
			if err != nil {
				return err
			}
			point0Values, err := hinter.GetConsecutiveValues(vm, point0, int16(6))
			if err != nil {
				return err
			}
			point1Values, err := hinter.GetConsecutiveValues(vm, point1, int16(3))
			if err != nil {
				return err
			}

			// [d0, d1, d2]
			var slopeValuesBig [3]*big.Int
			// [x.d0, x.d1, x.d2, y.d0, y.d1, y.d2]
			var point0ValuesBig [6]*big.Int
			// [x.d0, x.d1, x.d2]
			var point1ValuesBig [3]*big.Int

			for i := 0; i < 6; i++ {
				if i < 3 {
					slopeValue, err := slopeValues[i].FieldElement()
					if err != nil {
						return err
					}
					slopeValueBig := slopeValue.BigInt(new(big.Int))
					slopeValuesBig[i] = slopeValueBig

					point1Value, err := point1Values[i].FieldElement()
					if err != nil {
						return err
					}
					point1ValueBig := point1Value.BigInt(new(big.Int))
					point1ValuesBig[i] = point1ValueBig
				}

				point0Value, err := point0Values[i].FieldElement()
				if err != nil {
					return err
				}
				point0ValueBig := point0Value.BigInt(new(big.Int))
				point0ValuesBig[i] = point0ValueBig
			}

			//> slope = pack(ids.slope, PRIME)
			slopeBig, err := hintrunnerUtils.SecPPacked(slopeValuesBig[0], slopeValuesBig[1], slopeValuesBig[2], primeBig)
			if err != nil {
				return err
			}

			//> x0 = pack(ids.point0.x, PRIME)
			x0Big, err := hintrunnerUtils.SecPPacked(point0ValuesBig[0], point0ValuesBig[1], point0ValuesBig[2], primeBig)
			if err != nil {
				return err
			}

			//> x1 = pack(ids.point1.x, PRIME)
			x1Big, err := hintrunnerUtils.SecPPacked(point1ValuesBig[0], point1ValuesBig[1], point1ValuesBig[2], primeBig)
			if err != nil {
				return err
			}

			//> y0 = pack(ids.point0.y, PRIME)
			y0Big, err := hintrunnerUtils.SecPPacked(point0ValuesBig[3], point0ValuesBig[4], point0ValuesBig[5], primeBig)
			if err != nil {
				return err
			}

			//> value = new_x = (pow(slope, 2, SECP_P) - x0 - x1) % SECP_P

			secPBig, ok := utils.GetSecPBig()
			if !ok {
				return fmt.Errorf("GetSecPBig failed")
			}

			new_xBig := new(big.Int)
			new_xBig.Exp(slopeBig, big.NewInt(2), secPBig)
			new_xBig.Sub(new_xBig, x0Big)
			new_xBig.Sub(new_xBig, x1Big)
			new_xBig.Mod(new_xBig, secPBig)

			valueBig := new(big.Int)
			valueBig.Set(new_xBig)

			err = ctx.ScopeManager.AssignVariable("slope", slopeBig)
			if err != nil {
				return err
			}

			err = ctx.ScopeManager.AssignVariable("x0", x0Big)
			if err != nil {
				return err
			}

			err = ctx.ScopeManager.AssignVariable("x1", x1Big)
			if err != nil {
				return err
			}

			err = ctx.ScopeManager.AssignVariable("y0", y0Big)
			if err != nil {
				return err
			}

			err = ctx.ScopeManager.AssignVariable("new_x", new_xBig)
			if err != nil {
				return err
			}

			err = ctx.ScopeManager.AssignVariable("value", valueBig)
			if err != nil {
				return err
			}

			return nil
		},
	}
}

func createFastEcAddAssignNewXHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	slope, err := resolver.GetResOperander("slope")
	if err != nil {
		return nil, err
	}
	point0, err := resolver.GetResOperander("point0")
	if err != nil {
		return nil, err
	}
	point1, err := resolver.GetResOperander("point1")
	if err != nil {
		return nil, err
	}

	return newFastEcAddAssignNewXHint(slope, point0, point1), nil
}
