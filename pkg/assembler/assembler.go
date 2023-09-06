package assembler

import (
	"fmt"

	"github.com/alecthomas/participle/v2"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

var parser *participle.Parser[CasmProgram]

func CasmToBytecode(code string) ([]*f.Element, error) {
	if parser == nil {
		var err error
		parser, err = participle.Build[CasmProgram]()
		if err != nil {
			return nil, err
		}
	}

	casmAst, err := parser.ParseString("", code)
	if err != nil {
		return nil, err
	}

	return casmAstToBytecode(*casmAst)
}

//
// Functions that visit the AST in order to encode the instructions
//

const (
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
)

func casmAstToBytecode(casmAst CasmProgram) ([]*f.Element, error) {
	n := len(casmAst.Instructions)
	bytecode := make([]*f.Element, 0, n+(n/2)+1)
	for i := range casmAst.Instructions {
		err := instructionToBytecode(bytecode, casmAst.Instructions[i])
		if err != nil {
			return nil, err
		}
	}

	return bytecode, nil
}

func instructionToBytecode(bytecode []*f.Element, instruction Instruction) error {
	err := coreInstructionToBytecode(bytecode, instruction.Core)
	if err != nil {
		return err
	}

	if instruction.ApPlusOne {
		// check encoded instruction different than ap plus imm
		// and put app flag  == 2
	}

	return nil
}

func coreInstructionToBytecode(
	bytecode []*f.Element, instruction CoreInstruction,
) error {
	// If else over the different instructions type
	if instruction.AssertEq != nil {
		return assertEqToBytecode(bytecode, instruction.AssertEq)
	}
	if instruction.Jump != nil {

	}
	if instruction.Jnz != nil {

	}
	if instruction.Call != nil {

	}
	if instruction.Ret != nil {

	}
	if instruction.ApPlus != nil {

	}

	// this should never execute
	return fmt.Errorf("no core instruction detected")
}

func assertEqToBytecode(bytecode []*f.Element, assertEq *AssertEq) ([]*f.Element, error) {
	var encoded uint64 = 0

	// set opcode
	encoded = encoded | (1 << opcodeAssertEqBit)

	deref := assertEq.Lhs
	// set dst registry
	if deref.Name == "fp" {
		encoded = encoded | (1 << dstRegBit)
	} else if deref.Name != "ap" {
		return nil, fmt.Errorf("Unknown registry %s", deref.Name)
	}

	// set dst reg
	encoded, err := encodeDerefReg(encoded, deref, dstRegBit)
	if err != nil {
		return nil, err
	}
	// set dst offset
	dstOffset, err := deref.ParseOffset()
	if err != nil {
		return nil, err
	}
	encoded = encoded | uint64(dstOffset)

	// set op0 reg
	if assertEq.Rhs.Deref != nil {
		rhs := assertEq.Rhs.Deref
		encoded, err = encodeDerefReg(encoded, rhs, op0RegBit)
		if err != nil {
			return nil, err
		}
		op0Offset, err := rhs.ParseOffset()
		if err != nil {
			return nil, err
		}
		encoded = encoded | (uint64(op0Offset) << 16)
	} else {
		return nil, fmt.Errorf("Unknown expresion")
	}

	bytecode = append(bytecode, new(f.Element).SetUint64(encoded))
	return bytecode, nil
}

func encodeDerefReg(encoded uint64, deref *Deref, bit int) (uint64, error) {
	if deref.Name == "fp" {
		return encoded | (1 << dstRegBit), nil
	} else if deref.Name == "ap" {
		return encoded, nil
	}
	return 0, fmt.Errorf("Unknown registry %s", deref.Name)

}
