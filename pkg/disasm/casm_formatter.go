package disasm

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/NethermindEth/cairo-vm-go/pkg/assembler"
)

type casmFormatter struct {
	labels     map[int64]int
	funcLabels map[int64]int
}

func (cf *casmFormatter) printInstruction(w io.Writer, inst casmInstruction) error {
	var buf bytes.Buffer

	switch inst.Opcode {
	case assembler.OpCodeRet:
		buf.WriteString("ret")

	case assembler.OpCodeCall:
		var callSuffix string
		switch inst.PcUpdate {
		case assembler.PcUpdateJump:
			callSuffix = "abs"
		case assembler.PcUpdateJumpRel:
			callSuffix = "rel"
		}
		fmt.Fprintf(&buf, "call %s %+d", callSuffix, feltToInt64(inst.arg))

	case assembler.OpCodeAssertEq:
		buf.WriteString(cf.formatMemoryOperand(inst.DstRegister, int(inst.OffDest)))
		buf.WriteString(" = ")

		if inst.Res != assembler.Op1 {
			buf.WriteString(cf.formatMemoryOperand(inst.Op0Register, int(inst.OffOp0)))
		}
		switch inst.Res {
		case assembler.AddOperands:
			buf.WriteString(" + ")
		case assembler.MulOperands:
			buf.WriteString(" * ")
		}

		buf.WriteString(cf.formatOperand1(inst))

	case assembler.OpCodeNop:
		// Jumps use the same opcode=0.
		switch inst.PcUpdate {
		case assembler.PcUpdateJump:
			buf.WriteString("jmp abs " + cf.formatOperand1(inst))
		case assembler.PcUpdateJumpRel:
			buf.WriteString("jmp rel " + cf.formatOperand1(inst))
		case assembler.PcUpdateJnz:
			fmt.Fprintf(&buf, "jmp rel %s if %s != 0",
				cf.formatOperand1(inst),
				cf.formatMemoryOperand(inst.DstRegister, int(inst.OffDest)))
		}
		if inst.ApUpdate == assembler.AddRes && inst.Op1Source == assembler.Imm {
			// This is an "ap_add" pseudo-instruction encoding.
			//
			// See Display impl for AddApInstruction type in instruction.rs;
			// https://github.com/starkware-libs/cairo/blob/797781d8365445ad4a1ae8202881a61b6656bb98/crates/cairo-lang-casm/src/instructions.rs#L198
			fmt.Fprintf(&buf, "ap += %d", feltToInt64(inst.arg))
		}

	default:
		return fmt.Errorf("unexpected opcode: %v", inst.Opcode)
	}

	switch inst.ApUpdate {
	case assembler.Add1:
		buf.WriteString(", ap++")
	}

	_, err := w.Write(buf.Bytes())
	return err
}

func (cf *casmFormatter) formatOperand1(inst casmInstruction) string {
	var buf strings.Builder

	switch inst.Op1Source {
	case assembler.ApPlusOffOp1:
		buf.WriteString(cf.formatMemoryOperand(assembler.Ap, int(inst.OffOp1)))
	case assembler.FpPlusOffOp1:
		buf.WriteString(cf.formatMemoryOperand(assembler.Fp, int(inst.OffOp1)))
	case assembler.Imm:
		buf.WriteString(inst.arg.String())
	case assembler.Op0:
		// Things like [[fp+10]+20].
		buf.WriteString(cf.formatMemoryOperand2(inst.Op0Register, int(inst.OffOp0), int(inst.OffOp1)))
	}

	return buf.String()
}

func (cf *casmFormatter) formatMemoryOperand(reg assembler.Register, offset int) string {
	var buf strings.Builder
	buf.WriteString("[")
	buf.WriteString(strings.ToLower(reg.String()))
	if offset != 0 {
		fmt.Fprintf(&buf, "%+d", offset)
	}
	buf.WriteString("]")
	return buf.String()
}

func (cf *casmFormatter) formatMemoryOperand2(reg assembler.Register, offset, offset2 int) string {
	var buf strings.Builder

	buf.WriteString("[[")
	buf.WriteString(strings.ToLower(reg.String()))
	if offset != 0 {
		fmt.Fprintf(&buf, "%+d", offset)
	}
	buf.WriteString("]")
	if offset2 != 0 {
		fmt.Fprintf(&buf, "%+d", offset2)
	}
	buf.WriteString("]")

	return buf.String()
}
