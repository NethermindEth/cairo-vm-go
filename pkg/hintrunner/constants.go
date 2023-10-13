package hintrunner

import "github.com/holiman/uint256"

func MaxU128() uint256.Int {
	return uint256.Int{18446744073709551615, 18446744073709551615, 0, 0}
}
