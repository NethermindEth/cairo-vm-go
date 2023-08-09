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
	manager, err := mem.CreateMemoryManager(programBytecode)
	if err != nil {
		return nil, fmt.Errorf("error creating new virtual machine: %w", err)
	}

	return &VirtualMachine{
		Context{Fp: 0, Ap: 0, Pc: 0},
		manager,
		config,
	}, nil
}

// todo(rodro): add a cache mechanism for not decoding the same instruction twice

// todo(rodro): how to know when te execute a hint or normal instruction

func (vm *VirtualMachine) RunStep() error {
	return nil
}
func (vm *VirtualMachine) RunStepAt(pc uint) error {
	bytecodeInstruction, err := vm.MemoryManager.Memory.GetInstructionBytecode(pc)
	if err != nil {
		return fmt.Errorf("cannot load step at %d: %w", pc, err)
	}
	instruction, err := DecodeInstruction(bytecodeInstruction)
	if err != nil {
		return fmt.Errorf("cannot decode step at %d: %w", pc, err)
	}

	err = vm.RunInstruction(instruction)
	if err != nil {
		return fmt.Errorf("cannot run step at %d: %w", pc, err)
	}

	return nil
}

func (vm *VirtualMachine) RunInstruction(instruction *Instruction) error {
	switch instruction.Opcode {
	case AssertEq:
		vm.assertEqual(instruction)
	default:
		return fmt.Errorf("unimplemented opcode: %d", instruction.Opcode)
	}
	return nil
}

func (vm *VirtualMachine) RunHint() error {
	return nil
}

func (vm *VirtualMachine) assertEqual(instruction *Instruction) error {
        instruction.
}
