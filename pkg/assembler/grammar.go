package assembler

import (
	"fmt"
	"math"
)

// Grammar and AST

type CasmProgram struct {
	Ast []AstNode `@@*`
}

type AstNode struct {
	AssertEq  *AssertEq `( @@ |`
	Jnz       *Jnz      `  @@ |`
	Jump      *Jump     `  @@ )`
	ApPlusOne bool      `( "," @"ap" "+" "+" )? ";" |`
	Call      *Call     `( @@ |`
	Ret       *Ret      `  @@ |`
	ApPlus    *ApPlus   `  @@ ) ";"`
}

type AssertEq struct {
	Dst   *Deref      `@@`
	Value *Expression `"=" @@`
}

type Jump struct {
	JumpType string      `"jmp" @("rel" | "abs")`
	Value    *Expression `@@`
}

type Jnz struct {
	Value     *DerefOrImm `"jmp" "rel" @@`
	Condition *Deref      `"if" @@ "!" "=" "0"`
}

type Call struct {
	CallType string      `"call" @("rel" | "abs")`
	Value    *DerefOrImm `@@`
}

type Ret struct {
	Ret string `"ret"`
}

type ApPlus struct {
	Value *Expression `"ap" "+" "=" @@`
}

type Expression struct {
	DoubleDeref   *DoubleDeref   `@@ |`
	MathOperation *MathOperation `@@ |`
	Deref         *Deref         `@@ |`
	Immediate     *string        `@Int`
}

type Deref struct {
	Name   string  `"[" @("ap" | "fp")`
	Offset *Offset `@@? "]"`
}

type DoubleDeref struct {
	Deref  *Deref  `"[" @@`
	Offset *Offset `@@? "]"`
}

type Offset struct {
	Sign  string `@("+" | "-")`
	Value *int   `@Int`
}

type MathOperation struct {
	Lhs      *Deref      `@@`
	Operator string      `@("+" | "*")`
	Rhs      *DerefOrImm `@@`
}

type DerefOrImm struct {
	Deref     *Deref  `@@ |`
	Immediate *string `@Int`
}

// AST Functionality

type CoreInstructioner interface {
	Expression() Expressioner
}

func (instruction AstNode) Expression() Expressioner {
	switch {
	case instruction.AssertEq != nil:
		return instruction.AssertEq.Value
	case instruction.Jump != nil:
		return instruction.Jump.Value
	case instruction.Jnz != nil:
		return instruction.Jnz.Value
	case instruction.Call != nil:
		return instruction.Call.Value
	case instruction.ApPlus != nil:
		return instruction.ApPlus.Value
	default:
		// when instruction is Ret
		return nil
	}
}

type Expressioner interface {
	AsDeref() *Deref
	AsDoubleDeref() *DoubleDeref
	AsMathOperation() *MathOperation
	AsImmediate() *string
}

func (e *Expression) AsDeref() *Deref {
	return e.Deref
}

func (e *Expression) AsDoubleDeref() *DoubleDeref {
	return e.DoubleDeref
}

func (e *Expression) AsMathOperation() *MathOperation {
	return e.MathOperation
}

func (e *Expression) AsImmediate() *string {
	return e.Immediate
}

func (di *DerefOrImm) AsDeref() *Deref {
	return di.Deref
}

func (di *DerefOrImm) AsDoubleDeref() *DoubleDeref {
	return nil
}

func (di *DerefOrImm) AsMathOperation() *MathOperation {
	return nil
}
func (di *DerefOrImm) AsImmediate() *string {
	return di.Immediate
}

func (deref *Deref) IsFp() bool {
	return deref.Name == "fp"
}

func (deref *Deref) BiasedOffset() (uint16, error) {
	if deref.Offset == nil {
		return biasedZero, nil
	}
	return biasedOffset(deref.Offset.Sign == "-", *deref.Offset.Value)
}

func (dderef *DoubleDeref) BiasedOffset() (uint16, error) {
	if dderef.Offset == nil {
		return biasedZero, nil
	}
	return biasedOffset(dderef.Offset.Sign == "-", *dderef.Offset.Value)
}

func biasedOffset(neg bool, value int) (uint16, error) {
	if neg {
		value = -value
	}
	if value > math.MaxInt16 || value < math.MinInt16 {
		return 0, fmt.Errorf("offset value outside of (-2**16, 2**16)")
	}
	biasedOffset := uint16(value) ^ 0x8000
	return biasedOffset, nil
}
