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
	instructionList, err := astToInstruction(*casmAst)
	if err != nil {
		return nil, err
	}
	// Instrciton to bytecode
	return encodeInstructionListToBytecode(instructionList)
}

//
// Functions that visit the AST in order to encode the instructions
//

// const (
// 	// offsets
// 	op0Offset   = 16
// 	op1Offset   = 32
// 	flagsOffset = 48

// 	// flag values
// 	dstRegBit         = 48
// 	op0RegBit         = 49
// 	op1ImmBit         = 50
// 	op1FpBit          = 51
// 	op1ApBit          = 52
// 	resAddBit         = 53
// 	resMulBit         = 54
// 	pcJumpAbsBit      = 55
// 	pcJumpRelBit      = 56
// 	pcJnzBit          = 57
// 	apAddBit          = 58
// 	apAdd1Bit         = 59
// 	opcodeCallBit     = 60
// 	opcodeRetBit      = 61
// 	opcodeAssertEqBit = 62

// 	// default values
// 	biasedZero     uint16 = 0x8000
// 	biasedPlusOne  uint16 = 0x8001
// 	biasedMinusOne uint16 = 0x7FFF
// 	biasedMinusTwo uint16 = 0x7FFE
// )

// ================================ AST to InstrList ================================================
func astToInstruction(ast CasmProgram) ([]Instruction, error) {
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
	// todo: Find a way to first find operation type, then encode regs?
	encodeDst(&node, &instr)
	encodeOp0(&node, &instr, expr)
	encodeOp1(&node, &instr, expr)
	encodeFlags(&node, &instr, expr)
	return instr, nil
}

func encodeDst(node *AstNode, instr *Instruction) {
	if node.ApPlus != nil || node.Jump != nil {
		// dstOffset is not involved so it is set to fp - 1 as default value
		instr.UOffDest = biasedMinusOne
		instr.DstRegister = 0x01
		return
	}
	if node.Call != nil {
		// dstOffset is set to ap + 0
		instr.UOffDest = biasedZero
		return
	}
	if node.Ret != nil {
		// dstOffset is set as fp - 2
		instr.UOffDest = biasedMinusTwo
		instr.DstRegister = 0x01
		return
	}

	var deref *Deref
	if node.AssertEq != nil {
		deref = node.AssertEq.Dst
	} else if node.Jnz != nil {
		deref = node.Jnz.Condition
	}

	biasedOffset, err := deref.BiasedOffset()
	if err != nil {
		return
	}
	//encode |= uint64(biasedOffset)
	instr.UOffDest = biasedOffset
	if deref.IsFp() {
		instr.DstRegister = 0x01
	}
}

func encodeOp0(node *AstNode, instr *Instruction, expr Expressioner) {
	if node != nil && node.Call != nil {
		// op0 is set as [ap + 1] to store current pc
		instr.UOffOp0 = biasedPlusOne
		return
	}
	if (node != nil && (node.Jnz != nil || node.Ret != nil)) ||
		(expr.AsDeref() != nil || expr.AsImmediate() != nil) {
		// op0 is not involved, it is set as fp - 1 as default value
		instr.UOffOp0 = biasedMinusOne
		instr.Op0Register = 0x01
		return
	}

	var deref *Deref
	if expr.AsDoubleDeref() != nil {
		deref = expr.AsDoubleDeref().Deref
	} else {
		deref = expr.AsMathOperation().Lhs
	}

	biasedOffset, err := deref.BiasedOffset()
	if err != nil {
		return
	}
	//encode |= uint64(biasedOffset) << op0Offset
	instr.UOffOp0 = biasedOffset
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
		instr.UOffOp1 = biasedMinusOne
		instr.Op1Source = 0x02
		return
	}

	if expr.AsDeref() != nil {
		biasedOffset, err := expr.AsDeref().BiasedOffset()
		if err != nil {
			return
		}
		instr.UOffOp1 = biasedOffset
		if expr.AsDeref().IsFp() {
			instr.Op1Source = 0x02
		} else {
			instr.Op1Source = 0x04
		}
		return
	} else if expr.AsDoubleDeref() != nil {
		biasedOffset, err := expr.AsDoubleDeref().BiasedOffset()
		if err != nil {
			return
		}
		instr.UOffOp1 = biasedOffset
		return
	} else if expr.AsImmediate() != nil {
		// immediate is converted to Felt during bytecode conversion
		imm := expr.AsImmediate()
		instr.UOffOp1 = biasedPlusOne
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

// ================================ InstrList to Bytecode ================================================
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
	// Combine the offsets and flags into a single uint64
	var encoding uint64 = 0
	encoding |= uint64(instr.UOffDest)
	encoding |= uint64(instr.UOffOp0) << op0Offset
	encoding |= uint64(instr.UOffOp1) << op1Offset

	return encoding
}

func encodeInstructionFlags(instr *Instruction, encoding uint64) (uint64, error) {
	// Use flag register to encode the flags
	var flagsReg uint16
	// Encode the flag bits
	flagsReg = uint16(instr.DstRegister) << dstRegBit
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
