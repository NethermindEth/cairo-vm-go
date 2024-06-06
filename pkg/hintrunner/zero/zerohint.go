package zero

import (
	"fmt"
	"strconv"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	zero "github.com/NethermindEth/cairo-vm-go/pkg/parsers/zero"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
)

// GenericZeroHinter wraps an adhoc Cairo0 inline (pythonic) hint implementation.
type GenericZeroHinter struct {
	Name string
	Op   func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error
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
	// Math hints
	case isLeFeltCode:
		return createIsLeFeltHinter(resolver)
	case assertLtFeltCode:
		return createAssertLtFeltHinter(resolver)
	case assertNotZeroCode:
		return createAssertNotZeroHinter(resolver)
	case assertNNCode:
		return createAssertNNHinter(resolver)
	case assertNotEqualCode:
		return createAssertNotEqualHinter(resolver)
	case assert250bits:
		return createAssert250bitsHinter(resolver)
	case assertLeFeltCode:
		return createAssertLeFeltHinter(resolver)
	case assertLeFeltExcluded0Code:
		return createAssertLeFeltExcluded0Hinter()
	case assertLeFeltExcluded1Code:
		return createAssertLeFeltExcluded1Hinter()
	case assertLeFeltExcluded2Code:
		return createAssertLeFeltExcluded2Hinter()
	case isNNCode:
		return createIsNNHinter(resolver)
	case isNNOutOfRangeCode:
		return createIsNNOutOfRangeHinter(resolver)
	case isPositiveCode:
		return createIsPositiveHinter(resolver)
	case splitIntAssertRange:
		return createSplitIntAssertRangeHinter(resolver)
	case splitIntCode:
		return createSplitIntHinter(resolver)
	case signedDivRemCode:
		return createSignedDivRemHinter(resolver)
	case powCode:
		return createPowHinter(resolver)
	case splitFeltCode:
		return createSplitFeltHinter(resolver)
	case sqrtCode:
		return createSqrtHinter(resolver)
	case unsignedDivRemCode:
		return createUnsignedDivRemHinter(resolver)
	case isQuadResidueCode:
		return createIsQuadResidueHinter(resolver)
	// Uint256 hints
	case uint256AddCode:
		return createUint256AddHinter(resolver, false)
	case uint256AddLowCode:
		return createUint256AddHinter(resolver, true)
	case split64Code:
		return createSplit64Hinter(resolver)
	case uint256SignedNNCode:
		return createUint256SignedNNHinter(resolver)
	case uint256UnsignedDivRemCode:
		return createUint256UnsignedDivRemHinter(resolver)
	case uint256SqrtCode:
		return createUint256SqrtHinter(resolver)
	case uint256MulDivModCode:
		return createUint256MulDivModHinter(resolver)
	// Signature hints
	case verifyECDSASignatureCode:
		return createVerifyECDSASignatureHinter(resolver)
	case getPointFromXCode:
		return createGetPointFromXHinter(resolver)
	case divModNSafeDivCode:
		return createDivModSafeDivHinter()
	case importSecp256R1PCode:
		return createImportSecp256R1PHinter()
	case verifyZeroCode:
		return createVerifyZeroHinter(resolver)
	case divModNPackedDivmodV1Code:
		return createDivModNPackedDivmodV1Hinter(resolver)
	// EC hints
	case ecNegateCode:
		return createEcNegateHinter(resolver)
	case nondetBigint3V1Code:
		return createNondetBigint3V1Hinter(resolver)
	case fastEcAddAssignNewXCode:
		return createFastEcAddAssignNewXHinter(resolver)
	case fastEcAddAssignNewYCode:
		return createFastEcAddAssignNewYHinter()
	case ecDoubleSlopeV1Code:
		return createEcDoubleSlopeV1Hinter(resolver)
	case reduceV1Code:
		return createReduceV1Hinter(resolver)
	case computeSlopeV1Code:
		return createComputeSlopeV1Hinter(resolver)
	case ecDoubleAssignNewXV1:
		return createEcDoubleAssignNewXV1Hinter(resolver)
	case ecDoubleAssignNewYV1:
		return createEcDoubleAssignNewYV1Hinter()
	// Blake hints
	case blake2sAddUint256BigendCode:
		return createBlake2sAddUint256Hinter(resolver, true)
	case blake2sAddUint256Code:
		return createBlake2sAddUint256Hinter(resolver, false)
	case blake2sFinalizeCode:
		return createBlake2sFinalizeHinter(resolver)
	// Keccak hints
	case keccakWriteArgsCode:
		return createKeccakWriteArgsHinter(resolver)
	case blockPermutationCode:
		return createBlockPermutationHinter(resolver)
	// Usort hints
	case usortEnterScopeCode:
		return createUsortEnterScopeHinter()
	case usortVerifyMultiplicityAssertCode:
		return createUsortVerifyMultiplicityAssertHinter()
	case usortVerifyCode:
		return createUsortVerifyHinter(resolver)
	case usortVerifyMultiplicityBodyCode:
		return createUsortVerifyMultiplicityBodyHinter(resolver)
	case usortBodyCode:
		return createUsortBodyHinter(resolver)
	// Dictionaries hints
	case defaultDictNewCode:
		return createDefaultDictNewHinter(resolver)
	case dictReadCode:
		return createDictReadHinter(resolver)
	case squashDictCode:
		return createSquashDictHinter(resolver)
	case squashDictInnerAssertLenKeys:
		return createSquashDictInnerAssertLenKeysHinter()
	case squashDictInnerContinueLoop:
		return createSquashDictInnerContinueLoopHinter(resolver)
	case squashDictInnerSkipLoop:
		return createSquashDictInnerSkipLoopHinter(resolver)
	case squashDictInnerLenAssert:
		return createSquashDictInnerLenAssertHinter()
	case squashDictInnerNextKey:
		return createSquashDictInnerNextKeyHinter(resolver)
	case squashDictInnerUsedAccessesAssert:
		return createSquashDictInnerUsedAccessesAssertHinter(resolver)
	// Other hints
	case allocSegmentCode:
		return createAllocSegmentHinter()
	case memcpyContinueCopyingCode:
		return createMemContinueHinter(resolver, false)
	case memsetContinueLoopCode:
		return createMemContinueHinter(resolver, true)
	case vmEnterScopeCode:
		return createVMEnterScopeHinter()
	case memcpyEnterScopeCode:
		return createMemEnterScopeHinter(resolver, false)
	case memsetEnterScopeCode:
		return createMemEnterScopeHinter(resolver, true)
	case vmExitScopeCode:
		return createVMExitScopeHinter()
	case testAssignCode:
		return createTestAssignHinter(resolver)
	default:
		return nil, fmt.Errorf("not identified hint")
	}
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
