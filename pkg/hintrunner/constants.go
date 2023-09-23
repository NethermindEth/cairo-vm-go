package hintrunner

import (
	"math/big"

	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func MaxU128() *big.Int {
	return big.NewInt(0).SetBits([]big.Word{18446744073709551615, 18446744073709551615})
}

func MaxU128Felt() *f.Element {
	return &f.Element{
		18446744073700081697,
		17407,
		18446744073709551584,
		576460752142434864,
	}
}
