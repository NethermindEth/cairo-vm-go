package utils

import (
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
)

func WriteTo(vm *VM.VirtualMachine, segment int, offset uint64, val memory.MemoryValue) {
	err := vm.Memory.Write(segment, offset, &val)
	if err != nil {
		panic(err)
	}
}

func ReadFrom(vm *VM.VirtualMachine, segment int, offset uint64) memory.MemoryValue {
	val, err := vm.Memory.Read(segment, offset)
	if err != nil {
		panic(err)
	}
	return val
}
