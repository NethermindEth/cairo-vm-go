package utils

import (
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/holiman/uint256"
)

//
// Felt Constants
//

var FeltZero = fp.Element{}

var FeltOne = fp.Element{
	18446744073709551585, 18446744073709551615, 18446744073709551615, 576460752303422960,
}

// 1 << 127
// same as 2 ** 127
var Felt127 = fp.Element{18446744073704816641, 8703, 18446744073709551600, 576460752222928912}

// 1 << 128
// same as 2 ** 128
var FeltMax128 = fp.Element{18446744073700081665, 17407, 18446744073709551584, 576460752142434320}

// 2 ** 250
var FeltUpperBound = fp.Element{0xfffffff5cdf80011, 0x4cc3fff, 0xfffffffffffdbe00, 0x7ffff52ad780230}

//
// Uint256 Constants
//

var Uint256Zero = uint256.Int{}

var Uint256One = uint256.Int{1, 0, 0, 0}

var Uint256Max128 = uint256.Int{18446744073709551615, 18446744073709551615, 0, 0}

// Alpha and Beta are paremeters required by the elliptic curve used by Cairo
// extracted from pedersen_params.json in https://github.com/starkware-libs/cairo-lang
var Alpha = fp.One()

var Beta = fp.Element([]uint64{
	3863487492851900874,
	7432612994240712710,
	12360725113329547591,
	88155977965380735,
})
