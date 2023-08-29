package hintrunner

import (
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
)

type Hinter interface {
	Execute(vm *VM.VirtualMachine) *HintError
}

type AllocSegment struct {
	dst CellRefer
}

func (hint AllocSegment) Execute(vm *VM.VirtualMachine) *HintError {
	segmentIndex := vm.MemoryManager.Memory.AllocateEmptySegment()
	memAddress := memory.MemoryValueFromSegmentAndOffset(segmentIndex, 0)

	cell, err := hint.dst.Get(vm)
	if err != nil {
		return NewHintError("AllocSegment", err)
	}

	err = cell.Write(memAddress)
	if err != nil {
		return NewHintError("AllocSegment", err)
	}

	return nil
}
