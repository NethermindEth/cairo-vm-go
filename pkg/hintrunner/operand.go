package hintrunner

import (
	"fmt"
	"math/big"

	"github.com/NethermindEth/cairo-vm-go/pkg/safemath"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

const (
	apCellRefName   = "ApCellRef"
	fpCellRefName   = "FpCellRef"
	derefName       = "Deref"
	doubleDerefName = "DoubleDeref"
	immediateName   = "Immediate"
	binOpName       = "BinaryOperator"
)

//
// All CellRef definitions

type CellRefer interface {
	Get(vm *VM.VirtualMachine) (*memory.Cell, error)
}

type ApCellRef int16

type FpCellRef int16

func (ap ApCellRef) Get(vm *VM.VirtualMachine) (*memory.Cell, error) {
	res, overflow := safemath.SafeOffset(vm.Context.Ap, int16(ap))
	if overflow {
		return nil, NewOperandError(
			apCellRefName,
			fmt.Errorf("%d + %d is outside of the [0, 2**64) range", vm.Context.Ap, ap),
		)
	}
	return vm.MemoryManager.Memory.Peek(VM.ExecutionSegment, res)
}

func (fp FpCellRef) Get(vm *VM.VirtualMachine) (*memory.Cell, error) {
	res, overflow := safemath.SafeOffset(vm.Context.Fp, int16(fp))
	if overflow {
		return nil, NewOperandError(
			fpCellRefName,
			fmt.Errorf("%d + %d is outside of the [0, 2**64) range", vm.Context.Ap, fp),
		)
	}
	return vm.MemoryManager.Memory.Peek(VM.ExecutionSegment, res)
}

//
// All ResOperand definitions

type ResOperander interface {
	Resolve(vm *VM.VirtualMachine) (*memory.MemoryValue, error)
}

type Deref struct {
	deref CellRefer
}

type DoubleDeref struct {
	deref  CellRefer
	offset int16
}

type Immediate big.Int

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

func (deref Deref) Resolve(vm *VM.VirtualMachine) (*memory.MemoryValue, error) {
	cell, err := deref.deref.Get(vm)
	if err != nil {
		return nil, NewOperandError(derefName, err)
	}
	return cell.Read(), nil
}

func (dderef DoubleDeref) Resolve(vm *VM.VirtualMachine) (*memory.MemoryValue, error) {
	cell, err := dderef.deref.Get(vm)
	if err != nil {
		return nil, NewOperandError(doubleDerefName, err)
	}
	lhs := cell.Read()

	// Double deref implies the first value read must be an address
	address, err := lhs.ToMemoryAddress()
	if err != nil {
		return nil, NewOperandError(doubleDerefName, err)
	}

	newOffset, overflow := safemath.SafeOffset(address.Offset, dderef.offset)
	if overflow {
		return nil, NewOperandError(
			doubleDerefName,
			safemath.NewSafeOffsetError(address.Offset, dderef.offset),
		)
	}
	resAddr := memory.NewMemoryAddress(address.SegmentIndex, newOffset)

	value, err := vm.MemoryManager.Memory.ReadFromAddress(resAddr)
	if err != nil {
		return nil, NewOperandError(doubleDerefName, err)
	}

	return value, nil
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

func (bop BinaryOp) Resolve(vm *VM.VirtualMachine) (*memory.MemoryValue, error) {
	cell, err := bop.lhs.Get(vm)
	if err != nil {
		return nil, err
	}
	lhs := cell.Read()

	rhs, err := bop.rhs.Resolve(vm)
	if err != nil {
		return nil, err
	}

	switch bop.operator {
	case Add:
		return memory.EmptyMemoryValueAs(lhs.IsAddress()).Add(lhs, rhs)
	case Mul:
		return memory.EmptyMemoryValueAsFelt().Mul(lhs, rhs)
	}

	return nil, NewOperandError(
		"BinaryOp",
		fmt.Errorf("unknown binary operator id: %d", bop.operator),
	)
}
