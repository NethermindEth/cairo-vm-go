package vm

import (
	"testing"

	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
)

func TestVMCreation(t *testing.T) {
	dummyBytecode := []*f.Element{
		newElementPtr(2),
		newElementPtr(3),
		newElementPtr(5),
		newElementPtr(7),
	}
	vm, err := NewVirtualMachine(dummyBytecode, VirtualMachineConfig{false})
	require.NoError(t, err)

	assert.Len(t, vm.MemoryManager.Memory.Segments, 2)
	assert.Len(t, vm.MemoryManager.Memory.Segments[ProgramSegment].Data, len(dummyBytecode))
	assert.Empty(t, vm.MemoryManager.Memory.Segments[ExecutionSegment].Data)
}

// todo(rodro): test all possible ways of:
// - cellDst: with ap and fp (using positive and negative offsets)
// - cellOp0: with ap and fp (using positive and negative offsets)
// - cellOp1: all four different outputs (using positive and negative offsets accordingly)
// - calculate res: verify valid mulitplication and addition. Also verify nil output when correct
// - update PC: verify all four cases. Besides, when testing relative jump (with or without conditions) that a negative relative address
// - update AP: verify all posible cases, and when Res is a negative value
// - update FP: verify all posible cases, and when Res is a negative value

func TestGetCellApDst(t *testing.T) {
	vm := defaultVirtualMachine()

	// Prepare vm with dummy values
	const offDest = 15
	const ap = 30
	vm.Context.Ap = ap
	writeToDataSegment(vm, ap+offDest, mem.MemoryValueFromInt(200))

	instruction := Instruction{
		OffDest:     offDest,
		DstRegister: Ap,
	}

	addr, err := vm.getDstAddr(&instruction)
	require.NoError(t, err)

	mv, err := vm.MemoryManager.Memory.ReadFromAddress(&addr)
	require.NoError(t, err)
	assert.True(t, mv.Known())
	assert.Equal(t, mem.MemoryValueFromInt(200), mv)
}

func TestGetCellFpDst(t *testing.T) {
	vm := defaultVirtualMachine()

	// Prepare vm with dummy values
	const (
		offDest = 5
		ap      = 30
		fp      = 20
	)
	vm.Context.Ap = ap
	vm.Context.Fp = fp
	writeToDataSegment(vm, fp+offDest, mem.MemoryValueFromInt(123))

	instruction := Instruction{
		OffDest:     offDest,
		DstRegister: Fp,
	}

	addr, err := vm.getDstAddr(&instruction)
	require.NoError(t, err)

	mv, err := vm.MemoryManager.Memory.ReadFromAddress(&addr)
	require.NoError(t, err)
	assert.True(t, mv.Known())
	assert.Equal(t, mem.MemoryValueFromInt(123), mv)
}

func TestGetCellDstApNegativeOffset(t *testing.T) {
	vm := defaultVirtualMachine()

	const (
		offDest = -2
		ap      = 12
	)
	vm.Context.Ap = ap

	writeToDataSegment(vm, ap+offDest, mem.MemoryValueFromInt(100))

	instruction := Instruction{
		OffDest:     offDest,
		DstRegister: Ap,
	}

	addr, err := vm.getDstAddr(&instruction)
	require.NoError(t, err)

	mv, err := vm.MemoryManager.Memory.ReadFromAddress(&addr)
	require.NoError(t, err)
	assert.True(t, mv.Known())
	assert.Equal(t, mem.MemoryValueFromInt(100), mv)
}

func TestGetCellDstFpNegativeOffset(t *testing.T) {
	vm := defaultVirtualMachine()

	const (
		offDest = -19
		fp      = 33
	)
	vm.Context.Fp = fp

	writeToDataSegment(vm, fp+offDest, mem.MemoryValueFromInt(100))

	instruction := Instruction{
		OffDest:     offDest,
		DstRegister: Fp,
	}

	addr, err := vm.getDstAddr(&instruction)
	require.NoError(t, err)

	mv, err := vm.MemoryManager.Memory.ReadFromAddress(&addr)
	require.NoError(t, err)
	assert.True(t, mv.Known())
	assert.Equal(t, mem.MemoryValueFromInt(100), mv)
}

func TestGetApCellOp0(t *testing.T) {
	vm := defaultVirtualMachine()

	// Prepare vm with dummy values
	const (
		offOp0 = 15
		ap     = 30
	)
	vm.Context.Ap = ap
	writeToDataSegment(vm, ap+offOp0, mem.MemoryValueFromInt(123))

	instruction := Instruction{
		OffOp0:      offOp0,
		Op0Register: Ap,
	}

	addr, err := vm.getOp0Addr(&instruction)
	require.NoError(t, err)

	mv, err := vm.MemoryManager.Memory.ReadFromAddress(&addr)
	require.NoError(t, err)
	assert.True(t, mv.Known())
	assert.Equal(t, mem.MemoryValueFromInt(123), mv)
}

func TestGetImmCellOp1(t *testing.T) {
	vm := defaultVirtualMachineWithBytecode(
		[]*f.Element{
			newElementPtr(0),    // dummy
			newElementPtr(0),    // dummy
			newElementPtr(1234), // imm
		},
	)

	// Prepare vm with dummy values
	const offOp1 = 1                                               // target imm
	vm.Context.Pc = &mem.MemoryAddress{SegmentIndex: 0, Offset: 1} // "current instruction"

	instruction := Instruction{
		OffOp1:    offOp1,
		Op1Source: Imm,
	}

	addr, err := vm.getOp1Addr(&instruction, nil)
	require.NoError(t, err)

	mv, err := vm.MemoryManager.Memory.ReadFromAddress(&addr)
	require.NoError(t, err)
	assert.True(t, mv.Known())
	assert.Equal(t, mem.MemoryValueFromInt(1234), mv)
}

func TestGetOp0PosCellOp1(t *testing.T) {
	vm := defaultVirtualMachineWithBytecode(
		[]*f.Element{
			newElementPtr(0),   // dummy
			newElementPtr(0),   // dummy
			newElementPtr(0),   // dummy
			newElementPtr(333), // op0+offset
		},
	)
	writeToDataSegment(vm, 0, mem.MemoryValueFromSegmentAndOffset(0, 2))
	op0Addr := mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: 0}

	// Prepare vm with dummy values
	const offOp1 = 1 // target relative to op0 offset
	instruction := Instruction{
		OffOp1:    offOp1,
		Op1Source: Op0,
	}

	addr, err := vm.getOp1Addr(&instruction, &op0Addr)
	require.NoError(t, err)

	mv, err := vm.MemoryManager.Memory.ReadFromAddress(&addr)
	require.NoError(t, err)
	assert.True(t, mv.Known())
	assert.Equal(t, mem.MemoryValueFromInt(333), mv)
}

func TestGetOp0NegCellOp1(t *testing.T) {
	vm := defaultVirtualMachineWithBytecode(
		[]*f.Element{
			newElementPtr(0),   // dummy
			newElementPtr(0),   // dummy
			newElementPtr(0),   // dummy
			newElementPtr(444), // op0 - offset
		},
	)
	writeToDataSegment(vm, 0, mem.MemoryValueFromSegmentAndOffset(0, 4))
	op0Addr := mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: 0}

	// Prepare vm with dummy values
	const offOp1 = -1 // target relative to op0 offset
	instruction := Instruction{
		OffOp1:    offOp1,
		Op1Source: Op0,
	}

	addr, err := vm.getOp1Addr(&instruction, &op0Addr)
	require.NoError(t, err)

	mv, err := vm.MemoryManager.Memory.ReadFromAddress(&addr)
	require.NoError(t, err)
	assert.True(t, mv.Known())
	assert.Equal(t, mem.MemoryValueFromInt(444), mv)
}

func TestGetFpPosCellOp1(t *testing.T) {
	vm := defaultVirtualMachine()

	// Prepare vm with dummy values
	const offOp1 = 2  // target relative to Fp
	vm.Context.Fp = 7 // "frame pointer"
	instruction := Instruction{
		OffOp1:    offOp1,
		Op1Source: FpPlusOffOp1,
	}

	writeToDataSegment(vm, vm.Context.Fp+2, mem.MemoryValueFromInt(321)) //Write to Execution Segment at Fp+2

	addr, err := vm.getOp1Addr(&instruction, nil)
	require.NoError(t, err)

	mv, err := vm.MemoryManager.Memory.ReadFromAddress(&addr)
	require.NoError(t, err)
	assert.True(t, mv.Known())
	assert.Equal(t, mem.MemoryValueFromInt(321), mv)
}

func TestGetFpNegCellOp1(t *testing.T) {
	vm := defaultVirtualMachine()

	// Prepare vm with dummy values
	const offOp1 = -2 // target relative to Fp
	vm.Context.Fp = 7 // "frame pointer"
	instruction := Instruction{
		OffOp1:    offOp1,
		Op1Source: FpPlusOffOp1,
	}

	writeToDataSegment(vm, vm.Context.Fp-2, mem.MemoryValueFromInt(123)) //Write to Execution Segment at Fp-2

	addr, err := vm.getOp1Addr(&instruction, nil)
	require.NoError(t, err)

	mv, err := vm.MemoryManager.Memory.ReadFromAddress(&addr)
	require.NoError(t, err)
	assert.True(t, mv.Known())
	assert.Equal(t, mem.MemoryValueFromInt(123), mv)
}

func TestGetApPosCellOp1(t *testing.T) {
	vm := defaultVirtualMachine()

	// Prepare vm with dummy values
	vm.Context.Ap = 3 // "allocation pointer"
	const offOp1 = 2  // target relative to Ap
	instruction := Instruction{
		OffOp1:    offOp1,
		Op1Source: ApPlusOffOp1,
	}
	writeToDataSegment(vm, vm.Context.Ap+2, mem.MemoryValueFromInt(41)) //Write to Execution Segment at Ap+2

	addr, err := vm.getOp1Addr(&instruction, nil)
	require.NoError(t, err)

	mv, err := vm.MemoryManager.Memory.ReadFromAddress(&addr)
	require.NoError(t, err)
	assert.True(t, mv.Known())
	assert.Equal(t, mem.MemoryValueFromInt(41), mv)
}

func TestGetApNegCellOp1(t *testing.T) {
	vm := defaultVirtualMachine()

	// Prepare vm with dummy values
	vm.Context.Ap = 3 // "allocation pointer"
	const offOp1 = -2 // target relative to Ap
	instruction := Instruction{
		OffOp1:    offOp1,
		Op1Source: ApPlusOffOp1,
	}
	writeToDataSegment(vm, vm.Context.Ap-2, mem.MemoryValueFromInt(57)) //Write to Execution Segment at Ap-2

	addr, err := vm.getOp1Addr(&instruction, nil)
	require.NoError(t, err)

	mv, err := vm.MemoryManager.Memory.ReadFromAddress(&addr)
	require.NoError(t, err)
	assert.True(t, mv.Known())
	assert.Equal(t, mem.MemoryValueFromInt(57), mv)
}

func TestInferOperandSub(t *testing.T) {
	vm := defaultVirtualMachine()
	instruction := Instruction{
		Opcode: AssertEq,
		Res:    AddOperands,
	}
	writeToDataSegment(vm, 0, mem.MemoryValueFromSegmentAndOffset(3, 15)) //destCell
	writeToDataSegment(vm, 1, mem.MemoryValueFromSegmentAndOffset(3, 7))  //op1Cell
	dstAddr := mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: 0}
	op1Addr := mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: 1}
	op0Addr := mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: 2}

	expectedOp0Vaue := mem.MemoryValueFromSegmentAndOffset(3, 8)
	inferedRes, err := vm.inferOperand(&instruction, &dstAddr, &op0Addr, &op1Addr)
	require.NoError(t, err)
	assert.Equal(t, mem.MemoryValueFromSegmentAndOffset(3, 15), inferedRes)

	op0Value, err := vm.MemoryManager.Memory.PeekFromAddress(&op0Addr)
	require.NoError(t, err)
	assert.Equal(t, expectedOp0Vaue, op0Value)
}

func TestComputeAddRes(t *testing.T) {
	vm := defaultVirtualMachine()
	writeToDataSegment(vm, 0, mem.MemoryValueFromSegmentAndOffset(2, 10)) //op0Cell
	writeToDataSegment(vm, 1, mem.MemoryValueFromInt(15))                 //op1Cell
	op0Addr := mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: 0}
	op1Addr := mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: 1}

	instruction := Instruction{
		Res: AddOperands,
	}

	res, err := vm.computeRes(&instruction, &op0Addr, &op1Addr)
	require.NoError(t, err)

	expected := mem.MemoryValueFromSegmentAndOffset(2, 25)
	assert.Equal(t, expected, res)
}

func TestOpcodeAssertionAssertEq(t *testing.T) {
	vm := defaultVirtualMachine()
	dstAddr := mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: 0}

	instruction := Instruction{
		Opcode: AssertEq,
	}

	res := mem.MemoryValueFromSegmentAndOffset(2, 10)
	err := vm.opcodeAssertions(&instruction, &dstAddr, nil, &res)
	require.NoError(t, err)

	op0Value, err := vm.MemoryManager.Memory.PeekFromAddress(&dstAddr)
	require.NoError(t, err)
	assert.Equal(t, res, op0Value)
}

func TestUpdatePcNextInstr(t *testing.T) {
	vm := defaultVirtualMachine()

	vm.Context.Pc = &mem.MemoryAddress{SegmentIndex: 0, Offset: 3}
	instruction := Instruction{
		PcUpdate:  NextInstr,
		Op1Source: Op0, // anything but imm
	}

	nextPc, err := vm.updatePc(&instruction, nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, mem.MemoryAddress{SegmentIndex: 0, Offset: 4}, nextPc)
}

func TestUpdatePcNextInstrImm(t *testing.T) {
	vm := defaultVirtualMachine()

	vm.Context.Pc = &mem.MemoryAddress{SegmentIndex: 0, Offset: 3}
	instruction := Instruction{
		PcUpdate:  NextInstr,
		Op1Source: Imm,
	}

	nextPc, err := vm.updatePc(&instruction, nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, mem.MemoryAddress{SegmentIndex: 0, Offset: 5}, nextPc)
}

func TestUpdatePcJump(t *testing.T) {
	vm := defaultVirtualMachine()

	vm.Context.Pc = &mem.MemoryAddress{SegmentIndex: 0, Offset: 3}
	jumpAddr := uint64(10)
	res := mem.MemoryValueFromSegmentAndOffset(0, jumpAddr)

	instruction := Instruction{
		PcUpdate: Jump,
	}
	nextPc, err := vm.updatePc(&instruction, nil, nil, &res)

	require.NoError(t, err)
	assert.Equal(t, mem.MemoryAddress{SegmentIndex: 0, Offset: jumpAddr}, nextPc)
}

func TestUpdatePcJumpRel(t *testing.T) {
	vm := defaultVirtualMachine()

	vm.Context.Pc = &mem.MemoryAddress{SegmentIndex: 0, Offset: 3}
	relAddr := uint64(10)
	res := mem.MemoryValueFromInt(relAddr)

	instruction := Instruction{
		PcUpdate: JumpRel,
	}
	nextPc, err := vm.updatePc(&instruction, nil, nil, &res)

	require.NoError(t, err)
	assert.Equal(t, mem.MemoryAddress{SegmentIndex: 0, Offset: 3 + relAddr}, nextPc)
}

func TestUpdatePcJnz(t *testing.T) {
	vm := defaultVirtualMachine()
	relAddr := uint64(10)
	writeToDataSegment(vm, 0, mem.MemoryValueFromInt(10))      //dstCell
	writeToDataSegment(vm, 1, mem.MemoryValueFromInt(relAddr)) //op1Cell
	dstAddr := mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: 0}
	op1Addr := mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: 1}

	vm.Context.Pc = &mem.MemoryAddress{SegmentIndex: 0, Offset: 11}
	res := mem.MemoryValueFromInt(10)
	instruction := Instruction{
		PcUpdate:  Jnz,
		Op1Source: Op0,
	}
	nextPc, err := vm.updatePc(&instruction, &dstAddr, &op1Addr, &res)

	require.NoError(t, err)
	assert.Equal(t, mem.MemoryAddress{SegmentIndex: 0, Offset: 11 + relAddr}, nextPc)
}

func TestUpdatePcJnzDstZero(t *testing.T) {
	vm := defaultVirtualMachine()
	writeToDataSegment(vm, 0, mem.MemoryValueFromInt(0)) //dstCell
	dstAddr := mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: 0}

	vm.Context.Pc = &mem.MemoryAddress{SegmentIndex: 0, Offset: 11}

	instruction := Instruction{
		PcUpdate:  Jnz,
		Op1Source: Op0,
	}
	nextPc, err := vm.updatePc(&instruction, &dstAddr, nil, nil)

	require.NoError(t, err)
	assert.Equal(t, mem.MemoryAddress{SegmentIndex: 0, Offset: 11 + 1}, nextPc)
}

func TestUpdatePcJnzDstZeroImm(t *testing.T) {
	vm := defaultVirtualMachine()
	writeToDataSegment(vm, 0, mem.MemoryValueFromInt(0)) //dstCell
	dstAddr := mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: 0}

	vm.Context.Pc = &mem.MemoryAddress{SegmentIndex: 0, Offset: 9}

	instruction := Instruction{
		PcUpdate:  Jnz,
		Op1Source: Imm,
	}
	nextPc, err := vm.updatePc(&instruction, &dstAddr, nil, nil)

	require.NoError(t, err)
	assert.Equal(t, mem.MemoryAddress{SegmentIndex: 0, Offset: 9 + 2}, nextPc)
}

func TestUpdateApAddOne(t *testing.T) {
	vm := defaultVirtualMachine()

	vm.Context.Ap = 5
	instruction := Instruction{
		Opcode:   Nop,
		ApUpdate: Add1,
	}

	nextAp, err := vm.updateAp(&instruction, nil)
	require.NoError(t, err)
	assert.Equal(t, vm.Context.Ap+1, nextAp)
}

func TestUpdateFp(t *testing.T) {
	vm := defaultVirtualMachine()

	vm.Context.Fp = 5
	instruction := Instruction{
		Opcode: Nop,
	}

	nextFp, err := vm.updateFp(&instruction, nil)
	require.NoError(t, err)
	assert.Equal(t, vm.Context.Fp, nextFp)
}

func writeToDataSegment(vm *VirtualMachine, index uint64, value mem.MemoryValue) {
	err := vm.MemoryManager.Memory.Write(ExecutionSegment, index, &value)
	if err != nil {
		panic("error in test util: writeToDataSegment")
	}
}

func defaultVirtualMachine() *VirtualMachine {
	vm, _ := NewVirtualMachine(make([]*f.Element, 0), VirtualMachineConfig{false})
	return vm
}

func defaultVirtualMachineWithBytecode(bytecode []*f.Element) *VirtualMachine {
	vm, _ := NewVirtualMachine(bytecode, VirtualMachineConfig{false})
	return vm
}

// create a pointer to an Element
func newElementPtr(val uint64) *f.Element {
	element := f.NewElement(val)
	return &element
}
