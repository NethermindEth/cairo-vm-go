package builtins

import (
	"fmt"
	"math/big"

	"golang.org/x/crypto/sha3"
)

const (
	KECCAK_FULL_RATE_IN_U64S  = 17
	KECCAK_FULL_RATE_IN_BYTES = 136
	BYTES_IN_U64_WORD         = 8
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
func Cairo_Keccak(data []byte) ([]byte, error) {
	hash, err := Keccak256(data)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

func AddPadding(input []uint64, lastInputWord uint64, lastInputNumBytes int) ([]uint64, error) {
	wordsDivisor := KECCAK_FULL_RATE_IN_U64S
	lastBlockNumFullWords := len(input) % wordsDivisor

	var firstWordToAppend, firstPaddingBytePart, r uint64
	if lastInputNumBytes == 0 {
		firstWordToAppend = 1
	} else {
		switch lastInputNumBytes {
		case 1:
			firstPaddingBytePart = 0x100
		case 2:
			firstPaddingBytePart = 0x10000
		case 3:
			firstPaddingBytePart = 0x1000000
		case 4:
			firstPaddingBytePart = 0x100000000
		case 5:
			firstPaddingBytePart = 0x10000000000
		case 6:
			firstPaddingBytePart = 0x1000000000000
		case 7:
			firstPaddingBytePart = 0x100000000000000
		default:
			return nil, fmt.Errorf("keccak last input word >7b")
		}
		r = lastInputWord % firstPaddingBytePart
		firstWordToAppend = firstPaddingBytePart + r
	}

	//	// Debug print statements:
	fmt.Printf("firstPaddingBytePart: %x\n", firstPaddingBytePart)
	fmt.Printf("r: %x\n", r)
	fmt.Printf("firstWordToAppend: %x\n", firstWordToAppend)

	input = append(input, firstWordToAppend) // Moved outside the if-else block

	if lastBlockNumFullWords == KECCAK_FULL_RATE_IN_U64S-1 {
		input = append(input, 0x8000000000000000+firstWordToAppend)
		return input, nil
	}

	finalizePadding := func(input []uint64, numPaddingWords uint32) []uint64 {
		for i := uint32(0); i < numPaddingWords; i++ {
			if i == numPaddingWords-1 {
				input = append(input, 0x8000000000000000)
			} else {
				input = append(input, 0)
			}
		}
		return input
	}
	//Debug when error
	fmt.Println("Before finalizePadding:", input)
	input = finalizePadding(input, uint32(KECCAK_FULL_RATE_IN_U64S-1-lastBlockNumFullWords))
	fmt.Println("After finalizePadding:", input)
	return input, nil
}
