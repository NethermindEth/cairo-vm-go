package zero

import (
	"math/big"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	secp_utils "github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/utils"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/holiman/uint256"
)

func newEcNegateHint(point hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "EcNegate",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack
			//> y = pack(ids.point.y, PRIME) % SECP_P
			//> # The modulo operation in python always returns a nonnegative number.
			//> value = (-y) % SECP_P

			secPUint256 := secp_utils.GetSecPUint256()

			pointAddr, err := point.GetAddress(vm)
			if err != nil {
				return err
			}

			pointMemoryValues, err := hinter.GetConsecutiveValues(vm, pointAddr, int16(6))
			if err != nil {
				return err
			}

			// [y.d0, y.d1, y.d2]
			var pointYValues [3]*fp.Element

			for i := 0; i < 3; i++ {
				pointYValue, err := pointMemoryValues[i+3].FieldElement()
				if err != nil {
					return err
				}
				pointYValues[i] = pointYValue
			}

			//> y = pack(ids.point.y, PRIME) % SECP_P
			yUint256, err := secp_utils.SecPPacked(pointYValues)
			if err != nil {
				return err
			}

			y := yUint256.ToBig()
			secP := secPUint256.ToBig()

			//> value = (-y) % SECP_P
			y.Neg(y)
			y.Mod(y, secP)

			// //> value = (-y) % SECP_P
			// yUint256.Neg(yUint256)
			// yUint256.Mod(yUint256, &secPUint256)

			return ctx.ScopeManager.AssignVariables(map[string]any{"value": y, "SECP_P": secP})
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

			valueBig, err := ctx.ScopeManager.GetVariableValueAsBigInt("value")
			if err != nil {
				return err
			}

			valueUint256, _ := uint256.FromBig(valueBig)

			//> split(value)
			values, err := secp_utils.SecPSplit(valueUint256)
			if err != nil {
				return err
			}

			//> segments.write_arg(ids.res.address_, values)
			for i := 0; i < 3; i++ {
				valueAddr, err := address.AddOffset(int16(i))
				if err != nil {
					return err
				}

				valueFelt := new(fp.Element).SetBigInt(values[i].ToBig())
				valueMv := mem.MemoryValueFromFieldElement(valueFelt)

				err = vm.Memory.WriteToAddress(&valueAddr, &valueMv)
				if err != nil {
					return err
				}
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

			slopeBig, err := ctx.ScopeManager.GetVariableValueAsBigInt("slope")
			if err != nil {
				return err
			}
			x0Big, err := ctx.ScopeManager.GetVariableValueAsBigInt("x0")
			if err != nil {
				return err
			}
			new_xBig, err := ctx.ScopeManager.GetVariableValueAsBigInt("new_x")
			if err != nil {
				return err
			}
			y0Big, err := ctx.ScopeManager.GetVariableValueAsBigInt("y0")
			if err != nil {
				return err
			}
			secPBig, err := ctx.ScopeManager.GetVariableValueAsBigInt("SECP_P")
			if err != nil {
				return err
			}

			new_yBig := new(big.Int)
			new_yBig.Sub(x0Big, new_xBig)
			new_yBig.Mul(new_yBig, slopeBig)
			new_yBig.Sub(new_yBig, y0Big)
			new_yBig.Mod(new_yBig, secPBig)

			valueBig := new(big.Int)
			valueBig.Set(new_yBig)

			return ctx.ScopeManager.AssignVariables(map[string]any{"new_y": new_yBig, "value": valueBig})
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

			slopeAddr, err := slope.GetAddress(vm)
			if err != nil {
				return err
			}
			slopeMemoryValues, err := hinter.GetConsecutiveValues(vm, slopeAddr, int16(3))
			if err != nil {
				return err
			}

			point0Addr, err := point0.GetAddress(vm)
			if err != nil {
				return err
			}
			point0MemoryValues, err := hinter.GetConsecutiveValues(vm, point0Addr, int16(6))
			if err != nil {
				return err
			}

			point1Addr, err := point1.GetAddress(vm)
			if err != nil {
				return err
			}
			point1MemoryValues, err := hinter.GetConsecutiveValues(vm, point1Addr, int16(3))
			if err != nil {
				return err
			}

			// [d0, d1, d2]
			var slopeValues [3]*fp.Element
			// [x.d0, x.d1, x.d2]
			var point0XValues [3]*fp.Element
			// [y.d0, y.d1, y.d2]
			var point0YValues [3]*fp.Element
			// [x.d0, x.d1, x.d2]
			var point1XValues [3]*fp.Element

			for i := 0; i < 3; i++ {
				slopeValue, err := slopeMemoryValues[i].FieldElement()
				if err != nil {
					return err
				}
				slopeValues[i] = slopeValue

				point0XValue, err := point0MemoryValues[i].FieldElement()
				if err != nil {
					return err
				}
				point0XValues[i] = point0XValue

				point1XValue, err := point1MemoryValues[i].FieldElement()
				if err != nil {
					return err
				}
				point1XValues[i] = point1XValue
			}

			for i := 3; i < 6; i++ {
				point0YValue, err := point0MemoryValues[i].FieldElement()
				if err != nil {
					return err
				}
				point0YValues[i-3] = point0YValue
			}

			//> slope = pack(ids.slope, PRIME)
			slopeUint256, err := secp_utils.SecPPacked(slopeValues)
			if err != nil {
				return err
			}

			//> x0 = pack(ids.point0.x, PRIME)
			x0Uint256, err := secp_utils.SecPPacked(point0XValues)
			if err != nil {
				return err
			}

			//> x1 = pack(ids.point1.x, PRIME)
			x1Uint256, err := secp_utils.SecPPacked(point1XValues)
			if err != nil {
				return err
			}

			//> y0 = pack(ids.point0.y, PRIME)
			y0Uint256, err := secp_utils.SecPPacked(point0YValues)
			if err != nil {
				return err
			}

			//> value = new_x = (pow(slope, 2, SECP_P) - x0 - x1) % SECP_P

			secPUint256 := secp_utils.GetSecPUint256()

			new_xBig := new(big.Int)
			new_xBig.Exp(slopeUint256.ToBig(), big.NewInt(2), secPUint256.ToBig())
			new_xBig.Sub(new_xBig, x0Uint256.ToBig())
			new_xBig.Sub(new_xBig, x1Uint256.ToBig())
			new_xBig.Mod(new_xBig, secPUint256.ToBig())

			valueBig := new(big.Int)
			valueBig.Set(new_xBig)

			return ctx.ScopeManager.AssignVariables(map[string]any{"slope": slopeUint256.ToBig(), "x0": x0Uint256.ToBig(), "x1": x1Uint256.ToBig(), "y0": y0Uint256.ToBig(), "new_x": new_xBig, "value": valueBig})
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
