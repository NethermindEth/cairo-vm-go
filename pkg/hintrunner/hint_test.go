package hintrunner

import (
	"math/big"
	"testing"

	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/stretchr/testify/require"
)

func TestAllocSegment(t *testing.T) {
	vm, _ := defaultVirtualMachine()
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
	vm, _ := defaultVirtualMachine()
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
			vm, _ := defaultVirtualMachine()
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
			vm, _ := defaultVirtualMachine()
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
	vm, _ := defaultVirtualMachine()
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

func TestLinearSplit(t *testing.T) {
	vm, _ := defaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	value := Immediate(*big.NewInt(42*223344 + 14))
	scalar := Immediate(*big.NewInt(42))
	max_x := Immediate(*big.NewInt(9999999999))
	var x ApCellRef = 0
	var y ApCellRef = 1

	hint := LinearSplit{
		value:  value,
		scalar: scalar,
		max_x:  max_x,
		x:      x,
		y:      y,
	}

	err := hint.Execute(vm)
	require.NoError(t, err)
	xx := readFrom(vm, VM.ExecutionSegment, 0)
	require.Equal(t, xx, memory.MemoryValueFromInt(223344))
	yy := readFrom(vm, VM.ExecutionSegment, 1)
	require.Equal(t, yy, memory.MemoryValueFromInt(14))

	vm, _ = defaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	//Lower max_x
	max_x = Immediate(*big.NewInt(223343))
	hint = LinearSplit{
		value:  value,
		scalar: scalar,
		max_x:  max_x,
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
