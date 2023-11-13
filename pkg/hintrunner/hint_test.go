package hintrunner

import (
	"io"
	"math/big"
	"os"
	"testing"

	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/holiman/uint256"
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

	err := alloc1.Execute(vm, nil)
	require.Nil(t, err)
	require.Equal(t, 3, len(vm.Memory.Segments))
	require.Equal(
		t,
		memory.MemoryValueFromSegmentAndOffset(2, 0),
		readFrom(vm, VM.ExecutionSegment, vm.Context.Ap+5),
	)

	err = alloc2.Execute(vm, nil)
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

	lhs := Immediate(f.NewElement(13))

	hint := TestLessThan{
		dst: dst,
		lhs: lhs,
		rhs: rhs,
	}

	err := hint.Execute(vm, nil)
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
		lhsValue    f.Element
		expectedMsg string
	}{
		{f.NewElement(32), "Expected the hint to evaluate to False when lhs is larger"},
		{f.NewElement(17), "Expected the hint to evaluate to False when values are equal"},
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

			lhs := Immediate(tc.lhsValue)
			hint := TestLessThan{
				dst: dst,
				lhs: lhs,
				rhs: rhs,
			}

			err := hint.Execute(vm, nil)
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
		lhsValue    f.Element
		expectedMsg string
	}{
		{f.NewElement(13), "Expected the hint to evaluate to True when lhs is less than rhs"},
		{f.NewElement(23), "Expected the hint to evaluate to True when values are equal"},
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

			lhs := Immediate(tc.lhsValue)
			hint := TestLessThanOrEqual{
				dst: dst,
				lhs: lhs,
				rhs: rhs,
			}

			err := hint.Execute(vm, nil)
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

	lhs := Immediate(f.NewElement(32))

	hint := TestLessThanOrEqual{
		dst: dst,
		lhs: lhs,
		rhs: rhs,
	}

	err := hint.Execute(vm, nil)
	require.NoError(t, err)
	require.Equal(
		t,
		memory.EmptyMemoryValueAsFelt(),
		readFrom(vm, VM.ExecutionSegment, 1),
		"Expected the hint to evaluate to False when lhs is larger",
	)
}

func TestLinearSplit(t *testing.T) {
	vm := defaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	value := Immediate(f.NewElement(42*223344 + 14))
	scalar := Immediate(f.NewElement(42))
	maxX := Immediate(f.NewElement(9999999999))
	var x ApCellRef = 0
	var y ApCellRef = 1

	hint := LinearSplit{
		value:  value,
		scalar: scalar,
		maxX:   maxX,
		x:      x,
		y:      y,
	}

	err := hint.Execute(vm)
	require.NoError(t, err)
	xx := readFrom(vm, VM.ExecutionSegment, 0)
	require.Equal(t, xx, memory.MemoryValueFromInt(223344))
	yy := readFrom(vm, VM.ExecutionSegment, 1)
	require.Equal(t, yy, memory.MemoryValueFromInt(14))

	vm = defaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	//Lower max_x
	maxX = Immediate(f.NewElement(223343))
	hint = LinearSplit{
		value:  value,
		scalar: scalar,
		maxX:   maxX,
		x:      x,
		y:      y,
	}

	err = hint.Execute(vm)
	require.NoError(t, err)
	xx = readFrom(vm, VM.ExecutionSegment, 0)
	require.Equal(t, xx, memory.MemoryValueFromInt(223343))
	yy = readFrom(vm, VM.ExecutionSegment, 1)
	require.Equal(t, yy, memory.MemoryValueFromInt(14+42))
}

func TestWideMul128(t *testing.T) {
	vm := defaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	var dstLow ApCellRef = 1
	var dstHigh ApCellRef = 2

	lhsBytes := new(uint256.Int).Lsh(uint256.NewInt(1), 127).Bytes32()
	lhsFelt, err := f.BigEndian.Element(&lhsBytes)
	require.NoError(t, err)

	rhsFelt := f.NewElement(1<<8 + 1)

	lhs := Immediate(lhsFelt)
	rhs := Immediate(rhsFelt)

	hint := WideMul128{
		low:  dstLow,
		high: dstHigh,
		lhs:  lhs,
		rhs:  rhs,
	}

	err = hint.Execute(vm, nil)
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

	lhsBytes := new(uint256.Int).Lsh(uint256.NewInt(1), 128).Bytes32()
	lhsFelt, err := f.BigEndian.Element(&lhsBytes)
	require.NoError(t, err)

	lhs := Immediate(lhsFelt)
	rhs := Immediate(f.NewElement(1))

	hint := WideMul128{
		low:  dstLow,
		high: dstHigh,
		lhs:  lhs,
		rhs:  rhs,
	}

	err = hint.Execute(vm, nil)
	require.ErrorContains(t, err, "should be u128")
}

func TestDebugPrint(t *testing.T) {
	//Save the old stdout
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	vm := defaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	writeTo(vm, VM.ExecutionSegment, 0, memory.MemoryValueFromSegmentAndOffset(VM.ExecutionSegment, 2))
	writeTo(vm, VM.ExecutionSegment, 1, memory.MemoryValueFromSegmentAndOffset(VM.ExecutionSegment, 5))
	writeTo(vm, VM.ExecutionSegment, 2, memory.MemoryValueFromInt(10))
	writeTo(vm, VM.ExecutionSegment, 3, memory.MemoryValueFromInt(20))
	writeTo(vm, VM.ExecutionSegment, 4, memory.MemoryValueFromInt(30))

	var starRef ApCellRef = 0
	var endRef ApCellRef = 1
	start := Deref{starRef}
	end := Deref{endRef}
	hint := DebugPrint{
		start: start,
		end:   end,
	}
	expected := []byte("[DEBUG] a\n[DEBUG] 14\n[DEBUG] 1e\n")
	err := hint.Execute(vm)

	w.Close()
	out, _ := io.ReadAll(r)
	//Restore stdout at the end of the test
	os.Stdout = rescueStdout

	require.NoError(t, err)
	require.Equal(t, expected, out)
}

func TestSquareRoot(t *testing.T) {
	vm := defaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0
	var dst ApCellRef = 1

	value := Immediate(f.NewElement(36))
	hint := SquareRoot{
		value: value,
		dst:   dst,
	}

	err := hint.Execute(vm, nil)

	require.NoError(t, err)
	require.Equal(
		t,
		memory.MemoryValueFromInt(6),
		readFrom(vm, VM.ExecutionSegment, 1),
	)

	dst = 2
	value = Immediate(f.NewElement(30))
	hint = SquareRoot{
		value: value,
		dst:   dst,
	}

	err = hint.Execute(vm, nil)

	require.NoError(t, err)
	require.Equal(
		t,
		memory.MemoryValueFromInt(5),
		readFrom(vm, VM.ExecutionSegment, 2),
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

	valueLow := Immediate(f.NewElement(121))
	valueHigh := Immediate(f.NewElement(0))

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

	valueLow := Immediate(f.NewElement(0))
	valueHigh := Immediate(f.NewElement(1 << 8))

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

	valueLow := Immediate(f.NewElement(51))
	valueHigh := Immediate(f.NewElement(1024))

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

func TestAssertLeFindSmallArc(t *testing.T) {
	vm := defaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0
	// vm.Memory.Segments = append(vm.Memory.Segments, memory.EmptySegment())

	ctx := HintRunnerContext{
		DictionaryManager:         DictionaryManager{},
		SquashedDictionaryManager: SquashedDictionaryManager{},
		ExcludedArc:               0,
	}

	valA := Immediate(f.NewElement(12))
	valB := Immediate(f.NewElement(13))

	rangeCheckPtr := Deref{ApCellRef(0)}

	hint := AssertLeFindSmallArc{
		a:             valA,
		b:             valB,
		rangeCheckPtr: rangeCheckPtr,
	}

	err := hint.Execute(vm, &ctx)

	require.NoError(t, err)

	expected1 := memory.MemoryValueFromInt(0)
	expected2 := memory.MemoryValueFromInt(0)
	expected3 := memory.MemoryValueFromInt(0)
	expected4 := memory.MemoryValueFromInt(0)

	actual1 := readFrom(vm, VM.ExecutionSegment, 1)
	actual2 := readFrom(vm, VM.ExecutionSegment, 2)
	actual3 := readFrom(vm, VM.ExecutionSegment, 3)
	actual4 := readFrom(vm, VM.ExecutionSegment, 4)

	require.Equal(t, expected1, actual1)
	require.Equal(t, expected2, actual2)
	require.Equal(t, expected3, actual3)
	require.Equal(t, expected4, actual4)
}
