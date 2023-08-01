package vm

import (
	"math/big"
	"testing"

	f "github.com/NethermindEth/juno/core/felt"
	"github.com/stretchr/testify/assert"
)

func TestAssertEq(t *testing.T) {
	instruction := Instruction{
		Off0:        (*new(f.Felt).SetUint64(0)),
		Off1:        (*new(f.Felt).SetBigInt(big.NewInt(-1))),
		Off2:        (*new(f.Felt).SetUint64(1)),
		Imm:         (new(f.Felt).SetUint64(1)),
		DstRegister: Ap,
		Op0Register: Fp,
		Op1Addr:     Imm,
		Res:         Op1,
		PcUpdate:    NextInstr,
		ApUpdate:    Add1,
		FpUpdate:    SameFp,
		Opcode:      AssertEq,
	}

	decoded, err := DecodeInstruction(
		(new(f.Felt).SetBytes([]byte{0x48, 0x06, 0x80, 0x01, 0x7F, 0xFF, 0x80, 00})),
		(new(f.Felt).SetUint64(1)),
	)

	assert.NoError(t, err)
	assert.Equal(t, *decoded, instruction)
}

func TestJmp(t *testing.T) {
	instruction := Instruction{
		Off0:        (*new(f.Felt).SetBigInt(big.NewInt(-1))),
		Off1:        (*new(f.Felt).SetUint64(2)),
		Off2:        (*new(f.Felt).SetUint64(0)),
		Imm:         nil,
		DstRegister: Fp,
		Op0Register: Ap,
		Op1Addr:     FpPlustOff2,
		Res:         AddOperands,
		PcUpdate:    JumpRel,
		ApUpdate:    SameAp,
		FpUpdate:    SameFp,
		Opcode:      Nop,
	}

	decoded, err := DecodeInstruction(
		(new(f.Felt).SetBytes([]byte{0x01, 0x29, 0x80, 0x00, 0x80, 0x02, 0x7F, 0xFF})),
		nil,
	)

	assert.NoError(t, err)
	assert.Equal(t, *decoded, instruction)

}

func TestJnz(t *testing.T) {
	instruction := Instruction{
		Off0:        (*new(f.Felt).SetUint64(3)),
		Off1:        (*new(f.Felt).SetBigInt(big.NewInt(-1))),
		Off2:        (*new(f.Felt).SetBigInt(big.NewInt(-16))),
		Imm:         nil,
		DstRegister: Ap,
		Op0Register: Fp,
		Op1Addr:     FpPlustOff2,
		Res:         Unconstrained,
		PcUpdate:    Jnz,
		ApUpdate:    SameAp,
		FpUpdate:    SameFp,
		Opcode:      Nop,
	}

	decoded, err := DecodeInstruction(
		(new(f.Felt).SetBytes([]byte{0x02, 0x0A, 0x7F, 0xF0, 0x7F, 0xFF, 0x80, 0x03})),
		nil,
	)

	assert.NoError(t, err)
	assert.Equal(t, *decoded, instruction)
}

func TestCall(t *testing.T) {
	instruction := Instruction{
		Off0:        (*new(f.Felt).SetUint64(0)),
		Off1:        (*new(f.Felt).SetUint64(1)),
		Off2:        (*new(f.Felt).SetUint64(1)),
		Imm:         (new(f.Felt).SetUint64(1234)),
		DstRegister: Ap,
		Op0Register: Ap,
		Op1Addr:     Imm,
		Res:         Op1,
		PcUpdate:    JumpRel,
		ApUpdate:    Add2,
		FpUpdate:    ApPlus2,
		Opcode:      Call,
	}

	decoded, err := DecodeInstruction(
		(new(f.Felt).SetBytes([]byte{0x11, 0x04, 0x80, 0x01, 0x80, 0x01, 0x80, 0x00})),
		(new(f.Felt).SetUint64(1234)),
	)

	assert.NoError(t, err)
	assert.Equal(t, *decoded, instruction)
}

func TestRet(t *testing.T) {
	instruction := Instruction{
		Off0:        (*new(f.Felt).SetBigInt(big.NewInt(-2))),
		Off1:        (*new(f.Felt).SetBigInt(big.NewInt(-1))),
		Off2:        (*new(f.Felt).SetBigInt(big.NewInt(-1))),
		Imm:         nil,
		DstRegister: Fp,
		Op0Register: Fp,
		Op1Addr:     FpPlustOff2,
		Res:         Op1,
		PcUpdate:    Jump,
		ApUpdate:    SameAp,
		FpUpdate:    Dst,
		Opcode:      Ret,
	}

	decoded, err := DecodeInstruction(
		(new(f.Felt).SetBytes([]byte{0x20, 0x8B, 0x7F, 0xFF, 0x7F, 0xFF, 0x7F, 0xFE})),
		nil,
	)

	assert.NoError(t, err)
	assert.Equal(t, *decoded, instruction)
}

func TestAddAp(t *testing.T) {
	instruction := Instruction{
		Off0:        (*new(f.Felt).SetBigInt(big.NewInt(-1))),
		Off1:        (*new(f.Felt).SetBigInt(big.NewInt(-1))),
		Off2:        (*new(f.Felt).SetUint64(1)),
		Imm:         (new(f.Felt).SetUint64(123)),
		DstRegister: Fp,
		Op0Register: Fp,
		Op1Addr:     Imm,
		Res:         Op1,
		PcUpdate:    NextInstr,
		ApUpdate:    AddImm,
		FpUpdate:    SameFp,
		Opcode:      Nop,
	}

	decoded, err := DecodeInstruction(
		(new(f.Felt).SetBytes([]byte{0x04, 0x07, 0x80, 0x01, 0x7F, 0xFF, 0x7F, 0xFF})),
		(new(f.Felt).SetUint64(123)),
	)

	assert.NoError(t, err)
	assert.Equal(t, *decoded, instruction)
}
