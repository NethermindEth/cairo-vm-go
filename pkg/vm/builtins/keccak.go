package builtins

import (
	"fmt"
	"math/big"

	"golang.org/x/crypto/sha3"
)

// u128ToU64 converts a big.Int (representing a 128-bit integer) to a uint64,
// assuming the big.Int fits into a uint64.
func U128ToU64(input *big.Int) (uint64, error) {
	if input.BitLen() > 64 {
		return 0, fmt.Errorf("value too large to fit in uint64")
	}
	return input.Uint64(), nil
}

// u128Split splits a big.Int (representing a 128-bit integer) into two uint64s.
func U128Split(input *big.Int) (high, low uint64, err error) {
	divisor := new(big.Int)
	divisor.SetString("10000000000000000", 16) // Set the divisor using a hexadecimal string
	quotient, remainder := new(big.Int).DivMod(input, divisor, new(big.Int))
	high, err = U128ToU64(quotient)
	if err != nil {
		return 0, 0, err
	}
	low, err = U128ToU64(remainder)
	if err != nil {
		return 0, 0, err
	}
	return high, low, nil
}

// Keccak256 computes the Keccak-256 hash of the input data.
func Keccak256(data []byte) ([]byte, error) {
	hasher := sha3.NewLegacyKeccak256()
	if _, err := hasher.Write(data); err != nil {
		return nil, err // Return the error to the caller.
	}
	return hasher.Sum(nil), nil // Return the hash and a nil error.
}

// Cairo_Keccak computes and prints the Keccak-256 hash of the input data.
func Cairo_Keccak(data []byte) {
	hash, err := Keccak256(data)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("%x\n", hash)
}
