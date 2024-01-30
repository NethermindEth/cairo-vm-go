package zero

import (
	"fmt"
	"strconv"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/core"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	sn "github.com/NethermindEth/cairo-vm-go/pkg/parsers/starknet"
	zero "github.com/NethermindEth/cairo-vm-go/pkg/parsers/zero"
)

func GetZeroHints(cairoZeroJson *zero.ZeroProgram) (map[uint64]hinter.Hinter, error) {
	hints := make(map[uint64]hinter.Hinter)
	for counter, rawHints := range cairoZeroJson.Hints {
		pc, err := strconv.ParseUint(counter, 10, 64)
		if err != nil {
			return nil, err
		}

		// TODO: Check if it is possible to have more than one hint
		if len(rawHints) != 1 {
			return nil, fmt.Errorf("expected only 1 hint but got  %d", len(rawHints))
		}
		rawHint := rawHints[0]

		hint, err := GetHintFromCode(cairoZeroJson, rawHint, pc)
		if err != nil {
			return nil, err
		}

		hints[pc] = hint
	}

	return hints, nil
}

func GetHintFromCode(program *zero.ZeroProgram, rawHint zero.Hint, hintPC uint64) (hinter.Hinter, error){
	cellRefParams, resOpParams, err := GetParameters(program, rawHint, hintPC)
	if err != nil {
		return nil, err
	}

	switch rawHint.Code {
	case AllocSegmentCode:
		return CreateAllocSegmentHinter(cellRefParams, resOpParams)
	default:
		return nil, fmt.Errorf("Not identified hint")
	}
}

func CreateAllocSegmentHinter(cellRefParams []hinter.CellRefer, resOpParams []hinter.ResOperander) (hinter.Hinter, error) {
	if len(cellRefParams) + len(resOpParams) != 0 {
		return nil, fmt.Errorf("Expected no arguments for %s hint", sn.AllocSegmentName)
	}
	return &core.AllocSegment { Dst: hinter.ApCellRef(0) }, nil
}

func GetParameters(zeroProgram *zero.ZeroProgram, hint zero.Hint, hintPC uint64) ([]hinter.CellRefer, []hinter.ResOperander, error) {
	var cellRefParams []hinter.CellRefer
	var resOpParams []hinter.ResOperander
	for referenceName := range hint.FlowTrackingData.ReferenceIds {
		rawIdentifier, ok := zeroProgram.Identifiers[referenceName]
		if !ok {
			return nil, nil, fmt.Errorf("missing identifier %s", referenceName)
		}
		identifier, ok := rawIdentifier.(map[string]any)
		if !ok {
			return nil, nil, fmt.Errorf("wrong structure for identifier")
		}

		rawReferences, ok := identifier["references"]
		if !ok {
			return nil, nil, fmt.Errorf("identifier %s should have at least one reference", referenceName)
		}
		references, ok := rawReferences.([]zero.Reference)
		if !ok {
			return nil, nil, fmt.Errorf("expected a list of references")
		}

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
		switch result := param.(type){
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
