package zero

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	hintrunner_utils "github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/utils"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/holiman/uint256"
)

// EcNegate hint negates the y-coordinate of a point on an elliptic curve modulo SECP_P
//
// `newEcNegateHint` takes 1 operander as argument
//   - `point` is the point on an elliptic curve to operate on
//
// `newEcNegateHint` assigns the result as `value` in the current scope
func newEcNegateHint(point hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "EcNegate",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack
			//> y = pack(ids.point.y, PRIME) % SECP_P
			//> # The modulo operation in python always returns a nonnegative number.
			//> value = (-y) % SECP_P

			secPBig, ok := hintrunner_utils.GetSecPBig()
			if !ok {
				return fmt.Errorf("GetSecPBig failed")
			}

			pointAddr, err := point.GetAddress(vm)
			if err != nil {
				return err
			}

			pointYAddr, err := pointAddr.AddOffset(3)
			if err != nil {
				return err
			}

			pointYValues, err := vm.Memory.ResolveAsBigInt3(pointYAddr)
			if err != nil {
				return err
			}

			//> y = pack(ids.point.y, PRIME) % SECP_P
			yBig, err := hintrunner_utils.SecPPacked(pointYValues)
			if err != nil {
				return err
			}
			yBig.Mod(&yBig, &secPBig)

			//> value = (-y) % SECP_P
			yBig.Neg(&yBig)
			yBig.Mod(&yBig, &secPBig)

			return ctx.ScopeManager.AssignVariables(map[string]any{"value": &yBig})
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

// NondetBigint3V1 hint writes a value to a specified segment of memory
//
// `newNondetBigint3V1Hint` takes 1 operander as argument
//   - `res` is the location in memory where to write the result
//
// `newNondetBigint3V1Hint` uses `SecPSplit` to split the value in 3 felts and writes the result in memory
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
			values, err := hintrunner_utils.SecPSplit(valueBig)
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

// FastEcAddAssignNewX hint computes a new x-coordinate for fast elliptic curve addition
//
// `newFastEcAddAssignNewXHint` takes 3 operanders as arguments
//   - `slope` is the slope of the line connecting `point0` and `point1`
//   - `point0` and `point1` are 2 points on an elliptic curve
//
// `newFastEcAddAssignNewXHint` assigns the new x-coordinate as `value` in the current scope
// It also assigns `slope`, `x0`, `y0` and `new_x` in the current scope
// so that they are available in the current scope for FastEcAddAssignNewY hint
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

			point0Addr, err := point0.GetAddress(vm)
			if err != nil {
				return err
			}

			point1Addr, err := point1.GetAddress(vm)
			if err != nil {
				return err
			}

			point0YAddr, err := point0Addr.AddOffset(3)
			if err != nil {
				return err
			}

			slopeValues, err := vm.Memory.ResolveAsBigInt3(slopeAddr)
			if err != nil {
				return err
			}

			point0XValues, err := vm.Memory.ResolveAsBigInt3(point0Addr)
			if err != nil {
				return err
			}

			point1XValues, err := vm.Memory.ResolveAsBigInt3(point1Addr)
			if err != nil {
				return err
			}

			point0YValues, err := vm.Memory.ResolveAsBigInt3(point0YAddr)
			if err != nil {
				return err
			}

			//> slope = pack(ids.slope, PRIME)
			slopeBig, err := hintrunner_utils.SecPPacked(slopeValues)
			if err != nil {
				return err
			}

			//> x0 = pack(ids.point0.x, PRIME)
			x0Big, err := hintrunner_utils.SecPPacked(point0XValues)
			if err != nil {
				return err
			}

			//> x1 = pack(ids.point1.x, PRIME)
			x1Big, err := hintrunner_utils.SecPPacked(point1XValues)
			if err != nil {
				return err
			}

			//> y0 = pack(ids.point0.y, PRIME)
			y0Big, err := hintrunner_utils.SecPPacked(point0YValues)
			if err != nil {
				return err
			}

			//> value = new_x = (pow(slope, 2, SECP_P) - x0 - x1) % SECP_P

			secPBig, ok := hintrunner_utils.GetSecPBig()
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

			return ctx.ScopeManager.AssignVariables(map[string]any{"slope": &slopeBig, "x0": &x0Big, "y0": &y0Big, "new_x": new_xBig, "value": valueBig})
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

// FastEcAddAssignNewY hint computes a new y-coordinate for fast elliptic curve addition
// of two different points
// This hint is ultimately used for either multiplying a point with an integer with `ec_mul_by_uint256`
// Cairo function, or for adding two different points with `ec_add` Cairo function
//
// `newFastEcAddAssignNewYHint` doesn't take any operander as argument
// This hint follows the execution of `FastEcAddAssignNewX` hint when computing the addition of two given points
// This is why all variables are already accessible in the current scope
//
// `newFastEcAddAssignNewYHint` assigns the new y-coordinate as `value` in the current scope
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

			valueBig := ComputeYCoordinate(slopeBig, x0Big, new_xBig, y0Big, secPBig)

			return ctx.ScopeManager.AssignVariables(map[string]any{"value": valueBig})
		},
	}
}

func createFastEcAddAssignNewYHinter() (hinter.Hinter, error) {
	return newFastEcAddAssignNewYHint(), nil
}

// EcDoubleSlopeV1 hint computes the slope for doubling a point on an elliptic curve
//
// `newEcDoubleSlopeV1Hint` takes 1 operander as argument
//   - `point` is the point on an elliptic curve to operate on
//
// `newEcDoubleSlopeV1Hint` assigns the `slope` result as `value` in the current scope
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

			pointYAddr, err := pointAddr.AddOffset(3)
			if err != nil {
				return err
			}

			pointXValues, err := vm.Memory.ResolveAsBigInt3(pointAddr)
			if err != nil {
				return err
			}

			pointYValues, err := vm.Memory.ResolveAsBigInt3(pointYAddr)
			if err != nil {
				return err
			}

			//> x = pack(ids.point.x, PRIME)
			xBig, err := hintrunner_utils.SecPPacked(pointXValues)
			if err != nil {
				return err
			}

			//> y = pack(ids.point.y, PRIME)
			yBig, err := hintrunner_utils.SecPPacked(pointYValues)
			if err != nil {
				return err
			}

			secPBig, ok := hintrunner_utils.GetSecPBig()
			if !ok {
				return fmt.Errorf("GetSecPBig failed")
			}

			//> value = slope = ec_double_slope(point=(x, y), alpha=0, p=SECP_P)
			valueBig, err := hintrunner_utils.EcDoubleSlope(&xBig, &yBig, big.NewInt(0), &secPBig)
			if err != nil {
				return err
			}

			return ctx.ScopeManager.AssignVariables(map[string]any{"value": &valueBig})
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

// ReduceV1 hint reduces a packed value modulo the SECP256R1 prime
//
// `newReduceV1Hint` takes 1 operander as argument
//   - `x` is the packed value to be reduced
//
// `newReduceV1Hint` assigns the result as `value` in the current scope
func newReduceV1Hint(x hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "ReduceV1",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack
			//> value = pack(ids.x, PRIME) % SECP_P

			secPBig, ok := hintrunner_utils.GetSecPBig()
			if !ok {
				return fmt.Errorf("GetSecPBig failed")
			}

			xAddr, err := x.GetAddress(vm)
			if err != nil {
				return err
			}

			xValues, err := vm.Memory.ResolveAsBigInt3(xAddr)
			if err != nil {
				return err
			}

			xBig, err := hintrunner_utils.SecPPacked(xValues)
			if err != nil {
				return err
			}

			xBig.Mod(&xBig, &secPBig)
			valueBigIntPtr := new(big.Int).Set(&xBig)

			return ctx.ScopeManager.AssignVariable("value", valueBigIntPtr)
		},
	}
}

func createReduceV1Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	x, err := resolver.GetResOperander("x")
	if err != nil {
		return nil, err
	}

	return newReduceV1Hint(x), nil
}

// EcDoubleAssignNewXV1 hint computes a new x-coordinate for a point being doubled on an elliptic curve
//
// `newEcDoubleAssignNewXV1Hint` takes 2 operanders as arguments
//   - `slope` is the slope for doubling a point, computed with EcDoubleSlopeV1 hint
//   - `point` is the point on an elliptic curve to operate on
//
// `newEcDoubleAssignNewXV1Hint` assigns the `new_x` result as `value` in the current scope
// It also assigns `slope`, `x`, `y` and `new_x` in the current scope
// so that they are available in the current scope for EcDoubleAssignNewYV1 hint
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

			pointAddr, err := point.GetAddress(vm)
			if err != nil {
				return err
			}

			pointYAddr, err := pointAddr.AddOffset(3)
			if err != nil {
				return err
			}

			slopeValues, err := vm.Memory.ResolveAsBigInt3(slopeAddr)
			if err != nil {
				return err
			}

			pointXValues, err := vm.Memory.ResolveAsBigInt3(pointAddr)
			if err != nil {
				return err
			}

			pointYValues, err := vm.Memory.ResolveAsBigInt3(pointYAddr)
			if err != nil {
				return err
			}

			//> slope = pack(ids.slope, PRIME)
			slopeBig, err := hintrunner_utils.SecPPacked(slopeValues)
			if err != nil {
				return err
			}

			//> x = pack(ids.point.x, PRIME)
			xBig, err := hintrunner_utils.SecPPacked(pointXValues)
			if err != nil {
				return err
			}

			//> y = pack(ids.point.y, PRIME)
			yBig, err := hintrunner_utils.SecPPacked(pointYValues)
			if err != nil {
				return err
			}

			//> value = new_x = (pow(slope, 2, SECP_P) - 2 * x) % SECP_P
			secPBig, ok := hintrunner_utils.GetSecPBig()
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

			return ctx.ScopeManager.AssignVariables(map[string]any{"slope": &slopeBig, "x": &xBig, "y": &yBig, "new_x": new_xBig, "value": valueBig, "SECP_P": &secPBig})
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

// EcDoubleAssignNewYV1 hint computes a new y-coordinate when doubling a point
// on an elliptic curve
// This hint is ultimately used for either multiplying a point with an integer with `ec_mul_by_uint256`
// Cairo function, or for adding a given point to itself with `ec_add` Cairo function
//
// `newEcDoubleAssignNewYV1Hint` doesn't take any operander as argument
// This hint follows the execution of `EcDoubleAssignNewXV1` hint when doubling a point
// This is why all variables are already accessible in the current scope
//
// `newEcDoubleAssignNewYV1Hint` assigns the new y-coordinate as `value` in the current scope
func newEcDoubleAssignNewYV1Hint() hinter.Hinter {
	return &GenericZeroHinter{
		Name: "EcDoubleAssignNewYV1",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> value = new_y = (slope * (x - new_x) - y) % SECP256R1_P

			slopeBig, err := ctx.ScopeManager.GetVariableValueAsBigInt("slope")
			if err != nil {
				return err
			}
			xBig, err := ctx.ScopeManager.GetVariableValueAsBigInt("x")
			if err != nil {
				return err
			}
			new_xBig, err := ctx.ScopeManager.GetVariableValueAsBigInt("new_x")
			if err != nil {
				return err
			}
			yBig, err := ctx.ScopeManager.GetVariableValueAsBigInt("y")
			if err != nil {
				return err
			}
			secPBig, err := ctx.ScopeManager.GetVariableValueAsBigInt("SECP_P")
			if err != nil {
				return err
			}

			valueBig := ComputeYCoordinate(slopeBig, xBig, new_xBig, yBig, secPBig)

			return ctx.ScopeManager.AssignVariables(map[string]any{"value": valueBig})
		},
	}
}

func createEcDoubleAssignNewYV1Hinter() (hinter.Hinter, error) {
	return newEcDoubleAssignNewYV1Hint(), nil
}

// ComputeSlopeV1 hint computes the slope between two points on an elliptic curve
//
// `newComputeSlopeV1Hint` takes 2 operanders as arguments
//   - `point0` is the first point on an elliptic curve to operate on
//   - `point1` is the second point on an elliptic curve to operate on
//
// `newComputeSlopeV1Hint` assigns the `slope` result as `value` in the current scope
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

			point0XAddr, err := point0.GetAddress(vm)
			if err != nil {
				return err
			}

			point1XAddr, err := point1.GetAddress(vm)
			if err != nil {
				return err
			}

			point0YAddr, err := point0XAddr.AddOffset(3)
			if err != nil {
				return err
			}

			point1YAddr, err := point1XAddr.AddOffset(3)
			if err != nil {
				return err
			}

			point0XValues, err := vm.Memory.ResolveAsBigInt3(point0XAddr)
			if err != nil {
				return err
			}

			point1XValues, err := vm.Memory.ResolveAsBigInt3(point1XAddr)
			if err != nil {
				return err
			}

			point0YValues, err := vm.Memory.ResolveAsBigInt3(point0YAddr)
			if err != nil {
				return err
			}

			point1YValues, err := vm.Memory.ResolveAsBigInt3(point1YAddr)
			if err != nil {
				return err
			}

			//> x0 = pack(ids.point0.x, PRIME)
			x0Big, err := hintrunner_utils.SecPPacked(point0XValues)
			if err != nil {
				return err
			}

			//> x1 = pack(ids.point1.x, PRIME)
			x1Big, err := hintrunner_utils.SecPPacked(point1XValues)
			if err != nil {
				return err
			}

			//> y0 = pack(ids.point0.y, PRIME)
			y0Big, err := hintrunner_utils.SecPPacked(point0YValues)
			if err != nil {
				return err
			}

			//> y1 = pack(ids.point0.y, PRIME)
			y1Big, err := hintrunner_utils.SecPPacked(point1YValues)
			if err != nil {
				return err
			}

			secPBig, ok := hintrunner_utils.GetSecPBig()
			if !ok {
				return fmt.Errorf("GetSecPBig failed")
			}

			// value = slope = line_slope(point1=(x0, y0), point2=(x1, y1), p=SECP_P)
			slopeBig, err := hintrunner_utils.LineSlope(&x0Big, &y0Big, &x1Big, &y1Big, &secPBig)
			if err != nil {
				return err
			}

			value := new(big.Int).Set(&slopeBig)

			return ctx.ScopeManager.AssignVariables(map[string]any{"value": value})
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

func newEcMulInnerHint(scalar hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "EcMulInner",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> memory[ap] = (ids.scalar % PRIME) % 2

			scalarFelt, err := hinter.ResolveAsFelt(vm, scalar)
			if err != nil {
				return err
			}
			scalarBytes := scalarFelt.Bytes()

			resultUint256 := new(uint256.Int).SetBytes(scalarBytes[:])
			resultUint256.Mod(resultUint256, uint256.NewInt(2))
			resultFelt := new(fp.Element).SetBytes(resultUint256.Bytes())
			resultMv := mem.MemoryValueFromFieldElement(resultFelt)
			apAddr := vm.Context.AddressAp()
			return vm.Memory.WriteToAddress(&apAddr, &resultMv)
		},
	}
}

func createEcMulInnerHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	scalar, err := resolver.GetResOperander("scalar")
	if err != nil {
		return nil, err
	}

	return newEcMulInnerHint(scalar), nil
}

// IsZeroNondet hint computes whether a value is zero or not
//
// `newIsZeroNondetHint` doesn't take any operander as argument
//
// `newIsZeroNondetHint` writes to `[ap]` the result of the comparison
// i.e, 1 if `x == 0`, 0 otherwise
func newIsZeroNondetHint() hinter.Hinter {
	return &GenericZeroHinter{
		Name: "IsZeroConditional",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> x == 0

			x, err := ctx.ScopeManager.GetVariableValueAsBigInt("x")
			if err != nil {
				return err
			}

			apAddr := vm.Context.AddressAp()

			var v mem.MemoryValue

			if x.Cmp(big.NewInt(0)) == 0 {
				v = mem.MemoryValueFromFieldElement(&utils.FeltOne)
			} else {
				v = mem.MemoryValueFromFieldElement(&utils.FeltZero)
			}

			return vm.Memory.WriteToAddress(&apAddr, &v)
		},
	}
}

func createIsZeroNondetHinter() (hinter.Hinter, error) {
	return newIsZeroNondetHint(), nil
}

// IsZeroPack hint computes packed value modulo SECP_P prime
//
// `newIsZeroPackHint` takes 1 operander as argument
//   - `x` is the value that will be packed and taken modulo SECP_P prime
//
// `newIsZeroPackHint` assigns the result as `x` in the current scope
func newIsZeroPackHint(x hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "IsZeroPack",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack
			//> x = pack(ids.x, PRIME) % SECP_P

			xAddr, err := x.GetAddress(vm)
			if err != nil {
				return err
			}

			xValues, err := vm.Memory.ResolveAsBigInt3(xAddr)
			if err != nil {
				return err
			}

			secPBig, ok := hintrunner_utils.GetSecPBig()
			if !ok {
				return fmt.Errorf("GetSecPBig failed")
			}

			xPackedBig, err := hintrunner_utils.SecPPacked(xValues)
			if err != nil {
				return err
			}

			value := new(big.Int)
			value.Mod(&xPackedBig, &secPBig)

			if err := ctx.ScopeManager.AssignVariable("x", value); err != nil {
				return err
			}

			return nil
		},
	}
}

func createIsZeroPackHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	x, err := resolver.GetResOperander("x")
	if err != nil {
		return nil, err
	}

	return newIsZeroPackHint(x), nil
}

// IsZeroDivMod hint computes the division modulo SECP_P prime for a given packed value
//
// `newIsZeroDivModHint` doesn't take any operander as argument
//
// `newIsZeroDivModHint` assigns the result as `value` in the current scope
func newIsZeroDivModHint() hinter.Hinter {
	return &GenericZeroHinter{
		Name: "IsZeroDivMod",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_secp.secp_utils import SECP_P
			//> from starkware.python.math_utils import div_mod
			//> value = x_inv = div_mod(1, x, SECP_P)

			secPBig, ok := hintrunner_utils.GetSecPBig()
			if !ok {
				return fmt.Errorf("GetSecPBig failed")
			}

			x, err := ctx.ScopeManager.GetVariableValueAsBigInt("x")
			if err != nil {
				return err
			}

			resBig, err := hintrunner_utils.Divmod(big.NewInt(1), x, &secPBig)
			if err != nil {
				return err
			}

			return ctx.ScopeManager.AssignVariable("value", &resBig)
		},
	}
}

func createIsZeroDivModHinter() (hinter.Hinter, error) {
	return newIsZeroDivModHint(), nil
}

func newRandomEcPointHint(p, m, q, s hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "RandomEcPoint",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.crypto.signature.signature import ALPHA, BETA, FIELD_PRIME
			//> from starkware.python.math_utils import random_ec_point
			//> from starkware.python.utils import to_bytes
			//>
			//> # Define a seed for random_ec_point that's dependent on all the input, so that:
			//> #   (1) The added point s is deterministic.
			//> #   (2) It's hard to choose inputs for which the builtin will fail.
			//> seed = b"".join(map(to_bytes, [ids.p.x, ids.p.y, ids.m, ids.q.x, ids.q.y]))
			//> ids.s.x, ids.s.y = random_ec_point(FIELD_PRIME, ALPHA, BETA, seed)

			pAddr, err := p.GetAddress(vm)
			if err != nil {
				return err
			}
			pValues, err := vm.Memory.ResolveAsEcPoint(pAddr)
			if err != nil {
				return err
			}
			mFelt, err := hinter.ResolveAsFelt(vm, m)
			if err != nil {
				return err
			}
			qAddr, err := q.GetAddress(vm)
			if err != nil {
				return err
			}
			qValues, err := vm.Memory.ResolveAsEcPoint(qAddr)
			if err != nil {
				return err
			}

			var bytesArray []byte
			writeFeltToBytesArray := func(n *fp.Element) {
				for _, byteValue := range n.Bytes() {
					bytesArray = append(bytesArray, byteValue)
				}
			}
			for _, felt := range pValues {
				writeFeltToBytesArray(felt)
			}
			writeFeltToBytesArray(mFelt)
			for _, felt := range qValues {
				writeFeltToBytesArray(felt)
			}
			seed := sha256.Sum256(bytesArray)

			alphaBig := new(big.Int)
			utils.Alpha.BigInt(alphaBig)
			betaBig := new(big.Int)
			utils.Beta.BigInt(betaBig)
			fieldPrime, ok := hintrunner_utils.GetCairoPrime()

			for i := uint64(0); i < 100; i++ {
				iBytes := make([]byte, 10)
				binary.LittleEndian.PutUint64(iBytes, i)
				concatenated := append(seed[1:], iBytes...)
				hash := sha256.Sum256(concatenated)
				hashHex := hex.EncodeToString(hash[:])
				x := new(big.Int)
				x.SetString(hashHex, 16)

				yCoef := big.NewInt(1)
				if seed[0]&1 == 1 {
					yCoef.Neg(yCoef)
				}

				// Try to recover y
				if !ok {
					return fmt.Errorf("failed to get field prime value")
				}
				if y, err := hintrunner_utils.RecoverY(x, alphaBig, betaBig, &fieldPrime); err == nil {
					y.Mul(yCoef, y)
					y.Mod(y, &fieldPrime)

					sAddr, err := s.GetAddress(vm)
					if err != nil {
						return err
					}

					sXFelt := new(fp.Element).SetBigInt(x)
					sYFelt := new(fp.Element).SetBigInt(y)
					fmt.Println(sXFelt.String())
					fmt.Println(sYFelt.String())
					sXMv := mem.MemoryValueFromFieldElement(sXFelt)
					sYMv := mem.MemoryValueFromFieldElement(sYFelt)

					err = vm.Memory.WriteToNthStructField(sAddr, sXMv, 0)
					if err != nil {
						return err
					}
					return vm.Memory.WriteToNthStructField(sAddr, sYMv, 1)
				}
			}

			return fmt.Errorf("could not find a point on the curve")
		},
	}
}

func createRandomEcPointHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	p, err := resolver.GetResOperander("p")
	if err != nil {
		return nil, err
	}
	m, err := resolver.GetResOperander("m")
	if err != nil {
		return nil, err
	}
	q, err := resolver.GetResOperander("q")
	if err != nil {
		return nil, err
	}
	s, err := resolver.GetResOperander("s")
	if err != nil {
		return nil, err
	}

	return newRandomEcPointHint(p, m, q, s), nil
}
