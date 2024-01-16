package disasm

import (
	"github.com/NethermindEth/cairo-vm-go/pkg/assembler"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

type casmInstruction struct {
	*assembler.Instruction

	arg            *f.Element
	bytecodeOffset int64
}

func (inst *casmInstruction) JumpTarget() (int64, bool) {
	if inst.Opcode == assembler.OpCodeRet {
		return 0, false
	}
	if inst.PcUpdate == assembler.PcUpdateNextInstr {
		return 0, false
	}

	offset := feltToInt64(inst.arg)
	if inst.PcUpdate == assembler.PcUpdateJump {
		return offset, true
	}
	return inst.bytecodeOffset + offset, true
}

func (inst *casmInstruction) Size() int64 {
	// Note: OpCodeCall also has an immediate (call target).
	if inst.Op1Source == assembler.Imm {
		return 2
	}
	return 1
}
