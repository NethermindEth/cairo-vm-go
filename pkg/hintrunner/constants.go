package hintrunner

import (
	"github.com/holiman/uint256"

	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func MaxU128() *uint256.Int {
	return uint256.NewInt(0).SetBytes([]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255})
}

func MaxU128Felt() *f.Element {
	return &f.Element{
		18446744073700081697,
		17407,
		18446744073709551584,
		576460752142434864,
	}
}
