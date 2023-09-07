package assembler

import (
	"fmt"
	"math"
)

// Grammar and AST

type CasmProgram struct {
	Instructions []Instruction `@@`
}

type Instruction struct {
	Core      *CoreInstruction `@@`
	ApPlusOne bool             `(","@"ap++")?";" |`
	ApPlus    *ApPlus          `@@ ";"`
}

type CoreInstruction struct {
	AssertEq *AssertEq `@@ |`
	Jump     *Jump     `@@ |`
	Jnz      *Jnz      `@@ |`
	Call     *Call     `@@ |`
	Ret      *Ret      `@@ `
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
	Condition *Deref      `"if" @@ "!=" "0"`
}

type Call struct {
	CallType string      `"call" @("rel" | "abs")`
	Value    *DerefOrImm `@@`
}

type Ret struct {
	Ret string `"ret"`
}

type ApPlus struct {
	Value *Expression `"ap" "+=" @@`
}

type Expression struct {
	Deref         *Deref         `@@ |`
	DoubleDeref   *DoubleDeref   `@@ |`
	MathOperation *MathOperation `@@ |`
	Immediate     *string        `@String`
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
	Operator string      `("+" | "*")`
	Rhs      *DerefOrImm `@@`
}

type DerefOrImm struct {
	Deref     *Deref  `@@ |`
	Immediate *string `@String`
}

// AST Functionality

type CoreInstructioner interface {
	Expression() Expressioner
}

func (instruction Instruction) Unwrap() CoreInstructioner {
	if instruction.ApPlus != nil {
		return instruction.ApPlus
	} else if instruction.Core.AssertEq != nil {
		return instruction.Core.AssertEq
	} else if instruction.Core.Jump != nil {
		return instruction.Core.Jump
	} else if instruction.Core.Jnz != nil {
		return instruction.Core.Jnz
	} else if instruction.Core.Call != nil {
		return instruction.Core.Call
	} else {
		return instruction.Core.Ret
	}
}

func (assertEq *AssertEq) Expression() Expressioner {
	return assertEq.Value
}

func (jump *Jump) Expression() Expressioner {
	return jump.Value
}

func (jnz *Jnz) Expression() Expressioner {
	return jnz.Value
}

func (call *Call) Expression() Expressioner {
	return call.Value
}

func (ret *Ret) Expression() Expressioner {
	return nil
}
func (apPlus *ApPlus) Expression() Expressioner {
	return apPlus.Value
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
		return 0, nil
	}
	return biasedOffset(deref.Offset.Sign == "-", *deref.Offset.Value)
}

func (dderef *DoubleDeref) BiasedOffset() (uint16, error) {
	if dderef.Offset == nil {
		return 0, nil
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
