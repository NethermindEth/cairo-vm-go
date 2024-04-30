package utils

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func EcDoubleSlope(pointX, pointY, alpha, prime *big.Int) (big.Int, error) {
	// https://github.com/starkware-libs/cairo-lang/blob/efa9648f57568aad8f8a13fbf027d2de7c63c2c0/src/starkware/python/math_utils.py#L151

	if new(big.Int).Mod(pointY, prime).Cmp(big.NewInt(0)) == 0 {
		return *big.NewInt(0), errors.New("point[1] % p == 0")
	}

	n := big.NewInt(3)
	n.Mul(n, pointX)
	n.Mul(n, pointX)
	n.Add(n, alpha)

	m := big.NewInt(2)
	m.Mul(m, pointY)

	return Divmod(n, m, prime)
}

func LineSlope(point_aX, point_aY, point_bX, point_bY, prime *big.Int) (big.Int, error) {
	// https://github.com/lambdaclass/cairo-vm_in_go/blob/31c3628bc10ebc1628685b3cdfa72d0e938b533e/pkg/builtins/ec_op.go#L258

	modValue := new(big.Int).Mod(new(big.Int).Sub(point_aX, point_aY), prime)

	if modValue.Cmp(big.NewInt(0)) == 0 {
		return *big.NewInt(0), errors.New("the slope of the line is invalid")
	}

	// Compute the difference of y-coordinates
	n := new(big.Int).Sub(point_bX, point_bY)

	// Compute the difference of x-coordinates
	m := new(big.Int).Sub(point_aX, point_aY)

	return Divmod(n, m, prime)
}

func AsInt(valueFelt *fp.Element) big.Int {
	// https://github.com/starkware-libs/cairo-lang/blob/efa9648f57568aad8f8a13fbf027d2de7c63c2c0/src/starkware/cairo/common/math_utils.py#L8

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

func Divmod(n, m, p *big.Int) (big.Int, error) {
	// https://github.com/starkware-libs/cairo-lang/blob/efa9648f57568aad8f8a13fbf027d2de7c63c2c0/src/starkware/python/math_utils.py#L26

	a, _, c := igcdex(m, p)
	if c.Cmp(big.NewInt(1)) != 0 {
		return *big.NewInt(0), errors.New("no solution exists (gcd(m, p) != 1)")
	}
	res := new(big.Int)
	res.Mul(n, &a)
	res.Mod(res, p)
	return *res, nil
}

func igcdex(a, b *big.Int) (big.Int, big.Int, big.Int) {
	// https://github.com/sympy/sympy/blob/d91b8ad6d36a59a879cc70e5f4b379da5fdd46ce/sympy/core/intfunc.py#L362

	if a.Cmp(big.NewInt(0)) == 0 && b.Cmp(big.NewInt(0)) == 0 {
		return *big.NewInt(0), *big.NewInt(1), *big.NewInt(0)
	}
	g, x, y := gcdext(a, b)
	return x, y, g
}

func gcdext(a, b *big.Int) (big.Int, big.Int, big.Int) {
	// https://github.com/sympy/sympy/blob/d91b8ad6d36a59a879cc70e5f4b379da5fdd46ce/sympy/external/ntheory.py#L125

	if a.Cmp(big.NewInt(0)) == 0 || b.Cmp(big.NewInt(0)) == 0 {
		g := new(big.Int)
		if a.Cmp(big.NewInt(0)) == 0 {
			g.Abs(b)
		} else {
			g.Abs(a)
		}

		if g.Cmp(big.NewInt(0)) == 0 {
			return *big.NewInt(0), *big.NewInt(0), *big.NewInt(0)
		}
		return *g, *new(big.Int).Div(a, g), *new(big.Int).Div(b, g)
	}

	xSign, aSigned := sign(a)
	ySign, bSigned := sign(b)
	x, r := big.NewInt(1), big.NewInt(0)
	y, s := big.NewInt(0), big.NewInt(1)

	for bSigned.Sign() != 0 {
		q, c := new(big.Int).DivMod(&aSigned, &bSigned, new(big.Int))
		aSigned = bSigned
		bSigned = *c
		x, r = r, new(big.Int).Sub(x, new(big.Int).Mul(q, r))
		y, s = s, new(big.Int).Sub(y, new(big.Int).Mul(q, s))
	}

	return aSigned, *new(big.Int).Mul(x, big.NewInt(int64(xSign))), *new(big.Int).Mul(y, big.NewInt(int64(ySign)))
}

func sign(n *big.Int) (int, big.Int) {
	// https://github.com/sympy/sympy/blob/d91b8ad6d36a59a879cc70e5f4b379da5fdd46ce/sympy/external/ntheory.py#L119

	if n.Sign() < 0 {
		return -1, *new(big.Int).Abs(n)
	}
	return 1, *new(big.Int).Set(n)
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
