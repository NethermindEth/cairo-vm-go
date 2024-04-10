package utils

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func AsInt(valueFelt *fp.Element) *big.Int {
	// https://github.com/starkware-libs/cairo-lang/blob/efa9648f57568aad8f8a13fbf027d2de7c63c2c0/src/starkware/cairo/common/math_utils.py#L8

	var valueBig big.Int
	valueFelt.BigInt(&valueBig)
	return AsIntBig(&valueBig)
}

func AsIntBig(value *big.Int) *big.Int {
	boundBig := new(big.Int).Div(fp.Modulus(), big.NewInt(2))

	// val if val < prime // 2 else val - prime
	if value.Cmp(boundBig) == -1 {
		return value
	}
	return new(big.Int).Sub(value, fp.Modulus())
}
