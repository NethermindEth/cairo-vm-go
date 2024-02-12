package zero

import (
	"fmt"
	"strconv"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/core"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
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
	return &core.AllocSegment{Dst: hinter.ApCellRef(0)}, nil
}

func createTestAssignHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	arg, err := resolver.GetReference("a")
	if err != nil {
		return nil, err
	}

	a, ok := arg.(hinter.ResOperander)
	if !ok {
		return nil, fmt.Errorf("expected a ResOperander reference")
	}

	h := &GenericZeroHinter{
		Name: "TestAssign",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			apAddr := vm.Context.AddressAp()
			v, err := a.Resolve(vm)
			if err != nil {
				return err
			}
			return vm.Memory.WriteToAddress(&apAddr, &v)
		},
	}
	return h, nil
}

func getParameters(zeroProgram *zero.ZeroProgram, hint zero.Hint, hintPC uint64) (hintReferenceResolver, error) {
	resolver := NewReferenceResolver()

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
		param = param.ApplyApTracking(hint.FlowTrackingData.ApTracking, reference.ApTrackingData)
		if err := resolver.AddReference(referenceName, param); err != nil {
			return resolver, err
		}
	}

	return resolver, nil
}
