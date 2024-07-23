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
)

func newInvModPUint512Hint(x, xInverseModP hinter.ResOperander) hinter.Hinter {
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
			pack512 := func(lolow, loHigh, hiLow, hiHigh *fp.Element, numBitsShift int) big.Int {
				var loLowBig, loHighBig, hiLowBig, hiHighBig big.Int
				lolow.BigInt(&loLowBig)
				loHigh.BigInt(&loHighBig)
				hiLow.BigInt(&hiLowBig)
				hiHigh.BigInt(&hiHighBig)

				return *new(big.Int).Add(new(big.Int).Lsh(&hiHighBig, uint(numBitsShift)), &loLowBig).Add(new(big.Int).Lsh(&hiLowBig, uint(numBitsShift)), &loHighBig)
			}
			pack := func(low, high *fp.Element, numBitsShift int) big.Int {
				var lowBig, highBig big.Int
				low.BigInt(&lowBig)
				high.BigInt(&highBig)

				return *new(big.Int).Add(new(big.Int).Lsh(&highBig, uint(numBitsShift)), &lowBig)
			}

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

			x := pack512(xLoLow, xLoHigh, xHiLow, xHiHigh, 128)
			p := pack(pLow, pHigh, 128)

			xInverseModPBig := new(big.Int).Exp(&x, big.NewInt(-1), &p)

			split := func(num big.Int, numBitsShift uint16, length int) []fp.Element {
				a := make([]fp.Element, length)
				mask := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(numBitsShift)), big.NewInt(1))

				for i := 0; i < length; i++ {
					a[i] = *new(fp.Element).SetBigInt(new(big.Int).And(&num, mask))
					num.Rsh(&num, uint(numBitsShift))
				}

				return a
			}

			xInverseModPSplit := split(*xInverseModPBig, 128, 2)

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
