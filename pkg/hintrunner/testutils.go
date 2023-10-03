package hintrunner

import (
	"github.com/NethermindEth/cairo-vm-go/pkg/vm"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
)

func defaultVirtualMachine() (*vm.VirtualMachine, *memory.MemoryManager) {
	manager := memory.CreateMemoryManager()
	manager.Memory.AllocateEmptySegment()
	manager.Memory.AllocateEmptySegment()

	vm, err := vm.NewVirtualMachine(vm.Context{}, manager.Memory, vm.VirtualMachineConfig{})
	if err != nil {
		panic(err)
	}
	return vm, manager
}

func writeTo(vm *VM.VirtualMachine, segment uint64, offset uint64, val memory.MemoryValue) {
	_ = vm.Memory.Write(segment, offset, &val)
}

func readFrom(vm *VM.VirtualMachine, segment uint64, offset uint64) memory.MemoryValue {
	val, _ := vm.Memory.Read(segment, offset)
	return val
}
