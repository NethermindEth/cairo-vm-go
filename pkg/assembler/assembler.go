package assembler

import (
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

func casmAstToBytecode(casmAst CasmProgram) ([]*f.Element, error) {
	n := len(casmAst.Instructions)
	bytecode := make([]*f.Element, 0, n+(n/2)+1)
	for i := range casmAst.Instructions {

	}

	return bytecode, nil
}

func instructionToBytecode(instruction Instruction) (*f.Element, error) {

}
