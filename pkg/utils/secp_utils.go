package utils

import "math/big"

func GetSecp256R1_P() (*big.Int, bool) {
	return new(big.Int).SetString("115792089210356248762697446949407573530086143415290314195533631308867097853951", 10)
}
