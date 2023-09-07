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
	return encodeCasmProgram(*casmAst)
}

//
// Functions that visit the AST in order to encode the instructions
//

const (
	// offsets
	op0Offset = 16
	op1Offset = 32

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

func encodeCasmProgram(casmAst CasmProgram) ([]*f.Element, error) {
	n := len(casmAst.Instructions)
	bytecode := make([]*f.Element, 0, n+(n/2)+1)
	var err error
	for i := range casmAst.Instructions {
		bytecode, err = encodeInstruction(bytecode, casmAst.Instructions[i])
		if err != nil {
			return nil, err
		}
	}
	return bytecode, nil
}

func encodeInstruction(bytecode []*f.Element, instruction Instruction) ([]*f.Element, error) {
	var encode uint64 = 0
	expression := instruction.Unwrap().Expression()

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

func encodeDstReg(instr *Instruction, encode uint64) (uint64, error) {
	if instr.ApPlus != nil || instr.Core.Jump != nil {
		// dstOffset is not involved so it is set to fp - 1 as default value
		encode |= 1 << dstRegBit
		encode |= uint64(biasedMinusOne)
		return encode, nil
	}
	if instr.Core.Call != nil {
		// dstOffset is set to ap + 0
		encode |= uint64(biasedZero)
		return encode, nil
	}
	if instr.Core.Ret != nil {
		// dstOffset is set as fp - 2
		encode |= 1 << dstRegBit
		encode |= uint64(biasedMinusTwo)
		return encode, nil
	}

	var deref *Deref
	if instr.Core.AssertEq != nil {
		deref = instr.Core.AssertEq.Dst
	} else if instr.Core.Jnz != nil {
		deref = instr.Core.Jnz.Condition
	}

	biasedOffset, err := deref.BiasedOffset()
	if err != nil {
		return 0, err
	}
	encode |= uint64(biasedOffset)
	if deref.IsFp() {
		encode |= 1 << dstRegBit
	}

	return encode, nil

}

func encodeOp0Reg(instr *Instruction, expr Expressioner, encode uint64) (uint64, error) {
	if instr.Core != nil && instr.Core.Call != nil {
		// op0 is set as [ap + 1] to store current pc
		encode |= uint64(biasedPlusOne) << op0Offset
		return encode, nil
	}
	if (instr.Core != nil && (instr.Core.Jnz != nil || instr.Core.Ret != nil)) ||
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
// and offset of op1, an immeadiate if exists, and a possible error
func encodeOp1Source(inst *Instruction, expr Expressioner, encode uint64) (uint64, *f.Element, error) {
	if inst.Core != nil && inst.Core.Ret != nil {
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

func encodePcUpdate(instruction Instruction, encode uint64) uint64 {
	if instruction.Core == nil {
		return encode
	}

	if instruction.Core.Jump != nil || instruction.Core.Call != nil {
		var isAbs bool
		if instruction.Core.Jump != nil {
			isAbs = instruction.Core.Jump.JumpType == "abs"
		} else {
			isAbs = instruction.Core.Call.CallType == "abs"
		}
		if isAbs {
			encode |= 1 << pcJumpAbsBit
		} else {
			encode |= 1 << pcJumpRelBit
		}
	} else if instruction.Core.Jnz != nil {
		encode |= 1 << pcJnzBit
	} else if instruction.Core.Ret != nil {
		encode |= 1 << pcJumpAbsBit
	}

	return encode
}

func encodeApUpdate(instruction Instruction, encode uint64) uint64 {
	if instruction.ApPlus != nil {
		encode |= 1 << apAddBit
	} else if instruction.ApPlusOne {
		encode |= 1 << apAdd1Bit
	}
	return encode
}

func encodeOpCode(instruction Instruction, encode uint64) uint64 {
	if instruction.Core != nil {
		if instruction.Core.Call != nil {
			encode |= 1 << opcodeCallBit
		} else if instruction.Core.Ret != nil {
			encode |= 1 << opcodeRetBit
		} else if instruction.Core.AssertEq != nil {
			encode |= 1 << opcodeAssertEqBit
		}
	}
	return encode
}
