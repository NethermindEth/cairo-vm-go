package utils
package uint256

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func getBaseBig() uint256.Int {
	// 2**86
	return uint256.Int{
		0,
		0x400000,
		0,
		0,
	}
}

func GetSecPBig() uint256.Int {
	// 2**256 - 2**32 - 2**9 - 2**8 - 2**7 - 2**6 - 2**4 - 1
	return uint256.Int{
		0xFFFFFFFFFFFFFFFF,
		0xFFFFFFFFFFFFFFFF,
		0xFFFFFFFFFFFFFFFF,
		0xFFFFFFFEFFFFFC2F,
	}}

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

func GetBetaBig() uint256.Int {
	return uint256.NewInt(7)
}

func SecPSplit(num *big.Int) ([]*big.Int, error) {
	// https://github.com/starkware-libs/cairo-lang/blob/efa9648f57568aad8f8a13fbf027d2de7c63c2c0/src/starkware/cairo/common/cairo_secp/secp_utils.py#L14

	split := make([]*big.Int, 3)

	baseBig, ok := getBaseBig()
	if !ok {
		return nil, fmt.Errorf("getBaseBig failed")
	}

	var residue big.Int
	for i := 0; i < 3; i++ {
		num.DivMod(num, baseBig, &residue)
		split[i] = new(big.Int).Set(&residue)
	}

	if num.Cmp(big.NewInt(0)) != 0 {
		return nil, fmt.Errorf("num != 0")
	}

	return split, nil
}

func GetSecp256R1_P() uint256.Int {
	// 2**256 - 2**224 + 2**192 + 2**96 - 1
	return uint256.Int{
		0xFFFFFFFF00000001,
		0x0000000000000000,
		0x00000000FFFFFFFF,
		0xFFFFFFFFFFFFFFFF,
	}
}
