package hintrunner

import (
	"fmt"
	"math/big"

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
	u128Max := big.NewInt(1).Lsh(big.NewInt(1), 128)
	u128Max = big.NewInt(1).Sub(u128Max, big.NewInt(1))
	u128MaxFelt := &f.Element{}
	u128MaxFelt.SetBigInt(u128Max)

	lsh, err := hint.lhs.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve lhs operand %s: %v", hint.lhs, err)
	}

	rhs, err := hint.rhs.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve rhs operand %s: %v", hint.rhs, err)
	}

	lhsFelt, err := lsh.ToFieldElement()
	if err != nil {
		return err
	}
	if lhsFelt.Cmp(u128MaxFelt) > 0 {
		return fmt.Errorf("lhs operand %s should be u128", lhsFelt)
	}

	rhsFelt, err := rhs.ToFieldElement()
	if err != nil {
		return err
	}
	if rhsFelt.Cmp(u128MaxFelt) > 0 {
		return fmt.Errorf("lhs operand %s should be u128", rhsFelt)
	}

	mul := big.NewInt(1).Mul(lhsFelt.BigInt(big.NewInt(1)), rhsFelt.BigInt(big.NewInt(1)))

	low := big.NewInt(1)
	high := big.NewInt(1)
	low.And(mul, u128Max)
	high.Rsh(mul, 128)

	lowFelt := &f.Element{}
	lowFelt.SetBigInt(low)
	highFelt := &f.Element{}
	highFelt.SetBigInt(high)

	lowCell, err := hint.low.Get(vm)
	if err != nil {
		return fmt.Errorf("get destination cell: %v", err)
	}
	err = lowCell.Write(memory.MemoryValueFromFieldElement(lowFelt))
	if err != nil {
		return fmt.Errorf("write cell: %v", err)
	}

	highCell, err := hint.high.Get(vm)
	if err != nil {
		return fmt.Errorf("get destination cell: %v", err)
	}
	err = highCell.Write(memory.MemoryValueFromFieldElement(highFelt))
	if err != nil {
		return fmt.Errorf("write cell: %v", err)
	}

	return nil
}
