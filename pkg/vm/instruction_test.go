package vm

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/safemath"
	"github.com/stretchr/testify/require"

	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/assert"
)

func TestDecodeInstructionValues(t *testing.T) {
	offDest, offOp0, offOp1, flags := decodeInstructionValues(
		new(safemath.LazyFelt).SetUval(0x480680017fff8010).Uint64(),
	)
	assert.Equal(t, int16(16), offDest)
	assert.Equal(t, int16(-1), offOp0)
	assert.Equal(t, int16(1), offOp1)
	assert.Equal(t, uint16(0x4806), flags)
}

func TestAssertEq(t *testing.T) {
	expected := Instruction{
		OffDest:     0,
		OffOp0:      -1,
		OffOp1:      1,
		DstRegister: Ap,
		Op0Register: Fp,
		Op1Source:   Imm,
		Res:         Op1,
		PcUpdate:    NextInstr,
		ApUpdate:    Add1,
		Opcode:      AssertEq,
	}

	decoded, err := DecodeInstruction(
		new(safemath.LazyFelt).SetUval(0x480680017FFF8000),
	)

	require.NoError(t, err)
	assert.Equal(t, expected, *decoded)
}

func TestJmp(t *testing.T) {
	expected := Instruction{
		OffDest:     -1,
		OffOp0:      2,
		OffOp1:      0,
		DstRegister: Fp,
		Op0Register: Ap,
		Op1Source:   FpPlusOffOp1,
		Res:         AddOperands,
		PcUpdate:    JumpRel,
		ApUpdate:    SameAp,
		Opcode:      Nop,
	}

	decoded, err := DecodeInstruction(
		new(safemath.LazyFelt).SetUval(0x0129800080027FFF),
	)

	require.NoError(t, err)
	assert.Equal(t, expected, *decoded)
}

func TestJnz(t *testing.T) {
	expected := Instruction{
		OffDest:     3,
		OffOp0:      -1,
		OffOp1:      -16,
		DstRegister: Ap,
		Op0Register: Fp,
		Op1Source:   FpPlusOffOp1,
		Res:         Unconstrained,
		PcUpdate:    Jnz,
		ApUpdate:    SameAp,
		Opcode:      Nop,
	}

	decoded, err := DecodeInstruction(
		new(safemath.LazyFelt).SetUval(0x020A7FF07FFF8003),
	)

	require.NoError(t, err)
	require.Equal(t, expected, *decoded)
}

func TestCall(t *testing.T) {
	expected := Instruction{
		OffDest:     0,
		OffOp0:      1,
		OffOp1:      1,
		DstRegister: Ap,
		Op0Register: Ap,
		Op1Source:   Imm,
		Res:         Op1,
		PcUpdate:    JumpRel,
		ApUpdate:    Add2,
		Opcode:      Call,
	}

	decoded, err := DecodeInstruction(
		new(safemath.LazyFelt).SetUval(0x1104800180018000),
	)

	require.NoError(t, err)
	assert.Equal(t, expected, *decoded)
}

func TestRet(t *testing.T) {
	expected := Instruction{
		OffDest:     -2,
		OffOp0:      -1,
		OffOp1:      -1,
		DstRegister: Fp,
		Op0Register: Fp,
		Op1Source:   FpPlusOffOp1,
		Res:         Op1,
		PcUpdate:    Jump,
		ApUpdate:    SameAp,
		Opcode:      Ret,
	}

	decoded, err := DecodeInstruction(
		new(safemath.LazyFelt).SetUval(0x208B7FFF7FFF7FFE),
	)

	require.NoError(t, err)
	assert.Equal(t, expected, *decoded)
}

func TestAddAp(t *testing.T) {
	expected := Instruction{
		OffDest:     -1,
		OffOp0:      -1,
		OffOp1:      1,
		DstRegister: Fp,
		Op0Register: Fp,
		Op1Source:   Imm,
		Res:         Op1,
		PcUpdate:    NextInstr,
		ApUpdate:    AddImm,
		Opcode:      Nop,
	}

	decoded, err := DecodeInstruction(
		new(safemath.LazyFelt).SetUval(0x040780017FFF7FFF),
	)

	require.NoError(t, err)
	assert.Equal(t, expected, *decoded)
}

func TestBiggerThan64Bits(t *testing.T) {
	instruction := new(safemath.LazyFelt).SetFelt(new(f.Element).SetBigInt(big.NewInt(1).Lsh(big.NewInt(1), 64)))

	_, err := DecodeInstruction(instruction)

	require.Error(t, err)
	expectedError := fmt.Sprintf("error decoding instruction: %s is bigger than 64 bits", instruction.String())
	assert.EqualError(t, err, expectedError)
}

func TestInvalidOpOneAddress(t *testing.T) {
	instruction := new(safemath.LazyFelt).SetUval(0x040f800180018000)

	_, err := DecodeInstruction(instruction)

	require.Error(t, err)
	expectedError := fmt.Sprintf("error decoding op1_addr of instruction: decoding wrong sequence of bits: %v", []uint16{1, 1, 0})
	assert.EqualError(t, err, expectedError)
}

func TestInvalidPcUpdate(t *testing.T) {
	instruction := new(safemath.LazyFelt).SetUval(0x0587800180018000)

	_, err := DecodeInstruction(instruction)

	require.Error(t, err)
	expectedError := fmt.Sprintf("error decoding pc_update of instruction: decoding wrong sequence of bits: %v", []uint16{1, 1, 0})
	assert.EqualError(t, err, expectedError)
}

func TestInvalidResLogic(t *testing.T) {
	instruction := new(safemath.LazyFelt).SetUval(0x0467800180018000)

	_, err := DecodeInstruction(instruction)

	require.Error(t, err)
	expectedError := fmt.Sprintf("error decoding res_logic of instruction: decoding wrong sequence of bits: %v", []uint16{1, 1})
	assert.EqualError(t, err, expectedError)
}

func TestInvalidApUpdate(t *testing.T) {
	instruction := new(safemath.LazyFelt).SetUval(0x0C07800180018000)

	_, err := DecodeInstruction(instruction)

	require.Error(t, err)
	expectedError := fmt.Sprintf("error decoding ap_update of instruction: decoding wrong sequence of bits: %v", []uint16{1, 1})
	assert.EqualError(t, err, expectedError)
}

func TestInvalidOpcode(t *testing.T) {
	instruction := new(safemath.LazyFelt).SetUval(0x3407800180018000)

	_, err := DecodeInstruction(instruction)

	require.Error(t, err)
	expectedError := fmt.Sprintf("error decoding opcode of instruction: decoding wrong sequence of bits: %v", []uint16{1, 1, 0})
	assert.EqualError(t, err, expectedError)
}

func TestPcUpdateJnzInvalid(t *testing.T) {
	instructionInvalidRes := new(safemath.LazyFelt).SetUval(0x0627800180018000)
	instructionInvalidOpcode := new(safemath.LazyFelt).SetUval(0x1607800180018000)
	instructionInvalidApUpdate := new(safemath.LazyFelt).SetUval(0x0607800180018000)
	expectedError := "jnz opcode must have unconstrained res logic, no opcode, and no ap change"

	_, err := DecodeInstruction(instructionInvalidRes)

	require.Error(t, err)
	assert.EqualError(t, err, expectedError)

	_, err = DecodeInstruction(instructionInvalidOpcode)

	require.Error(t, err)
	assert.EqualError(t, err, expectedError)

	_, err = DecodeInstruction(instructionInvalidApUpdate)

	require.Error(t, err)
	assert.EqualError(t, err, expectedError)
}

func TestCallInvalidApUpdate(t *testing.T) {
	instruction := new(safemath.LazyFelt).SetUval(0x1507800180018001)
	expectedError := "CALL must have ap_update = ADD2"

	_, err := DecodeInstruction(instruction)

	require.Error(t, err)
	assert.EqualError(t, err, expectedError)
}
