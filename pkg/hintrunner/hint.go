package hintrunner

import (
	"fmt"
	"sort"

	"github.com/holiman/uint256"

	"github.com/NethermindEth/cairo-vm-go/pkg/safemath"
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
	mask := MaxU128()

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

	sqrt := valueFelt.Sqrt(valueFelt)

	dstAddr, err := hint.dst.Get(vm)
	if err != nil {
		return fmt.Errorf("get destination cell: %w", err)
	}

	dstVal := mem.MemoryValueFromFieldElement(sqrt)
	err = vm.Memory.WriteToAddress(&dstAddr, &dstVal)
	if err != nil {
		return fmt.Errorf("write cell: %w", err)
	}
	return nil
}

type AllocFelt252Dict struct {
	SegmentArenaPtr ResOperander
}

func (hint *AllocFelt252Dict) String() string {
	return "AllocFelt252Dict"
}
func (hint *AllocFelt252Dict) Execute(vm *VM.VirtualMachine, ctx *HintRunnerContext) error {
	arenaPtr, err := ResolveAsAddress(vm, hint.SegmentArenaPtr)
	if err != nil {
		return fmt.Errorf("resolve segment arena pointer: %w", err)
	}

	// find for the amount of initialized dicts
	initializedDictsOffset, overflow := safemath.SafeOffset(arenaPtr.Offset, -2)
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
	segmentInfoOffset, overflow := safemath.SafeOffset(arenaPtr.Offset, -3)
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
	return "FeltGetSegmentArenaIndex"
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
		safemath.Reverse(val)
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
	if biggestKey.Cmp(&safemath.Max128) > 0 {
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
	firstKey := ctx.SquashedDictionaryManager.LastKey()
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

	lastIndex := f.NewElement(
		ctx.SquashedDictionaryManager.LastIndex(),
	)
	mv := mem.MemoryValueFromFieldElement(&lastIndex)

	return vm.Memory.WriteToAddress(&rangeCheckPtr, &mv)
}
