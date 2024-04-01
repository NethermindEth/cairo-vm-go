package utils

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"math/big"
	"math/rand"

	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func RandomFeltElement(rand *rand.Rand) f.Element {
	b := [32]byte{}
	binary.BigEndian.PutUint64(b[24:32], rand.Uint64())
	binary.BigEndian.PutUint64(b[16:24], rand.Uint64())
	binary.BigEndian.PutUint64(b[8:16], rand.Uint64())
	//Limit to 59 bits so at max we have a 251 bit number
	binary.BigEndian.PutUint64(b[0:8], rand.Uint64()>>5)
	f, _ := f.BigEndian.Element(&b)
	return f
}

func RandomFeltElementU128(rand *rand.Rand) f.Element {
	b := [32]byte{}
	binary.BigEndian.PutUint64(b[24:32], rand.Uint64())
	binary.BigEndian.PutUint64(b[16:24], rand.Uint64())
	f, _ := f.BigEndian.Element(&b)
	return f
}

func DefaultRandGenerator() *rand.Rand {
	return rand.New(rand.NewSource(0))
}

//
// Math Utils
//

// https://github.com/starkware-libs/cairo-lang/blob/efa9648f57568aad8f8a13fbf027d2de7c63c2c0/src/starkware/python/math_utils.py#L26
func div_mod(n, m, p *big.Int) (*big.Int, error) {
	a, _, c := igcdex(m, p)
	if c.Cmp(big.NewInt(1)) != 0 {
		return nil, errors.New("no solution exists (gcd(m, p) != 1)")
	}
	res := new(big.Int)
	res.Mul(n, a)
	res.Mod(res, p)
	return res, nil
}

// https://github.com/sympy/sympy/blob/d91b8ad6d36a59a879cc70e5f4b379da5fdd46ce/sympy/core/intfunc.py#L362
func igcdex(a, b *big.Int) (*big.Int, *big.Int, *big.Int) {
	if a.Sign() == 0 && b.Sign() == 0 {
		return big.NewInt(0), big.NewInt(1), big.NewInt(0)
	}
	g, x, y := gcdext(a, b)
	return x, y, g
}

// https://github.com/sympy/sympy/blob/d91b8ad6d36a59a879cc70e5f4b379da5fdd46ce/sympy/external/ntheory.py#L119
func sign(n *big.Int) (int, *big.Int) {
	if n.Sign() < 0 {
		return -1, new(big.Int).Abs(n)
	}
	return 1, new(big.Int).Set(n)
}

// https://github.com/sympy/sympy/blob/d91b8ad6d36a59a879cc70e5f4b379da5fdd46ce/sympy/external/ntheory.py#L125
func gcdext(a, b *big.Int) (*big.Int, *big.Int, *big.Int) {
	if a.Sign() == 0 || b.Sign() == 0 {
		g := new(big.Int)
		if a.Sign() == 0 {
			g.Abs(b)
		} else {
			g.Abs(a)
		}

		if g.Sign() == 0 {
			return big.NewInt(0), big.NewInt(0), big.NewInt(0)
		}
		return g, new(big.Int).Div(a, g), new(big.Int).Div(b, g)
	}

	xSign, a := sign(a)
	ySign, b := sign(b)
	x, r := big.NewInt(1), big.NewInt(0)
	y, s := big.NewInt(0), big.NewInt(1)

	for b.Sign() != 0 {
		q, c := new(big.Int).DivMod(a, b, new(big.Int))
		a, b = b, c
		x, r = r, new(big.Int).Sub(x, new(big.Int).Mul(q, r))
		y, s = s, new(big.Int).Sub(y, new(big.Int).Mul(q, s))
	}

	return a, new(big.Int).Mul(x, big.NewInt(int64(xSign))), new(big.Int).Mul(y, big.NewInt(int64(ySign)))
}

//
// EC Utils
//

// https://github.com/starkware-libs/cairo-lang/blob/efa9648f57568aad8f8a13fbf027d2de7c63c2c0/src/starkware/python/math_utils.py#L151
func EcDoubleSlope(pointX, pointY, alpha, p *big.Int) (*big.Int, error) {
	if new(big.Int).Mod(pointY, p).Sign() == 0 {
		return nil, errors.New("point[1] % p == 0")
	}

	n := big.NewInt(3)
	n.Mul(n, pointX)
	n.Mul(n, pointX)
	n.Add(n, alpha)

	m := big.NewInt(2)
	m.Mul(m, pointY)

	return div_mod(n, m, p)
}

// https://github.com/starkware-libs/cairo-lang/blob/efa9648f57568aad8f8a13fbf027d2de7c63c2c0/src/starkware/cairo/common/math_utils.py#L8
func asInt(value *big.Int, prime *big.Int) *big.Int {
	asIntBig := new(big.Int)
	primeBy2 := new(big.Int).Div(prime, big.NewInt(2))
	if value.Cmp(primeBy2) != -1 {
		asIntBig.Sub(value, prime)
	} else {
		asIntBig.Set(value)
	}
	return asIntBig
}

// https://github.com/starkware-libs/cairo-lang/blob/efa9648f57568aad8f8a13fbf027d2de7c63c2c0/src/starkware/cairo/common/cairo_secp/secp_utils.py#L28
func SecPPacked(d0, d1, d2, prime *big.Int) (*big.Int, error) {
	baseBig, ok := utils.GetEcBaseBig()
	if !ok {
		return nil, fmt.Errorf("GetEcBaseBig failed")
	}

	packedBig := new(big.Int)
	for idx, dBig := range []*big.Int{d0, d1, d2} {
		dBig = asInt(dBig, prime)
		valueToAddBig := new(big.Int).Exp(baseBig, big.NewInt(int64(idx)), nil)
		valueToAddBig.Mul(valueToAddBig, dBig)
		packedBig.Add(packedBig, valueToAddBig)
	}

	return packedBig, nil
}
