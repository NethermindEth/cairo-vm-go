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

type DebugPrint struct {
	start ResOperander
	end   ResOperander
}

func (hint DebugPrint) Execute(vm *VM.VirtualMachine) error {
	start, err := hint.start.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve start operand %s: %v", hint.start, err)
	}

	startAddr, err := start.MemoryAddress()
	if err != nil {
		return fmt.Errorf("start memory address: %v", err)
	}

	end, err := hint.end.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve end operand %s: %v", hint.end, err)
	}
	endAddr, err := end.MemoryAddress()
	if err != nil {
		return fmt.Errorf("end memory address: %v", err)
	}

	if startAddr.Offset > endAddr.Offset {
		return fmt.Errorf("start cannot be greater than end")
	}

	current := startAddr.Offset
	for current < endAddr.Offset {
		v, err := vm.Memory.ReadFromAddress(&memory.MemoryAddress{
			SegmentIndex: startAddr.SegmentIndex,
			Offset:       current,
		})
		if err != nil {
			return err
		}

		field, _ := v.FieldElement()
		fmt.Printf("[DEBUG] %s\n", field.Text(16))
		current += 1
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

	dstAddr, err := hint.dst.Get(vm)
	if err != nil {
		return fmt.Errorf("get destination cell: %v", err)
	}

	dstVal := memory.MemoryValueFromFieldElement(sqrt)
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

	valueHigh, err := hint.valueHigh.Resolve(vm)
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
	value := uint256.Int(valueHighFelt.Bits())
	value.Lsh(&value, 128)
	value.Add(&value, &valueLowU256)

	// root = math.isqrt(value)
	root := uint256.Int{}
	root.Sqrt(&value)

	// remainder = value - root ** 2
	root2 := uint256.Int{}
	root2.Mul(&root, &root)
	remainder := uint256.Int{}
	remainder.Sub(&value, &root2)

	// memory{sqrt0} = root & 0xFFFFFFFFFFFFFFFF
	// memory{sqrt1} = root >> 64
	mask64 := uint256.NewInt(0xFFFFFFFFFFFFFFFF)
	rootMasked := uint256.Int{}
	rootMasked.And(&root, mask64)
	rootShifted := root.Rsh(&root, 64)

	sqrt0 := f.Element{}
	sqrt0.SetBytes(rootMasked.Bytes())

	sqrt1 := f.Element{}
	sqrt1.SetBytes(rootShifted.Bytes())

	sqrt0Addr, err := hint.sqrt0.Get(vm)
	if err != nil {
		return fmt.Errorf("get sqrt0 cell: %v", err)
	}

	sqrt1Addr, err := hint.sqrt1.Get(vm)
	if err != nil {
		return fmt.Errorf("get sqrt1 cell: %v", err)
	}

	sqrt0Val := memory.MemoryValueFromFieldElement(&sqrt0)
	err = vm.Memory.WriteToAddress(&sqrt0Addr, &sqrt0Val)
	if err != nil {
		return fmt.Errorf("write sqrt0 cell: %v", err)
	}

	sqrt1Val := memory.MemoryValueFromFieldElement(&sqrt1)
	err = vm.Memory.WriteToAddress(&sqrt1Addr, &sqrt1Val)
	if err != nil {
		return fmt.Errorf("write sqrt1 cell: %v", err)
	}

	// memory{remainder_low} = remainder & 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF
	// memory{remainder_high} = remainder >> 128
	mask128 := uint256.NewInt(0xFFFFFFFFFFFFFFFF)
	mask128.Lsh(mask128, 64)
	mask128.Or(mask128, mask64)
	remainderMasked := uint256.Int{}
	remainderMasked.And(&remainder, mask128)
	remainderLow := f.Element{}
	remainderLow.SetBytes(remainderMasked.Bytes())

	remainderShifted := uint256.Int{}
	remainderShifted.Rsh(&remainder, 128)
	remainderHigh := f.Element{}
	remainderHigh.SetBytes(remainderShifted.Bytes())

	remainderLowAddr, err := hint.remainderLow.Get(vm)
	if err != nil {
		return fmt.Errorf("get remainderLow cell: %v", err)
	}

	remainderHighAddr, err := hint.remainderHigh.Get(vm)
	if err != nil {
		return fmt.Errorf("get remainderHigh cell: %v", err)
	}

	remainderLowVal := memory.MemoryValueFromFieldElement(&remainderLow)
	err = vm.Memory.WriteToAddress(&remainderLowAddr, &remainderLowVal)
	if err != nil {
		return fmt.Errorf("write remainderLow cell: %v", err)
	}

	remainderHighVal := memory.MemoryValueFromFieldElement(&remainderHigh)
	err = vm.Memory.WriteToAddress(&remainderHighAddr, &remainderHighVal)
	if err != nil {
		return fmt.Errorf("write remainderHigh cell: %v", err)
	}

	// memory{sqrt_mul_2_minus_remainder_ge_u128} = root * 2 - remainder >= 2**128
	rootMul2 := uint256.Int{}
	rootMul2.Lsh(&root, 1)
	lhs := uint256.Int{}
	lhs.Sub(&rootMul2, &remainder)

	rhs := uint256.NewInt(1)
	rhs.Lsh(rhs, 128)
	result := rhs.Gt(&lhs)
	result = !result

	sqrtMul2MinusRemainderGeU128 := f.Element{}
	if result {
		sqrtMul2MinusRemainderGeU128.SetOne()
	}

	sqrtMul2MinusRemainderGeU128Addr, err := hint.sqrtMul2MinusRemainderGeU128.Get(vm)
	if err != nil {
		return fmt.Errorf("get sqrtMul2MinusRemainderGeU128Addr cell: %v", err)
	}

	sqrtMul2MinusRemainderGeU128AddrVal := memory.MemoryValueFromFieldElement(&sqrtMul2MinusRemainderGeU128)
	err = vm.Memory.WriteToAddress(&sqrtMul2MinusRemainderGeU128Addr, &sqrtMul2MinusRemainderGeU128AddrVal)
	if err != nil {
		return fmt.Errorf("write sqrtMul2MinusRemainderGeU128Addr cell: %v", err)
	}

	return nil
}
