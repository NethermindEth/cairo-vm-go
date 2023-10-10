package hintrunner

import (
	"math/big"
	"testing"

	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/stretchr/testify/assert"
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

func TestAllocSegmentOfSize(t *testing.T) {
	vm, _ := defaultVirtualMachine()
	vm.Context.Ap = 3
	vm.Context.Fp = 0

	const desiredSize = 200

	var ap ApCellRef = 5
	var fp FpCellRef = 9

	alloc1 := AllocSegmentOfSize{dst: ap, size: desiredSize}
	alloc2 := AllocSegmentOfSize{dst: fp, size: desiredSize}

	err := alloc1.Execute(vm)
	require.Nil(t, err)

	assert.Equal(t, 3, len(vm.Memory.Segments), "A new segment should be added to memory after alloc1")
	assert.Equal(t, desiredSize, cap(vm.Memory.Segments[2].Data), "The segment's capacity should match the desired size")

	assert.Equal(
		t,
		memory.MemoryValueFromSegmentAndOffset(2, 0),
		readFrom(vm, VM.ExecutionSegment, vm.Context.Ap+5),
		"Memory address should be as expected after alloc1",
	)

	err = alloc2.Execute(vm)
	require.Nil(t, err)

	assert.Equal(t, 4, len(vm.Memory.Segments), "A new segment should be added to memory after alloc2")
	assert.Equal(t, desiredSize, cap(vm.Memory.Segments[3].Data), "The segment's capacity should match the desired size")

	assert.Equal(
		t,
		memory.MemoryValueFromSegmentAndOffset(3, 0),
		readFrom(vm, VM.ExecutionSegment, vm.Context.Fp+9),
		"Memory address should be as expected after alloc2",
	)
}

func TestTestLessThanFalse(t *testing.T) {
	vm, _ := defaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0
	writeTo(vm, VM.ExecutionSegment, 0, memory.MemoryValueFromInt(17))

	var dst ApCellRef = 1

	lhs := Immediate(*big.NewInt(32))

	var rhsRef FpCellRef = 0
	rhs := Deref{rhsRef}

	hint := TestLessThan{
		dst: dst,
		lhs: lhs,
		rhs: rhs,
	}

	err := hint.Execute(vm)
	require.Nil(t, err)
	require.Equal(
		t,
		memory.EmptyMemoryValueAsFelt(),
		readFrom(vm, VM.ExecutionSegment, 1),
	)
}

func TestTestLessThanTrue(t *testing.T) {
	vm, _ := defaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0
	writeTo(vm, VM.ExecutionSegment, 0, memory.MemoryValueFromInt(23))

	var dst ApCellRef = 1

	lhs := Immediate(*big.NewInt(13))

	var rhsRef FpCellRef = 0
	rhs := Deref{rhsRef}

	hint := TestLessThan{
		dst: dst,
		lhs: lhs,
		rhs: rhs,
	}

	err := hint.Execute(vm)
	require.Nil(t, err)
	require.Equal(
		t,
		memory.MemoryValueFromInt(1),
		readFrom(vm, VM.ExecutionSegment, 1),
	)
}
