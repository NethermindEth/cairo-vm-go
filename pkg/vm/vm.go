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
	Step          uint64
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
	manager.Memory.Segments = append(manager.Memory.Segments, mem.EmptySegmentWithCapacity(10))
	// 2 (dataSegment) <- segment where ap and fp move around
	manager.Memory.Segments = append(manager.Memory.Segments, mem.EmptySegmentWithLength(1))

	return &VirtualMachine{
		Context:       Context{Fp: 0, Ap: 0, Pc: 0},
		Step:          0,
		MemoryManager: manager,
		Config:        config,
	}, nil
}

// todo(rodro): add a cache mechanism for not decoding the same instruction twice

// todo(rodro): how to know when te execute a hint or normal instruction

func (vm *VirtualMachine) RunStep() error {
	memoryValue, err := vm.MemoryManager.Memory.Read(programSegment, vm.Context.Pc)
	if err != nil {
		return fmt.Errorf("cannot load step at %d: %w", vm.Context.Pc, err)
	}

	bytecodeInstruction, err := memoryValue.ToFieldElement()
	if err != nil {
		return fmt.Errorf("cannot unwrap step at %d: %w", vm.Context.Pc, err)
	}

	instruction, err := DecodeInstruction(bytecodeInstruction)
	if err != nil {
		return fmt.Errorf("cannot decode step at %d: %w", vm.Context.Pc, err)
	}

	err = vm.RunInstruction(instruction)
	if err != nil {
		return fmt.Errorf("cannot run step at %d: %w", vm.Context.Pc, err)
	}

	return nil
}
func (vm *VirtualMachine) RunStepAt(pc uint64) error {
	vm.Context.Pc = pc
	return vm.RunStep()
}

func (vm *VirtualMachine) RunInstruction(instruction *Instruction) error {
	// todo(rodro): any OffOpX can be negative, a better math system is required due to
	// substraction. Also it will need to handle overflows and underflows
	dstCell, err := vm.getCellDst(instruction)
	if err != nil {
		return err
	}

	op0Cell, err := vm.getCellOp0(instruction)
	if err != nil {
		return err
	}

	op1Cell, err := vm.getCellOp1(instruction, op0Cell)
	if err != nil {
		return err
	}

	res, err := vm.computeRes(instruction, op0Cell, op1Cell)
	if err != nil {
		return err
	}

	err = vm.opcodeAssertions(instruction, dstCell, op0Cell, res)
	if err != nil {
		return err
	}

	next_pc, err := vm.updatePc(instruction, dstCell, op1Cell, res)
	if err != nil {
		return err
	}

	next_ap, err := vm.updateAp(instruction, res)
	if err != nil {
		return err
	}

	next_fp, err := vm.updateFp(instruction, res)
	if err != nil {
		return err
	}

	vm.Context.Pc = next_pc
	vm.Context.Ap = next_ap
	vm.Context.Fp = next_fp

	vm.Step += 1
	return nil
}

func (vm *VirtualMachine) RunHint() error {
	return nil
}

func (vm *VirtualMachine) getCellDst(instruction *Instruction) (*mem.Cell, error) {
	var dstRegister uint64
	if instruction.DstRegister == Ap {
		dstRegister = vm.Context.Ap
	} else {
		dstRegister = vm.Context.Fp
	}

	// todo(rodro): fix this math
	return vm.MemoryManager.Memory.Peek(dataSegment, dstRegister+uint64(instruction.OffDest))
}

func (vm *VirtualMachine) getCellOp0(instruction *Instruction) (*mem.Cell, error) {
	var op0Register uint64
	if instruction.Op0Register == Ap {
		op0Register = vm.Context.Ap
	} else {
		op0Register = vm.Context.Fp
	}
	// todo(rodro): fix this math
	offset := op0Register + uint64(instruction.OffOp0)
	return vm.MemoryManager.Memory.Peek(dataSegment, op0Register+offset)
}

func (vm *VirtualMachine) getCellOp1(instruction *Instruction, op0Cell *mem.Cell) (*mem.Cell, error) {
	var op1Address *mem.MemoryAddress
	switch instruction.Op1Source {
	case Op0:
		// in this case Op0 is being used as an address, and must be of unwrapped as is
		op0Address, err := op0Cell.Read().ToMemoryAddress()
		if err != nil {
			return nil, fmt.Errorf("expected op0 to be an address: %w", err)
		}
		op1Address = mem.CreateMemoryAddress(op0Address.SegmentIndex, op0Address.Offset)
	case Imm:
		// todo(rodro): would it be sensitive to check instruction.OffOp1 == 1?
		op1Address = mem.CreateMemoryAddress(programSegment, vm.Context.Pc)
	case FpPlusOffOp1:
		op1Address = mem.CreateMemoryAddress(programSegment, vm.Context.Fp)
	case ApPlusOffOp1:
		op1Address = mem.CreateMemoryAddress(programSegment, vm.Context.Ap)
	}
	// todo(rodro): fix this math
	op1Address.Offset += uint64(instruction.OffOp1)

	return vm.MemoryManager.Memory.PeekFromAddress(op1Address)
}

func (vm *VirtualMachine) computeRes(
	instruction *Instruction, op0Cell *mem.Cell, op1Cell *mem.Cell,
) (*mem.MemoryValue, error) {
	if instruction.PcUpdate == Jnz {
		if instruction.Res == Op1 && instruction.Opcode == Nop && instruction.ApUpdate == AddImm {
			return nil, nil
		}
		return nil, fmt.Errorf("invalid flag combination calculating res")
	}

	switch instruction.Res {
	case Op1:
		return op1Cell.Read(), nil
	case AddOperands:
		op0 := op0Cell.Read()
		op1 := op0Cell.Read()
		return mem.EmptyMemoryValueAs(op0.IsAddress()).Add(op0, op1)
	case MulOperands:
		op0 := op0Cell.Read()
		op1 := op0Cell.Read()
		return mem.EmptyMemoryValueAsFelt().Mul(op0, op1)
	}

	return nil, fmt.Errorf("unknown res flag value: %d", instruction.Res)
}

func (vm *VirtualMachine) updatePc(
	instruction *Instruction,
	dstCell *mem.Cell,
	op1Cell *mem.Cell,
	res *mem.MemoryValue,
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
		dest, err := dstCell.Read().ToFieldElement()
		if err != nil {
			return 0, err
		}

		if dest.IsZero() {
			return vm.Context.Pc + uint64(instruction.Size()), nil
		}

		// todo(rodro): math check when relAddr is negative
		relAddr, err := res.Uint64()
		if err != nil {
			return 0, err
		}
		return vm.Context.Pc + relAddr, nil

	}
	return 0, fmt.Errorf("unkwon pc update value: %d", instruction.PcUpdate)
}

func (vm *VirtualMachine) opcodeAssertions(
	instruction *Instruction,
	dstCell *mem.Cell,
	op0Cell *mem.Cell,
	res *mem.MemoryValue,
) error {
	switch instruction.Opcode {
	case Call:
		err := op0Cell.Write(
			mem.MemoryValueFromUint(vm.Context.Pc + uint64(instruction.Size())),
		)
		if err != nil {
			return err
		}

		err = dstCell.Write(mem.MemoryValueFromUint(vm.Context.Fp))
		if err != nil {
			return err
		}
	case AssertEq:
		err := dstCell.Write(res)
		if err != nil {
			return err
		}
	}
	return nil
}

func (vm *VirtualMachine) updateAp(instruction *Instruction, res *mem.MemoryValue) (uint64, error) {
	if instruction.Opcode == Call {
		if instruction.ApUpdate == Add2 {
			return vm.Context.Ap + 2, nil
		}
		return 0, fmt.Errorf("cannot update ap, invalid flag combination: Call & Add2")
	}

	switch instruction.ApUpdate {
	case SameAp:
		return vm.Context.Ap, nil
	case AddImm:
		res64, err := res.Uint64()
		if err != nil {
			return 0, err
		}
		return vm.Context.Ap + res64, nil
	case Add1:
		return vm.Context.Ap + 1, nil
	}
	return 0, fmt.Errorf("cannot update ap, unknown ApUpdate flag: %d", instruction.ApUpdate)
}

func (vm *VirtualMachine) updateFp(instruction *Instruction, res *mem.MemoryValue) (uint64, error) {
	switch instruction.Opcode {
	case Call:
		return vm.Context.Ap + 2, nil
	case Ret:
		res64, err := res.Uint64()
		if err != nil {
			return 0, err
		}
		return vm.Context.Fp + res64, nil
	default:
		return vm.Context.Fp, nil
	}
}
