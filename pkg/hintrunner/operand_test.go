package hintrunner

import (
	"math/big"
	"testing"

	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/stretchr/testify/require"
)

func TestGetAp(t *testing.T) {
	vm := defaultVirtualMachine()
	vm.Context.Ap = 5
	writeTo(vm, VM.ExecutionSegment, vm.Context.Ap+7, memory.MemoryValueFromInt(11))

	var apCell ApCellRef = 7
	cell, err := apCell.Get(vm)

	require.NoError(t, err)

	value := cell.Read()
	require.Equal(t, memory.MemoryValueFromInt(11), value)
}

func TestGetFp(t *testing.T) {
	vm := defaultVirtualMachine()
	vm.Context.Fp = 15
	writeTo(vm, VM.ExecutionSegment, vm.Context.Fp-7, memory.MemoryValueFromInt(11))

	var fpCell FpCellRef = -7
	cell, err := fpCell.Get(vm)

	require.NoError(t, err)

	value := cell.Read()
	require.Equal(t, memory.MemoryValueFromInt(11), value)
}

func TestResolveDeref(t *testing.T) {
	vm := defaultVirtualMachine()
	vm.Context.Ap = 5
	writeTo(vm, VM.ExecutionSegment, vm.Context.Ap+7, memory.MemoryValueFromInt(11))

	var apCell ApCellRef = 7
	deref := Deref{apCell}

	value, err := deref.Resolve(vm)

	require.NoError(t, err)
	require.Equal(t, memory.MemoryValueFromInt(11), value)
}

func TestResolveDoubleDerefPositiveOffset(t *testing.T) {
	vm := defaultVirtualMachine()
	vm.Context.Ap = 5
	writeTo(
		vm,
		VM.ExecutionSegment, vm.Context.Ap+7,
		memory.MemoryValueFromSegmentAndOffset(VM.ExecutionSegment, 0),
	)
	writeTo(
		vm,
		VM.ExecutionSegment, 14,
		memory.MemoryValueFromInt(13),
	)

	var apCell ApCellRef = 7
	dderf := DoubleDeref{apCell, 14}

	value, err := dderf.Resolve(vm)
	require.NoError(t, err)
	require.Equal(t, memory.MemoryValueFromInt(13), value)
}

func TestResolveDoubleDerefNegativeOffset(t *testing.T) {
	vm := defaultVirtualMachine()
	vm.Context.Ap = 5
	writeTo(
		vm,
		VM.ExecutionSegment, vm.Context.Ap+7,
		memory.MemoryValueFromSegmentAndOffset(VM.ExecutionSegment, 20),
	)
	writeTo(
		vm,
		VM.ExecutionSegment, 6,
		memory.MemoryValueFromInt(13),
	)

	var apCell ApCellRef = 7
	dderf := DoubleDeref{apCell, -14}

	value, err := dderf.Resolve(vm)
	require.NoError(t, err)
	require.Equal(t, memory.MemoryValueFromInt(13), value)
}

func TestResolveImmediate(t *testing.T) {
	// Immediate does not need the vm for resolving itself
	var vm *VM.VirtualMachine = nil

	imm := Immediate(*big.NewInt(99))

	solved, err := imm.Resolve(vm)
	require.NoError(t, err)
	require.Equal(t, memory.MemoryValueFromInt(99), solved)
}

func TestResolveAddOp(t *testing.T) {
	vm := defaultVirtualMachine()
	// Set the information used by the lhs
	vm.Context.Fp = 0
	vm.Context.Ap = 5
	writeTo(
		vm,
		VM.ExecutionSegment, vm.Context.Ap+7,
		memory.MemoryValueFromSegmentAndOffset(4, 29),
	)
	// Set the information used by the rhs
	writeTo(
		vm,
		VM.ExecutionSegment, vm.Context.Fp+20,
		memory.MemoryValueFromInt(30),
	)

	// lhs
	var ap ApCellRef = 7

	// Rhs
	var fp FpCellRef = 20
	deref := Deref{fp}

	operator := Add

	bop := BinaryOp{
		operator: operator,
		lhs:      ap,
		rhs:      deref,
	}

	res, err := bop.Resolve(vm)
	require.NoError(t, err)
	require.Equal(t, memory.MemoryValueFromSegmentAndOffset(4, 59), res)
}

func TestResolveMulOp(t *testing.T) {
	vm := defaultVirtualMachine()
	// Set the information used by the lhs
	vm.Context.Fp = 0
	vm.Context.Ap = 5
	writeTo(
		vm,
		VM.ExecutionSegment, vm.Context.Ap+7,
		memory.MemoryValueFromInt(100),
	)
	// Set the information used by the rhs
	writeTo(
		vm,
		VM.ExecutionSegment, vm.Context.Fp+20,
		memory.MemoryValueFromInt(5),
	)

	// lhs
	var ap ApCellRef = 7

	// Rhs
	var fp FpCellRef = 20
	deref := Deref{fp}

	operator := Mul

	bop := BinaryOp{
		operator: operator,
		lhs:      ap,
		rhs:      deref,
	}

	res, err := bop.Resolve(vm)
	require.NoError(t, err)
	require.Equal(t, memory.MemoryValueFromInt(500), res)

}
