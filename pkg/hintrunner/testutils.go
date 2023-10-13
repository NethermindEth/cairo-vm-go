package hintrunner

import (
	"github.com/NethermindEth/cairo-vm-go/pkg/vm"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
)

func defaultVirtualMachine() *vm.VirtualMachine {
	memory := memory.InitializeEmptyMemory()
	memory.AllocateEmptySegment()
	memory.AllocateEmptySegment()

	vm, err := vm.NewVirtualMachine(vm.Context{}, memory, vm.VirtualMachineConfig{})
	if err != nil {
		panic(err)
	}
	return vm
}

func writeTo(vm *VM.VirtualMachine, segment uint64, offset uint64, val memory.MemoryValue) {
	err := vm.Memory.Write(segment, offset, &val)
	if err != nil {
		panic(err)
	}
}

func readFrom(vm *VM.VirtualMachine, segment uint64, offset uint64) memory.MemoryValue {
	val, err := vm.Memory.Read(segment, offset)
	if err != nil {
		panic(err)
	}
	return val
}
