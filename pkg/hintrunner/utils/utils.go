package utils

import (
	"encoding/binary"
	"fmt"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"math/big"
	"math/rand"

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

//
// EC Utils
//

// def as_int(val, prime):
//
//	assert_integer(val)
//	return val if val < prime // 2 else val - prime
func AsInt(value *big.Int, prime *big.Int) *big.Int {
	asIntBig := new(big.Int)
	primeBy2 := new(big.Int).Div(prime, big.NewInt(2))
	if value.Cmp(primeBy2) != -1 {
		asIntBig.Sub(value, prime)
	} else {
		asIntBig.Set(value)
	}
	return asIntBig
}

// def pack(z, prime):
//
//	limbs = z.d0, z.d1, z.d2
//	return sum(as_int(limb, prime) * (BASE**i) for i, limb in enumerate(limbs))
func SecPPacked(d0, d1, d2, prime *big.Int) (*big.Int, error) {
	baseBig, ok := utils.GetEcBaseBig()
	if !ok {
		return nil, fmt.Errorf("GetEcBaseBig failed")
	}

	packedBig := new(big.Int)
	for idx, dBig := range []*big.Int{d0, d1, d2} {
		dBig = AsInt(dBig, prime)
		valueToAddBig := new(big.Int).Exp(baseBig, big.NewInt(int64(idx)), nil)
		valueToAddBig.Mul(valueToAddBig, dBig)
		packedBig.Add(packedBig, valueToAddBig)
	}

	return packedBig, nil
}
