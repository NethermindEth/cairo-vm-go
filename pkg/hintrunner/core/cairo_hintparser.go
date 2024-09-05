package core

import (
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/parsers/starknet"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func parseCellRefer(cr starknet.CellRef) hinter.Reference {
	switch cr.Register {
	case starknet.AP:
		return hinter.ApCellRef(cr.Offset)
	case starknet.FP:
		return hinter.FpCellRef(cr.Offset)
	}
	return nil
}

func parseDeref(ro starknet.ResOperand) hinter.Deref {
	cr := parseCellRefer(ro.ResOperand.(*starknet.Deref).Deref)
	return hinter.Deref{
		Deref: cr,
	}
}

func parseDoubleDeref(ro starknet.ResOperand) hinter.DoubleDeref {
	dd := ro.ResOperand.(*starknet.DoubleDeref)
	offset := int16(dd.Inner.Offset)
	cr := parseCellRefer(dd.Inner.CellRef)
	deref := hinter.Deref{
		Deref: cr,
	}
	return hinter.DoubleDeref{
		Offset: offset,
		Deref:  deref,
	}
}

func parseImmediate(ro starknet.ResOperand) hinter.Immediate {
	val := ro.ResOperand.(*starknet.Immediate).Immediate
	valFelt := new(fp.Element).SetBigInt(val)
	return hinter.Immediate(*valFelt)
}

func parseBinOp(ro starknet.ResOperand) hinter.BinaryOp {
	binOp := ro.ResOperand.(*starknet.BinOp).BinOp
	a := hinter.Deref{
		Deref: parseCellRefer(binOp.A),
	}
	var b hinter.Reference = nil
	switch binOp.B.Inner.(type) {
	case *starknet.Deref:
		b = &hinter.Deref{
			Deref: parseCellRefer(binOp.B.Inner.(*starknet.Deref).Deref),
		}
	case *starknet.Immediate:
		val := binOp.B.Inner.(*starknet.Immediate).Immediate
		valFelt := new(fp.Element).SetBigInt(val)
		b = hinter.Immediate(*valFelt)
	}
	var operation hinter.Operator = 0
	switch binOp.Op {
	case starknet.Add:
		operation = hinter.Add
	case starknet.Mul:
		operation = hinter.Add
	}
	return hinter.BinaryOp{
		Operator: operation,
		Lhs:      a,
		Rhs:      b,
	}
}

func parseResOperand(ro starknet.ResOperand) hinter.Reference {
	switch ro.Name {
	case starknet.DerefName:
		return parseDeref(ro)
	case starknet.DoubleDerefName:
		return parseDoubleDeref(ro)
	case starknet.ImmediateName:
		return parseImmediate(ro)
	case starknet.BinOpName:
		return parseBinOp(ro)
	}
	return nil
}

func GetHintByName(hint starknet.Hint) (hinter.Hinter, error) {
	switch hint.Name {
	case starknet.AllocSegmentName:
		args := hint.Args.(*starknet.AllocSegment)
		return &AllocSegment{
			Dst: parseCellRefer(args.Dst),
		}, nil
	case starknet.TestLessThanName:
		args := hint.Args.(*starknet.TestLessThan)
		return &TestLessThan{
			lhs: parseResOperand(args.Lhs),
			rhs: parseResOperand(args.Rhs),
			dst: parseCellRefer(args.Dst),
		}, nil
	case starknet.TestLessThanOrEqualName:
		args := hint.Args.(*starknet.TestLessThanOrEqual)
		return &TestLessThanOrEqual{
			lhs: parseResOperand(args.Lhs),
			rhs: parseResOperand(args.Rhs),
			dst: parseCellRefer(args.Dst),
		}, nil
	case starknet.LinearSplitName:
		args := hint.Args.(*starknet.LinearSplit)
		return &LinearSplit{
			value:  parseResOperand(args.Value),
			scalar: parseResOperand(args.Scalar),
			maxX:   parseResOperand(args.MaxX),
			x:      parseCellRefer(args.X),
			y:      parseCellRefer(args.Y),
		}, nil
	case starknet.WideMul128Name:
		args := hint.Args.(*starknet.WideMul128)
		return &WideMul128{
			lhs:  parseResOperand(args.Lhs),
			rhs:  parseResOperand(args.Rhs),
			high: parseCellRefer(args.High),
			low:  parseCellRefer(args.Low),
		}, nil
	case starknet.DivModName:
		args := hint.Args.(*starknet.DivMod)
		return &DivMod{
			lhs:       parseResOperand(args.Lhs),
			rhs:       parseResOperand(args.Rhs),
			quotient:  parseCellRefer(args.Quotient),
			remainder: parseCellRefer(args.Remainder),
		}, nil
	case starknet.Uint256DivModName:
		args := hint.Args.(*starknet.Uint256DivMod)
		return &Uint256DivMod{
			dividend0:  parseResOperand(args.Dividend0),
			dividend1:  parseResOperand(args.Dividend1),
			divisor0:   parseResOperand(args.Divisor0),
			divisor1:   parseResOperand(args.Divisor1),
			quotient0:  parseCellRefer(args.Quotient0),
			quotient1:  parseCellRefer(args.Quotient1),
			remainder0: parseCellRefer(args.Remainder0),
			remainder1: parseCellRefer(args.Remainder1),
		}, nil
	case starknet.DebugPrintName:
		args := hint.Args.(*starknet.DebugPrint)
		return &DebugPrint{
			start: parseResOperand(args.Start),
			end:   parseResOperand(args.End),
		}, nil
	case starknet.SquareRootName:
		args := hint.Args.(*starknet.SquareRoot)
		return &SquareRoot{
			value: parseResOperand(args.Value),
			dst:   parseCellRefer(args.Dst),
		}, nil
	case starknet.Uint256SquareRootName:
		args := hint.Args.(*starknet.Uint256SquareRoot)
		return &Uint256SquareRoot{
			valueLow:                     parseResOperand(args.ValueLow),
			valueHigh:                    parseResOperand(args.ValueHigh),
			sqrt0:                        parseCellRefer(args.Sqrt0),
			sqrt1:                        parseCellRefer(args.Sqrt1),
			remainderLow:                 parseCellRefer(args.RemainderLow),
			remainderHigh:                parseCellRefer(args.RemainderHigh),
			sqrtMul2MinusRemainderGeU128: parseCellRefer(args.SqrtMul2MinusRemainderGeU128),
		}, nil
	case starknet.AllocFelt252DictName:
		args := hint.Args.(*starknet.AllocFelt252Dict)
		return &AllocFelt252Dict{
			SegmentArenaPtr: parseResOperand(args.SegmentArenaPtr),
		}, nil
	case starknet.Felt252DictEntryInitName:
		args := hint.Args.(*starknet.Felt252DictEntryInit)
		return &Felt252DictEntryInit{
			DictPtr: parseResOperand(args.DictPtr),
			Key:     parseResOperand(args.Key),
		}, nil
	case starknet.Felt252DictEntryUpdateName:
		args := hint.Args.(*starknet.Felt252DictEntryUpdate)
		return &Felt252DictEntryUpdate{
			DictPtr: parseResOperand(args.DictPtr),
			Value:   parseResOperand(args.Value),
		}, nil
	case starknet.GetSegmentArenaIndexName:
		args := hint.Args.(*starknet.GetSegmentArenaIndex)
		return &GetSegmentArenaIndex{
			DictEndPtr: parseResOperand(args.DictEndPtr),
			DictIndex:  parseCellRefer(args.DictIndex),
		}, nil
	case starknet.InitSquashDataName:
		args := hint.Args.(*starknet.InitSquashData)
		return &InitSquashData{
			DictAccesses: parseResOperand(args.DictAccesses),
			NumAccesses:  parseResOperand(args.NAccesses),
			BigKeys:      parseCellRefer(args.BigKeys),
			FirstKey:     parseCellRefer(args.FirstKey),
		}, nil
	case starknet.GetCurrentAccessIndexName:
		args := hint.Args.(*starknet.GetCurrentAccessIndex)
		return &GetCurrentAccessIndex{
			RangeCheckPtr: parseResOperand(args.RangeCheckPtr),
		}, nil
	case starknet.ShouldSkipSquashLoopName:
		args := hint.Args.(*starknet.ShouldSkipSquashLoop)
		return &ShouldSkipSquashLoop{
			ShouldSkipLoop: parseCellRefer(args.ShouldSkipLoop),
		}, nil
	case starknet.GetCurrentAccessDeltaName:
		args := hint.Args.(*starknet.GetCurrentAccessDelta)
		return &GetCurrentAccessDelta{
			IndexDeltaMinusOne: parseCellRefer(args.IndexDeltaMinus1),
		}, nil
	case starknet.ShouldContinueSquashLoopName:
		args := hint.Args.(*starknet.ShouldContinueSquashLoop)
		return &ShouldContinueSquashLoop{
			ShouldContinue: parseCellRefer(args.ShouldContinue),
		}, nil
	case starknet.GetNextDictKeyName:
		args := hint.Args.(*starknet.GetNextDictKey)
		return &GetNextDictKey{
			NextKey: parseCellRefer(args.NextKey),
		}, nil
	case starknet.Uint512DivModByUint256Name:
		args := hint.Args.(*starknet.Uint512DivModByUint256)
		return &Uint512DivModByUint256{
			dividend0:  parseResOperand(args.Dividend0),
			dividend1:  parseResOperand(args.Dividend1),
			dividend2:  parseResOperand(args.Dividend2),
			dividend3:  parseResOperand(args.Dividend3),
			divisor0:   parseResOperand(args.Divisor0),
			divisor1:   parseResOperand(args.Divisor1),
			quotient0:  parseCellRefer(args.Quotient0),
			quotient1:  parseCellRefer(args.Quotient1),
			quotient2:  parseCellRefer(args.Quotient2),
			quotient3:  parseCellRefer(args.Quotient3),
			remainder0: parseCellRefer(args.Remainder0),
			remainder1: parseCellRefer(args.Remainder1),
		}, nil
	case starknet.AllocConstantSizeName:
		args := hint.Args.(*starknet.AllocConstantSize)
		return &AllocConstantSize{
			Dst:  parseCellRefer(args.Dst),
			Size: parseResOperand(args.Size),
		}, nil
	case starknet.AssertLeFindSmallArcsName:
		args := hint.Args.(*starknet.AssertLeFindSmallArcs)
		return &AssertLeFindSmallArc{
			A:             parseResOperand(args.A),
			B:             parseResOperand(args.B),
			RangeCheckPtr: parseResOperand(args.RangeCheckPtr),
		}, nil
	case starknet.AssertLeIsFirstArcExcludedName:
		args := hint.Args.(*starknet.AssertLeIsFirstArcExcluded)
		return &AssertLeIsFirstArcExcluded{
			SkipExcludeAFlag: parseCellRefer(args.SkipExcludeAFlag),
		}, nil
	case starknet.AssertLeIsSecondArcExcludedName:
		args := hint.Args.(*starknet.AssertLeIsSecondArcExcluded)
		return &AssertLeIsSecondArcExcluded{
			SkipExcludeBMinusA: parseCellRefer(args.SkipExcludeBMinusA),
		}, nil
	case starknet.RandomEcPointName:
		args := hint.Args.(*starknet.RandomEcPoint)
		return &RandomEcPoint{
			x: parseCellRefer(args.X),
			y: parseCellRefer(args.Y),
		}, nil
	case starknet.FieldSqrtName:
		args := hint.Args.(*starknet.FieldSqrt)
		return &FieldSqrt{
			val:  parseResOperand(args.Val),
			sqrt: parseCellRefer(args.Sqrt),
		}, nil
	default:
		return nil, fmt.Errorf("unknown hint: %v", hint.Name)
	}
}
