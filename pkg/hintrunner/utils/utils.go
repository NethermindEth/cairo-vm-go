package utils

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func RandomFeltElement(rand *rand.Rand) f.Element {
	b := [32]byte{}
	binary.BigEndian.PutUint64(b[24:32], rand.Uint64())
	binary.BigEndian.PutUint64(b[16:24], rand.Uint64())
	binary.BigEndian.PutUint64(b[8:16], rand.Uint64())
	//Limit to 59 bits so at max we have a 251 bit number
	binary.BigEndian.PutUint64(b[0:8], rand.Uint64()>>5)
	f, _ := f.BigEndian.Element(&b)
	return f
}

func RandomFeltElementU128(rand *rand.Rand) f.Element {
	b := [32]byte{}
	binary.BigEndian.PutUint64(b[24:32], rand.Uint64())
	binary.BigEndian.PutUint64(b[16:24], rand.Uint64())
	f, _ := f.BigEndian.Element(&b)
	return f
}

func DefaultRandGenerator() *rand.Rand {
	return rand.New(rand.NewSource(0))
}

func ToSafeUint32(mv *memory.MemoryValue) (uint32, error) {
	valueUint64, err := mv.Uint64()
	if err != nil {
		return 0, err
	}
	if valueUint64 > math.MaxUint32 {
		return 0, fmt.Errorf("value out of range")
	}
	return uint32(valueUint64), nil
}
