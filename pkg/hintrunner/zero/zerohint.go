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
			hint, err := GetHintFromCode(cairoZeroJson, rawHint)
			if err != nil {
				return nil, err
			}

			hints[pc] = append(hints[pc], hint)
		}
	}

	return hints, nil
}

func GetHintFromCode(program *zero.ZeroProgram, rawHint zero.Hint) (hinter.Hinter, error) {
	resolver, err := getParameters(program, rawHint)
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
	case assert250bitsCode:
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
	case splitIntAssertRangeCode:
		return createSplitIntAssertRangeHinter(resolver)
	case splitIntCode:
		return createSplitIntHinter(resolver)
	case signedDivRemCode:
		return createSignedDivRemHinter(resolver)
	case powCode:
		return createPowHinter(resolver)
	case signedPowCode:
		return createSignedPowHinter(resolver)
	case splitFeltCode:
		return createSplitFeltHinter(resolver)
	case sqrtCode:
		return createSqrtHinter(resolver)
	case unsignedDivRemCode:
		return createUnsignedDivRemHinter(resolver)
	case isQuadResidueCode:
		return createIsQuadResidueHinter(resolver)
	case getHighLenCode:
		return createGetHighLenHinter(resolver)
	case split128Code:
		return createSplit128Hinter(resolver)
	case is250BitsCode:
		return createIs250BitsHinter(resolver)
	// Uint256 hints
	case uint128AddCode:
		return createUint128AddHinter(resolver)
	case uint128SqrtCode:
		return createUint128SqrtHinter(resolver)
	case uint256AddCode:
		return createUint256AddHinter(resolver)
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
	case uint256SubCode:
		return createUint256SubHinter(resolver)
	case uint256UnsignedDivRemExpandedCode:
		return createUint256UnsignedDivRemExpandedHinter(resolver)
	case splitXXCode:
		return createSplitXXHinter(resolver)
	case invModPUint512Code:
		return createInvModPUint512Hinter(resolver)
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
	case verifyZeroV2Code:
		return createVerifyZeroHinter(resolver)
	case verifyZeroV3Code:
		return createVerifyZeroV3Hinter(resolver)
	case verifyZeroAltCode:
		return createVerifyZeroHinter(resolver)
	case divModNPackedDivmodV1Code:
		return createDivModNPackedDivmodV1Hinter(resolver)
	case importSECP256R1NCode:
		return createImportSECP256R1NHinter()
	// EC hints
	case bigIntToUint256Code:
		return createBigIntToUint256Hinter(resolver)
	case ecNegateCode:
		return createEcNegateHinter(resolver)
	case divModNSafeDivPlusOneCode:
		return createDivModNSafeDivPlusOneHinter()
	case divModNPackedDivModExternalNCode:
		return createDivModNPackedDivModExternalNHinter(resolver)
	case nondetBigint3V1Code:
		return createNondetBigint3V1Hinter(resolver)
	case fastEcAddAssignNewXCode:
		return createFastEcAddAssignNewXHinter(resolver)
	case fastEcAddAssignNewXV2Code:
		return createFastEcAddAssignNewXV2Hinter(resolver)
	case fastEcAddAssignNewXV3Code:
		return createFastEcAddAssignNewXV3Hinter(resolver)
	case fastEcAddAssignNewYCode:
		return createFastEcAddAssignNewYHinter()
	case ecDoubleSlopeV1Code:
		return createEcDoubleSlopeV1Hinter(resolver)
	case ecDoubleSlopeV3Code:
		return createEcDoubleSlopeV3Hinter(resolver)
	case reduceV1Code:
		return createReduceHinter(resolver)
	case reduceV2Code:
		return createReduceHinter(resolver)
	case reduceEd25519Code:
		return createReduceEd25519Hinter(resolver)
	case computeSlopeV1Code:
		return createComputeSlopeV1Hinter(resolver)
	case computeSlopeV2Code:
		return createComputeSlopeV2Hinter(resolver)
	case computeSlopeV3Code:
		return createComputeSlopeV3Hinter(resolver)
	case ecDoubleAssignNewXV1Code:
		return createEcDoubleAssignNewXV1Hinter(resolver)
	case ecDoubleAssignNewXV2Code:
		return createEcDoubleAssignNewXV2Hinter(resolver)
	case ecDoubleAssignNewXV4Code:
		return createEcDoubleAssignNewXV4Hinter(resolver)
	case ecDoubleAssignNewYV1Code:
		return createEcDoubleAssignNewYV1Hinter()
	case ecMulInnerCode:
		return createEcMulInnerHinter(resolver)
	case isZeroNondetCode:
		return createIsZeroNondetHinter()
	case isZeroPackCode:
		return createIsZeroPackHinter(resolver)
	case isZeroDivModCode:
		return createIsZeroDivModHinter()
	case recoverYCode:
		return createRecoverYHinter(resolver)
	case randomEcPointCode:
		return createRandomEcPointHinter(resolver)
	case chainedEcOpCode:
		return createChainedEcOpHinter(resolver)
	case bigIntPackDivModCode:
		return createBigIntPackDivModHinter(resolver)
	case bigIntSaveDivCode:
		return createBigIntSaveDivHinter(resolver)
	case ecRecoverDivModNPackedCode:
		return createEcRecoverDivModNPackedHinter(resolver)
	case ecRecoverSubABCode:
		return createEcRecoverSubABHinter(resolver)
	case ecRecoverProductModCode:
		return createEcRecoverProductModHinter(resolver)
	case ecRecoverProductDivMCode:
		return createEcRecoverProductDivMHinter()
	// Blake hints
	case blake2sAddUint256BigendCode:
		return createBlake2sAddUint256Hinter(resolver, true)
	case blake2sAddUint256Code:
		return createBlake2sAddUint256Hinter(resolver, false)
	case blake2sFinalizeCode:
		return createBlake2sFinalizeHinter(resolver)
	case blake2sFinalizeV2Code:
		return createBlake2sFinalizeHinter(resolver)
	case blake2sFinalizeV3Code:
		return createBlake2sFinalizeV3Hinter(resolver)
	case blake2sComputeCode:
		return createBlake2sComputeHinter(resolver)
	case blake2sCompressCode:
		return createBlake2sCompressHinter(resolver)
	// Sha256 hints
	case packedSha256Code:
		return createPackedSha256Hinter(resolver)
	case sha256ChunkCode:
		return createSha256ChunkHinter(resolver)
	case finalizeSha256Code:
		return createFinalizeSha256Hinter(resolver)
	// Keccak hints
	case keccakWriteArgsCode:
		return createKeccakWriteArgsHinter(resolver)
	case cairoKeccakFinalizeCode:
		return createCairoKeccakFinalizeHinter(resolver)
	case cairoKeccakFinalizeBlockSize1000Code:
		return createCairoKeccakFinalizeHinter(resolver)
	case unsafeKeccakCode:
		return createUnsafeKeccakHinter(resolver)
	case unsafeKeccakFinalizeCode:
		return createUnsafeKeccakFinalizeHinter(resolver)
	case compareKeccakFullRateInBytesCode:
		return createCompareKeccakFullRateInBytesNondetHinter(resolver)
	case blockPermutationCode:
		return createBlockPermutationHinter(resolver)
	case compareBytesInWordCode:
		return createCompareBytesInWordNondetHinter(resolver)
	case splitInput3Code:
		return createSplitInput3Hinter(resolver)
	case splitInput6Code:
		return createSplitInput6Hinter(resolver)
	case splitInput9Code:
		return createSplitInput9Hinter(resolver)
	case splitInput12Code:
		return createSplitInput12Hinter(resolver)
	case splitInput15Code:
		return createSplitInput15Hinter(resolver)
	case splitOutputMidLowHighCode:
		return createSplitOutputMidLowHighHinter(resolver)
	case splitOutput0Code:
		return createSplitOutput0Hinter(resolver)
	case SplitNBytesCode:
		return createSplitNBytesHinter(resolver)
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
	case dictNewCode:
		return createDictNewHinter()
	case defaultDictNewCode:
		return createDefaultDictNewHinter(resolver)
	case dictReadCode:
		return createDictReadHinter(resolver)
	case dictSquashCopyDictCode:
		return createDictSquashCopyDictHinter(resolver)
	case dictWriteCode:
		return createDictWriteHinter(resolver)
	case dictUpdateCode:
		return createDictUpdateHinter(resolver)
	case squashDictCode:
		return createSquashDictHinter(resolver)
	case squashDictInnerAssertLenKeysCode:
		return createSquashDictInnerAssertLenKeysHinter()
	case squashDictInnerCheckAccessIndexCode:
		return createSquashDictInnerCheckAccessIndexHinter(resolver)
	case squashDictInnerContinueLoopCode:
		return createSquashDictInnerContinueLoopHinter(resolver)
	case squashDictInnerFirstIterationCode:
		return createSquashDictInnerFirstIterationHinter(resolver)
	case squashDictInnerSkipLoopCode:
		return createSquashDictInnerSkipLoopHinter(resolver)
	case squashDictInnerLenAssertCode:
		return createSquashDictInnerLenAssertHinter()
	case squashDictInnerNextKeyCode:
		return createSquashDictInnerNextKeyHinter(resolver)
	case squashDictInnerUsedAccessesAssertCode:
		return createSquashDictInnerUsedAccessesAssertHinter(resolver)
	case dictSquashUpdatePtrCode:
		return createDictSquashUpdatePtrHinter(resolver)
	// Other hints
	case allocSegmentCode:
		return createAllocSegmentHinter()
	case memcpyContinueCopyingCode:
		return createMemContinueHinter(resolver, false)
	case memsetContinueLoopCode:
		return createMemContinueHinter(resolver, true)
	case memcpyEnterScopeCode:
		return createMemEnterScopeHinter(resolver, false)
	case memsetEnterScopeCode:
		return createMemEnterScopeHinter(resolver, true)
	case searchSortedLowerCode:
		return createSearchSortedLowerHinter(resolver)
	case vmEnterScopeCode:
		return createVMEnterScopeHinter()
	case vmExitScopeCode:
		return createVMExitScopeHinter()
	case getFeltBitLengthCode:
		return createGetFeltBitLengthHinter(resolver)
	case setAddCode:
		return createSetAddHinter(resolver)
	case testAssignCode:
		return createTestAssignHinter(resolver)
	case findElementCode:
		return createFindElementHinter(resolver)
	case nondetElementsOverTwoCode:
		return createNondetElementsOverXHinter(resolver, 2)
	case nondetElementsOverTenCode:
		return createNondetElementsOverXHinter(resolver, 10)
	case normalizeAddressCode:
		return createNormalizeAddressHinter(resolver)
	case sha256AndBlake2sInputCode:
		return createSha256AndBlake2sInputHinter(resolver)
	default:
		return nil, fmt.Errorf("not identified hint: \n%s", rawHint.Code)
	}
}

func getParameters(zeroProgram *zero.ZeroProgram, hint zero.Hint) (hintReferenceResolver, error) {
	resolver := NewReferenceResolver()

	for referenceName, id := range hint.FlowTrackingData.ReferenceIds {
		rawIdentifier, ok := zeroProgram.Identifiers[referenceName]
		if !ok {
			return resolver, fmt.Errorf("missing identifier %s", referenceName)
		}

		if len(rawIdentifier.References) == 0 {
			return resolver, fmt.Errorf("identifier %s should have at least one reference", referenceName)
		}
		if int(id) >= len(zeroProgram.ReferenceManager.References) {
			return resolver, fmt.Errorf("invalid reference id %d", id)
		}
		reference := zeroProgram.ReferenceManager.References[id]

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
