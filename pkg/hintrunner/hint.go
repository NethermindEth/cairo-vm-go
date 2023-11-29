package hintrunner

import (
	"fmt"
	"sort"

	"github.com/holiman/uint256"

	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"

	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

type Hinter interface {
	fmt.Stringer

	Execute(vm *VM.VirtualMachine, ctx *HintRunnerContext) error
}

type AllocSegment struct {
	dst CellRefer
}

func (hint *AllocSegment) String() string {
	return "AllocSegment"
}

func (hint *AllocSegment) Execute(vm *VM.VirtualMachine, _ *HintRunnerContext) error {
	newSegment := vm.Memory.AllocateEmptySegment()
	memAddress := mem.MemoryValueFromMemoryAddress(&newSegment)

	regAddr, err := hint.dst.Get(vm)
	if err != nil {
		return fmt.Errorf("get register %s: %w", hint.dst, err)
	}

	err = vm.Memory.WriteToAddress(&regAddr, &memAddress)
	if err != nil {
		return fmt.Errorf("write to address %s: %w", regAddr, err)
	}

	return nil
}

type TestLessThan struct {
	dst CellRefer
	lhs ResOperander
	rhs ResOperander
}

func (hint *TestLessThan) String() string {
	return "TestLessThan"
}

func (hint *TestLessThan) Execute(vm *VM.VirtualMachine, _ *HintRunnerContext) error {
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

	mv := mem.MemoryValueFromFieldElement(&resFelt)
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

func (hint *TestLessThanOrEqual) String() string {
	return "TestLessThanOrEqual"
}

func (hint *TestLessThanOrEqual) Execute(vm *VM.VirtualMachine, _ *HintRunnerContext) error {
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

	mv := mem.MemoryValueFromFieldElement(&resFelt)
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
	mv := mem.MemoryValueFromFieldElement(xFiled)
	err = vm.Memory.WriteToAddress(&xAddr, &mv)
	if err != nil {
		return fmt.Errorf("write to x address %s: %w", xAddr, err)
	}

	mv = mem.MemoryValueFromFieldElement(yFiled)
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

func (hint *WideMul128) String() string {
	return "WideMul128"
}

func (hint *WideMul128) Execute(vm *VM.VirtualMachine, _ *HintRunnerContext) error {
	mask := &utils.Uint256Max128

	lhs, err := hint.lhs.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve lhs operand %s: %w", hint.lhs, err)
	}
	rhs, err := hint.rhs.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve rhs operand %s: %w", hint.rhs, err)
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

	if lhsU256.Gt(mask) {
		return fmt.Errorf("lhs operand %s should be u128", lhsFelt)
	}
	if rhsU256.Gt(mask) {
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
		return fmt.Errorf("get destination cell: %w", err)
	}
	mvLow := mem.MemoryValueFromFieldElement(&low)
	err = vm.Memory.WriteToAddress(&lowAddr, &mvLow)
	if err != nil {
		return fmt.Errorf("write cell: %v", err)
	}

	highAddr, err := hint.high.Get(vm)
	if err != nil {
		return fmt.Errorf("get destination cell: %w", err)
	}
	mvHigh := mem.MemoryValueFromFieldElement(&high)
	err = vm.Memory.WriteToAddress(&highAddr, &mvHigh)
	if err != nil {
		return fmt.Errorf("write cell: %w", err)
	}
	return nil
}

type DivMod struct {
	lhs       ResOperander
	rhs       ResOperander
	quotient  CellRefer
	remainder CellRefer
}

func (hint DivMod) String() string {
	return "DivMod"
}

func (hint DivMod) Execute(vm *VM.VirtualMachine) error {
	lhsVal, err := hint.lhs.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve lhs operand %s: %v", hint.lhs, err)
	}

	rhsVal, err := hint.rhs.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve rhs operand %s: %v", hint.rhs, err)
	}

	lhsFelt, err := lhsVal.FieldElement()
	if err != nil {
		return err
	}

	rhsFelt, err := rhsVal.FieldElement()
	if err != nil {
		return err
	}

	if rhsFelt.IsZero() {
		return fmt.Errorf("cannot be divide by zero, rhs: %v", rhsFelt)
	}

	lhsvalue := uint256.Int(lhsFelt.Bits())
	rhsvalue := uint256.Int(rhsFelt.Bits())

	// get quotient
	quo := uint256.Int{}
	quo.Div(&lhsvalue, &rhsvalue)

	quotient := f.Element{}
	quoVal := quo.Uint64()
	quotient.SetUint64(quoVal)

	quotientAddr, err := hint.quotient.Get(vm)
	if err != nil {
		return fmt.Errorf("get quotient cell: %v", err)
	}

	quotientVal := mem.MemoryValueFromFieldElement(&quotient)
	err = vm.Memory.WriteToAddress(&quotientAddr, &quotientVal)
	if err != nil {
		return fmt.Errorf("write cell: %v", err)
	}

	// get remainder: lhs - (rhs * quotient)
	temp := uint256.Int{}
	temp.Mul(&rhsvalue, &quo)

	rem := uint256.Int{}
	rem.Sub(&lhsvalue, &temp)

	remainder := f.Element{}
	remVal := rem.Uint64()
	remainder.SetUint64(remVal)

	remainderAddr, err := hint.remainder.Get(vm)
	if err != nil {
		return fmt.Errorf("get remainder cell: %v", err)
	}

	remainderVal := mem.MemoryValueFromFieldElement(&remainder)
	err = vm.Memory.WriteToAddress(&remainderAddr, &remainderVal)
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
		v, err := vm.Memory.ReadFromAddress(&mem.MemoryAddress{
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

func (hint *SquareRoot) String() string {
	return "SquareRoot"
}

func (hint *SquareRoot) Execute(vm *VM.VirtualMachine, _ *HintRunnerContext) error {
	value, err := hint.value.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve value operand %s: %w", hint.value, err)
	}

	valueFelt, err := value.FieldElement()
	if err != nil {
		return err
	}

	// Need to do this conversion to handle non-square values
	valueU256 := uint256.Int(valueFelt.Bits())
	valueU256.Sqrt(&valueU256)

	sqrt := f.Element{}
	sqrt.SetBytes(valueU256.Bytes())

	dstAddr, err := hint.dst.Get(vm)
	if err != nil {
		return fmt.Errorf("get destination cell: %w", err)
	}

	dstVal := mem.MemoryValueFromFieldElement(&sqrt)
	err = vm.Memory.WriteToAddress(&dstAddr, &dstVal)
	if err != nil {
		return fmt.Errorf("write cell: %w", err)
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

	sqrt0Val := mem.MemoryValueFromFieldElement(&sqrt0)
	err = vm.Memory.WriteToAddress(&sqrt0Addr, &sqrt0Val)
	if err != nil {
		return fmt.Errorf("write sqrt0 cell: %v", err)
	}

	sqrt1Val := mem.MemoryValueFromFieldElement(&sqrt1)
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

	remainderLowVal := mem.MemoryValueFromFieldElement(&remainderLow)
	err = vm.Memory.WriteToAddress(&remainderLowAddr, &remainderLowVal)
	if err != nil {
		return fmt.Errorf("write remainderLow cell: %v", err)
	}

	remainderHighVal := mem.MemoryValueFromFieldElement(&remainderHigh)
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

	sqrtMul2MinusRemainderGeU128AddrVal := mem.MemoryValueFromFieldElement(&sqrtMul2MinusRemainderGeU128)
	err = vm.Memory.WriteToAddress(&sqrtMul2MinusRemainderGeU128Addr, &sqrtMul2MinusRemainderGeU128AddrVal)
	if err != nil {
		return fmt.Errorf("write sqrtMul2MinusRemainderGeU128Addr cell: %v", err)
	}

	return nil
}

//
// Dictionary Hints
//

type AllocFelt252Dict struct {
	SegmentArenaPtr ResOperander
}

func (hint *AllocFelt252Dict) String() string {
	return "AllocFelt252Dict"
}
func (hint *AllocFelt252Dict) Execute(vm *VM.VirtualMachine, ctx *HintRunnerContext) error {
	InitializeDictionaryManagerIfNot(ctx)

	arenaPtr, err := ResolveAsAddress(vm, hint.SegmentArenaPtr)
	if err != nil {
		return fmt.Errorf("resolve segment arena pointer: %w", err)
	}

	// find for the amount of initialized dicts
	initializedDictsOffset, overflow := utils.SafeOffset(arenaPtr.Offset, -2)
	if overflow {
		return fmt.Errorf("look for initialized dicts: overflow: %s - 2", arenaPtr)
	}
	initializedDictsFelt, err := vm.Memory.Read(arenaPtr.SegmentIndex, initializedDictsOffset)
	if err != nil {
		return fmt.Errorf("read initialized dicts: %w", err)
	}
	initializedDicts, err := initializedDictsFelt.Uint64()
	if err != nil {
		return fmt.Errorf("read initialized dicts: %w", err)
	}

	// find for the segment info pointer
	segmentInfoOffset, overflow := utils.SafeOffset(arenaPtr.Offset, -3)
	if overflow {
		return fmt.Errorf("look for segment info pointer: overflow: %s - 3", arenaPtr)
	}
	segmentInfoMv, err := vm.Memory.Read(arenaPtr.SegmentIndex, segmentInfoOffset)
	if err != nil {
		return fmt.Errorf("read segment info pointer: %w", err)
	}
	segmentInfoPtr, err := segmentInfoMv.MemoryAddress()
	if err != nil {
		return fmt.Errorf("expected pointer to segment info but got a felt: %w", err)
	}

	// with the segment info pointer and the number of initialized dictionaries we know
	// where to write the new dictionary
	newDictAddress := ctx.DictionaryManager.NewDictionary(vm)
	mv := mem.MemoryValueFromMemoryAddress(&newDictAddress)
	insertOffset := segmentInfoPtr.Offset + initializedDicts*3
	if err = vm.Memory.Write(segmentInfoPtr.SegmentIndex, insertOffset, &mv); err != nil {
		return fmt.Errorf("write new dictionary to segment info: %w", err)
	}
	return nil
}

type Felt252DictEntryInit struct {
	DictPtr ResOperander
	Key     ResOperander
}

func (hint Felt252DictEntryInit) String() string {
	return "Felt252DictEntryInit"
}

func (hint *Felt252DictEntryInit) Execute(vm *VM.VirtualMachine, ctx *HintRunnerContext) error {
	dictPtr, err := ResolveAsAddress(vm, hint.DictPtr)
	if err != nil {
		return fmt.Errorf("resolve dictionary pointer: %w", err)
	}

	key, err := ResolveAsFelt(vm, hint.Key)
	if err != nil {
		return fmt.Errorf("resolve key: %w", err)
	}

	prevValue, err := ctx.DictionaryManager.At(&dictPtr, &key)
	if err != nil {
		return fmt.Errorf("get dictionary entry: %w", err)
	}
	if prevValue == nil {
		mv := mem.EmptyMemoryValueAsFelt()
		prevValue = &mv
		_ = ctx.DictionaryManager.Set(&dictPtr, &key, prevValue)
	}
	return vm.Memory.Write(dictPtr.SegmentIndex, dictPtr.Offset+1, prevValue)
}

type Felt252DictEntryUpdate struct {
	DictPtr ResOperander
	Value   ResOperander
}

func (hint Felt252DictEntryUpdate) String() string {
	return "Felt252DictEntryUpdate"
}

func (hint *Felt252DictEntryUpdate) Execute(vm *VM.VirtualMachine, ctx *HintRunnerContext) error {
	dictPtr, err := ResolveAsAddress(vm, hint.DictPtr)
	if err != nil {
		return fmt.Errorf("resolve dictionary pointer: %w", err)
	}

	keyPtr, err := dictPtr.AddOffset(-3)
	if err != nil {
		return fmt.Errorf("get key pointer: %w", err)
	}
	keyMv, err := vm.Memory.ReadFromAddress(&keyPtr)
	if err != nil {
		return fmt.Errorf("read key pointer: %w", err)
	}
	key, err := keyMv.FieldElement()
	if err != nil {
		return fmt.Errorf("expected key to be a field element: %w", err)
	}

	value, err := hint.Value.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve value: %w", err)
	}

	return ctx.DictionaryManager.Set(&dictPtr, key, &value)
}

type GetSegmentArenaIndex struct {
	DictIndex  CellRefer
	DictEndPtr ResOperander
}

func (hint *GetSegmentArenaIndex) String() string {
	return "GetSegmentArenaIndex"
}

func (hint *GetSegmentArenaIndex) Execute(vm *VM.VirtualMachine, ctx *HintRunnerContext) error {
	dictIndex, err := hint.DictIndex.Get(vm)
	if err != nil {
		return fmt.Errorf("get dict index: %w", err)
	}

	dictEndPtr, err := ResolveAsAddress(vm, hint.DictEndPtr)
	if err != nil {
		return fmt.Errorf("resolve dict end pointer: %w", err)
	}

	dict, err := ctx.DictionaryManager.GetDictionary(&dictEndPtr)
	if err != nil {
		return fmt.Errorf("get dictionary: %w", err)
	}

	initNum := mem.MemoryValueFromUint(dict.InitNumber())
	return vm.Memory.WriteToAddress(&dictIndex, &initNum)
}

//
// Squashed Dictionary Hints
//

type InitSquashData struct {
	FirstKey     CellRefer
	BigKeys      CellRefer
	DictAccesses ResOperander
	NumAccesses  ResOperander
}

func (hint *InitSquashData) String() string {
	return "InitSquashData"
}

func (hint *InitSquashData) Execute(vm *VM.VirtualMachine, ctx *HintRunnerContext) error {
	// todo(rodro): Don't know if it could be called multiple times, or
	err := InitializeSquashedDictionaryManager(ctx)
	if err != nil {
		return err
	}

	dictAccessPtr, err := ResolveAsAddress(vm, hint.DictAccesses)
	if err != nil {
		return fmt.Errorf("resolve dict access: %w", err)
	}

	numAccess, err := ResolveAsUint64(vm, hint.NumAccesses)
	if err != nil {
		return fmt.Errorf("resolve num access: %w", err)
	}

	const dictAccessSize = 3
	for i := uint64(0); i < numAccess; i++ {
		keyPtr := mem.MemoryAddress{
			SegmentIndex: dictAccessPtr.SegmentIndex,
			Offset:       dictAccessPtr.Offset + i*dictAccessSize,
		}
		key, err := vm.Memory.ReadFromAddressAsElement(&keyPtr)
		if err != nil {
			return fmt.Errorf("reading key at %s: %w", keyPtr, err)
		}

		ctx.SquashedDictionaryManager.Insert(&key, i)
	}
	for key, val := range ctx.SquashedDictionaryManager.KeyToIndices {
		// reverse each indice access list per key
		utils.Reverse(val)
		// store each key
		ctx.SquashedDictionaryManager.Keys = append(ctx.SquashedDictionaryManager.Keys, key)
	}

	// sort the keys in descending order
	sort.Slice(ctx.SquashedDictionaryManager.Keys, func(i, j int) bool {
		return ctx.SquashedDictionaryManager.Keys[i].Cmp(&ctx.SquashedDictionaryManager.Keys[j]) < 0
	})

	// if the first key is bigger than 2^128, signal it
	bigKeysAddr, err := hint.BigKeys.Get(vm)
	if err != nil {
		return fmt.Errorf("get big keys address: %w", err)
	}
	biggestKey := ctx.SquashedDictionaryManager.Keys[0]
	cmpRes := mem.MemoryValueFromUint[uint64](0)
	if biggestKey.Cmp(&utils.FeltMax128) > 0 {
		cmpRes = mem.MemoryValueFromUint[uint64](1)
	}
	err = vm.Memory.WriteToAddress(&bigKeysAddr, &cmpRes)
	if err != nil {
		return fmt.Errorf("write big keys address: %w", err)
	}

	// store the left most, smaller key
	firstKeyAddr, err := hint.FirstKey.Get(vm)
	if err != nil {
		return fmt.Errorf("get first key address: %w", err)
	}
	firstKey, err := ctx.SquashedDictionaryManager.LastKey()
	if err != nil {
		return fmt.Errorf("get first key: %w", err)
	}

	mv := mem.MemoryValueFromFieldElement(&firstKey)
	return vm.Memory.WriteToAddress(&firstKeyAddr, &mv)
}

type GetCurrentAccessIndex struct {
	RangeCheckPtr ResOperander
}

func (hint *GetCurrentAccessIndex) String() string {
	return "GetCurrentAccessIndex"
}

func (hint *GetCurrentAccessIndex) Execute(vm *VM.VirtualMachine, ctx *HintRunnerContext) error {
	rangeCheckPtr, err := ResolveAsAddress(vm, hint.RangeCheckPtr)
	if err != nil {
		return fmt.Errorf("resolve range check pointer: %w", err)
	}

	lastIndex64, err := ctx.SquashedDictionaryManager.LastIndex()
	if err != nil {
		return fmt.Errorf("get last index: %w", err)
	}

	lastIndex := f.NewElement(lastIndex64)
	mv := mem.MemoryValueFromFieldElement(&lastIndex)

	return vm.Memory.WriteToAddress(&rangeCheckPtr, &mv)
}

type ShouldSkipSquashLoop struct {
	ShouldSkipLoop CellRefer
}

func (hint *ShouldSkipSquashLoop) String() string {
	return "ShouldSkipSquashLoop"
}

func (hint *ShouldSkipSquashLoop) Execute(vm *VM.VirtualMachine, ctx *HintRunnerContext) error {
	shouldSkipLoopAddr, err := hint.ShouldSkipLoop.Get(vm)
	if err != nil {
		return fmt.Errorf("get should skip loop address: %w", err)
	}

	var shouldSkipLoop f.Element
	if lastIndices, err := ctx.SquashedDictionaryManager.LastIndices(); err == nil && len(lastIndices) > 1 {
		shouldSkipLoop.SetOne()
	} else if err != nil {
		return fmt.Errorf("get last indices: %w", err)
	}

	mv := mem.MemoryValueFromFieldElement(&shouldSkipLoop)
	return vm.Memory.WriteToAddress(&shouldSkipLoopAddr, &mv)
}

type GetCurrentAccessDelta struct {
	IndexDeltaMinusOne CellRefer
}

func (hint *GetCurrentAccessDelta) String() string {
	return "GetCurrentAccessDelta"
}

func (hint *GetCurrentAccessDelta) Execute(vm *VM.VirtualMachine, ctx *HintRunnerContext) error {
	indexDeltaPtr, err := hint.IndexDeltaMinusOne.Get(vm)
	if err != nil {
		return fmt.Errorf("get index delta address: %w", err)
	}

	previousKeyIndex, err := ctx.SquashedDictionaryManager.PopIndex()
	if err != nil {
		return fmt.Errorf("pop index: %w", err)
	}

	currentKeyIndex, err := ctx.SquashedDictionaryManager.LastIndex()
	if err != nil {
		return fmt.Errorf("get last index: %w", err)
	}

	// todo(rodro): could previousKeyIndex be bigger than currentKeyIndex?
	indexDeltaMinusOne := currentKeyIndex - previousKeyIndex - 1
	mv := mem.MemoryValueFromUint(indexDeltaMinusOne)

	return vm.Memory.WriteToAddress(&indexDeltaPtr, &mv)
}

type ShouldContinueSquashLoop struct {
	ShouldContinue CellRefer
}

func (hint *ShouldContinueSquashLoop) String() string {
	return "ShouldContinueSquashLoop"
}

func (hint *ShouldContinueSquashLoop) Execute(vm *VM.VirtualMachine, ctx *HintRunnerContext) error {
	shouldContinuePtr, err := hint.ShouldContinue.Get(vm)
	if err != nil {
		return fmt.Errorf("get should continue address: %w", err)
	}

	var shouldContinueLoop f.Element
	if lastIndices, err := ctx.SquashedDictionaryManager.LastIndices(); err == nil && len(lastIndices) <= 1 {
		shouldContinueLoop.SetOne()
	} else if err != nil {
		return fmt.Errorf("get last indices: %w", err)
	}

	mv := mem.MemoryValueFromFieldElement(&shouldContinueLoop)
	return vm.Memory.WriteToAddress(&shouldContinuePtr, &mv)
}

type GetNextDictKey struct {
	NextKey CellRefer
}

func (hint *GetNextDictKey) String() string {
	return "GetNextDictKey"
}

func (hint *GetNextDictKey) Execute(vm *VM.VirtualMachine, ctx *HintRunnerContext) error {
	nextKeyAddr, err := hint.NextKey.Get(vm)
	if err != nil {
		return fmt.Errorf("get next key address: %w", err)
	}

	nextKey, err := ctx.SquashedDictionaryManager.PopKey()
	if err != nil {
		return fmt.Errorf("pop key: %w", err)
	}

	mv := mem.MemoryValueFromFieldElement(&nextKey)
	return vm.Memory.WriteToAddress(&nextKeyAddr, &mv)
}
