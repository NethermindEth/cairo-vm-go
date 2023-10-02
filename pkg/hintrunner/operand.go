package hintrunner

import (
	"fmt"
	"math/big"

	"github.com/NethermindEth/cairo-vm-go/pkg/safemath"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

//
// All CellRef definitions

type CellRefer interface {
	fmt.Stringer

	Get(vm *VM.VirtualMachine) (memory.MemoryAddress, error)
}

type ApCellRef int16

func (ap ApCellRef) String() string {
	return fmt.Sprintf("ApCellRef(%d)", ap)
}

func (ap ApCellRef) Get(vm *VM.VirtualMachine) (memory.MemoryAddress, error) {
	res, overflow := safemath.SafeOffset(vm.Context.Ap, int16(ap))
	if overflow {
		return memory.MemoryAddress{}, safemath.NewSafeOffsetError(vm.Context.Ap, int16(ap))
	}
	return memory.MemoryAddress{SegmentIndex: VM.ExecutionSegment, Offset: res}, nil
}

type FpCellRef int16

func (fp FpCellRef) String() string {
	return fmt.Sprintf("FpCellRef(%d)", fp)
}

func (fp FpCellRef) Get(vm *VM.VirtualMachine) (memory.MemoryAddress, error) {
	res, overflow := safemath.SafeOffset(vm.Context.Fp, int16(fp))
	if overflow {
		return memory.MemoryAddress{}, safemath.NewSafeOffsetError(vm.Context.Ap, int16(fp))
	}
	return memory.MemoryAddress{SegmentIndex: VM.ExecutionSegment, Offset: res}, nil
}

//
// All ResOperand definitions

type ResOperander interface {
	fmt.Stringer

	Resolve(vm *VM.VirtualMachine) (memory.MemoryValue, error)
}

type Deref struct {
	deref CellRefer
}

func (deref Deref) String() string {
	return "Deref"
}

func (deref Deref) Resolve(vm *VM.VirtualMachine) (memory.MemoryValue, error) {
	address, err := deref.deref.Get(vm)
	if err != nil {
		return memory.MemoryValue{}, fmt.Errorf("get cell: %w", err)
	}
	return vm.MemoryManager.Memory.ReadFromAddress(&address)
}

type DoubleDeref struct {
	deref  CellRefer
	offset int16
}

func (dderef DoubleDeref) Resolve(vm *VM.VirtualMachine) (memory.MemoryValue, error) {
	lhsAddr, err := dderef.deref.Get(vm)
	if err != nil {
		return memory.MemoryValue{}, fmt.Errorf("get lhs address %s: %w", lhsAddr, err)
	}
	lhs, err := vm.MemoryManager.Memory.ReadFromAddress(&lhsAddr)
	if err != nil {
		return memory.MemoryValue{}, fmt.Errorf("read lhs address %s: %w", lhsAddr, err)
	}

	// Double deref implies the left hand side read must be an address
	address, err := lhs.ToMemoryAddress()
	if err != nil {
		return memory.MemoryValue{}, err
	}

	newOffset, overflow := safemath.SafeOffset(address.Offset, dderef.offset)
	if overflow {
		return memory.MemoryValue{}, safemath.NewSafeOffsetError(address.Offset, dderef.offset)
	}
	resAddr := memory.MemoryAddress{
		SegmentIndex: address.SegmentIndex,
		Offset:       newOffset,
	}

	value, err := vm.MemoryManager.Memory.ReadFromAddress(&resAddr)
	if err != nil {
		return memory.MemoryValue{}, fmt.Errorf("read result at %s: %w", resAddr, err)
	}

	return value, nil
}

type Immediate big.Int

func (imm Immediate) String() string {
	return "Immediate"
}

// todo(rodro): Specs from Starkware stablish this can be uint256 and not a felt.
// Should we respect that, or go straight to felt?
func (imm Immediate) Resolve(vm *VM.VirtualMachine) (memory.MemoryValue, error) {
	felt := &f.Element{}
	bigInt := (big.Int)(imm)
	// todo(rodro): do we require to check that big int is lesser than P, or do we
	// just take: big_int `mod` P?
	felt.SetBigInt(&bigInt)

	return memory.MemoryValueFromFieldElement(felt), nil
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

func (bop BinaryOp) Resolve(vm *VM.VirtualMachine) (memory.MemoryValue, error) {
	lhsAddr, err := bop.lhs.Get(vm)
	if err != nil {
		return memory.MemoryValue{}, fmt.Errorf("get lhs address %s: %w", bop.lhs, err)
	}
	lhs, err := vm.MemoryManager.Memory.ReadFromAddress(&lhsAddr)
	if err != nil {
		return memory.MemoryValue{}, fmt.Errorf("read lhs address %s: %v", lhsAddr, err)
	}

	rhs, err := bop.rhs.Resolve(vm)
	if err != nil {
		return memory.MemoryValue{}, fmt.Errorf("resolve rhs operand %s: %v", rhs, err)
	}

	switch bop.operator {
	case Add:
		mv := memory.EmptyMemoryValueAs(lhs.IsAddress() || rhs.IsAddress())
		err := mv.Add(&lhs, &rhs)
		return mv, err
	case Mul:
		mv := memory.EmptyMemoryValueAsFelt()
		err := mv.Mul(&lhs, &rhs)
		return mv, err
	default:
		return memory.MemoryValue{}, fmt.Errorf("unknown binary operator: %d", bop.operator)
	}
}
