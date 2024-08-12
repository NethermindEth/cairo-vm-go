package zero

import (
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"github.com/alecthomas/participle/v2"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

// Grammar for hint references:
//
// [cast(expr, type)]
// arithExp => term | addTerm
// addTerm => term (+|-) arithExp
// term => Exp | ProdExp
// prodExp => Exp * Exp
// exp => cellRef | deref | dderef | int

var parser *participle.Parser[IdentifierExp] = participle.MustBuild[IdentifierExp](participle.UseLookahead(20))

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
	AddExp  *AddExp  `@@ |`
	TermExp *TermExp `@@`
}

type AddExp struct {
	LeftExp  *TermExp  `@@`
	Operator string    `@("+" | "-")`
	RightExp *ArithExp `@@`
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
	IntExp     *OffsetExp  `@@ |`
	CellRefExp *CellRefExp `@@ |`
	DerefExp   *DerefExp   `@@ |`
	DDerefExp  *DDerefExp  `@@`
}

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
	Number    *int `@Int |`
	NegNumber *int `"(" "-" @Int ")"`
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
	DerefExp       *DerefExp       `"[" @@ "]" |`
	DerefOffsetExp *DerefOffsetExp `"[" @@ "]"`
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
				offset, err := utils.Int16FromFelt((*fp.Element)(&right))
				if err == nil {
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
	switch {
	case expression.AddExp != nil:
		return expression.AddExp.Evaluate()
	case expression.TermExp != nil:
		return expression.TermExp.Evaluate()
	default:
		return nil, fmt.Errorf("unexpected identifier value")
	}
}

func (expression AddExp) Evaluate() (hinter.Reference, error) {
	leftExp, err := expression.LeftExp.Evaluate()
	if err != nil {
		return nil, err
	}
	rightExp, err := expression.RightExp.Evaluate()
	if err != nil {
		return nil, err
	}

	if rightResult, ok := rightExp.(hinter.Immediate); ok {
		switch leftResult := leftExp.(type) {
		case hinter.CellRefer:
			off, err := utils.Int16FromFelt((*fp.Element)(&rightResult))
			if err == nil {
				if expression.Operator == "-" {
					off = -off
				}

				switch cellRef := leftResult.(type) {
				case hinter.ApCellRef:
					oldOffset := int16(cellRef)
					return hinter.ApCellRef(off + oldOffset), nil
				case hinter.FpCellRef:
					oldOffset := int16(cellRef)
					return hinter.FpCellRef(off + oldOffset), nil
				}
			}
		case hinter.Immediate:
			lFelt := (*fp.Element)(&leftResult)
			rFelt := (*fp.Element)(&rightResult)
			result := new(fp.Element).Add(lFelt, rFelt)
			return hinter.Immediate(*result), nil
		}
	}

	operator, err := parseOperator(expression.Operator)
	if err != nil {
		return nil, err
	}

	// This is necesary since leftExp and rightExp are References and BinaryOp requires ResOperanders
	if leftOp, ok := leftExp.(hinter.ResOperander); ok {
		if rightOp, ok := rightExp.(hinter.ResOperander); ok {
			return hinter.BinaryOp{
				Operator: operator,
				Lhs:      leftOp,
				Rhs:      rightOp,
			}, nil
		}
	}
	return nil, fmt.Errorf("unexpected addition expression")
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
		return hinter.Immediate(*new(fp.Element).SetInt64(int64(*intExp))), nil
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
	offset := int16(*offsetValue)
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

func (expression OffsetExp) Evaluate() (*int, error) {
	switch {
	case expression.Number != nil:
		return expression.Number, nil
	case expression.NegNumber != nil:
		negNumber := -*expression.NegNumber
		return &negNumber, nil
	default:
		return nil, fmt.Errorf("expected a number")
	}
}

func (expression DerefExp) Evaluate() (hinter.Deref, error) {
	cellRefExp, err := expression.CellRefExp.Evaluate()
	if err != nil {
		return hinter.Deref{}, err
	}
	cellRef, ok := cellRefExp.(hinter.CellRefer)
	if !ok {
		return hinter.Deref{}, fmt.Errorf("expected a CellRefer expression but got %s", cellRefExp)
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
		offset := int16(*offsetValue)
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
