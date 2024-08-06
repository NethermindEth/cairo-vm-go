package disasm

import (
	"fmt"
	"strings"

	"github.com/NethermindEth/cairo-vm-go/pkg/assembler"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

type disassembler struct {
	bytecode []*fp.Element

	config Config

	instructions []casmInstruction

	labels     map[int64]int
	funcLabels map[int64]int

	formatter *casmFormatter

	prog Program
}

func (d *disassembler) Disassemble() (*Program, error) {
	type step struct {
		name string
		fn   func() error
	}
	steps := []step{
		{"decode instructions", d.decodeInstructionsStep},
		{"collect labels", d.collectLabelsStep},
		{"construct program", d.constructProgramStep},
	}
	for _, s := range steps {
		if err := s.fn(); err != nil {
			return nil, fmt.Errorf("%s: %w", s.name, err)
		}
	}

	return &d.prog, nil
}

func (d *disassembler) decodeInstructionsStep() error {
	offset := int64(0)

	for offset < int64(len(d.bytecode)) {
		b := d.bytecode[offset]
		decoded, err := assembler.DecodeInstruction(b)
		if err != nil {
			return fmt.Errorf("offset %d (%q): %w", offset, b.String(), err)
		}
		inst := casmInstruction{
			Instruction:    decoded,
			bytecodeOffset: offset,
		}
		if inst.Size() == 2 {
			inst.arg = d.bytecode[offset+1]
		}
		d.instructions = append(d.instructions, inst)
		offset += inst.Size()
	}

	return nil
}

func (d *disassembler) collectLabelsStep() error {
	d.labels = make(map[int64]int)
	d.funcLabels = make(map[int64]int)

	for _, inst := range d.instructions {
		jumpTarget, ok := inst.JumpTarget()
		if !ok {
			continue
		}
		labels := d.labels
		if inst.Opcode == assembler.OpCodeCall {
			labels = d.funcLabels
		}
		if _, ok := labels[jumpTarget]; ok {
			continue
		}
		id := len(labels)
		labels[jumpTarget] = id
	}

	return nil
}

func (d *disassembler) constructProgramStep() error {
	d.formatter = &casmFormatter{
		labels:     d.labels,
		funcLabels: d.funcLabels,
	}

	for _, inst := range d.instructions {
		funcLabelID, ok := d.funcLabels[inst.bytecodeOffset]
		if ok {
			d.pushCommentLine("<F%d> (pc=%d)", funcLabelID, inst.bytecodeOffset)
		}
		labelID, ok := d.labels[inst.bytecodeOffset]
		if ok {
			d.pushCommentLine("<L%d> (pc=%d)", labelID, inst.bytecodeOffset)
		}
		if err := d.pushDisasmLine(inst); err != nil {
			return err
		}
	}

	return nil
}

func (d *disassembler) pushDisasmLine(inst casmInstruction) error {
	var buf strings.Builder

	buf.WriteString(strings.Repeat(" ", d.config.Indent))

	if err := d.formatter.printInstruction(&buf, inst); err != nil {
		return err
	}

	comments, err := d.collectInstComments(inst)
	if err != nil {
		return err
	}

	d.prog.Lines = append(d.prog.Lines, Line{
		Text:     buf.String(),
		Comments: comments,
	})

	return nil
}

func (d *disassembler) pushCommentLine(format string, args ...any) {
	s := fmt.Sprintf(format, args...)
	d.prog.Lines = append(d.prog.Lines, Line{
		Comments: []string{s},
	})
}

func (d *disassembler) collectInstComments(inst casmInstruction) ([]string, error) {
	var comments []string

	isAddApPseudo := false

	switch inst.Opcode {
	case assembler.OpCodeCall:
		jumpTarget, _ := inst.JumpTarget()
		if funcLabelID, ok := d.funcLabels[jumpTarget]; ok {
			comments = append(comments, fmt.Sprintf("calls F%d", funcLabelID))
		}

	case assembler.OpCodeAssertEq:
		if inst.Op1Source == assembler.Imm {
			// Try to recognize the division.
			// So, instead of just this:
			// > assert [fp+1] = [fp] * 2894802230932904970957858226476056084498485772265277359978473644908697616385
			// ...the user sees this (note the comment):
			// > assert [fp+1] = [fp] * 2894802230932904970957858226476056084498485772265277359978473644908697616385 // div 5
			imm := inst.arg
			if inst.Res == assembler.MulOperands && !imm.IsUint64() {
				dividend := fp.NewElement(0)
				dividend.Inverse(imm)
				// If divident is a very large number, then we could got it wrong.
				if dividend.IsUint64() {
					comments = append(comments, "div "+dividend.String())
				}
			}
		}

	case assembler.OpCodeNop:
		jumpTarget, ok := inst.JumpTarget()
		if ok {
			if labelID, ok := d.labels[jumpTarget]; ok {
				comments = append(comments, fmt.Sprintf("targets L%d", labelID))
			}
		}
		if inst.ApUpdate == assembler.AddRes && inst.Op1Source == assembler.Imm {
			isAddApPseudo = true
		}

	case assembler.OpCodeRet:
		// Nothing to do.

	default:
		return nil, fmt.Errorf("unexpected opcode: %v", inst.Opcode)
	}

	switch inst.ApUpdate {
	case assembler.Add2:
		comments = append(comments, "ap += 2")
	case assembler.AddRes:
		if !isAddApPseudo {
			if inst.Op1Source == assembler.Imm {
				comments = append(comments, fmt.Sprintf("ap += %d", feltToInt64(inst.arg)))
			} else {
				comments = append(comments, "ap += $result")
			}
		}
	case assembler.SameAp, assembler.Add1:
		// Nothing to do.
	default:
		return nil, fmt.Errorf("unexpected ap update: %v", inst.ApUpdate)
	}

	return comments, nil
}
