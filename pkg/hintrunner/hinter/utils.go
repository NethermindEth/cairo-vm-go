package hinter

import (
	"fmt"

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