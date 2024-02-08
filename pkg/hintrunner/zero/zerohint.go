package zero

import (
	"fmt"
	"strconv"
	"strings"

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
	resolver, err := getParameters(program, rawHint, hintPC)
	if err != nil {
		return nil, err
	}

	switch rawHint.Code {
	case AllocSegmentCode:
		return CreateAllocSegmentHinter(resolver)
	case TestAssignCode:
		return createTestAssignHinter(resolver)
	default:
		return nil, fmt.Errorf("Not identified hint")
	}
}

func CreateAllocSegmentHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	if resolver.NumResOperanders()+resolver.NumCellRefers() != 0 {
		return nil, fmt.Errorf("Expected no arguments for %s hint", sn.AllocSegmentName)
	}
	return &core.AllocSegment{Dst: hinter.ApCellRef(0)}, nil
}

func createTestAssignHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	if resolver.NumResOperanders() < 1 {
		return nil, fmt.Errorf("Expected at least 1 ResOperander")
	}

	arg := resolver.GetResOperander("a")

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

// shortSymbolName turns a full symbol name like "a.b.c" into just "c".
func shortSymbolName(name string) string {
	i := strings.LastIndexByte(name, '.')
	if i != -1 {
		return name[i+1:]
	}
	return name
}

func getParameters(zeroProgram *zero.ZeroProgram, hint zero.Hint, hintPC uint64) (hintReferenceResolver, error) {
	var resolver hintReferenceResolver

	for referenceName := range hint.FlowTrackingData.ReferenceIds {
		rawIdentifier, ok := zeroProgram.Identifiers[referenceName]
		if !ok {
			return resolver, fmt.Errorf("missing identifier %s", referenceName)
		}

		if len(rawIdentifier.References) == 0 {
			return resolver, fmt.Errorf("identifier %s should have at least one reference", referenceName)
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
			return resolver, fmt.Errorf("identifier %s should have a reference with pc smaller or equal than %d", referenceName, hintPC)
		}

		param, err := ParseIdentifier(reference.Value)
		if err != nil {
			return resolver, err
		}
		param = applyApTracking(zeroProgram, hint, reference, param)
		switch result := param.(type) {
		case hinter.CellRefer:
			resolver.AddCellRefer(shortSymbolName(referenceName), result)
		case hinter.ResOperander:
			resolver.AddResOperander(shortSymbolName(referenceName), result)
		default:
			return resolver, fmt.Errorf("unexpected type for identifier value %s", reference.Value)
		}
	}

	return resolver, nil
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
