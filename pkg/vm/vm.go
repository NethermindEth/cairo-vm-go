package vm

import (
	"fmt"

	a "github.com/NethermindEth/cairo-vm-go/pkg/assembler"
	safemath "github.com/NethermindEth/cairo-vm-go/pkg/safemath"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
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
	Pc mem.MemoryAddress
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

func (ctx *Context) AddressAp() mem.MemoryAddress {
	return mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: ctx.Ap}
}

func (ctx *Context) AddressFp() mem.MemoryAddress {
	return mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: ctx.Fp}
}

func (ctx *Context) AddressPc() mem.MemoryAddress {
	return mem.MemoryAddress{SegmentIndex: ctx.Pc.SegmentIndex, Offset: ctx.Pc.Offset}
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
	Context Context
	Memory  *mem.Memory
	Step    uint64
	Trace   []Context
	config  VirtualMachineConfig
	// instructions cache
	instructions map[uint64]*a.Instruction
}

// NewVirtualMachine creates a VM from the program bytecode using a specified config.
func NewVirtualMachine(initialContext Context, memory *mem.Memory, config VirtualMachineConfig) (*VirtualMachine, error) {
	// Initialize the trace if necesary
	var trace []Context
	if config.ProofMode {
		trace = make([]Context, 0)
	}

	return &VirtualMachine{
		Context:      initialContext,
		Memory:       memory,
		Trace:        trace,
		config:       config,
		instructions: make(map[uint64]*a.Instruction),
	}, nil
}

// todo(rodro): add a cache mechanism for not decoding the same instruction twice

func (vm *VirtualMachine) RunStep(hintRunner HintRunner) error {
	// if instruction is not in cache, redecode and store it
	instruction, ok := vm.instructions[vm.Context.Pc.Offset]
	if !ok {
		memoryValue, err := vm.Memory.ReadFromAddress(&vm.Context.Pc)
		if err != nil {
			return fmt.Errorf("reading instruction: %w", err)
		}

		bytecodeInstruction, err := memoryValue.FieldElement()
		if err != nil {
			return fmt.Errorf("reading instruction: %w", err)
		}

		instruction, err = a.DecodeInstruction(bytecodeInstruction)
		if err != nil {
			return fmt.Errorf("decoding instruction: %w", err)
		}
		vm.instructions[vm.Context.Pc.Offset] = instruction
	}

	// store the trace before state change
	if vm.config.ProofMode {
		vm.Trace = append(vm.Trace, vm.Context)
	}

	err := vm.RunInstruction(instruction)
	if err != nil {
		return fmt.Errorf("running instruction: %w", err)
	}

	vm.Step++
	return nil
}

func (vm *VirtualMachine) RunInstruction(instruction *a.Instruction) error {
	dstAddr, err := vm.getDstAddr(instruction)
	if err != nil {
		return fmt.Errorf("dst cell: %w", err)
	}

	op0Addr, err := vm.getOp0Addr(instruction)
	if err != nil {
		return fmt.Errorf("op0 cell: %w", err)
	}

	op1Addr, err := vm.getOp1Addr(instruction, &op0Addr)
	if err != nil {
		return fmt.Errorf("op1 cell: %w", err)
	}

	res, err := vm.inferOperand(instruction, &dstAddr, &op0Addr, &op1Addr)
	if err != nil {
		return fmt.Errorf("res infer: %w", err)
	}
	if !res.Known() {
		res, err = vm.computeRes(instruction, &op0Addr, &op1Addr)
		if err != nil {
			return fmt.Errorf("compute res: %w", err)
		}
	}

	err = vm.opcodeAssertions(instruction, &dstAddr, &op0Addr, &res)
	if err != nil {
		return fmt.Errorf("opcode assertions: %w", err)
	}

	nextPc, err := vm.updatePc(instruction, &dstAddr, &op1Addr, &res)
	if err != nil {
		return fmt.Errorf("pc update: %w", err)
	}

	nextAp, err := vm.updateAp(instruction, &res)
	if err != nil {
		return fmt.Errorf("ap update: %w", err)
	}

	nextFp, err := vm.updateFp(instruction, &dstAddr)
	if err != nil {
		return fmt.Errorf("fp update: %w", err)
	}

	vm.Context.Pc = nextPc
	vm.Context.Ap = nextAp
	vm.Context.Fp = nextFp

	return nil
}

// It returns the current trace entry, the public memory, and the occurrence of an error
func (vm *VirtualMachine) ExecutionTrace() ([]Trace, error) {
	if !vm.config.ProofMode {
		return nil, fmt.Errorf("proof mode is off")
	}

	return vm.relocateTrace(), nil
}

func (vm *VirtualMachine) getDstAddr(instruction *a.Instruction) (mem.MemoryAddress, error) {
	var dstRegister uint64
	if instruction.DstRegister == a.Ap {
		dstRegister = vm.Context.Ap
	} else {
		dstRegister = vm.Context.Fp
	}

	addr, isOverflow := safemath.SafeOffset(dstRegister, instruction.OffDest)
	if isOverflow {
		return mem.UnknownAddress, fmt.Errorf("offset overflow: %d + %d", dstRegister, instruction.OffDest)
	}
	return mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: addr}, nil
}

func (vm *VirtualMachine) getOp0Addr(instruction *a.Instruction) (mem.MemoryAddress, error) {
	var op0Register uint64
	if instruction.Op0Register == a.Ap {
		op0Register = vm.Context.Ap
	} else {
		op0Register = vm.Context.Fp
	}

	addr, isOverflow := safemath.SafeOffset(op0Register, instruction.OffOp0)
	if isOverflow {
		return mem.UnknownAddress, fmt.Errorf("offset overflow: %d + %d", op0Register, instruction.OffOp0)
	}
	return mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: addr}, nil
}

func (vm *VirtualMachine) getOp1Addr(instruction *a.Instruction, op0Addr *mem.MemoryAddress) (mem.MemoryAddress, error) {
	var op1Address mem.MemoryAddress
	switch instruction.Op1Source {
	case a.Op0:
		// in this case Op0 is being used as an address, and must be of unwrapped as it
		op0Value, err := vm.Memory.ReadFromAddress(op0Addr)
		if err != nil {
			return mem.UnknownAddress, fmt.Errorf("cannot read op0: %w", err)
		}

		op0Address, err := op0Value.MemoryAddress()
		if err != nil {
			return mem.UnknownAddress, fmt.Errorf("op0 is not an address: %w", err)
		}
		op1Address = mem.MemoryAddress{SegmentIndex: op0Address.SegmentIndex, Offset: op0Address.Offset}
	case a.Imm:
		op1Address = vm.Context.AddressPc()
	case a.FpPlusOffOp1:
		op1Address = vm.Context.AddressFp()
	case a.ApPlusOffOp1:
		op1Address = vm.Context.AddressAp()
	}

	addr, isOverflow := safemath.SafeOffset(op1Address.Offset, instruction.OffOp1)
	if isOverflow {
		return mem.UnknownAddress, fmt.Errorf("offset overflow: %d + %d", op1Address.Offset, instruction.OffOp1)
	}
	op1Address.Offset = addr
	return op1Address, nil
}

// when there is an assertion with a substraction or division like : x = y - z
// the compiler treats it as y = x + z. This means that the VM knows the
// dstCell value and either op0Cell xor op1Cell. This function infers the
// unknow operand as well as the `res` auxiliar value
func (vm *VirtualMachine) inferOperand(
	instruction *a.Instruction, dstAddr *mem.MemoryAddress, op0Addr *mem.MemoryAddress, op1Addr *mem.MemoryAddress,
) (mem.MemoryValue, error) {
	if instruction.Opcode != a.OpCodeAssertEq ||
		(instruction.Res == a.Unconstrained) {
		return mem.MemoryValue{}, nil
	}

	dstValue, err := vm.Memory.PeekFromAddress(dstAddr)
	if err != nil {
		return mem.MemoryValue{}, fmt.Errorf("cannot read dst: %w", err)
	}

	if !dstValue.Known() {
		return mem.MemoryValue{}, nil // let computeRes try to handle it
	}

	op0Value, err := vm.Memory.PeekFromAddress(op0Addr)
	if err != nil {
		return mem.MemoryValue{}, fmt.Errorf("cannot read op0: %w", err)
	}
	op1Value, err := vm.Memory.PeekFromAddress(op1Addr)
	if err != nil {
		return mem.MemoryValue{}, fmt.Errorf("cannot read op1: %w", err)
	}

	if op0Value.Known() && op1Value.Known() {
		return mem.MemoryValue{}, nil
	}

	if instruction.Res == a.Op1 && !op1Value.Known() {
		if err = vm.Memory.WriteToAddress(op1Addr, &dstValue); err != nil {
			return mem.MemoryValue{}, err
		}
		return dstValue, nil
	}

	var knownOpValue mem.MemoryValue
	var unknownOpAddr *mem.MemoryAddress
	if op0Value.Known() {
		knownOpValue = op0Value
		unknownOpAddr = op1Addr
	} else {
		knownOpValue = op1Value
		unknownOpAddr = op0Addr
	}

	var missingVal mem.MemoryValue
	if instruction.Res == a.AddOperands {
		missingVal = mem.EmptyMemoryValueAs(dstValue.IsAddress())
		err = missingVal.Sub(&dstValue, &knownOpValue)
	} else {
		missingVal = mem.EmptyMemoryValueAsFelt()
		err = missingVal.Div(&dstValue, &knownOpValue)
	}
	if err != nil {
		return mem.MemoryValue{}, err
	}

	if err = vm.Memory.WriteToAddress(unknownOpAddr, &missingVal); err != nil {
		return mem.MemoryValue{}, err
	}
	return dstValue, nil
}

func (vm *VirtualMachine) computeRes(
	instruction *a.Instruction, op0Addr *mem.MemoryAddress, op1Addr *mem.MemoryAddress,
) (mem.MemoryValue, error) {
	switch instruction.Res {
	case a.Unconstrained:
		return mem.MemoryValue{}, nil
	case a.Op1:
		return vm.Memory.ReadFromAddress(op1Addr)
	default:
		op0, err := vm.Memory.ReadFromAddress(op0Addr)
		if err != nil {
			return mem.MemoryValue{}, fmt.Errorf("cannot read op0: %w", err)
		}

		op1, err := vm.Memory.ReadFromAddress(op1Addr)
		if err != nil {
			return mem.MemoryValue{}, fmt.Errorf("cannot read op1: %w", err)
		}

		res := mem.EmptyMemoryValueAs(op0.IsAddress() || op1.IsAddress())
		if instruction.Res == a.AddOperands {
			err = res.Add(&op0, &op1)
		} else if instruction.Res == a.MulOperands {
			err = res.Mul(&op0, &op1)
		} else {
			return mem.MemoryValue{}, fmt.Errorf("invalid res flag value: %d", instruction.Res)
		}
		return res, err
	}
}

func (vm *VirtualMachine) opcodeAssertions(
	instruction *a.Instruction,
	dstAddr *mem.MemoryAddress,
	op0Addr *mem.MemoryAddress,
	res *mem.MemoryValue,
) error {
	switch instruction.Opcode {
	case a.OpCodeCall:
		fpAddr := vm.Context.AddressFp()
		fpMv := mem.MemoryValueFromMemoryAddress(&fpAddr)
		// Store at [ap] the current fp
		if err := vm.Memory.WriteToAddress(dstAddr, &fpMv); err != nil {
			return err
		}

		apMv := mem.MemoryValueFromSegmentAndOffset(
			vm.Context.Pc.SegmentIndex,
			vm.Context.Pc.Offset+uint64(instruction.Size()),
		)
		// Write in [ap + 1] the next instruction to execute
		if err := vm.Memory.WriteToAddress(op0Addr, &apMv); err != nil {
			return err
		}
	case a.OpCodeAssertEq:
		// assert that the calculated res is stored in dst
		if err := vm.Memory.WriteToAddress(dstAddr, res); err != nil {
			return err
		}
	}
	return nil
}

func (vm *VirtualMachine) updatePc(
	instruction *a.Instruction,
	dstAddr *mem.MemoryAddress,
	op1Addr *mem.MemoryAddress,
	res *mem.MemoryValue,
) (mem.MemoryAddress, error) {
	switch instruction.PcUpdate {
	case a.PcUpdateNextInstr:
		return mem.MemoryAddress{
			SegmentIndex: vm.Context.Pc.SegmentIndex,
			Offset:       vm.Context.Pc.Offset + uint64(instruction.Size()),
		}, nil
	case a.PcUpdateJump:
		addr, err := res.MemoryAddress()
		if err != nil {
			return mem.UnknownAddress, fmt.Errorf("absolute jump: %w", err)
		}
		return *addr, nil
	case a.PcUpdateJumpRel:
		val, err := res.FieldElement()
		if err != nil {
			return mem.UnknownAddress, fmt.Errorf("relative jump: %w", err)
		}
		newPc := vm.Context.Pc
		err = newPc.Add(&newPc, val)
		return newPc, err
	case a.PcUpdateJnz:
		destMv, err := vm.Memory.ReadFromAddress(dstAddr)
		if err != nil {
			return mem.UnknownAddress, err
		}
		dest, err := destMv.FieldElement()
		if err != nil {
			return mem.UnknownAddress, err
		}

		if dest.IsZero() {
			return mem.MemoryAddress{
				SegmentIndex: vm.Context.Pc.SegmentIndex,
				Offset:       vm.Context.Pc.Offset + uint64(instruction.Size()),
			}, nil
		}

		op1Mv, err := vm.Memory.ReadFromAddress(op1Addr)
		if err != nil {
			return mem.UnknownAddress, err
		}

		val, err := op1Mv.FieldElement()
		if err != nil {
			return mem.UnknownAddress, err
		}

		newPc := vm.Context.Pc
		err = newPc.Add(&newPc, val)
		return newPc, err

	}
	return mem.UnknownAddress, fmt.Errorf("unkwon pc update value: %d", instruction.PcUpdate)
}

func (vm *VirtualMachine) updateAp(instruction *a.Instruction, res *mem.MemoryValue) (uint64, error) {
	switch instruction.ApUpdate {
	case a.SameAp:
		return vm.Context.Ap, nil
	case a.AddRes:
		apFelt := new(f.Element).SetUint64(vm.Context.Ap) // Convert ap value to felt

		resFelt, err := res.FieldElement() // Extract the f.Element from MemoryValue
		if err != nil {
			return 0, err
		}

		newAp := new(f.Element).Add(apFelt, resFelt) // Calculate newAp as the addition of apFelt and resFelt
		if !newAp.IsUint64() {
			return 0, fmt.Errorf("resulting AP value is too large to fit in uint64")
		}
		return newAp.Uint64(), nil // Return the addition as uint64
	case a.Add1:
		return vm.Context.Ap + 1, nil
	case a.Add2:
		return vm.Context.Ap + 2, nil
	}
	return 0, fmt.Errorf("cannot update ap, unknown ApUpdate flag: %d", instruction.ApUpdate)
}

func (vm *VirtualMachine) updateFp(instruction *a.Instruction, dstAddr *mem.MemoryAddress) (uint64, error) {
	switch instruction.Opcode {
	case a.OpCodeCall:
		// [ap] and [ap + 1] are written to memory
		return vm.Context.Ap + 2, nil
	case a.OpCodeRet:
		// [dst] should be a memory address of the form (executionSegment, fp - 2)
		destMv, err := vm.Memory.ReadFromAddress(dstAddr)
		if err != nil {
			return 0, err
		}

		dst, err := destMv.MemoryAddress()
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
	totalBytecode := vm.Memory.Segments[ProgramSegment].Len() + 1
	for i := range vm.Trace {
		relocatedTrace[i] = vm.Trace[i].Relocate(totalBytecode)
	}
	return relocatedTrace
}

// It returns all segments in memory but relocated as a single segment
// Each element is a pointer to a field element, if the cell was not accessed,
// nil is stored instead
func (vm *VirtualMachine) RelocateMemory() []*f.Element {
	segmentsOffsets, maxMemoryUsed := vm.Memory.RelocationOffsets()
	// the prover expect first element of the relocated memory to start at index 1,
	// this way we fill relocatedMemory starting from zero, but the actual value
	// returned has nil as its first element.
	relocatedMemory := make([]*f.Element, maxMemoryUsed)
	for i, segment := range vm.Memory.Segments {
		for j := uint64(0); j < segment.Len(); j++ {
			cell := segment.Data[j]
			if !cell.Known() {
				continue
			}

			var felt *f.Element
			if cell.IsAddress() {
				addr, _ := cell.MemoryAddress()
				felt = addr.Relocate(segmentsOffsets)
			} else {
				felt, _ = cell.FieldElement()
			}

			relocatedMemory[segmentsOffsets[i]+j] = felt
		}
	}
	return relocatedMemory
}
