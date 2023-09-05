package hintrunner

import (
	"testing"

	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/stretchr/testify/require"
)

func TestExistingHint(t *testing.T) {
	vm := defaultVirtualMachine()
	vm.Context.Ap = 3

	var ap ApCellRef = 5
	allocHint := AllocSegment{ap}

	hr := NewHintRunner(map[uint64]Hinter{
		10: allocHint,
	})

	vm.Context.Pc = 10
	err := hr.RunHint(vm)
	require.Nil(t, err)
	require.Equal(
		t,
		memory.MemoryValueFromSegmentAndOffset(2, 0),
		readFrom(vm, VM.ExecutionSegment, vm.Context.Ap+5),
	)
}

func TestNoHint(t *testing.T) {
	vm := defaultVirtualMachine()
	vm.Context.Ap = 3

	var ap ApCellRef = 5
	allocHint := AllocSegment{ap}

	hr := NewHintRunner(map[uint64]Hinter{
		10: allocHint,
	})

	vm.Context.Pc = 100
	err := hr.RunHint(vm)
	require.Nil(t, err)
	require.Equal(t, 2, len(vm.MemoryManager.Memory.Segments))
}
