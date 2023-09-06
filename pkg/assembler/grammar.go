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
	ApPlusOne bool             `(","@"ap++")?";"`
}

type CoreInstruction struct {
	AssertEq *AssertEq `@@ |`
	Jump     *Jump     `@@ |`
	Jnz      *Jnz      `@@ |`
	Call     *Call     `@@ |`
	Ret      *Ret      `@@ |`
	ApPlus   *ApPlus   `@@`
}

type AssertEq struct {
	Lhs *Deref      `@@`
	Rhs *Expression `"=" @@`
}

type Jump struct {
	JumpType string      `"jmp" @("rel" | "abs")`
	Value    *Expression `@@`
}

type Jnz struct {
	Value     Expression `"jmp" "rel" @@`
	Condition Expression `"if" @@ "!=" "0"`
}

type Call struct {
	CallType string     `"call" @("rel" | "abs")`
	Address  Expression `@@`
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
	Immediate     *int           `@Int`
}

type Deref struct {
	Name   string `"[" @("ap" | "fp")`
	Sign   string `@("+" | "-")?`
	Offset *int   `@Int? "]"`
}

type DoubleDeref struct {
	Deref  *Deref `"[" @@`
	Sign   string `@("+" | "-")?`
	Offset *int   `@Int? "]"`
}

type MathOperation struct {
	Register   *Deref      `@@`
	Operation  string      `("+" | "-")`
	DerefOrImm *DerefOrImm `@@`
}

type DerefOrImm struct {
	Deref     *Deref `@@ |`
	Immediate *int   `@Int`
}

// AST Functionality

func (deref *Deref) ParseOffset() (uint16, error) {
	if deref.Offset == nil {
		return 0, nil
	}
	offset := *deref.Offset
	if deref.Sign == "-" {
		offset = -offset
	} else if deref.Sign != "+" {
		return 0, fmt.Errorf("missing sign in deref offset")
	}

	if offset > math.MaxInt16 || offset < math.MinInt16 {
		return 0, fmt.Errorf("offset value outside of (-2**16, 2**16)")
	}

	biasedOffset := uint16(offset) ^ 0x8000
	return biasedOffset, nil
}
