package utils

import "math/big"

func asInt(value *big.Int, prime *big.Int) *big.Int {
	// https://github.com/starkware-libs/cairo-lang/blob/efa9648f57568aad8f8a13fbf027d2de7c63c2c0/src/starkware/cairo/common/math_utils.py#L8
	
	asIntBig := new(big.Int)
	primeBy2 := new(big.Int).Div(prime, big.NewInt(2))
	if value.Cmp(primeBy2) != -1 {
		asIntBig.Sub(value, prime)
	} else {
		asIntBig.Set(value)
	}
	return asIntBig
}
