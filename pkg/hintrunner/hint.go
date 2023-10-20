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

type SquareRoot struct {
	value ResOperander
	dst   CellRefer
}

func (hint SquareRoot) String() string {
	return "SquareRoot"
}

func (hint SquareRoot) Execute(vm *VM.VirtualMachine) error {
	value, err := hint.value.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve value operand %s: %v", hint.value, err)
	}

	valueFelt, err := value.FieldElement()
	if err != nil {
		return err
	}

	sqrt := valueFelt.Sqrt(valueFelt)

	dst := f.Element{}
	dst.Set(sqrt)

	dstAddr, err := hint.dst.Get(vm)
	if err != nil {
		return fmt.Errorf("get destination cell: %v", err)
	}
	dstVal := memory.MemoryValueFromFieldElement(&dst)
	err = vm.Memory.WriteToAddress(&dstAddr, &dstVal)
	if err != nil {
		return fmt.Errorf("write cell: %v", err)
	}
	return nil
}

type Uint256SquareRoot struct {
	valueLow                     ResOperander
	valueHigh                    ResOperander
	sqrt0                        CellRefer
	sqrt1                        CellRefer
	remainderLow                 CellRefer
	remainderHigh                CellRefer
	sqrtMul2MinusRemainderGeU128 CellRefer
}

func (hint Uint256SquareRoot) String() string {
	return "Uint256SquareRoot"
}

func (hint Uint256SquareRoot) Execute(vm *VM.VirtualMachine) error {
	valueLow, err := hint.valueLow.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve valueLow operand %s: %v", hint.valueLow, err)
	}

	valueHigh, err := hint.valueLow.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve valueHigh operand %s: %v", hint.valueHigh, err)
	}

	valueLowFelt, err := valueLow.FieldElement()
	if err != nil {
		return err
	}

	valueHighFelt, err := valueHigh.FieldElement()
	if err != nil {
		return err
	}

	// value = {value_low} + {value_high} * 2**128
	valueLowU256 := uint256.Int(valueLowFelt.Bits())
	valueHighU256 := uint256.Int(valueHighFelt.Bits())

	shifted := valueLowU256.Lsh(&valueLowU256, 128)
	value := valueHighU256.Add(shifted, &valueHighU256)

	// root = math.isqrt(value)
	root := value.Sqrt(value)
	// remainder = value - root ** 2
	// TODO: Is there any Power method?
	root2 := root.Mul(root, root)
	remainder := value.Sub(value, root2)

	// memory{sqrt0} = root & 0xFFFFFFFFFFFFFFFF
	mask64 := uint256.NewInt(0xFFFFFFFFFFFFFFFF)
	rootMasked := root.And(root, mask64)
	sqrt0 := f.Element{}
	sqrt0.SetBytes(rootMasked.Bytes())

	sqrt0Addr, err := hint.sqrt0.Get(vm)
	if err != nil {
		return fmt.Errorf("get sqrt0 cell: %v", err)
	}

	sqrt0Val := memory.MemoryValueFromFieldElement(&sqrt0)
	err = vm.Memory.WriteToAddress(&sqrt0Addr, &sqrt0Val)
	if err != nil {
		return fmt.Errorf("get remainderHigh cell: %v", err)
	}

	// memory{sqrt1} = root >> 64
	rootShifted := root.Rsh(root, 64)
	sqrt1 := f.Element{}
	sqrt1.SetBytes(rootShifted.Bytes())

	sqrt1Addr, err := hint.sqrt1.Get(vm)
	if err != nil {
		return fmt.Errorf("get sqrt0 cell: %v", err)
	}

	sqrt1Val := memory.MemoryValueFromFieldElement(&sqrt1)
	err = vm.Memory.WriteToAddress(&sqrt1Addr, &sqrt1Val)
	if err != nil {
		return fmt.Errorf("get remainderHigh cell: %v", err)
	}

	// memory{remainder_low} = remainder & 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF
	mask128 := MaxU128()
	remainderMasked := remainder.And(remainder, &mask128)
	remainderLow := f.Element{}
	remainderLow.SetBytes(remainderMasked.Bytes())

	remainderLowAddr, err := hint.remainderLow.Get(vm)
	if err != nil {
		return fmt.Errorf("get remainderLow cell: %v", err)
	}

	remainderLowVal := memory.MemoryValueFromFieldElement(&remainderLow)
	err = vm.Memory.WriteToAddress(&remainderLowAddr, &remainderLowVal)
	if err != nil {
		return fmt.Errorf("get remainderHigh cell: %v", err)
	}

	// memory{remainder_high} = remainder >> 128
	remainderShifted := remainder.Rsh(remainder, 128)
	remainderHigh := f.Element{}
	remainderHigh.SetBytes(remainderShifted.Bytes())

	remainderHighAddr, err := hint.remainderHigh.Get(vm)
	if err != nil {
		return fmt.Errorf("get remainderHigh cell: %v", err)
	}

	remainderHighVal := memory.MemoryValueFromFieldElement(&remainderHigh)
	err = vm.Memory.WriteToAddress(&remainderHighAddr, &remainderHighVal)
	if err != nil {
		return fmt.Errorf("get remainderHigh cell: %v", err)
	}

	// memory{sqrt_mul_2_minus_remainder_ge_u128} = root * 2 - remainder >= 2**128
	rootMul2 := root.Lsh(root, 1)
	lhs := rootMul2.Sub(rootMul2, remainder)
	rhs := mask128.Mul(&mask128, &mask128)
	result := rhs.Gt(lhs)
	result = !result
	sqrtMul2MinusRemainderGeU128 := f.Element{}
	if result {
		sqrtMul2MinusRemainderGeU128.SetOne()
	} else {
		sqrtMul2MinusRemainderGeU128.SetZero()
	}

	sqrtMul2MinusRemainderGeU128Addr, err := hint.sqrtMul2MinusRemainderGeU128.Get(vm)
	if err != nil {
		return fmt.Errorf("get remainderHigh cell: %v", err)
	}

	sqrtMul2MinusRemainderGeU128AddrVal := memory.MemoryValueFromFieldElement(&sqrtMul2MinusRemainderGeU128)
	err = vm.Memory.WriteToAddress(&sqrtMul2MinusRemainderGeU128Addr, &sqrtMul2MinusRemainderGeU128AddrVal)
	if err != nil {
		return fmt.Errorf("get remainderHigh cell: %v", err)
	}

	return nil
}
