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

func (hint DivMod) Execute(vm *VM.VirtualMachine) error {

}

type Uint256DivMod struct {
	Dividend0  ResOperander
	Dividend1  ResOperander
	Divisor0   ResOperander
	Divisor1   ResOperander
	Quotient0  CellRefer
	Quotient1  CellRefer
	Remainder0 CellRefer
	Remainder1 CellRefer
}

func (hint Uint256DivMod) String() string {
	return "Uint256DivMod"
}

func (hint Uint256DivMod) Execute(vm *VM.VirtualMachine) error {

	u128MaxFelt := MaxU128Felt()

	dividend0, err := hint.Dividend0.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve dividend0 operand %s: %v", hint.Dividend0, err)
	}

	dividend1, err := hint.Dividend1.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve dividend1 operand %s: %v", hint.Dividend1, err)
	}

	divisor0, err := hint.Divisor0.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve divisor0 operand %s: %v", hint.Divisor0, err)
	}

	divisor1, err := hint.Divisor1.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve divisor1 operand %s: %v", hint.Divisor1, err)
	}

	dividend0Felt, err := dividend0.ToFieldElement()
	if err != nil {
		return err
	}

	dividend1Felt, err := dividend1.ToFieldElement()
	if err != nil {
		return err
	}

	divisor0Felt, err := divisor0.ToFieldElement()
	if err != nil {
		return err
	}

	divisor1Felt, err := divisor1.ToFieldElement()
	if err != nil {
		return err
	}

	if dividend0Felt.Cmp(u128MaxFelt) > 0 {
		return fmt.Errorf("dividend0 operand %s should be u128", dividend0Felt)
	}

	if dividend1Felt.Cmp(u128MaxFelt) > 0 {
		return fmt.Errorf("dividend1 operand %s should be u128", dividend1Felt)
	}

	if divisor0Felt.Cmp(u128MaxFelt) > 0 {
		return fmt.Errorf("divisor0 operand %s should be u128", divisor0Felt)
	}

	if divisor1Felt.Cmp(u128MaxFelt) > 0 {
		return fmt.Errorf("divisor1 operand %s should be u128", divisor1Felt)
	}

	dividend := big.NewInt(1).Add(dividend0Felt.BigInt(big.NewInt(1)), dividend1Felt.BigInt(big.NewInt(1)))
	divisor := big.NewInt(1).Add(divisor0Felt.BigInt(big.NewInt(1)), divisor1Felt.BigInt(big.NewInt(1)))

	quotient, remainder := big.NewInt(1).DivMod(dividend, divisor, dividend)
	mask := MaxU128()

	quotientLow := big.NewInt(1)
	quotientHigh := big.NewInt(1)
	quotientLow.And(quotient, mask)
	quotientHigh.Rsh(quotient, 128)

	remainderLow := big.NewInt(1)
	remainderHigh := big.NewInt(1)
	remainderLow.And(remainder, mask)
	remainderHigh.Rsh(remainder, 128)

	quotientLowFelt := &f.Element{}
	quotientLowFelt.SetBigInt(quotientLow)
	quotientHighFelt := &f.Element{}
	quotientHighFelt.SetBigInt(quotientHigh)

	remainderLowFelt := &f.Element{}
	remainderLowFelt.SetBigInt(remainderLow)
	remainderHighFelt := &f.Element{}
	remainderHighFelt.SetBigInt(remainderHigh)

	quotientLowCell, err := hint.Quotient0.Get(vm)
	if err != nil {
		return fmt.Errorf("get destination cell: %v", err)
	}
	err = quotientLowCell.Write(memory.MemoryValueFromFieldElement(quotientLowFelt))
	if err != nil {
		return fmt.Errorf("write cell: %v", err)
	}

	quotientHighCell, err := hint.Quotient1.Get(vm)
	if err != nil {
		return fmt.Errorf("get destination cell: %v", err)
	}
	err = quotientHighCell.Write(memory.MemoryValueFromFieldElement(quotientHighFelt))
	if err != nil {
		return fmt.Errorf("write cell: %v", err)
	}

	remainderLowCell, err := hint.Remainder0.Get(vm)
	if err != nil {
		return fmt.Errorf("get destination cell: %v", err)
	}
	err = remainderLowCell.Write(memory.MemoryValueFromFieldElement(remainderLowFelt))
	if err != nil {
		return fmt.Errorf("write cell: %v", err)
	}

	remainderHighCell, err := hint.Remainder1.Get(vm)
	if err != nil {
		return fmt.Errorf("get destination cell: %v", err)
	}
	err = remainderHighCell.Write(memory.MemoryValueFromFieldElement(remainderHighFelt))
	if err != nil {
		return fmt.Errorf("write cell: %v", err)
	}

	return nil

}
