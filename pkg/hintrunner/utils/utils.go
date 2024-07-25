package utils

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"math/rand"

	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func RandomFeltElement(rand *rand.Rand) fp.Element {
	b := [32]byte{}
	binary.BigEndian.PutUint64(b[24:32], rand.Uint64())
	binary.BigEndian.PutUint64(b[16:24], rand.Uint64())
	binary.BigEndian.PutUint64(b[8:16], rand.Uint64())
	//Limit to 59 bits so at max we have a 251 bit number
	binary.BigEndian.PutUint64(b[0:8], rand.Uint64()>>5)
	f, _ := fp.BigEndian.Element(&b)
	return f
}

func RandomFeltElementU128(rand *rand.Rand) fp.Element {
	b := [32]byte{}
	binary.BigEndian.PutUint64(b[24:32], rand.Uint64())
	binary.BigEndian.PutUint64(b[16:24], rand.Uint64())
	f, _ := fp.BigEndian.Element(&b)
	return f
}

func DefaultRandGenerator() *rand.Rand {
	return rand.New(rand.NewSource(0))
}

func ToSafeUint32(mv *mem.MemoryValue) (uint32, error) {
	valueUint64, err := mv.Uint64()
	if err != nil {
		return 0, err
	}
	if valueUint64 > math.MaxUint32 {
		return 0, fmt.Errorf("value out of range")
	}
	return uint32(valueUint64), nil
}

func RandomEcPoint(vm *VM.VirtualMachine, bytesArray []byte, sAddr mem.MemoryAddress) error {
	seed := sha256.Sum256(bytesArray)

	alphaBig := new(big.Int)
	utils.Alpha.BigInt(alphaBig)
	betaBig := new(big.Int)
	utils.Beta.BigInt(betaBig)
	fieldPrime, ok := GetCairoPrime()
	if !ok {
		return fmt.Errorf("GetCairoPrime failed")
	}

	for i := uint64(0); i < 100; i++ {
		iBytes := make([]byte, 10)
		binary.LittleEndian.PutUint64(iBytes, i)
		concatenated := append(seed[1:], iBytes...)
		hash := sha256.Sum256(concatenated)
		hashHex := hex.EncodeToString(hash[:])
		x := new(big.Int)
		x.SetString(hashHex, 16)

		yCoef := big.NewInt(1)
		if seed[0]&1 == 1 {
			yCoef.Neg(yCoef)
		}

		// Try to recover y
		if !ok {
			return fmt.Errorf("failed to get field prime value")
		}
		if y, err := RecoverY(x, betaBig, &fieldPrime); err == nil {
			y.Mul(yCoef, y)
			y.Mod(y, &fieldPrime)

			sXFelt := new(fp.Element).SetBigInt(x)
			sYFelt := new(fp.Element).SetBigInt(y)
			sXMv := mem.MemoryValueFromFieldElement(sXFelt)
			sYMv := mem.MemoryValueFromFieldElement(sYFelt)

			err = vm.Memory.WriteToNthStructField(sAddr, sXMv, 0)
			if err != nil {
				return err
			}
			return vm.Memory.WriteToNthStructField(sAddr, sYMv, 1)
		}
	}

	return fmt.Errorf("could not find a point on the curve")
}
