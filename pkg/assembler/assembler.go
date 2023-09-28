package assembler

import (
	"github.com/NethermindEth/cairo-vm-go/pkg/vm"

	"github.com/alecthomas/participle/v2"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

var parser *participle.Parser[CasmProgram] = participle.MustBuild[CasmProgram](
	// mandatory lookahead to disambiguate between productions:
	// expr -> [reg + n] + [reg + m] and
	// expr -> [reg + n]
	participle.UseLookahead(5),
)

//func CasmToBytecode(code string) ([]*f.Element, error) {
//	casmAst, err := parser.ParseString("", code)
//
//	if err != nil {
//		return nil, err
//	}
//	bytecode, _ := encodeCasmProgram(*casmAst)
//	//fmt.Println("Code: ", code)
//	//fmt.Println("CASM AST: ", casmAst)
//	//fmt.Println("bytecode: ", bytecode)
//	return bytecode, nil
//}

func CasmToBytecode(code string) ([]*f.Element, error) {
	casmAst, err := parser.ParseString("", code)
	if err != nil {
		return nil, err
	}
	// Ast To Instruction List
	instructionList, err := astToInstruction(*casmAst)
	// Instrciton to bytecode
	return encodeInstructionListToBytecode(instructionList)
}

type Instruction vm.Instruction

//
// Functions that visit the AST in order to encode the instructions
//

const (
	// offsets
	op0Offset   = 16
	op1Offset   = 32
	flagsOffset = 48

	// flag values
	dstRegBit         = 48
	op0RegBit         = 49
	op1ImmBit         = 50
	op1FpBit          = 51
	op1ApBit          = 52
	resAddBit         = 53
	resMulBit         = 54
	pcJumpAbsBit      = 55
	pcJumpRelBit      = 56
	pcJnzBit          = 57
	apAddBit          = 58
	apAdd1Bit         = 59
	opcodeCallBit     = 60
	opcodeRetBit      = 61
	opcodeAssertEqBit = 62

	// default values
	biasedZero     uint16 = 0x8000
	biasedPlusOne  uint16 = 0x8001
	biasedMinusOne uint16 = 0x7FFF
	biasedMinusTwo uint16 = 0x7FFE
)

// ================================ AST to InstrList ================================================
func astToInstruction(ast CasmProgram) ([]Instruction, error) {
	// Vist ast
	n := len(ast.Ast)
	// Slice with length 0 and capacity n
	instructionList := make([]Instruction, 0, n)
	// iterate over the AST
	// Call encodeInstruction2 to turn the node into a Instruction
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
	// todo: Find a way to first find operation type, then encode regs
	// encode dst
	encodeDst(&node, &instr)
	// encode Op0
	encodeOp0(&node, &instr, expr)
	// encode Op1
	encodeOp1(&node, &instr, expr)
	// encode Flags
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
		//encode |= uint64(biasedPlusOne) << op0Offset
		instr.UOffOp0 = biasedPlusOne
		return
	}
	if (node != nil && (node.Jnz != nil || node.Ret != nil)) ||
		(expr.AsDeref() != nil || expr.AsImmediate() != nil) {
		// op0 is not involved, it is set as fp - 1 as default value
		//encode |= 1 << op0RegBit
		//encode |= uint64(biasedMinusOne) << op0Offset
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
		//encode |= uint64(biasedMinusOne) << op1Offset
		//encode |= 1 << op1FpBit
		instr.UOffOp1 = biasedMinusOne
		instr.Op1Source = 0x02
		return
	}

	if expr.AsDeref() != nil {
		biasedOffset, err := expr.AsDeref().BiasedOffset()
		if err != nil {
			return
		}
		//encode |= uint64(biasedOffset) << op1Offset
		instr.UOffOp1 = biasedOffset
		if expr.AsDeref().IsFp() {
			//encode |= 1 << op1FpBit
			instr.Op1Source = 0x02
		} else {
			//encode |= 1 << op1ApBit
			instr.Op1Source = 0x04
		}
		return
	} else if expr.AsDoubleDeref() != nil {
		biasedOffset, err := expr.AsDoubleDeref().BiasedOffset()
		if err != nil {
			return
		}
		//encode |= uint64(biasedOffset) << op1Offset
		instr.UOffOp1 = biasedOffset
		return
	} else if expr.AsImmediate() != nil {
		//imm, err := new(f.Element).SetString(*expr.AsImmediate())
		// Lets turn immediate to Felt later
		imm := expr.AsImmediate()
		//if err != nil {
		//	return
		//}
		//encode |= uint64(biasedPlusOne) << op1Offset
		instr.UOffOp1 = biasedPlusOne
		instr.Op1Source = 0x01
		instr.Imm = *imm
		//return encode | 1<<op1ImmBit, imm, nil
		return
	} else {
		//  if it is a math operation, the op1 source is set by the right hand side
		//return encodeOp1(node, expr.AsMathOperation().Rhs, encode)
		encodeOp1(node, instr, expr.AsMathOperation().Rhs)
	}
}

func encodeFlags(node *AstNode, instr *Instruction, expression Expressioner) {
	// Encode ResLogic
	if expression != nil && expression.AsMathOperation() != nil {
		if expression.AsMathOperation().Operator == "+" {
			//encode |= 1 << resAddBit
			instr.Res = 0x01
		} else {
			//encode |= 1 << resMulBit
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
			//encode |= 1 << pcJumpAbsBit
			instr.PcUpdate = 0x01
		} else {
			//encode |= 1 << pcJumpRelBit
			instr.PcUpdate = 0x02
		}
	} else if node.Jnz != nil {
		//encode |= 1 << pcJnzBit
		instr.PcUpdate = 0x04
	} else if node.Ret != nil {
		//encode |= 1 << pcJumpAbsBit
		instr.PcUpdate = 0x01
	}
	// Encode ApUpdate
	if node.ApPlus != nil {
		//encode |= 1 << apAddBit
		instr.ApUpdate = 0x01
	} else if node.ApPlusOne {
		//encode |= 1 << apAdd1Bit
		//fmt.Println("HERE")
		instr.ApUpdate = 0x02
	}
	// Encode Opcode
	if node.Call != nil {
		//encode |= 1 << opcodeCallBit
		instr.Opcode = 0x01
	} else if node.Ret != nil {
		//encode |= 1 << opcodeRetBit
		instr.Opcode = 0x02
	} else if node.AssertEq != nil {
		//encode |= 1 << opcodeAssertEqBit
		instr.Opcode = 0x04
	}
}

// ================================ InstrList to Bytecode ================================================
func encodeInstructionListToBytecode(instruction []Instruction) ([]*f.Element, error) {
	n := len(instruction)
	bytecodes := make([]*f.Element, 0, n+(n/2)+1)
	//var err error
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
	rawInstruction := encodeInstructionValues(instruction)

	// Encode the flags
	rawInstruction, err := encodeInstructionFlags(instruction, rawInstruction)
	if err != nil {
		return nil, err
	}

	// Create a new f.Element from the raw instruction
	element := new(f.Element).SetUint64(rawInstruction)

	return element, nil
}

func encodeInstructionValues(instr *Instruction) uint64 {
	// Combine the offsets and flags into a single uint64
	var encoding uint64 = 0
	encoding |= uint64(instr.UOffDest)
	encoding |= uint64(instr.UOffOp0) << op0Offset
	encoding |= uint64(instr.UOffOp1) << op1Offset
	//encoding |= uint64(instr.) << flagsOffset

	// Combine all the offsets and flags
	//encoding := offDstEncUint64 | offOp0EncUint64 | offOp1EncUint64 | flagsUint64

	// Apply the 2's complement encoding. Why this line?
	//encoding ^= encoding ^ 0x0000800080008000

	return encoding
}

func encodeInstructionFlags(instr *Instruction, encoding uint64) (uint64, error) {
	encoding |= uint64(instr.DstRegister) << dstRegBit
	encoding |= uint64(instr.Op0Register) << op0RegBit
	encoding |= uint64(instr.Op1Source) << op1ImmBit
	encoding |= uint64(instr.Res) << resAddBit
	encoding |= uint64(instr.PcUpdate) << pcJumpAbsBit
	encoding |= uint64(instr.ApUpdate) << apAddBit
	encoding |= uint64(instr.Opcode) << opcodeCallBit
	return encoding, nil
}

// ===================================================== OLD ============================================
func encodeCasmProgram(casmAst CasmProgram) ([]*f.Element, error) {
	n := len(casmAst.Ast)
	bytecode := make([]*f.Element, 0, n+(n/2)+1)
	var err error
	for i := range casmAst.Ast {
		bytecode, err = encodeInstruction(bytecode, casmAst.Ast[i])
		if err != nil {
			return nil, err
		}
	}
	return bytecode, nil
}

func encodeInstruction(bytecode []*f.Element, instruction AstNode) ([]*f.Element, error) {
	var encode uint64 = 0
	expression := instruction.Expression()
	//fmt.Println("EXPRESSION: ", expression)

	encode, err := encodeDstReg(&instruction, encode)
	if err != nil {
		return nil, err
	}

	encode, err = encodeOp0Reg(&instruction, expression, encode)
	if err != nil {
		return nil, err
	}

	encode, imm, err := encodeOp1Source(&instruction, expression, encode)
	if err != nil {
		return nil, err
	}

	encode = encodeResLogic(expression, encode) |
		encodePcUpdate(instruction, encode) |
		encodeApUpdate(instruction, encode) |
		encodeOpCode(instruction, encode)

	encodeAsFelt := new(f.Element).SetUint64(encode)

	bytecode = append(bytecode, encodeAsFelt)
	if imm != nil {
		bytecode = append(bytecode, imm)
	}

	return bytecode, nil
}

func encodeDstReg(instr *AstNode, encode uint64) (uint64, error) {
	if instr.ApPlus != nil || instr.Jump != nil {
		// dstOffset is not involved so it is set to fp - 1 as default value
		encode |= 1 << dstRegBit
		encode |= uint64(biasedMinusOne)
		return encode, nil
	}
	if instr.Call != nil {
		// dstOffset is set to ap + 0
		encode |= uint64(biasedZero)
		return encode, nil
	}
	if instr.Ret != nil {
		// dstOffset is set as fp - 2
		encode |= 1 << dstRegBit
		encode |= uint64(biasedMinusTwo)
		return encode, nil
	}

	var deref *Deref
	if instr.AssertEq != nil {
		deref = instr.AssertEq.Dst
	} else if instr.Jnz != nil {
		deref = instr.Jnz.Condition
	}

	biasedOffset, err := deref.BiasedOffset()
	if err != nil {
		return 0, err
	}
	encode |= uint64(biasedOffset)
	if deref.IsFp() {
		encode |= 1 << dstRegBit
	}
	//fmt.Println("HERE: ")
	return encode, nil

}

func encodeOp0Reg(instr *AstNode, expr Expressioner, encode uint64) (uint64, error) {
	if instr != nil && instr.Call != nil {
		// op0 is set as [ap + 1] to store current pc
		encode |= uint64(biasedPlusOne) << op0Offset
		return encode, nil
	}
	if (instr != nil && (instr.Jnz != nil || instr.Ret != nil)) ||
		(expr.AsDeref() != nil || expr.AsImmediate() != nil) {
		// op0 is not involved, it is set as fp - 1 as default value
		encode |= 1 << op0RegBit
		encode |= uint64(biasedMinusOne) << op0Offset
		return encode, nil
	}

	var deref *Deref
	if expr.AsDoubleDeref() != nil {
		deref = expr.AsDoubleDeref().Deref
	} else {
		deref = expr.AsMathOperation().Lhs
	}

	biasedOffset, err := deref.BiasedOffset()
	if err != nil {
		return 0, err
	}
	encode |= uint64(biasedOffset) << op0Offset
	if deref.IsFp() {
		encode |= 1 << op0RegBit
	}

	return encode, nil
}

// Given the expression and the current encode returns an updated encode with the corresponding bit
// and offset of op1, an immediate if exists, and a possible error
func encodeOp1Source(inst *AstNode, expr Expressioner, encode uint64) (uint64, *f.Element, error) {
	if inst != nil && inst.Ret != nil {
		// op1 is set as [fp - 1], where we read the previous pc
		encode |= uint64(biasedMinusOne) << op1Offset
		encode |= 1 << op1FpBit
		return encode, nil, nil
	}

	if expr.AsDeref() != nil {
		biasedOffset, err := expr.AsDeref().BiasedOffset()
		if err != nil {
			return 0, nil, err
		}
		encode |= uint64(biasedOffset) << op1Offset
		if expr.AsDeref().IsFp() {
			encode |= 1 << op1FpBit
		} else {
			encode |= 1 << op1ApBit
		}
		return encode, nil, nil
	} else if expr.AsDoubleDeref() != nil {
		biasedOffset, err := expr.AsDoubleDeref().BiasedOffset()
		if err != nil {
			return 0, nil, err
		}
		encode |= uint64(biasedOffset) << op1Offset
		return encode, nil, nil
	} else if expr.AsImmediate() != nil {
		imm, err := new(f.Element).SetString(*expr.AsImmediate())
		if err != nil {
			return 0, nil, err
		}
		encode |= uint64(biasedPlusOne) << op1Offset
		return encode | 1<<op1ImmBit, imm, nil
	} else {
		//  if it is a math operation, the op1 source is set by the right hand side
		return encodeOp1Source(inst, expr.AsMathOperation().Rhs, encode)
	}
}

func encodeResLogic(expression Expressioner, encode uint64) uint64 {
	if expression != nil && expression.AsMathOperation() != nil {
		if expression.AsMathOperation().Operator == "+" {
			encode |= 1 << resAddBit
		} else {
			encode |= 1 << resMulBit
		}
	}
	return encode
}

func encodePcUpdate(instruction AstNode, encode uint64) uint64 {
	if instruction.Jump != nil || instruction.Call != nil {
		var isAbs bool
		if instruction.Jump != nil {
			isAbs = instruction.Jump.JumpType == "abs"
		} else {
			isAbs = instruction.Call.CallType == "abs"
		}
		if isAbs {
			encode |= 1 << pcJumpAbsBit
		} else {
			encode |= 1 << pcJumpRelBit
		}
	} else if instruction.Jnz != nil {
		encode |= 1 << pcJnzBit
	} else if instruction.Ret != nil {
		encode |= 1 << pcJumpAbsBit
	}

	return encode
}

func encodeApUpdate(instruction AstNode, encode uint64) uint64 {
	if instruction.ApPlus != nil {
		encode |= 1 << apAddBit
	} else if instruction.ApPlusOne {
		encode |= 1 << apAdd1Bit
	}
	return encode
}

func encodeOpCode(instruction AstNode, encode uint64) uint64 {
	if instruction.Call != nil {
		encode |= 1 << opcodeCallBit
	} else if instruction.Ret != nil {
		encode |= 1 << opcodeRetBit
	} else if instruction.AssertEq != nil {
		encode |= 1 << opcodeAssertEqBit
	}
	return encode
}
