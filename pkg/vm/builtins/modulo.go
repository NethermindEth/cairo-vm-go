package builtins

import (
	"fmt"
	"math/big"

	"github.com/NethermindEth/cairo-vm-go/pkg/utils"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

const ModuloName = "Mod"
const cellsPerModulo = 16
const inputCellsPerModulo = 8
const instancesPerComponentModulo = 16

const OFFSETS_PTR_OFFSET = 5
const N_OFFSET = 6
const N_WORDS = 4
const CELLS_PER_MOD = 7
const FILL_MEMORY_MAX = 100000
const VALUES_PTR_OFFSET = 4

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
	addOp Operation = "add"
	subOp Operation = "invAdd"
	mulOp Operation = "mul"
	divOp Operation = "invMul"
)

type ModBuiltin struct {
	ratio          uint64
	modBuiltinType ModBuiltinType
	wordBitLen     uint64
	batchSize      uint64
	shift          big.Int
	shiftPowers    [N_WORDS]big.Int
}

func (k *ModBuiltin) CheckWrite(segment *memory.Segment, offset uint64, value *memory.MemoryValue) error {
	return nil
}

func (k *ModBuiltin) InferValue(segment *memory.Segment, offset uint64) error {
	return nil
}

func (k *ModBuiltin) String() string {
	if k.modBuiltinType == Add {
		return string(Add) + ModuloName
	} else {
		return string(Mul) + ModuloName
	}
}

func (k *ModBuiltin) GetAllocatedSize(segmentUsedSize uint64, vmCurrentStep uint64) (uint64, error) {
	return getBuiltinAllocatedSize(segmentUsedSize, vmCurrentStep, k.ratio, inputCellsPerKeccak, instancesPerComponentKeccak, cellsPerKeccak)
}

func (m *ModBuiltin) readNWordsValue(memory *memory.Memory, addr memory.MemoryAddress) ([N_WORDS]fp.Element, big.Int, error) {
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
		if word.Cmp(&m.shift) >= 0 {
			return words, *value, fmt.Errorf("Word exceeds mod builtin word bit length") // Replace with proper RunnerError handling
		}

		words[i] = wordFelt
		value.Add(value, new(big.Int).Mul(&word, &m.shiftPowers[i]))
	}

	return words, *value, nil
}

func (m *ModBuiltin) readInputs(mem *memory.Memory, addr memory.MemoryAddress) (*ModBuiltinInputs, error) {
	valuesPtrAddr, err := addr.AddOffset(int16(VALUES_PTR_OFFSET))
	if err != nil {
		return nil, err
	}
	valuesPtr, err := mem.ReadAsAddress(&valuesPtrAddr)
	if err != nil {
		return nil, err
	}
	offsetsPtrAddr, err := addr.AddOffset(int16(OFFSETS_PTR_OFFSET))
	if err != nil {
		return nil, err
	}
	offsetsPtr, err := mem.ReadAsAddress(&offsetsPtrAddr)
	if err != nil {
		return nil, err
	}
	nFelt, err := mem.ReadAsElement(addr.SegmentIndex, addr.Offset+N_OFFSET)
	if err != nil {
		return nil, err
	}
	n := nFelt.Uint64()
	if n < 1 {
		return nil, fmt.Errorf("ModuloBuiltin: n must be at least 1")
	}
	pValues, p, err := m.readNWordsValue(mem, addr)
	if err != nil {
		return nil, err
	}
	return &ModBuiltinInputs{
		p:          p,
		pValues:    pValues,
		valuesPtr:  valuesPtr,
		n:          n,
		offsetsPtr: offsetsPtr,
	}, nil
}

func (m *ModBuiltin) fillInputs(mem *memory.Memory, builtinPtr memory.MemoryAddress, inputs ModBuiltinInputs) error {
	if inputs.n > FILL_MEMORY_MAX {
		return fmt.Errorf("fill memory max exceeded")
	}

	nInstances, err := utils.SafeDiv(inputs.n, m.batchSize)
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
		newAddr, err := inputs.offsetsPtr.AddOffset(3 * int16(instance) * int16(m.batchSize))
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
		val := fp.NewElement(inputs.n - m.batchSize*uint64(instance))
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

func (m *ModBuiltin) WriteNWordsValue(mem *memory.Memory, addr memory.MemoryAddress, value *big.Int) error {
	for i := 0; i < N_WORDS; i++ {
		word := new(big.Int).Mod(value, &m.shift)
		modAddr, err := addr.AddOffset(int16(i))
		if err != nil {
			return err
		}
		mv := memory.MemoryValueFromFieldElement(new(fp.Element).SetBigInt(word))
		if err := mem.WriteToAddress(&modAddr, &mv); err != nil {
			return err
		}
		value.Div(value, &m.shift)
	}
	if value.Sign() != 0 {
		return fmt.Errorf("WriteNWordsValueNotZero")
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
		_, value, err := m.readNWordsValue(mem, addr)
		if err != nil {
			return false, err
		}
		values = append(values, &value)
	}

	a, b, c := values[0], values[1], values[2]

	applyOp := func(a, b *big.Int, op Operation) big.Int {
		switch op {
		case addOp:
			return *new(big.Int).Add(a, b)
		case subOp:
			return *new(big.Int).Sub(a, b)
		case mulOp:
			return *new(big.Int).Mul(a, b)
		case divOp:
			return *new(big.Int).Div(a, b)
		default:
			return *new(big.Int)
		}
	}

	switch {
	case a != nil && b != nil && c == nil:
		value := applyOp(a, b, op)
		value.Mod(&value, &inputs.p)
		if err := m.WriteNWordsValue(mem, addresses[2], &value); err != nil {
			return false, err
		}
		return true, nil
	case a != nil && b == nil && c != nil:
		value := applyOp(c, a, invOp)
		value.Mod(&value, &inputs.p)
		if err := m.WriteNWordsValue(mem, addresses[1], &value); err != nil {
			return false, err
		}
		return true, nil
	case a == nil && b != nil && c != nil:
		value := applyOp(c, b, invOp)
		value.Mod(&value, &inputs.p)
		if err := m.WriteNWordsValue(mem, addresses[0], &value); err != nil {
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
	addModBuiltinRunner, ok := addModBuiltinSegment.BuiltinRunner.(*ModBuiltin)
	if !ok {
		return fmt.Errorf("addModBuiltinRunner is not a ModBuiltin")
	}
	mulModBuiltinRunner, ok := mulModBuiltinSegment.BuiltinRunner.(*ModBuiltin)
	if !ok {
		return fmt.Errorf("mulModBuiltinRunner is not a ModBuiltin")
	}
	if addModBuiltinRunner.wordBitLen != mulModBuiltinRunner.wordBitLen {
		return fmt.Errorf("AddMod and MulMod wordBitLen mismatch")
	}

	addModBuiltinInputs, err := addModBuiltinRunner.readInputs(mem, addModBuiltinAddr)
	if err != nil {
		return err
	}
	addModBuiltinRunner.fillInputs(mem, addModBuiltinAddr, *addModBuiltinInputs)
	addModBuiltinRunner.fillOffsets(mem, addModBuiltinInputs.offsetsPtr, nAddModsIndex, addModBuiltinInputs.n-nAddModsIndex)

	mulModBuiltinInputs, err := mulModBuiltinRunner.readInputs(mem, mulModBuiltinAddr)
	if err != nil {
		return err
	}
	mulModBuiltinRunner.fillInputs(mem, mulModBuiltinAddr, *mulModBuiltinInputs)
	mulModBuiltinRunner.fillOffsets(mem, mulModBuiltinInputs.offsetsPtr, nMulModsIndex, mulModBuiltinInputs.n-nMulModsIndex)

	addModIndex, mulModIndex := uint64(0), uint64(0)
	for addModIndex < nAddModsIndex || mulModIndex < nMulModsIndex {
		ok, err := addModBuiltinRunner.fillValue(mem, *addModBuiltinInputs, int(addModIndex), addOp, subOp)
		if err != nil {
			return err
		}
		if addModIndex < nAddModsIndex && ok {
			addModIndex++
			continue
		}
		ok, err = mulModBuiltinRunner.fillValue(mem, *mulModBuiltinInputs, int(mulModIndex), mulOp, divOp)
		if err != nil {
			return err
		}
		if mulModIndex < nMulModsIndex && ok {
			mulModIndex++
		}
	}

	// POTENTIALY: add n_computed_mul_gates features in the future

	return nil
}
