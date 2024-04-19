package utils

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func AsInt(valueFelt *fp.Element) big.Int {
	var valueBig big.Int
	valueFelt.BigInt(&valueBig)
	return AsIntBig(&valueBig)
}

func AsIntBig(value *big.Int) big.Int {
	boundBig := new(big.Int).Div(fp.Modulus(), big.NewInt(2))

	// val if val < prime // 2 else val - prime
	if value.Cmp(boundBig) == -1 {
		return *value
	}
	return *new(big.Int).Sub(value, fp.Modulus())
}

func SafeDiv(x, y *big.Int) (big.Int, error) {
	if y.Cmp(big.NewInt(0)) == 0 {
		return *big.NewInt(0), fmt.Errorf("Division by zero.")
	}
	if new(big.Int).Mod(x, y).Cmp(big.NewInt(0)) != 0 {
		return *big.NewInt(0), fmt.Errorf("%v is not divisible by %v.", x, y)
	}
	return *new(big.Int).Div(x, y), nil
}
