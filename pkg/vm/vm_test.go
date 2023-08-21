package vm

import (
	"github.com/stretchr/testify/require"
	"testing"

	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/assert"
)

func TestVMCreation(t *testing.T) {
	const bytecodeSize = 4
	dummyBytecode := [bytecodeSize]*f.Element{
		newElementPtr(2),
		newElementPtr(3),
		newElementPtr(5),
		newElementPtr(7),
	}

	vm, err := NewVirtualMachine(dummyBytecode[:], VirtualMachineConfig{false, false})
	require.NoError(t, err)
	assert.NotNil(t, vm)

	assert.Len(t, vm.MemoryManager.Memory.Segments, 2)
	assert.Len(t, vm.MemoryManager.Memory.Segments[programSegment].Data, bytecodeSize)
	assert.Empty(t, vm.MemoryManager.Memory.Segments[executionSegment].Data)
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
	vm, err := NewVirtualMachine(make([]*f.Element, 0), VirtualMachineConfig{false, false})
	require.NoError(t, err)
	assert.NotNil(t, vm)

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
	assert.NotNil(t, cell)

	assert.True(t, cell.Accessed)
	assert.Equal(t, mem.MemoryValueFromInt(200), cell.Read())

}

func TestGetCellFpDst(t *testing.T) {
	vm, err := NewVirtualMachine(make([]*f.Element, 0), VirtualMachineConfig{false, false})
	require.NoError(t, err)
	assert.NotNil(t, vm)

	// Prepare vm with dummy values
	const offDest = 5
	const ap = 30
	const fp = 20
	vm.Context.Ap = ap
	vm.Context.Fp = fp
	writeToDataSegment(vm, fp+offDest, mem.MemoryValueFromInt(123))

	instruction := Instruction{
		OffDest:     offDest,
		DstRegister: Fp,
	}

	cell, err := vm.getCellDst(&instruction)
	require.NoError(t, err)
	assert.NotNil(t, cell)

	assert.True(t, cell.Accessed)
	assert.Equal(t, mem.MemoryValueFromInt(123), cell.Read())
}

func TestGetApCellOp0(t *testing.T) {
	vm, err := NewVirtualMachine(make([]*f.Element, 0), VirtualMachineConfig{false, false})
	require.NoError(t, err)
	assert.NotNil(t, vm)

	// Prepare vm with dummy values
	const offOp0 = 15
	const ap = 30
	vm.Context.Ap = ap
	writeToDataSegment(vm, ap+offOp0, mem.MemoryValueFromInt(123))

	instruction := Instruction{
		OffOp0:      offOp0,
		Op0Register: Ap,
	}

	cell, err := vm.getCellOp0(&instruction)
	require.NoError(t, err)
	assert.NotNil(t, cell)

	assert.True(t, cell.Accessed)
	assert.Equal(t, mem.MemoryValueFromInt(123), cell.Read())
}

func TestGetImmCellOp1(t *testing.T) {
	vm, err := NewVirtualMachine(
		[]*f.Element{
			newElementPtr(0),    // dummy
			newElementPtr(0),    // dummy
			newElementPtr(1234), // imm
		},
		VirtualMachineConfig{false, false},
	)
	require.NoError(t, err)
	assert.NotNil(t, vm)

	// Prepare vm with dummy values
	const offOp1 = 1  // target imm
	vm.Context.Pc = 1 // "current instruction"

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

func TestInferOperandSub(t *testing.T) {
	vm, err := NewVirtualMachine(make([]*f.Element, 0), VirtualMachineConfig{false, false})
	require.NoError(t, err)
	assert.NotNil(t, vm)

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
	vm, err := NewVirtualMachine(make([]*f.Element, 0), VirtualMachineConfig{false, false})
	require.NoError(t, err)
	assert.NotNil(t, vm)

	instruction := Instruction{
		Res: AddOperands,
	}

	cellOp0 := &mem.Cell{
		Accessed: true,
		Value: mem.MemoryValueFromMemoryAddress(
			mem.CreateMemoryAddress(2, 10),
		),
	}

	cellOp1 := &mem.Cell{
		Accessed: true,
		Value:    mem.MemoryValueFromInt(15),
	}

	res, err := vm.computeRes(&instruction, cellOp0, cellOp1)
	require.NoError(t, err)

	expected := mem.MemoryValueFromMemoryAddress(
		mem.CreateMemoryAddress(2, 25),
	)

	assert.Equal(t, expected, res)
}

func (vm *VirtualMachine) TestOpcodeAssertionAssertEq(t *testing.T) {
	vm, err := NewVirtualMachine(make([]*f.Element, 0), VirtualMachineConfig{false, false})
	require.NoError(t, err)
	assert.NotNil(t, vm)

	instruction := Instruction{
		Opcode: AssertEq,
	}

	dstCell := mem.Cell{}
	res := mem.MemoryValueFromMemoryAddress(mem.CreateMemoryAddress(2, 10))

	err = vm.opcodeAssertions(&instruction, &dstCell, nil, res)
	require.NoError(t, err)
	assert.Equal(
		t,
		mem.Cell{
			Accessed: true,
			Value:    mem.MemoryValueFromMemoryAddress(mem.CreateMemoryAddress(2, 10))},
		dstCell,
	)
}

func (vm *VirtualMachine) TestUpdatePcNextInstr(t *testing.T) {
	vm, err := NewVirtualMachine(make([]*f.Element, 0), VirtualMachineConfig{false, false})
	require.NoError(t, err)
	assert.NotNil(t, vm)

	vm.Context.Pc = 3
	instruction := Instruction{
		PcUpdate: NextInstr,
	}

	nextPc, err := vm.updatePc(&instruction, nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, vm.Context.Pc+1, nextPc)
}

func (vm *VirtualMachine) TestUpdatePcNextInstrImm(t *testing.T) {
	vm, err := NewVirtualMachine(make([]*f.Element, 0), VirtualMachineConfig{false, false})
	require.NoError(t, err)
	assert.NotNil(t, vm)

	vm.Context.Pc = 3
	instruction := Instruction{
		PcUpdate:  NextInstr,
		Op1Source: Imm,
	}

	nextPc, err := vm.updatePc(&instruction, nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, vm.Context.Pc+2, nextPc)
}

func (vm *VirtualMachine) TestUpdateApAddOne(t *testing.T) {
	vm, err := NewVirtualMachine(make([]*f.Element, 0), VirtualMachineConfig{false, false})
	require.NoError(t, err)
	assert.NotNil(t, vm)

	vm.Context.Ap = 5
	instruction := Instruction{
		Opcode:   Nop,
		ApUpdate: Add1,
	}

	nextAp, err := vm.updateAp(&instruction, nil)
	require.NoError(t, err)
	assert.Equal(t, vm.Context.Ap+1, nextAp)
}

func (vm *VirtualMachine) TestUpdateFp(t *testing.T) {
	vm, err := NewVirtualMachine(make([]*f.Element, 0), VirtualMachineConfig{false, false})
	require.NoError(t, err)
	assert.NotNil(t, vm)

	vm.Context.Fp = 5
	instruction := Instruction{
		Opcode: Nop,
	}

	nextFp, err := vm.updateFp(&instruction, nil)
	require.NoError(t, err)
	assert.Equal(t, vm.Context.Fp, nextFp)
}

func writeToDataSegment(vm *VirtualMachine, index uint64, value *mem.MemoryValue) {
	err := vm.MemoryManager.Memory.Write(executionSegment, index, value)
	if err != nil {
		panic("error in test util: writeToDataSegment")
	}
}

// create a pointer to an Element
func newElementPtr(val uint64) *f.Element {
	element := f.NewElement(val)
	return &element
}
