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

/*
* Casm to instruction list in assembler.go
* Instruction list to bytecode in instruction.go
 */
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

//
// Functions that visit the AST in order to encode Instruction list.
//

func astToInstruction(ast *CasmProgram) ([]Instruction, error) {
	// Vist ast
	n := len(ast.Ast)
	// Slice with length 0 and capacity n
	instructionList := make([]Instruction, 0, n)
	// iterate over the AST
	for i := range ast.Ast {
		instruction, err := encodeNodeToInstr(ast.Ast[i])
		if err != nil {
			return nil, err
		}
		// Append instruction to list
		instructionList = append(instructionList, instruction)
	}
	return instructionList, nil
}

func encodeNodeToInstr(node AstNode) (Instruction, error) {
	var instr Instruction
	expr := node.Expression()
	encodeDst(&node, &instr)
	encodeOp0(&node, &instr, expr)
	encodeOp1(&node, &instr, expr)
	encodeFlags(&node, &instr, expr)
	return instr, nil
}

func encodeDst(node *AstNode, instr *Instruction) {
	if node.ApPlus != nil || node.Jump != nil {
		// dstOffset is not involved so it is set to fp - 1 as default value
		instr.OffDest = -1
		instr.DstRegister = 0x01
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
		instr.DstRegister = 0x01
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
		instr.DstRegister = 0x01
	}
}

func encodeOp0(node *AstNode, instr *Instruction, expr Expressioner) {
	if node != nil && node.Call != nil {
		// op0 is set as [ap + 1] to store current pc
		instr.OffOp0 = 1
		return
	}
	if (node != nil && (node.Jnz != nil || node.Ret != nil)) ||
		(expr.AsDeref() != nil || expr.AsImmediate() != nil) {
		// op0 is not involved, it is set as fp - 1 as default value
		// instr.UOffOp0 = biasedMinusOne
		instr.OffOp0 = -1
		instr.Op0Register = 0x01
		return
	}

	var deref *Deref
	if expr.AsDoubleDeref() != nil {
		deref = expr.AsDoubleDeref().Deref
	} else {
		deref = expr.AsMathOperation().Lhs
	}

	// biasedOffset, err := deref.BiasedOffset()
	offset, err := deref.SignedOffset()
	if err != nil {
		return
	}
	//encode |= uint64(biasedOffset) << op0Offset
	instr.OffOp0 = offset
	if deref.IsFp() {
		//encode |= 1 << op0RegBit
		instr.Op0Register = 1
	}
}

// Given the expression and the current encode returns an updated encode with the corresponding bit
// and offset of op1, an immediate if exists, and a possible error
func encodeOp1(node *AstNode, instr *Instruction, expr Expressioner) {
	if node != nil && node.Ret != nil {
		// op1 is set as [fp - 1], where we read the previous pc
		// instr.UOffOp1 = biasedMinusOne
		instr.OffOp1 = -1
		instr.Op1Source = 0x02
		return
	}

	if expr.AsDeref() != nil {
		offset, err := expr.AsDeref().SignedOffset()
		if err != nil {
			return
		}
		// instr.UOffOp1 = biasedOffset
		instr.OffOp1 = offset
		if expr.AsDeref().IsFp() {
			instr.Op1Source = 0x02
		} else {
			instr.Op1Source = 0x04
		}
		return
	} else if expr.AsDoubleDeref() != nil {
		// biasedOffset, err := expr.AsDoubleDeref().BiasedOffset()
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
		instr.Op1Source = 0x01
		instr.Imm = *imm
		return
	} else {
		//  if it is a math operation, the op1 source is set by the right hand side
		encodeOp1(node, instr, expr.AsMathOperation().Rhs)
	}
}

func encodeFlags(node *AstNode, instr *Instruction, expression Expressioner) {
	// Encode ResLogic
	if expression != nil && expression.AsMathOperation() != nil {
		if expression.AsMathOperation().Operator == "+" {
			instr.Res = 0x01
		} else {
			instr.Res = 0x02
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
			instr.PcUpdate = 0x01
		} else {
			instr.PcUpdate = 0x02
		}
	} else if node.Jnz != nil {
		instr.PcUpdate = 0x04
	} else if node.Ret != nil {
		instr.PcUpdate = 0x01
	}

	// Encode ApUpdate
	if node.ApPlus != nil {
		instr.ApUpdate = 0x01
	} else if node.ApPlusOne {
		instr.ApUpdate = 0x02
	}

	// Encode Opcode
	if node.Call != nil {
		instr.Opcode = 0x01
	} else if node.Ret != nil {
		instr.Opcode = 0x02
	} else if node.AssertEq != nil {
		instr.Opcode = 0x04
	}
}
