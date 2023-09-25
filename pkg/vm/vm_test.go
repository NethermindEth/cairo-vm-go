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

	cell, err := vm.getCellDst(&instruction)
	require.NoError(t, err)

	assert.True(t, cell.Accessed)
	assert.Equal(t, mem.MemoryValueFromInt(200), cell.Read())
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

	cell, err := vm.getCellDst(&instruction)
	require.NoError(t, err)

	assert.True(t, cell.Accessed)
	assert.Equal(t, mem.MemoryValueFromInt(123), cell.Read())
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

	cell, err := vm.getCellDst(&instruction)

	require.NoError(t, err)
	assert.True(t, cell.Accessed)
	assert.Equal(t, mem.MemoryValueFromInt(100), cell.Read())
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

	cell, err := vm.getCellDst(&instruction)
	require.NoError(t, err)
	assert.True(t, cell.Accessed)
	assert.Equal(t, mem.MemoryValueFromInt(100), cell.Read())
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

	cell, err := vm.getCellOp0(&instruction)
	require.NoError(t, err)

	assert.True(t, cell.Accessed)
	assert.Equal(t, mem.MemoryValueFromInt(123), cell.Read())
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
	const offOp1 = 1                           // target imm
	vm.Context.Pc = mem.NewMemoryAddress(0, 1) // "current instruction"

	instruction := Instruction{
		OffOp1:    offOp1,
		Op1Source: Imm,
	}

	cell, err := vm.getCellOp1(&instruction, nil)
	require.NoError(t, err)
	assert.NotNil(t, cell)

	assert.True(t, cell.Accessed)
	assert.Equal(t, mem.MemoryValueFromInt(1234), cell.Read())
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

	// Prepare vm with dummy values
	const offOp1 = 1 // target relative to op0 offset
	op0Cell := &mem.Cell{
		Accessed: true,
		Value:    mem.MemoryValueFromSegmentAndOffset(0, 2),
	}

	instruction := Instruction{
		OffOp1:    offOp1,
		Op1Source: Op0,
	}

	cell, err := vm.getCellOp1(&instruction, op0Cell)
	require.NoError(t, err)
	assert.NotNil(t, cell)

	assert.True(t, cell.Accessed)
	assert.Equal(t, mem.MemoryValueFromInt(333), cell.Read())
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

	// Prepare vm with dummy values
	const offOp1 = -1 // target relative to op0 offset
	op0Cell := &mem.Cell{
		Accessed: true,
		Value:    mem.MemoryValueFromSegmentAndOffset(0, 4),
	}

	instruction := Instruction{
		OffOp1:    offOp1,
		Op1Source: Op0,
	}

	cell, err := vm.getCellOp1(&instruction, op0Cell)
	require.NoError(t, err)
	assert.NotNil(t, cell)

	assert.True(t, cell.Accessed)
	assert.Equal(t, mem.MemoryValueFromInt(444), cell.Read())
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

	cell, err := vm.getCellOp1(&instruction, nil)
	require.NoError(t, err)
	assert.NotNil(t, cell)

	assert.True(t, cell.Accessed)
	assert.Equal(t, mem.MemoryValueFromInt(321), cell.Read())
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

	cell, err := vm.getCellOp1(&instruction, nil)
	require.NoError(t, err)
	assert.NotNil(t, cell)

	assert.True(t, cell.Accessed)
	assert.Equal(t, mem.MemoryValueFromInt(123), cell.Read())
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

	cell, err := vm.getCellOp1(&instruction, nil)
	require.NoError(t, err)
	assert.NotNil(t, cell)

	assert.True(t, cell.Accessed)
	assert.Equal(t, mem.MemoryValueFromInt(41), cell.Read())
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

	cell, err := vm.getCellOp1(&instruction, nil)
	require.NoError(t, err)
	assert.NotNil(t, cell)

	assert.True(t, cell.Accessed)
	assert.Equal(t, mem.MemoryValueFromInt(57), cell.Read())
}

func TestInferOperandSub(t *testing.T) {
	vm := defaultVirtualMachine()
	instruction := Instruction{
		Opcode: AssertEq,
		Res:    AddOperands,
	}

	dstCell := &mem.Cell{
		Accessed: true,
		Value:    mem.MemoryValueFromSegmentAndOffset(3, 15),
	}
	op1Cell := &mem.Cell{
		Accessed: true,
		Value:    mem.MemoryValueFromSegmentAndOffset(3, 7),
	}

	// unknown cell to infer
	op0Cell := &mem.Cell{}
	expectedOp0Cell := &mem.Cell{
		Accessed: true,
		Value:    mem.MemoryValueFromSegmentAndOffset(3, 8),
	}

	inferedRes, err := vm.inferOperand(&instruction, dstCell, op0Cell, op1Cell)
	require.NoError(t, err)

	assert.Equal(t, dstCell.Value, inferedRes)
	assert.Equal(t, expectedOp0Cell, op0Cell)
}

func TestComputeAddRes(t *testing.T) {
	vm := defaultVirtualMachine()

	instruction := Instruction{
		Res: AddOperands,
	}

	cellOp0 := &mem.Cell{
		Accessed: true,
		Value: mem.MemoryValueFromMemoryAddress(
			mem.NewMemoryAddress(2, 10),
		),
	}

	cellOp1 := &mem.Cell{
		Accessed: true,
		Value:    mem.MemoryValueFromInt(15),
	}

	res, err := vm.computeRes(&instruction, cellOp0, cellOp1)
	require.NoError(t, err)

	expected := mem.MemoryValueFromMemoryAddress(
		mem.NewMemoryAddress(2, 25),
	)

	assert.Equal(t, expected, res)
}

func TestOpcodeAssertionAssertEq(t *testing.T) {
	vm := defaultVirtualMachine()

	instruction := Instruction{
		Opcode: AssertEq,
	}

	dstCell := mem.Cell{}
	res := mem.MemoryValueFromMemoryAddress(mem.NewMemoryAddress(2, 10))

	err := vm.opcodeAssertions(&instruction, &dstCell, nil, res)
	require.NoError(t, err)
	assert.Equal(
		t,
		mem.Cell{
			Accessed: true,
			Value:    mem.MemoryValueFromMemoryAddress(mem.NewMemoryAddress(2, 10))},
		dstCell,
	)
}

func TestUpdatePcNextInstr(t *testing.T) {
	vm := defaultVirtualMachine()

	vm.Context.Pc = mem.NewMemoryAddress(0, 3)
	instruction := Instruction{
		PcUpdate:  NextInstr,
		Op1Source: Op0, // anything but imm
	}

	nextPc, err := vm.updatePc(&instruction, nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, mem.NewMemoryAddress(0, 4), nextPc)
}

func TestUpdatePcNextInstrImm(t *testing.T) {
	vm := defaultVirtualMachine()

	vm.Context.Pc = mem.NewMemoryAddress(0, 3)
	instruction := Instruction{
		PcUpdate:  NextInstr,
		Op1Source: Imm,
	}

	nextPc, err := vm.updatePc(&instruction, nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, mem.NewMemoryAddress(0, 5), nextPc)
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

func writeToDataSegment(vm *VirtualMachine, index uint64, value *mem.MemoryValue) {
	err := vm.MemoryManager.Memory.Write(ExecutionSegment, index, value)
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
