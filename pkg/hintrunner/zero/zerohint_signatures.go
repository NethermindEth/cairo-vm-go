package zero

import (
	"fmt"
	"math/big"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	secp_utils "github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/utils"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/builtins"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/holiman/uint256"
)

func newVerifyECDSASignatureHinter(ecdsaPtr, signature_r, signature_s hinter.ResOperander) hinter.Hinter {
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
	return newVerifyECDSASignatureHinter(ecdsaPtr, signature_r, signature_s), nil
}

func newGetPointFromXHinter(xCube, v hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "VerifyECDSASignature",
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

			xCubeMemoryValues, err := hinter.GetConsecutiveValues(vm, xCubeAddr, 3)
			if err != nil {
				return err
			}
			v, err := hinter.ResolveAsFelt(vm, v)
			if err != nil {
				return err
			}

			//> from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack
			secpUint256 := secp_utils.GetSecPUint256()

			//> x_cube_int = pack(ids.x_cube, PRIME) % SECP_P
			var xCubeValues [3]*fp.Element
			for i := 0; i < 3; i++ {
				xCubeValues[i], err = xCubeMemoryValues[i].FieldElement()
				if err != nil {
					return err
				}
			}
			xCubeUint256, err := secp_utils.SecPPacked(xCubeValues)
			if err != nil {
				return err
			}
			xCubeUint256.Mod(xCubeUint256, &secpUint256)

			//> y_square_int = (x_cube_int + ids.BETA) % SECP_P
			beta := secp_utils.GetBetaUint256()
			ySquareUint256 := uint256.NewInt(0).Add(xCubeUint256, &beta)
			ySquareUint256.Mod(ySquareUint256, &secpUint256)

			//> y = pow(y_square_int, (SECP_P + 1) // 4, SECP_P)
			exponent := uint256.NewInt(0).Div(uint256.NewInt(0).Add(&secpUint256, uint256.NewInt(1)), uint256.NewInt(4))
			yBig := new(big.Int).Exp(ySquareUint256.ToBig(), exponent.ToBig(), secpUint256.ToBig())
			vBig := v.BigInt(new(big.Int))
			secpBig := secpUint256.ToBig()

			//> if ids.v % 2 == y % 2:
			//>	 value = y
			//> else:
			//>	 value = (-y) % SECP_P
			value := new(big.Int)
			if vBig.Bit(0) == yBig.Bit(0) {
				value.Set(yBig)
			} else {
				value.Mod(value.Neg(yBig), secpBig)
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
	return newGetPointFromXHinter(xCube, v), nil
}

func newImportSecp256R1PHinter() hinter.Hinter {
	return &GenericZeroHinter{
		Name: "Secp256R1",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_secp.secp256r1_utils import SECP256R1_P as SECP_P
			SECP256R1_P := secp_utils.GetSecp256R1_P()
			SECP256R1_PBig := SECP256R1_P.ToBig()

			return ctx.ScopeManager.AssignVariable("SECP_P", SECP256R1_PBig)
		},
	}
}

func createImportSecp256R1PHinter() (hinter.Hinter, error) {
	return newImportSecp256R1PHinter(), nil
}
