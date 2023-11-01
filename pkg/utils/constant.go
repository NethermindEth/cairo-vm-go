package safemath

import (
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/holiman/uint256"
)

var FeltZero = fp.Element{}

var FeltOne = fp.Element{
	18446744073709551585, 18446744073709551615, 18446744073709551615, 576460752303422960,
}

// 1 << 128
var FeltMax128 = fp.Element{18446744073700081665, 17407, 18446744073709551584, 576460752142434320}

var Uint256Max128 = uint256.Int{18446744073709551615, 18446744073709551615, 0, 0}
