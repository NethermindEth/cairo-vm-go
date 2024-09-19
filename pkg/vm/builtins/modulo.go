package builtins

import (
	"fmt"
	"math/big"

	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

const VALUES_PTR_OFFSET = 4
const OFFSETS_PTR_OFFSET = 5
const N_OFFSET = 6
const N_WORDS = 4
const CELLS_PER_MOD = 7
const FILL_MEMORY_MAX = 100000

type ModInstanceDef struct {
	ratio      *uint32
	wordBitLen uint32
	batchSize  uint
}

func (*ModInstanceDef) New(ratio *uint32, wordBitLen uint32, batchSize uint) *ModInstanceDef {
	return &ModInstanceDef{
		ratio:      ratio,
		wordBitLen: wordBitLen,
		batchSize:  batchSize,
	}
}

type ModBuiltinInputs struct {
	p          big.Int
	pValues    [N_WORDS]fp.Element
	valuesPtr  memory.MemoryAddress
	offsetsPtr memory.MemoryAddress
	n          uint64
}

type ModBuiltinType string

const (
	Add ModBuiltinType = "Add"
	Mul ModBuiltinType = "Mul"
)

type Operation string

const (
	AddOp    Operation = "add"
	SubOp    Operation = "sub"
	MulOp    Operation = "mul"
	DivModOp Operation = "divmod"
)

type ModBuiltin struct {
	builtinType      ModBuiltinType
	base             uint64
	stop_ptr         *uint
	instanceDef      ModInstanceDef
	included         bool
	zeroSegmentIndex uint
	zeroSegmentSize  uint
	shift            big.Int
	shiftPowers      [N_WORDS]big.Int
}

func (m *ModBuiltin) CheckWrite(segment *memory.Segment, offset uint64, value *memory.MemoryValue) error {
	return nil
}

func (m *ModBuiltin) InferValue(segment *memory.Segment, offset uint64) error {
	return nil
}

func (m *ModBuiltin) GetAllocatedSize(segmentUsedSize uint64, vmCurrentStep uint64) (uint64, error) {
	return 0, nil
}

func max(a, b uint) uint {
	if a > b {
		return a
	}
	return b
}

func (*ModBuiltin) New(builtinType ModBuiltinType, included bool, instanceDef ModInstanceDef) *ModBuiltin {
	shift := new(big.Int).Lsh(big.NewInt(1), uint(instanceDef.wordBitLen))
	shiftPowers := [N_WORDS]big.Int{}
	for i := 0; i < N_WORDS; i++ {
		shiftPowers[i] = *new(big.Int).Exp(shift, big.NewInt(int64(i)), nil)
	}
	return &ModBuiltin{
		builtinType:      builtinType,
		base:             0,
		stop_ptr:         nil,
		instanceDef:      instanceDef,
		included:         included,
		zeroSegmentIndex: 0,
		zeroSegmentSize:  max(N_WORDS, instanceDef.batchSize*3),
		shift:            *shift,
		shiftPowers:      shiftPowers,
	}
}

func (mbr *ModBuiltin) NewAddMod(instanceDef *ModInstanceDef, included bool) *ModBuiltin {
	return mbr.New(Add, included, *instanceDef)
}

func (mbr *ModBuiltin) NewMulMod(instanceDef *ModInstanceDef, included bool) *ModBuiltin {
	return mbr.New(Mul, included, *instanceDef)
}

func (mbr *ModBuiltin) String() string {
	switch mbr.builtinType {
	case Add:
		return "add_mod_builtin"
	case Mul:
		return "mul_mod_builtin"
	default:
		return "unknown"
	}
}

func (mbr *ModBuiltin) ReadNWordsValue(memory *memory.Memory, addr memory.MemoryAddress) ([N_WORDS]fp.Element, big.Int, error) {
	var words [N_WORDS]fp.Element
	value := new(big.Int).SetInt64(0)

	for i := 0; i < N_WORDS; i++ {
		newAddr, err := addr.AddOffset(int16(i))
		if err != nil {
			return words, *value, err
		}

		wordFelt, err := memory.ReadAsElement(newAddr.SegmentIndex, newAddr.Offset)
		if err != nil {
			return words, *value, err
		}

		var word big.Int
		wordFelt.BigInt(&word)
		if word.Cmp(&mbr.shift) >= 0 {
			return words, *value, fmt.Errorf("word exceeds mod builtin word bit length")
		}

		words[i] = wordFelt
		value.Add(value, new(big.Int).Mul(&word, &mbr.shiftPowers[i]))
	}

	return words, *value, nil
}

func (m *ModBuiltin) readInputs(mem *memory.Memory, addr memory.MemoryAddress) (ModBuiltinInputs, error) {
	valuesPtrAddr, err := addr.AddOffset(int16(VALUES_PTR_OFFSET))
	if err != nil {
		return ModBuiltinInputs{}, err
	}
	valuesPtr, err := mem.ReadAsAddress(&valuesPtrAddr)
	if err != nil {
		return ModBuiltinInputs{}, err
	}
	offsetsPtrAddr, err := addr.AddOffset(int16(OFFSETS_PTR_OFFSET))
	if err != nil {
		return ModBuiltinInputs{}, err
	}
	offsetsPtr, err := mem.ReadAsAddress(&offsetsPtrAddr)
	if err != nil {
		return ModBuiltinInputs{}, err
	}
	nFelt, err := mem.ReadAsElement(addr.SegmentIndex, addr.Offset+N_OFFSET)
	if err != nil {
		return ModBuiltinInputs{}, err
	}
	n := nFelt.Uint64()
	if n < 1 {
		return ModBuiltinInputs{}, fmt.Errorf("moduloBuiltin: n must be at least 1")
	}
	pValues, p, err := m.ReadNWordsValue(mem, addr)
	if err != nil {
		return ModBuiltinInputs{}, err
	}
	return ModBuiltinInputs{
		p:          p,
		pValues:    pValues,
		valuesPtr:  valuesPtr,
		n:          n,
		offsetsPtr: offsetsPtr,
	}, nil
}

func (mbr *ModBuiltin) ComputeValue(memory memory.Memory, valuesPtr memory.MemoryAddress, offsetsPtr memory.MemoryAddress, indexInBatch uint, index uint) (big.Int, error) {
	newOffsetPtr, err := offsetsPtr.AddOffset(int16(index + 3*indexInBatch))
	if err != nil {
		return big.Int{}, err
	}
	offset, err := memory.ReadFromAddress(&newOffsetPtr)
	if err != nil {
		return big.Int{}, err
	}
	offsetFelt, err := offset.Uint64()
	if err != nil {
		return big.Int{}, err
	}
	valueAddr, err := valuesPtr.AddOffset(int16(offsetFelt))
	if err != nil {
		return big.Int{}, err
	}
	_, value, err := mbr.ReadNWordsValue(&memory, valueAddr)
	if err != nil {
		return big.Int{}, err
	}
	return value, nil
}

func (mbr *ModBuiltin) ReadMemoryVars(memory memory.Memory, valuesPtr memory.MemoryAddress, offsetsPtr memory.MemoryAddress, indexInBatch uint) (big.Int, big.Int, big.Int, error) {
	a, err := mbr.ComputeValue(memory, valuesPtr, offsetsPtr, indexInBatch, 0)
	if err != nil {
		return big.Int{}, big.Int{}, big.Int{}, err
	}
	b, err := mbr.ComputeValue(memory, valuesPtr, offsetsPtr, indexInBatch, 1)
	if err != nil {
		return big.Int{}, big.Int{}, big.Int{}, err
	}
	c, err := mbr.ComputeValue(memory, valuesPtr, offsetsPtr, indexInBatch, 2)
	if err != nil {
		return big.Int{}, big.Int{}, big.Int{}, err
	}
	return a, b, c, nil
}

func (m *ModBuiltin) fillInputs(mem *memory.Memory, builtinPtr memory.MemoryAddress, inputs ModBuiltinInputs) error {
	if inputs.n > FILL_MEMORY_MAX {
		return fmt.Errorf("fill memory max exceeded")
	}

	nInstances, err := utils.SafeDivUint64(inputs.n, uint64(m.instanceDef.batchSize))
	if err != nil {
		return err
	}

	for instance := 1; instance < int(nInstances); instance++ {
		instancePtr, err := builtinPtr.AddOffset(int16(instance * CELLS_PER_MOD))
		if err != nil {
			return err
		}

		for i := 0; i < N_WORDS; i++ {
			addr, err := instancePtr.AddOffset(int16(i))
			if err != nil {
				return err
			}
			mv := memory.MemoryValueFromFieldElement(&inputs.pValues[i])
			if err := mem.WriteToAddress(&addr, &mv); err != nil {
				return err
			}
		}

		addr, err := instancePtr.AddOffset(VALUES_PTR_OFFSET)
		if err != nil {
			return err
		}
		mv := memory.MemoryValueFromMemoryAddress(&inputs.valuesPtr)
		if err := mem.WriteToAddress(&addr, &mv); err != nil {
			return err
		}

		addr, err = instancePtr.AddOffset(OFFSETS_PTR_OFFSET)
		if err != nil {
			return err
		}
		newAddr, err := inputs.offsetsPtr.AddOffset(3 * int16(instance) * int16(m.instanceDef.batchSize))
		if err != nil {
			return err
		}
		mv = memory.MemoryValueFromMemoryAddress(&newAddr)
		if err := mem.WriteToAddress(&addr, &mv); err != nil {
			return err
		}

		addr, err = instancePtr.AddOffset(N_OFFSET)
		if err != nil {
			return err
		}
		val := fp.NewElement(inputs.n - uint64(m.instanceDef.batchSize)*uint64(instance))
		mv = memory.MemoryValueFromFieldElement(&val)
		if err := mem.WriteToAddress(&addr, &mv); err != nil {
			return err
		}
	}

	return nil
}

func (m *ModBuiltin) fillOffsets(mem *memory.Memory, offsetsPtr memory.MemoryAddress, index, nCopies uint64) error {
	if nCopies == 0 {
		return nil
	}

	for i := 0; i < 3; i++ {
		addr, err := offsetsPtr.AddOffset(int16(i))
		if err != nil {
			return err
		}

		offset, err := mem.ReadAsAddress(&addr)
		if err != nil {
			return err
		}

		for copyI := 0; copyI < int(nCopies); copyI++ {
			copyAddr, err := offsetsPtr.AddOffset(int16(3*(index+uint64(copyI)) + uint64(i)))
			if err != nil {
				return err
			}
			mv := memory.MemoryValueFromMemoryAddress(&offset)
			if err := mem.WriteToAddress(&copyAddr, &mv); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *ModBuiltin) writeNWordsValue(mem *memory.Memory, addr memory.MemoryAddress, value big.Int) error {
	for i := 0; i < N_WORDS; i++ {
		word := new(big.Int).Mod(&value, &m.shift)
		modAddr, err := addr.AddOffset(int16(i))
		if err != nil {
			return err
		}
		mv := memory.MemoryValueFromFieldElement(new(fp.Element).SetBigInt(word))
		if err := mem.WriteToAddress(&modAddr, &mv); err != nil {
			return err
		}
		value.Div(&value, &m.shift)
	}
	if value.Sign() != 0 {
		return fmt.Errorf("writeNWordsValue: value should be zero")
	}
	return nil
}

func (m *ModBuiltin) fillValue(mem *memory.Memory, inputs ModBuiltinInputs, index int, op, invOp Operation) (bool, error) {
	addresses := make([]memory.MemoryAddress, 0, 3)
	values := make([]*big.Int, 0, 3)

	for i := 0; i < 3; i++ {
		addr, err := inputs.offsetsPtr.AddOffset(int16(3*index + i))
		if err != nil {
			return false, err
		}
		offsetFelt, err := mem.ReadAsElement(addr.SegmentIndex, addr.Offset)
		if err != nil {
			return false, err
		}
		offset := offsetFelt.Uint64()
		addr, err = inputs.valuesPtr.AddOffset(int16(offset))
		if err != nil {
			return false, err
		}
		addresses = append(addresses, addr)
		_, value, err := m.ReadNWordsValue(mem, addr)
		if err != nil {
			return false, err
		}
		values = append(values, &value)
	}

	a, b, c := values[0], values[1], values[2]

	applyOp := func(a, b *big.Int, op Operation) big.Int {
		switch op {
		case AddOp:
			return *new(big.Int).Add(a, b)
		case SubOp:
			return *new(big.Int).Sub(a, b)
		case MulOp:
			return *new(big.Int).Mul(a, b)
		case DivModOp:
			return *new(big.Int).Div(a, b)
		default:
			return *new(big.Int)
		}
	}

	switch {
	case a != nil && b != nil && c == nil:
		value := applyOp(a, b, op)
		value.Mod(&value, &inputs.p)
		if err := m.writeNWordsValue(mem, addresses[2], value); err != nil {
			return false, err
		}
		return true, nil
	case a != nil && b == nil && c != nil:
		value := applyOp(c, a, invOp)
		value.Mod(&value, &inputs.p)
		if err := m.writeNWordsValue(mem, addresses[1], value); err != nil {
			return false, err
		}
		return true, nil
	case a == nil && b != nil && c != nil:
		value := applyOp(c, b, invOp)
		value.Mod(&value, &inputs.p)
		if err := m.writeNWordsValue(mem, addresses[0], value); err != nil {
			return false, err
		}
		return true, nil
	case a != nil && b != nil && c != nil:
		return true, nil
	default:
		return false, nil
	}
}

func FillMemory(mem *memory.Memory, addModBuiltinAddr memory.MemoryAddress, nAddModsIndex uint64, mulModBuiltinAddr memory.MemoryAddress, nMulModsIndex uint64) error {
	addModBuiltinSegment, ok := mem.FindSegmentWithBuiltin("AddMod")
	if ok {
		return fmt.Errorf("AddMod builtin segment doesn't exist")
	}
	mulModBuiltinSegment, ok := mem.FindSegmentWithBuiltin("MulMod")
	if ok {
		return fmt.Errorf("MulMod builtin segment doesn't exist")
	}
	addModBuiltin, ok := addModBuiltinSegment.BuiltinRunner.(*ModBuiltin)
	if !ok {
		return fmt.Errorf("addModBuiltin is not a ModBuiltin")
	}
	mulModBuiltin, ok := mulModBuiltinSegment.BuiltinRunner.(*ModBuiltin)
	if !ok {
		return fmt.Errorf("mulModBuiltin is not a ModBuiltin")
	}
	if addModBuiltin.instanceDef.wordBitLen != mulModBuiltin.instanceDef.wordBitLen {
		return fmt.Errorf("AddMod and MulMod wordBitLen mismatch")
	}

	addModBuiltinInputs, err := addModBuiltin.readInputs(mem, addModBuiltinAddr)
	if err != nil {
		return err
	}
	if err := addModBuiltin.fillInputs(mem, addModBuiltinAddr, addModBuiltinInputs); err != nil {
		return err
	}
	if err := addModBuiltin.fillOffsets(mem, addModBuiltinInputs.offsetsPtr, nAddModsIndex, addModBuiltinInputs.n-nAddModsIndex); err != nil {
		return err
	}

	mulModBuiltinInputs, err := mulModBuiltin.readInputs(mem, mulModBuiltinAddr)
	if err != nil {
		return err
	}
	if err := mulModBuiltin.fillInputs(mem, mulModBuiltinAddr, mulModBuiltinInputs); err != nil {
		return err
	}
	if err := mulModBuiltin.fillOffsets(mem, mulModBuiltinInputs.offsetsPtr, nMulModsIndex, mulModBuiltinInputs.n-nMulModsIndex); err != nil {
		return err
	}

	addModIndex, mulModIndex := uint64(0), uint64(0)
	for addModIndex < nAddModsIndex {
		ok, err := addModBuiltin.fillValue(mem, addModBuiltinInputs, int(addModIndex), AddOp, SubOp)
		if err != nil {
			return err
		}
		if ok {
			addModIndex++
		}
	}

	for mulModIndex < nMulModsIndex {
		ok, err = mulModBuiltin.fillValue(mem, mulModBuiltinInputs, int(mulModIndex), MulOp, DivModOp)
		if err != nil {
			return err
		}
		if ok {
			mulModIndex++
		}
	}
	// POTENTIALY: add n_computed_mul_gates features in the future

	return nil
}
