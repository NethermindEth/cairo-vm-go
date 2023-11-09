package vm

import (
	"encoding/binary"
	"testing"

	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	a "github.com/NethermindEth/cairo-vm-go/pkg/assembler"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
)

// ===========================================
// Test State Transition Individual Components
// ===========================================

func TestGetCellApDst(t *testing.T) {
	vm := defaultVirtualMachine()

	// Prepare vm with dummy values
	const offDest = 15
	const ap = 30
	vm.Context.Ap = ap
	writeToDataSegment(vm, ap+offDest, 200)

	instruction := a.Instruction{
		OffDest:     offDest,
		DstRegister: a.Ap,
	}

	addr, err := vm.getDstAddr(&instruction)
	require.NoError(t, err)

	mv, err := vm.Memory.ReadFromAddress(&addr)
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
	writeToDataSegment(vm, fp+offDest, 123)

	instruction := a.Instruction{
		OffDest:     offDest,
		DstRegister: a.Fp,
	}

	addr, err := vm.getDstAddr(&instruction)
	require.NoError(t, err)

	mv, err := vm.Memory.ReadFromAddress(&addr)
	require.NoError(t, err)
	assert.True(t, mv.Known())
	assert.Equal(t, mem.MemoryValueFromInt(123), mv)
}

func TestGetCellApDstWithDifferentOffsets(t *testing.T) {
	vm := defaultVirtualMachine()
	offsets := []int{-10, -5, 0, 5, 10}

	for _, offset := range offsets {
		const ap = 30
		vm.Context.Ap = ap

		writeToDataSegment(vm, uint64(ap+offset), 200)

		instruction := a.Instruction{
			OffDest:     int16(offset),
			DstRegister: a.Ap,
		}

		addr, err := vm.getDstAddr(&instruction)
		require.NoError(t, err)

		mv, err := vm.Memory.ReadFromAddress(&addr)
		require.NoError(t, err)
		assert.True(t, mv.Known())
		assert.Equal(t, mem.MemoryValueFromInt(200), mv)
	}
}

func TestGetCellDstApNegativeOffset(t *testing.T) {
	vm := defaultVirtualMachine()

	const (
		offDest = -2
		ap      = 12
	)
	vm.Context.Ap = ap

	writeToDataSegment(vm, ap+offDest, 100)

	instruction := a.Instruction{
		OffDest:     offDest,
		DstRegister: a.Ap,
	}

	addr, err := vm.getDstAddr(&instruction)
	require.NoError(t, err)

	mv, err := vm.Memory.ReadFromAddress(&addr)
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

	writeToDataSegment(vm, fp+offDest, 100)

	instruction := a.Instruction{
		OffDest:     offDest,
		DstRegister: a.Fp,
	}

	addr, err := vm.getDstAddr(&instruction)
	require.NoError(t, err)

	mv, err := vm.Memory.ReadFromAddress(&addr)
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
	writeToDataSegment(vm, ap+offOp0, 123)

	instruction := a.Instruction{
		OffOp0:      offOp0,
		Op0Register: a.Ap,
	}

	addr, err := vm.getOp0Addr(&instruction)
	require.NoError(t, err)

	mv, err := vm.Memory.ReadFromAddress(&addr)
	require.NoError(t, err)
	assert.True(t, mv.Known())
	assert.Equal(t, mem.MemoryValueFromInt(123), mv)
}

func TestGetApCellOp0NegOff(t *testing.T) {
	// Op0 & Ap & Negative case
	vm := defaultVirtualMachine()

	// Prepare vm with dummy values
	const (
		offOp0 = -12
		ap     = 20
	)
	vm.Context.Ap = ap
	writeToDataSegment(vm, ap+offOp0, 155)

	instruction := a.Instruction{
		OffOp0:      offOp0,
		Op0Register: a.Ap,
	}

	addr, err := vm.getOp0Addr(&instruction)
	require.NoError(t, err)

	mv, err := vm.Memory.ReadFromAddress(&addr)
	require.NoError(t, err)
	assert.True(t, mv.Known())
	assert.Equal(t, mem.MemoryValueFromInt(155), mv)
}

func TestGetFpCellOp0(t *testing.T) {
	// Op0 & Fp & Positive case
	vm := defaultVirtualMachine()

	// Prepare vm with dummy values
	const (
		offOp0 = 26
		fp     = 74
	)
	vm.Context.Fp = fp
	writeToDataSegment(vm, fp+offOp0, 365)

	instruction := a.Instruction{
		OffOp0:      offOp0,
		Op0Register: a.Fp,
	}

	addr, err := vm.getOp0Addr(&instruction)
	require.NoError(t, err)

	mv, err := vm.Memory.ReadFromAddress(&addr)
	require.NoError(t, err)
	assert.True(t, mv.Known())
	assert.Equal(t, mem.MemoryValueFromInt(365), mv)
}

func TestGetFpCellOp0NegOff(t *testing.T) {
	// Op0 & Fp & Negative case
	vm := defaultVirtualMachine()

	// Prepare vm with dummy values
	const (
		offOp0 = -15
		fp     = 67
	)
	vm.Context.Fp = fp
	writeToDataSegment(vm, fp+offOp0, 286)

	instruction := a.Instruction{
		OffOp0:      offOp0,
		Op0Register: a.Fp,
	}

	addr, err := vm.getOp0Addr(&instruction)
	require.NoError(t, err)

	mv, err := vm.Memory.ReadFromAddress(&addr)
	require.NoError(t, err)
	assert.True(t, mv.Known())
	assert.Equal(t, mem.MemoryValueFromInt(286), mv)
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
	const offOp1 = 1                                              // target imm
	vm.Context.Pc = mem.MemoryAddress{SegmentIndex: 0, Offset: 1} // "current instruction"

	instruction := a.Instruction{
		OffOp1:    offOp1,
		Op1Source: a.Imm,
	}

	addr, err := vm.getOp1Addr(&instruction, nil)
	require.NoError(t, err)

	mv, err := vm.Memory.ReadFromAddress(&addr)
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
	writeToDataSegment(vm, 0, &mem.MemoryAddress{SegmentIndex: 0, Offset: 2})
	op0Addr := mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: 0}

	// Prepare vm with dummy values
	const offOp1 = 1 // target relative to op0 offset
	instruction := a.Instruction{
		OffOp1:    offOp1,
		Op1Source: a.Op0,
	}

	addr, err := vm.getOp1Addr(&instruction, &op0Addr)
	require.NoError(t, err)

	mv, err := vm.Memory.ReadFromAddress(&addr)
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
	writeToDataSegment(vm, 0, &mem.MemoryAddress{SegmentIndex: 0, Offset: 4})
	op0Addr := mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: 0}

	// Prepare vm with dummy values
	const offOp1 = -1 // target relative to op0 offset
	instruction := a.Instruction{
		OffOp1:    offOp1,
		Op1Source: a.Op0,
	}

	addr, err := vm.getOp1Addr(&instruction, &op0Addr)
	require.NoError(t, err)

	mv, err := vm.Memory.ReadFromAddress(&addr)
	require.NoError(t, err)
	assert.True(t, mv.Known())
	assert.Equal(t, mem.MemoryValueFromInt(444), mv)
}

func TestGetFpPosCellOp1(t *testing.T) {
	vm := defaultVirtualMachine()

	// Prepare vm with dummy values
	const offOp1 = 2  // target relative to Fp
	vm.Context.Fp = 7 // "frame pointer"
	instruction := a.Instruction{
		OffOp1:    offOp1,
		Op1Source: a.FpPlusOffOp1,
	}

	writeToDataSegment(vm, vm.Context.Fp+2, 321) //Write to Execution Segment at Fp+2

	addr, err := vm.getOp1Addr(&instruction, nil)
	require.NoError(t, err)

	mv, err := vm.Memory.ReadFromAddress(&addr)
	require.NoError(t, err)
	assert.True(t, mv.Known())
	assert.Equal(t, mem.MemoryValueFromInt(321), mv)
}

func TestGetFpNegCellOp1(t *testing.T) {
	vm := defaultVirtualMachine()

	// Prepare vm with dummy values
	const offOp1 = -2 // target relative to Fp
	vm.Context.Fp = 7 // "frame pointer"
	instruction := a.Instruction{
		OffOp1:    offOp1,
		Op1Source: a.FpPlusOffOp1,
	}

	writeToDataSegment(vm, vm.Context.Fp-2, 123) //Write to Execution Segment at Fp-2

	addr, err := vm.getOp1Addr(&instruction, nil)
	require.NoError(t, err)

	mv, err := vm.Memory.ReadFromAddress(&addr)
	require.NoError(t, err)
	assert.True(t, mv.Known())
	assert.Equal(t, mem.MemoryValueFromInt(123), mv)
}

func TestGetApPosCellOp1(t *testing.T) {
	vm := defaultVirtualMachine()

	// Prepare vm with dummy values
	vm.Context.Ap = 3 // "allocation pointer"
	const offOp1 = 2  // target relative to Ap
	instruction := a.Instruction{
		OffOp1:    offOp1,
		Op1Source: a.ApPlusOffOp1,
	}
	writeToDataSegment(vm, vm.Context.Ap+2, 41) //Write to Execution Segment at Ap+2

	addr, err := vm.getOp1Addr(&instruction, nil)
	require.NoError(t, err)

	mv, err := vm.Memory.ReadFromAddress(&addr)
	require.NoError(t, err)
	assert.True(t, mv.Known())
	assert.Equal(t, mem.MemoryValueFromInt(41), mv)
}

func TestGetApNegCellOp1(t *testing.T) {
	vm := defaultVirtualMachine()

	// Prepare vm with dummy values
	vm.Context.Ap = 3 // "allocation pointer"
	const offOp1 = -2 // target relative to Ap
	instruction := a.Instruction{
		OffOp1:    offOp1,
		Op1Source: a.ApPlusOffOp1,
	}
	writeToDataSegment(vm, vm.Context.Ap-2, 57) //Write to Execution Segment at Ap-2

	addr, err := vm.getOp1Addr(&instruction, nil)
	require.NoError(t, err)

	mv, err := vm.Memory.ReadFromAddress(&addr)
	require.NoError(t, err)
	assert.True(t, mv.Known())
	assert.Equal(t, mem.MemoryValueFromInt(57), mv)
}

func TestInferOperandSub(t *testing.T) {
	vm := defaultVirtualMachine()
	instruction := a.Instruction{
		Opcode: a.OpCodeAssertEq,
		Res:    a.AddOperands,
	}
	writeToDataSegment(vm, 0, &mem.MemoryAddress{SegmentIndex: 3, Offset: 15}) //destCell
	writeToDataSegment(vm, 1, &mem.MemoryAddress{SegmentIndex: 3, Offset: 7})  //op1Cell
	dstAddr := mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: 0}
	op1Addr := mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: 1}
	op0Addr := mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: 2}

	expectedOp0Vaue := mem.MemoryValueFromSegmentAndOffset(3, 8)
	inferedRes, err := vm.inferOperand(&instruction, &dstAddr, &op0Addr, &op1Addr)
	require.NoError(t, err)
	assert.Equal(t, mem.MemoryValueFromSegmentAndOffset(3, 15), inferedRes)

	op0Value, err := vm.Memory.PeekFromAddress(&op0Addr)
	require.NoError(t, err)
	assert.Equal(t, expectedOp0Vaue, op0Value)
}

func TestInferResOp1(t *testing.T) {
	vm := defaultVirtualMachine()
	instruction := a.Instruction{
		Opcode: a.OpCodeAssertEq,
		Res:    a.Op1,
	}
	writeToDataSegment(vm, 0, 1337) //destCell
	dstAddr := mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: 0}
	op1Addr := mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: 1}
	op0Addr := mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: 2}

	expectedOp1Vaue := mem.MemoryValueFromInt(1337)
	inferedRes, err := vm.inferOperand(&instruction, &dstAddr, &op0Addr, &op1Addr)
	require.NoError(t, err)
	assert.Equal(t, mem.MemoryValueFromInt(1337), inferedRes)

	op0Value, err := vm.Memory.PeekFromAddress(&op1Addr)
	require.NoError(t, err)
	assert.Equal(t, expectedOp1Vaue, op0Value)
}

func TestComputeResUnconstrained(t *testing.T) {
	vm := defaultVirtualMachine()
	instruction := a.Instruction{Res: a.Unconstrained}
	res, err := vm.computeRes(&instruction, nil, nil)
	require.NoError(t, err)
	require.False(t, res.Known())
}

func TestComputeResOp1(t *testing.T) {
	vm := defaultVirtualMachine()
	instruction := a.Instruction{Res: a.Op1}

	writeToDataSegment(vm, 3, 15)
	op1Addr := mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: 3}

	res, err := vm.computeRes(&instruction, nil, &op1Addr)
	require.NoError(t, err)

	expected := mem.MemoryValueFromInt(15)
	assert.Equal(t, expected, res)
}

func TestComputeAddResAddrToFelt(t *testing.T) {
	vm := defaultVirtualMachine()
	instruction := a.Instruction{Res: a.AddOperands}

	op0Addr := writeToDataSegment(
		vm, 3, &mem.MemoryAddress{SegmentIndex: 2, Offset: 10},
	)
	op1Addr := writeToDataSegment(vm, 8, 15)

	res, err := vm.computeRes(&instruction, &op0Addr, &op1Addr)
	require.NoError(t, err)

	expected := mem.MemoryValueFromSegmentAndOffset(2, 25)
	assert.Equal(t, expected, res)
}

func TestComputeAddResFeltToAddr(t *testing.T) {
	vm := defaultVirtualMachine()
	instruction := a.Instruction{Res: a.AddOperands}

	op0Addr := writeToDataSegment(vm, 2, 8)
	op1Addr := writeToDataSegment(
		vm, 5, &mem.MemoryAddress{SegmentIndex: 2, Offset: 7},
	)

	res, err := vm.computeRes(&instruction, &op0Addr, &op1Addr)
	require.NoError(t, err)
	expected := mem.MemoryValueFromSegmentAndOffset(2, 15)
	assert.Equal(t, expected, res)
}

func TestComputeAddResBothAddrs(t *testing.T) {
	vm := defaultVirtualMachine()
	instruction := a.Instruction{Res: a.AddOperands}

	op0Addr := writeToDataSegment(
		vm, 3, &mem.MemoryAddress{SegmentIndex: 2, Offset: 10},
	)
	op1Addr := writeToDataSegment(
		vm, 4, &mem.MemoryAddress{SegmentIndex: 2, Offset: 15},
	)

	_, err := vm.computeRes(&instruction, &op0Addr, &op1Addr)
	require.Error(t, err) // Expecting an error since adding two addresses is not allowed
}

func TestComputeAddResBothFelts(t *testing.T) {
	vm := defaultVirtualMachine()
	instruction := a.Instruction{Res: a.AddOperands}

	op0Addr := writeToDataSegment(vm, 3, 10)
	op1Addr := writeToDataSegment(vm, 4, 15)

	res, err := vm.computeRes(&instruction, &op0Addr, &op1Addr)
	require.NoError(t, err)
	expected := mem.MemoryValueFromInt(25)
	assert.Equal(t, expected, res)
}

// Felt should be Positive or Negative. Thus four test cases
func TestComputeMulResPosToPosFelt(t *testing.T) {
	//Positive Felt to Positive Felt compute
	vm := defaultVirtualMachine()
	instruction := a.Instruction{Res: a.MulOperands}

	op0Addr := writeToDataSegment(vm, 3, 10)
	op1Addr := writeToDataSegment(vm, 4, 15)

	res, err := vm.computeRes(&instruction, &op0Addr, &op1Addr)
	require.NoError(t, err)
	expected := mem.MemoryValueFromInt(150)
	assert.Equal(t, expected, res)
}

func TestComputeMulResNegToPosFelts(t *testing.T) {
	vm := defaultVirtualMachine()
	instruction := a.Instruction{Res: a.MulOperands}
	//Negative to Positive
	op0Addr := writeToDataSegment(vm, 3, -10)
	op1Addr := writeToDataSegment(vm, 4, 15)

	res, err := vm.computeRes(&instruction, &op0Addr, &op1Addr)
	require.NoError(t, err)
	expected := mem.MemoryValueFromInt(-150)
	assert.Equal(t, expected, res)
}

func TestComputeMulResPosToNegFelt(t *testing.T) {
	vm := defaultVirtualMachine()
	instruction := a.Instruction{Res: a.MulOperands}
	//Positive to Negative
	op0Addr := writeToDataSegment(vm, 3, 10)
	op1Addr := writeToDataSegment(vm, 4, -15)

	res, err := vm.computeRes(&instruction, &op0Addr, &op1Addr)
	require.NoError(t, err)
	expected := mem.MemoryValueFromInt(-150)
	assert.Equal(t, expected, res)
}

func TestComputeMulResNegToNegFelt(t *testing.T) {
	vm := defaultVirtualMachine()
	instruction := a.Instruction{Res: a.MulOperands}
	//Netagive to Negative
	op0Addr := writeToDataSegment(vm, 3, -10)
	op1Addr := writeToDataSegment(vm, 4, -15)

	res, err := vm.computeRes(&instruction, &op0Addr, &op1Addr)
	require.NoError(t, err)
	expected := mem.MemoryValueFromInt(150)
	assert.Equal(t, expected, res)
}

// Multiplication does not involve addresses
// three failing cases
func TestComputeMulResAddrToFelt(t *testing.T) {
	vm := defaultVirtualMachine()
	instruction := a.Instruction{Res: a.MulOperands}

	op0Addr := writeToDataSegment(
		vm, 3, &mem.MemoryAddress{SegmentIndex: 2, Offset: 10},
	)
	op1Addr := writeToDataSegment(vm, 4, 15)

	_, err := vm.computeRes(&instruction, &op0Addr, &op1Addr)
	require.Error(t, err) // Expecting an error since multiplying an address with a felt is not allowed
}

func TestComputeMulResFeltToAddr(t *testing.T) {
	vm := defaultVirtualMachine()
	instruction := a.Instruction{Res: a.MulOperands}

	op0Addr := writeToDataSegment(vm, 3, 10)
	op1Addr := writeToDataSegment(
		vm, 4, &mem.MemoryAddress{SegmentIndex: 2, Offset: 15},
	)

	_, err := vm.computeRes(&instruction, &op0Addr, &op1Addr)
	require.Error(t, err)
}

func TestComputeMulResBothAddrs(t *testing.T) {
	vm := defaultVirtualMachine()
	instruction := a.Instruction{Res: a.MulOperands}

	op0Addr := writeToDataSegment(
		vm, 3, &mem.MemoryAddress{SegmentIndex: 2, Offset: 10},
	)
	op1Addr := writeToDataSegment(
		vm, 4, &mem.MemoryAddress{SegmentIndex: 2, Offset: 15},
	)

	_, err := vm.computeRes(&instruction, &op0Addr, &op1Addr)
	require.Error(t, err) // Expecting an error since multiplying two addresses is not allowed
}

func TestOpcodeAssertionAssertEq(t *testing.T) {
	vm := defaultVirtualMachine()
	dstAddr := mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: 0}

	instruction := a.Instruction{
		Opcode: a.OpCodeAssertEq,
	}

	res := mem.MemoryValueFromSegmentAndOffset(2, 10)
	err := vm.opcodeAssertions(&instruction, &dstAddr, nil, &res)
	require.NoError(t, err)

	op0Value, err := vm.Memory.PeekFromAddress(&dstAddr)
	require.NoError(t, err)
	assert.Equal(t, res, op0Value)
}

func TestUpdatePcNextInstr(t *testing.T) {
	vm := defaultVirtualMachine()

	vm.Context.Pc = mem.MemoryAddress{SegmentIndex: 0, Offset: 3}
	instruction := a.Instruction{
		PcUpdate:  a.PcUpdateNextInstr,
		Op1Source: a.Op0, // anything but imm
	}

	nextPc, err := vm.updatePc(&instruction, nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, mem.MemoryAddress{SegmentIndex: 0, Offset: 4}, nextPc)
}

func TestUpdatePcNextInstrImm(t *testing.T) {
	vm := defaultVirtualMachine()

	vm.Context.Pc = mem.MemoryAddress{SegmentIndex: 0, Offset: 3}
	instruction := a.Instruction{
		PcUpdate:  a.PcUpdateNextInstr,
		Op1Source: a.Imm,
	}

	nextPc, err := vm.updatePc(&instruction, nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, mem.MemoryAddress{SegmentIndex: 0, Offset: 5}, nextPc)
}

func TestUpdatePcJump(t *testing.T) {
	vm := defaultVirtualMachine()

	instruction := a.Instruction{
		PcUpdate: a.PcUpdateJump,
	}

	jumpAddrs := []mem.MemoryValue{
		mem.MemoryValueFromInt(10),
		mem.MemoryValueFromSegmentAndOffset(4, 5),
	}
	expectedPcs := []mem.MemoryAddress{
		{SegmentIndex: 0, Offset: 10},
		{SegmentIndex: 4, Offset: 5},
	}

	for i := range jumpAddrs {
		vm.Context.Pc = mem.MemoryAddress{SegmentIndex: 0, Offset: 3}
		jumpAddr := jumpAddrs[i]
		expectedPc := expectedPcs[i]

		nextPc, err := vm.updatePc(&instruction, nil, nil, &jumpAddr)
		require.NoError(t, err)
		assert.Equal(t, expectedPc, nextPc)
	}
}

func TestUpdatePcJumpRel(t *testing.T) {
	vm := defaultVirtualMachine()

	vm.Context.Pc = mem.MemoryAddress{SegmentIndex: 0, Offset: 3}
	relAddr := uint64(10)
	res := mem.MemoryValueFromInt(relAddr)

	instruction := a.Instruction{
		PcUpdate: a.PcUpdateJumpRel,
	}
	nextPc, err := vm.updatePc(&instruction, nil, nil, &res)

	require.NoError(t, err)
	assert.Equal(t, mem.MemoryAddress{SegmentIndex: 0, Offset: 3 + relAddr}, nextPc)
}

func TestUpdatePcJnz(t *testing.T) {
	vm := defaultVirtualMachine()
	relAddr := uint64(10)
	writeToDataSegment(vm, 0, 10)      //dstCell
	writeToDataSegment(vm, 1, relAddr) //op1Cell
	dstAddr := mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: 0}
	op1Addr := mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: 1}

	vm.Context.Pc = mem.MemoryAddress{SegmentIndex: 0, Offset: 11}
	res := mem.MemoryValueFromInt(10)
	instruction := a.Instruction{
		PcUpdate:  a.PcUpdateJnz,
		Op1Source: a.Op0,
	}
	nextPc, err := vm.updatePc(&instruction, &dstAddr, &op1Addr, &res)

	require.NoError(t, err)
	assert.Equal(t, mem.MemoryAddress{SegmentIndex: 0, Offset: 11 + relAddr}, nextPc)
}

func TestUpdatePcJnzDstZero(t *testing.T) {
	vm := defaultVirtualMachine()
	writeToDataSegment(vm, 0, 0) //dstCell
	dstAddr := mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: 0}

	vm.Context.Pc = mem.MemoryAddress{SegmentIndex: 0, Offset: 11}

	instruction := a.Instruction{
		PcUpdate:  a.PcUpdateJnz,
		Op1Source: a.Op0,
	}
	nextPc, err := vm.updatePc(&instruction, &dstAddr, nil, nil)

	require.NoError(t, err)
	assert.Equal(t, mem.MemoryAddress{SegmentIndex: 0, Offset: 11 + 1}, nextPc)
}

func TestUpdatePcJnzDstZeroImm(t *testing.T) {
	vm := defaultVirtualMachine()
	writeToDataSegment(vm, 0, 0) //dstCell
	dstAddr := mem.MemoryAddress{SegmentIndex: ExecutionSegment, Offset: 0}

	vm.Context.Pc = mem.MemoryAddress{SegmentIndex: 0, Offset: 9}

	instruction := a.Instruction{
		PcUpdate:  a.PcUpdateJnz,
		Op1Source: a.Imm,
	}
	nextPc, err := vm.updatePc(&instruction, &dstAddr, nil, nil)

	require.NoError(t, err)
	assert.Equal(t, mem.MemoryAddress{SegmentIndex: 0, Offset: 9 + 2}, nextPc)
}

func TestUpdateApSameAp(t *testing.T) {
	vm := defaultVirtualMachine()

	vm.Context.Ap = 5
	instruction := a.Instruction{
		Opcode:   a.OpCodeNop,
		ApUpdate: a.SameAp,
	}

	nextAp, err := vm.updateAp(&instruction, nil)
	require.NoError(t, err)
	assert.Equal(t, vm.Context.Ap, nextAp)
}

func TestUpdateApAddImmPos(t *testing.T) {
	vm := defaultVirtualMachine()

	vm.Context.Ap = 5
	instruction := a.Instruction{
		Opcode:   a.OpCodeNop,
		ApUpdate: a.AddRes,
	}

	res := mem.MemoryValueFromInt(7)

	nextAp, err := vm.updateAp(&instruction, &res)
	require.NoError(t, err)
	assert.Equal(t, uint64(12), nextAp)
}

func TestUpdateApAddImmNeg(t *testing.T) {
	vm := defaultVirtualMachine()

	vm.Context.Ap = 10
	instruction := a.Instruction{
		Opcode:   a.OpCodeNop,
		ApUpdate: a.AddRes,
	}

	res := mem.MemoryValueFromInt(-3)

	nextAp, err := vm.updateAp(&instruction, &res)
	require.NoError(t, err)
	assert.Equal(t, uint64(7), nextAp)
}

func TestUpdateApAddOne(t *testing.T) {
	vm := defaultVirtualMachine()

	vm.Context.Ap = 5
	instruction := a.Instruction{
		Opcode:   a.OpCodeNop,
		ApUpdate: a.Add1,
	}

	nextAp, err := vm.updateAp(&instruction, nil)
	require.NoError(t, err)
	assert.Equal(t, vm.Context.Ap+1, nextAp)
}

func TestUpdateApAddTwo(t *testing.T) {
	vm := defaultVirtualMachine()

	vm.Context.Ap = 5
	instruction := a.Instruction{
		Opcode:   a.OpCodeNop,
		ApUpdate: a.Add2,
	}

	nextAp, err := vm.updateAp(&instruction, nil)
	require.NoError(t, err)
	assert.Equal(t, vm.Context.Ap+2, nextAp)
}

func TestUpdateFp(t *testing.T) {
	vm := defaultVirtualMachine()

	vm.Context.Fp = 5
	instruction := a.Instruction{
		Opcode: a.OpCodeNop,
	}

	nextFp, err := vm.updateFp(&instruction, nil)
	require.NoError(t, err)
	assert.Equal(t, vm.Context.Fp, nextFp)
}

// =====================================
// Test State Transition Full Execution
// =====================================

type noHintRunner struct{}

func (r *noHintRunner) RunHint(_ *VirtualMachine) error {
	return nil
}

func TestRunStepInstructions(t *testing.T) {
	hintrunner := noHintRunner{}
	setInitialReg := func(vm *VirtualMachine, regvals ...uint64) {
		if len(regvals) != 3 {
			panic("expected three register values")
		}
		vm.Context.Ap = regvals[0]
		vm.Context.Fp = regvals[1]
		vm.Context.Pc = mem.MemoryAddress{SegmentIndex: 0, Offset: regvals[2]}
	}

	t.Run("assign right to left;", func(t *testing.T) {
		vm := defaultVirtualMachineWithCode("[ap + 1] = [ap];")
		setInitialReg(vm, 1, 1, 0)

		writeToDataSegment(vm, vm.Context.Ap, 2)

		err := vm.RunStep(&hintrunner)
		require.NoError(t, err)

		mv, err := vm.Memory.Read(ExecutionSegment, vm.Context.Ap+1)
		require.NoError(t, err)
		assert.Equal(t, mem.MemoryValueFromInt(2), mv)
	})

	t.Run("assign left to right", func(t *testing.T) {
		vm := defaultVirtualMachineWithCode("[ap + 1] = [ap];")
		setInitialReg(vm, 1, 1, 0)

		writeToDataSegment(vm, vm.Context.Ap+1, 2)

		err := vm.RunStep(&hintrunner)
		require.NoError(t, err)

		mv, err := vm.Memory.Read(ExecutionSegment, vm.Context.Ap)
		require.NoError(t, err)
		assert.Equal(t, mem.MemoryValueFromInt(2), mv)
	})

	t.Run("addition", func(t *testing.T) {
		vm := defaultVirtualMachineWithCode("[ap] = [ap - 1] + [fp];")
		setInitialReg(vm, 3, 1, 0)

		writeToDataSegment(vm, vm.Context.Fp, 10)
		writeToDataSegment(vm, vm.Context.Ap-1, 5)

		err := vm.RunStep(&hintrunner)
		require.NoError(t, err)

		mv, err := vm.Memory.Read(ExecutionSegment, vm.Context.Ap)
		require.NoError(t, err)
		assert.Equal(t, mem.MemoryValueFromInt(15), mv)
	})

	t.Run("substraction", func(t *testing.T) {
		vm := defaultVirtualMachineWithCode("[ap] = [ap - 1] + [fp];")
		setInitialReg(vm, 3, 1, 0)

		writeToDataSegment(vm, vm.Context.Fp, 10)
		writeToDataSegment(vm, vm.Context.Ap, 5)

		err := vm.RunStep(&hintrunner)
		require.NoError(t, err)

		mv, err := vm.Memory.Read(ExecutionSegment, vm.Context.Ap-1)
		require.NoError(t, err)
		assert.Equal(t, mem.MemoryValueFromInt(-5), mv)
	})

	t.Run("multiplication", func(t *testing.T) {
		vm := defaultVirtualMachineWithCode("[ap] = [ap - 1] * [fp];")
		setInitialReg(vm, 3, 1, 0)

		writeToDataSegment(vm, vm.Context.Fp, 5)
		writeToDataSegment(vm, vm.Context.Ap-1, 10)

		err := vm.RunStep(&hintrunner)
		require.NoError(t, err)

		mv, err := vm.Memory.Read(ExecutionSegment, vm.Context.Ap)
		require.NoError(t, err)
		assert.Equal(t, mem.MemoryValueFromInt(50), mv)
	})

	t.Run("division", func(t *testing.T) {
		vm := defaultVirtualMachineWithCode("[ap] = [ap + 1] * [fp];")
		setInitialReg(vm, 3, 1, 0)

		writeToDataSegment(vm, vm.Context.Fp, 2)
		writeToDataSegment(vm, vm.Context.Ap, 10)

		err := vm.RunStep(&hintrunner)
		require.NoError(t, err)

		mv, err := vm.Memory.Read(ExecutionSegment, vm.Context.Ap+1)
		require.NoError(t, err)
		assert.Equal(t, mem.MemoryValueFromInt(5), mv)
	})

	t.Run("test advancing ap with expression", func(t *testing.T) {
		vm := defaultVirtualMachineWithCode("ap += [fp + 4] + [fp];")
		setInitialReg(vm, 1, 1, 0)

		writeToDataSegment(vm, vm.Context.Fp+4, 5)
		writeToDataSegment(vm, vm.Context.Fp, 5)

		err := vm.RunStep(&hintrunner)
		require.NoError(t, err)

		assert.Equal(t, vm.Context.Ap, uint64(11))
	})

	t.Run("test advancing ap with immediate", func(t *testing.T) {
		vm := defaultVirtualMachineWithCode("ap += 123;")
		setInitialReg(vm, 1, 1, 0)

		err := vm.RunStep(&hintrunner)
		require.NoError(t, err)

		assert.Equal(t, vm.Context.Ap, uint64(124))
	})

	t.Run("test abs jump with immediate", func(t *testing.T) {
		vm := defaultVirtualMachineWithCode("jmp abs 15;")
		setInitialReg(vm, 1, 1, 0)

		err := vm.RunStep(&hintrunner)
		require.NoError(t, err)

		assert.Equal(t, vm.Context.Pc.Offset, uint64(15))
	})
	
	t.Run("test abs jump with address", func(t *testing.T) {
		vm := defaultVirtualMachineWithCode("jmp abs [ap];")
		setInitialReg(vm, 1, 1, 0)

		writeToDataSegment(vm, vm.Context.Ap, &mem.MemoryAddress{SegmentIndex: 15, Offset: 18})

		err := vm.RunStep(&hintrunner)
		require.NoError(t, err)

		assert.Equal(t, vm.Context.Pc, mem.MemoryAddress{SegmentIndex: 15, Offset: 18})
	})

	t.Run("test rel jump with expression", func(t *testing.T) {
		vm := defaultVirtualMachineWithCode("jmp rel [ap + 1] + [fp];")
		setInitialReg(vm, 1, 1, 0)

		writeToDataSegment(vm, vm.Context.Ap+1, 1)
		writeToDataSegment(vm, vm.Context.Fp, 7)

		err := vm.RunStep(&hintrunner)
		require.NoError(t, err)

		assert.Equal(t, vm.Context.Pc.Offset, uint64(8))
	})

	t.Run("test rel jump with immediate", func(t *testing.T) {
		vm := defaultVirtualMachineWithCode("jmp rel 123;")
		setInitialReg(vm, 1, 1, 0)

		err := vm.RunStep(&hintrunner)
		require.NoError(t, err)

		assert.Equal(t, vm.Context.Pc.Offset, uint64(123))
	})

	t.Run("test conditional jump, if <op> != 0", func(t *testing.T) {
		vm := defaultVirtualMachineWithCode("jmp rel [ap - 1] if [fp + 2] != 0;")
		setInitialReg(vm, 1, 1, 0)

		writeToDataSegment(vm, vm.Context.Ap-1, 4)
		writeToDataSegment(vm, vm.Context.Fp+2, 5)

		err := vm.RunStep(&hintrunner)
		require.NoError(t, err)

		assert.Equal(t, vm.Context.Pc.Offset, uint64(4))
	})

	t.Run("test conditional jump, if <op> == 0", func(t *testing.T) {
		vm := defaultVirtualMachineWithCode("jmp rel [ap - 1] if [fp + 2] != 0;")
		setInitialReg(vm, 1, 1, 0)

		writeToDataSegment(vm, vm.Context.Ap-1, 3)
		writeToDataSegment(vm, vm.Context.Fp+2, 0)

		err := vm.RunStep(&hintrunner)
		require.NoError(t, err)

		assert.Equal(t, vm.Context.Pc.Offset, uint64(1))
	})

	t.Run("test 'call abs <address>'", func(t *testing.T) {
		vm := defaultVirtualMachineWithCode("call abs [fp + 4];")
		setInitialReg(vm, 1, 4, 0)

		writeToDataSegment(vm, vm.Context.Fp+4, &mem.MemoryAddress{SegmentIndex: 5, Offset: 8})

		err := vm.RunStep(&hintrunner)
		require.NoError(t, err)

		assert.Equal(t, vm.Context.Pc, mem.MemoryAddress{SegmentIndex: 5, Offset: 8})

		// ap is advanced by 2 and fp is set to the new ap
		assert.Equal(t, vm.Context.Ap, vm.Context.Fp, uint64(3))
	})

	t.Run("test call rel with immediate", func(t *testing.T) {
		vm := defaultVirtualMachineWithCode("call rel 123;")
		setInitialReg(vm, 2, 1, 0)

		err := vm.RunStep(&hintrunner)
		require.NoError(t, err)

		assert.Equal(t, vm.Context.Pc, mem.MemoryAddress{SegmentIndex: 0, Offset: 123})

		// ap is advanced by 2 and fp is set to the new ap
		assert.Equal(t, vm.Context.Fp, vm.Context.Ap, uint64(4))
	})

	t.Run("test call rel with address", func(t *testing.T) {
		vm := defaultVirtualMachineWithCode("call rel [ap + 4];")
		setInitialReg(vm, 1, 4, 0)

		writeToDataSegment(vm, vm.Context.Ap+4, 5)

		err := vm.RunStep(&hintrunner)
		require.NoError(t, err)

		assert.Equal(t, vm.Context.Pc, mem.MemoryAddress{SegmentIndex: 0, Offset: 5})

		// ap is advanced by 2 and fp is set to the new ap
		assert.Equal(t, vm.Context.Ap, vm.Context.Fp, uint64(3))
	})

	t.Run("test ret", func(t *testing.T) {
		vm := defaultVirtualMachineWithCode("ret;")
		setInitialReg(vm, 1, 3, 0)

		// The value from [fp - 1] is read. This is the return PC.
		writeToDataSegment(vm, vm.Context.Fp-1, &mem.MemoryAddress{SegmentIndex: 5, Offset: 6})
		writeToDataSegment(vm, vm.Context.Ap, &mem.MemoryAddress{SegmentIndex: 3, Offset: 46})

		err := vm.RunStep(&hintrunner)
		require.NoError(t, err)

		// the ret instruction reads the value from [fp - 1] to determine the next program counter (PC)
		assert.Equal(t, vm.Context.Pc, mem.MemoryAddress{SegmentIndex: 5, Offset: 6})

		// the value of ap is written to fp
		assert.Equal(t, vm.Context.Fp, uint64(46))
	})
}

// ======================
// Test Memory Relocation
// ======================

func TestMemoryRelocationWithFelt(t *testing.T) {
	// segment 0: [2, -, -, 3]
	// segment 3: [5, -, 7, -, 11, 13]
	// relocated: [-, 2, -, -, 3, 5, -, 7, -, 11, 13]
	vm := defaultVirtualMachine()

	updateMemoryWithValues(
		vm.Memory,
		[]memoryWrite{
			// segment zero
			{0, 0, uint64(2)},
			{0, 3, uint64(3)},
			// segment three
			{3, 0, uint64(5)},
			{3, 2, uint64(7)},
			{3, 4, uint64(11)},
			{3, 5, uint64(13)},
		},
	)

	res := vm.RelocateMemory()

	expected := []*f.Element{
		nil,
		// segment zero
		new(f.Element).SetUint64(2),
		nil,
		nil,
		new(f.Element).SetUint64(3),
		// segment three
		new(f.Element).SetUint64(5),
		nil,
		new(f.Element).SetUint64(7),
		nil,
		new(f.Element).SetUint64(11),
		new(f.Element).SetUint64(13),
	}

	require.Equal(t, len(expected), len(res))
	require.Equal(t, expected, res)
}

func TestMemoryRelocationWithAddress(t *testing.T) {
	// segment 0: [-, 1, -, 1:5] (4)
	// segment 1: [1, 4:3, 7, -, -, 13] (10)
	// segment 2: [0:1] (11)
	// segment 3: [2:0] (12)
	// segment 4: [0:0, 1:1, 1:5, 15] (16)
	// relocated: [
	//      dummy:  -,
	//      zero:   -,  1, -, 10,
	//      one:    1, 16, 7,  -, -, 13,
	//      two:    2,
	//      three: 11,
	//      four:   1,  6, 10, 15,
	// ]

	vm := defaultVirtualMachine()
	updateMemoryWithValues(
		vm.Memory,
		[]memoryWrite{
			// segment zero
			{0, 1, uint64(1)},
			{0, 3, &mem.MemoryAddress{SegmentIndex: 1, Offset: 5}},
			// segment one
			{1, 0, uint64(1)},
			{1, 1, &mem.MemoryAddress{SegmentIndex: 4, Offset: 3}},
			{1, 2, uint64(7)},
			{1, 5, uint64(13)},
			// segment two
			{2, 0, &mem.MemoryAddress{SegmentIndex: 0, Offset: 1}},
			// segment three
			{3, 0, &mem.MemoryAddress{SegmentIndex: 2, Offset: 0}},
			// segment four
			{4, 0, &mem.MemoryAddress{SegmentIndex: 0, Offset: 0}},
			{4, 1, &mem.MemoryAddress{SegmentIndex: 1, Offset: 1}},
			{4, 2, &mem.MemoryAddress{SegmentIndex: 1, Offset: 5}},
			{4, 3, uint64(15)},
		},
	)

	res := vm.RelocateMemory()

	expected := []*f.Element{
		nil,
		// segment zero
		nil,
		new(f.Element).SetUint64(1),
		nil,
		new(f.Element).SetUint64(10),
		// segment one
		new(f.Element).SetUint64(1),
		new(f.Element).SetUint64(16),
		new(f.Element).SetUint64(7),
		nil,
		nil,
		new(f.Element).SetUint64(13),
		// segment two
		new(f.Element).SetUint64(2),
		// segment three
		new(f.Element).SetUint64(11),
		// segment 4
		new(f.Element).SetUint64(1),
		new(f.Element).SetUint64(6),
		new(f.Element).SetUint64(10),
		new(f.Element).SetUint64(15),
	}

	require.Equal(t, len(expected), len(res))
	require.Equal(t, expected, res)
}

// ==============================
// Test Trace and Memory Encoding
// ==============================

func TestTraceEncodingDecoding(t *testing.T) {
	trace := []Trace{
		{Ap: 1, Fp: 2, Pc: 3},
		{Ap: 4, Fp: 5, Pc: 6},
		{Ap: 9, Fp: 8, Pc: 7},
	}

	encodedTrace := EncodeTrace(trace)

	expected := make([]byte, len(trace)*3*8)
	// first context
	binary.LittleEndian.PutUint64(expected[0:8], 1)
	binary.LittleEndian.PutUint64(expected[8:16], 2)
	binary.LittleEndian.PutUint64(expected[16:24], 3)
	// second context
	binary.LittleEndian.PutUint64(expected[24:32], 4)
	binary.LittleEndian.PutUint64(expected[32:40], 5)
	binary.LittleEndian.PutUint64(expected[40:48], 6)
	// third context
	binary.LittleEndian.PutUint64(expected[48:56], 9)
	binary.LittleEndian.PutUint64(expected[56:64], 8)
	binary.LittleEndian.PutUint64(expected[64:72], 7)

	// test encoding
	require.Equal(
		t,
		expected,
		encodedTrace,
	)

	// test decoding
	decodedTrace := DecodeTrace(encodedTrace)
	require.Equal(
		t,
		trace,
		decodedTrace,
	)

}

func TestMemoryEncodingDecoding(t *testing.T) {
	memory := []*f.Element{
		new(f.Element).SetUint64(4),
		new(f.Element).SetUint64(15),
		nil,
		nil,
		new(f.Element).SetUint64(8),
		nil,
		new(f.Element).SetUint64(2),
	}

	encodedMemory := EncodeMemory(memory)

	// the array size depends on the ammount of non nil elements
	// it stores (addres, felt) encoded in little endian in a consecutive way
	expected := make([]byte, 4*(8+32))

	//first element
	binary.LittleEndian.PutUint64(expected[0:8], 0)
	f.LittleEndian.PutElement((*[32]byte)(expected[8:40]), *new(f.Element).SetUint64(4))
	//second element
	binary.LittleEndian.PutUint64(expected[40:48], 1)
	f.LittleEndian.PutElement((*[32]byte)(expected[48:80]), *new(f.Element).SetUint64(15))
	//third element
	binary.LittleEndian.PutUint64(expected[80:88], 4)
	f.LittleEndian.PutElement((*[32]byte)(expected[88:120]), *new(f.Element).SetUint64(8))
	//fourth element
	binary.LittleEndian.PutUint64(expected[120:128], 6)
	f.LittleEndian.PutElement((*[32]byte)(expected[128:160]), *new(f.Element).SetUint64(2))

	require.Equal(
		t,
		len(expected),
		len(encodedMemory),
	)
	require.Equal(
		t,
		expected,
		encodedMemory,
	)

	// testing decoding
	decodedMemory := DecodeMemory(encodedMemory)
	require.Equal(
		t,
		memory,
		decodedMemory,
	)
}

// ==============
// Util Functions
// ==============

func writeToDataSegment(vm *VirtualMachine, index uint64, value any) mem.MemoryAddress {
	mv, err := mem.MemoryValueFromAny(value)
	if err != nil {
		panic(err)
	}

	err = vm.Memory.Write(ExecutionSegment, index, &mv)
	if err != nil {
		panic(err)
	}
	return mem.MemoryAddress{
		SegmentIndex: ExecutionSegment,
		Offset:       index,
	}
}

func defaultVirtualMachine() *VirtualMachine {
	return defaultVirtualMachineWithBytecode(nil)
}

func defaultVirtualMachineWithCode(code string) *VirtualMachine {
	bytecode, err := a.CasmToBytecode(code)
	if err != nil {
		panic(err)
	}

	return defaultVirtualMachineWithBytecode(bytecode)
}

func defaultVirtualMachineWithBytecode(bytecode []*f.Element) *VirtualMachine {
	memory := mem.InitializeEmptyMemory()
	_, err := memory.AllocateSegment(bytecode)
	if err != nil {
		panic(err)
	}

	memory.AllocateEmptySegment()

	vm, err := NewVirtualMachine(Context{}, memory, VirtualMachineConfig{})
	if err != nil {
		panic(err)
	}
	return vm
}

// create a pointer to an Element
func newElementPtr(val uint64) *f.Element {
	element := f.NewElement(val)
	return &element
}

type memoryWrite struct {
	SegmentIndex uint64
	Offset       uint64
	Value        any
}

func updateMemoryWithValues(memory *mem.Memory, valuesToWrite []memoryWrite) {
	var max_segment uint64 = 0
	for _, toWrite := range valuesToWrite {
		// wrap any inside a memory value
		val, err := mem.MemoryValueFromAny(toWrite.Value)
		if err != nil {
			panic(err)
		}

		// if the destination segment does not exist, create it
		for toWrite.SegmentIndex >= max_segment {
			max_segment += 1
			memory.AllocateEmptySegment()
		}

		// write the memory val
		err = memory.Write(toWrite.SegmentIndex, toWrite.Offset, &val)
		if err != nil {
			panic(err)
		}

	}
}
