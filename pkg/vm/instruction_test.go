package vm

import (
	"testing"

	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/assert"
)

func TestAssertEq(t *testing.T) {
	instruction := Instruction{
		OffDest: 0,
		OffOp0:  -1,
		OffOp1:  1,
		// Imm:         (new(f.Element).SetUint64(1)),
		DstRegister: Ap,
		Op0Register: Fp,
		Op1Source:   Imm,
		Res:         Op1,
		PcUpdate:    NextInstr,
		ApUpdate:    Add1,
		FpUpdate:    SameFp,
		Opcode:      AssertEq,
	}

	decoded, err := DecodeInstruction(
		(new(f.Element).SetBytes([]byte{0x48, 0x06, 0x80, 0x01, 0x7F, 0xFF, 0x80, 00})),
		// (new(f.Element).SetUint64(1)),
	)

	assert.NoError(t, err)
	assert.Equal(t, *decoded, instruction)
}

func TestJmp(t *testing.T) {
	instruction := Instruction{
		OffDest: -1,
		OffOp0:  2,
		OffOp1:  0,
		// Imm:         nil,
		DstRegister: Fp,
		Op0Register: Ap,
		Op1Source:   FpPlusOffOp1,
		Res:         AddOperands,
		PcUpdate:    JumpRel,
		ApUpdate:    SameAp,
		FpUpdate:    SameFp,
		Opcode:      Nop,
	}

	decoded, err := DecodeInstruction(
		(new(f.Element).SetBytes([]byte{0x01, 0x29, 0x80, 0x00, 0x80, 0x02, 0x7F, 0xFF})),
		// nil,
	)

	assert.NoError(t, err)
	assert.Equal(t, *decoded, instruction)

}

func TestJnz(t *testing.T) {
	instruction := Instruction{
		OffDest: 3,
		OffOp0:  -1,
		OffOp1:  -16,
		// Imm:         nil,
		DstRegister: Ap,
		Op0Register: Fp,
		Op1Source:   FpPlusOffOp1,
		Res:         Unconstrained,
		PcUpdate:    Jnz,
		ApUpdate:    SameAp,
		FpUpdate:    SameFp,
		Opcode:      Nop,
	}

	decoded, err := DecodeInstruction(
		(new(f.Element).SetBytes([]byte{0x02, 0x0A, 0x7F, 0xF0, 0x7F, 0xFF, 0x80, 0x03})),
		// nil,
	)

	assert.NoError(t, err)
	assert.Equal(t, *decoded, instruction)
}

func TestCall(t *testing.T) {
	instruction := Instruction{
		OffDest: 0,
		OffOp0:  1,
		OffOp1:  1,
		// Imm:         (new(f.Element).SetUint64(1234)),
		DstRegister: Ap,
		Op0Register: Ap,
		Op1Source:   Imm,
		Res:         Op1,
		PcUpdate:    JumpRel,
		ApUpdate:    Add2,
		FpUpdate:    ApPlus2,
		Opcode:      Call,
	}

	decoded, err := DecodeInstruction(
		(new(f.Element).SetBytes([]byte{0x11, 0x04, 0x80, 0x01, 0x80, 0x01, 0x80, 0x00})),
		// (new(f.Element).SetUint64(1234)),
	)

	assert.NoError(t, err)
	assert.Equal(t, *decoded, instruction)
}

func TestRet(t *testing.T) {
	instruction := Instruction{
		OffDest: -2,
		OffOp0:  -1,
		OffOp1:  -1,
		// Imm:         nil,
		DstRegister: Fp,
		Op0Register: Fp,
		Op1Source:   FpPlusOffOp1,
		Res:         Op1,
		PcUpdate:    Jump,
		ApUpdate:    SameAp,
		FpUpdate:    Dst,
		Opcode:      Ret,
	}

	decoded, err := DecodeInstruction(
		(new(f.Element).SetBytes([]byte{0x20, 0x8B, 0x7F, 0xFF, 0x7F, 0xFF, 0x7F, 0xFE})),
		// nil,
	)

	assert.NoError(t, err)
	assert.Equal(t, *decoded, instruction)
}

func TestAddAp(t *testing.T) {
	instruction := Instruction{
		OffDest: -1,
		OffOp0:  -1,
		OffOp1:  1,
		// Imm:         (new(f.Element).SetUint64(123)),
		DstRegister: Fp,
		Op0Register: Fp,
		Op1Source:   Imm,
		Res:         Op1,
		PcUpdate:    NextInstr,
		ApUpdate:    AddImm,
		FpUpdate:    SameFp,
		Opcode:      Nop,
	}

	decoded, err := DecodeInstruction(
		(new(f.Element).SetBytes([]byte{0x04, 0x07, 0x80, 0x01, 0x7F, 0xFF, 0x7F, 0xFF})),
		// (new(f.Element).SetUint64(123)),
	)

	assert.NoError(t, err)
	assert.Equal(t, *decoded, instruction)
}
