package hintrunner

import (
	"math/big"
	"testing"

	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/require"
)

func TestAllocSegment(t *testing.T) {
	vm := defaultVirtualMachine()
	vm.Context.Ap = 3
	vm.Context.Fp = 0

	var ap ApCellRef = 5
	var fp FpCellRef = 9

	alloc1 := AllocSegment{ap}
	alloc2 := AllocSegment{fp}

	err := alloc1.Execute(vm)
	require.Nil(t, err)
	require.Equal(t, 3, len(vm.Memory.Segments))
	require.Equal(
		t,
		memory.MemoryValueFromSegmentAndOffset(2, 0),
		readFrom(vm, VM.ExecutionSegment, vm.Context.Ap+5),
	)

	err = alloc2.Execute(vm)
	require.Nil(t, err)
	require.Equal(t, 4, len(vm.Memory.Segments))
	require.Equal(
		t,
		memory.MemoryValueFromSegmentAndOffset(3, 0),
		readFrom(vm, VM.ExecutionSegment, vm.Context.Fp+9),
	)

}

func TestTestLessThanTrue(t *testing.T) {
	vm := defaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0
	writeTo(vm, VM.ExecutionSegment, 0, memory.MemoryValueFromInt(23))

	var dst ApCellRef = 1
	var rhsRef FpCellRef = 0
	rhs := Deref{rhsRef}

	lhs := Immediate(*big.NewInt(13))

	hint := TestLessThan{
		dst: dst,
		lhs: lhs,
		rhs: rhs,
	}

	err := hint.Execute(vm)
	require.NoError(t, err)
	require.Equal(
		t,
		memory.MemoryValueFromInt(1),
		readFrom(vm, VM.ExecutionSegment, 1),
		"Expected the hint to evaluate to True when lhs is less than rhs",
	)
}
func TestTestLessThanFalse(t *testing.T) {
	testCases := []struct {
		lhsValue    *big.Int
		expectedMsg string
	}{
		{big.NewInt(32), "Expected the hint to evaluate to False when lhs is larger"},
		{big.NewInt(17), "Expected the hint to evaluate to False when values are equal"},
	}

	for _, tc := range testCases {
		t.Run(tc.expectedMsg, func(t *testing.T) {
			vm := defaultVirtualMachine()
			vm.Context.Ap = 0
			vm.Context.Fp = 0
			writeTo(vm, VM.ExecutionSegment, 0, memory.MemoryValueFromInt(17))

			var dst ApCellRef = 1
			var rhsRef FpCellRef = 0
			rhs := Deref{rhsRef}

			lhs := Immediate(*tc.lhsValue)
			hint := TestLessThan{
				dst: dst,
				lhs: lhs,
				rhs: rhs,
			}

			err := hint.Execute(vm)
			require.NoError(t, err)
			require.Equal(
				t,
				memory.EmptyMemoryValueAsFelt(),
				readFrom(vm, VM.ExecutionSegment, 1),
				tc.expectedMsg,
			)
		})
	}
}

func TestTestLessThanOrEqTrue(t *testing.T) {
	testCases := []struct {
		lhsValue    *big.Int
		expectedMsg string
	}{
		{big.NewInt(13), "Expected the hint to evaluate to True when lhs is less than rhs"},
		{big.NewInt(23), "Expected the hint to evaluate to True when values are equal"},
	}

	for _, tc := range testCases {
		t.Run(tc.expectedMsg, func(t *testing.T) {
			vm := defaultVirtualMachine()
			vm.Context.Ap = 0
			vm.Context.Fp = 0
			writeTo(vm, VM.ExecutionSegment, 0, memory.MemoryValueFromInt(23))

			var dst ApCellRef = 1
			var rhsRef FpCellRef = 0
			rhs := Deref{rhsRef}

			lhs := Immediate(*tc.lhsValue)
			hint := TestLessThanOrEqual{
				dst: dst,
				lhs: lhs,
				rhs: rhs,
			}

			err := hint.Execute(vm)
			require.NoError(t, err)
			require.Equal(
				t,
				memory.MemoryValueFromInt(1),
				readFrom(vm, VM.ExecutionSegment, 1),
				tc.expectedMsg,
			)
		})
	}
}

func TestTestLessThanOrEqFalse(t *testing.T) {
	vm := defaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0
	writeTo(vm, VM.ExecutionSegment, 0, memory.MemoryValueFromInt(17))

	var dst ApCellRef = 1
	var rhsRef FpCellRef = 0
	rhs := Deref{rhsRef}

	lhs := Immediate(*big.NewInt(32))

	hint := TestLessThanOrEqual{
		dst: dst,
		lhs: lhs,
		rhs: rhs,
	}

	err := hint.Execute(vm)
	require.NoError(t, err)
	require.Equal(
		t,
		memory.EmptyMemoryValueAsFelt(),
		readFrom(vm, VM.ExecutionSegment, 1),
		"Expected the hint to evaluate to False when lhs is larger",
	)
}

func TestWideMul128(t *testing.T) {
	vm := defaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	var dstLow ApCellRef = 1
	var dstHigh ApCellRef = 2

	lhs := Immediate(*big.NewInt(1).Lsh(big.NewInt(1), 127))
	rhs := Immediate(*big.NewInt(1<<8 + 1))

	hint := WideMul128{
		low:  dstLow,
		high: dstHigh,
		lhs:  lhs,
		rhs:  rhs,
	}

	err := hint.Execute(vm)
	require.Nil(t, err)

	low := &f.Element{}
	low.SetBigInt(big.NewInt(1).Lsh(big.NewInt(1), 127))

	require.Equal(
		t,
		memory.MemoryValueFromFieldElement(low),
		readFrom(vm, VM.ExecutionSegment, 1),
	)
	require.Equal(
		t,
		memory.MemoryValueFromInt(1<<7),
		readFrom(vm, VM.ExecutionSegment, 2),
	)
}

func TestWideMul128IncorrectRange(t *testing.T) {
	vm := defaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	var dstLow ApCellRef = 1
	var dstHigh ApCellRef = 2

	lhs := Immediate(*big.NewInt(1).Lsh(big.NewInt(1), 128))
	rhs := Immediate(*big.NewInt(1))

	hint := WideMul128{
		low:  dstLow,
		high: dstHigh,
		lhs:  lhs,
		rhs:  rhs,
	}

	err := hint.Execute(vm)
	require.ErrorContains(t, err, "should be u128")
}

func TestSquareRoot(t *testing.T) {
	vm := defaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	var dst ApCellRef = 1

	value := Immediate(*big.NewInt(36))
	hint := SquareRoot{
		value: value,
		dst:   dst,
	}

	err := hint.Execute(vm)

	require.NoError(t, err)
	require.Equal(
		t,
		memory.MemoryValueFromInt(6),
		readFrom(vm, VM.ExecutionSegment, 1),
	)
}

func TestUint256SquareRootLow(t *testing.T) {
	vm := defaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	var sqrt0 ApCellRef = 1
	var sqrt1 ApCellRef = 2
	var remainderLow ApCellRef = 3
	var remainderHigh ApCellRef = 4
	var sqrtMul2MinusRemainderGeU128 ApCellRef = 5

	valueLow := Immediate(*big.NewInt(121))
	valueHigh := Immediate(*big.NewInt(0))

	hint := Uint256SquareRoot{
		valueLow:                     valueLow,
		valueHigh:                    valueHigh,
		sqrt0:                        sqrt0,
		sqrt1:                        sqrt1,
		remainderLow:                 remainderLow,
		remainderHigh:                remainderHigh,
		sqrtMul2MinusRemainderGeU128: sqrtMul2MinusRemainderGeU128,
	}

	err := hint.Execute(vm)

	require.NoError(t, err)

	expectedSqrt0 := memory.MemoryValueFromInt(11)
	expectedSqrt1 := memory.MemoryValueFromInt(0)
	expectedRemainderLow := memory.MemoryValueFromInt(0)
	expectedRemainderHigh := memory.MemoryValueFromInt(0)
	expectedSqrtMul2MinusRemainderGeU128 := memory.MemoryValueFromInt(0)

	actualSqrt0 := readFrom(vm, VM.ExecutionSegment, 1)
	actualSqrt1 := readFrom(vm, VM.ExecutionSegment, 2)
	actualRemainderLow := readFrom(vm, VM.ExecutionSegment, 3)
	actualRemainderHigh := readFrom(vm, VM.ExecutionSegment, 4)
	actualSqrtMul2MinusRemainderGeU128 := readFrom(vm, VM.ExecutionSegment, 5)

	require.Equal(t, expectedSqrt0, actualSqrt0)
	require.Equal(t, expectedSqrt1, actualSqrt1)
	require.Equal(t, expectedRemainderLow, actualRemainderLow)
	require.Equal(t, expectedRemainderHigh, actualRemainderHigh)
	require.Equal(t, expectedSqrtMul2MinusRemainderGeU128, actualSqrtMul2MinusRemainderGeU128)
}

func TestUint256SquareRootHigh(t *testing.T) {
	vm := defaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	var sqrt0 ApCellRef = 1
	var sqrt1 ApCellRef = 2
	var remainderLow ApCellRef = 3
	var remainderHigh ApCellRef = 4
	var sqrtMul2MinusRemainderGeU128 ApCellRef = 5

	valueLow := Immediate(*big.NewInt(0))
	valueHigh := Immediate(*big.NewInt(1 << 8))

	hint := Uint256SquareRoot{
		valueLow:                     valueLow,
		valueHigh:                    valueHigh,
		sqrt0:                        sqrt0,
		sqrt1:                        sqrt1,
		remainderLow:                 remainderLow,
		remainderHigh:                remainderHigh,
		sqrtMul2MinusRemainderGeU128: sqrtMul2MinusRemainderGeU128,
	}

	err := hint.Execute(vm)

	require.NoError(t, err)

	expectedSqrt0 := memory.MemoryValueFromInt(0)
	expectedSqrt1 := memory.MemoryValueFromInt(16)
	expectedRemainderLow := memory.MemoryValueFromInt(0)
	expectedRemainderHigh := memory.MemoryValueFromInt(0)
	expectedSqrtMul2MinusRemainderGeU128 := memory.MemoryValueFromInt(0)

	actualSqrt0 := readFrom(vm, VM.ExecutionSegment, 1)
	actualSqrt1 := readFrom(vm, VM.ExecutionSegment, 2)
	actualRemainderLow := readFrom(vm, VM.ExecutionSegment, 3)
	actualRemainderHigh := readFrom(vm, VM.ExecutionSegment, 4)
	actualSqrtMul2MinusRemainderGeU128 := readFrom(vm, VM.ExecutionSegment, 5)

	require.Equal(t, expectedSqrt0, actualSqrt0)
	require.Equal(t, expectedSqrt1, actualSqrt1)
	require.Equal(t, expectedRemainderLow, actualRemainderLow)
	require.Equal(t, expectedRemainderHigh, actualRemainderHigh)
	require.Equal(t, expectedSqrtMul2MinusRemainderGeU128, actualSqrtMul2MinusRemainderGeU128)
}

func TestUint256SquareRoot(t *testing.T) {
	vm := defaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	var sqrt0 ApCellRef = 1
	var sqrt1 ApCellRef = 2
	var remainderLow ApCellRef = 3
	var remainderHigh ApCellRef = 4
	var sqrtMul2MinusRemainderGeU128 ApCellRef = 5

	valueLow := Immediate(*big.NewInt(51))
	valueHigh := Immediate(*big.NewInt(1024))

	hint := Uint256SquareRoot{
		valueLow:                     valueLow,
		valueHigh:                    valueHigh,
		sqrt0:                        sqrt0,
		sqrt1:                        sqrt1,
		remainderLow:                 remainderLow,
		remainderHigh:                remainderHigh,
		sqrtMul2MinusRemainderGeU128: sqrtMul2MinusRemainderGeU128,
	}

	err := hint.Execute(vm)

	require.NoError(t, err)

	expectedSqrt0 := memory.MemoryValueFromInt(0)
	expectedSqrt1 := memory.MemoryValueFromInt(32)
	expectedRemainderLow := memory.MemoryValueFromInt(51)
	expectedRemainderHigh := memory.MemoryValueFromInt(0)
	expectedSqrtMul2MinusRemainderGeU128 := memory.MemoryValueFromInt(0)

	actualSqrt0 := readFrom(vm, VM.ExecutionSegment, 1)
	actualSqrt1 := readFrom(vm, VM.ExecutionSegment, 2)
	actualRemainderLow := readFrom(vm, VM.ExecutionSegment, 3)
	actualRemainderHigh := readFrom(vm, VM.ExecutionSegment, 4)
	actualSqrtMul2MinusRemainderGeU128 := readFrom(vm, VM.ExecutionSegment, 5)

	require.Equal(t, expectedSqrt0, actualSqrt0)
	require.Equal(t, expectedSqrt1, actualSqrt1)
	require.Equal(t, expectedRemainderLow, actualRemainderLow)
	require.Equal(t, expectedRemainderHigh, actualRemainderHigh)
	require.Equal(t, expectedSqrtMul2MinusRemainderGeU128, actualSqrtMul2MinusRemainderGeU128)
}
