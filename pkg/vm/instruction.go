package vm

import (
	"fmt"
	"math/big"

	f "github.com/NethermindEth/juno/core/felt"
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

	Imm *f.Felt

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
	if instr.Imm != nil {
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
	//reservedBit        = 15
	offsetBits    = 16
	numberOfFlags = 15
)

func decodeInstructionValues(encoding *big.Int) (flags uint16, off0Enc uint16, off1Enc uint16, off2Enc uint16, err error) {
	if encoding.Cmp(new(big.Int).Lsh(big.NewInt(1), uint(3*offsetBits+numberOfFlags))) >= 0 {
		return 0, 0, 0, 0, fmt.Errorf("unsupported instruction")
	}

	// After this we can safely assume encoding < 2^63
	var uintEncoding = encoding.Uint64()

	// first, second and third 16 bits of the instruction encoding respectively
	off0Enc = uint16(uintEncoding & (1<<offsetBits - 1))
	off1Enc = uint16((uintEncoding >> offsetBits) & (1<<offsetBits - 1))
	off2Enc = uint16((uintEncoding >> (2 * offsetBits)) & (1<<offsetBits - 1))
	// bits 48..63
	flags = uint16(uintEncoding >> (3 * offsetBits))
	err = nil

	return
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

func DecodeInstruction(instruction *f.Felt, imm *f.Felt) (*Instruction, error) {
	var instr *Instruction = new(Instruction)

	// break down the instruction into 4 16-bit segments
	flags, off0Enc, off1Enc, off2Enc, err := decodeInstructionValues(instruction.BigInt(big.NewInt(0)))

	if err != nil {
		return nil, fmt.Errorf("error decoding an instruction: %w", err)
	}

	instr.DstRegister = Register((flags >> dstRegBit) & 1)
	instr.Op0Register = Register((flags >> op0RegBit) & 1)

	op1Addr, err := oneHot((flags>>op1ImmBit)&1, (flags>>op1ApBit)&1, (flags>>op1FpBit)&1)
	if err != nil {
		return nil, fmt.Errorf("error decoding op1_addr of instruction: %w", err)
	}

	instr.Op1Addr = Op1Addr(op1Addr)

	// if the address to draw op1 from is set to be the imm
	// check the imm argument
	if instr.Op1Addr == Imm {
		if imm == nil {
			return nil, fmt.Errorf("op1_addr is Op1Addr.IMM, but no immediate given")
		} else {
			var immFelt f.Felt
			instr.Imm = immFelt.Set(imm)
		}
	} else {
		instr.Imm = nil
	}

	pcUpdate, err := oneHot((flags>>pcJumpAbsBit)&1, (flags>>pcJumpRelBit)&1, (flags>>pcJnzBit)&1)
	if err != nil {
		return nil, fmt.Errorf("error decoding pc_update of instruction: %w", err)
	}

	instr.PcUpdate = PcUpdate(pcUpdate)

	var defaultResLogic ResLogic

	// (0, 0) bits at pc_update corespond to different
	// scenarios depending on the instruction.
	// For JNZ the result is not constrained
	if instr.PcUpdate == Jnz {
		defaultResLogic = Unconstrained
	} else {
		defaultResLogic = Op1
	}

	res, err := oneHot((flags>>resAddBit)&1, (flags>>resMulBit)&1)
	if err != nil {
		return nil, fmt.Errorf("error decoding res_logic of instruction: %w", err)
	}

	if res == 2 {
		instr.Res = defaultResLogic
	} else {
		instr.Res = ResLogic(res)
	}

	// Moreover, the result must be unconstrained in case of JNZ
	if instr.PcUpdate == Jnz && instr.Res != Unconstrained {
		return nil, fmt.Errorf("jnz opcode must have Unconstrained res logic")
	}

	apUpdate, err := oneHot((flags>>apAddBit)&1, (flags>>apAdd1Bit)&1)
	if err != nil {
		return nil, fmt.Errorf("error decoding ap_update of instruction: %w", err)
	}

	instr.ApUpdate = ApUpdate(apUpdate)

	opcode, err := oneHot((flags>>opcodeCallBit)&1, (flags>>opcodeRetBit)&1, (flags>>opcodeAssertEqBit)&1)
	if err != nil {
		return nil, fmt.Errorf("error decoding opcode of instruction: %w", err)
	}

	instr.Opcode = Opcode(opcode)

	if instr.Opcode == Call {
		// (0, 0) bits for ap_update also stand for different
		// behaviour in different opcodes.
		// Call treats (0, 0) as ADD2 logic
		if instr.ApUpdate != SameAp {
			return nil, fmt.Errorf("CALL must have ap_update = ADD2")
		}
		instr.ApUpdate = Add2
	}

	switch instr.Opcode {
	case Call:
		instr.FpUpdate = ApPlus2
	case Ret:
		instr.FpUpdate = Dst
	default:
		instr.FpUpdate = SameFp
	}

	// Turning unsigned offsets into signed ones
	instr.Off0 = int16(int(off0Enc) - (1 << (offsetBits - 1)))
	instr.Off1 = int16(int(off1Enc) - (1 << (offsetBits - 1)))
	instr.Off2 = int16(int(off2Enc) - (1 << (offsetBits - 1)))

	return instr, nil
}
