package hintrunner

import (
	"fmt"
	"io"
	"math/big"
	"os"
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

func TestDebugPrint(t *testing.T) {

	//Save the old stdout
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	vm := defaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0
	v := memory.MemoryValueFromInt(10)
	vm.Memory.Write(VM.ExecutionSegment, 0, &v)
	v = memory.MemoryValueFromInt(20)
	vm.Memory.Write(VM.ExecutionSegment, 1, &v)
	v = memory.MemoryValueFromInt(30)
	vm.Memory.Write(VM.ExecutionSegment, 2, &v)

	start := Immediate(*big.NewInt(0))
	end := Immediate(*big.NewInt(2))
	hint := DebugPrint{
		start: start,
		end:   end,
	}
	err := hint.Execute(vm)

	w.Close()
	out, _ := io.ReadAll(r)
	//Restore stdout at the end of the test
	os.Stdout = rescueStdout

	fmt.Println(out)
	require.NoError(t, err)
}
