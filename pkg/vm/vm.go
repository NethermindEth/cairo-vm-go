package vm

import (
	"fmt"

	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

const (
	programSegment = iota
	executionSegment
	dataSegment
)

type Context struct {
	Fp uint64
	Ap uint64
	Pc uint64
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

// todo(rodro): add a cache mechanism for not decoding the same instruction twice

// todo(rodro): how to know when te execute a hint or normal instruction

func (vm *VirtualMachine) RunStep() error {
	return nil
}
func (vm *VirtualMachine) RunStepAt(pc uint64) error {
	memoryValue, err := vm.MemoryManager.Memory.Read(programSegment, pc)
	if err != nil {
		return fmt.Errorf("cannot load step at %d: %w", pc, err)
	}

	bytecodeInstruction, err := memoryValue.ToFieldElement()
	if err != nil {
		return fmt.Errorf("cannot unwrap step at %d: %w", pc, err)
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
	op0, err := vm.computeOp0(instruction)
	if err != nil {
		return err
	}

	op1, err := vm.computeOp1(instruction, op0)
	if err != nil {
		return err
	}

	res, err := vm.computeRes(instruction, op0, op1)
	if err != nil {
		return err
	}
	dstAddress := vm.computeDstAddress(instruction)

	return vm.executeOpcode(instruction, res, &dstAddress)
}

func (vm *VirtualMachine) RunHint() error {
	return nil
}

func (vm *VirtualMachine) computeOp0(instruction *Instruction) (*mem.MemoryValue, error) {
	var op0Register uint64
	if instruction.Op0Register == Ap {
		op0Register = vm.Context.Ap
	} else {
		op0Register = vm.Context.Fp
	}
	op0Address := mem.CreateMemoryAddress(dataSegment, op0Register+uint64(instruction.OffOp0))

	return vm.MemoryManager.Memory.ReadFromAddress(op0Address)
}

func (vm *VirtualMachine) computeOp1(instruction *Instruction, op0 *mem.MemoryValue) (*mem.MemoryValue, error) {
	var op1Address *mem.MemoryAddress
	switch instruction.Op1Source {
	case Op0:
		// in this case Op0 is being used as an address, and must be of that type
		op0Address, err := op0.ToMemoryAddress()
		if err != nil {
			return nil, fmt.Errorf("expected op0 to be an address: %w", err)
		}
		op1Address = mem.CreateMemoryAddress(op0Address.SegmentIndex, op0Address.Offset)
	case Imm:
		// todo(rodro): would it be sensitive to check instruction.OffOp1 == 1
		op1Address = mem.CreateMemoryAddress(programSegment, vm.Context.Pc)
	case FpPlusOffOp1:
		op1Address = mem.CreateMemoryAddress(programSegment, vm.Context.Fp)
	case ApPlusOffOp1:
		op1Address = mem.CreateMemoryAddress(programSegment, vm.Context.Ap)
	}
	op1Address.Offset += uint64(instruction.OffOp1)

	return vm.MemoryManager.Memory.ReadFromAddress(op1Address)
}

func (vm *VirtualMachine) computeRes(
	instruction *Instruction, op0 *mem.MemoryValue, op1 *mem.MemoryValue,
) (*mem.MemoryValue, error) {
	switch instruction.Res {
	case Op1:
		return op1, nil
	case AddOperands:
		return mem.EmptyMemoryValueAs(op0.IsAddress()).Add(op0, op1)
	case MulOperands:
		return mem.EmptyMemoryValueAs(op0.IsAddress()).Mul(op0, op1)
	}
	return nil, fmt.Errorf("unknown res")
}

func (vm *VirtualMachine) computeDstAddress(instruction *Instruction) mem.MemoryAddress {
	var dstRegister uint64
	if instruction.DstRegister == Ap {
		dstRegister = vm.Context.Ap
	} else {
		dstRegister = vm.Context.Fp
	}

	// todo(rodro): this naive sum should be change, what if offset is neg
	return *mem.CreateMemoryAddress(dataSegment, dstRegister+uint64(instruction.OffDest))
}

func (vm *VirtualMachine) executeOpcode(
	instruction *Instruction, res *mem.MemoryValue, dstAddress *mem.MemoryAddress,
) error {
	if instruction.Opcode == Call {
		// assert op0 == pc + instruction size
		// asert dst == fp
		// next_fp = ap + 2
		// update ap
	} else {
		switch instruction.Opcode {
		case Nop:
		case Ret:
			// not implemented
		case AssertEq:
			return vm.MemoryManager.Memory.WriteToAddress(dstAddress, res)
		}
	}
	return nil

}
