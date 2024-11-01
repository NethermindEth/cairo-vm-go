package vm

import (
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"

	a "github.com/NethermindEth/cairo-vm-go/pkg/assembler"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
)

func DefaultVirtualMachine() *VirtualMachine {
	return defaultVirtualMachineWithBytecode(nil)
}

func defaultVirtualMachineWithCode(code string) *VirtualMachine {
	bytecode, _, err := a.CasmToBytecode(code)
	if err != nil {
		panic(err)
	}

	return defaultVirtualMachineWithBytecode(bytecode)
}

func defaultVirtualMachineWithBytecode(bytecode []*f.Element) *VirtualMachine {
	memory := mem.InitializeEmptyMemory()
	_, err := memory.AllocateSegment(bytecode)
	if err != nil {
		panic(err)
	}

	memory.AllocateEmptySegment()

	vm, err := NewVirtualMachine(Context{}, memory, VirtualMachineConfig{})
	if err != nil {
		panic(err)
	}
	return vm
}
