package assembler

import (
	"testing"

	"github.com/stretchr/testify/require"

	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/assert"
)

// func TestDecodeInstructionValues(t *testing.T) {
// 	offDest, offOp0, offOp1, flags := decodeInstructionValues(
// 		new(f.Element).SetBytes([]byte{0x48, 0x06, 0x80, 0x01, 0x7f, 0xff, 0x80, 0x10}).Uint64(),
// 	)
// 	assert.Equal(t, int16(16), offDest)
// 	assert.Equal(t, int16(-1), offOp0)
// 	assert.Equal(t, int16(1), offOp1)
// 	assert.Equal(t, uint16(0x4806), flags)
// }

func TestAssertEq(t *testing.T) {
	expected := Instruction{
		OffDest: 1,
		OffOp0:  -1,
		OffOp1:  1,
		// Imm:         "5",
		DstRegister: 1,
		Op0Register: 1,
		Op1Source:   1,
		Res:         0,
		PcUpdate:    0,
		ApUpdate:    0,
		Opcode:      4,
	}

	decoded, err := DecodeInstruction(
		// new(f.Element).SetBytes([]byte{0x48, 0x06, 0x80, 0x01, 0x7F, 0xFF, 0x80, 0x01}),
		new(f.Element).SetBytes([]byte{0x40, 0x07, 0x80, 0x01, 0x7F, 0xFF, 0x80, 0x01}),
	)

	require.NoError(t, err)
	assert.Equal(t, expected, *decoded)
}

// func TestJmp(t *testing.T) {
// 	expected := Instruction{
// 		OffDest:     -1,
// 		OffOp0:      2,
// 		OffOp1:      0,
// 		DstRegister: Fp,
// 		Op0Register: Ap,
// 		Op1Source:   FpPlusOffOp1,
// 		Res:         AddOperands,
// 		PcUpdate:    PcUpdateJumpRel,
// 		ApUpdate:    SameAp,
// 		Opcode:      OpCodeNop,
// 	}

// 	decoded, err := DecodeInstruction(
// 		new(f.Element).SetBytes([]byte{0x01, 0x29, 0x80, 0x00, 0x80, 0x02, 0x7F, 0xFF}),
// 	)

// 	require.NoError(t, err)
// 	assert.Equal(t, expected, *decoded)
// }

// func TestJnzInstr(t *testing.T) {
// 	expected := Instruction{
// 		OffDest:     3,
// 		OffOp0:      -1,
// 		OffOp1:      -16,
// 		DstRegister: Ap,
// 		Op0Register: Fp,
// 		Op1Source:   FpPlusOffOp1,
// 		Res:         Unconstrained,
// 		PcUpdate:    PcUpdateJnz,
// 		ApUpdate:    SameAp,
// 		Opcode:      OpCodeNop,
// 	}

// 	decoded, err := DecodeInstruction(
// 		new(f.Element).SetBytes([]byte{0x02, 0x0A, 0x7F, 0xF0, 0x7F, 0xFF, 0x80, 0x03}),
// 	)

// 	require.NoError(t, err)
// 	require.Equal(t, expected, *decoded)
// }

// func TestCall(t *testing.T) {
// 	expected := Instruction{
// 		OffDest:     0,
// 		OffOp0:      1,
// 		OffOp1:      1,
// 		DstRegister: Ap,
// 		Op0Register: Ap,
// 		Op1Source:   Imm,
// 		Res:         Op1,
// 		PcUpdate:    PcUpdateJumpRel,
// 		ApUpdate:    Add2,
// 		Opcode:      OpCodeCall,
// 	}

// 	decoded, err := DecodeInstruction(
// 		new(f.Element).SetBytes([]byte{0x11, 0x04, 0x80, 0x01, 0x80, 0x01, 0x80, 0x00}),
// 	)

// 	require.NoError(t, err)
// 	assert.Equal(t, expected, *decoded)
// }

// func TestRetInstr(t *testing.T) {
// 	expected := Instruction{
// 		OffDest:     -2,
// 		OffOp0:      -1,
// 		OffOp1:      -1,
// 		DstRegister: Fp,
// 		Op0Register: Fp,
// 		Op1Source:   FpPlusOffOp1,
// 		Res:         Op1,
// 		PcUpdate:    PcUpdateJump,
// 		ApUpdate:    SameAp,
// 		Opcode:      OpCodeRet,
// 	}

// 	decoded, err := DecodeInstruction(
// 		new(f.Element).SetBytes([]byte{0x20, 0x8B, 0x7F, 0xFF, 0x7F, 0xFF, 0x7F, 0xFE}),
// 	)

// 	require.NoError(t, err)
// 	assert.Equal(t, expected, *decoded)
// }

// func TestAddAp(t *testing.T) {
// 	expected := Instruction{
// 		OffDest:     -1,
// 		OffOp0:      -1,
// 		OffOp1:      1,
// 		DstRegister: Fp,
// 		Op0Register: Fp,
// 		Op1Source:   Imm,
// 		Res:         Op1,
// 		PcUpdate:    PcUpdateNextInstr,
// 		ApUpdate:    AddRes,
// 		Opcode:      OpCodeNop,
// 	}

// 	decoded, err := DecodeInstruction(
// 		new(f.Element).SetBytes([]byte{0x04, 0x07, 0x80, 0x01, 0x7F, 0xFF, 0x7F, 0xFF}),
// 	)

// 	require.NoError(t, err)
// 	assert.Equal(t, expected, *decoded)
// }

// func TestBiggerThan64Bits(t *testing.T) {
// 	instruction := new(f.Element).SetBigInt(big.NewInt(1).Lsh(big.NewInt(1), 64))

// 	_, err := DecodeInstruction(instruction)

// 	require.Error(t, err)
// 	assert.ErrorContains(t, err, "is bigger than 64 bits")
// }

// func TestInvalidOpOneAddress(t *testing.T) {
// 	instruction := new(f.Element).SetBytes([]byte{0x04, 0x0f, 0x80, 0x01, 0x80, 0x01, 0x80, 0x00})

// 	_, err := DecodeInstruction(instruction)

// 	require.Error(t, err)
// 	assert.ErrorContains(t, err, "op1 source")
// 	assert.ErrorContains(t, err, "wrong sequence of bits")
// }

// func TestInvalidPcUpdate(t *testing.T) {
// 	instruction := new(f.Element).SetBytes([]byte{0x05, 0x87, 0x80, 0x01, 0x80, 0x01, 0x80, 0x00})

// 	_, err := DecodeInstruction(instruction)

// 	require.Error(t, err)
// 	assert.ErrorContains(t, err, "pc update")
// 	assert.ErrorContains(t, err, "wrong sequence of bits")
// }

// func TestInvalidResLogic(t *testing.T) {
// 	instruction := new(f.Element).SetBytes([]byte{0x04, 0x67, 0x80, 0x01, 0x80, 0x01, 0x80, 0x00})

// 	_, err := DecodeInstruction(instruction)

// 	require.Error(t, err)
// 	assert.ErrorContains(t, err, "res logic")
// 	assert.ErrorContains(t, err, "wrong sequence of bits")
// }

// func TestInvalidApUpdate(t *testing.T) {
// 	instruction := new(f.Element).SetBytes([]byte{0x0C, 0x07, 0x80, 0x01, 0x80, 0x01, 0x80, 0x00})

// 	_, err := DecodeInstruction(instruction)

// 	require.Error(t, err)
// 	assert.ErrorContains(t, err, "ap update")
// 	assert.ErrorContains(t, err, "wrong sequence of bits")
// }

// func TestInvalidOpcode(t *testing.T) {
// 	instruction := new(f.Element).SetBytes([]byte{0x34, 0x07, 0x80, 0x01, 0x80, 0x01, 0x80, 0x00})

// 	_, err := DecodeInstruction(instruction)

// 	require.Error(t, err)
// 	assert.ErrorContains(t, err, "opcode")
// 	assert.ErrorContains(t, err, "wrong sequence of bits")
// }

// func TestPcUpdateJnzInvalid(t *testing.T) {
// 	instructionInvalidRes := new(f.Element).SetBytes([]byte{0x06, 0x27, 0x80, 0x01, 0x80, 0x01, 0x80, 0x00})
// 	instructionInvalidOpcode := new(f.Element).SetBytes([]byte{0x16, 0x07, 0x80, 0x01, 0x80, 0x01, 0x80, 0x00})
// 	instructionInvalidApUpdate := new(f.Element).SetBytes([]byte{0x06, 0x07, 0x80, 0x01, 0x80, 0x01, 0x80, 0x00})

// 	expectedError := "jnz opcode must have unconstrained res logic, no opcode, and no ap change"

// 	_, err := DecodeInstruction(instructionInvalidRes)

// 	require.Error(t, err)
// 	assert.ErrorContains(t, err, expectedError)

// 	_, err = DecodeInstruction(instructionInvalidOpcode)

// 	require.Error(t, err)
// 	assert.ErrorContains(t, err, expectedError)

// 	_, err = DecodeInstruction(instructionInvalidApUpdate)

// 	require.Error(t, err)
// 	assert.ErrorContains(t, err, expectedError)
// }

// func TestCallInvalidApUpdate(t *testing.T) {
// 	instruction := new(f.Element).SetBytes([]byte{0x15, 0x07, 0x80, 0x01, 0x80, 0x01, 0x80, 0x01})

// 	_, err := DecodeInstruction(instruction)

// 	require.Error(t, err)
// 	assert.ErrorContains(t, err, "CALL must have ap_update = ADD2")
// }
