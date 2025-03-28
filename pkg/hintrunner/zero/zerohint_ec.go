package zero

import (
	"fmt"
	"math/big"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	secp_utils "github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/utils"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/holiman/uint256"
)

// GetHighLen hint calculates the highest bit length between `scalar_u.d2` and `scalar_v.d2`,
// subtracts 1 from the result, and assigns it to `ids.len_hi`
//
// `newGetHighLenHint` takes 3 operanders as arguments:
//   - `len_hi`: the variable that will store the result of the bit-length calculation
//   - `scalar_u.d2`: the first scalar value
//   - `scalar_v.d2`: the second scalar value
func newGetHighLenHint(len_hi, scalar_u, scalar_v hinter.Reference) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "GetHighLen",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> ids.len_hi = max(ids.scalar_u.d2.bit_length(), ids.scalar_v.d2.bit_length())-1

			scalarUAddr, err := scalar_u.Get(vm)
			if err != nil {
				return fmt.Errorf("failed to resolve scalar_u address: %w", err)
			}

			scalarVAddr, err := scalar_v.Get(vm)
			if err != nil {
				return fmt.Errorf("failed to resolve scalar_v address: %w", err)
			}

			scalarUValues, err := vm.Memory.ResolveAsBigInt3(scalarUAddr)
			if err != nil {
				return fmt.Errorf("failed to resolve scalar_u.d2: %w", err)
			}

			scalarVValues, err := vm.Memory.ResolveAsBigInt3(scalarVAddr)
			if err != nil {
				return fmt.Errorf("failed to resolve scalar_v.d2: %w", err)
			}

			var scalarUD2 big.Int
			_ = *scalarUValues[2].BigInt(&scalarUD2)

			var scalarVD2 big.Int
			_ = *scalarVValues[2].BigInt(&scalarVD2)

			bitLenU := scalarUD2.BitLen()
			bitLenV := scalarVD2.BitLen()

			lenHi := max(bitLenU, bitLenV) - 1

			lenHiAddr, err := len_hi.Get(vm)
			if err != nil {
				return fmt.Errorf("failed to get address of len_hi: %w", err)
			}

			lenHiMv := mem.MemoryValueFromInt(lenHi)

			return vm.Memory.WriteToAddress(&lenHiAddr, &lenHiMv)
		},
	}
}

func createGetHighLenHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	len_hi, err := resolver.GetReference("len_hi")
	if err != nil {
		return nil, err
	}

	scalar_u, err := resolver.GetReference("scalar_u")
	if err != nil {
		return nil, err
	}

	scalar_v, err := resolver.GetReference("scalar_v")
	if err != nil {
		return nil, err
	}

	return newGetHighLenHint(len_hi, scalar_u, scalar_v), nil
}

// BigIntToUint256 hint guesses the low part of the `x` uint256 variable
//
// `newBigIntToUint256Hint` takes 2 operanders as arguments
//   - `low` is the variable that will store the low part of the uint256 result
//   - `x` is the BigInt variable to convert to uint256
func newBigIntToUint256Hint(low, x hinter.Reference) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "BigIntToUint256",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> ids.low = (ids.x.d0 + ids.x.d1 * ids.BASE) & ((1 << 128) - 1)

			lowAddr, err := low.Get(vm)
			if err != nil {
				return err
			}

			xAddr, err := x.Get(vm)
			if err != nil {
				return err
			}

			xBigInt, err := vm.Memory.ResolveAsBigInt3(xAddr)
			if err != nil {
				return err
			}

			var xD0Big big.Int
			var xD1Big big.Int

			xBigInt[0].BigInt(&xD0Big)
			xBigInt[1].BigInt(&xD1Big)

			baseBig, ok := secp_utils.GetBaseBig()
			if !ok {
				return fmt.Errorf("getBaseBig failed")
			}

			var operand big.Int
			operand.Mul(&xD1Big, &baseBig)
			operand.Add(&operand, &xD0Big)

			mask := new(big.Int).Lsh(big.NewInt(1), 128)
			mask = new(big.Int).Sub(mask, big.NewInt(1))

			lowBigInt := new(big.Int).And(&operand, mask)
			lowValue := mem.MemoryValueFromFieldElement(new(fp.Element).SetBigInt(lowBigInt))

			err = vm.Memory.WriteToAddress(&lowAddr, &lowValue)
			if err != nil {
				return err
			}

			return nil
		},
	}
}

func createBigIntToUint256Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	low, err := resolver.GetReference("low")
	if err != nil {
		return nil, err
	}

	x, err := resolver.GetReference("x")
	if err != nil {
		return nil, err
	}

	return newBigIntToUint256Hint(low, x), nil
}

// EcNegate hint negates the y-coordinate of a point on an elliptic curve modulo SECP_P
//
// `newEcNegateHint` takes 1 operander as argument
//   - `point` is the point on an elliptic curve to operate on
//
// `newEcNegateHint` assigns the result as `value` in the current scope
func newEcNegateHint(point hinter.Reference) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "EcNegate",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack
			//>
			//> y = pack(ids.point.y, PRIME) % SECP_P
			//> # The modulo operation in python always returns a nonnegative number.
			//> value = (-y) % SECP_P

			secPBig, ok := secp_utils.GetSecPBig()
			if !ok {
				return fmt.Errorf("GetSecPBig failed")
			}

			pointAddr, err := point.Get(vm)
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
			yBig, err := secp_utils.SecPPacked(pointYValues)
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
	point, err := resolver.GetReference("point")
	if err != nil {
		return nil, err
	}

	return newEcNegateHint(point), nil
}

// DivModeNSafeDivPlusOne performs a safe division of the result obtained from
// the multiplication of `res` and `b` subtracted by `a`, by `N`. It then adds 1
// to the final result to ensure safety and prevent division by zero errors.
//
// `newDivModNSafeDivPlusOneHint` doens't take any operander as argument
//
// `DivModeNSafeDivPlusOne` assigns the result as `value` in the current scope.
func newDivModNSafeDivPlusOneHint() hinter.Hinter {
	return &GenericZeroHinter{
		Name: "DivModNSafeDivPlusOne",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> value = k_plus_one = safe_div(res * b - a, N) + 1
			valueBig := new(big.Int)

			resBig, err := hinter.GetVariableAs[*big.Int](&ctx.ScopeManager, "res")
			if err != nil {
				return err
			}

			aBig, err := hinter.GetVariableAs[*big.Int](&ctx.ScopeManager, "a")
			if err != nil {
				return err
			}

			bBig, err := hinter.GetVariableAs[*big.Int](&ctx.ScopeManager, "b")
			if err != nil {
				return err
			}

			nBig, err := hinter.GetVariableAs[*big.Int](&ctx.ScopeManager, "N")
			if err != nil {
				return err
			}

			valueBig.Mul(resBig, bBig)
			valueBig.Sub(valueBig, aBig)

			newValueBig, err := utils.SafeDiv(valueBig, nBig)
			if err != nil {
				return err
			}

			newValueBig.Add(&newValueBig, big.NewInt(1))
			return ctx.ScopeManager.AssignVariable("value", &newValueBig)
		},
	}
}

func createDivModNSafeDivPlusOneHinter() (hinter.Hinter, error) {
	return newDivModNSafeDivPlusOneHint(), nil
}

// DivModNPackedDivModExternalN computes the div_mod operation for a given packed value.
// `newDivModNPackedDivModExternalN` takes 2 operander as arguments
//   - `a` is the value that will be packed and taken prime
//   - `b` is the value that will be packed and taken prime
//
// `DivModNPackedDivModExternalN` assigns the result as `value` in the current scope.
func newDivModNPackedDivModExternalN(a, b hinter.Reference) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "DivModNPackedDivModExternalN",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_secp.secp_utils import pack
			//> from starkware.python.math_utils import div_mod, safe_div
			//> a = pack(ids.a, PRIME)
			//> b = pack(ids.b, PRIME)
			//> value = res = div_mod(a, b, N)

			aAddr, err := a.Get(vm)
			if err != nil {
				return err
			}

			aValues, err := vm.Memory.ResolveAsBigInt3(aAddr)
			if err != nil {
				return err
			}

			aBig, err := secp_utils.SecPPacked(aValues)
			if err != nil {
				return err
			}

			bAddr, err := b.Get(vm)
			if err != nil {
				return err
			}

			bValues, err := vm.Memory.ResolveAsBigInt3(bAddr)
			if err != nil {
				return err
			}

			bBig, err := secp_utils.SecPPacked(bValues)
			if err != nil {
				return err
			}

			nBig, err := hinter.GetVariableAs[*big.Int](&ctx.ScopeManager, "N")
			if err != nil {
				return err
			}

			newValueBig, err := secp_utils.Divmod(&aBig, &bBig, nBig)
			if err != nil {
				return err
			}

			resBig := new(big.Int).Set(&newValueBig)

			return ctx.ScopeManager.AssignVariables(map[string]any{"value": &newValueBig, "res": resBig, "a": &aBig, "b": &bBig})
		},
	}
}

func createDivModNPackedDivModExternalNHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	a, err := resolver.GetReference("a")
	if err != nil {
		return nil, err
	}

	b, err := resolver.GetReference("b")
	if err != nil {
		return nil, err
	}

	return newDivModNPackedDivModExternalN(a, b), nil
}

// NondetBigint3V1 hint writes a value to a specified segment of memory
//
// `newNondetBigint3V1Hint` takes 1 operander as argument
//   - `res` is the location in memory where to write the result
//
// `newNondetBigint3V1Hint` uses `SecPSplit` to split the value in 3 felts and writes the result in memory
func newNondetBigint3V1Hint(res hinter.Reference) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "NondetBigint3V1",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_secp.secp_utils import split
			//>
			//> segments.write_arg(ids.res.address_, split(value))

			address, err := res.Get(vm)
			if err != nil {
				return err
			}

			valueBig, err := hinter.GetVariableAs[*big.Int](&ctx.ScopeManager, "value")
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
	res, err := resolver.GetReference("res")
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
//
// There are 3 versions of FastEcAddAssignNewX hint
// V1 uses Secp256k1 curve
// V2 uses Curve25519 curve with SECP_P = 2**255 - 19
// V3 is similar to V1 but uses `pt0` and `pt1` for operanders where V1 and V2 use `point0` and `point1`
func newFastEcAddAssignNewXHint(slope, point0, point1 hinter.Reference, secPBig big.Int) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "FastEcAddAssignNewX",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			// V1
			//> from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack
			//>
			//> slope = pack(ids.slope, PRIME)
			//> x0 = pack(ids.point0.x, PRIME)
			//> x1 = pack(ids.point1.x, PRIME)
			//> y0 = pack(ids.point0.y, PRIME)
			//>
			//> value = new_x = (pow(slope, 2, SECP_P) - x0 - x1) % SECP_P

			// V2
			//> from starkware.cairo.common.cairo_secp.secp_utils import pack
			//> SECP_P = 2**255 - 19
			//>
			//> slope = pack(ids.slope, PRIME)
			//> x0 = pack(ids.point0.x, PRIME)
			//> x1 = pack(ids.point1.x, PRIME)
			//> y0 = pack(ids.point0.y, PRIME)
			//>
			//> value = new_x = (pow(slope, 2, SECP_P) - x0 - x1) % SECP_P

			slopeAddr, err := slope.Get(vm)
			if err != nil {
				return err
			}

			point0Addr, err := point0.Get(vm)
			if err != nil {
				return err
			}

			point1Addr, err := point1.Get(vm)
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
			new_xBig := new(big.Int)
			new_xBig.Exp(&slopeBig, big.NewInt(2), &secPBig)
			new_xBig.Sub(new_xBig, &x0Big)
			new_xBig.Sub(new_xBig, &x1Big)
			new_xBig.Mod(new_xBig, &secPBig)

			valueBig := new(big.Int)
			valueBig.Set(new_xBig)

			return ctx.ScopeManager.AssignVariables(map[string]any{"slope": &slopeBig, "x0": &x0Big, "y0": &y0Big, "new_x": new_xBig, "value": valueBig, "SECP_P": &secPBig})
		},
	}
}

func createFastEcAddAssignNewXHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	slope, err := resolver.GetReference("slope")
	if err != nil {
		return nil, err
	}

	point0, err := resolver.GetReference("point0")
	if err != nil {
		return nil, err
	}

	point1, err := resolver.GetReference("point1")
	if err != nil {
		return nil, err
	}

	secPBig, ok := secp_utils.GetSecPBig()
	if !ok {
		return nil, fmt.Errorf("GetSecPBig failed")
	}

	return newFastEcAddAssignNewXHint(slope, point0, point1, secPBig), nil
}

func createFastEcAddAssignNewXV2Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	slope, err := resolver.GetReference("slope")
	if err != nil {
		return nil, err
	}

	point0, err := resolver.GetReference("point0")
	if err != nil {
		return nil, err
	}

	point1, err := resolver.GetReference("point1")
	if err != nil {
		return nil, err
	}

	//> SECP_P = 2**255-19
	secPBig, ok := secp_utils.GetCurve25519PBig()
	if !ok {
		return nil, fmt.Errorf("GetSecPBig failed")
	}

	return newFastEcAddAssignNewXHint(slope, point0, point1, secPBig), nil
}

func createFastEcAddAssignNewXV3Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	slope, err := resolver.GetReference("slope")
	if err != nil {
		return nil, err
	}

	point0, err := resolver.GetReference("pt0")
	if err != nil {
		return nil, err
	}

	point1, err := resolver.GetReference("pt1")
	if err != nil {
		return nil, err
	}

	secPBig, ok := secp_utils.GetSecPBig()
	if !ok {
		return nil, fmt.Errorf("GetSecPBig failed")
	}

	return newFastEcAddAssignNewXHint(slope, point0, point1, secPBig), nil
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

			slopeBig, err := hinter.GetVariableAs[*big.Int](&ctx.ScopeManager, "slope")
			if err != nil {
				return err
			}

			x0Big, err := hinter.GetVariableAs[*big.Int](&ctx.ScopeManager, "x0")
			if err != nil {
				return err
			}

			new_xBig, err := hinter.GetVariableAs[*big.Int](&ctx.ScopeManager, "new_x")
			if err != nil {
				return err
			}

			y0Big, err := hinter.GetVariableAs[*big.Int](&ctx.ScopeManager, "y0")
			if err != nil {
				return err
			}

			secPBig, err := hinter.GetVariableAs[*big.Int](&ctx.ScopeManager, "SECP_P")
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
func newEcDoubleSlopeV1Hint(point hinter.Reference) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "EcDoubleSlopeV1",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack
			//> from starkware.python.math_utils import ec_double_slope
			//>
			//> # Compute the slope.
			//> x = pack(ids.point.x, PRIME)
			//> y = pack(ids.point.y, PRIME)
			//> value = slope = ec_double_slope(point=(x, y), alpha=0, p=SECP_P)

			pointAddr, err := point.Get(vm)
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
			if new(big.Int).Mod(&yBig, &secPBig).Cmp(big.NewInt(0)) == 0 {
				return fmt.Errorf("point[1] modulo p == 0")
			}

			valueBig, err := secp_utils.EcDoubleSlope(&xBig, &yBig, big.NewInt(0), &secPBig)
			if err != nil {
				return err
			}

			return ctx.ScopeManager.AssignVariables(map[string]any{"value": &valueBig})
		},
	}
}

func createEcDoubleSlopeV1Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	point, err := resolver.GetReference("point")
	if err != nil {
		return nil, err
	}

	return newEcDoubleSlopeV1Hint(point), nil
}

// EcDoubleSlopeV3 hint computes the slope for doubling a point on the elliptic curve
//
// `newEcDoubleSlopeV3Hint` takes 1 operander as argument
//   - `pt` is the point on an elliptic curve to operate on
//
// `newEcDoubleSlopeV3Hint` assigns the `slope` result as `value` in the current scope
// This version differs from EcDoubleSlopeV1 by the name of the operander (`point` for V1, `pt` for V3)
// and the computation of the slope : V1 uses a dedicated utility function with an additionnal check
// while V3 executes the modular division directly
func newEcDoubleSlopeV3Hint(point hinter.Reference) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "EcDoubleSlopeV3",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack
			//> from starkware.python.math_utils import div_mod
			//>
			//> # Compute the slope.
			//> x = pack(ids.pt.x, PRIME)
			//> y = pack(ids.pt.y, PRIME)
			//> value = slope = div_mod(3 * x ** 2, 2 * y, SECP_P)

			pointAddr, err := point.Get(vm)
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

			//> x = pack(ids.pt.x, PRIME)
			xBig, err := secp_utils.SecPPacked(pointXValues)
			if err != nil {
				return err
			}

			//> y = pack(ids.pt.y, PRIME)
			yBig, err := secp_utils.SecPPacked(pointYValues)
			if err != nil {
				return err
			}

			secPBig, ok := secp_utils.GetSecPBig()
			if !ok {
				return fmt.Errorf("GetSecPBig failed")
			}

			//> value = slope = div_mod(3 * x ** 2, 2 * y, SECP_P)
			valueBig, err := secp_utils.EcDoubleSlope(&xBig, &yBig, big.NewInt(0), &secPBig)
			if err != nil {
				return err
			}

			return ctx.ScopeManager.AssignVariables(map[string]any{"value": &valueBig})
		},
	}
}

func createEcDoubleSlopeV3Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	point, err := resolver.GetReference("pt")
	if err != nil {
		return nil, err
	}

	return newEcDoubleSlopeV3Hint(point), nil
}

// Reduce hint reduces a packed value modulo the SECP256K1 prime
//
// `newReduceHint` takes 1 operander as argument
//   - `x` is the packed value to be reduced
//
// `newReduceHint` assigns the result as `value` in the current scope
// This implementation is valid for ReduceV1 and ReduceV2
func newReduceHint(x hinter.Reference) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "Reduce",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack (V1) //> from starkware.cairo.common.cairo_secp.secp_utils import pack (V2)
			//>
			//> value = pack(ids.x, PRIME) % SECP_P

			secPBig, ok := secp_utils.GetSecPBig()
			if !ok {
				return fmt.Errorf("GetSecPBig failed")
			}

			xAddr, err := x.Get(vm)
			if err != nil {
				return err
			}

			xValues, err := vm.Memory.ResolveAsBigInt3(xAddr)
			if err != nil {
				return err
			}

			xBig, err := secp_utils.SecPPacked(xValues)
			if err != nil {
				return err
			}

			xBig.Mod(&xBig, &secPBig)
			valueBigIntPtr := new(big.Int).Set(&xBig)

			return ctx.ScopeManager.AssignVariable("value", valueBigIntPtr)
		},
	}
}

func createReduceHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	x, err := resolver.GetReference("x")
	if err != nil {
		return nil, err
	}

	return newReduceHint(x), nil
}

// ReduceEd25519 hint reduces a packed value modulo the Curve25519 prime
//
// `newReduceEd25519Hint` takes 1 operander as argument
//   - `x` is the packed value to be reduced
//
// `newReduceEd25519Hint` assigns the result as `value` in the current scope
func newReduceEd25519Hint(x hinter.Reference) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "ReduceEd25519",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_secp.secp_utils import pack
			//> SECP_P=2**255-19
			//>
			//> value = pack(ids.x, PRIME) % SECP_P

			secPBig, ok := secp_utils.GetCurve25519PBig()
			if !ok {
				return fmt.Errorf("GetSecPBig failed")
			}

			xAddr, err := x.Get(vm)
			if err != nil {
				return err
			}

			xValues, err := vm.Memory.ResolveAsBigInt3(xAddr)
			if err != nil {
				return err
			}

			xBig, err := secp_utils.SecPPacked(xValues)
			if err != nil {
				return err
			}

			xBig.Mod(&xBig, &secPBig)
			valueBigIntPtr := new(big.Int).Set(&xBig)

			return ctx.ScopeManager.AssignVariable("value", valueBigIntPtr)
		},
	}
}

func createReduceEd25519Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	x, err := resolver.GetReference("x")
	if err != nil {
		return nil, err
	}

	return newReduceEd25519Hint(x), nil
}

// EcDoubleAssignNewX hint computes a new x-coordinate for a point being doubled on an elliptic curve
//
// `newEcDoubleAssignNewXHint` takes 2 operanders as arguments
//   - `slope` is the slope for doubling a point, computed with EcDoubleSlopeV1 hint
//   - `point` is the point on an elliptic curve to operate on for V1, `pt` for V2
//
// `newEcDoubleAssignNewXHint` assigns the `new_x` result as `value` in the current scope
// It also assigns `slope`, `x`, `y` and `new_x` in the current scope
// so that they are available in the current scope for EcDoubleAssignNewYV1 hint
//
// This implementation is valid for EcDoubleAssignNewX V1,V2 and V4, only the operander differs
// with `point` used for V1,V2 and `pt` used for V4 and for V2 SECP_P has to be already in scope
// contrary to V1
func newEcDoubleAssignNewXHint(slope, point hinter.Reference) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "EcDoubleAssignNewX",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			// V1
			//> from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack
			//>
			//> slope = pack(ids.slope, PRIME)
			//> x = pack(ids.point.x, PRIME)
			//> y = pack(ids.point.y, PRIME)
			//>
			//> value = new_x = (pow(slope, 2, SECP_P) - 2 * x) % SECP_P

			// V2
			//> from starkware.cairo.common.cairo_secp.secp_utils import pack
			//>
			//> slope = pack(ids.slope, PRIME)
			//> x = pack(ids.point.x, PRIME)
			//> y = pack(ids.point.y, PRIME)
			//>
			//> value = new_x = (pow(slope, 2, SECP_P) - 2 * x) % SECP_P

			// V4
			//> from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack
			//>
			//> slope = pack(ids.slope, PRIME)
			//> x = pack(ids.pt.x, PRIME)
			//> y = pack(ids.pt.y, PRIME)
			//>
			//> value = new_x = (pow(slope, 2, SECP_P) - 2 * x) % SECP_P

			slopeAddr, err := slope.Get(vm)
			if err != nil {
				return err
			}

			pointAddr, err := point.Get(vm)
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

			return ctx.ScopeManager.AssignVariables(map[string]any{"slope": &slopeBig, "x": &xBig, "y": &yBig, "new_x": new_xBig, "value": valueBig, "SECP_P": &secPBig})
		},
	}
}

func createEcDoubleAssignNewXV1Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	slope, err := resolver.GetReference("slope")
	if err != nil {
		return nil, err
	}

	point, err := resolver.GetReference("point")
	if err != nil {
		return nil, err
	}

	return newEcDoubleAssignNewXHint(slope, point), nil
}

func createEcDoubleAssignNewXV4Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	slope, err := resolver.GetReference("slope")
	if err != nil {
		return nil, err
	}

	point, err := resolver.GetReference("pt")
	if err != nil {
		return nil, err
	}

	return newEcDoubleAssignNewXHint(slope, point), nil
}

func createEcDoubleAssignNewXV2Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	slope, err := resolver.GetReference("slope")
	if err != nil {
		return nil, err
	}

	point, err := resolver.GetReference("point")
	if err != nil {
		return nil, err
	}

	return newEcDoubleAssignNewXHint(slope, point), nil
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

			slopeBig, err := hinter.GetVariableAs[*big.Int](&ctx.ScopeManager, "slope")
			if err != nil {
				return err
			}

			xBig, err := hinter.GetVariableAs[*big.Int](&ctx.ScopeManager, "x")
			if err != nil {
				return err
			}

			new_xBig, err := hinter.GetVariableAs[*big.Int](&ctx.ScopeManager, "new_x")
			if err != nil {
				return err
			}

			yBig, err := hinter.GetVariableAs[*big.Int](&ctx.ScopeManager, "y")
			if err != nil {
				return err
			}

			secPBig, err := hinter.GetVariableAs[*big.Int](&ctx.ScopeManager, "SECP_P")
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

// ComputeSlopeV1 hint computes the slope between two points on the Secp256k1 elliptic curve
//
// `newComputeSlopeV1Hint` takes 2 operanders as arguments
//   - `point0` is the first point on an elliptic curve to operate on
//   - `point1` is the second point on an elliptic curve to operate on
//
// `newComputeSlopeV1Hint` assigns the `slope` result as `value` in the current scope
func newComputeSlopeV1Hint(point0, point1 hinter.Reference) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "ComputeSlopeV1",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack
			//> from starkware.python.math_utils import line_slope
			//>
			//> # Compute the slope.
			//> x0 = pack(ids.point0.x, PRIME)
			//> y0 = pack(ids.point0.y, PRIME)
			//> x1 = pack(ids.point1.x, PRIME)
			//> y1 = pack(ids.point1.y, PRIME)
			//> value = slope = line_slope(point1=(x0, y0), point2=(x1, y1), p=SECP_P)

			point0XAddr, err := point0.Get(vm)
			if err != nil {
				return err
			}

			point1XAddr, err := point1.Get(vm)
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

			modValue := new(big.Int).Mod(new(big.Int).Sub(&x0Big, &x1Big), &secPBig)

			if modValue.Cmp(big.NewInt(0)) == 0 {
				return fmt.Errorf("the slope of the line is invalid")
			}

			slopeBig, err := secp_utils.LineSlope(&x0Big, &y0Big, &x1Big, &y1Big, &secPBig)
			if err != nil {
				return err
			}

			value := new(big.Int).Set(&slopeBig)

			return ctx.ScopeManager.AssignVariables(map[string]any{"value": value})
		},
	}
}

func createComputeSlopeV1Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	point0, err := resolver.GetReference("point0")
	if err != nil {
		return nil, err
	}

	point1, err := resolver.GetReference("point1")
	if err != nil {
		return nil, err
	}

	return newComputeSlopeV1Hint(point0, point1), nil
}

// ComputeSlopeV2 hint computes the slope between two points on the Curve25519 curve
//
// `newComputeSlopeV2Hint` takes 2 operanders as arguments
//   - `point0` is the first point on an elliptic curve to operate on
//   - `point1` is the second point on an elliptic curve to operate on
//
// `newComputeSlopeV2Hint` assigns the `slope` result as `value` in the current scope
// // This version uses Curve25519 curve with SECP_P = 2**255 - 19
func newComputeSlopeV2Hint(point0, point1 hinter.Reference) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "ComputeSlopeV2",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.python.math_utils import line_slope
			//> from starkware.cairo.common.cairo_secp.secp_utils import pack
			//> SECP_P = 2**255-19
			//> # Compute the slope.
			//> x0 = pack(ids.point0.x, PRIME)
			//> y0 = pack(ids.point0.y, PRIME)
			//> x1 = pack(ids.point1.x, PRIME)
			//> y1 = pack(ids.point1.y, PRIME)
			//> value = slope = line_slope(point1=(x0, y0), point2=(x1, y1), p=SECP_P)

			point0XAddr, err := point0.Get(vm)
			if err != nil {
				return err
			}

			point1XAddr, err := point1.Get(vm)
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

			//> y1 = pack(ids.point1.y, PRIME)
			y1Big, err := secp_utils.SecPPacked(point1YValues)
			if err != nil {
				return err
			}

			//> SECP_P = 2**255-19
			secPBig, ok := secp_utils.GetCurve25519PBig()
			if !ok {
				return fmt.Errorf("GetSecPBig failed")
			}

			// value = slope = line_slope(point1=(x0, y0), point2=(x1, y1), p=SECP_P)

			modValue := new(big.Int).Mod(new(big.Int).Sub(&x0Big, &x1Big), &secPBig)

			if modValue.Cmp(big.NewInt(0)) == 0 {
				return fmt.Errorf("the slope of the line is invalid")
			}

			slopeBig, err := secp_utils.LineSlope(&x0Big, &y0Big, &x1Big, &y1Big, &secPBig)
			if err != nil {
				return err
			}

			value := new(big.Int).Set(&slopeBig)

			return ctx.ScopeManager.AssignVariables(map[string]any{"value": value})
		},
	}
}

func createComputeSlopeV2Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	point0, err := resolver.GetReference("point0")
	if err != nil {
		return nil, err
	}

	point1, err := resolver.GetReference("point1")
	if err != nil {
		return nil, err
	}

	return newComputeSlopeV2Hint(point0, point1), nil
}

// ComputeSlopeV3 hint computes the slope between two points on the Secp256k1 elliptic curve
//
// `newComputeSlopeV3Hint` takes 2 operanders as arguments
//   - `pt0` is the first point on an elliptic curve to operate on
//   - `pt1` is the second point on an elliptic curve to operate on
//
// `newComputeSlopeV3Hint` assigns the `slope` result as `value` in the current scope
//
// This version differs from ComputeSlopeV1 by the name of the operanders (`point0` and `point1` for V1, `pt0` and `pt1` for V3)
// and the computation of the slope : V1 uses a dedicated utility function with an additionnal check while V3 executes
// the modular division directly
func newComputeSlopeV3Hint(point0, point1 hinter.Reference) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "ComputeSlopeV3",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack
			//> from starkware.python.math_utils import div_mod
			//>
			//> # Compute the slope.
			//> x0 = pack(ids.pt0.x, PRIME)
			//> y0 = pack(ids.pt0.y, PRIME)
			//> x1 = pack(ids.pt1.x, PRIME)
			//> y1 = pack(ids.pt1.y, PRIME)
			//> value = slope = div_mod(y0 - y1, x0 - x1, SECP_P)

			point0XAddr, err := point0.Get(vm)
			if err != nil {
				return err
			}

			point1XAddr, err := point1.Get(vm)
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

			//> value = slope = div_mod(y0 - y1, x0 - x1, SECP_P)
			slopeBig, err := secp_utils.LineSlope(&x0Big, &y0Big, &x1Big, &y1Big, &secPBig)
			if err != nil {
				return err
			}

			return ctx.ScopeManager.AssignVariables(map[string]any{"value": &slopeBig})
		},
	}
}

func createComputeSlopeV3Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	point0, err := resolver.GetReference("pt0")
	if err != nil {
		return nil, err
	}

	point1, err := resolver.GetReference("pt1")
	if err != nil {
		return nil, err
	}

	return newComputeSlopeV3Hint(point0, point1), nil
}

func newEcMulInnerHint(scalar hinter.Reference) hinter.Hinter {
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
	scalar, err := resolver.GetReference("scalar")
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
		Name: "IsZeroNondet",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> python hint in cairo file: "x == 0"
			//> compiled file hint: "memory[ap] = to_felt_or_relocatable(x == 0)"

			x, err := hinter.GetVariableAs[*big.Int](&ctx.ScopeManager, "x")
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
func newIsZeroPackHint(x hinter.Reference) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "IsZeroPack",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack
			//>
			//> x = pack(ids.x, PRIME) % SECP_P

			xAddr, err := x.Get(vm)
			if err != nil {
				return err
			}

			xValues, err := vm.Memory.ResolveAsBigInt3(xAddr)
			if err != nil {
				return err
			}

			secPBig, ok := secp_utils.GetSecPBig()
			if !ok {
				return fmt.Errorf("GetSecPBig failed")
			}

			xPackedBig, err := secp_utils.SecPPacked(xValues)
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
	x, err := resolver.GetReference("x")
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
			//>
			//> value = x_inv = div_mod(1, x, SECP_P)

			secPBig, ok := secp_utils.GetSecPBig()
			if !ok {
				return fmt.Errorf("GetSecPBig failed")
			}

			x, err := hinter.GetVariableAs[*big.Int](&ctx.ScopeManager, "x")
			if err != nil {
				return err
			}

			resBig, err := secp_utils.Divmod(big.NewInt(1), x, &secPBig)
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

// RecoverY hint Recovers the y coordinate of a point on the elliptic curve
// y^2 = x^3 + alpha * x + beta (mod field_prime) of a given x coordinate.
//
// `newRecoverYHint` takes 2 operanders as arguments
//   - `x` is the x coordinate of an elliptic curve point
//   - `p` is one of the two EC points with the given x coordinate (x, y)
func newRecoverYHint(x, p hinter.Reference) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "RecoverY",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.crypto.signature.signature import ALPHA, BETA, FIELD_PRIME
			//> from starkware.python.math_utils import recover_y
			//> ids.p.x = ids.x
			//> # This raises an exception if `x` is not on the curve.
			//> ids.p.y = recover_y(ids.x, ALPHA, BETA, FIELD_PRIME)

			pXAddr, err := p.Get(vm)
			if err != nil {
				return err
			}

			pYAddr, err := pXAddr.AddOffset(1)
			if err != nil {
				return err
			}

			xFelt, err := hinter.ResolveAsFelt(vm, x)
			if err != nil {
				return err
			}

			valueX := mem.MemoryValueFromFieldElement(xFelt)

			err = vm.Memory.WriteToAddress(&pXAddr, &valueX)
			if err != nil {
				return err
			}

			betaBigInt := new(big.Int)
			utils.Beta.BigInt(betaBigInt)

			fieldPrimeBigInt, ok := secp_utils.GetCairoPrime()
			if !ok {
				return fmt.Errorf("GetCairoPrime failed")
			}

			xBigInt := new(big.Int)
			xFelt.BigInt(xBigInt)

			// y^2 = x^3 + alpha * x + beta (mod field_prime)
			resultBigInt, err := secp_utils.RecoverY(xBigInt, betaBigInt, &fieldPrimeBigInt)
			if err != nil {
				return err
			}
			resultFelt := new(fp.Element).SetBigInt(resultBigInt)
			resultMv := mem.MemoryValueFromFieldElement(resultFelt)
			return vm.Memory.WriteToAddress(&pYAddr, &resultMv)
		},
	}
}

func createRecoverYHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	x, err := resolver.GetReference("x")
	if err != nil {
		return nil, err
	}

	p, err := resolver.GetReference("p")
	if err != nil {
		return nil, err
	}

	return newRecoverYHint(x, p), nil
}

// RandomEcPoint hint returns a random non-zero point on the STARK curve
// y^2 = x^3 + alpha * x + beta (mod field_prime).
// The point is created deterministically from the seed.
//
// `newRandomEcPointHint` takes 4 operanders as arguments
//   - `p` is an EC point used for seed generation
//   - `m` the multiplication coefficient of Q used for seed generation
//   - `q` an EC point used for seed generation
//   - `s` is where the generated random EC point is written to
func newRandomEcPointHint(p, m, q, s hinter.Reference) hinter.Hinter {
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

			pAddr, err := p.Get(vm)
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

			qAddr, err := q.Get(vm)
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

			sAddr, err := s.Get(vm)
			if err != nil {
				return err
			}

			return secp_utils.RandomEcPoint(vm, bytesArray, sAddr)
		},
	}
}

func createRandomEcPointHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	p, err := resolver.GetReference("p")
	if err != nil {
		return nil, err
	}

	m, err := resolver.GetReference("m")
	if err != nil {
		return nil, err
	}

	q, err := resolver.GetReference("q")
	if err != nil {
		return nil, err
	}

	s, err := resolver.GetReference("s")
	if err != nil {
		return nil, err
	}

	return newRandomEcPointHint(p, m, q, s), nil
}

// ChainedEcOp hint returns a random non-zero point on the STARK curve
// in the context of chained ecop operations.
// The point is created deterministically from the seed.
//
// `newChainedEcOpHint` takes 5 operanders as arguments
//   - `len` is the number of chained elements in the chained ecop operation
//   - `p` is an EC point used for seed generation
//   - `m` the multiplication coefficient of Q used for seed generation
//   - `q` an EC point used for seed generation
//   - `s` is where the generated random EC point is written to
func newChainedEcOpHint(len, p, m, q, s hinter.Reference) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "ChainedEcOp",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.crypto.signature.signature import ALPHA, BETA, FIELD_PRIME
			//> from starkware.python.math_utils import random_ec_point
			//> from starkware.python.utils import to_bytes
			//>
			//> n_elms = ids.len
			//> assert isinstance(n_elms, int) and n_elms >= 0, \
			//> 	f'Invalid value for len. Got: {n_elms}.'
			//> if '__chained_ec_op_max_len' in globals():
			//> 	assert n_elms <= __chained_ec_op_max_len, \
			//> 		f'chained_ec_op() can only be used with len<={__chained_ec_op_max_len}. ' \
			//> 		f'Got: n_elms={n_elms}.'
			//>
			//> # Define a seed for random_ec_point that's dependent on all the input, so that:
			//> #   (1) The added point s is deterministic.
			//> #   (2) It's hard to choose inputs for which the builtin will fail.
			//> seed = b"".join(
			//> 	map(
			//> 		to_bytes,
			//> 		[
			//> 			ids.p.x,
			//> 			ids.p.y,
			//> 			*memory.get_range(ids.m, n_elms),
			//> 			*memory.get_range(ids.q.address_, 2 * n_elms),
			//> 		],
			//> 	)
			//> )
			//> ids.s.x, ids.s.y = random_ec_point(FIELD_PRIME, ALPHA, BETA, seed)

			nElms, err := hinter.ResolveAsUint64(vm, len)
			if err != nil {
				return err
			}

			if nElms == 0 {
				return fmt.Errorf("invalid value for len. Got: %v", nElms)
			}

			chainedEcOpMaxLen := uint64(1000)
			if nElms > chainedEcOpMaxLen {
				return fmt.Errorf("f'chained_ec_op() can only be used with len<=%d.\n Got: n_elms=%d", chainedEcOpMaxLen, nElms)
			}

			pAddr, err := p.Get(vm)
			if err != nil {
				return err
			}

			mAddr, err := hinter.ResolveAsAddress(vm, m)
			if err != nil {
				return err
			}

			qAddr, err := hinter.ResolveAsAddress(vm, q)
			if err != nil {
				return err
			}

			pValues, err := vm.Memory.ResolveAsEcPoint(pAddr)
			if err != nil {
				return err
			}

			firstRange, err := vm.Memory.GetConsecutiveMemoryValues(*mAddr, nElms)
			if err != nil {
				return err
			}

			secondRange, err := vm.Memory.GetConsecutiveMemoryValues(*qAddr, 2*nElms)
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

			for _, element := range firstRange {
				writeFeltToBytesArray(&element.Felt)
			}
			for _, element := range secondRange {
				writeFeltToBytesArray(&element.Felt)

			}

			sAddr, err := s.Get(vm)
			if err != nil {
				return err
			}

			return secp_utils.RandomEcPoint(vm, bytesArray, sAddr)
		},
	}
}

func createChainedEcOpHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	len, err := resolver.GetReference("len")
	if err != nil {
		return nil, err
	}

	p, err := resolver.GetReference("p")
	if err != nil {
		return nil, err
	}

	m, err := resolver.GetReference("m")
	if err != nil {
		return nil, err
	}

	q, err := resolver.GetReference("q")
	if err != nil {
		return nil, err
	}

	s, err := resolver.GetReference("s")
	if err != nil {
		return nil, err
	}

	return newChainedEcOpHint(len, p, m, q, s), nil
}

// EcRecoverDivModNPacked hint stores the value of div_mod(x, s, N) to scope.
//
// `newEcRecoverDivModNPackedHint` takes 3 operanders as arguments
//   - `n` is an EC point
//   - `x` is an EC point
//   - `s` is an EC point
func newEcRecoverDivModNPackedHint(n, x, s hinter.Reference) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "EcRecoverDivModNPacked",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_secp.secp_utils import pack
			//> from starkware.python.math_utils import div_mod, safe_div
			//>
			//> N = pack(ids.n, PRIME)
			//> x = pack(ids.x, PRIME) % N
			//> s = pack(ids.s, PRIME) % N
			//> value = res = div_mod(x, s, N)

			nAddr, err := n.Get(vm)
			if err != nil {
				return err
			}

			xAddr, err := x.Get(vm)
			if err != nil {
				return err
			}

			sAddr, err := s.Get(vm)
			if err != nil {
				return err
			}

			nValues, err := vm.Memory.ResolveAsBigInt3(nAddr)
			if err != nil {
				return err
			}

			xValues, err := vm.Memory.ResolveAsBigInt3(xAddr)
			if err != nil {
				return err
			}

			sValues, err := vm.Memory.ResolveAsBigInt3(sAddr)
			if err != nil {
				return err
			}

			//> N = pack(ids.n, PRIME)
			nPackedBig, err := secp_utils.SecPPacked(nValues)
			if err != nil {
				return err
			}

			//> x = pack(ids.x, PRIME) % N
			xPackedBig, err := secp_utils.SecPPacked(xValues)
			if err != nil {
				return err
			}
			xPackedBig.Mod(&xPackedBig, &nPackedBig)

			//> s = pack(ids.s, PRIME) % N
			sPackedBig, err := secp_utils.SecPPacked(sValues)
			if err != nil {
				return err
			}
			sPackedBig.Mod(&sPackedBig, &nPackedBig)

			//> value = res = div_mod(x, s, N)
			resBig, err := secp_utils.Divmod(&xPackedBig, &sPackedBig, &nPackedBig)
			if err != nil {
				return err
			}

			valueBig := new(big.Int)
			valueBig.Set(&resBig)

			return ctx.ScopeManager.AssignVariables(map[string]any{"res": &resBig, "value": valueBig})
		},
	}
}

func createEcRecoverDivModNPackedHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	n, err := resolver.GetReference("n")
	if err != nil {
		return nil, err
	}

	x, err := resolver.GetReference("x")
	if err != nil {
		return nil, err
	}

	s, err := resolver.GetReference("s")
	if err != nil {
		return nil, err
	}

	return newEcRecoverDivModNPackedHint(n, x, s), nil
}

// EcRecoverSubAB hint stores the value of a - b to scope.
//
// `newEcRecoverSubABHint` takes 2 operanders as arguments
//   - `a` is an EC point
//   - `b` is an EC point
func newEcRecoverSubABHint(a, b hinter.Reference) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "EcRecoverSubAB",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_secp.secp_utils import pack
			//> from starkware.python.math_utils import div_mod, safe_div
			//>
			//> a = pack(ids.a, PRIME)
			//> b = pack(ids.b, PRIME)
			//>
			//> value = res = a - b

			aAddr, err := a.Get(vm)
			if err != nil {
				return err
			}

			bAddr, err := b.Get(vm)
			if err != nil {
				return err
			}

			aValues, err := vm.Memory.ResolveAsBigInt3(aAddr)
			if err != nil {
				return err
			}

			bValues, err := vm.Memory.ResolveAsBigInt3(bAddr)
			if err != nil {
				return err
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

			//> value = res = a - b
			resBig := new(big.Int)
			resBig.Sub(&aPackedBig, &bPackedBig)
			valueBig := new(big.Int)
			valueBig.Set(resBig)

			return ctx.ScopeManager.AssignVariables(map[string]any{"res": resBig, "value": valueBig})
		},
	}
}

func createEcRecoverSubABHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	a, err := resolver.GetReference("a")
	if err != nil {
		return nil, err
	}

	b, err := resolver.GetReference("b")
	if err != nil {
		return nil, err
	}

	return newEcRecoverSubABHint(a, b), nil
}

// EcRecoverProductMod hint stores the value of (a * b) % m to scope.
//
// `newEcRecoverProductModHint` takes 3 operanders as arguments
//   - `a` is an EC point
//   - `b` is an EC point
//   - `m` is an EC point
func newEcRecoverProductModHint(a, b, m hinter.Reference) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "EcRecoverProductMod",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_secp.secp_utils import pack
			//> from starkware.python.math_utils import div_mod, safe_div
			//>
			//> a = pack(ids.a, PRIME)
			//> b = pack(ids.b, PRIME)
			//> product = a * b
			//> m = pack(ids.m, PRIME)
			//>
			//> value = res = product % m

			aAddr, err := a.Get(vm)
			if err != nil {
				return err
			}

			bAddr, err := b.Get(vm)
			if err != nil {
				return err
			}

			mAddr, err := m.Get(vm)
			if err != nil {
				return err
			}

			aValues, err := vm.Memory.ResolveAsBigInt3(aAddr)
			if err != nil {
				return err
			}

			bValues, err := vm.Memory.ResolveAsBigInt3(bAddr)
			if err != nil {
				return err
			}

			mValues, err := vm.Memory.ResolveAsBigInt3(mAddr)
			if err != nil {
				return err
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

			//> m = pack(ids.m, PRIME)
			mPackedBig, err := secp_utils.SecPPacked(mValues)
			if err != nil {
				return err
			}

			//> product = a * b
			productBig := new(big.Int)
			productBig.Mul(&aPackedBig, &bPackedBig)

			//> value = res = product % m
			resBig := new(big.Int)
			resBig.Mod(productBig, &mPackedBig)

			valueBig := new(big.Int)
			valueBig.Set(resBig)

			return ctx.ScopeManager.AssignVariables(map[string]any{"m": &mPackedBig, "product": productBig, "res": resBig, "value": valueBig})
		},
	}
}

func createEcRecoverProductModHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	a, err := resolver.GetReference("a")
	if err != nil {
		return nil, err
	}

	b, err := resolver.GetReference("b")
	if err != nil {
		return nil, err
	}

	m, err := resolver.GetReference("m")
	if err != nil {
		return nil, err
	}

	return newEcRecoverProductModHint(a, b, m), nil
}

// EcRecoverProductDivM hint fetches product and m scope variables
// and stores the result of their division in scope variables value and k
//
// `newEcRecoverProductDivMHint` takes no arguments
func newEcRecoverProductDivMHint() hinter.Hinter {
	return &GenericZeroHinter{
		Name: "EcRecoverProductDivM",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> value = k = product // m

			productBig, err := hinter.GetVariableAs[*big.Int](&ctx.ScopeManager, "product")
			if err != nil {
				return err
			}

			mBig, err := hinter.GetVariableAs[*big.Int](&ctx.ScopeManager, "m")
			if err != nil {
				return err
			}

			kBig := new(big.Int)
			kBig.Div(productBig, mBig)

			valueBig := new(big.Int)
			valueBig.Set(kBig)

			return ctx.ScopeManager.AssignVariables(map[string]any{"k": kBig, "value": valueBig})
		},
	}
}

func createEcRecoverProductDivMHinter() (hinter.Hinter, error) {
	return newEcRecoverProductDivMHint(), nil
}

// BigIntPackDivMod hint divides two values modulo a prime number
//
// `newBigIntPackDivModHint` takes 3 operanders as arguments
//   - `P` is the prime modulus
//   - `x` is the numerator
//   - `y` is the denominator
//
// `newBigIntPackDivModHint` assigns the result as `value` in the current scope
func newBigIntPackDivModHint(x, y, p hinter.Reference) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "BigIntPackDivMod",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_secp.secp_utils import pack
			//> from starkware.cairo.common.math_utils import as_int
			//> from starkware.python.math_utils import div_mod, safe_div
			//>
			//> p = pack(ids.P, PRIME)
			//> x = pack(ids.x, PRIME) + as_int(ids.x.d3, PRIME) * ids.BASE ** 3 + as_int(ids.x.d4, PRIME) * ids.BASE ** 4
			//> y = pack(ids.y, PRIME)
			//>
			//> value = res = div_mod(x, y, p)

			pAddr, err := p.Get(vm)
			if err != nil {
				return err
			}

			pValues, err := vm.Memory.ResolveAsBigInt3(pAddr)
			if err != nil {
				return err
			}

			xAddr, err := x.Get(vm)
			if err != nil {
				return err
			}

			xValues, err := vm.Memory.ResolveAsBigInt5(xAddr)
			if err != nil {
				return err
			}

			yAddr, err := y.Get(vm)
			if err != nil {
				return err
			}

			yValues, err := vm.Memory.ResolveAsBigInt3(yAddr)
			if err != nil {
				return err
			}

			var xD3Big big.Int
			var xD4Big big.Int

			xValues[3].BigInt(&xD3Big)
			xValues[4].BigInt(&xD4Big)

			base, ok := secp_utils.GetBaseBig()
			if !ok {
				return fmt.Errorf("getBaseBig failed")
			}

			//> p = pack(ids.P, PRIME)
			pPacked, err := secp_utils.SecPPacked(pValues)
			if err != nil {
				return err
			}

			//> x = pack(ids.x, PRIME) + as_int(ids.x.d3, PRIME) * ids.BASE ** 3 + as_int(ids.x.d4, PRIME) * ids.BASE ** 4
			xPacked, err := secp_utils.SecPPackedBigInt5(xValues)
			if err != nil {
				return err
			}

			base3Big := new(big.Int)

			base3Big.Exp(&base, big.NewInt(3), big.NewInt(0))

			base4Big := new(big.Int)

			base4Big.Exp(&base, big.NewInt(4), big.NewInt(0))

			xBig := new(big.Int)

			xBig.Mul(&xD3Big, base3Big)

			xBig.Add(xBig, xBig.Mul(&xD4Big, base4Big))

			xBig.Add(xBig, &xPacked)

			//> y = pack(ids.y, PRIME)
			yPacked, err := secp_utils.SecPPacked(yValues)
			if err != nil {
				return err
			}

			//> value = res = div_mod(x, y, p)
			res, err := secp_utils.Divmod(xBig, &yPacked, &pPacked)
			if err != nil {
				return err
			}

			var value = new(big.Int).Set(&res)

			return ctx.ScopeManager.AssignVariables(map[string]any{"value": value, "res": &res, "x": xBig, "y": &yPacked, "p": &pPacked})
		},
	}
}

func createBigIntPackDivModHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	x, err := resolver.GetReference("x")
	if err != nil {
		return nil, err
	}

	y, err := resolver.GetReference("y")
	if err != nil {
		return nil, err
	}

	p, err := resolver.GetReference("P")
	if err != nil {
		return nil, err
	}

	return newBigIntPackDivModHint(x, y, p), nil
}

// BigIntSafeDiv hint safely divides two numbers and assigns the result based on a condition
//
// `newBigIntSafeDivHint` does not take any arguments
//
// `newBigIntSafeDivHint` assigns the result as `value` and sets `flag` based on the result in the current scope
func newBigIntSafeDivHint(flag hinter.Reference) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "BigIntSafeDiv",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> k = safe_div(res * y - x, p)
			//> value = k if k > 0 else 0 - k
			//> ids.flag = 1 if k > 0 else 0

			flagAddr, err := flag.Get(vm)
			if err != nil {
				return err
			}

			x, err := hinter.GetVariableAs[*big.Int](&ctx.ScopeManager, "x")
			if err != nil {
				return err
			}

			y, err := hinter.GetVariableAs[*big.Int](&ctx.ScopeManager, "y")
			if err != nil {
				return err
			}

			p, err := hinter.GetVariableAs[*big.Int](&ctx.ScopeManager, "p")
			if err != nil {
				return err
			}

			res, err := hinter.GetVariableAs[*big.Int](&ctx.ScopeManager, "res")
			if err != nil {
				return err
			}

			//> k = safe_div(res * y - x, p)
			tmp := new(big.Int)
			tmp.Mul(res, y)
			tmp.Sub(tmp, x)
			k := new(big.Int).Div(tmp, p)

			//> value = k if k > 0 else 0 - k
			value := new(big.Int).Abs(k)

			//> ids.flag = 1 if k > 0 else 0
			flagBigInt := big.NewInt(0)
			if k.Sign() > 0 {
				flagBigInt.SetInt64(1)
			}

			flagValue := mem.MemoryValueFromFieldElement(new(fp.Element).SetBigInt(flagBigInt))

			err = ctx.ScopeManager.AssignVariable("value", value)
			if err != nil {
				return err
			}

			err = vm.Memory.WriteToAddress(&flagAddr, &flagValue)
			if err != nil {
				return err
			}

			return nil
		},
	}
}

func createBigIntSaveDivHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	flag, err := resolver.GetReference("flag")
	if err != nil {
		return nil, err
	}

	return newBigIntSafeDivHint(flag), nil
}
