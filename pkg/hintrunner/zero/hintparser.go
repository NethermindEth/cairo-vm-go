package zero

import (
	"fmt"
	"math/big"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

// Hint references follow the format "cast(<expression>, <type>)". It also allows an
// external dereference such as "[cast(<expression>, <type>)]". The <expression> in
// the hint reference is interpreted as an arithmetic expression, so the root of the
// grammar defined in this file would be `arithExp`
//
// Grammar:
// arithExp => term (('+'|'-') term)*
// term     => exp | prodExp
// prodExp  => exp '*' exp
// exp      => cellRef | deref | dderef | int
// cellRef  => ('ap'|'fp') ('+'|'-') int
// deref    => [cellRef]
// dderef   => [deref ('+'|'-') int]

var (
	basicLexer = lexer.MustSimple([]lexer.SimpleRule{
		{"Number", `\d+`},
		{"Ident", `[a-zA-Z_]\w*`},
		{"Punct", `[-[!@#$%^&*()+_={}\|:;"'<,>.?/]|]`},
		{"whitespace", `[ \t]+`},
	})
	parser = participle.MustBuild[IdentifierExp](
		participle.Lexer(basicLexer),
		participle.UseLookahead(20),
	)
)

type IdentifierExp struct {
	DerefCastExp *DerefCastExp `@@ |`
	CastExp      *CastExp      `@@`
}

type DerefCastExp struct {
	CastExp *CastExp `"[" @@ "]"`
}

type CastExp struct {
	ValueExp *ArithExp `"cast" "(" @@ ","`
	CastType []string  `@Ident ("." @Ident)* ("*")? ("*")? ")"`
}

type ArithExp struct {
	TermExp *TermExp `@@`
	AddExp  []AddExp `@@*`
}

type AddExp struct {
	Operator string   `@("+" | "-")`
	TermExp  *TermExp `@@`
}

type TermExp struct {
	ProdExp *ProdExp    `@@ |`
	Exp     *Expression `@@`
}

type ProdExp struct {
	LeftExp  *Expression `@@`
	Operator string      `"*"`
	RightExp *Expression `@@`
}

type Expression struct {
	DDerefExp  *DDerefExp  `@@ |`
	DerefExp   *DerefExp   `@@ |`
	CellRefExp *CellRefExp `@@ |`
	IntExp     *OffsetExp  `@@`
}

// CellRefSimple represents the structure of a CellRef in its natural form.
// A CellRefSimple cannot be an Expression by itself if it has an offset,
// since the parser will interpret this as a sum of terms instead.
// That's why CellRefExp is also defined. Notice that in the case where there
// is an offset, the whole expression is expected to be enclosed in parenthesis.
type CellRefSimple struct {
	RegisterOffset *RegisterOffset `@@ |`
	Register       string          `@("ap" | "fp")`
}

type CellRefExp struct {
	RegisterOffset *RegisterOffset `"(" @@ ")" |`
	Register       string          `@("ap" | "fp")`
}

type RegisterOffset struct {
	Register string     `@("ap" | "fp")`
	Operator string     `@("+" | "-")`
	Offset   *OffsetExp `@@`
}

type OffsetExp struct {
	Number    string `@Number |`
	NegNumber string `"(" "-" @Number ")"`
}

type DerefExp struct {
	CellRefExp *CellRefSimple `"[" @@ "]"`
}

type DerefOffsetExp struct {
	DerefExp *DerefExp  `@@`
	Operator string     `@("+" | "-")`
	Offset   *OffsetExp `@@`
}

type DDerefExp struct {
	DerefOffsetExp *DerefOffsetExp `"[" @@ "]" |`
	DerefExp       *DerefExp       `"[" @@ "]"`
}

// AST Functionality
func (expression IdentifierExp) Evaluate() (hinter.Reference, error) {
	switch {
	case expression.DerefCastExp != nil:
		return expression.DerefCastExp.Evaluate()
	case expression.CastExp != nil:
		return expression.CastExp.Evaluate()
	default:
		return nil, fmt.Errorf("unexpected identifier value")
	}
}

func (expression DerefCastExp) Evaluate() (hinter.Reference, error) {
	value, err := expression.CastExp.ValueExp.Evaluate()
	if err != nil {
		return nil, err
	}

	switch result := value.(type) {
	case hinter.CellRefer:
		return hinter.Deref{Deref: result}, nil
	case hinter.Deref:
		return hinter.DoubleDeref{
				Deref:  result,
				Offset: 0,
			},
			nil
	case hinter.BinaryOp:
		if left, ok := result.Lhs.(hinter.Deref); ok {
			if right, ok := result.Rhs.(hinter.Immediate); ok {
				if offset, ok := utils.Int16FromFelt((*fp.Element)(&right)); ok {
					return hinter.DoubleDeref{
							Deref:  left,
							Offset: offset,
						},
						nil
				}
			}
		}
		return nil, fmt.Errorf("invalid binary operation inside a deref")
	default:
		return nil, fmt.Errorf("unexpected deref expression")
	}
}

func (expression CastExp) Evaluate() (hinter.Reference, error) {
	return expression.ValueExp.Evaluate()
}

func (expression ArithExp) Evaluate() (hinter.Reference, error) {
	leftExp, err := expression.TermExp.Evaluate()
	if err != nil {
		return nil, err
	}

	if leftResult, ok := leftExp.(hinter.CellRefer); ok {
		// Binary Operation does not support CellRef in the left hand side
		// so the expression has to follow the pattern:
		// reg + off + off + ... + off
		for _, term := range expression.AddExp {
			rightExp, err := term.TermExp.Evaluate()
			if err != nil {
				return nil, err
			}
			rightResult, ok := rightExp.(hinter.Immediate)
			if !ok {
				return nil, fmt.Errorf("invalid arithmetic expression")
			}

			off, ok := utils.Int16FromFelt((*fp.Element)(&rightResult))
			if !ok {
				return nil, fmt.Errorf("invalid arithmetic expression")
			}

			if term.Operator == "-" {
				off = -off
			}

			switch cellRef := leftResult.(type) {
			case hinter.ApCellRef:
				oldOffset := int16(cellRef)
				leftResult = hinter.ApCellRef(off + oldOffset)
				continue
			case hinter.FpCellRef:
				oldOffset := int16(cellRef)
				leftResult = hinter.FpCellRef(off + oldOffset)
				continue
			}
		}
		return leftResult, nil
	} else {
		for _, term := range expression.AddExp {
			rightExp, err := term.TermExp.Evaluate()
			if err != nil {
				return nil, err
			}

			op, err := parseOperator(term.Operator)
			if err != nil {
				return nil, err
			}

			if leftResult, ok := leftExp.(hinter.ResOperander); ok {
				if rightResult, ok := rightExp.(hinter.ResOperander); ok {
					leftExp = hinter.BinaryOp{
						Operator: op,
						Lhs:      leftResult,
						Rhs:      rightResult,
					}
					continue
				}
			}
			return nil, fmt.Errorf("invalid arithmetic expression")
		}
		return leftExp, nil
	}

}

func (expression TermExp) Evaluate() (hinter.Reference, error) {
	switch {
	case expression.ProdExp != nil:
		return expression.ProdExp.Evaluate()
	case expression.Exp != nil:
		return expression.Exp.Evaluate()
	default:
		return nil, fmt.Errorf("unexpected identifier value")
	}
}

func (expression ProdExp) Evaluate() (hinter.Reference, error) {
	leftExp, err := expression.LeftExp.Evaluate()
	if err != nil {
		return nil, err
	}
	rightExp, err := expression.RightExp.Evaluate()
	if err != nil {
		return nil, err
	}

	if leftOp, ok := leftExp.(hinter.ResOperander); ok {
		if rightOp, ok := rightExp.(hinter.ResOperander); ok {
			return hinter.BinaryOp{
				Operator: hinter.Mul,
				Lhs:      leftOp,
				Rhs:      rightOp,
			}, nil
		}
	}
	return nil, fmt.Errorf("unexpected product expression")
}

func (expression Expression) Evaluate() (hinter.Reference, error) {
	switch {
	case expression.IntExp != nil:
		intExp, err := expression.IntExp.Evaluate()
		if err != nil {
			return nil, err
		}
		return hinter.Immediate(*new(fp.Element).SetBigInt(intExp)), nil
	case expression.CellRefExp != nil:
		return expression.CellRefExp.Evaluate()
	case expression.DerefExp != nil:
		return expression.DerefExp.Evaluate()
	case expression.DDerefExp != nil:
		return expression.DDerefExp.Evaluate()
	default:
		return nil, fmt.Errorf("unexpected expression value")
	}
}

func (expression CellRefSimple) Evaluate() (hinter.CellRefer, error) {
	if expression.RegisterOffset != nil {
		return expression.RegisterOffset.Evaluate()
	}

	return EvaluateRegister(expression.Register, 0)
}

func (expression CellRefExp) Evaluate() (hinter.CellRefer, error) {
	if expression.RegisterOffset != nil {
		return expression.RegisterOffset.Evaluate()
	}

	return EvaluateRegister(expression.Register, 0)
}

func (expression RegisterOffset) Evaluate() (hinter.CellRefer, error) {
	offsetValue, _ := expression.Offset.Evaluate()
	offset, ok := utils.Int16FromBigInt(offsetValue)
	if !ok {
		return nil, fmt.Errorf("offset does not fit in int16")
	}
	if expression.Operator == "-" {
		offset = -offset
	}

	return EvaluateRegister(expression.Register, offset)
}

func EvaluateRegister(register string, offset int16) (hinter.CellRefer, error) {
	switch register {
	case "ap":
		return hinter.ApCellRef(offset), nil
	case "fp":
		return hinter.FpCellRef(offset), nil
	default:
		return nil, fmt.Errorf("invalid offset value")
	}
}

func (expression OffsetExp) Evaluate() (*big.Int, error) {
	switch {
	case expression.Number != "":
		bigIntValue, ok := new(big.Int).SetString(expression.Number, 10)
		if !ok {
			return nil, fmt.Errorf("expected a number")
		}
		return bigIntValue, nil
	case expression.NegNumber != "":
		bigIntValue, ok := new(big.Int).SetString(expression.NegNumber, 10)
		if !ok {
			return nil, fmt.Errorf("expected a number")
		}
		negNumber := bigIntValue.Neg(bigIntValue)
		return negNumber, nil
	default:
		return nil, fmt.Errorf("expected a number")
	}
}

func (expression DerefExp) Evaluate() (hinter.Deref, error) {
	cellRef, err := expression.CellRefExp.Evaluate()
	if err != nil {
		return hinter.Deref{}, err
	}
	return hinter.Deref{Deref: cellRef}, nil
}

func (expression DDerefExp) Evaluate() (hinter.DoubleDeref, error) {
	switch {
	case expression.DerefExp != nil:
		derefExp, err := expression.DerefExp.Evaluate()
		if err != nil {
			return hinter.DoubleDeref{}, err
		}
		return hinter.DoubleDeref{
			Deref:  derefExp,
			Offset: 0,
		}, nil
	case expression.DerefOffsetExp != nil:
		derefExp, err := expression.DerefOffsetExp.DerefExp.Evaluate()
		if err != nil {
			return hinter.DoubleDeref{}, err
		}
		offsetValue, err := expression.DerefOffsetExp.Offset.Evaluate()
		if err != nil {
			return hinter.DoubleDeref{}, err
		}
		offset, ok := utils.Int16FromBigInt(offsetValue)
		if !ok {
			return hinter.DoubleDeref{}, fmt.Errorf("offset does not fit in int16")
		}
		if expression.DerefOffsetExp.Operator == "-" {
			offset = -offset
		}
		return hinter.DoubleDeref{
			Deref:  derefExp,
			Offset: offset,
		}, nil

	default:
		return hinter.DoubleDeref{}, fmt.Errorf("unexpected double deref expression")
	}
}

func ParseIdentifier(value string) (hinter.Reference, error) {
	identifierExp, err := parser.ParseString("", value)
	if err != nil {
		return nil, err
	}

	return identifierExp.Evaluate()
}

func parseOperator(op string) (hinter.Operator, error) {
	switch op {
	case "+":
		return hinter.Add, nil
	case "*":
		return hinter.Mul, nil
	default:
		return 0, fmt.Errorf("unexpected op: %q", op)
	}
}
