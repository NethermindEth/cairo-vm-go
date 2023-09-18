package assembler

import (
	"fmt"
	"testing"

	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/assert"
)

func TestAssertEqRegister(t *testing.T) {
	encode := parseSingleInstruction("[ap + 0] = [fp + 0], ap++;")

	// verify offsets
	dstOffset := uint16(encode)
	assert.Equal(t, biased(0), dstOffset)

	op0Offset := uint16(encode >> 16)
	assert.Equal(t, biased(-1), op0Offset)

	op1Offset := uint16(encode >> 32)
	assert.Equal(t, biased(0), op1Offset)

	// verify flags
	assert.True(t, (encode>>dstRegBit)&1 == 0)
	assert.True(t, (encode>>op0RegBit)&1 == 1)
	assert.True(
		t,
		(encode>>op1ImmBit)&1 == 0 &&
			(encode>>op1FpBit)&1 == 1 &&
			(encode>>op1ApBit)&1 == 0,
	)
	assert.True(
		t, (encode>>resAddBit)&1 == 0 && (encode>>resMulBit)&1 == 0,
	)
	assert.True(
		t,
		(encode>>pcJumpAbsBit)&1 == 0 &&
			(encode>>pcJumpRelBit)&1 == 0 &&
			(encode>>pcJnzBit)&1 == 0,
	)
	assert.True(
		t, (encode>>apAddBit)&1 == 0 && (encode>>apAdd1Bit)&1 == 1,
	)
	assert.True(
		t,
		(encode>>opcodeRetBit)&1 == 0 &&
			(encode>>opcodeCallBit)&1 == 0 &&
			(encode>>opcodeAssertEqBit)&1 == 1,
	)
}

func TestAssertEqImm(t *testing.T) {
	encode, imm := parseImmediateInstruction("[fp + 1] = 5;")

	// verify imm
	assert.Equal(t, uint64(5), imm.Uint64())

	// verify offsets
	dstOffset := uint16(encode)
	assert.Equal(t, biased(1), dstOffset)

	op0Offset := uint16(encode >> 16)
	assert.Equal(t, biased(-1), op0Offset)

	op1Offset := uint16(encode >> 32)
	assert.Equal(t, biased(1), op1Offset)

	// verify flags
	assert.True(t, (encode>>dstRegBit)&1 == 1)
	assert.True(t, (encode>>op0RegBit)&1 == 1)
	assert.True(
		t,
		(encode>>op1ImmBit)&1 == 1 &&
			(encode>>op1FpBit)&1 == 0 &&
			(encode>>op1ApBit)&1 == 0,
	)
	assert.True(
		t, (encode>>resAddBit)&1 == 0 && (encode>>resMulBit)&1 == 0,
	)
	assert.True(
		t,
		(encode>>pcJumpAbsBit)&1 == 0 &&
			(encode>>pcJumpRelBit)&1 == 0 &&
			(encode>>pcJnzBit)&1 == 0,
	)
	assert.True(
		t, (encode>>apAddBit)&1 == 0 && (encode>>apAdd1Bit)&1 == 0,
	)
	assert.True(
		t,
		(encode>>opcodeRetBit)&1 == 0 &&
			(encode>>opcodeCallBit)&1 == 0 &&
			(encode>>opcodeAssertEqBit)&1 == 1,
	)

}

func TestAssertEqDoubleDeref(t *testing.T) {
	encode := parseSingleInstruction("[ap + 1] = [[ap - 2] - 3];")

	// verify offsets
	dstOffset := uint16(encode)
	assert.Equal(t, biased(1), dstOffset)

	op0Offset := uint16(encode >> 16)
	assert.Equal(t, biased(-2), op0Offset)

	op1Offset := uint16(encode >> 32)
	assert.Equal(t, biased(-3), op1Offset)

	// verify flags
	assert.True(t, (encode>>dstRegBit)&1 == 0)
	assert.True(t, (encode>>op0RegBit)&1 == 0)
	assert.True(
		t,
		(encode>>op1ImmBit)&1 == 0 &&
			(encode>>op1FpBit)&1 == 0 &&
			(encode>>op1ApBit)&1 == 0,
	)
	assert.True(
		t, (encode>>resAddBit)&1 == 0 && (encode>>resMulBit)&1 == 0,
	)
	assert.True(
		t,
		(encode>>pcJumpAbsBit)&1 == 0 &&
			(encode>>pcJumpRelBit)&1 == 0 &&
			(encode>>pcJnzBit)&1 == 0,
	)
	assert.True(
		t, (encode>>apAddBit)&1 == 0 && (encode>>apAdd1Bit)&1 == 0,
	)
	assert.True(
		t,
		(encode>>opcodeRetBit)&1 == 0 &&
			(encode>>opcodeCallBit)&1 == 0 &&
			(encode>>opcodeAssertEqBit)&1 == 1,
	)
}

func TestAssertEqMathOperation(t *testing.T) {
	encode := parseSingleInstruction("[fp - 10] = [ap + 2] * [ap - 3];")

	// verify offsets
	dstOffset := uint16(encode)
	assert.Equal(t, biased(-10), dstOffset)

	op0Offset := uint16(encode >> 16)
	assert.Equal(t, biased(2), op0Offset)

	op1Offset := uint16(encode >> 32)
	assert.Equal(t, biased(-3), op1Offset)

	// verify flags
	assert.True(t, (encode>>dstRegBit)&1 == 1)
	assert.True(t, (encode>>op0RegBit)&1 == 0)
	assert.True(
		t,
		(encode>>op1ImmBit)&1 == 0 &&
			(encode>>op1FpBit)&1 == 0 &&
			(encode>>op1ApBit)&1 == 1,
	)
	assert.True(
		t, (encode>>resAddBit)&1 == 0 && (encode>>resMulBit)&1 == 1,
	)
	assert.True(
		t,
		(encode>>pcJumpAbsBit)&1 == 0 &&
			(encode>>pcJumpRelBit)&1 == 0 &&
			(encode>>pcJnzBit)&1 == 0,
	)
	assert.True(
		t, (encode>>apAddBit)&1 == 0 && (encode>>apAdd1Bit)&1 == 0,
	)
	assert.True(
		t,
		(encode>>opcodeRetBit)&1 == 0 &&
			(encode>>opcodeCallBit)&1 == 0 &&
			(encode>>opcodeAssertEqBit)&1 == 1,
	)
}

func TestCallAbs(t *testing.T) {
	encode, imm := parseImmediateInstruction("call abs 123;")

	// verify imm
	assert.Equal(t, uint64(123), imm.Uint64())

	// verify offsets
	dstOffset := uint16(encode)
	assert.Equal(t, biased(0), dstOffset)

	op0Offset := uint16(encode >> 16)
	assert.Equal(t, biased(1), op0Offset)

	op1Offset := uint16(encode >> 32)
	assert.Equal(t, biased(1), op1Offset)

	// verify flags
	assert.True(t, (encode>>dstRegBit)&1 == 0)
	assert.True(t, (encode>>op0RegBit)&1 == 0)
	assert.True(
		t,
		(encode>>op1ImmBit)&1 == 1 &&
			(encode>>op1FpBit)&1 == 0 &&
			(encode>>op1ApBit)&1 == 0,
	)
	assert.True(
		t, (encode>>resAddBit)&1 == 0 && (encode>>resMulBit)&1 == 0,
	)
	assert.True(
		t,
		(encode>>pcJumpAbsBit)&1 == 1 &&
			(encode>>pcJumpRelBit)&1 == 0 &&
			(encode>>pcJnzBit)&1 == 0,
	)
	assert.True(
		t, (encode>>apAddBit)&1 == 0 && (encode>>apAdd1Bit)&1 == 0,
	)
	assert.True(
		t,
		(encode>>opcodeRetBit)&1 == 0 &&
			(encode>>opcodeCallBit)&1 == 1 &&
			(encode>>opcodeAssertEqBit)&1 == 0,
	)
}

func TestCallRel(t *testing.T) {
	encode := parseSingleInstruction("call rel [ap - 3];")

	// verify offsets
	dstOffset := uint16(encode)
	assert.Equal(t, biased(0), dstOffset)

	op0Offset := uint16(encode >> 16)
	assert.Equal(t, biased(1), op0Offset)

	op1Offset := uint16(encode >> 32)
	assert.Equal(t, biased(-3), op1Offset)

	// verify flags
	assert.True(t, (encode>>dstRegBit)&1 == 0)
	assert.True(t, (encode>>op0RegBit)&1 == 0)
	assert.True(
		t,
		(encode>>op1ImmBit)&1 == 0 &&
			(encode>>op1FpBit)&1 == 0 &&
			(encode>>op1ApBit)&1 == 1,
	)
	assert.True(
		t, (encode>>resAddBit)&1 == 0 && (encode>>resMulBit)&1 == 0,
	)
	assert.True(
		t,
		(encode>>pcJumpAbsBit)&1 == 0 &&
			(encode>>pcJumpRelBit)&1 == 1 &&
			(encode>>pcJnzBit)&1 == 0,
	)
	assert.True(
		t, (encode>>apAddBit)&1 == 0 && (encode>>apAdd1Bit)&1 == 0,
	)
	assert.True(
		t,
		(encode>>opcodeRetBit)&1 == 0 &&
			(encode>>opcodeCallBit)&1 == 1 &&
			(encode>>opcodeAssertEqBit)&1 == 0,
	)
}

func TestRet(t *testing.T) {
	encode := parseSingleInstruction("ret;")

	// verify offsets
	dstOffset := uint16(encode)
	assert.Equal(t, biased(-2), dstOffset)

	op0Offset := uint16(encode >> 16)
	assert.Equal(t, biased(-1), op0Offset)

	op1Offset := uint16(encode >> 32)
	assert.Equal(t, biased(-1), op1Offset)

	// verify flags
	assert.True(t, (encode>>dstRegBit)&1 == 1)
	assert.True(t, (encode>>op0RegBit)&1 == 1)
	assert.True(
		t,
		(encode>>op1ImmBit)&1 == 0 &&
			(encode>>op1FpBit)&1 == 1 &&
			(encode>>op1ApBit)&1 == 0,
	)
	assert.True(
		t, (encode>>resAddBit)&1 == 0 && (encode>>resMulBit)&1 == 0,
	)
	assert.True(
		t,
		(encode>>pcJumpAbsBit)&1 == 1 &&
			(encode>>pcJumpRelBit)&1 == 0 &&
			(encode>>pcJnzBit)&1 == 0,
	)
	assert.True(
		t, (encode>>apAddBit)&1 == 0 && (encode>>apAdd1Bit)&1 == 0,
	)
	assert.True(
		t,
		(encode>>opcodeRetBit)&1 == 1 &&
			(encode>>opcodeCallBit)&1 == 0 &&
			(encode>>opcodeAssertEqBit)&1 == 0,
	)
}

func TestJumpAbs(t *testing.T) {
	encode := parseSingleInstruction("jmp abs [fp - 5] + [fp + 3];")

	// verify offsets
	dstOffset := uint16(encode)
	assert.Equal(t, biased(-1), dstOffset)

	op0Offset := uint16(encode >> 16)
	assert.Equal(t, biased(-5), op0Offset)

	op1Offset := uint16(encode >> 32)
	assert.Equal(t, biased(3), op1Offset)

	// verify flags
	assert.True(t, (encode>>dstRegBit)&1 == 1)
	assert.True(t, (encode>>op0RegBit)&1 == 1)
	assert.True(
		t,
		(encode>>op1ImmBit)&1 == 0 &&
			(encode>>op1FpBit)&1 == 1 &&
			(encode>>op1ApBit)&1 == 0,
	)
	assert.True(
		t, (encode>>resAddBit)&1 == 1 && (encode>>resMulBit)&1 == 0,
	)
	assert.True(
		t,
		(encode>>pcJumpAbsBit)&1 == 1 &&
			(encode>>pcJumpRelBit)&1 == 0 &&
			(encode>>pcJnzBit)&1 == 0,
	)
	assert.True(
		t, (encode>>apAddBit)&1 == 0 && (encode>>apAdd1Bit)&1 == 0,
	)
	assert.True(
		t,
		(encode>>opcodeRetBit)&1 == 0 &&
			(encode>>opcodeCallBit)&1 == 0 &&
			(encode>>opcodeAssertEqBit)&1 == 0,
	)
}

func TestJnz(t *testing.T) {
	encode := parseSingleInstruction("jmp rel [ap - 2] if [fp - 7] != 0;")

	// verify offsets
	dstOffset := uint16(encode)
	assert.Equal(t, biased(-7), dstOffset)

	op0Offset := uint16(encode >> 16)
	assert.Equal(t, biased(-1), op0Offset)

	op1Offset := uint16(encode >> 32)
	assert.Equal(t, biased(-2), op1Offset)

	// verify flags
	assert.True(t, (encode>>dstRegBit)&1 == 1)
	assert.True(t, (encode>>op0RegBit)&1 == 1)
	assert.True(
		t,
		(encode>>op1ImmBit)&1 == 0 &&
			(encode>>op1FpBit)&1 == 0 &&
			(encode>>op1ApBit)&1 == 1,
	)
	assert.True(
		t, (encode>>resAddBit)&1 == 0 && (encode>>resMulBit)&1 == 0,
	)
	assert.True(
		t,
		(encode>>pcJumpAbsBit)&1 == 0 &&
			(encode>>pcJumpRelBit)&1 == 0 &&
			(encode>>pcJnzBit)&1 == 1,
	)
	assert.True(
		t, (encode>>apAddBit)&1 == 0 && (encode>>apAdd1Bit)&1 == 0,
	)
	assert.True(
		t,
		(encode>>opcodeRetBit)&1 == 0 &&
			(encode>>opcodeCallBit)&1 == 0 &&
			(encode>>opcodeAssertEqBit)&1 == 0,
	)
}

func TestAddApImm(t *testing.T) {
	encode, imm := parseImmediateInstruction("ap += 150;")

	// verify imm
	assert.Equal(t, uint64(150), imm.Uint64())

	// verify offsets
	dstOffset := uint16(encode)
	assert.Equal(t, biased(-1), dstOffset)

	op0Offset := uint16(encode >> 16)
	assert.Equal(t, biased(-1), op0Offset)

	op1Offset := uint16(encode >> 32)
	assert.Equal(t, biased(1), op1Offset)

	// verify flags
	assert.True(t, (encode>>dstRegBit)&1 == 1)
	assert.True(t, (encode>>op0RegBit)&1 == 1)
	assert.True(
		t,
		(encode>>op1ImmBit)&1 == 1 &&
			(encode>>op1FpBit)&1 == 0 &&
			(encode>>op1ApBit)&1 == 0,
	)
	assert.True(
		t, (encode>>resAddBit)&1 == 0 && (encode>>resMulBit)&1 == 0,
	)
	assert.True(
		t,
		(encode>>pcJumpAbsBit)&1 == 0 &&
			(encode>>pcJumpRelBit)&1 == 0 &&
			(encode>>pcJnzBit)&1 == 0,
	)
	assert.True(
		t, (encode>>apAddBit)&1 == 1 && (encode>>apAdd1Bit)&1 == 0,
	)
	assert.True(
		t,
		(encode>>opcodeRetBit)&1 == 0 &&
			(encode>>opcodeCallBit)&1 == 0 &&
			(encode>>opcodeAssertEqBit)&1 == 0,
	)

}

func parseImmediateInstruction(casmCode string) (uint64, *f.Element) {
	instructions, err := CasmToBytecode(casmCode)
	if err != nil {
		panic(err)
	}

	if len(instructions) != 2 {
		panic(fmt.Errorf("Expected a sized 2 instruction, got %d", len(instructions)))
	}

	return instructions[0].Uint64(), instructions[1]
}

func parseSingleInstruction(casmCode string) uint64 {
	instructions, err := CasmToBytecode(casmCode)
	if err != nil {
		panic(err)
	}

	if len(instructions) != 1 {
		panic(fmt.Errorf("Expected 1 instruction, got %d", len(instructions)))
	}
	return instructions[0].Uint64()
}

func biased(num int16) uint16 {
	return uint16(num) ^ 0x8000
}
