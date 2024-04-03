package utils

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func AsInt(valueFelt *fp.Element) *big.Int {
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
