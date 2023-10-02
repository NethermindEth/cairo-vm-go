package vm

import (
	"fmt"

	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"

	"github.com/NethermindEth/cairo-vm-go/pkg/builtinrunner"
	safemath "github.com/NethermindEth/cairo-vm-go/pkg/safemath"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
)

const (
	ProgramSegment = iota
	ExecutionSegment
)

// Required by the VM to run hints.
//
// HintRunner is defined as an external component of the VM so any user
// could define its own, allowing the use of custom hints
type HintRunner interface {
	RunHint(vm *VirtualMachine) error
}

// Represents the current execution context of the vm
type Context struct {
	Pc *mem.MemoryAddress
	Fp uint64
	Ap uint64
}

func (ctx *Context) String() string {
	return fmt.Sprintf(
		"Context {pc: %d:%d, fp: %d, ap: %d}",
		ctx.Pc.SegmentIndex,
		ctx.Pc.Offset,
		ctx.Fp,
		ctx.Ap,
	)
}

func (ctx *Context) AddressAp() *mem.MemoryAddress {
	return &mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: ctx.Ap}
}

func (ctx *Context) AddressFp() *mem.MemoryAddress {
	return &mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: ctx.Fp}
}

func (ctx *Context) AddressPc() *mem.MemoryAddress {
	return &mem.MemoryAddress{SegmentIndex: ctx.Pc.SegmentIndex, Offset: ctx.Pc.Offset}
}

// relocates pc, ap and fp to be their real address value
// that is, pc + 1, ap + programSegmentOffset, fp + programSegmentOffset
func (ctx *Context) Relocate(executionSegmentOffset uint64) Trace {
	return Trace{
		// todo(rodro): this should be improved upon
		Pc: ctx.Pc.Offset + 1,
		Ap: ctx.Ap + executionSegmentOffset,
		Fp: ctx.Fp + executionSegmentOffset,
	}
}

type Trace struct {
	Pc uint64
	Fp uint64
	Ap uint64
}

// This type represents the current execution context of the vm
type VirtualMachineConfig struct {
	// If true, the vm outputs the trace and the relocated memory at the end of execution
	ProofMode bool
}

type VirtualMachine struct {
	Context       Context
	MemoryManager *mem.MemoryManager
	Step          uint64
	Trace         []Context
	BuiltinRunner builtinrunner.BuiltinRunner
	config        VirtualMachineConfig
	// instructions cache
	instructions map[uint64]*Instruction

	// builtins
}

// NewVirtualMachine creates a VM from the program bytecode using a specified config.
func NewVirtualMachine(programBytecode []*f.Element, config VirtualMachineConfig) (*VirtualMachine, error) {
	// Initialize memory with to initial segments:
	// the first one for the program segment and
	// the second one to keep track of the execution
	manager := mem.CreateMemoryManager()
	// 0 (programSegment) <- segment where the bytecode is stored
	_, err := manager.Memory.AllocateSegment(programBytecode)
	if err != nil {
		return nil, fmt.Errorf("error loading bytecode: %w", err)
	}
	// 1 (executionSegment) <- segment where ap and fp move around
	manager.Memory.AllocateEmptySegment()

	// Initialize the trace if necesary
	var trace []Context
	if config.ProofMode {
		trace = make([]Context, 0)
	}

	return &VirtualMachine{
		Context: Context{
			Fp: 0,
			Ap: 0,
			Pc: &mem.MemoryAddress{
				SegmentIndex: ProgramSegment,
				Offset:       0,
			},
		},
		Step:          0,
		MemoryManager: manager,
		Trace:         trace,
		config:        config,
		instructions:  make(map[uint64]*Instruction, len(programBytecode)),
	}, nil
}

// todo(rodro): add a cache mechanism for not decoding the same instruction twice

func (vm *VirtualMachine) RunStep(hintRunner HintRunner) error {
	// Run hint
	err := hintRunner.RunHint(vm)
	if err != nil {
		return fmt.Errorf("pc %d: %w", vm.Context.Pc, err)
	}

	instruction, ok := vm.instructions[vm.Context.Pc.Offset]
	if !ok {
		// ddecode instruction
		memoryValue, err := vm.MemoryManager.Memory.ReadFromAddress(vm.Context.Pc)
		if err != nil {
			return fmt.Errorf("pc %d: %w", vm.Context.Pc, err)
		}

		bytecodeInstruction, err := memoryValue.ToFieldElement()
		if err != nil {
			return fmt.Errorf("pc %d: %w", vm.Context.Pc, err)
		}

		instruction, err = DecodeInstruction(bytecodeInstruction)
		if err != nil {
			return fmt.Errorf("pc %d: %w", vm.Context.Pc, err)
		}
		vm.instructions[vm.Context.Pc.Offset] = instruction
	}

	// store the trace before state change
	if vm.config.ProofMode {
		vm.Trace = append(vm.Trace, vm.Context)
	}

	err = vm.RunInstruction(instruction)
	if err != nil {
		return fmt.Errorf("pc %d: %w", vm.Context.Pc, err)
	}

	vm.Step++
	return nil
}

func (vm *VirtualMachine) RunInstruction(instruction *Instruction) error {
	dstCell, err := vm.getCellDst(instruction)
	if err != nil {
		return fmt.Errorf("dst cell: %w", err)
	}

	op0Cell, err := vm.getCellOp0(instruction)
	if err != nil {
		return fmt.Errorf("op0 cell: %w", err)
	}

	op1Cell, err := vm.getCellOp1(instruction, op0Cell)
	if err != nil {
		return fmt.Errorf("op1 cell: %w", err)
	}

	//Here do stuff
	if op0Cell.Read().IsAddress() {
		v, _ := op0Cell.Read().ToMemoryAddress()
		for _, i := range vm.BuiltinRunner.Builtins() {
			if v.SegmentIndex == i.Segment() {
				i.Run()
			}
		}
	}

	res, err := vm.inferOperand(instruction, dstCell, op0Cell, op1Cell)
	if err != nil {
		return fmt.Errorf("res infer: %w", err)
	}

	if res == nil {
		res, err = vm.computeRes(instruction, op0Cell, op1Cell)
		if err != nil {
			return fmt.Errorf("compute res: %w", err)
		}
	}

	err = vm.opcodeAssertions(instruction, dstCell, op0Cell, res)
	if err != nil {
		return fmt.Errorf("opcode assertions: %w", err)
	}

	nextPc, err := vm.updatePc(instruction, dstCell, op1Cell, res)
	if err != nil {
		return fmt.Errorf("pc update: %w", err)
	}

	nextAp, err := vm.updateAp(instruction, res)
	if err != nil {
		return fmt.Errorf("ap update: %w", err)
	}

	nextFp, err := vm.updateFp(instruction, dstCell)
	if err != nil {
		return fmt.Errorf("fp update: %w", err)
	}

	vm.Context.Pc = nextPc
	vm.Context.Ap = nextAp
	vm.Context.Fp = nextFp

	return nil
}

// It returns the current trace entry, the public memory, and the occurrence of an error
func (vm *VirtualMachine) Proof() ([]Trace, []*f.Element, error) {
	if !vm.config.ProofMode {
		return nil, nil, fmt.Errorf("proof mode is off")
	}

	relocatedTrace := vm.relocateTrace()

	relocatedMemory := vm.MemoryManager.RelocateMemory()

	return relocatedTrace, relocatedMemory, nil
}

func (vm *VirtualMachine) getCellDst(instruction *Instruction) (*mem.Cell, error) {
	var dstRegister *mem.MemoryAddress
	if instruction.DstRegister == Ap {
		dstRegister = vm.Context.AddressAp()
	} else {
		dstRegister = vm.Context.AddressFp()
	}

	addr, isOverflow := safemath.SafeOffset(dstRegister.Offset, instruction.OffDest)
	if isOverflow {
		return nil, fmt.Errorf("offset overflow: %d + %d", dstRegister, instruction.OffDest)
	}
	return vm.MemoryManager.Memory.Peek(dstRegister.SegmentIndex, addr)
}

func (vm *VirtualMachine) getCellOp0(instruction *Instruction) (*mem.Cell, error) {
	var op0Register *mem.MemoryAddress
	if instruction.Op0Register == Ap {
		op0Register = vm.Context.AddressAp()
	} else {
		op0Register = vm.Context.AddressFp()
	}

	addr, isOverflow := safemath.SafeOffset(op0Register.Offset, instruction.OffOp0)
	if isOverflow {
		return nil, fmt.Errorf("offset overflow: %d + %d", op0Register, instruction.OffOp0)
	}
	return vm.MemoryManager.Memory.Peek(op0Register.SegmentIndex, addr)
}

func (vm *VirtualMachine) getCellOp1(instruction *Instruction, op0Cell *mem.Cell) (*mem.Cell, error) {
	var op1Address *mem.MemoryAddress
	switch instruction.Op1Source {
	case Op0:
		// in this case Op0 is being used as an address, and must be of unwrapped as it
		op0Address, err := op0Cell.Read().ToMemoryAddress()
		if err != nil {
			return nil, fmt.Errorf("op0 is not an address: %w", err)
		}
		op1Address = mem.NewMemoryAddress(op0Address.SegmentIndex, op0Address.Offset)
	case Imm:
		op1Address = vm.Context.AddressPc()
	case FpPlusOffOp1:
		op1Address = vm.Context.AddressFp()
	case ApPlusOffOp1:
		op1Address = vm.Context.AddressAp()
	}

	addr, isOverflow := safemath.SafeOffset(op1Address.Offset, instruction.OffOp1)
	if isOverflow {
		return nil, fmt.Errorf("offset overflow: %d + %d", op1Address.Offset, instruction.OffOp1)
	}
	op1Address.Offset = addr

	return vm.MemoryManager.Memory.PeekFromAddress(op1Address)
}

// when there is an assertion with a substraction or division like : x = y - z
// the compiler treats it as y = x + z. This means that the VM knows the
// dstCell value and either op0Cell xor op1Cell. This function infers the
// unknow operand as well as the `res` auxiliar value
func (vm *VirtualMachine) inferOperand(
	instruction *Instruction, dstCell *mem.Cell, op0Cell *mem.Cell, op1Cell *mem.Cell,
) (*mem.MemoryValue, error) {
	if instruction.Opcode != AssertEq ||
		(instruction.Res != AddOperands && instruction.Res != MulOperands) ||
		(op0Cell.Accessed && op1Cell.Accessed) {
		return nil, nil
	}
	if !dstCell.Accessed {
		return nil, fmt.Errorf("dst cell is unknown")
	}

	var knownOpCell *mem.Cell
	var unknownOpCell *mem.Cell
	if op0Cell.Accessed {
		knownOpCell = op0Cell
		unknownOpCell = op1Cell
	} else {
		knownOpCell = op1Cell
		unknownOpCell = op0Cell
	}

	var missingVal *mem.MemoryValue
	var err error
	dst := dstCell.Read()
	if instruction.Res == AddOperands {
		missingVal, err = mem.EmptyMemoryValueAs(dst.IsAddress()).Sub(dst, knownOpCell.Read())
	} else {
		missingVal, err = mem.EmptyMemoryValueAsFelt().Div(dst, knownOpCell.Read())
	}
	if err != nil {
		return nil, err
	}

	err = unknownOpCell.Write(missingVal)
	if err != nil {
		return nil, err
	}

	return dst, nil
}

func (vm *VirtualMachine) computeRes(
	instruction *Instruction, op0Cell *mem.Cell, op1Cell *mem.Cell,
) (*mem.MemoryValue, error) {
	switch instruction.Res {
	case Unconstrained:
		return nil, nil
	case Op1:
		return op1Cell.Read(), nil
	case AddOperands:
		op0 := op0Cell.Read()
		op1 := op1Cell.Read()
		return mem.EmptyMemoryValueAs(op0.IsAddress()).Add(op0, op1)
	case MulOperands:
		op0 := op0Cell.Read()
		op1 := op1Cell.Read()
		return mem.EmptyMemoryValueAsFelt().Mul(op0, op1)
	}

	return nil, fmt.Errorf("invalid res flag value: %d", instruction.Res)
}

func (vm *VirtualMachine) opcodeAssertions(
	instruction *Instruction,
	dstCell *mem.Cell,
	op0Cell *mem.Cell,
	res *mem.MemoryValue,
) error {
	switch instruction.Opcode {
	case Call:
		// Store at [ap] the current fp
		err := dstCell.Write(mem.MemoryValueFromMemoryAddress(vm.Context.AddressFp()))
		if err != nil {
			return err
		}
		err = dstCell.Write(mem.MemoryValueFromMemoryAddress(vm.Context.AddressFp()))
		if err != nil {
			return err
		}

		// Write in [ap + 1] the next instruction to execute
		err = op0Cell.Write(
			mem.MemoryValueFromSegmentAndOffset(
				vm.Context.Pc.SegmentIndex,
				vm.Context.Pc.Offset+uint64(instruction.Size()),
			),
		)
		if err != nil {
			return err
		}
		err = dstCell.Write(mem.MemoryValueFromMemoryAddress(vm.Context.AddressFp()))
		if err != nil {
			return err
		}
	case AssertEq:
		// assert that the calculated res is stored in dst
		err := dstCell.Write(res)
		if err != nil {
			return err
		}
	}
	return nil
}

func (vm *VirtualMachine) updatePc(
	instruction *Instruction,
	dstCell *mem.Cell,
	op1Cell *mem.Cell,
	res *mem.MemoryValue,
) (*mem.MemoryAddress, error) {
	switch instruction.PcUpdate {
	case NextInstr:
		return mem.NewMemoryAddress(
			vm.Context.Pc.SegmentIndex,
			vm.Context.Pc.Offset+uint64(instruction.Size()),
		), nil
	case Jump:
		addr, err := res.ToMemoryAddress()
		if err != nil {
			return nil, fmt.Errorf("absolute jump: %w", err)
		}
		return addr, nil
	case JumpRel:
		val, err := res.ToFieldElement()
		if err != nil {
			return nil, fmt.Errorf("relative jump: %w", err)
		}
		return new(mem.MemoryAddress).Add(vm.Context.Pc, val)
	case Jnz:
		dest, err := dstCell.Read().ToFieldElement()
		if err != nil {
			return nil, err
		}

		if dest.IsZero() {
			return mem.NewMemoryAddress(vm.Context.Pc.SegmentIndex, vm.Context.Pc.Offset+uint64(instruction.Size())), nil
		}
		val, err := op1Cell.Read().ToFieldElement()
		if err != nil {
			return nil, err
		}
		return new(mem.MemoryAddress).Add(vm.Context.Pc, val)

	}
	return nil, fmt.Errorf("unkwon pc update value: %d", instruction.PcUpdate)
}

func (vm *VirtualMachine) updateAp(instruction *Instruction, res *mem.MemoryValue) (uint64, error) {
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
	case Add2:
		return vm.Context.Ap + 2, nil
	}
	return 0, fmt.Errorf("cannot update ap, unknown ApUpdate flag: %d", instruction.ApUpdate)
}

func (vm *VirtualMachine) updateFp(instruction *Instruction, dstCell *mem.Cell) (uint64, error) {
	switch instruction.Opcode {
	case Call:
		// [ap] and [ap + 1] are written to memory
		return vm.Context.Ap + 2, nil
	case Ret:
		// [dst] should be a memory address of the form (executionSegment, fp - 2)
		dst, err := dstCell.Read().ToMemoryAddress()
		if err != nil {
			return 0, fmt.Errorf("ret: %w", err)
		}
		return dst.Offset, nil
	default:
		return vm.Context.Fp, nil
	}
}

func (vm *VirtualMachine) relocateTrace() []Trace {
	// one is added, because prover expect that the first element to be on
	// indexed on 1 instead of 0
	relocatedTrace := make([]Trace, len(vm.Trace))
	totalBytecode := vm.MemoryManager.Memory.Segments[ProgramSegment].Len() + 1
	for i := range vm.Trace {
		relocatedTrace[i] = vm.Trace[i].Relocate(totalBytecode)
	}
	return relocatedTrace
}
