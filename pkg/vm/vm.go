package vm

import (
	f "github.com/NethermindEth/juno/core/felt"
)

type Context struct {
	Fp uint
	Ap uint
	Pc uint
}

type VirtualMachineConfig struct {
	Trace bool
	// Todo(rodro): Update this property to include all builtins
	Builtins bool
}

type VirtualMachine struct {
	Context              Context
	MemorySegmentManager MemorySegmentManager
	Config               VirtualMachineConfig
}

// NewVirtualMachine creates a VM from the program bytecode using a specified config.
func NewVirtualMachine(programBytecode *[]f.Felt, config VirtualMachineConfig) *VirtualMachine {
	return &VirtualMachine{
		Context{Fp: 0, Ap: 0, Pc: 0},
		CreateMemorySegmentManager(programBytecode),
		config,
	}
}

// RunInstructionFrom executes the program starting at the specified Program Counter.
func (vm *VirtualMachine) RunStep() error {
	return nil
}

func (vm *VirtualMachine) RunInstruction() error {
	return nil
}
