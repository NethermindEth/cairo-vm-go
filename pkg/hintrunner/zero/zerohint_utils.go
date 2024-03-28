package zero

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func AsInt(valueFelt fp.Element) big.Int {
	var valueBig big.Int
	valueFelt.BigInt(&valueBig)
	boundBig := new(big.Int).Div(fp.Modulus(), big.NewInt(2))

	// val if val < prime // 2 else val - prime
	if valueBig.Cmp(boundBig) == -1 {
		return valueBig
	}
	negativeValueBig := new(big.Int).Sub(&valueBig, fp.Modulus())
	return *negativeValueBig
}
