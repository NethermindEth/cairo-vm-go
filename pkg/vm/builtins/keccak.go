package builtins

import (
	"fmt"

	"golang.org/x/crypto/sha3"
)

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
