package vm

import (
	"fmt"

	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

type Register uint8

const (
	Ap Register = iota
	Fp
)

type Op1Src uint8

const (
	Imm Op1Src = iota
	FpPlusOffOp1
	ApPlusOffOp1
	Op0
)

type ResLogic uint8

const (
	AddOperands ResLogic = iota
	MulOperands
	Unconstrained
	Op1
)

type PcUpdate uint8

const (
	Jump PcUpdate = iota
	JumpRel
	Jnz
	NextInstr
)

type ApUpdate uint8

const (
	AddImm ApUpdate = iota
	Add1
	SameAp
	Add2
)

type Opcode uint8

const (
	Call Opcode = iota
	Ret
	AssertEq
	Nop
)

type Instruction struct {
	OffDest int16
	OffOp0  int16
	OffOp1  int16

	DstRegister Register
	Op0Register Register

	Op1Source Op1Src

	Res ResLogic

	// How to update registries after instruction execution
	PcUpdate PcUpdate
	ApUpdate ApUpdate

	// Defines which instruction needs to be executed
	Opcode Opcode
}

func (instr Instruction) Size() uint8 {
	if instr.Op1Source == Imm {
		return 2
	}
	return 1
}

const (
	dstRegBit         = 0
	op0RegBit         = 1
	op1ImmBit         = 2
	op1FpBit          = 3
	op1ApBit          = 4
	resAddBit         = 5
	resMulBit         = 6
	pcJumpAbsBit      = 7
	pcJumpRelBit      = 8
	pcJnzBit          = 9
	apAddBit          = 10
	apAdd1Bit         = 11
	opcodeCallBit     = 12
	opcodeRetBit      = 13
	opcodeAssertEqBit = 14
	offsetBits        = 16
)

func DecodeInstruction(rawInstruction *f.Element) (*Instruction, error) {
	if !rawInstruction.IsUint64() {
		return nil, fmt.Errorf("error decoding instruction: %d is bigger than 64 bits", *rawInstruction)
	}
	off0Enc, off1Enc, off2Enc, flags := decodeInstructionValues(rawInstruction.Uint64())

	// Create empty instruction
	instruction := new(Instruction)

	// Add unsigned offsets as signed ones
	instruction.OffDest = off0Enc
	instruction.OffOp0 = off1Enc
	instruction.OffOp1 = off2Enc

	err := decodeInstructionFlags(instruction, flags)
	if err != nil {
		return nil, err
	}

	return instruction, nil
}

// break the instruction into 4 segments of 16 bits
// |         off0            |
// |         off1            |
// |         off2            |
// |         flags           |
func decodeInstructionValues(encoding uint64) (
	off0Enc int16, off1Enc int16, off2Enc int16, flags uint16,
) {
	encodingWith2sComplement := encoding ^ 0x0000800080008000
	// first, second and third 16 bits of the instruction encoding respectively
	off0Enc = int16(encodingWith2sComplement & (1<<offsetBits - 1))
	off1Enc = int16((encodingWith2sComplement >> offsetBits) & (1<<offsetBits - 1))
	off2Enc = int16((encodingWith2sComplement >> (2 * offsetBits)) & (1<<offsetBits - 1))
	// bits 48..63
	flags = uint16(encodingWith2sComplement >> (3 * offsetBits))
	return
}

// Update instruction fields according to flags
// | dst | op0 | op1 src |  res  |   pc   |   ap   |  opcode  |  - |
// | reg | reg |         | logic | update | update |          |  - |
// |-----|-----|---------|-------|--------|--------|----------|----|
// |  0  |  1  | 2  3  4 |  5 6  | 7  8 9 | 10  11 | 12 13 14 | 15 |
func decodeInstructionFlags(instruction *Instruction, flags uint16) error {
	// Extract instruction flags
	instruction.DstRegister = Register((flags >> dstRegBit) & 1)
	instruction.Op0Register = Register((flags >> op0RegBit) & 1)

	op1Addr, err := oneHot((flags>>op1ImmBit)&1, (flags>>op1FpBit)&1, (flags>>op1ApBit)&1)
	if err != nil {
		return fmt.Errorf("error decoding op1_addr of instruction: %w", err)
	}
	instruction.Op1Source = Op1Src(op1Addr)

	pcUpdate, err := oneHot((flags>>pcJumpAbsBit)&1, (flags>>pcJumpRelBit)&1, (flags>>pcJnzBit)&1)
	if err != nil {
		return fmt.Errorf("error decoding pc_update of instruction: %w", err)
	}
	instruction.PcUpdate = PcUpdate(pcUpdate)

	var defaultResLogic ResLogic
	// (0, 0) bits at pc_update corespond to different
	// scenarios depending on the instruction.
	// For JNZ the result is not constrained
	if instruction.PcUpdate == Jnz {
		defaultResLogic = Unconstrained
	} else {
		defaultResLogic = Op1
	}

	res, err := oneHot((flags>>resAddBit)&1, (flags>>resMulBit)&1)
	if err != nil {
		return fmt.Errorf("error decoding res_logic of instruction: %w", err)
	}

	if res == 2 {
		instruction.Res = defaultResLogic
	} else {
		instruction.Res = ResLogic(res)
	}

	apUpdate, err := oneHot((flags>>apAddBit)&1, (flags>>apAdd1Bit)&1)
	if err != nil {
		return fmt.Errorf("error decoding ap_update of instruction: %w", err)
	}
	instruction.ApUpdate = ApUpdate(apUpdate)

	opcode, err := oneHot((flags>>opcodeCallBit)&1, (flags>>opcodeRetBit)&1, (flags>>opcodeAssertEqBit)&1)
	if err != nil {
		return fmt.Errorf("error decoding opcode of instruction: %w", err)
	}
	instruction.Opcode = Opcode(opcode)

	// for pc udpate Jnz, res should be unconstrainded, no opcode, and ap should update with Imm
	if instruction.PcUpdate == Jnz &&
		(instruction.Res != Unconstrained ||
			instruction.Opcode != Nop ||
			instruction.ApUpdate != AddImm) {
		return fmt.Errorf(
			"jnz opcode must have unconstrained res logic, no opcode, and ap should update using an Imm",
		)
	}

	if instruction.Opcode == Call {
		// (0, 0) bits for ap_update also stand for different
		// behaviour in different opcodes.
		// Call treats (0, 0) as ADD2 logic
		if instruction.ApUpdate != SameAp {
			return fmt.Errorf("CALL must have ap_update = ADD2")
		}
		instruction.ApUpdate = Add2
	}

	return nil
}

// Given []uint16 of 0s or 1s returns the set bit if there's only one such
// and return len(bits) in case there's no set bits.
// If there are more than 1 set bits return an error.
func oneHot(bits ...uint16) (uint16, error) {
	var checkSum uint16 = 0
	setBit := len(bits)

	// checking
	for i, bit := range bits {
		checkSum += bit

		if bit == 1 {
			setBit = i
		}
	}

	if checkSum > 1 {
		return 0, fmt.Errorf("decoding wrong sequence of bits: %v", bits)
	}

	return uint16(setBit), nil
}
