package utils

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/holiman/uint256"
)

func getBaseUint256() uint256.Int {
	// 2**86
	return uint256.Int{
		0,
		0x400000,
		0,
		0,
	}
}

func GetSecPUint256() uint256.Int {
	// 2**256 - 2**32 - 2**9 - 2**8 - 2**7 - 2**6 - 2**4 - 1
	return uint256.Int{
		0xFFFFFFFEFFFFFC2F,
		0xFFFFFFFFFFFFFFFF,
		0xFFFFFFFFFFFFFFFF,
		0xFFFFFFFFFFFFFFFF,
	}
}

func SecPPacked(limbs [3]*fp.Element) (uint256.Int, error) {
	// https://github.com/starkware-libs/cairo-lang/blob/efa9648f57568aad8f8a13fbf027d2de7c63c2c0/src/starkware/cairo/common/cairo_secp/secp_utils.py#L28

	base := getBaseUint256()

	packed := uint256.NewInt(0)
	for idx, limb := range limbs {
		limbBytes := limb.Bytes()
		limbUint256 := new(uint256.Int).SetBytes(limbBytes[:])
		valueToAdd := uint256.NewInt(0).Exp(&base, uint256.NewInt(uint64(int64(idx))))
		valueToAdd.Mul(valueToAdd, limbUint256)
		packed.Add(packed, valueToAdd)
	}

	return *packed, nil
}

func GetBetaUint256() uint256.Int {
	return uint256.Int{
		0x7,
		0,
		0,
		0,
	}
}

func GetNBig() big.Int {
	// https://github.com/starkware-libs/cairo-lang/blob/efa9648f57568aad8f8a13fbf027d2de7c63c2c0/src/starkware/cairo/common/cairo_secp/secp_utils.py#L9

	NBig, _ := new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141", 16)
	return *NBig
}

func SecPSplit(num *uint256.Int) ([]uint256.Int, error) {
	// https://github.com/starkware-libs/cairo-lang/blob/efa9648f57568aad8f8a13fbf027d2de7c63c2c0/src/starkware/cairo/common/cairo_secp/secp_utils.py#L14

	split := make([]uint256.Int, 3)

	baseUint256 := getBaseUint256()

	for i := 0; i < 3; i++ {
		var residue uint256.Int
		num.DivMod(num, &baseUint256, &residue)
		split[i] = residue
	}

	if num.Cmp(uint256.NewInt(0)) != 0 {
		return nil, fmt.Errorf("num != 0")
	}

	return split, nil
}

func GetSecp256R1_P() uint256.Int {
	// 2**256 - 2**224 + 2**192 + 2**96 - 1
	return uint256.Int{
		0xFFFFFFFFFFFFFFFF,
		0x00000000FFFFFFFF,
		0x0000000000000000,
		0xFFFFFFFF00000001,
	}
}
