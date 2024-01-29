package zero

import (
	"fmt"

	"github.com/alecthomas/participle/v2"
	op "github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
)

var parser *participle.Parser[IdentifierExp] = participle.MustBuild[IdentifierExp](participle.UseLookahead(10))

// Possible cases extracted from https://github.com/lambdaclass/cairo-vm_in_go/blob/main/pkg/hints/hint_utils/hint_reference.go#L41
// Immediate: cast(number, type)
// Reference no deref 1 offset: cast(reg + off, type)
// Reference no deref 2 offsets: cast(reg + off1 + off2, type)
// Reference with deref 1 offset: cast([reg + off1], type)
// Reference with deref 2 offsets: cast([reg + off1] + off2, type)
// Two references with deref: cast([reg + off1] + [reg + off2], type)
// Reference off omitted: cast(reg, type)
// Reference with deref off omitted: cast([reg], type)
// Reference with deref 2 offsets off1 omitted: cast([reg] + off2, type)
// 2 dereferences off1 omitted: cast([reg] + [reg + off2], type)
// 2 dereferences off2 omitted: cast([reg + off1] + [reg], type)
// 2 dereferences both offs omitted: cast([reg] + [reg], type)
// Reference no dereference 2 offsets - + : cast(reg - off1 + off2, type)

// Note: The same cases apply with an external dereference. Example: [cast(number, type)]

type IdentifierExp struct {
	DerefCastExp *DerefCastExp `@@ |`
	CastExp      *CastExp      `@@`
}

type DerefCastExp struct {
	CastExp *CastExp `"[" @@ "]"`
}

type CastExp struct {
	ValueExpr *Expression `"cast" "(" @@ ","`
	CastType  []string  `@Ident ("." @Ident)* ("*")? ("*")? ")"`
}

type Expression struct {
	BinOpExp   *BinOpExp   `@@ |`
	CellRefExp *CellRefExp `"(" @@ ")" | @@ |`
	DerefExp   *DerefExp   `@@`
}

type CellRefExp struct {
	RegisterOffset *RegisterOffset `@@ |`
	Register       string          `@("ap" | "fp")`
}

type RegisterOffset struct {
	Register string     `@("ap" | "fp")`
	Operator string     `@("+" | "-")`
	Offset   *OffsetExp `@@`
}

type DerefExp struct {
	CellRefExp *CellRefExp `"[" @@ "]"`
}

type BinOpExp struct {
	LeftExp  *LeftExp   `@@ "+"`
	RightExp *RightExp  `@@`
}

type OffsetExp struct {
	Number    *int `@Int |`
	NegNumber *int `"(" "-" @Int ")"`
}

type LeftExp struct {
	CellRefExp *RegisterOffset `@@ |`
	DerefExp   *DerefExp   `@@`
}

type RightExp struct {
	DerefExp *DerefExp  `@@ |`
	Offset   *OffsetExp `@@`
}

type DerefOffset struct {
	Deref  op.Deref
	Offset *int
}
type DerefDeref struct {
	LeftDeref  op.Deref
	RightDeref op.Deref
}


// AST Functionality
func (expression IdentifierExp) Evaluate() (any, error) {
	switch {
	case expression.DerefCastExp != nil:
		return expression.DerefCastExp.Evaluate()
	case expression.CastExp != nil:
		return expression.CastExp.Evaluate()
	default:
		return nil, fmt.Errorf("unexpected identifier value")
	}
}

func (expression DerefCastExp) Evaluate() (any, error) {
	value, err := expression.CastExp.ValueExpr.Evaluate()
	if err != nil {
		return nil, err
	}

	switch result := value.(type) {
	case op.CellRefer:
		return op.Deref{Deref: result}, nil
	case op.Deref:
		return op.DoubleDeref{
			Deref:  result.Deref,
			Offset: 0,
		},
		nil		
	case DerefOffset:
		return op.DoubleDeref{
			Deref:  result.Deref.Deref,
			Offset: int16(*result.Offset),
		},
		nil
	default:
		return nil, fmt.Errorf("unexpected identifier value")
	}
}

func (expression CastExp) Evaluate() (any, error) {
	value, err := expression.ValueExpr.Evaluate()
	if err != nil {
		return nil, err
	}

	switch result := value.(type) {
	case op.CellRefer:
		return result, nil
	case op.Deref:
		return result, nil
	case DerefOffset:
		return op.BinaryOp{
			Operator: 0,
			Lhs:      result.Deref.Deref,
			Rhs:      op.Immediate{
				uint64(0),
				uint64(0),
				uint64(0),
				uint64(*result.Offset),
			},
		}, nil
	case DerefDeref:
		return op.BinaryOp{
			Operator: 0,
			Lhs:      result.LeftDeref.Deref,
			Rhs:      result.RightDeref,
		}, nil
	default:
		return nil, fmt.Errorf("unexpected identifier value")
	}
}

func (expression Expression) Evaluate() (any, error) {
	switch {
	case expression.CellRefExp != nil:
		return expression.CellRefExp.Evaluate()
	case expression.DerefExp != nil:
		return expression.DerefExp.Evaluate()
	case expression.BinOpExp != nil:
		return expression.BinOpExp.Evaluate()
	default:
		return nil, fmt.Errorf("unexpected expression value")
	}
}

func (expression RegisterOffset) Evaluate() (any, error) {
	offsetValue, _ := expression.Offset.Evaluate()
	offset := int16(*offsetValue)
	if expression.Operator == "-" {
		offset = -offset
	}
	
	return EvaluateRegister(expression.Register, offset)
}

func (expression CellRefExp) Evaluate() (any, error) {
	if expression.RegisterOffset != nil {
		return expression.RegisterOffset.Evaluate()
	}
	
	return EvaluateRegister(expression.Register, 0)
}

func EvaluateRegister(register string, offset int16) (op.CellRefer, error) {
	switch register{
	case "ap":
		return op.ApCellRef(offset), nil
	case "fp":
		return op.FpCellRef(offset), nil
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
		return nil, fmt.Errorf("Expected a number")
	}
}

func (expression DerefExp) Evaluate() (any, error) {
	cellRefExp, err := expression.CellRefExp.Evaluate()
	if err != nil {
		return nil, err
	}
	cellRef, ok := cellRefExp.(op.CellRefer)
	if !ok {
		return nil, fmt.Errorf("Expected a CellRefer expression but got %s", cellRefExp)
	}
	return op.Deref{Deref: cellRef}, nil
}

func (expression BinOpExp) Evaluate() (any, error) {
	leftExp, err := expression.LeftExp.Evaluate()
	if err != nil {
		return nil, err
	}

	rightExp, err := expression.RightExp.Evaluate()
	if err != nil{
		return nil, err
	}

	switch lResult := leftExp.(type){
	case op.CellRefer:
		offset, ok := rightExp.(*int)
		if !ok {
			return nil, fmt.Errorf("invalid type operation")
		}
		offsetValue := int16(*offset)

		var cellRefOffset int16
		switch register := lResult.(type) {
		case op.ApCellRef:
			cellRefOffset = int16(register)
		case op.FpCellRef:
			cellRefOffset = int16(register)
		}

		offsetValue = offsetValue + cellRefOffset
		switch lResult.(type) {
		case op.ApCellRef:
			return op.ApCellRef(offsetValue), nil
		case op.FpCellRef:
			return op.FpCellRef(offsetValue), nil
		}
	
	case op.Deref:
		switch rResult := rightExp.(type) {
		case op.Deref:
			return DerefDeref{
				lResult,
				rResult,
			}, nil
		case *int:
			return DerefOffset{
				lResult,
				rResult,
			}, nil
		}
	}

	return nil, fmt.Errorf("invalid binary operation")
}

func (expression LeftExp) Evaluate() (any, error) {
	switch{
	case expression.CellRefExp != nil:
		return expression.CellRefExp.Evaluate()
	case expression.DerefExp != nil:
		return expression.DerefExp.Evaluate()
	}
	return nil, fmt.Errorf("Unexpected left expression in binary operation")
}

func (expression RightExp) Evaluate() (any, error) {
	switch{
	case expression.DerefExp != nil:
		return expression.DerefExp.Evaluate()
	case expression.Offset != nil:
		return expression.Offset.Evaluate()
	}
	return nil, fmt.Errorf("Unexpected right expression in binary operation")
}

func ParseIdentifier(value string) (any, error) {
	identifierExp, err := parser.ParseString("", value)
	if err != nil {
		return nil, err
	}

	return identifierExp.Evaluate()
}