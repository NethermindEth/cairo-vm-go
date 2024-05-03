package zero

import (
	"fmt"
	"math/big"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	secp_utils "github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/utils"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
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
			yBig, err := secp_utils.SecPPacked(pointYValues)
			if err != nil {
				return err
			}
			yBig.Mod(&yBig, &secPBig)

			//> value = (-y) % SECP_P
			yBig.Neg(&yBig)
			yBig.Mod(&yBig, &secPBig)

			return ctx.ScopeManager.AssignVariables(map[string]any{"value": &yBig, "SECP_P": &secPBig})
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

			//> split(value)
			values, err := secp_utils.SecPSplit(valueBig)
			if err != nil {
				return err
			}

			//> segments.write_arg(ids.res.address_, values)
			for i := 0; i < 3; i++ {
				valueAddr, err := address.AddOffset(int16(i))
				if err != nil {
					return err
				}

				valueFelt := new(fp.Element).SetBigInt(&values[i])
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
			slopeBig, err := secp_utils.SecPPacked(slopeValues)
			if err != nil {
				return err
			}

			//> x0 = pack(ids.point0.x, PRIME)
			x0Big, err := secp_utils.SecPPacked(point0XValues)
			if err != nil {
				return err
			}

			//> x1 = pack(ids.point1.x, PRIME)
			x1Big, err := secp_utils.SecPPacked(point1XValues)
			if err != nil {
				return err
			}

			//> y0 = pack(ids.point0.y, PRIME)
			y0Big, err := secp_utils.SecPPacked(point0YValues)
			if err != nil {
				return err
			}

			//> value = new_x = (pow(slope, 2, SECP_P) - x0 - x1) % SECP_P

			secPBig, ok := secp_utils.GetSecPBig()
			if !ok {
				return fmt.Errorf("GetSecPBig failed")
			}

			new_xBig := new(big.Int)
			new_xBig.Exp(&slopeBig, big.NewInt(2), &secPBig)
			new_xBig.Sub(new_xBig, &x0Big)
			new_xBig.Sub(new_xBig, &x1Big)
			new_xBig.Mod(new_xBig, &secPBig)

			valueBig := new(big.Int)
			valueBig.Set(new_xBig)

			return ctx.ScopeManager.AssignVariables(map[string]any{"slope": &slopeBig, "x0": &x0Big, "x1": &x1Big, "y0": &y0Big, "new_x": new_xBig, "value": valueBig})
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

func newEcDoubleSlopeV1Hint(point hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "EcDoubleSlopeV1",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack
			//> from starkware.python.math_utils import ec_double_slope
			//
			//> # Compute the slope.
			//> x = pack(ids.point.x, PRIME)
			//> y = pack(ids.point.y, PRIME)
			//> value = slope = ec_double_slope(point=(x, y), alpha=0, p=SECP_P)

			pointAddr, err := point.GetAddress(vm)
			if err != nil {
				return err
			}
			pointMemoryValues, err := hinter.GetConsecutiveValues(vm, pointAddr, int16(6))
			if err != nil {
				return err
			}

			// [x.d0, x.d1, x.d2]
			var pointXValues [3]*fp.Element
			// [y.d0, y.d1, y.d2]
			var pointYValues [3]*fp.Element

			for i := 0; i < 3; i++ {
				pointValue, err := pointMemoryValues[i].FieldElement()
				if err != nil {
					return err
				}
				pointXValues[i] = pointValue
			}
			for i := 3; i < 6; i++ {
				pointValue, err := pointMemoryValues[i].FieldElement()
				if err != nil {
					return err
				}
				pointYValues[i-3] = pointValue
			}

			//> x = pack(ids.point.x, PRIME)
			xBig, err := secp_utils.SecPPacked(pointXValues)
			if err != nil {
				return err
			}

			//> y = pack(ids.point.y, PRIME)
			yBig, err := secp_utils.SecPPacked(pointYValues)
			if err != nil {
				return err
			}

			secPBig, ok := secp_utils.GetSecPBig()
			if !ok {
				return fmt.Errorf("GetSecPBig failed")
			}

			//> value = slope = ec_double_slope(point=(x, y), alpha=0, p=SECP_P)
			valueBig, err := secp_utils.EcDoubleSlope(&xBig, &yBig, big.NewInt(0), &secPBig)
			if err != nil {
				return err
			}

			slopeBig := new(big.Int).Set(&valueBig)

			return ctx.ScopeManager.AssignVariables(map[string]any{"x": &xBig, "y": &yBig, "value": &valueBig, "slope": slopeBig})
		},
	}
}

func createEcDoubleSlopeV1Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	point, err := resolver.GetResOperander("point")
	if err != nil {
		return nil, err
	}

	return newEcDoubleSlopeV1Hint(point), nil
}

func newEcDoubleAssignNewXV1Hint(slope, point hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "EcDoubleAssignNewXV1",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack

			//> slope = pack(ids.slope, PRIME)
			//> x = pack(ids.point.x, PRIME)
			//> y = pack(ids.point.y, PRIME)

			//> value = new_x = (pow(slope, 2, SECP_P) - 2 * x) % SECP_P

			slopeAddr, err := slope.GetAddress(vm)
			if err != nil {
				return err
			}
			slopeMemoryValues, err := hinter.GetConsecutiveValues(vm, slopeAddr, int16(3))
			if err != nil {
				return err
			}

			pointAddr, err := point.GetAddress(vm)
			if err != nil {
				return err
			}
			pointMemoryValues, err := hinter.GetConsecutiveValues(vm, pointAddr, int16(6))
			if err != nil {
				return err
			}

			// [d0, d1, d2]
			var slopeValues [3]*fp.Element
			// [x.d0, x.d1, x.d2]
			var pointXValues [3]*fp.Element
			// [y.d0, y.d1, y.d2]
			var pointYValues [3]*fp.Element

			for i := 0; i < 3; i++ {
				slopeValue, err := slopeMemoryValues[i].FieldElement()
				if err != nil {
					return err
				}
				slopeValues[i] = slopeValue

				pointXValue, err := pointMemoryValues[i].FieldElement()
				if err != nil {
					return err
				}
				pointXValues[i] = pointXValue

				pointYValue, err := pointMemoryValues[i+3].FieldElement()
				if err != nil {
					return err
				}
				pointYValues[i] = pointYValue
			}

			//> slope = pack(ids.slope, PRIME)
			slopeBig, err := secp_utils.SecPPacked(slopeValues)
			if err != nil {
				return err
			}

			//> x = pack(ids.point.x, PRIME)
			xBig, err := secp_utils.SecPPacked(pointXValues)
			if err != nil {
				return err
			}

			//> y = pack(ids.point.y, PRIME)
			yBig, err := secp_utils.SecPPacked(pointYValues)
			if err != nil {
				return err
			}

			//> value = new_x = (pow(slope, 2, SECP_P) - 2 * x) % SECP_P
			secPBig, ok := secp_utils.GetSecPBig()
			if !ok {
				return fmt.Errorf("GetSecPBig failed")
			}

			multRes := new(big.Int)
			multRes.Mul(big.NewInt(2), &xBig)

			new_xBig := new(big.Int)
			new_xBig.Exp(&slopeBig, big.NewInt(2), &secPBig)
			new_xBig.Sub(new_xBig, multRes)
			new_xBig.Mod(new_xBig, &secPBig)

			valueBig := new(big.Int)
			valueBig.Set(new_xBig)

			return ctx.ScopeManager.AssignVariables(map[string]any{"slope": &slopeBig, "x": &xBig, "y": &yBig, "new_x": new_xBig, "value": valueBig})
		},
	}
}

func createEcDoubleAssignNewXV1Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	slope, err := resolver.GetResOperander("slope")
	if err != nil {
		return nil, err
	}
	point, err := resolver.GetResOperander("point")
	if err != nil {
		return nil, err
	}

	return newEcDoubleAssignNewXV1Hint(slope, point), nil
}

func newComputeSlopeV1Hint(point0, point1 hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "ComputeSlopeV1",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack
			//> from starkware.python.math_utils import line_slope

			//> # Compute the slope.
			//> x0 = pack(ids.point0.x, PRIME)
			//> y0 = pack(ids.point0.y, PRIME)
			//> x1 = pack(ids.point1.x, PRIME)
			//> y1 = pack(ids.point1.y, PRIME)
			//> value = slope = line_slope(point1=(x0, y0), point2=(x1, y1), p=SECP_P)

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
			point1MemoryValues, err := hinter.GetConsecutiveValues(vm, point1Addr, int16(6))
			if err != nil {
				return err
			}

			// [x.d0, x.d1, x.d2]
			var point0XValues [3]*fp.Element
			// [y.d0, y.d1, y.d2]
			var point0YValues [3]*fp.Element
			// [x.d0, x.d1, x.d2]
			var point1XValues [3]*fp.Element
			// [y.d0, y.d1, y.d2]
			var point1YValues [3]*fp.Element

			for i := 0; i < 3; i++ {
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

				point1YValue, err := point1MemoryValues[i].FieldElement()
				if err != nil {
					return err
				}
				point1YValues[i-3] = point1YValue
			}

			//> x0 = pack(ids.point0.x, PRIME)
			x0Big, err := secp_utils.SecPPacked(point0XValues)
			if err != nil {
				return err
			}

			//> x1 = pack(ids.point1.x, PRIME)
			x1Big, err := secp_utils.SecPPacked(point1XValues)
			if err != nil {
				return err
			}

			//> y0 = pack(ids.point0.y, PRIME)
			y0Big, err := secp_utils.SecPPacked(point0YValues)
			if err != nil {
				return err
			}

			//> y1 = pack(ids.point0.y, PRIME)
			y1Big, err := secp_utils.SecPPacked(point1YValues)
			if err != nil {
				return err
			}

			secPBig, ok := secp_utils.GetSecPBig()
			if !ok {
				return fmt.Errorf("GetSecPBig failed")
			}

			// value = slope = line_slope(point1=(x0, y0), point2=(x1, y1), p=SECP_P)
			slopeBig, err := secp_utils.LineSlope(&x0Big, &y0Big, &x1Big, &y1Big, &secPBig)
			if err != nil {
				return err
			}

			value := new(big.Int).Set(&slopeBig)

			return ctx.ScopeManager.AssignVariables(map[string]any{"value": value, "slope": value})
		},
	}
}

func createComputeSlopeV1Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	point0, err := resolver.GetResOperander("point0")
	if err != nil {
		return nil, err
	}

	point1, err := resolver.GetResOperander("point1")
	if err != nil {
		return nil, err
	}

	return newComputeSlopeV1Hint(point0, point1), nil
}
