package assembler

import (
	"fmt"
	"math"
)

// Grammar and AST

type CasmProgram struct {
	InstructionList []InstructionNode `@@*`
}

type InstructionNode struct {
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
	DoubleDeref   *DoubleDeref    `@@ |`
	MathOperation *MathOperation  `@@ |`
	Deref         *Deref          `@@ |`
	Immediate     *ImmediateValue `@@`
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

type ImmediateValue struct {
	Sign  string `@("+" | "-")?`
	Value *int   `@Int`
}

// AST Functionality

type CoreInstructioner interface {
	Expression() Expressioner
}

func (instruction InstructionNode) Expression() Expressioner {
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
	if e.Immediate == nil {
		return nil
	}
	if e.Immediate.Sign != "" {
		immediate := signedString(e.Immediate.Sign == "-", *e.Immediate.Value)
		return &immediate
	}
	immediate := signedString(false, *e.Immediate.Value)
	return &immediate
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

func (deref *Deref) SignedOffset() (int16, error) {
	if deref.Offset == nil {
		return 0, nil
	}
	return signedOffset(deref.Offset.Sign == "-", *deref.Offset.Value)
}

func (dderef *DoubleDeref) SignedOffset() (int16, error) {
	if dderef.Offset == nil {
		return 0, nil
	}
	return signedOffset(dderef.Offset.Sign == "-", *dderef.Offset.Value)
}

func signedOffset(neg bool, value int) (int16, error) {
	if neg {
		value = -value
	}
	if value > math.MaxInt16 || value < math.MinInt16 {
		return 0, fmt.Errorf("offset value outside of (-2**16, 2**16)")
	}
	return int16(value), nil
}

func signedString(neg bool, value int) string {
	if neg {
		return fmt.Sprintf("-%d", value)
	}
	return fmt.Sprintf("%d", value)
}
