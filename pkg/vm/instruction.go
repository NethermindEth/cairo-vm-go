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

type Op1Addr uint8

const (
	Imm Op1Addr = iota
	ApPlusOff2
	FpPlustOff2
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

type FpUpdate uint8

const (
	ApPlus2 FpUpdate = iota
	Dst
	SameFp
)

type Opcode uint8

const (
	Call Opcode = iota
	Ret
	AssertEq
	Nop
)

type Instruction struct {
	Off0 int16
	Off1 int16
	Off2 int16

	// Imm *f.Element

	DstRegister Register
	Op0Register Register

	Op1Addr Op1Addr

	Res      ResLogic
	PcUpdate PcUpdate
	ApUpdate ApUpdate
	FpUpdate FpUpdate
	Opcode   Opcode
}

func (instr Instruction) Size() uint8 {
	if instr.Op1Addr == Imm {
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
	instruction.Off0 = int16(int(off0Enc) - (1 << (offsetBits - 1)))
	instruction.Off1 = int16(int(off1Enc) - (1 << (offsetBits - 1)))
	instruction.Off2 = int16(int(off2Enc) - (1 << (offsetBits - 1)))

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
	off0Enc uint16, off1Enc uint16, off2Enc uint16, flags uint16,
) {
	// first, second and third 16 bits of the instruction encoding respectively
	off0Enc = uint16(encoding & (1<<offsetBits - 1))
	off1Enc = uint16((encoding >> offsetBits) & (1<<offsetBits - 1))
	off2Enc = uint16((encoding >> (2 * offsetBits)) & (1<<offsetBits - 1))
	// bits 48..63
	flags = uint16(encoding >> (3 * offsetBits))
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

	op1Addr, err := oneHot((flags>>op1ImmBit)&1, (flags>>op1ApBit)&1, (flags>>op1FpBit)&1)
	if err != nil {
		return fmt.Errorf("error decoding op1_addr of instruction: %w", err)
	}
	instruction.Op1Addr = Op1Addr(op1Addr)

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

	// The result must be unconstrained in case of JNZ
	if instruction.PcUpdate == Jnz && instruction.Res != Unconstrained {
		return fmt.Errorf("jnz opcode must have Unconstrained res logic")
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

	if instruction.Opcode == Call {
		// (0, 0) bits for ap_update also stand for different
		// behaviour in different opcodes.
		// Call treats (0, 0) as ADD2 logic
		if instruction.ApUpdate != SameAp {
			return fmt.Errorf("CALL must have ap_update = ADD2")
		}
		instruction.ApUpdate = Add2
	}

	switch instruction.Opcode {
	case Call:
		instruction.FpUpdate = ApPlus2
	case Ret:
		instruction.FpUpdate = Dst
	default:
		instruction.FpUpdate = SameFp
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
