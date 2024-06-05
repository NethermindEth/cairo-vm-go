package zero

import (
	"fmt"
	"math/big"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	secp_utils "github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/utils"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/builtins"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

// VerifyZero hint verifies that a packed value is zero modulo the SECP256R1 prime
// and stores in memory the quotient of the modular divison of the packed value by
// SECP256R1 prime
//
// `newVerifyZeroHint` takes 2 operanders as arguments
//   - `value` is the value that will be verified
//   - `q` is the variable that will store the quotient of the modular division
//
// `newVerifyZeroHint` writes the quotient of the modular division of the packed value
// by SECP256R1 prime to the memory address corresponding to `q`
func newVerifyZeroHint(val, q hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "VerifyZero",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack
			//> q, r = divmod(pack(ids.val, PRIME), SECP_P)
			//> assert r == 0, f"verify_zero: Invalid input {ids.val.d0, ids.val.d1, ids.val.d2}."
			//> ids.q = q % PRIME

			//> from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack

			valAddr, err := val.GetAddress(vm)
			if err != nil {
				return err
			}

			valValues, err := vm.Memory.ResolveAsBigInt3(valAddr)
			if err != nil {
				return err
			}

			secPBig, ok := secp_utils.GetSecPBig()
			if !ok {
				return fmt.Errorf("GetSecPBig failed")
			}

			//> q, r = divmod(pack(ids.val, PRIME), SECP_P)
			packedValue, err := secp_utils.SecPPacked(valValues)
			if err != nil {
				return err
			}
			qBig, rBig := new(big.Int), new(big.Int)
			qBig.DivMod(&packedValue, &secPBig, rBig)

			//> assert r == 0, f"verify_zero: Invalid input {ids.val.d0, ids.val.d1, ids.val.d2}."
			if rBig.Cmp(big.NewInt(0)) != 0 {
				return fmt.Errorf("verify_zero: Invalid input (%v, %v, %v)", valValues[0], valValues[1], valValues[2])
			}

			//> ids.q = q % PRIME
			qBig.Mod(qBig, fp.Modulus())
			qFelt := new(fp.Element).SetBigInt(qBig)
			qAddr, err := q.GetAddress(vm)
			if err != nil {
				return err
			}
			qMv := mem.MemoryValueFromFieldElement(qFelt)
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

// VerifyECDSASignature hint writes an ECDSA signature to a given address
//
// `newVerifyECDSASignatureHint` takes 3 operanders as arguments
//   - `ecdsaPtr` is the pointer variable that stores the address
//     where to write the signature
//   - `signature_r` and `signature_s` are the r and s parts of the signature
//
// `newVerifyECDSASignatureHint` uses the ECDSA builtin to perform this operation
func newVerifyECDSASignatureHint(ecdsaPtr, signature_r, signature_s hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "VerifyECDSASignature",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> ecdsa_builtin.add_signature(ids.ecdsa_ptr.address_, (ids.signature_r, ids.signature_s))
			ecdsaPtrAddr, err := hinter.ResolveAsAddress(vm, ecdsaPtr)
			if err != nil {
				return err
			}
			signature_rFelt, err := hinter.ResolveAsFelt(vm, signature_r)
			if err != nil {
				return err
			}
			signature_sFelt, err := hinter.ResolveAsFelt(vm, signature_s)
			if err != nil {
				return err
			}
			ECDSA_segment, ok := vm.Memory.FindSegmentWithBuiltin(builtins.ECDSAName)
			if !ok {
				return fmt.Errorf("ECDSA segment not found")
			}
			ECDSA_builtinRunner := (ECDSA_segment.BuiltinRunner).(*builtins.ECDSA)
			return ECDSA_builtinRunner.AddSignature(ecdsaPtrAddr.Offset, signature_rFelt, signature_sFelt)
		},
	}
}

func createVerifyECDSASignatureHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	ecdsaPtr, err := resolver.GetResOperander("ecdsa_ptr")
	if err != nil {
		return nil, err
	}

	signature_r, err := resolver.GetResOperander("signature_r")
	if err != nil {
		return nil, err
	}

	signature_s, err := resolver.GetResOperander("signature_s")
	if err != nil {
		return nil, err
	}

	return newVerifyECDSASignatureHint(ecdsaPtr, signature_r, signature_s), nil
}

// GetPointFromX hint calculates the y-coordinate of a point
// on an elliptic curve for a given x-coordinate
//
// `newGetPointFromXHint` takes 2 operanders as arguments
//   - `xCube` is the x-coordinate value used to calculate the point
//   - `v` is the parity of the `y` result, it should be either 0 or 1
//
// `newGetPointFromXHint` assigns the y-coordinate as `value` in the current scope
func newGetPointFromXHint(xCube, v hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "GetPointFromX",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack
			//> x_cube_int = pack(ids.x_cube, PRIME) % SECP_P
			//> y_square_int = (x_cube_int + ids.BETA) % SECP_P
			//> y = pow(y_square_int, (SECP_P + 1) // 4, SECP_P)
			//> if ids.v % 2 == y % 2:
			//>	 value = y
			//> else:
			//>	 value = (-y) % SECP_P
			xCubeAddr, err := xCube.GetAddress(vm)
			if err != nil {
				return err
			}

			v, err := hinter.ResolveAsFelt(vm, v)
			if err != nil {
				return err
			}

			//> from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack
			secpBig, _ := secp_utils.GetSecPBig()

			xCubeValues, err := vm.Memory.ResolveAsBigInt3(xCubeAddr)
			if err != nil {
				return err
			}

			xCubeIntBig, err := secp_utils.SecPPacked(xCubeValues)
			if err != nil {
				return err
			}
			xCubeIntBig.Mod(&xCubeIntBig, &secpBig)

			//> y_square_int = (x_cube_int + ids.BETA) % SECP_P
			betaBig := secp_utils.GetBetaBig()
			ySquareIntBig := new(big.Int).Add(&xCubeIntBig, &betaBig)
			ySquareIntBig.Mod(ySquareIntBig, &secpBig)

			//> y = pow(y_square_int, (SECP_P + 1) // 4, SECP_P)
			exponent := new(big.Int).Div(new(big.Int).Add(&secpBig, big.NewInt(1)), big.NewInt(4))
			y := new(big.Int).Exp(ySquareIntBig, exponent, &secpBig)
			vBig := v.BigInt(new(big.Int))

			//> if ids.v % 2 == y % 2:
			//>	 value = y
			//> else:
			//>	 value = (-y) % SECP_P
			value := new(big.Int)
			if vBig.Bit(0) == y.Bit(0) {
				value.Set(y)
			} else {
				value.Mod(value.Neg(y), &secpBig)
			}
			return ctx.ScopeManager.AssignVariable("value", value)
		},
	}
}

func createGetPointFromXHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	xCube, err := resolver.GetResOperander("x_cube")
	if err != nil {
		return nil, err
	}

	v, err := resolver.GetResOperander("v")
	if err != nil {
		return nil, err
	}

	return newGetPointFromXHint(xCube, v), nil
}

// ImportSecp256R1P hint imports the `SECP_P` constant from SECP256R1
// curve utilities in the current scope
//
// `newImportSecp256R1PHint` doesn't take any operander as argument
//
// `newGetPointFromXHint` assigns `SECP_P` variable in the current scope
func newImportSecp256R1PHint() hinter.Hinter {
	return &GenericZeroHinter{
		Name: "ImportSecp256R1P",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_secp.secp256r1_utils import SECP256R1_P as SECP_P
			SECP256R1_PBig, ok := secp_utils.GetSecp256R1_P()
			if !ok {
				return fmt.Errorf("SECP256R1_P failed")
			}

			return ctx.ScopeManager.AssignVariable("SECP_P", &SECP256R1_PBig)
		},
	}
}

func createImportSecp256R1PHinter() (hinter.Hinter, error) {
	return newImportSecp256R1PHint(), nil
}

// DivModSafeDiv hint computes a safe division in the context of the N constant
//
// `newDivModSafeDivHint` doesn't take any operander as argument
//
// `newDivModSafeDivHint` assigns the result `k` as `value` in the current scope
func newDivModSafeDivHint() hinter.Hinter {
	return &GenericZeroHinter{
		Name: "DivModSafeDiv",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> value = k = safe_div(res * b - a, N)

			res, err := ctx.ScopeManager.GetVariableValueAsBigInt("res")
			if err != nil {
				return err
			}

			a, err := ctx.ScopeManager.GetVariableValueAsBigInt("a")
			if err != nil {
				return err
			}

			b, err := ctx.ScopeManager.GetVariableValueAsBigInt("b")
			if err != nil {
				return err
			}

			N, err := ctx.ScopeManager.GetVariableValueAsBigInt("N")
			if err != nil {
				return err
			}

			divisor := new(big.Int).Sub(new(big.Int).Mul(res, b), a)
			value, err := secp_utils.SafeDiv(divisor, N)
			if err != nil {
				return err
			}

			return ctx.ScopeManager.AssignVariable("value", &value)
		},
	}
}

func createDivModSafeDivHinter() (hinter.Hinter, error) {
	return newDivModSafeDivHint(), nil
}

// DivModNPackedDivmodV1 hint caculates the division modulo N for packed values
//
// `newDivModNPackedDivmodV1Hint` takes 2 operanders as arguments
//   - `a` is the packed value that will be divided
//   - `b` is the packed value that will divide `a`
//
// `newDivModNPackedDivmodV1Hint` assigns the result `res` as `value` in the current scope
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

			bAddr, err := b.GetAddress(vm)
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

			nBig, ok := secp_utils.GetN()
			if !ok {
				return fmt.Errorf("GetN failed")
			}

			//> value = res = div_mod(a, b, N)
			resBig, err := secp_utils.Divmod(&aPackedBig, &bPackedBig, &nBig)
			if err != nil {
				return err
			}
			valueBig := new(big.Int).Set(&resBig)

			return ctx.ScopeManager.AssignVariable("value", valueBig)
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
