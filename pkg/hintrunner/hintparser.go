package hintrunner

import (
	"fmt"
)

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
	Deref  Deref
	Offset *int
}
type DerefDeref struct {
	LeftDeref  Deref
	RightDeref Deref
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
	case CellRefer:
		return Deref{result}, nil
	case Deref:
		return DoubleDeref{
			result.deref,
			0,
		},
		nil		
	case DerefOffset:
		return DoubleDeref{
			result.Deref.deref,
			int16(*result.Offset),
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
	case CellRefer:
		return result, nil
	case Deref:
		return result, nil
	case DerefOffset:
		return BinaryOp{
			0,
			result.Deref.deref,
			Immediate{
				uint64(0),
				uint64(0),
				uint64(0),
				uint64(*result.Offset),
			},
		}, nil
	case DerefDeref:
		return BinaryOp{
			0,
			result.LeftDeref.deref,
			result.RightDeref,
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

func EvaluateRegister(register string, offset int16) (CellRefer, error) {
	switch register{
	case "ap":
		return ApCellRef(offset), nil
	case "fp":
		return FpCellRef(offset), nil
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
	cellRef, ok := cellRefExp.(CellRefer)
	if !ok {
		return nil, fmt.Errorf("Expected a CellRefer expression but got %s", cellRefExp)
	}
	return Deref{cellRef}, nil
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
	case CellRefer:
		offset, ok := rightExp.(*int)
		if !ok {
			return nil, fmt.Errorf("invalid type operation")
		}
		offsetValue := int16(*offset)

		var cellRefOffset int16
		switch register := lResult.(type) {
		case ApCellRef:
			cellRefOffset = int16(register)
		case FpCellRef:
			cellRefOffset = int16(register)
		}

		offsetValue = offsetValue + cellRefOffset
		switch lResult.(type) {
		case ApCellRef:
			return ApCellRef(offsetValue), nil
		case FpCellRef:
			return FpCellRef(offsetValue), nil
		}
	
	case Deref:
		switch rResult := rightExp.(type) {
		case Deref:
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
