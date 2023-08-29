package hintrunner

import (
	"fmt"
	"math/big"

	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

//
// All CellRef definitions

type CellRefer interface {
	Get(vm *VM.VirtualMachine) (*memory.Cell, error)
}

type ApCellRef int16

type FpCellRef int16

func (ap ApCellRef) Get(vm *VM.VirtualMachine) (*memory.Cell, error) {
	// todo(rodro): fix maths with safemath from ilia
	offset := vm.Context.Ap + uint64(ap)
	return vm.MemoryManager.Memory.Peek(VM.ExecutionSegment, offset)
}

func (fp FpCellRef) Get(vm *VM.VirtualMachine) (*memory.Cell, error) {
	// todo(rodro): fix maths with safemath from ilia
	offset := vm.Context.Fp + uint64(fp)
	return vm.MemoryManager.Memory.Peek(VM.ExecutionSegment, offset)
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
		return nil, err
	}
	return cell.Read(), nil
}

func (dderef DoubleDeref) Resolve(vm *VM.VirtualMachine) (*memory.MemoryValue, error) {
	cell, err := dderef.deref.Get(vm)
	if err != nil {
		return nil, err
	}
	lhs := cell.Read()

	var res *memory.MemoryValue
	if dderef.offset >= 0 {
		rhs := memory.MemoryValueFromInt(dderef.offset)
		res, err = memory.EmptyMemoryValueAs(lhs.IsAddress()).Add(lhs, rhs)
	} else {
		rhs := memory.MemoryValueFromInt(Abs(dderef.offset))
		res, err = memory.EmptyMemoryValueAs(lhs.IsAddress()).Sub(lhs, rhs)
	}
	if err != nil {
		return nil, err
	}

	return res, nil
}

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

	return nil, fmt.Errorf("Unknown operator: %d", bop.operator)
}
