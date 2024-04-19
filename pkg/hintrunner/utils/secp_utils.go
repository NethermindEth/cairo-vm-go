package utils

import (
	"fmt"

	"math/big"

	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func getBaseBig() (big.Int, bool) {
	// 2**86
	base, ok := new(big.Int).SetString("77371252455336267181195264", 10)
	return *base, ok
}

func GetSecPBig() (big.Int, bool) {
	// 2**256 - 2**32 - 2**9 - 2**8 - 2**7 - 2**6 - 2**4 - 1
	secP, ok := new(big.Int).SetString("115792089237316195423570985008687907853269984665640564039457584007908834671663", 10)
	return *secP, ok
}

func SecPPacked(limbs [3]*fp.Element) (big.Int, error) {
	// https://github.com/starkware-libs/cairo-lang/blob/efa9648f57568aad8f8a13fbf027d2de7c63c2c0/src/starkware/cairo/common/cairo_secp/secp_utils.py#L28

	baseBig, ok := getBaseBig()
	if !ok {
		return *big.NewInt(0), fmt.Errorf("getBaseBig failed")
	}

	packedBig := new(big.Int)
	for idx, limb := range limbs {
		limbBig := AsInt(limb)
		valueToAddBig := new(big.Int).Exp(&baseBig, big.NewInt(int64(idx)), nil)
		valueToAddBig.Mul(valueToAddBig, &limbBig)
		packedBig.Add(packedBig, valueToAddBig)
	}

	return *packedBig, nil
}

func GetBetaBig() big.Int {
	return *big.NewInt(7)
}

func GetNBig() big.Int {
	// https://github.com/starkware-libs/cairo-lang/blob/efa9648f57568aad8f8a13fbf027d2de7c63c2c0/src/starkware/cairo/common/cairo_secp/secp_utils.py#L9

	NBig, _ := new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141", 16)
	return *NBig
}

func SecPSplit(num *big.Int) ([]big.Int, error) {
	// https://github.com/starkware-libs/cairo-lang/blob/efa9648f57568aad8f8a13fbf027d2de7c63c2c0/src/starkware/cairo/common/cairo_secp/secp_utils.py#L14

	split := make([]big.Int, 3)

	baseBig, ok := getBaseBig()
	if !ok {
		return nil, fmt.Errorf("getBaseBig failed")
	}

	var residue big.Int
	for i := 0; i < 3; i++ {
		num.DivMod(num, &baseBig, &residue)
		splitVal := new(big.Int).Set(&residue)
		split[i] = *splitVal
	}

	if num.Cmp(big.NewInt(0)) != 0 {
		return nil, fmt.Errorf("num != 0")
	}

	return split, nil
}

func GetSecp256R1_P() (big.Int, bool) {
	// 2**256 - 2**224 + 2**192 + 2**96 - 1
	secp256R1_P, ok := new(big.Int).SetString("115792089210356248762697446949407573530086143415290314195533631308867097853951", 10)
	return *secp256R1_P, ok
}
