package assembler

// import "github.com/alecthomas/participle/v2"

type CasmProgram struct {
	Instructions []Instruction
}

type Instruction struct {
	Core      CoreInstruction "@@"
	ApPlusOne bool            `(,@"ap++")?;`
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
	Lhs Deref      `@@`
	Rhs Expression `"=" @@`
}

type Jump struct {
	JumpType string     `"jmp" @Ident`
	Value    Expression `@@`
}

type Jnz struct {
	Value     Expression `"jmp" "rel" @Ident`
	Condition Expression `"if" @@ "!=" "0"`
}

type Call struct {
	Address Expression
}

type Ret struct {
	Ret string `"ret"`
}

type ApPlus struct {
	Value *Expression `"ap" "+=" @@`
}

type Expression struct {
	Deref       *Deref       `@@ |`
	DoubleDeref *DoubleDeref `@@ |`
	Immediate   *int         `@Int`
}

type Deref struct {
	Name   string `"[" "@Ident"`
	Sign   string `@("+" | "-")?`
	Offset *int   `@Int? "]"`
}

type DoubleDeref struct {
	Deref  *Deref `"[" @@"`
	Sign   string `@("+" | "-")?`
	Offset *int   `@Int? "]"`
}

type MathOperation struct {
	Register   *Deref      `@@`
	Operation  string      `("+" | "-")`
	DerefOrImm *DerefOrImm `@@`
}

type DerefOrImm struct {
	Deref     *Deref "@@ |"
	Immediate *int   "@Int"
}
