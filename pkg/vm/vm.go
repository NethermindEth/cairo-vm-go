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
func NewVirtualMachine(programBytecode []*f.Element, config VirtualMachineConfig) (*VirtualMachine, error) {
	manager, err := mem.CreateMemoryManager()
	if err != nil {
		return nil, fmt.Errorf("error creating new virtual machine: %w", err)
	}

	// 0 (programSegment) <- segment where the bytecode is stored
	_, err = manager.Memory.AllocateSegment(programBytecode)
	if err != nil {
		return nil, fmt.Errorf("error loading bytecode: %w", err)
	}

	// 1 (executionSegment) <- segment where the stack trace will be stored
	manager.Memory.AllocateEmptySegment()
	// 2 (dataSegment) <- segment where ap and fp move around
	manager.Memory.AllocateEmptySegment()

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

	next_pc, err := vm.updatePc(instruction, res, dstAddress)
	if err != nil {
		return err
	}

	vm.Context.Pc = next_pc

	return nil
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
	// todo: OffOp0 can be negative, a better system is required, perhaps a substraction
	op0Address := mem.CreateMemoryAddress(dataSegment, op0Register+uint64(instruction.OffOp0))

	return vm.MemoryManager.Memory.ReadFromAddress(op0Address)
}

func (vm *VirtualMachine) computeOp1(instruction *Instruction, op0 *mem.MemoryValue) (*mem.MemoryValue, error) {
	var op1Address *mem.MemoryAddress
	switch instruction.Op1Source {
	case Op0:
		// in this case Op0 is being used as an address, and must be of unwrapped as is
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
	if instruction.PcUpdate == Jnz {
		if instruction.Res == Op1 && instruction.Opcode == Nop && instruction.ApUpdate == AddImm {
			return op1, nil
		}
		return nil, fmt.Errorf("invalid flag combination calculating res")
	}

	switch instruction.Res {
	case Op1:
		return op1, nil
	case AddOperands:
		return mem.EmptyMemoryValueAs(op0.IsAddress()).Add(op0, op1)
	case MulOperands:
		return mem.EmptyMemoryValueAs(op0.IsAddress()).Mul(op0, op1)
	}

	return nil, fmt.Errorf("unknown res flag value: %d", instruction.Res)
}

func (vm *VirtualMachine) computeDstAddress(instruction *Instruction) *mem.MemoryAddress {
	var dstRegister uint64
	if instruction.DstRegister == Ap {
		dstRegister = vm.Context.Ap
	} else {
		dstRegister = vm.Context.Fp
	}

	// todo(rodro): this naive sum should be changed because the offset can be negative as well
	// todo(rodro): there is a need to check for underflow as well
	return mem.CreateMemoryAddress(dataSegment, dstRegister+uint64(instruction.OffDest))
}

func (vm *VirtualMachine) updatePc(
	instruction *Instruction,
	res *mem.MemoryValue,
	destAddr *mem.MemoryAddress,
) (uint64, error) {
	switch instruction.PcUpdate {
	case NextInstr:
		return vm.Context.Pc + uint64(instruction.Size()), nil
	case Jump:
		return res.Uint64()
	case JumpRel:
		relAddr, err := res.Uint64()
		if err != nil {
			return 0, err
		}
		return vm.Context.Pc + relAddr, nil
	case Jnz:
		destValue, err := vm.MemoryManager.Memory.ReadFromAddress(destAddr)
		if err != nil {
			return 0, err
		}
		dest, err := destValue.Uint64()
		if err != nil {
			return 0, err
		}

		if dest == 0 {
			return vm.Context.Pc + uint64(instruction.Size()), nil
		}

		relAddr, err := res.Uint64()
		if err != nil {
			return 0, err
		}
		return vm.Context.Pc + relAddr, nil

	}
	return 0, fmt.Errorf("unkwon pc update value: %d", instruction.PcUpdate)
}

func (vm *VirtualMachine) updateAp(instruction *Instruction) uint64 {
	return 0
}
