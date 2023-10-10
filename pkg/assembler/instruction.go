package assembler

import (
	"fmt"

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
	case Op0:
		return "Op0"
	case Imm:
		return "Imm"
	case FpPlusOffOp1:
		return "Fp"
	case ApPlusOffOp1:
		return "Ap"
	default:
		return string(opSrc)
	}
}

const (
	Op0 Op1Src = iota
	Imm
	FpPlusOffOp1
	_
	ApPlusOffOp1
)

type ResLogic uint8

func (res ResLogic) String() string {
	switch res {
	case Op1:
		return "Op1"
	case AddOperands:
		return "Add"
	case MulOperands:
		return "Mul"
	case Unconstrained:
		return "Unconstrained"
	default:
		return string(res)
	}
}

const (
	Op1 ResLogic = iota
	AddOperands
	MulOperands
	Unconstrained
)

type PcUpdate uint8

func (res PcUpdate) String() string {
	switch res {
	case PcUpdateNextInstr:
		return "Next instr"
	case PcUpdateJump:
		return "Jump Abs"
	case PcUpdateJumpRel:
		return "Jump Rel"
	case PcUpdateJnz:
		return "Jnz"
	default:
		return string(res)
	}
}

const (
	PcUpdateNextInstr PcUpdate = iota
	PcUpdateJump
	PcUpdateJumpRel
	_
	PcUpdateJnz
)

type ApUpdate uint8

func (ap ApUpdate) String() string {
	switch ap {
	case SameAp:
		return "Same Ap"
	case AddRes:
		return "Add Res"
	case Add1:
		return "Add 1"
	case Add2:
		return "Add 2"
	default:
		return string(ap)
	}
}

const (
	SameAp ApUpdate = iota
	AddRes
	Add1
	_
	Add2
)

type Opcode uint8

func (op Opcode) String() string {
	switch op {
	case OpCodeNop:
		return "Nop"
	case OpCodeCall:
		return "Call"
	case OpCodeRet:
		return "Ret"
	case OpCodeAssertEq:
		return "Assert"
	default:
		return string(op)
	}
}

const (
	OpCodeNop Opcode = iota
	OpCodeCall
	OpCodeRet
	_
	OpCodeAssertEq
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

	// Immediate value for 2 word instructions
	Imm string
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
	// Offsets
	op0Offset   = 16
	op1Offset   = 32
	flagsOffset = 48

	// Relative to flagsOffset
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

	// Default values
	// biasedZero     uint16 = 0x8000
	// biasedPlusOne  uint16 = 0x8001
	// biasedMinusOne uint16 = 0x7FFF
	// biasedMinusTwo uint16 = 0x7FFE
)

/*
*    Decode the bytecode into an instruction
 */
func DecodeInstruction(rawInstruction *f.Element) (*Instruction, error) {
	if !rawInstruction.IsUint64() {
		return nil, fmt.Errorf("%s is bigger than 64 bits", rawInstruction.Text(10))
	}
	offDstEnc, offOp0Enc, offOp1Enc, flags := decodeInstructionValues(rawInstruction.Uint64())

	// Create empty instruction
	instruction := new(Instruction)

	// Add unsigned offsets as signed ones
	instruction.OffDest = offDstEnc
	instruction.OffOp0 = offOp0Enc
	instruction.OffOp1 = offOp1Enc

	err := decodeInstructionFlags(instruction, flags)
	if err != nil {
		return nil, fmt.Errorf("flags: %w", err)
	}

	return instruction, nil
}

// break the instruction into 4 segments of 16 bits
// |         off0            |
// |         off1            |
// |         off2            |
// |         flags           |
func decodeInstructionValues(encoding uint64) (
	offDstEnc int16, offOp0Enc int16, offOp1Enc int16, flags uint16,
) {
	encodingWith2sComplement := encoding ^ 0x0000800080008000
	// first, second and third 16 bits of the instruction encoding respectively
	offDstEnc = int16(encodingWith2sComplement)
	offOp0Enc = int16(encodingWith2sComplement >> op0Offset)
	offOp1Enc = int16(encodingWith2sComplement >> op1Offset)
	// bits 48..63
	flags = uint16(encodingWith2sComplement >> flagsOffset)
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

	// op1Addr := flags & (1<<op1ImmBit | 1<<op1FpBit | 1<<op1ApBit)
	op1Addr := flags & (1<<op1ImmBit | 1<<op1FpBit | 1<<op1ApBit) >> op1ImmBit
	if op1Addr == 3 {
		return fmt.Errorf("op1 source: wrong sequence of bits")
	}
	instruction.Op1Source = Op1Src(op1Addr)

	pcUpdate := flags & (1<<pcJumpAbsBit | 1<<pcJumpRelBit | 1<<pcJnzBit) >> pcJumpAbsBit
	if pcUpdate == 3 {
		return fmt.Errorf("pc update: wrong sequence of bits")
	}
	instruction.PcUpdate = PcUpdate(pcUpdate)

	var defaultResLogic ResLogic
	// (0, 0) bits at pc_update corespond to different
	// scenarios depending on the instruction.
	// For JNZ the result is not constrained
	if instruction.PcUpdate == PcUpdateJnz {
		defaultResLogic = Unconstrained
	} else {
		defaultResLogic = Op1
	}

	res := flags & (1<<resAddBit | 1<<resMulBit) >> resAddBit
	if res == 3 {
		return fmt.Errorf("res logic: wrong sequence of bits")
	}

	if res == 0 {
		instruction.Res = defaultResLogic
	} else {
		instruction.Res = ResLogic(res)
	}

	apUpdate := flags & (1<<apAddBit | 1<<apAdd1Bit) >> apAddBit
	if apUpdate == 3 {
		return fmt.Errorf("ap update: wrong sequence of bits")
	}
	instruction.ApUpdate = ApUpdate(apUpdate)

	opcode := flags & (1<<opcodeCallBit | 1<<opcodeRetBit | 1<<opcodeAssertEqBit) >> opcodeCallBit
	if opcode == 3 {
		return fmt.Errorf("opcode: wrong sequence of bits")
	}
	instruction.Opcode = Opcode(opcode)

	// for pc udpate Jnz, res should be unconstrainded, no opcode, and ap should update with Imm
	if instruction.PcUpdate == PcUpdateJnz &&
		(instruction.Res != Unconstrained ||
			instruction.Opcode != OpCodeNop ||
			instruction.ApUpdate != SameAp) {
		return fmt.Errorf(
			"jnz opcode must have unconstrained res logic, no opcode, and no ap change",
		)
	}

	if instruction.Opcode == OpCodeCall {
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

/*
*    Instruction list into bytecode functions
 */
func encodeInstructionListToBytecode(instruction []Instruction) ([]*f.Element, error) {
	n := len(instruction)
	bytecodes := make([]*f.Element, 0, n+(n/2)+1)

	for i := range instruction {
		bytecode, err := encodeOneInstruction(&instruction[i])
		if err != nil {
			return nil, err
		}
		bytecodes = append(bytecodes, bytecode)
		if instruction[i].Imm != "" {
			imm, err := new(f.Element).SetString(instruction[i].Imm)
			if err != nil {
				return nil, err
			}
			bytecodes = append(bytecodes, imm)
		}
	}
	return bytecodes, nil
}

// break the instruction into 4 segments of 16 bits
// | 	   flags       | 	   offOp1      | 	   offOp0      | 	   offDst     |
func encodeOneInstruction(instruction *Instruction) (*f.Element, error) {
	// Get the offsets
	// Combine the offsets and flags into a single uint64
	rawInstruction := encodeOffsets(instruction)

	// Encode the flags
	rawInstruction, err := encodeInstructionFlags(instruction, rawInstruction)
	if err != nil {
		return nil, err
	}

	// Create a new f.Element from the raw instruction
	element := new(f.Element).SetUint64(rawInstruction)

	return element, nil
}

func encodeOffsets(instr *Instruction) uint64 {
	// Find biased version of the offsets
	// then encode them as bytecode
	biasedOffset := findBiasedOffset(instr.OffDest)
	encoding := uint64(biasedOffset)
	biasedOffset = findBiasedOffset(instr.OffOp0)
	encoding |= uint64(biasedOffset) << op0Offset
	biasedOffset = findBiasedOffset(instr.OffOp1)
	encoding |= uint64(biasedOffset) << op1Offset
	return encoding
}

func encodeInstructionFlags(instr *Instruction, encoding uint64) (uint64, error) {
	// Use a seperate flag register to encode the flags
	// To help with relative bit offsets
	flagsReg := uint16(instr.DstRegister) << dstRegBit
	flagsReg |= uint16(instr.Op0Register) << op0RegBit
	flagsReg |= uint16(instr.Op1Source) << op1ImmBit
	flagsReg |= uint16(instr.Res) << resAddBit
	flagsReg |= uint16(instr.PcUpdate) << pcJumpAbsBit
	flagsReg |= uint16(instr.ApUpdate) << apAddBit
	flagsReg |= uint16(instr.Opcode) << opcodeCallBit
	// Finally OR them with the 64 bit encoding with the flagsOffset
	encoding |= uint64(flagsReg) << flagsOffset
	return encoding, nil
}

func findBiasedOffset(value int16) uint16 {
	biasedOffset := uint16(value) ^ 0x8000
	return biasedOffset
}
