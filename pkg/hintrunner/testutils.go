package hintrunner

import (
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func defaultVirtualMachine() *VM.VirtualMachine {
	vm, _ := VM.NewVirtualMachine(make([]*f.Element, 0), VM.VirtualMachineConfig{})
	return vm
}

func writeTo(vm *VM.VirtualMachine, segment uint64, offset uint64, val memory.MemoryValue) {
	_ = vm.MemoryManager.Memory.Write(segment, offset, &val)
}

func readFrom(vm *VM.VirtualMachine, segment uint64, offset uint64) memory.MemoryValue {
	val, _ := vm.MemoryManager.Memory.Read(segment, offset)
	return val
}
