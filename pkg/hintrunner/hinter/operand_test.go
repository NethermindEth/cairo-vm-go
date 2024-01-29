package hinter

import (
	"testing"

	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/utils"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/require"
)

func TestGetAp(t *testing.T) {
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 5
	utils.WriteTo(vm, VM.ExecutionSegment, vm.Context.Ap+7, memory.MemoryValueFromInt(11))

	var apReg ApCellRef = 7
	apAddr, err := apReg.Get(vm)

	require.NoError(t, err)

	value, err := vm.Memory.ReadFromAddress(&apAddr)
	require.NoError(t, err)

	require.Equal(t, memory.MemoryValueFromInt(11), value)
}

func TestGetFp(t *testing.T) {
	vm := VM.DefaultVirtualMachine()
	vm.Context.Fp = 15
	utils.WriteTo(vm, VM.ExecutionSegment, vm.Context.Fp-7, memory.MemoryValueFromInt(11))

	var fpReg FpCellRef = -7
	fpAddr, err := fpReg.Get(vm)
	require.NoError(t, err)

	value, err := vm.Memory.ReadFromAddress(&fpAddr)
	require.NoError(t, err)

	require.Equal(t, memory.MemoryValueFromInt(11), value)
}

func TestResolveDeref(t *testing.T) {
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 5
	utils.WriteTo(vm, VM.ExecutionSegment, vm.Context.Ap+7, memory.MemoryValueFromInt(11))

	var apCell ApCellRef = 7
	deref := Deref{apCell}

	value, err := deref.Resolve(vm)
	require.NoError(t, err)

	require.Equal(t, memory.MemoryValueFromInt(11), value)
}

func TestResolveDoubleDerefPositiveOffset(t *testing.T) {
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 5
	utils.WriteTo(
		vm,
		VM.ExecutionSegment, vm.Context.Ap+7,
		memory.MemoryValueFromSegmentAndOffset(VM.ExecutionSegment, 0),
	)
	utils.WriteTo(
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
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 5
	utils.WriteTo(
		vm,
		VM.ExecutionSegment, vm.Context.Ap+7,
		memory.MemoryValueFromSegmentAndOffset(VM.ExecutionSegment, 20),
	)
	utils.WriteTo(
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

	imm := Immediate(f.NewElement(99))

	solved, err := imm.Resolve(vm)
	require.NoError(t, err)
	require.Equal(t, memory.MemoryValueFromInt(99), solved)
}

func TestResolveAddOp(t *testing.T) {
	vm := VM.DefaultVirtualMachine()
	// Set the information used by the lhs
	vm.Context.Fp = 0
	vm.Context.Ap = 5
	utils.WriteTo(
		vm,
		VM.ExecutionSegment, vm.Context.Ap+7,
		memory.MemoryValueFromSegmentAndOffset(4, 29),
	)
	// Set the information used by the rhs
	utils.WriteTo(
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
		Operator: operator,
		Lhs:      ap,
		Rhs:      deref,
	}

	res, err := bop.Resolve(vm)
	require.NoError(t, err)
	require.Equal(t, memory.MemoryValueFromSegmentAndOffset(4, 59), res)
}

func TestResolveMulOp(t *testing.T) {
	vm := VM.DefaultVirtualMachine()
	// Set the information used by the lhs
	vm.Context.Fp = 0
	vm.Context.Ap = 5
	utils.WriteTo(
		vm,
		VM.ExecutionSegment, vm.Context.Ap+7,
		memory.MemoryValueFromInt(100),
	)
	// Set the information used by the rhs
	utils.WriteTo(
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
		Operator: operator,
		Lhs:      ap,
		Rhs:      deref,
	}

	res, err := bop.Resolve(vm)
	require.NoError(t, err)
	require.Equal(t, memory.MemoryValueFromInt(500), res)

}
