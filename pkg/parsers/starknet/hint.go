package starknet

import (
	"encoding/json"
	"fmt"
	"math/big"
)

// Source of the code:
// https://github.com/starkware-libs/cairo/blob/main/crates/cairo-lang-casm/src/hints/mod.rs
// https://github.com/starkware-libs/cairo/blob/main/crates/cairo-lang-casm/src/operand.rs

type HintName string

const (
	// Starknet hints
	SystemCallName HintName = "SystemCall"
	CheatcodeName  HintName = "Cheatcode"
	// Core hints
	AllocSegmentName                HintName = "AllocSegment"
	TestLessThanName                HintName = "TestLessThan"
	TestLessThanOrEqualName         HintName = "TestLessThanOrEqual"
	WideMul128Name                  HintName = "WideMul128"
	DivModName                      HintName = "DivMod"
	Uint256DivModName               HintName = "Uint256DivMod"
	Uint512DivModByUint256Name      HintName = "Uint512DivModByUint256"
	SquareRootName                  HintName = "SquareRoot"
	Uint256SquareRootName           HintName = "Uint256SquareRoot"
	LinearSplitName                 HintName = "LinearSplit"
	AllocFelt252DictName            HintName = "AllocFelt252Dict"
	Felt252DictEntryInitName        HintName = "Felt252DictEntryInit"
	Felt252DictEntryUpdateName      HintName = "Felt252DictEntryUpdate"
	GetSegmentArenaIndexName        HintName = "GetSegmentArenaIndex"
	InitSquashDataName              HintName = "InitSquashData"
	GetCurrentAccessIndexName       HintName = "GetCurrentAccessIndex"
	ShouldSkipSquashLoopName        HintName = "ShouldSkipSquashLoop"
	GetCurrentAccessDeltaName       HintName = "GetCurrentAccessDelta"
	ShouldContinueSquashLoopName    HintName = "ShouldContinueSquashLoop"
	GetNextDictKeyName              HintName = "GetNextDictKey"
	AssertLeFindSmallArcsName       HintName = "AssertLeFindSmallArcs"
	AssertLeIsFirstArcExcludedName  HintName = "AssertLeIsFirstArcExcluded"
	AssertLeIsSecondArcExcludedName HintName = "AssertLeIsSecondArcExcluded"
	RandomEcPointName               HintName = "RandomEcPoint"
	FieldSqrtName                   HintName = "FieldSqrt"
	DebugPrintName                  HintName = "DebugPrint"
	AllocConstantSizeName           HintName = "AllocConstantSize"
	// Name Deprecated hints
	AssertCurrentAccessIndicesIsEmptyName HintName = "AssertCurrentAccessIndicesIsEmpty"
	AssertAllAccessesUsedName             HintName = "AssertAllAccessesUsed"
	AssertAllKeysUsedName                 HintName = "AssertAllKeysUsed"
	AssertLeAssertThirdArcExcludedName    HintName = "AssertLeAssertThirdArcExcluded"
	AssertLtAssertValidInputName          HintName = "AssertLtAssertValidInput"
	Felt252DictReadName                   HintName = "Felt252DictRead"
	Felt252DictWriteName                  HintName = "Felt252DictWrite"
)

// Starknet hints
type SystemCall struct {
	System ResOperand `json:"system" validate:"required"`
}

type Cheatcode struct {
	Selector    *big.Int   `json:"selector" validate:"required"`
	InputStart  ResOperand `json:"input_start" validate:"required"`
	InputEnd    ResOperand `json:"input_end" validate:"required"`
	OutputStart CellRef    `json:"output_start" validate:"required"`
	OutputEnd   CellRef    `json:"output_end" validate:"required"`
}

// Core hints

type AllocSegment struct {
	Dst CellRef `json:"dst" validate:"required"`
}

type TestLessThan struct {
	Lhs ResOperand `json:"lhs" validate:"required"`
	Rhs ResOperand `json:"rhs" validate:"required"`
	Dst CellRef    `json:"dst" validate:"required"`
}

type TestLessThanOrEqual struct {
	Lhs ResOperand `json:"lhs" validate:"required"`
	Rhs ResOperand `json:"rhs" validate:"required"`
	Dst CellRef    `json:"dst" validate:"required"`
}

type WideMul128 struct {
	Lhs  ResOperand `json:"lhs" validate:"required"`
	Rhs  ResOperand `json:"rhs" validate:"required"`
	High CellRef    `json:"high" validate:"required"`
	Low  CellRef    `json:"low" validate:"required"`
}

type DivMod struct {
	Lhs       ResOperand `json:"lhs" validate:"required"`
	Rhs       ResOperand `json:"rhs" validate:"required"`
	Quotient  CellRef    `json:"quotient" validate:"required"`
	Remainder CellRef    `json:"remainder" validate:"required"`
}

type Uint256DivMod struct {
	Dividend0  ResOperand `json:"dividend0" validate:"required"`
	Dividend1  ResOperand `json:"dividend1" validate:"required"`
	Divisor0   ResOperand `json:"divisor0" validate:"required"`
	Divisor1   ResOperand `json:"divisor1" validate:"required"`
	Quotient0  CellRef    `json:"quotient0" validate:"required"`
	Quotient1  CellRef    `json:"quotient1" validate:"required"`
	Remainder0 CellRef    `json:"remainder0" validate:"required"`
	Remainder1 CellRef    `json:"remainder1" validate:"required"`
}

type Uint512DivModByUint256 struct {
	Dividend0  ResOperand `json:"dividend0" validate:"required"`
	Dividend1  ResOperand `json:"dividend1" validate:"required"`
	Dividend2  ResOperand `json:"dividend2" validate:"required"`
	Dividend3  ResOperand `json:"dividend3" validate:"required"`
	Divisor0   ResOperand `json:"divisor0" validate:"required"`
	Divisor1   ResOperand `json:"divisor1" validate:"required"`
	Quotient0  CellRef    `json:"quotient0" validate:"required"`
	Quotient1  CellRef    `json:"quotient1" validate:"required"`
	Quotient2  CellRef    `json:"quotient2" validate:"required"`
	Quotient3  CellRef    `json:"quotient3" validate:"required"`
	Remainder0 CellRef    `json:"remainder0" validate:"required"`
	Remainder1 CellRef    `json:"remainder1" validate:"required"`
}

type SquareRoot struct {
	Value ResOperand `json:"value" validate:"required"`
	Dst   CellRef    `json:"dst" validate:"required"`
}

type Uint256SquareRoot struct {
	ValueLow                     ResOperand `json:"value_low" validate:"required"`
	ValueHigh                    ResOperand `json:"value_high" validate:"required"`
	Sqrt0                        CellRef    `json:"sqrt0" validate:"required"`
	Sqrt1                        CellRef    `json:"sqrt1" validate:"required"`
	RemainderLow                 CellRef    `json:"remainder_low" validate:"required"`
	RemainderHigh                CellRef    `json:"remainder_high" validate:"required"`
	SqrtMul2MinusRemainderGeU128 CellRef    `json:"sqrt_mul_2_minus_remainder_ge_u128" validate:"required"`
}

type LinearSplit struct {
	Value  ResOperand `json:"value" validate:"required"`
	Scalar ResOperand `json:"scalar" validate:"required"`
	MaxX   ResOperand `json:"max_x" validate:"required"`
	X      CellRef    `json:"x" validate:"required"`
	Y      CellRef    `json:"y" validate:"required"`
}

type AllocFelt252Dict struct {
	SegmentArenaPtr ResOperand `json:"segment_arena_ptr" validate:"required"`
}

type Felt252DictEntryInit struct {
	DictPtr ResOperand `json:"dict_ptr" validate:"required"`
	Key     ResOperand `json:"key" validate:"required"`
}

type Felt252DictEntryUpdate struct {
	DictPtr ResOperand `json:"dict_ptr" validate:"required"`
	Value   ResOperand `json:"value" validate:"required"`
}

type GetSegmentArenaIndex struct {
	DictEndPtr ResOperand `json:"dict_end_ptr" validate:"required"`
	DictIndex  CellRef    `json:"dict_index" validate:"required"`
}

type InitSquashData struct {
	DictAccesses ResOperand `json:"dict_accesses" validate:"required"`
	PtrDiff      ResOperand `json:"ptr_diff" validate:"required"`
	NAccesses    ResOperand `json:"n_accesses" validate:"required"`
	BigKeys      CellRef    `json:"big_keys" validate:"required"`
	FirstKey     CellRef    `json:"first_key" validate:"required"`
}

type GetCurrentAccessIndex struct {
	RangeCheckPtr ResOperand `json:"range_check_ptr" validate:"required"`
}

type ShouldSkipSquashLoop struct {
	ShouldSkipLoop CellRef `json:"should_skip_loop" validate:"required"`
}

type GetCurrentAccessDelta struct {
	IndexDeltaMinus1 CellRef `json:"index_delta_minus_1" validate:"required"`
}

type ShouldContinueSquashLoop struct {
	ShouldContinue CellRef `json:"should_continue" validate:"required"`
}

type GetNextDictKey struct {
	NextKey CellRef `json:"next_key" validate:"required"`
}

type AssertLeFindSmallArcs struct {
	RangeCheckPtr ResOperand `json:"range_check_ptr" validate:"required"`
	A             ResOperand `json:"a" validate:"required"`
	B             ResOperand `json:"b" validate:"required"`
}

type AssertLeIsFirstArcExcluded struct {
	SkipExcludeAFlag CellRef `json:"skip_exclude_a_flag" validate:"required"`
}

type AssertLeIsSecondArcExcluded struct {
	SkipExcludeBMinusA CellRef `json:"skip_exclude_b_minus_a" validate:"required"`
}

type RandomEcPoint struct {
	X CellRef `json:"x" validate:"required"`
	Y CellRef `json:"y" validate:"required"`
}

type FieldSqrt struct {
	Val  ResOperand `json:"val" validate:"required"`
	Sqrt CellRef    `json:"sqrt" validate:"required"`
}

type DebugPrint struct {
	Start ResOperand `json:"start" validate:"required"`
	End   ResOperand `json:"end" validate:"required"`
}

type AllocConstantSize struct {
	Size ResOperand `json:"size" validate:"required"`
	Dst  CellRef    `json:"dst" validate:"required"`
}

// Deprecated hints

type AssertCurrentAccessIndicesIsEmpty struct{}

type AssertAllAccessesUsed struct {
	NUsedAccesses CellRef `json:"n_used_accesses" validate:"required"`
}

type AssertAllKeysUsed struct{}

type AssertLeAssertThirdArcExcluded struct{}

type AssertLtAssertValidInput struct {
	A ResOperand `json:"a" validate:"required"`
	B ResOperand `json:"b" validate:"required"`
}

type Felt252DictRead struct {
	DictPtr  ResOperand `json:"dict_ptr" validate:"required"`
	Key      ResOperand `json:"key" validate:"required"`
	ValueDst CellRef    `json:"value_dst" validate:"required"`
}

type Felt252DictWrite struct {
	DictPtr ResOperand `json:"dict_ptr" validate:"required"`
	Key     ResOperand `json:"key" validate:"required"`
	Value   ResOperand `json:"value" validate:"required"`
}

// Operands

type Register string

const (
	AP Register = "AP"
	FP Register = "FP"
)

type CellRef struct {
	Register Register `json:"register" validate:"required"`
	Offset   int      `json:"offset" validate:"required"`
}

type ResOperandName string

const (
	DerefName       ResOperandName = "Deref"
	DoubleDerefName ResOperandName = "DoubleDeref"
	ImmediateName   ResOperandName = "Immediate"
	BinOpName       ResOperandName = "BinOp"
)

type Operand interface{}

type ResOperand struct {
	Name       ResOperandName
	ResOperand Operand `validate:"required"`
}

func (ro *ResOperand) UnmarshalJSON(data []byte) error {
	var resOp map[string]json.RawMessage

	err := json.Unmarshal(data, &resOp)
	if err != nil {
		return err
	}

	var op any
	var name ResOperandName
	for k := range resOp {
		switch ResOperandName(k) {
		case DerefName:
			op = &Deref{}
			name = DerefName
		case DoubleDerefName:
			op = &DoubleDeref{}
			name = DoubleDerefName
		case ImmediateName:
			op = &Immediate{}
			name = ImmediateName
		case BinOpName:
			op = &BinOp{}
			name = BinOpName
		default:
			return fmt.Errorf("unknown res operand %s", k)
		}

		if err = json.Unmarshal(data, op); err != nil {
			return err
		}
		break
	}

	ro.ResOperand = op
	ro.Name = name
	return nil
}

type Deref struct {
	Deref CellRef `validate:"required"`
}

type InnerDoubleDeref struct {
	CellRef CellRef `validate:"required"`
	Offset  int     `validate:"required"`
}

func (i *InnerDoubleDeref) UnmarshalJSON(data []byte) error {
	var s []any
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	byteCell, err := json.Marshal(s[0])
	if err != nil {
		return err
	}
	var cell CellRef
	err = json.Unmarshal(byteCell, &cell)
	if err != nil {
		return err
	}
	i.CellRef = cell

	offset, ok := s[1].(float64)
	if !ok {
		return fmt.Errorf("convert offset %v to float64", s[1])
	}
	i.Offset = int(offset)

	return nil
}

type DoubleDeref struct {
	Inner InnerDoubleDeref `json:"DoubleDeref" validate:"required"`
}

type Immediate struct {
	Immediate *big.Int `validate:"required"`
}

func (i *Immediate) UnmarshalJSON(data []byte) error {
	str := struct {
		Immediate string
	}{}
	err := json.Unmarshal(data, &str)
	if err != nil {
		return err
	}

	immediate, ok := new(big.Int).SetString(str.Immediate, 0)
	if !ok {
		return fmt.Errorf("convert immediate value %s to big.Int", str.Immediate)
	}

	i.Immediate = immediate
	return nil
}

type BinOp struct {
	BinOp BinOpOperand `validate:"required"`
}

type Operation string

const (
	Add Operation = "Add"
	Mul Operation = "Mul"
)

type BinOpOperand struct {
	Op Operation        `json:"op" validate:"required"`
	A  CellRef          `json:"a" validate:"required"`
	B  DerefOrImmediate `json:"b" validate:"required"`
}

type InnerDerefOrImmediate interface{}

type DerefOrImmediate struct {
	Inner InnerDerefOrImmediate `json:"DerefOrImmediate" validate:"required"`
}

func (d *DerefOrImmediate) UnmarshalJSON(data []byte) error {
	var rawDerefOrImmediate map[string]json.RawMessage

	err := json.Unmarshal(data, &rawDerefOrImmediate)
	if err != nil {
		return err
	}

	var op any
	for k := range rawDerefOrImmediate {
		switch ResOperandName(k) {
		case DerefName:
			op = &Deref{}
		case ImmediateName:
			op = &Immediate{}
		default:
			return fmt.Errorf("expected deref or immediate, got %s", k)
		}
		break
	}

	if err = json.Unmarshal(data, op); err != nil {
		return err
	}

	d.Inner = op
	return nil
}

type HintArgs interface{}

type Hint struct {
	Name HintName `validate:"required"`
	Args HintArgs `validate:"required"`
}

func (h *Hint) UnmarshalJSON(data []byte) error {
	var rawHint map[string]json.RawMessage
	err := json.Unmarshal(data, &rawHint)
	if err != nil {
		return err
	}

	for k, v := range rawHint {
		h.Name = HintName(k)
		var args any

		switch h.Name {
		// Starknet hints
		case SystemCallName:
			args = &SystemCall{}
		case CheatcodeName:
			args = &Cheatcode{}
		// Core hints
		case AllocSegmentName:
			args = &AllocSegment{}
		case TestLessThanName:
			args = &TestLessThan{}
		case TestLessThanOrEqualName:
			args = &TestLessThanOrEqual{}
		case WideMul128Name:
			args = &WideMul128{}
		case DivModName:
			args = &DivMod{}
		case Uint256DivModName:
			args = &Uint256DivMod{}
		case Uint512DivModByUint256Name:
			args = &Uint512DivModByUint256{}
		case SquareRootName:
			args = &SquareRoot{}
		case Uint256SquareRootName:
			args = &Uint256SquareRoot{}
		case LinearSplitName:
			args = &LinearSplit{}
		case AllocFelt252DictName:
			args = &AllocFelt252Dict{}
		case Felt252DictEntryInitName:
			args = &Felt252DictEntryInit{}
		case Felt252DictEntryUpdateName:
			args = &Felt252DictEntryUpdate{}
		case GetSegmentArenaIndexName:
			args = &GetSegmentArenaIndex{}
		case InitSquashDataName:
			args = &InitSquashData{}
		case GetCurrentAccessIndexName:
			args = &GetCurrentAccessIndex{}
		case ShouldSkipSquashLoopName:
			args = &ShouldSkipSquashLoop{}
		case GetCurrentAccessDeltaName:
			args = &GetCurrentAccessDelta{}
		case ShouldContinueSquashLoopName:
			args = &ShouldContinueSquashLoop{}
		case GetNextDictKeyName:
			args = &GetNextDictKey{}
		case AssertLeFindSmallArcsName:
			args = &AssertLeFindSmallArcs{}
		case AssertLeIsFirstArcExcludedName:
			args = &AssertLeIsFirstArcExcluded{}
		case AssertLeIsSecondArcExcludedName:
			args = &AssertLeIsSecondArcExcluded{}
		case RandomEcPointName:
			args = &RandomEcPoint{}
		case FieldSqrtName:
			args = &FieldSqrt{}
		case DebugPrintName:
			args = &DebugPrint{}
		case AllocConstantSizeName:
			args = &AllocConstantSize{}
		// Deprecated hints
		case AssertCurrentAccessIndicesIsEmptyName:
			args = &AssertCurrentAccessIndicesIsEmpty{}
		case AssertAllAccessesUsedName:
			args = &AssertAllAccessesUsed{}
		case AssertAllKeysUsedName:
			args = &AssertAllKeysUsed{}
		case AssertLeAssertThirdArcExcludedName:
			args = &AssertLeAssertThirdArcExcluded{}
		case AssertLtAssertValidInputName:
			args = &AssertLtAssertValidInput{}
		case Felt252DictReadName:
			args = &Felt252DictRead{}
		case Felt252DictWriteName:
			args = &Felt252DictWrite{}
		default:
			return fmt.Errorf("unknown hint %s", k)
		}

		if err = json.Unmarshal(v, args); err != nil {
			return err
		}
		h.Args = args
		break
	}
	return nil
}
