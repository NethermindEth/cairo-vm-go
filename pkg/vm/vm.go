package vm

import (
	"encoding/binary"
	"fmt"
	"math"

	asmb "github.com/NethermindEth/cairo-vm-go/pkg/assembler"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
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
		"Context {pc: %s, fp: %d, ap: %d}",
		ctx.Pc,
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
	// If true, the vm outputs the trace and the relocated memory at the end of execution and finalize segments
	// in order for the prover to create a proof
	ProofMode bool
	// If true, the vm collects the relocated trace at the end of execution, without finalizing segments
	CollectTrace bool
}

type VirtualMachine struct {
	Context Context
	Memory  *mem.Memory
	Step    uint64
	Trace   []Context
	Config  VirtualMachineConfig
	// instructions cache
	instructions map[uint64]*asmb.Instruction
	// RcLimitsMin and RcLimitsMax define the range of values of instructions offsets, used for checking the number of potential range checks holes
	RcLimitsMin uint16
	RcLimitsMax uint16
}

func (vm *VirtualMachine) PrintMemory() {
	for i, _ := range vm.Memory.Segments {
		for j, cell := range vm.Memory.Segments[i].Data {
			fmt.Printf("%d:%d %s\n", i, j, cell)
		}
	}
}

// NewVirtualMachine creates a VM from the program bytecode using a specified config.
func NewVirtualMachine(
	initialContext Context, memory *mem.Memory, config VirtualMachineConfig,
) (*VirtualMachine, error) {

	// Initialize the trace if necesary
	var trace []Context
	if config.ProofMode || config.CollectTrace {
		// starknet defines a limit on the maximum number of computational steps that a transaction can contain when processed on the Starknet network.
		// https://docs.starknet.io/tools/limits-and-triggers/
		trace = make([]Context, 0, 10000000)
	}

	return &VirtualMachine{
		Context:      initialContext,
		Memory:       memory,
		Trace:        trace,
		Config:       config,
		instructions: make(map[uint64]*asmb.Instruction),
		RcLimitsMin:  math.MaxUint16,
		RcLimitsMax:  0,
	}, nil
}

func (vm *VirtualMachine) RunStep(hintRunner HintRunner) error {
	// first run the hint
	err := hintRunner.RunHint(vm)
	if err != nil {
		return err
	}

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

		instruction, err = asmb.DecodeInstruction(bytecodeInstruction)
		if err != nil {
			return fmt.Errorf("decoding instruction: %w", err)
		}
		vm.instructions[vm.Context.Pc.Offset] = instruction
	}

	// store the trace before state change
	if vm.Config.ProofMode || vm.Config.CollectTrace {
		vm.Trace = append(vm.Trace, vm.Context)
	}

	err = vm.RunInstruction(instruction)
	if err != nil {
		return fmt.Errorf("running instruction: %w", err)
	}

	vm.Step++
	return nil
}

const RC_OFFSET_BITS = 16

//go:nosplit
func (vm *VirtualMachine) RunInstruction(instruction *asmb.Instruction) error {

	var off0 int = int(instruction.OffDest) + (1 << (RC_OFFSET_BITS - 1))
	var off1 int = int(instruction.OffOp0) + (1 << (RC_OFFSET_BITS - 1))
	var off2 int = int(instruction.OffOp1) + (1 << (RC_OFFSET_BITS - 1))

	value := uint16(utils.Max(off0, utils.Max(off1, off2)))
	vm.RcLimitsMax = utils.Max(vm.RcLimitsMax, value)
	value = uint16(utils.Min(off0, utils.Min(off1, off2)))
	vm.RcLimitsMin = utils.Min(vm.RcLimitsMin, value)
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
		return fmt.Errorf("infer res: %w", err)
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

func (vm *VirtualMachine) getDstAddr(instruction *asmb.Instruction) (mem.MemoryAddress, error) {
	var dstRegister uint64
	if instruction.DstRegister == asmb.Ap {
		dstRegister = vm.Context.Ap
	} else {
		dstRegister = vm.Context.Fp
	}

	addr, isOverflow := utils.SafeOffset(dstRegister, instruction.OffDest)
	if isOverflow {
		return mem.UnknownAddress, fmt.Errorf("offset overflow: %d + %d", dstRegister, instruction.OffDest)
	}
	return mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: addr}, nil
}

func (vm *VirtualMachine) getOp0Addr(instruction *asmb.Instruction) (mem.MemoryAddress, error) {
	var op0Register uint64
	if instruction.Op0Register == asmb.Ap {
		op0Register = vm.Context.Ap
	} else {
		op0Register = vm.Context.Fp
	}

	addr, isOverflow := utils.SafeOffset(op0Register, instruction.OffOp0)
	if isOverflow {
		return mem.UnknownAddress,
			fmt.Errorf("offset overflow: %d + %d", op0Register, instruction.OffOp0)
	}
	return mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: addr}, nil
}

func (vm *VirtualMachine) getOp1Addr(instruction *asmb.Instruction, op0Addr *mem.MemoryAddress) (mem.MemoryAddress, error) {
	var op1Address mem.MemoryAddress
	switch instruction.Op1Source {
	case asmb.Op0:
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
	case asmb.Imm:
		op1Address = vm.Context.AddressPc()
	case asmb.FpPlusOffOp1:
		op1Address = vm.Context.AddressFp()
	case asmb.ApPlusOffOp1:
		op1Address = vm.Context.AddressAp()
	}

	newOffset, isOverflow := utils.SafeOffset(op1Address.Offset, instruction.OffOp1)
	if isOverflow {
		return mem.UnknownAddress, fmt.Errorf("offset overflow: %d + %d", op1Address.Offset, instruction.OffOp1)
	}
	op1Address.Offset = newOffset
	return op1Address, nil
}

// when there is an assertion with a substraction or division like : x = y - z
// the compiler treats it as y = x + z. This means that the VM knows the
// dstCell value and either op0Cell or op1Cell. This function infers the
// unknow operand as well as the `res` auxiliar value
func (vm *VirtualMachine) inferOperand(
	instruction *asmb.Instruction, dstAddr *mem.MemoryAddress, op0Addr *mem.MemoryAddress, op1Addr *mem.MemoryAddress,
) (mem.MemoryValue, error) {
	if instruction.Opcode != asmb.OpCodeAssertEq ||
		instruction.Res == asmb.Unconstrained ||
		!vm.Memory.KnownValueAtAddress(dstAddr) {
		return mem.MemoryValue{}, nil
	}

	// we known dst value is known due to previous check
	dstValue, _ := vm.Memory.PeekFromAddress(dstAddr)

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

	if instruction.Res == asmb.Op1 && !op1Value.Known() {
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
	if instruction.Res == asmb.AddOperands {
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
	instruction *asmb.Instruction, op0Addr *mem.MemoryAddress, op1Addr *mem.MemoryAddress,
) (mem.MemoryValue, error) {
	switch instruction.Res {
	case asmb.Unconstrained:
		return mem.MemoryValue{}, nil
	case asmb.Op1:
		op1, err := vm.Memory.ReadFromAddress(op1Addr)
		if err != nil {
			return mem.UnknownValue, fmt.Errorf("cannot read op1: %w", err)
		}
		return op1, nil

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
		if instruction.Res == asmb.AddOperands {
			err = res.Add(&op0, &op1)
		} else if instruction.Res == asmb.MulOperands {
			err = res.Mul(&op0, &op1)
		} else {
			return mem.MemoryValue{}, fmt.Errorf("invalid res flag value: %d", instruction.Res)
		}
		return res, err
	}
}

func (vm *VirtualMachine) opcodeAssertions(
	instruction *asmb.Instruction,
	dstAddr *mem.MemoryAddress,
	op0Addr *mem.MemoryAddress,
	res *mem.MemoryValue,
) error {
	switch instruction.Opcode {
	case asmb.OpCodeCall:
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
	case asmb.OpCodeAssertEq:
		// assert that the calculated res is stored in dst
		if err := vm.Memory.WriteToAddress(dstAddr, res); err != nil {
			return err
		}
	}
	return nil
}

func (vm *VirtualMachine) updatePc(
	instruction *asmb.Instruction,
	dstAddr *mem.MemoryAddress,
	op1Addr *mem.MemoryAddress,
	res *mem.MemoryValue,
) (mem.MemoryAddress, error) {
	switch instruction.PcUpdate {
	case asmb.PcUpdateNextInstr:
		return mem.MemoryAddress{
			SegmentIndex: vm.Context.Pc.SegmentIndex,
			Offset:       vm.Context.Pc.Offset + uint64(instruction.Size()),
		}, nil
	case asmb.PcUpdateJump:
		// both address and felt are allowed here. It can be a felt when used
		// with an immediate or a memory address holding a felt. It can be an address
		// when a memory address holds a memory address
		if addr, err := res.MemoryAddress(); err == nil {
			return *addr, nil
		} else if val, err := res.Uint64(); err == nil {
			return mem.MemoryAddress{
				SegmentIndex: vm.Context.Pc.SegmentIndex,
				Offset:       val,
			}, nil
		} else {
			return mem.UnknownAddress,
				fmt.Errorf("absolute jump: invalid jump location: %w", err)
		}

	case asmb.PcUpdateJumpRel:
		val, err := res.FieldElement()
		if err != nil {
			return mem.UnknownAddress, fmt.Errorf("relative jump: %w", err)
		}
		newPc := vm.Context.Pc
		err = newPc.Add(&newPc, val)
		return newPc, err
	case asmb.PcUpdateJnz:
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

func (vm *VirtualMachine) updateAp(instruction *asmb.Instruction, res *mem.MemoryValue) (uint64, error) {
	switch instruction.ApUpdate {
	case asmb.SameAp:
		return vm.Context.Ap, nil
	case asmb.AddRes:
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
	case asmb.Add1:
		return vm.Context.Ap + 1, nil
	case asmb.Add2:
		return vm.Context.Ap + 2, nil
	}
	return 0, fmt.Errorf("cannot update ap, unknown ApUpdate flag: %d", instruction.ApUpdate)
}

func (vm *VirtualMachine) updateFp(instruction *asmb.Instruction, dstAddr *mem.MemoryAddress) (uint64, error) {
	switch instruction.Opcode {
	case asmb.OpCodeCall:
		// [ap] and [ap + 1] are written to memory
		return vm.Context.Ap + 2, nil
	case asmb.OpCodeRet:
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

// It returns the trace after relocation, i.e, relocates pc, ap and fp for each step
// to be their real address value
func (vm *VirtualMachine) RelocateTrace(relocatedTrace *[]Trace) {
	// one is added, because prover expect that the first element to be
	// indexed on 1 instead of 0
	totalBytecode := vm.Memory.Segments[ProgramSegment].Len() + 1
	for i := range vm.Trace {
		(*relocatedTrace)[i] = vm.Trace[i].Relocate(totalBytecode)
	}
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
		for j := uint64(0); j < segment.RealLen(); j++ {
			if !segment.Data[j].Known() {
				continue
			}

			var felt *f.Element
			if segment.Data[j].IsAddress() {
				addr, _ := segment.Data[j].MemoryAddress()
				felt = addr.Relocate(segmentsOffsets)
			} else {
				felt, _ = segment.Data[j].FieldElement()
			}
			relocatedMemory[segmentsOffsets[i]+j] = felt
		}
	}
	return relocatedMemory
}

const ctxSize = 3 * 8

func EncodeTrace(trace []Trace) []byte {
	content := make([]byte, 0, len(trace)*ctxSize)
	for i := range trace {
		content = binary.LittleEndian.AppendUint64(content, trace[i].Ap)
		content = binary.LittleEndian.AppendUint64(content, trace[i].Fp)
		content = binary.LittleEndian.AppendUint64(content, trace[i].Pc)
	}
	return content
}

func DecodeTrace(content []byte) []Trace {
	trace := make([]Trace, 0, len(content)/ctxSize)
	for i := 0; i < len(content); i += ctxSize {
		trace = append(
			trace,
			Trace{
				Ap: binary.LittleEndian.Uint64(content[i : i+8]),
				Fp: binary.LittleEndian.Uint64(content[i+8 : i+16]),
				Pc: binary.LittleEndian.Uint64(content[i+16 : i+24]),
			},
		)
	}
	return trace
}

const addrSize = 8
const feltSize = 32

// Encode the relocated memory in the (address, value) form
// in a consecutive way
func EncodeMemory(memory []*f.Element) []byte {
	// Check non nil elements for optimal array size
	nonNilElms := 0
	for i := range memory {
		if memory[i] != nil {
			nonNilElms++
		}
	}
	content := make([]byte, nonNilElms*(addrSize+feltSize))

	count := 0
	for i := range memory {
		if memory[i] == nil {
			continue
		}
		// set the right content index
		j := count * (addrSize + feltSize)
		// store the address
		binary.LittleEndian.PutUint64(content[j:j+addrSize], uint64(i))
		// store the field element
		f.LittleEndian.PutElement(
			(*[32]byte)(content[j+addrSize:j+addrSize+feltSize]),
			*memory[i],
		)

		// increase the number of elements stored
		count++
	}
	return content
}

// DecodeMemory decodes an encoded memory byte array back to a memory array of felts
func DecodeMemory(content []byte) []*f.Element {
	if len(content) == 0 {
		return make([]*f.Element, 0)
	}

	// calculate the max memory index
	// the content byte array has addresses and its associated value
	// in a (address, value) form
	// but the data isn't necessarily sorted by address,
	// so we scan through the entire memory file, find the largest memory index
	// and use it to initialize the memory array below
	lastMemIndex := uint64(0)
	for i := 0; i < len(content); i += addrSize + feltSize {
		memIndex := binary.LittleEndian.Uint64(content[i : i+addrSize])
		if memIndex > lastMemIndex {
			lastMemIndex = memIndex
		}
	}

	// create the memory array with the same length as the max memory index
	memory := make([]*f.Element, lastMemIndex+1)

	// decode the content and store it in memory
	for i := 0; i < len(content); i += addrSize + feltSize {
		memIndex := binary.LittleEndian.Uint64(content[i : i+addrSize])
		felt, err := f.LittleEndian.Element((*[32]byte)(content[i+addrSize : i+addrSize+feltSize]))
		if err != nil {
			panic(err)
		}
		memory[memIndex] = &felt
	}
	return memory
}
