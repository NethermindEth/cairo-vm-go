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

	Get(vm *VM.VirtualMachine) (*memory.Cell, error)
}

type ApCellRef int16

func (ap ApCellRef) String() string {
	return "ApCellRef"
}

func (ap ApCellRef) Get(vm *VM.VirtualMachine) (*memory.Cell, error) {
	res, overflow := safemath.SafeOffset(vm.Context.Ap, int16(ap))
	if overflow {
		return nil, safemath.NewSafeOffsetError(vm.Context.Ap, int16(ap))
	}
	return vm.MemoryManager.Memory.Peek(VM.ExecutionSegment, res)
}

type FpCellRef int16

func (fp FpCellRef) String() string {
	return "FpCellRef"
}

func (fp FpCellRef) Get(vm *VM.VirtualMachine) (*memory.Cell, error) {
	res, overflow := safemath.SafeOffset(vm.Context.Fp, int16(fp))
	if overflow {
		return nil, safemath.NewSafeOffsetError(vm.Context.Ap, int16(fp))
	}
	return vm.MemoryManager.Memory.Peek(VM.ExecutionSegment, res)
}

//
// All ResOperand definitions

type ResOperander interface {
	fmt.Stringer

	Resolve(vm *VM.VirtualMachine) (*memory.MemoryValue, error)
}

type Deref struct {
	deref CellRefer
}

func (deref Deref) String() string {
	return "Deref"
}

func (deref Deref) Resolve(vm *VM.VirtualMachine) (*memory.MemoryValue, error) {
	cell, err := deref.deref.Get(vm)
	if err != nil {
		return nil, fmt.Errorf("get cell: %v", err)
	}
	return cell.Read(), nil
}

type DoubleDeref struct {
	deref  CellRefer
	offset int16
}

func (dderef DoubleDeref) Resolve(vm *VM.VirtualMachine) (*memory.MemoryValue, error) {
	cell, err := dderef.deref.Get(vm)
	if err != nil {
		return nil, fmt.Errorf("get cell: %v", err)
	}
	lhs := cell.Read()

	// Double deref implies the first value read must be an address
	address, err := lhs.ToMemoryAddress()
	if err != nil {
		return nil, err
	}

	newOffset, overflow := safemath.SafeOffset(address.Offset, dderef.offset)
	if overflow {
		return nil, safemath.NewSafeOffsetError(address.Offset, dderef.offset)
	}
	resAddr := memory.NewMemoryAddress(address.SegmentIndex, newOffset)

	value, err := vm.MemoryManager.Memory.ReadFromAddress(resAddr)
	if err != nil {
		return nil, fmt.Errorf("read cell: %v", err)
	}

	return value, nil
}

type Immediate big.Int

func (imm Immediate) String() string {
	return "Immediate"
}

// todo(rodro): Specs from Starkware stablish this can be uint256 and not a felt.
// Should we respect that, or go straight to felt?
func (imm Immediate) Resolve(vm *VM.VirtualMachine) (*memory.MemoryValue, error) {
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

func (bop BinaryOp) Resolve(vm *VM.VirtualMachine) (*memory.MemoryValue, error) {
	cell, err := bop.lhs.Get(vm)
	if err != nil {
		return nil, fmt.Errorf("get lhs operand %s: %v", bop.lhs, err)
	}
	lhs := cell.Read()

	rhs, err := bop.rhs.Resolve(vm)
	if err != nil {
		return nil, fmt.Errorf("resolve rhs operand %s: %v", rhs, err)
	}

	switch bop.operator {
	case Add:
		return memory.EmptyMemoryValueAs(lhs.IsAddress()).Add(lhs, rhs)
	case Mul:
		return memory.EmptyMemoryValueAsFelt().Mul(lhs, rhs)
	default:
		return nil, fmt.Errorf("unknown binary operator: %d", bop.operator)
	}
}
