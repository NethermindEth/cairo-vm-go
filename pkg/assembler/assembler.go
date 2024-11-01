package assembler

import (
	"github.com/alecthomas/participle/v2"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

var parser *participle.Parser[CasmProgram] = participle.MustBuild[CasmProgram](
	// mandatory lookahead to disambiguate between productions:
	// expr -> [reg + n] + [reg + m] and
	// expr -> [reg + n]
	// also required for:
	// instr -> jmp rel <expr> and
	// instr -> jmp rel <expr> if <val> != 0
	//
	// an extra +1 (7->8) step is required for an optionally negative offset (see #186):
	// jmp rel [fp + -111]; without lookahead=8, it won't be parsed sucessfully
	participle.UseLookahead(8),
)

// Given a CASM program it returns its encoded bytecode
func CasmToBytecode(code string) ([]*f.Element, uint8, error) {
	casmAst, err := parser.ParseString("", code)
	if err != nil {
		return nil, 0, err
	}
	// Ast To Instruction List
	wordList, err := astToInstruction(casmAst)
	if err != nil {
		return nil, 0, err
	}
	// Instruction to bytecode
	return encodeInstructionListToBytecode(wordList)
}

// Given a CASM ast it returns a list of instructions
func astToInstruction(ast *CasmProgram) ([]Word, error) {
	n := len(ast.InstructionList)
	wordList := make([]Word, 0, n)
	for i := range ast.InstructionList {
		instruction, imm, err := nodeToInstruction(ast.InstructionList[i])
		if err != nil {
			return nil, err
		}
		wordList = append(wordList, instruction)
		if imm != "" {
			wordList = append(wordList, imm)
		}
	}
	return wordList, nil
}

// Given an Instruction Node return an Instruction and possible Immediate
func nodeToInstruction(node InstructionNode) (Word, Immediate, error) {
	var instr Instruction
	var imm Immediate
	expr := node.Expression()
	err := setInstructionDst(&node, &instr)
	if err != nil {
		return nil, "", err
	}
	err = setInstructionOp0(&node, &instr, expr)
	if err != nil {
		return nil, "", err
	}
	imm, err = setInstructionOp1(&node, &instr, expr)
	if err != nil {
		return nil, "", err
	}
	err = setInstructionFlags(&node, &instr, expr)
	if err != nil {
		return nil, "", err
	}
	return instr, imm, nil
}

func setInstructionDst(node *InstructionNode, instr *Instruction) error {
	if node.ApPlus != nil || node.Jump != nil {
		// dstOffset is not involved so it is set to fp - 1 as default value
		instr.OffDest = -1
		instr.DstRegister = Fp
		return nil
	}
	if node.Call != nil {
		// dstOffset is set to ap + 0
		instr.OffDest = 0
		return nil
	}
	if node.Ret != nil {
		// dstOffset is set as fp - 2
		instr.OffDest = -2
		instr.DstRegister = Fp
		return nil
	}

	var deref *Deref
	if node.AssertEq != nil {
		deref = node.AssertEq.Dst
	} else if node.Jnz != nil {
		deref = node.Jnz.Condition
	}

	offset, err := deref.SignedOffset()
	if err != nil {
		return err
	}
	instr.OffDest = offset
	if deref.IsFp() {
		instr.DstRegister = Fp
	} else {
		instr.DstRegister = Ap
	}
	return nil
}

func setInstructionOp0(node *InstructionNode, instr *Instruction, expr Expressioner) error {
	if node != nil && node.Call != nil {
		// op0 is set as [ap + 1] to store current pc
		instr.OffOp0 = 1
		return nil
	}
	if (node != nil && (node.Jnz != nil || node.Ret != nil)) ||
		(expr.AsDeref() != nil || expr.AsImmediate() != nil) {
		// op0 is not involved, it is set as fp - 1 as default value
		instr.OffOp0 = -1
		instr.Op0Register = Fp
		return nil
	}

	var deref *Deref
	if expr.AsDoubleDeref() != nil {
		deref = expr.AsDoubleDeref().Deref
	} else {
		deref = expr.AsMathOperation().Lhs
	}

	offset, err := deref.SignedOffset()
	if err != nil {
		return err
	}
	instr.OffOp0 = offset
	if deref.IsFp() {
		instr.Op0Register = Fp
	} else {
		instr.Op0Register = Ap
	}
	return nil
}

// Given an instruction node and an instruction set the corresponding offset and
// register of op1. It returns an immediate if it exists and the ocurrence of an
// error.
func setInstructionOp1(node *InstructionNode, instr *Instruction, expr Expressioner) (Immediate, error) {
	if node != nil && node.Ret != nil {
		// op1 is set as [fp - 1], where we read the previous pc
		instr.OffOp1 = -1
		instr.Op1Source = FpPlusOffOp1
		return "", nil
	}

	if expr.AsDeref() != nil {
		offset, err := expr.AsDeref().SignedOffset()
		if err != nil {
			return "", err
		}
		instr.OffOp1 = offset
		if expr.AsDeref().IsFp() {
			instr.Op1Source = FpPlusOffOp1
		} else {
			instr.Op1Source = ApPlusOffOp1
		}
		return "", nil
	} else if expr.AsDoubleDeref() != nil {
		offset, err := expr.AsDoubleDeref().SignedOffset()
		if err != nil {
			return "", err
		}
		instr.OffOp1 = offset
		return "", nil
	} else if expr.AsImmediate() != nil {
		// immediate is converted to Felt during bytecode conversion
		imm := expr.AsImmediate()
		instr.OffOp1 = 1
		instr.Op1Source = Imm
		// instr.Imm = *imm
		return *imm, nil
	}
	//  if it is a math operation, the op1 source is set by the right hand side
	return setInstructionOp1(node, instr, expr.AsMathOperation().Rhs)
}

func setInstructionFlags(node *InstructionNode, instr *Instruction, expression Expressioner) error {
	// Set ResLogic
	if expression != nil && expression.AsMathOperation() != nil {
		if expression.AsMathOperation().Operator == "+" {
			instr.Res = AddOperands
		} else {
			instr.Res = MulOperands
		}
	}

	// Set PcUpdate
	if node.Jump != nil || node.Call != nil {
		var isAbs bool
		if node.Jump != nil {
			isAbs = node.Jump.JumpType == "abs"
		} else {
			isAbs = node.Call.CallType == "abs"
		}
		if isAbs {
			instr.PcUpdate = PcUpdateJump
		} else {
			instr.PcUpdate = PcUpdateJumpRel
		}
	} else if node.Jnz != nil {
		instr.PcUpdate = PcUpdateJnz
	} else if node.Ret != nil {
		instr.PcUpdate = PcUpdateJump
	}

	// Set ApUpdate
	if node.ApPlus != nil {
		instr.ApUpdate = AddRes
	} else if node.ApPlusOne {
		instr.ApUpdate = Add1
	}

	// Set Opcode
	if node.Call != nil {
		instr.Opcode = OpCodeCall
	} else if node.Ret != nil {
		instr.Opcode = OpCodeRet
	} else if node.AssertEq != nil {
		instr.Opcode = OpCodeAssertEq
	}
	return nil
}
