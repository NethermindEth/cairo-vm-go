package hintrunner

import (
	"github.com/NethermindEth/cairo-vm-go/pkg/safemath"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
)

func defaultVirtualMachine() *VM.VirtualMachine {
	vm, _ := VM.NewVirtualMachine(make([]*safemath.LazyFelt, 0), VM.VirtualMachineConfig{})
	return vm
}

func writeTo(vm *VM.VirtualMachine, segment uint64, offset uint64, val *memory.MemoryValue) {
	_ = vm.MemoryManager.Memory.Write(segment, offset, val)
}

func readFrom(vm *VM.VirtualMachine, segment uint64, offset uint64) *memory.MemoryValue {
	val, _ := vm.MemoryManager.Memory.Read(segment, offset)
	return val
}
