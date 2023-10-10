package assembler

import (
	"github.com/alecthomas/participle/v2"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

var parser *participle.Parser[CasmProgram] = participle.MustBuild[CasmProgram](
	// mandatory lookahead to disambiguate between productions:
	// expr -> [reg + n] + [reg + m] and
	// expr -> [reg + n]
	participle.UseLookahead(5),
)

func CasmToBytecode(code string) ([]*f.Element, error) {
	casmAst, err := parser.ParseString("", code)
	if err != nil {
		return nil, err
	}
	// Ast To Instruction List
	instructionList, err := astToInstruction(casmAst)
	if err != nil {
		return nil, err
	}
	// Instruction to bytecode
	return encodeInstructionListToBytecode(instructionList)
}

/*
*    Casm to instruction list functions
 */
func astToInstruction(ast *CasmProgram) ([]Instruction, error) {
	// Vist ast
	n := len(ast.InstructionList)
	// Slice with length 0 and capacity n
	instructionList := make([]Instruction, 0, n)
	// iterate over the AST
	for i := range ast.InstructionList {
		instruction, err := encodeNodeToInstr(ast.InstructionList[i])
		if err != nil {
			return nil, err
		}
		// Append instruction to list
		instructionList = append(instructionList, instruction)
	}
	return instructionList, nil
}

func encodeNodeToInstr(node InstructionNode) (Instruction, error) {
	var instr Instruction
	expr := node.Expression()
	encodeDst(&node, &instr)
	encodeOp0(&node, &instr, expr)
	encodeOp1(&node, &instr, expr)
	encodeFlags(&node, &instr, expr)
	return instr, nil
}

func encodeDst(node *InstructionNode, instr *Instruction) {
	if node.ApPlus != nil || node.Jump != nil {
		// dstOffset is not involved so it is set to fp - 1 as default value
		instr.OffDest = -1
		instr.DstRegister = Fp
		return
	}
	if node.Call != nil {
		// dstOffset is set to ap + 0
		instr.OffDest = 0
		return
	}
	if node.Ret != nil {
		// dstOffset is set as fp - 2
		instr.OffDest = -2
		instr.DstRegister = Fp
		return
	}

	var deref *Deref
	if node.AssertEq != nil {
		deref = node.AssertEq.Dst
	} else if node.Jnz != nil {
		deref = node.Jnz.Condition
	}

	offset, err := deref.SignedOffset()
	if err != nil {
		return
	}
	instr.OffDest = offset
	if deref.IsFp() {
		instr.DstRegister = Fp
	} else {
		instr.DstRegister = Ap
	}
}

func encodeOp0(node *InstructionNode, instr *Instruction, expr Expressioner) {
	if node != nil && node.Call != nil {
		// op0 is set as [ap + 1] to store current pc
		instr.OffOp0 = 1
		return
	}
	if (node != nil && (node.Jnz != nil || node.Ret != nil)) ||
		(expr.AsDeref() != nil || expr.AsImmediate() != nil) {
		// op0 is not involved, it is set as fp - 1 as default value
		instr.OffOp0 = -1
		instr.Op0Register = Fp
		return
	}

	var deref *Deref
	if expr.AsDoubleDeref() != nil {
		deref = expr.AsDoubleDeref().Deref
	} else {
		deref = expr.AsMathOperation().Lhs
	}

	offset, err := deref.SignedOffset()
	if err != nil {
		return
	}
	instr.OffOp0 = offset
	if deref.IsFp() {
		instr.Op0Register = Fp
	} else {
		instr.Op0Register = Ap
	}
}

// Given the expression and the current encode returns an updated encode with the corresponding bit
// and offset of op1, an immediate if exists, and a possible error
func encodeOp1(node *InstructionNode, instr *Instruction, expr Expressioner) {
	if node != nil && node.Ret != nil {
		// op1 is set as [fp - 1], where we read the previous pc
		instr.OffOp1 = -1
		instr.Op1Source = FpPlusOffOp1
		return
	}

	if expr.AsDeref() != nil {
		offset, err := expr.AsDeref().SignedOffset()
		if err != nil {
			return
		}
		instr.OffOp1 = offset
		if expr.AsDeref().IsFp() {
			instr.Op1Source = FpPlusOffOp1
		} else {
			instr.Op1Source = ApPlusOffOp1
		}
		return
	} else if expr.AsDoubleDeref() != nil {
		offset, err := expr.AsDoubleDeref().SignedOffset()
		if err != nil {
			return
		}
		instr.OffOp1 = offset
		return
	} else if expr.AsImmediate() != nil {
		// immediate is converted to Felt during bytecode conversion
		imm := expr.AsImmediate()
		instr.OffOp1 = 1
		instr.Op1Source = Imm
		instr.Imm = *imm
		return
	} else {
		//  if it is a math operation, the op1 source is set by the right hand side
		encodeOp1(node, instr, expr.AsMathOperation().Rhs)
	}
}

func encodeFlags(node *InstructionNode, instr *Instruction, expression Expressioner) {
	// Encode ResLogic
	if expression != nil && expression.AsMathOperation() != nil {
		if expression.AsMathOperation().Operator == "+" {
			instr.Res = AddOperands
		} else {
			instr.Res = MulOperands
		}
	}

	// Encode PcUpdate
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

	// Encode ApUpdate
	if node.ApPlus != nil {
		instr.ApUpdate = AddRes
	} else if node.ApPlusOne {
		instr.ApUpdate = Add1
	}

	// Encode Opcode
	if node.Call != nil {
		instr.Opcode = OpCodeCall
	} else if node.Ret != nil {
		instr.Opcode = OpCodeRet
	} else if node.AssertEq != nil {
		instr.Opcode = OpCodeAssertEq
	}
}
