package vm

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	f "github.com/NethermindEth/juno/core/felt"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"golang.org/x/exp/constraints"
)

type Program struct {
	Prime    *f.Felt
	Bytecode *[]string
}

type JSONContent struct {
	Prime           string
	CompilerVersion string
	Bytecode        []string
	// todo(rodro): Add remaining Json fields
	// Hints
	// EntryPointsByType
	// Contructors
	// L1 Headers
}

func ProgramFromFile(pathToFile string) (*Program, error) {
	content, error := os.ReadFile(pathToFile)
	if error != nil {
		return nil, error
	}
	return ProgramFromBytes(content)
}

func ProgramFromBytes(content []byte) (*Program, error) {
	var jsonContent JSONContent
	err := json.Unmarshal(content, &jsonContent)
	if err != nil {
		return nil, err
	}

    prime, err := parsePrime(jsonContent.Prime)
	if err != nil {
		return nil, err
	}

    return &Program{Prime: prime, Bytecode: &jsonContent.Bytecode}, nil
}

func parsePrime(prime string) (*f.Felt, error) {

	decodedPrime, err := hex.DecodeString(prime)
	if err != nil {
		return &f.Zero, err
	}

    wordAmount := len(decodedPrime) / 4

	words := make([]uint64, wordAmount)
	for i := 0; i < wordAmount; i++ {
		word_i, n := binary.Uvarint(decodedPrime[i*64 : max((i+1)*64, len(decodedPrime))])
		if n <= 0 {
			return &f.Zero, fmt.Errorf("Prime parsing: invalid number: %d", n)
		}
		words[i] = word_i
	}

    x := fp.Element(words)
    return f.NewFelt(&x), nil
}

func max[T constraints.Ordered](x, y T) T {
    if x > y {
        return x
    }
    return y
}
