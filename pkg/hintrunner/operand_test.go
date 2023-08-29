package hintrunner

import (
	"testing"

	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
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
	// todo(rodro): do this with a negative offset, when safemath merged
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
	// todo(rodro): do this with a negative offset, when safemath merged
}

func TestResolveImmediate(t *testing.T) {

}

func TestResolveAddOp(t *testing.T) {

}

func TestResolveMulOp(t *testing.T) {

}

func defaultVirtualMachine() *VM.VirtualMachine {
	vm, _ := VM.NewVirtualMachine(make([]*f.Element, 0), VM.VirtualMachineConfig{})
	return vm
}

func writeTo(vm *VM.VirtualMachine, segment uint64, offset uint64, val *memory.MemoryValue) {
	_ = vm.MemoryManager.Memory.Write(segment, offset, val)
}
