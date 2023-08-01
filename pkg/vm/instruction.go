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
	Op1 ResLogic = iota
	AddOperands
	MulOperands
	Unconstrained
)

type PcUpdate uint8

const (
	NextInstr PcUpdate = iota
	Jump
	JumpRel
	Jnz
)

type ApUpdate uint8

const (
	SameAp ApUpdate = iota
	AddImm
	Add1
	Add2
)

type FpUpdate uint8

const (
	SameFp FpUpdate = iota
	ApPlus2
	Dst
)

type Opcode uint8

const (
	Nop Opcode = iota
	AssertEq
	Call
	Ret
)

type Instruction struct {
	Off0 f.Felt
	Off1 f.Felt
	Off2 f.Felt

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

var (
	dstRegBit         uint = 0
	op0RegBit         uint = 1
	op1ImmBit         uint = 2
	op1FpBit          uint = 3
	op1ApBit          uint = 4
	resAddBit         uint = 5
	resMulBit         uint = 6
	pcJumpAbsBit      uint = 7
	pcJumpRelBit      uint = 8
	pcJnzBit          uint = 9
	apAddBit          uint = 10
	apAdd1Bit         uint = 11
	opcodeCallBit     uint = 12
	opcodeRetBit      uint = 13
	opcodeAssertEqBit uint = 14
	//reservedBit       uint = 15
	offsetBits    uint = 16
	numberOfFlags uint = 15
)

func decodeInstructionValues(encoding *big.Int) (flags *big.Int, off0_enc *big.Int, off1_enc *big.Int, off2_enc *big.Int, err error) {
	if encoding.Cmp(new(big.Int).Lsh(big.NewInt(1), 3*offsetBits+numberOfFlags)) >= 0 {
		return nil, nil, nil, nil, fmt.Errorf("unsupported instruction")
	}

	off0_enc = big.NewInt(0)
	off1_enc = big.NewInt(0)
	off2_enc = big.NewInt(0)
	flags = big.NewInt(0)

	off0_enc.And(encoding, big.NewInt(1<<offsetBits-1))
	off1_enc.And(encoding.Rsh(encoding, offsetBits), big.NewInt(1<<offsetBits-1))
	off2_enc.And(encoding.Rsh(encoding, offsetBits), big.NewInt(1<<offsetBits-1))
	flags.Rsh(encoding, offsetBits)
	err = nil

	return
}

func DecodeInstruction(instruction *f.Felt, imm *f.Felt) (*Instruction, error) {
	var instr *Instruction = new(Instruction)

	var encoding big.Int
	instruction.BigInt(&encoding)

	flags, off0_enc, off1_enc, off2_enc, err := decodeInstructionValues(new(big.Int).Set(&encoding))

	if err != nil {
		return nil, fmt.Errorf("error during decoding an instruction: %w", err)
	}

	var flag big.Int
	if flag.And(flag.Rsh(flags, dstRegBit), big.NewInt(1)).Cmp(big.NewInt(1)) == 0 {
		instr.DstRegister = Fp
	} else {
		instr.DstRegister = Ap
	}

	if flag.And(flag.Rsh(flags, op0RegBit), big.NewInt(1)).Cmp(big.NewInt(1)) == 0 {
		instr.Op0Register = Fp
	} else {
		instr.Op0Register = Ap
	}

	instr.Op1Addr = map[[3]uint64]Op1Addr{
		[...]uint64{1, 0, 0}: Imm,
		[...]uint64{0, 1, 0}: ApPlusOff2,
		[...]uint64{0, 0, 1}: FpPlustOff2,
		[...]uint64{0, 0, 0}: Op0,
	}[[...]uint64{
		flag.And(flag.Rsh(flags, op1ImmBit), big.NewInt(1)).Uint64(),
		flag.And(flag.Rsh(flags, op1ApBit), big.NewInt(1)).Uint64(),
		flag.And(flag.Rsh(flags, op1FpBit), big.NewInt(1)).Uint64()}]

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

	instr.PcUpdate = map[[3]uint64]PcUpdate{
		[...]uint64{1, 0, 0}: Jump,
		[...]uint64{0, 1, 0}: JumpRel,
		[...]uint64{0, 0, 1}: Jnz,
		[...]uint64{0, 0, 0}: NextInstr,
	}[[...]uint64{
		flag.And(flag.Rsh(flags, pcJumpAbsBit), big.NewInt(1)).Uint64(),
		flag.And(flag.Rsh(flags, pcJumpRelBit), big.NewInt(1)).Uint64(),
		flag.And(flag.Rsh(flags, pcJnzBit), big.NewInt(1)).Uint64()}]

	var defaultResLogic ResLogic

	if instr.PcUpdate == Jnz {
		defaultResLogic = Unconstrained
	} else {
		defaultResLogic = Op1
	}

	instr.Res = map[[2]uint64]ResLogic{
		[...]uint64{1, 0}: AddOperands,
		[...]uint64{0, 1}: MulOperands,
		[...]uint64{0, 0}: defaultResLogic,
	}[[...]uint64{
		flag.And(flag.Rsh(flags, resAddBit), big.NewInt(1)).Uint64(),
		flag.And(flag.Rsh(flags, resMulBit), big.NewInt(1)).Uint64()}]

	if instr.PcUpdate == Jnz && instr.Res != Unconstrained {
		return nil, fmt.Errorf("jnz opcode must have Unconstrained res logic")
	}

	instr.ApUpdate = map[[2]uint64]ApUpdate{
		[...]uint64{1, 0}: AddImm,
		[...]uint64{0, 1}: Add1,
		[...]uint64{0, 0}: SameAp,
	}[[...]uint64{
		flag.And(flag.Rsh(flags, apAddBit), big.NewInt(1)).Uint64(),
		flag.And(flag.Rsh(flags, apAdd1Bit), big.NewInt(1)).Uint64()}]

	instr.Opcode = map[[3]uint64]Opcode{
		[...]uint64{1, 0, 0}: Call,
		[...]uint64{0, 1, 0}: Ret,
		[...]uint64{0, 0, 1}: AssertEq,
		[...]uint64{0, 0, 0}: Nop,
	}[[...]uint64{
		flag.And(flag.Rsh(flags, opcodeCallBit), big.NewInt(1)).Uint64(),
		flag.And(flag.Rsh(flags, opcodeRetBit), big.NewInt(1)).Uint64(),
		flag.And(flag.Rsh(flags, opcodeAssertEqBit), big.NewInt(1)).Uint64()}]

	if instr.Opcode == Call {
		if instr.ApUpdate != SameAp {
			return nil, fmt.Errorf("CALL must have update_ap is ADD2")
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

	var offset f.Felt
	instr.Off0 = *offset.SetBigInt(off0_enc.Sub(off0_enc, big.NewInt(1<<(offsetBits-1))))
	instr.Off1 = *offset.SetBigInt(off1_enc.Sub(off1_enc, big.NewInt(1<<(offsetBits-1))))
	instr.Off2 = *offset.SetBigInt(off2_enc.Sub(off2_enc, big.NewInt(1<<(offsetBits-1))))

	return instr, nil
}
