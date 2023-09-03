package hintrunner

import (
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

type Hinter interface {
	Execute(vm *VM.VirtualMachine) *HintError
}

const allocSegmentName = "AllocSegment"

type AllocSegment struct {
	dst CellRefer
}

func (hint AllocSegment) Execute(vm *VM.VirtualMachine) *HintError {
	segmentIndex := vm.MemoryManager.Memory.AllocateEmptySegment()
	memAddress := memory.MemoryValueFromSegmentAndOffset(segmentIndex, 0)

	cell, err := hint.dst.Get(vm)
	if err != nil {
		return NewHintError(allocSegmentName, err)
	}

	err = cell.Write(memAddress)
	if err != nil {
		return NewHintError(allocSegmentName, err)
	}

	return nil
}

const testLessThanName = "TestLessThan"

type TestLessThan struct {
	dst CellRefer
	lhs ResOperander
	rhs ResOperander
}

func (hint TestLessThan) Execute(vm *VM.VirtualMachine) *HintError {
	lhsVal, err := hint.lhs.Resolve(vm)
	if err != nil {
		return NewHintError(testLessThanName, err)
	}

	rhsVal, err := hint.rhs.Resolve(vm)
	if err != nil {
		return NewHintError(testLessThanName, err)
	}

	lhsFelt, err := lhsVal.ToFieldElement()
	if err != nil {
		return NewHintError(testLessThanName, err)
	}

	rhsFelt, err := rhsVal.ToFieldElement()
	if err != nil {
		return NewHintError(testLessThanName, err)
	}

	resFelt := f.Element{}
	if lhsFelt.Cmp(rhsFelt) <= 0 {
		resFelt.SetOne()
	}

	dstCell, err := hint.dst.Get(vm)
	if err != nil {
		return NewHintError(testLessThanName, err)
	}

	err = dstCell.Write(memory.MemoryValueFromFieldElement(&resFelt))
	if err != nil {
		return NewHintError(testLessThanName, err)
	}

	return nil
}
