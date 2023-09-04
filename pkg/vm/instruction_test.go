package vm

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/assert"
)

func TestDecodeInstructionValues(t *testing.T) {
	offDest, offOp0, offOp1, flags := decodeInstructionValues(
		new(f.Element).SetBytes([]byte{0x48, 0x06, 0x80, 0x01, 0x7f, 0xff, 0x80, 0x10}).Uint64(),
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
		new(f.Element).SetBytes([]byte{0x48, 0x06, 0x80, 0x01, 0x7F, 0xFF, 0x80, 0x00}),
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
		new(f.Element).SetBytes([]byte{0x01, 0x29, 0x80, 0x00, 0x80, 0x02, 0x7F, 0xFF}),
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
		ApUpdate:    AddImm,
		Opcode:      Nop,
	}

	decoded, err := DecodeInstruction(
		new(f.Element).SetBytes([]byte{0x06, 0x0A, 0x7F, 0xF0, 0x7F, 0xFF, 0x80, 0x03}),
	)

	require.NoError(t, err)
	assert.Equal(t, expected, *decoded)
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
		new(f.Element).SetBytes([]byte{0x11, 0x04, 0x80, 0x01, 0x80, 0x01, 0x80, 0x00}),
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
		new(f.Element).SetBytes([]byte{0x20, 0x8B, 0x7F, 0xFF, 0x7F, 0xFF, 0x7F, 0xFE}),
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
		new(f.Element).SetBytes([]byte{0x04, 0x07, 0x80, 0x01, 0x7F, 0xFF, 0x7F, 0xFF}),
	)

	require.NoError(t, err)
	assert.Equal(t, expected, *decoded)
}

func TestBiggerThan64Bits(t *testing.T) {
	instruction := new(f.Element).SetBigInt(big.NewInt(1).Lsh(big.NewInt(1), 64))

	_, err := DecodeInstruction(instruction)

	require.Error(t, err)
	expectedError := fmt.Sprintf("error decoding instruction: %d is bigger than 64 bits", *instruction)
	assert.EqualError(t, err, expectedError)
}

func TestInvalidOpOneAddress(t *testing.T) {
	instruction := new(f.Element).SetBytes([]byte{0x04, 0x0f, 0x80, 0x01, 0x80, 0x01, 0x80, 0x00})

	_, err := DecodeInstruction(instruction)

	require.Error(t, err)
	expectedError := fmt.Sprintf("error decoding op1_addr of instruction: decoding wrong sequence of bits: %v", []uint16{1, 1, 0})
	assert.EqualError(t, err, expectedError)
}

func TestInvalidPcUpdate(t *testing.T) {
	instruction := new(f.Element).SetBytes([]byte{0x05, 0x87, 0x80, 0x01, 0x80, 0x01, 0x80, 0x00})

	_, err := DecodeInstruction(instruction)

	require.Error(t, err)
	expectedError := fmt.Sprintf("error decoding pc_update of instruction: decoding wrong sequence of bits: %v", []uint16{1, 1, 0})
	assert.EqualError(t, err, expectedError)
}

func TestInvalidResLogic(t *testing.T) {
	instruction := new(f.Element).SetBytes([]byte{0x04, 0x67, 0x80, 0x01, 0x80, 0x01, 0x80, 0x00})

	_, err := DecodeInstruction(instruction)

	require.Error(t, err)
	expectedError := fmt.Sprintf("error decoding res_logic of instruction: decoding wrong sequence of bits: %v", []uint16{1, 1})
	assert.EqualError(t, err, expectedError)
}

func TestInvalidApUpdate(t *testing.T) {
	instruction := new(f.Element).SetBytes([]byte{0x0C, 0x07, 0x80, 0x01, 0x80, 0x01, 0x80, 0x00})

	_, err := DecodeInstruction(instruction)

	require.Error(t, err)
	expectedError := fmt.Sprintf("error decoding ap_update of instruction: decoding wrong sequence of bits: %v", []uint16{1, 1})
	assert.EqualError(t, err, expectedError)
}

func TestInvalidOpcode(t *testing.T) {
	instruction := new(f.Element).SetBytes([]byte{0x34, 0x07, 0x80, 0x01, 0x80, 0x01, 0x80, 0x00})

	_, err := DecodeInstruction(instruction)

	require.Error(t, err)
	expectedError := fmt.Sprintf("error decoding opcode of instruction: decoding wrong sequence of bits: %v", []uint16{1, 1, 0})
	assert.EqualError(t, err, expectedError)
}

func TestPcUpdateJnzInvalid(t *testing.T) {
	instructionInvalidRes := new(f.Element).SetBytes([]byte{0x06, 0x27, 0x80, 0x01, 0x80, 0x01, 0x80, 0x00})
	instructionInvalidOpcode := new(f.Element).SetBytes([]byte{0x16, 0x07, 0x80, 0x01, 0x80, 0x01, 0x80, 0x00})
	instructionInvalidApUpdate := new(f.Element).SetBytes([]byte{0x02, 0x07, 0x80, 0x01, 0x80, 0x01, 0x80, 0x00})
	expectedError := "jnz opcode must have unconstrained res logic, no opcode, and ap should update using an Imm"

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
	instruction := new(f.Element).SetBytes([]byte{0x15, 0x07, 0x80, 0x01, 0x80, 0x01, 0x80, 0x01})
	expectedError := "CALL must have ap_update = ADD2"

	_, err := DecodeInstruction(instruction)

	require.Error(t, err)
	assert.EqualError(t, err, expectedError)
}
