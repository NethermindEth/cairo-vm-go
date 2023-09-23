package vm

import (
	"fmt"
	"math/bits"

	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

type Register uint8

func (reg Register) String() string {
	if reg == Ap {
		return "Ap"
	} else if reg == Fp {
		return "Fp"
	}
	return string(reg)
}

const (
	Ap Register = iota
	Fp
)

type Op1Src uint8

func (opSrc Op1Src) String() string {
	switch opSrc {
	case Imm:
		return "Imm"
	case FpPlusOffOp1:
		return "Fp"
	case ApPlusOffOp1:
		return "Ap"
	case Op0:
		return "Op0"
	default:
		return string(opSrc)
	}
}

const (
	Imm Op1Src = iota
	FpPlusOffOp1
	ApPlusOffOp1
	Op0
)

type ResLogic uint8

func (res ResLogic) String() string {
	switch res {
	case AddOperands:
		return "Add"
	case MulOperands:
		return "Mul"
	case Unconstrained:
		return "Unconstrained"
	case Op1:
		return "Op1"
	default:
		return string(res)
	}
}

const (
	AddOperands ResLogic = iota
	MulOperands
	Unconstrained
	Op1
)

type PcUpdate uint8

func (res PcUpdate) String() string {
	switch res {
	case Jump:
		return "Jump Abs"
	case JumpRel:
		return "Jump Rel"
	case Jnz:
		return "Jnz"
	case NextInstr:
		return "Next instr"
	default:
		return string(res)
	}
}

const (
	Jump PcUpdate = iota
	JumpRel
	Jnz
	NextInstr
)

type ApUpdate uint8

func (ap ApUpdate) String() string {
	switch ap {
	case AddImm:
		return "Add Imm"
	case Add1:
		return "Add 1"
	case SameAp:
		return "Same Ap"
	case Add2:
		return "Add 2"
	default:
		return string(ap)
	}
}

const (
	AddImm ApUpdate = iota
	Add1
	SameAp
	Add2
)

type Opcode uint8

func (op Opcode) String() string {
	switch op {
	case Call:
		return "Call"
	case Ret:
		return "Ret"
	case AssertEq:
		return "Assert"
	case Nop:
		return "Nop"
	default:
		return string(op)
	}
}

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

func (i Instruction) String() string {
	return fmt.Sprintf(`
        Dst Offset: %d
        Dst Register: %s
        Op0 Offset: %d
        Op0 Register: %s
        Op1 Offset: %d
        Op1 Source: %s
        ResLogic: %s
        Pc Update: %s
        Ap Update: %s
        Opcode: %s
    `,
		i.OffDest,
		i.DstRegister,
		i.OffOp0,
		i.Op0Register,
		i.OffOp1,
		i.Op1Source,
		i.Res,
		i.PcUpdate,
		i.ApUpdate,
		i.Opcode,
	)
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
	off0Enc = int16(encodingWith2sComplement)
	off1Enc = int16(encodingWith2sComplement >> offsetBits)
	off2Enc = int16(encodingWith2sComplement >> (2 * offsetBits))
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

	op1Addr, err := oneHot(flags&(1<<op1ImmBit|1<<op1FpBit|1<<op1ApBit), op1ImmBit, 3)
	if err != nil {
		return fmt.Errorf("error decoding op1_addr of instruction: %w", err)
	}
	instruction.Op1Source = Op1Src(op1Addr)

	pcUpdate, err := oneHot(flags&(1<<pcJumpAbsBit|1<<pcJumpRelBit|1<<pcJnzBit), pcJumpAbsBit, 3)
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

	res, err := oneHot(flags&(1<<resAddBit|1<<resMulBit), resAddBit, 2)
	if err != nil {
		return fmt.Errorf("error decoding res_logic of instruction: %w", err)
	}

	if res == 2 {
		instruction.Res = defaultResLogic
	} else {
		instruction.Res = ResLogic(res)
	}

	apUpdate, err := oneHot(flags&(1<<apAddBit|1<<apAdd1Bit), apAddBit, 2)
	if err != nil {
		return fmt.Errorf("error decoding ap_update of instruction: %w", err)
	}
	instruction.ApUpdate = ApUpdate(apUpdate)

	opcode, err := oneHot(flags&(1<<opcodeCallBit|1<<opcodeRetBit|1<<opcodeAssertEqBit), opcodeCallBit, 3)
	if err != nil {
		return fmt.Errorf("error decoding opcode of instruction: %w", err)
	}
	instruction.Opcode = Opcode(opcode)

	// for pc udpate Jnz, res should be unconstrainded, no opcode, and ap should update with Imm
	if instruction.PcUpdate == Jnz &&
		(instruction.Res != Unconstrained ||
			instruction.Opcode != Nop ||
			instruction.ApUpdate != SameAp) {
		fmt.Println(instruction)
		return fmt.Errorf(
			"jnz opcode must have unconstrained res logic, no opcode, and no ap change",
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

// Given a uint16 returns the set bit offset by offset if there's only one such
// and return maxLen in case there's no set bits.
// If there are more than 1 set bits return an error.
func oneHot(bin, offset, maxLen uint16) (uint16, error) {
	numberOfBits := bits.OnesCount16(bin)

	if numberOfBits > 1 {
		digits := make([]uint8, 0, maxLen)
		bin = bin >> offset

		for i := 0; i < int(maxLen); i++ {
			digits = append(digits, uint8(bin%2))
			bin >>= 1
		}
		return 0, fmt.Errorf("decoding wrong sequence of bits: %b", digits)
	}

	len := bits.Len16(bin)

	if len == 0 {
		return maxLen, nil
	}

	return uint16(len-int(offset)) - 1, nil
}
