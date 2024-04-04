package utils

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func getBaseBig() (*big.Int, bool) {
	// 2**86
	return new(big.Int).SetString("77371252455336267181195264", 10)
}

func GetSecPBig() (*big.Int, bool) {
	// 2**256 - 2**32 - 2**9 - 2**8 - 2**7 - 2**6 - 2**4 - 1
	return new(big.Int).SetString("115792089237316195423570985008687907853269984665640564039457584007908834671663", 10)
}

func SecPPacked(limbs [3]*fp.Element) (*big.Int, error) {
	// https://github.com/starkware-libs/cairo-lang/blob/efa9648f57568aad8f8a13fbf027d2de7c63c2c0/src/starkware/cairo/common/cairo_secp/secp_utils.py#L28

	baseBig, ok := getBaseBig()
	if !ok {
		return nil, fmt.Errorf("getBaseBig failed")
	}

	packedBig := new(big.Int)
	for idx, limb := range limbs {
		limbBig := AsInt(limb)
		valueToAddBig := new(big.Int).Exp(baseBig, big.NewInt(int64(idx)), nil)
		valueToAddBig.Mul(valueToAddBig, limbBig)
		packedBig.Add(packedBig, valueToAddBig)
	}

	return packedBig, nil
}

func GetBetaBig() *big.Int {
	return big.NewInt(7)
}
