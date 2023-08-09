package vm

import (
	"fmt"

	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
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
	Context       Context
	MemoryManager *mem.MemoryManager
	Config        VirtualMachineConfig
}

// NewVirtualMachine creates a VM from the program bytecode using a specified config.
func NewVirtualMachine(programBytecode *[]f.Element, config VirtualMachineConfig) (*VirtualMachine, error) {
	manager, err := mem.CreateMemoryManager()
	if err != nil {
		return nil, fmt.Errorf("error creating new virtual machine: %w", err)
	}

	_, err = manager.Memory.AllocateSegment(programBytecode)
	if err != nil {
		return nil, fmt.Errorf("error loading bytecode: %w", err)
	}

	return &VirtualMachine{
		Context{Fp: 0, Ap: 0, Pc: 0},
		manager,
		config,
	}, nil
}

// RunInstructionFrom executes the program starting at the specified Program Counter.
func (vm *VirtualMachine) RunStep() error {
	return nil
}

func (vm *VirtualMachine) RunInstruction(instruction *Instruction) error {
	return nil
}

func (vm *VirtualMachine) RunHint(instruction *Instruction) error {
	return nil
}
