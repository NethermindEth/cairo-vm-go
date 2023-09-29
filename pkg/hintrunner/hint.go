package hintrunner

import (
	"fmt"

	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

type Hinter interface {
	fmt.Stringer

	Execute(vm *VM.VirtualMachine) error
}

type AllocSegment struct {
	dst CellRefer
}

func (hint AllocSegment) String() string {
	return "AllocSegment"
}

func (hint AllocSegment) Execute(vm *VM.VirtualMachine) error {
	segmentIndex := vm.MemoryManager.Memory.AllocateEmptySegment()
	memAddress := memory.MemoryValueFromSegmentAndOffset(segmentIndex, 0)

	cell, err := hint.dst.Get(vm)
	if err != nil {
		return fmt.Errorf("get destination cell: %v", err)
	}

	err = cell.Write(memAddress)
	if err != nil {
		return fmt.Errorf("write cell: %v", err)
	}

	return nil
}

type TestLessThan struct {
	dst CellRefer
	lhs ResOperander
	rhs ResOperander
}

func (hint TestLessThan) String() string {
	return "TestLessThan"
}

func (hint TestLessThan) Execute(vm *VM.VirtualMachine) error {
	lhsVal, err := hint.lhs.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve lhs operand %s: %v", hint.lhs, err)
	}

	rhsVal, err := hint.rhs.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve rhs operand %s: %v", hint.rhs, err)
	}

	lhsFelt, err := lhsVal.ToFieldElement()
	if err != nil {
		return err
	}

	rhsFelt, err := rhsVal.ToFieldElement()
	if err != nil {
		return err
	}

	resFelt := f.Element{}
	if lhsFelt.Cmp(rhsFelt) <= 0 {
		resFelt.SetOne()
	}

	dstCell, err := hint.dst.Get(vm)
	if err != nil {
		return fmt.Errorf("get destination cell: %v", err)
	}

	err = dstCell.Write(memory.MemoryValueFromFieldElement(&resFelt))
	if err != nil {
		return fmt.Errorf("write cell: %v", err)
	}

	return nil
}

type DivMod struct {
	lhs       ResOperander
	rhs       ResOperander
	quotient  CellRefer
	remainder CellRefer
}

func (hint DivMod) String() string {
	return "DivMod"
}

// Computes lhs/rhs and returns the quotient and remainder. Note: the hint may be used to write an already assigned memory cell
func (hint DivMod) Execute(vm *VM.VirtualMachine) error {
	lhsVal, err := hint.lhs.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve lhs operand %s: %v", hint.lhs, err)
	}

	rhsVal, err := hint.rhs.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve rhs operand %s: %v", hint.rhs, err)
	}

	lhsFelt, err := lhsVal.ToFieldElement()
	if err != nil {
		return err
	}

	rhsFelt, err := rhsVal.ToFieldElement()
	if err != nil {
		return err
	}

	if rhsFelt == new(f.Element).SetInt64(0) {
		return fmt.Errorf("Cannot divide by zero, rhs=%v", rhsFelt)
	}

	resFelt := f.Element{}
	resFelt.Div(lhsFelt, rhsFelt)

	// todo: how to get remainder??

	quoCell, err := hint.quotient.Get(vm)
	if err != nil {
		return fmt.Errorf("get quotient cell: %v", err)
	}

	err = quoCell.Write(memory.MemoryValueFromFieldElement(&resFelt))
	if err != nil {
		return fmt.Errorf("write cell: %v", err)
	}

	return nil
}
