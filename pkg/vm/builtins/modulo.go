package builtins

import (
	"fmt"
	"math/big"
	
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
	AddOp Operation = "add"
	SubOp Operation = "sub"
	MulOp Operation = "mul"
	DivModOp Operation = "divmod"
)

type ModBuiltinRunner struct {
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

func (*ModBuiltinRunner) New(builtinType ModBuiltinType, included bool, instanceDef ModInstanceDef) *ModBuiltinRunner {
	shift := new(big.Int).Lsh(big.NewInt(1), uint(instanceDef.wordBitLen))
	shiftPowers := [N_WORDS]big.Int{}
	for i := 0; i < N_WORDS; i++ {
		shiftPowers[i] = *new(big.Int).Exp(shift, big.NewInt(int64(i)), nil)
	}
	return &ModBuiltinRunner{
		builtinType:           builtinType,
		base:                  0,
		stop_ptr:              nil,
		instanceDef:           instanceDef,
		included: 			   included,
		zeroSegmentIndex:      0,
		zeroSegmentSize:       max(N_WORDS, instanceDef.batchSize * 3),
		shift:                 *shift,
		shiftPowers: 		   shiftPowers,
	}
}

func (mbr *ModBuiltinRunner) NewAddMod(instanceDef *ModInstanceDef, included bool) *ModBuiltinRunner {
	return mbr.New(Add, included, *instanceDef)
}

func (mbr *ModBuiltinRunner) NewMulMod(instanceDef *ModInstanceDef, included bool) *ModBuiltinRunner {
	return mbr.New(Mul, included, *instanceDef)
}

func (mbr *ModBuiltinRunner) Name() string {
	switch mbr.builtinType {
		case Add:
			return "add_mod_builtin"
		case Mul:
			return "mul_mod_builtin"
		default:
			return "unknown"
	}
}

func (mbr *ModBuiltinRunner) ReadNWordsValue(memory *memory.Memory, addr memory.MemoryAddress) ([N_WORDS]fp.Element, big.Int, error) {
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

func (m *ModBuiltinRunner) readInputs(mem *memory.Memory, addr memory.MemoryAddress) (ModBuiltinInputs, error) {
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

func (mbr *ModBuiltinRunner) ComputeValue(memory memory.Memory, valuesPtr memory.MemoryAddress, offsetsPtr memory.MemoryAddress, indexInBatch uint, index uint) (big.Int, error) {
	newOffsetPtr, err := offsetsPtr.AddOffset(int16(index + 3 * indexInBatch))
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

func (mbr *ModBuiltinRunner) ReadMemoryVars(memory memory.Memory, valuesPtr memory.MemoryAddress, offsetsPtr memory.MemoryAddress, indexInBatch uint) (big.Int, big.Int, big.Int, error) {
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

