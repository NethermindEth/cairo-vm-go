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
	ValueExpr *Expression `"cast(" @@`
	CastType  string      `"," @String ")"`
}

type Expression struct {
	CellRefExp *CellRefExp `"("@@")" | @@ |`
	DerefExp   *DerefExp   `@@ |`
	BinOpExp   *BinOpExp   `@@`
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
	LeftExp  *LeftExp   `@@`
	RightExp *RightExp  `"+" @@`
}

type OffsetExp struct {
	Number    *int `@Int`
	NegNumber *int `"(-" @Int ")"`
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
func (expression IdentifierExp) Evaluate() (interface{}, error) {
	switch {
	case expression.DerefCastExp != nil:
		return expression.DerefCastExp.Evaluate()
	case expression.CastExp != nil:
		return expression.CastExp.Evaluate()
	default:
		return nil, fmt.Errorf("unexpected identifier value")
	}
}

func (expression DerefCastExp) Evaluate() (interface{}, error) {
	result, err := expression.CastExp.ValueExpr.Evaluate()
	if err != nil {
		return nil, err
	}

	switch result.(type) {
	case CellRefer:
		cellRefValue := result.(CellRefer)
		return Deref{cellRefValue}, nil
	case Deref:
		derefValue := result.(Deref)
		return DoubleDeref{
			derefValue.deref,
			0,
		},
		nil		
	case DerefOffset:
		derefOffset := result.(DerefOffset)
		return DoubleDeref{
			derefOffset.Deref.deref,
			int16(*derefOffset.Offset),
		},
		nil
	default:
		return nil, fmt.Errorf("unexpected identifier value")
	}
}

func (expression CastExp) Evaluate() (interface{}, error) {
	result, err := expression.ValueExpr.Evaluate()
	if err != nil {
		return nil, err
	}

	switch result.(type) {
	case CellRefer:
		return result, nil
	case Deref:
		return result, nil
	case DerefOffset:
		derefOffset := result.(DerefOffset)
		return BinaryOp{
			0,
			derefOffset.Deref.deref,
			Immediate{
				uint64(0),
				uint64(0),
				uint64(0),
				uint64(*derefOffset.Offset),
			},
		}, nil
	case DerefDeref:
		derefDeref := result.(DerefDeref)
		return BinaryOp{
			0,
			derefDeref.LeftDeref.deref,
			derefDeref.RightDeref,
		}, nil
	default:
		return nil, fmt.Errorf("unexpected identifier value")
	}
}

func (expression Expression) Evaluate() (interface{}, error) {
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

func (expression RegisterOffset) Evaluate() (interface{}, error) {
	offsetValue, _ := expression.Offset.Evaluate()
	offset := int16(*offsetValue)
	if expression.Operator == "-" {
		offset = -offset
	}
	
	return EvaluateRegister(expression.Register, offset)
}

func (expression CellRefExp) Evaluate() (interface{}, error) {
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
	result := *expression.Number - *expression.NegNumber
	return &result, nil
}

func (expression DerefExp) Evaluate() (interface{}, error) {
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

func (expression BinOpExp) Evaluate() (interface{}, error) {
	leftExp, err := expression.LeftExp.Evaluate()
	if err != nil {
		return nil, err
	}

	rightExp, err := expression.RightExp.Evaluate()
	if err != nil{
		return nil, err
	}

	switch leftExp.(type){
	case CellRefer:
		cellRef := leftExp.(CellRefer)
		offset, ok := rightExp.(*int)
		if !ok {
			return nil, fmt.Errorf("invalid type operation")
		}
		offsetValue := int16(*offset)

		var cellRefOffset int16
		switch cellRef.(type) {
		case ApCellRef:
			apCellRef := cellRef.(ApCellRef)
			cellRefOffset = int16(apCellRef)
		case FpCellRef:
			fpCellRef := cellRef.(FpCellRef)
			cellRefOffset = int16(fpCellRef)
		}

		offsetValue = offsetValue + cellRefOffset
		switch cellRef.(type) {
		case ApCellRef:
			return ApCellRef(offsetValue), nil
		case FpCellRef:
			return FpCellRef(offsetValue), nil
		}
	
	case Deref:
		deref := leftExp.(Deref)
		switch rightExp.(type) {
		case Deref:
			rDeref := rightExp.(Deref)
			return DerefDeref{
				deref,
				rDeref,
			}, nil
		case *int:
			offset := rightExp.(*int)
			return DerefOffset{
				deref,
				offset,
			}, nil
		}
	}

	return nil, fmt.Errorf("invalid binary operation")
}

func (expression LeftExp) Evaluate() (interface{}, error) {
	switch{
	case expression.CellRefExp != nil:
		return expression.CellRefExp.Evaluate()
	case expression.DerefExp != nil:
		return expression.DerefExp.Evaluate()
	}
	return nil, fmt.Errorf("Unexpected left expression in binary operation")
}

func (expression RightExp) Evaluate() (interface{}, error) {
	switch{
	case expression.DerefExp != nil:
		return expression.DerefExp.Evaluate()
	case expression.Offset != nil:
		return expression.Offset.Evaluate()
	}
	return nil, fmt.Errorf("Unexpected right expression in binary operation")
}
