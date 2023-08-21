package vm

import (
	"github.com/stretchr/testify/require"
	"testing"

	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/assert"
)

func TestDecodeInstructionValues(t *testing.T) {
	offDest, offOp0, offOp1, flags := decodeInstructionValues(
		new(f.Element).SetBytes([]byte{0x48, 0x06, 0x80, 0x01, 0x7f, 0xff, 0x80, 0x10}).Uint64(),
	)
	assert.Equal(t, uint16(0x8010), offDest)
	assert.Equal(t, uint16(0x7fff), offOp0)
	assert.Equal(t, uint16(0x8001), offOp1)
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
