package hintrunner

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/core"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/utils"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/stretchr/testify/require"
)

func TestExistingHint(t *testing.T) {
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 3

	var ap hinter.ApCellRef = 5
	allocHint := core.AllocSegment{Dst: ap}

	hr := NewHintRunner(map[uint64][]hinter.Hinter{
		10: {&allocHint},
	}, nil)

	vm.Context.Pc = memory.MemoryAddress{
		SegmentIndex: 0,
		Offset:       10,
	}
	err := hr.RunHint(vm)
	require.Nil(t, err)
	require.Equal(
		t,
		memory.MemoryValueFromSegmentAndOffset(2, 0),
		utils.ReadFrom(vm, VM.ExecutionSegment, vm.Context.Ap+5),
	)
}

func TestNoHint(t *testing.T) {
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 3

	var ap hinter.ApCellRef = 5
	allocHint := core.AllocSegment{Dst: ap}

	hr := NewHintRunner(map[uint64][]hinter.Hinter{
		10: {&allocHint},
	}, nil)

	vm.Context.Pc = memory.MemoryAddress{
		SegmentIndex: 0,
		Offset:       100,
	}
	err := hr.RunHint(vm)
	require.Nil(t, err)
	require.Equal(t, 2, len(vm.Memory.Segments))
}
