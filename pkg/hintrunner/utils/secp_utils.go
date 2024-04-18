package utils

import (
	"fmt"

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

	baseUint256 := getBaseUint256()

	packedUint256 := uint256.NewInt(0)
	for idx, limb := range limbs {
		limbUint256, _ := uint256.FromBig(AsInt(limb))
		valueToAddUint256 := uint256.NewInt(0).Exp(&baseUint256, uint256.NewInt(uint64(int64(idx))))
		valueToAddUint256.Mul(valueToAddUint256, limbUint256)
		packedUint256.Add(packedUint256, valueToAddUint256)
	}

	return *packedUint256, nil
}

func GetBetaUint256() uint256.Int {
	return *uint256.NewInt(7)
}

func SecPSplit(num *uint256.Int) ([]uint256.Int, error) {
	// https://github.com/starkware-libs/cairo-lang/blob/efa9648f57568aad8f8a13fbf027d2de7c63c2c0/src/starkware/cairo/common/cairo_secp/secp_utils.py#L14

	split := make([]uint256.Int, 3)

	baseUint256 := getBaseUint256()

	for i := 0; i < 3; i++ {
		var residue uint256.Int
		num.DivMod(num, &baseUint256, &residue)
		item := uint256.NewInt(0).Set(&residue)
		split[i] = *item
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
