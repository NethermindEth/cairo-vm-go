package safemath

import (
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/holiman/uint256"
)

// Note: Using functions instead of `var` since they get inlined anyway

func FeltZero() fp.Element {
	return fp.Element{}
}

func FeltOne() fp.Element {
	return fp.Element{
		18446744073709551585, 18446744073709551615, 18446744073709551615, 576460752303422960,
	}
}

func FeltMax128() fp.Element {
	return fp.Element{
		18446744073700081665, 17407, 18446744073709551584, 576460752142434320,
	}
}

func Uint256Max128() uint256.Int {
	return uint256.Int{18446744073709551615, 18446744073709551615, 0, 0}
}
