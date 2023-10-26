package hintrunner

import (
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/safemath"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

//
// All CellRef definitions

type CellRefer interface {
	fmt.Stringer

	Get(vm *VM.VirtualMachine) (mem.MemoryAddress, error)
}

type ApCellRef int16

func (ap ApCellRef) String() string {
	return fmt.Sprintf("ApCellRef(%d)", ap)
}

func (ap ApCellRef) Get(vm *VM.VirtualMachine) (mem.MemoryAddress, error) {
	res, overflow := safemath.SafeOffset(vm.Context.Ap, int16(ap))
	if overflow {
		return mem.UnknownAddress, safemath.NewSafeOffsetError(vm.Context.Ap, int16(ap))
	}
	return mem.MemoryAddress{SegmentIndex: VM.ExecutionSegment, Offset: res}, nil
}

type FpCellRef int16

func (fp FpCellRef) String() string {
	return fmt.Sprintf("FpCellRef(%d)", fp)
}

func (fp FpCellRef) Get(vm *VM.VirtualMachine) (mem.MemoryAddress, error) {
	res, overflow := safemath.SafeOffset(vm.Context.Fp, int16(fp))
	if overflow {
		return mem.MemoryAddress{}, safemath.NewSafeOffsetError(vm.Context.Ap, int16(fp))
	}
	return mem.MemoryAddress{SegmentIndex: VM.ExecutionSegment, Offset: res}, nil
}

//
// All ResOperand definitions

type ResOperander interface {
	fmt.Stringer

	Resolve(vm *VM.VirtualMachine) (mem.MemoryValue, error)
}

type Deref struct {
	deref CellRefer
}

func (deref Deref) String() string {
	return "Deref"
}

func (deref Deref) Resolve(vm *VM.VirtualMachine) (mem.MemoryValue, error) {
	address, err := deref.deref.Get(vm)
	if err != nil {
		return mem.MemoryValue{}, fmt.Errorf("get cell: %w", err)
	}
	return vm.Memory.ReadFromAddress(&address)
}

type DoubleDeref struct {
	deref  CellRefer
	offset int16
}

func (dderef DoubleDeref) Resolve(vm *VM.VirtualMachine) (mem.MemoryValue, error) {
	lhsAddr, err := dderef.deref.Get(vm)
	if err != nil {
		return mem.UnknownValue, fmt.Errorf("get lhs address %s: %w", lhsAddr, err)
	}
	lhs, err := vm.Memory.ReadFromAddress(&lhsAddr)
	if err != nil {
		return mem.UnknownValue, fmt.Errorf("read lhs address %s: %w", lhsAddr, err)
	}

	// Double deref implies the left hand side read must be an address
	address, err := lhs.MemoryAddress()
	if err != nil {
		return mem.UnknownValue, err
	}

	newOffset, overflow := safemath.SafeOffset(address.Offset, dderef.offset)
	if overflow {
		return mem.UnknownValue, safemath.NewSafeOffsetError(address.Offset, dderef.offset)
	}
	resAddr := mem.MemoryAddress{
		SegmentIndex: address.SegmentIndex,
		Offset:       newOffset,
	}

	value, err := vm.Memory.ReadFromAddress(&resAddr)
	if err != nil {
		return mem.UnknownValue, fmt.Errorf("read result at %s: %w", resAddr, err)
	}

	return value, nil
}

type Immediate f.Element

func (imm Immediate) String() string {
	return "Immediate"
}

// Should we respect that, or go straight to felt?
func (imm Immediate) Resolve(vm *VM.VirtualMachine) (mem.MemoryValue, error) {
	felt := f.Element(imm)
	return mem.MemoryValueFromFieldElement(&felt), nil
}

type Operator uint8

const (
	Add Operator = iota
	Mul
)

type BinaryOp struct {
	operator Operator
	lhs      CellRefer
	rhs      ResOperander // (except DoubleDeref and BinaryOp)
}

func (bop BinaryOp) String() string {
	return "BinaryOperator"
}

func (bop BinaryOp) Resolve(vm *VM.VirtualMachine) (mem.MemoryValue, error) {
	lhsAddr, err := bop.lhs.Get(vm)
	if err != nil {
		return mem.UnknownValue, fmt.Errorf("get lhs address %s: %w", bop.lhs, err)
	}
	lhs, err := vm.Memory.ReadFromAddress(&lhsAddr)
	if err != nil {
		return mem.UnknownValue, fmt.Errorf("read lhs address %s: %w", lhsAddr, err)
	}

	rhs, err := bop.rhs.Resolve(vm)
	if err != nil {
		return mem.UnknownValue, fmt.Errorf("resolve rhs operand %s: %w", rhs, err)
	}

	switch bop.operator {
	case Add:
		mv := mem.EmptyMemoryValueAs(lhs.IsAddress() || rhs.IsAddress())
		err := mv.Add(&lhs, &rhs)
		return mv, err
	case Mul:
		mv := mem.EmptyMemoryValueAsFelt()
		err := mv.Mul(&lhs, &rhs)
		return mv, err
	default:
		return mem.UnknownValue, fmt.Errorf("unknown binary operator: %d", bop.operator)
	}
}
