package hintrunner

import (
	"encoding/binary"
	"fmt"
	"math/rand"

	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func ResolveAsAddress(vm *VM.VirtualMachine, op ResOperander) (mem.MemoryAddress, error) {
	mv, err := op.Resolve(vm)
	if err != nil {
		return mem.UnknownAddress, fmt.Errorf("%s: %w", op, err)
	}

	addr, err := mv.MemoryAddress()
	if err != nil {
		return mem.UnknownAddress, fmt.Errorf("%s: %w", op, err)
	}

	return *addr, nil
}

func ResolveAsFelt(vm *VM.VirtualMachine, op ResOperander) (f.Element, error) {
	mv, err := op.Resolve(vm)
	if err != nil {
		return f.Element{}, fmt.Errorf("%s: %w", op, err)
	}

	felt, err := mv.FieldElement()
	if err != nil {
		return f.Element{}, fmt.Errorf("%s: %w", op, err)
	}

	return *felt, nil
}

func ResolveAsUint64(vm *VM.VirtualMachine, op ResOperander) (uint64, error) {
	mv, err := op.Resolve(vm)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	uint64Value, err := mv.Uint64()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return uint64Value, nil
}

func randomFeltElement(rand *rand.Rand) f.Element {
	b := [32]byte{}
	binary.BigEndian.PutUint64(b[24:32], rand.Uint64())
	binary.BigEndian.PutUint64(b[16:24], rand.Uint64())
	binary.BigEndian.PutUint64(b[8:16], rand.Uint64())
	//Limit to 59 bits so at max we have a 251 bit number
	binary.BigEndian.PutUint64(b[0:8], rand.Uint64()>>5)
	f, _ := f.BigEndian.Element(&b)
	return f
}

func randomFeltElementU128(rand *rand.Rand) f.Element {
	b := [32]byte{}
	binary.BigEndian.PutUint64(b[24:32], rand.Uint64())
	binary.BigEndian.PutUint64(b[16:24], rand.Uint64())
	f, _ := f.BigEndian.Element(&b)
	return f
}

func defaultRandGenerator() *rand.Rand {
	return rand.New(rand.NewSource(0))
}
