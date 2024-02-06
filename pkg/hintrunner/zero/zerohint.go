package zero

import (
	"fmt"
	"strconv"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/core"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	sn "github.com/NethermindEth/cairo-vm-go/pkg/parsers/starknet"
	zero "github.com/NethermindEth/cairo-vm-go/pkg/parsers/zero"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
)

// GenericZeroHinter wraps an adhoc Cairo0 inline (pythonic) hint implementation.
type GenericZeroHinter struct {
	Name string
	Op   func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error
}

func (hint *GenericZeroHinter) String() string {
	return hint.Name
}

func (hint *GenericZeroHinter) Execute(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
	return hint.Op(vm, ctx)
}

func GetZeroHints(cairoZeroJson *zero.ZeroProgram) (map[uint64][]hinter.Hinter, error) {
	hints := make(map[uint64][]hinter.Hinter)
	for counter, rawHints := range cairoZeroJson.Hints {
		pc, err := strconv.ParseUint(counter, 10, 64)
		if err != nil {
			return nil, err
		}

		for _, rawHint := range rawHints {
			hint, err := GetHintFromCode(cairoZeroJson, rawHint, pc)
			if err != nil {
				return nil, err
			}

			hints[pc] = append(hints[pc], hint)
		}
	}

	return hints, nil
}

func GetHintFromCode(program *zero.ZeroProgram, rawHint zero.Hint, hintPC uint64) (hinter.Hinter, error) {
	cellRefParams, resOpParams, err := GetParameters(program, rawHint, hintPC)
	if err != nil {
		return nil, err
	}

	switch rawHint.Code {
	case AllocSegmentCode:
		return CreateAllocSegmentHinter(cellRefParams, resOpParams)
	case TestAssignCode:
		return createTestAssignHinter(cellRefParams, resOpParams)
	default:
		return nil, fmt.Errorf("Not identified hint")
	}
}

func CreateAllocSegmentHinter(cellRefParams []hinter.CellRefer, resOpParams []hinter.ResOperander) (hinter.Hinter, error) {
	if len(cellRefParams)+len(resOpParams) != 0 {
		return nil, fmt.Errorf("Expected no arguments for %s hint", sn.AllocSegmentName)
	}
	return &core.AllocSegment{Dst: hinter.ApCellRef(0)}, nil
}

func createTestAssignHinter(cellRefParams []hinter.CellRefer, resOpParams []hinter.ResOperander) (hinter.Hinter, error) {
	if len(resOpParams) < 1 {
		return nil, fmt.Errorf("Expected at least 1 ResOperander")
	}
	if len(cellRefParams) != 0 {
		return nil, fmt.Errorf("Expected 0 CellRefers (got %d)", len(cellRefParams))
	}

	// Given a Cairo0 code like this:
	//
	//	func fp_args_sum(arg1: felt, arg2: felt) -> felt {
	//		let a = arg1 + arg2;
	//		%{ memory[ap] = ids.a %}
	//		return [ap];
	//	}
	//
	// We get ids.a defined as a reference that refers to 2 function arguments.
	// When we execute GetParameters(), it will return 3 ResOperander for the hint:
	// one for the ids.a itself, and two for the args (as they're referenced from it).
	//
	//	"__main__.fp_args_sum.a"="cast([fp + (-4)] + [fp + (-3)], felt)"
	//	"__main__.fp_args_sum.arg1"="[cast(fp + (-4), felt*)]"
	//	"__main__.fp_args_sum.arg2"="[cast(fp + (-3), felt*)]"
	//
	// It's not entirely clear how to handle this situation yet.
	// It looks like we usually need exactly 1 reference per literal identifier (like ids.a),
	// but we may not need references that are not present in the hint's code directly.
	//
	// For now, I'll just take the first one (for tests purposes) and leave this issue for discussion.
	// Right now I want to test the ApTracking address calculation, not hint's reference collection.
	arg := resOpParams[0]

	h := &GenericZeroHinter{
		Name: "TestAssign",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			apAddr := vm.Context.AddressAp()
			v, err := arg.Resolve(vm)
			if err != nil {
				return err
			}
			return vm.Memory.WriteToAddress(&apAddr, &v)
		},
	}
	return h, nil
}

func GetParameters(zeroProgram *zero.ZeroProgram, hint zero.Hint, hintPC uint64) ([]hinter.CellRefer, []hinter.ResOperander, error) {
	var cellRefParams []hinter.CellRefer
	var resOpParams []hinter.ResOperander
	for referenceName := range hint.FlowTrackingData.ReferenceIds {
		rawIdentifier, ok := zeroProgram.Identifiers[referenceName]
		if !ok {
			return nil, nil, fmt.Errorf("missing identifier %s", referenceName)
		}

		if len(rawIdentifier.References) == 0 {
			return nil, nil, fmt.Errorf("identifier %s should have at least one reference", referenceName)
		}
		references := rawIdentifier.References

		// Go through the references in reverse order to get the one with biggest pc smaller or equal to the hint pc
		var reference zero.Reference
		ok = false
		for i := len(references) - 1; i >= 0; i-- {
			if references[i].Pc <= hintPC {
				reference = references[i]
				ok = true
				break
			}
		}
		if !ok {
			return nil, nil, fmt.Errorf("identifier %s should have a reference with pc smaller or equal than %d", referenceName, hintPC)
		}

		param, err := ParseIdentifier(reference.Value)
		if err != nil {
			return nil, nil, err
		}
		param = applyApTracking(zeroProgram, hint, reference, param)
		switch result := param.(type) {
		case hinter.CellRefer:
			cellRefParams = append(cellRefParams, result)
		case hinter.ResOperander:
			resOpParams = append(resOpParams, result)
		default:
			return nil, nil, fmt.Errorf("unexpected type for identifier value %s", reference.Value)
		}
	}

	return cellRefParams, resOpParams, nil
}

func applyApTracking(p *zero.ZeroProgram, h zero.Hint, ref zero.Reference, v any) any {
	// We can't make an inplace modification because the v's underlying type is not a pointer type.
	// Therefore, we need to return it from the function.
	// This makes this function less elegant: it requires type asserts, etc.

	switch v := v.(type) {
	case hinter.ApCellRef:
		if h.FlowTrackingData.ApTracking.Group != ref.ApTrackingData.Group {
			return v // Group mismatched: nothing to adjust
		}
		newOffset := v - hinter.ApCellRef(h.FlowTrackingData.ApTracking.Offset-ref.ApTrackingData.Offset)
		return hinter.ApCellRef(newOffset)

	case hinter.Deref:
		v.Deref = applyApTracking(p, h, ref, v.Deref).(hinter.CellRefer)
		return v

	case hinter.DoubleDeref:
		v.Deref = applyApTracking(p, h, ref, v.Deref).(hinter.CellRefer)
		return v

	case hinter.BinaryOp:
		v.Lhs = applyApTracking(p, h, ref, v.Lhs).(hinter.CellRefer)
		v.Rhs = applyApTracking(p, h, ref, v.Rhs).(hinter.ResOperander)
		return v

	default:
		// This case covers type that we don't need to visit.
		// E.g. FpCellRef, Immediate.
		return v
	}
}
