package zero

import (
	"math/big"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

const (
	P_LOW  = "201385395114098847380338600778089168199"
	P_HIGH = "64323764613183177041862057485226039389"

	BITSHIFT = 128
)

// InvModPUint512 hint computes the inverse modulo a prime number `p` of 512 bits
// `newInvModPUint512Hint` takes 2 operanders as arguments
//   - `x` is the `uint512` variable that will be inverted modulo `p`
//   - `x_inverse_mod_p` is the variable that will store the result of the hint in memory
func newInvModPUint512Hint(x, xInverseModP hinter.Reference) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "InvModPUint512",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> def pack_512(u, num_bits_shift: int) -> int:
			//>     limbs = (u.d0, u.d1, u.d2, u.d3)
			//>     return sum(limb << (num_bits_shift * i) for i, limb in enumerate(limbs))
			//>
			//> x = pack_512(ids.x, num_bits_shift = 128)
			//> p = ids.p.low + (ids.p.high << 128)
			//> x_inverse_mod_p = pow(x,-1, p)
			//>
			//> x_inverse_mod_p_split = (x_inverse_mod_p & ((1 << 128) - 1), x_inverse_mod_p >> 128)
			//>
			//> ids.x_inverse_mod_p.low = x_inverse_mod_p_split[0]
			//> ids.x_inverse_mod_p.high = x_inverse_mod_p_split[1]

			xLoLow, xLoHigh, xHiLow, xHiHigh, err := GetUint512AsFelts(vm, x)
			if err != nil {
				return err
			}
			pLow, err := new(fp.Element).SetString(P_LOW)
			if err != nil {
				return err
			}
			pHigh, err := new(fp.Element).SetString(P_HIGH)
			if err != nil {
				return err
			}

			x := Pack(BITSHIFT, xLoLow, xLoHigh, xHiLow, xHiHigh)
			p := Pack(BITSHIFT, pLow, pHigh)

			xInverseModPBig := new(big.Int).Exp(&x, big.NewInt(-1), &p)

			// split big.Int into two fp.Elements
			xInverseModPSplit := make([]fp.Element, 2)
			mask := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(BITSHIFT)), big.NewInt(1))

			xInverseModPSplit[0] = *new(fp.Element).SetBigInt(new(big.Int).And(xInverseModPBig, mask))
			xInverseModPBig.Rsh(xInverseModPBig, uint(BITSHIFT))
			xInverseModPSplit[1] = *new(fp.Element).SetBigInt(xInverseModPBig)

			resAddr, err := xInverseModP.GetAddress(vm)
			if err != nil {
				return err
			}
			return vm.Memory.WriteUint256ToAddress(resAddr, &xInverseModPSplit[0], &xInverseModPSplit[1])
		},
	}
}

func createInvModPUint512Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	x, err := resolver.GetResOperander("x")
	if err != nil {
		return nil, err
	}

	xInverseModP, err := resolver.GetResOperander("x_inverse_mod_p")
	if err != nil {
		return nil, err
	}

	return newInvModPUint512Hint(x, xInverseModP), nil
}
