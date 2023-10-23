package hintrunner

import (
	"fmt"

	"github.com/holiman/uint256"

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
	segmentIndex := vm.Memory.AllocateEmptySegment()
	memAddress := memory.MemoryValueFromSegmentAndOffset(segmentIndex, 0)

	regAddr, err := hint.dst.Get(vm)
	if err != nil {
		return fmt.Errorf("get register %s: %w", hint.dst, err)
	}

	err = vm.Memory.WriteToAddress(&regAddr, &memAddress)
	if err != nil {
		return fmt.Errorf("write to address %s: %v", regAddr, err)
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
		return fmt.Errorf("resolve lhs operand %s: %w", hint.lhs, err)
	}

	rhsVal, err := hint.rhs.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve rhs operand %s: %w", hint.rhs, err)
	}

	lhsFelt, err := lhsVal.FieldElement()
	if err != nil {
		return err
	}

	rhsFelt, err := rhsVal.FieldElement()
	if err != nil {
		return err
	}

	resFelt := f.Element{}
	if lhsFelt.Cmp(rhsFelt) < 0 {
		resFelt.SetOne()
	}

	dstAddr, err := hint.dst.Get(vm)
	if err != nil {
		return fmt.Errorf("get dst address %s: %w", dstAddr, err)
	}

	mv := memory.MemoryValueFromFieldElement(&resFelt)
	err = vm.Memory.WriteToAddress(&dstAddr, &mv)
	if err != nil {
		return fmt.Errorf("write to dst address %s: %w", dstAddr, err)
	}

	return nil
}

type TestLessThanOrEqual struct {
	dst CellRefer
	lhs ResOperander
	rhs ResOperander
}

func (hint TestLessThanOrEqual) String() string {
	return "TestLessThanOrEqual"
}

func (hint TestLessThanOrEqual) Execute(vm *VM.VirtualMachine) error {
	lhsVal, err := hint.lhs.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve lhs operand %s: %w", hint.lhs, err)
	}

	rhsVal, err := hint.rhs.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve rhs operand %s: %w", hint.rhs, err)
	}

	lhsFelt, err := lhsVal.FieldElement()
	if err != nil {
		return err
	}

	rhsFelt, err := rhsVal.FieldElement()
	if err != nil {
		return err
	}

	resFelt := f.Element{}
	if lhsFelt.Cmp(rhsFelt) <= 0 {
		resFelt.SetOne()
	}

	dstAddr, err := hint.dst.Get(vm)
	if err != nil {
		return fmt.Errorf("get dst address %s: %w", dstAddr, err)
	}

	mv := memory.MemoryValueFromFieldElement(&resFelt)
	err = vm.Memory.WriteToAddress(&dstAddr, &mv)
	if err != nil {
		return fmt.Errorf("write to dst address %s: %w", dstAddr, err)
	}

	return nil
}

type LinearSplit struct {
	value  ResOperander
	scalar ResOperander
	maxX   ResOperander
	x      CellRefer
	y      CellRefer
}

func (hint LinearSplit) Execute(vm *VM.VirtualMachine) error {
	value, err := hint.value.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve value operand %s: %w", hint.value, err)
	}
	valueField, err := value.FieldElement()
	if err != nil {
		return err
	}
	scalar, err := hint.scalar.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve scalar operand %s: %w", hint.scalar, err)
	}
	scalarField, err := scalar.FieldElement()
	if err != nil {
		return err
	}

	maxX, err := hint.maxX.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve max_x operand %s: %w", hint.maxX, err)
	}
	maxXField, err := maxX.FieldElement()
	if err != nil {
		return err
	}

	scalarBytes := scalarField.Bytes()
	valueBytes := valueField.Bytes()
	maxXBytes := maxXField.Bytes()
	scalarUint := new(uint256.Int).SetBytes(scalarBytes[:])
	valueUint := new(uint256.Int).SetBytes(valueBytes[:])
	maxXUint := new(uint256.Int).SetBytes(maxXBytes[:])

	x := (&uint256.Int{}).Div(valueUint, scalarUint)

	if x.Cmp(maxXUint) > 0 {
		x.Set(maxXUint)
	}

	y := &uint256.Int{}
	y = y.Sub(valueUint, y.Mul(scalarUint, x))

	xAddr, err := hint.x.Get(vm)
	if err != nil {
		return fmt.Errorf("get x address %s: %w", xAddr, err)
	}

	yAddr, err := hint.y.Get(vm)
	if err != nil {
		return fmt.Errorf("get y address %s: %w", yAddr, err)
	}

	xFiled := &f.Element{}
	yFiled := &f.Element{}
	xFiled.SetBytes(x.Bytes())
	yFiled.SetBytes(y.Bytes())
	mv := memory.MemoryValueFromFieldElement(xFiled)
	err = vm.Memory.WriteToAddress(&xAddr, &mv)
	if err != nil {
		return fmt.Errorf("write to x address %s: %w", xAddr, err)
	}

	mv = memory.MemoryValueFromFieldElement(yFiled)
	err = vm.Memory.WriteToAddress(&yAddr, &mv)
	if err != nil {
		return fmt.Errorf("write to y address %s: %w", yAddr, err)

	}
	return nil
}

type WideMul128 struct {
	lhs  ResOperander
	rhs  ResOperander
	high CellRefer
	low  CellRefer
}

func (hint WideMul128) String() string {
	return "WideMul128"
}

func (hint WideMul128) Execute(vm *VM.VirtualMachine) error {
	mask := MaxU128()

	lhs, err := hint.lhs.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve lhs operand %s: %v", hint.lhs, err)
	}
	rhs, err := hint.rhs.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve rhs operand %s: %v", hint.rhs, err)
	}

	lhsFelt, err := lhs.FieldElement()
	if err != nil {
		return err
	}
	rhsFelt, err := rhs.FieldElement()
	if err != nil {
		return err
	}

	lhsU256 := uint256.Int(lhsFelt.Bits())
	rhsU256 := uint256.Int(rhsFelt.Bits())

	if lhsU256.Gt(&mask) {
		return fmt.Errorf("lhs operand %s should be u128", lhsFelt)
	}
	if rhsU256.Gt(&mask) {
		return fmt.Errorf("rhs operand %s should be u128", rhsFelt)
	}

	mul := lhsU256.Mul(&lhsU256, &rhsU256)

	bytes := mul.Bytes32()

	low := f.Element{}
	low.SetBytes(bytes[16:])

	high := f.Element{}
	high.SetBytes(bytes[:16])

	lowAddr, err := hint.low.Get(vm)
	if err != nil {
		return fmt.Errorf("get destination cell: %v", err)
	}
	mvLow := memory.MemoryValueFromFieldElement(&low)
	err = vm.Memory.WriteToAddress(&lowAddr, &mvLow)
	if err != nil {
		return fmt.Errorf("write cell: %v", err)
	}

	highAddr, err := hint.high.Get(vm)
	if err != nil {
		return fmt.Errorf("get destination cell: %v", err)
	}
	mvHigh := memory.MemoryValueFromFieldElement(&high)
	err = vm.Memory.WriteToAddress(&highAddr, &mvHigh)
	if err != nil {
		return fmt.Errorf("write cell: %v", err)
	}
	return nil
}
