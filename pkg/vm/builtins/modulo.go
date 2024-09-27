package builtins

import (
	"fmt"
	"math/big"

	"github.com/NethermindEth/cairo-vm-go/pkg/utils"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

const ModuloName = "Mod"

// These are the offsets in the array, which is used here as ModBuiltinInputs :
// INPUT_NAMES = [
//
//	"p0",
//	"p1",
//	"p2",
//	"p3",
//	"values_ptr",
//	"offsets_ptr",
//	"n",
//
// ]
const VALUES_PTR_OFFSET = 4
const OFFSETS_PTR_OFFSET = 5
const N_OFFSET = 6

// This is the number of felts in a UInt384 struct
const N_WORDS = 4

// number of memory cells per modulo builtin
// 4(felts) + 1(values_ptr) + 1(offsets_ptr) + 1(n) = 7
const CELLS_PER_MOD = 7

// The maximum n value that the function fill_memory accepts
const MAX_N = 100000

// Represents a 384-bit unsigned integer d0 + 2**96 * d1 + 2**192 * d2 + 2**288 * d3
// where each di is in [0, 2**96).
//
// struct UInt384 {
//     d0: felt,
//     d1: felt,
//     d2: felt,
//     d3: felt,
// }
// Instead of introducing UInt384, we use [N_WORDS]fp.Element to represent the 384-bit integer.

type ModBuiltinInputs struct {
	// The modulus.
	p       big.Int
	pValues [N_WORDS]fp.Element
	// A pointer to input values, the intermediate results and the output.
	valuesPtr memory.MemoryAddress
	// A pointer to offsets inside the values array, defining the circuit.
	// The offsets array should contain 3 * n elements.
	offsetsPtr memory.MemoryAddress
	// The number of operations to perform.
	n uint64
}

type ModBuiltinType string

const (
	Add ModBuiltinType = "Add"
	Mul ModBuiltinType = "Mul"
)

type ModBuiltin struct {
	ratio uint64
	// Add | Mul
	modBuiltinType ModBuiltinType
	// number of bits in a word
	wordBitLen uint64
	batchSize  uint64
	// shift by the number of bits present in a word
	shift big.Int
	// powers required to do the corresponding shift
	shiftPowers [N_WORDS]big.Int
	// k value that bounds p when finding unknown value in fillValue function
	kBound *big.Int
}

func NewModBuiltin(ratio uint64, wordBitLen uint64, batchSize uint64, modBuiltinType ModBuiltinType) *ModBuiltin {
	shift := new(big.Int).Lsh(big.NewInt(1), uint(wordBitLen))
	shiftPowers := [N_WORDS]big.Int{}
	shiftPowers[0] = *big.NewInt(1)
	for i := 1; i < N_WORDS; i++ {
		shiftPowers[i].Mul(&shiftPowers[i-1], shift)
	}
	kBound := big.NewInt(2)
	if modBuiltinType == Mul {
		kBound = nil
	}
	return &ModBuiltin{
		ratio:          ratio,
		modBuiltinType: modBuiltinType,
		wordBitLen:     wordBitLen,
		batchSize:      batchSize,
		shift:          *shift,
		shiftPowers:    shiftPowers,
		kBound:         kBound,
	}
}

// TODO: Implement CheckWrite
func (m *ModBuiltin) CheckWrite(segment *memory.Segment, offset uint64, value *memory.MemoryValue) error {
	return nil
}

// TODO: Implement CheckRead
func (m *ModBuiltin) InferValue(segment *memory.Segment, offset uint64) error {
	return nil
}

func (m *ModBuiltin) String() string {
	return string(m.modBuiltinType) + ModuloName
}

func (m *ModBuiltin) GetAllocatedSize(segmentUsedSize uint64, vmCurrentStep uint64) (uint64, error) {
	return 0, nil
}

// Reads N_WORDS from memory, starting at address = addr.
// Returns the words and the value if all words are in memory.
// Verifies that all words are integers and are bounded by 2**wordBitLen.
func (m *ModBuiltin) readNWordsValue(memory *memory.Memory, addr memory.MemoryAddress) ([N_WORDS]fp.Element, *big.Int, error) {
	var words [N_WORDS]fp.Element
	value := new(big.Int).SetInt64(0)

	for i := 0; i < N_WORDS; i++ {
		newAddr, err := addr.AddOffset(int16(i))
		if err != nil {
			return [N_WORDS]fp.Element{}, nil, err
		}

		wordFelt, err := memory.ReadAsElement(newAddr.SegmentIndex, newAddr.Offset)
		if err != nil {
			return [N_WORDS]fp.Element{}, nil, err
		}

		var word big.Int
		wordFelt.BigInt(&word)
		if word.Cmp(&m.shift) >= 0 {
			return [N_WORDS]fp.Element{}, nil, fmt.Errorf("expected integer at address %d:%d to be smaller than 2^%d. Got: %s", newAddr.SegmentIndex, newAddr.Offset, m.wordBitLen, word.String())
		}

		words[i] = wordFelt
		value = new(big.Int).Add(value, new(big.Int).Mul(&word, &m.shiftPowers[i]))
	}

	return words, value, nil
}

// Reads the inputs to the builtin (p, p_values, values_ptr, offsets_ptr, n) from the memory at address = addr.
// Returns an instance of ModBuiltinInputs and asserts that it exists in memory.
// If `read_n` is false, avoid reading and validating the value of 'n'.
func (m *ModBuiltin) readInputs(mem *memory.Memory, addr memory.MemoryAddress, read_n bool) (ModBuiltinInputs, error) {
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
	n := uint64(0)
	if read_n {
		nFelt, err := mem.ReadAsElement(addr.SegmentIndex, addr.Offset+N_OFFSET)
		if err != nil {
			return ModBuiltinInputs{}, err
		}
		n = nFelt.Uint64()
		if n < 1 {
			return ModBuiltinInputs{}, fmt.Errorf("moduloBuiltin: Expected n >= 1. Got: %d", n)
		}
	}
	pValues, p, err := m.readNWordsValue(mem, addr)
	if err != nil {
		return ModBuiltinInputs{}, err
	}
	return ModBuiltinInputs{
		p:          *p,
		pValues:    pValues,
		valuesPtr:  valuesPtr,
		n:          n,
		offsetsPtr: offsetsPtr,
	}, nil
}

// Fills the inputs to the instances of the builtin given the inputs to the first instance.
func (m *ModBuiltin) fillInputs(mem *memory.Memory, builtinPtr memory.MemoryAddress, inputs ModBuiltinInputs) error {
	if inputs.n > MAX_N {
		return fmt.Errorf("fill memory max exceeded")
	}

	nInstances, err := utils.SafeDivUint64(inputs.n, m.batchSize)
	if err != nil {
		return err
	}

	for instance := 1; instance < int(nInstances); instance++ {
		instancePtr, err := builtinPtr.AddOffset(int16(instance * CELLS_PER_MOD))
		if err != nil {
			return err
		}

		// Filling the 4 values of a UInt384 struct
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

		// This denotes the number of operations left
		// n for new instance = original n - batch_size * (number of instances passed)
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

// Copies the first offsets into memory, nCopies times.
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

// Given a value, writes its n_words to memory, starting at address = addr.
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

// Fills a value in the values table, if exactly one value is missing.
// Returns true on success or if all values are already known.
// Given known, res, p fillValue tries to compute the minimal integer operand x which
// satisfies the equation op(x,known) = res + k*p for some k in {0,1,...,self.k_bound-1}.
func (m *ModBuiltin) fillValue(mem *memory.Memory, inputs ModBuiltinInputs, index int, op ModBuiltinType) (bool, error) {
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
		// do not check for error, as the value might not be in memory
		_, value, _ := m.readNWordsValue(mem, addr)
		values = append(values, value)
	}

	a, b, c := values[0], values[1], values[2]

	// 2 ** 384 (max value that can be stored in 4 felts)
	intLim := new(big.Int).Lsh(big.NewInt(1), uint(m.wordBitLen)*N_WORDS)
	kBound := m.kBound
	if kBound == nil {
		kBound = new(big.Int).Set(intLim)
	}

	switch {
	case a != nil && b != nil && c == nil:
		var value big.Int
		if op == Add {
			value = *new(big.Int).Add(a, b)
		} else {
			value = *new(big.Int).Mul(a, b)
		}
		if new(big.Int).Sub(&value, new(big.Int).Mul((new(big.Int).Sub(kBound, big.NewInt(1))), &inputs.p)).Cmp(new(big.Int).Sub(intLim, big.NewInt(1))) == 1 {
			return false, fmt.Errorf("%s builtin: Expected a %s b - %d * p <= %d", m.String(), m.modBuiltinType, kBound.Sub(kBound, big.NewInt(1)), intLim.Sub(intLim, big.NewInt(1)))
		}
		if value.Cmp(new(big.Int).Mul(kBound, &inputs.p)) < 0 {
			value.Mod(&value, &inputs.p)
		} else {
			value.Sub(&value, new(big.Int).Mul(new(big.Int).Sub(kBound, big.NewInt(1)), &inputs.p))
		}
		if err := m.writeNWordsValue(mem, addresses[2], value); err != nil {
			return false, err
		}
		return true, nil
	case a != nil && b == nil && c != nil:
		var value big.Int
		if op == Add {
			// Right now only k = 2 is an option, hence as we stated above that x + known can only take values
			// from res to res + (k - 1) * p, hence known <= res + p
			if a.Cmp(new(big.Int).Add(c, &inputs.p)) > 0 {
				return false, fmt.Errorf("%s builtin: addend greater than sum + p: %d > %d + %d", m.String(), a, c, &inputs.p)
			} else {
				if a.Cmp(c) <= 0 {
					value = *new(big.Int).Sub(c, a)
				} else {
					value = *new(big.Int).Sub(c.Add(c, &inputs.p), a)
				}
			}
		} else {
			x, _, gcd := utils.Igcdex(a, &inputs.p)
			// if gcd != 1, the known value is 0, in which case the res must be 0
			if gcd.Cmp(big.NewInt(1)) != 0 {
				value = *new(big.Int).Div(&inputs.p, &gcd)
			} else {
				value = *new(big.Int).Mul(c, &x)
				value = *value.Mod(&value, &inputs.p)
				tmpK, err := utils.SafeDiv(new(big.Int).Sub(new(big.Int).Mul(a, &value), c), &inputs.p)
				if err != nil {
					return false, err
				}
				if tmpK.Cmp(kBound) >= 0 {
					return false, fmt.Errorf("%s builtin: ((%d * q) - %d) / %d > %d for any q > 0, such that %d * q = %d (mod %d) ", m.String(), a, c, &inputs.p, kBound, a, c, &inputs.p)
				}
				if tmpK.Cmp(big.NewInt(0)) < 0 {
					value = *value.Add(&value, new(big.Int).Mul(&inputs.p, new(big.Int).Div(new(big.Int).Sub(a, new(big.Int).Sub(&tmpK, big.NewInt(1))), a)))
				}
			}
		}
		if err := m.writeNWordsValue(mem, addresses[1], value); err != nil {
			return false, err
		}
		return true, nil
	case a == nil && b != nil && c != nil:
		var value big.Int
		if op == Add {
			// Right now only k = 2 is an option, hence as we stated above that x + known can only take values
			// from res to res + (k - 1) * p, hence known <= res + p
			if b.Cmp(new(big.Int).Add(c, &inputs.p)) > 0 {
				return false, fmt.Errorf("%s builtin: addend greater than sum + p: %d > %d + %d", m.String(), b, c, &inputs.p)
			} else {
				if b.Cmp(c) <= 0 {
					value = *new(big.Int).Sub(c, b)
				} else {
					value = *new(big.Int).Sub(c.Add(c, &inputs.p), b)
				}
			}
		} else {
			x, _, gcd := utils.Igcdex(b, &inputs.p)
			// if gcd != 1, the known value is 0, in which case the res must be 0
			if gcd.Cmp(big.NewInt(1)) != 0 {
				value = *new(big.Int).Div(&inputs.p, &gcd)
			} else {
				value = *new(big.Int).Mul(c, &x)
				value = *value.Mod(&value, &inputs.p)
				tmpK, err := utils.SafeDiv(new(big.Int).Sub(new(big.Int).Mul(b, &value), c), &inputs.p)
				if err != nil {
					return false, err
				}
				if tmpK.Cmp(kBound) >= 0 {
					return false, fmt.Errorf("%s builtin: ((%d * q) - %d) / %d > %d for any q > 0, such that %d * q = %d (mod %d) ", m.String(), b, c, &inputs.p, kBound, b, c, &inputs.p)
				}
				if tmpK.Cmp(big.NewInt(0)) < 0 {
					value = *value.Add(&value, new(big.Int).Mul(&inputs.p, new(big.Int).Div(new(big.Int).Sub(b, new(big.Int).Sub(&tmpK, big.NewInt(1))), b)))
				}
			}
		}
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

// Fills the memory with inputs to the builtin instances based on the inputs to the
// first instance, pads the offsets table to fit the number of operations written in the
// input to the first instance, and calculates missing values in the values table.
//
// The number of operations written to the input of the first instance n should be at
// least n and a multiple of batch_size. Previous offsets are copied to the end of the
// offsets table to make its length 3n'.
func FillMemory(mem *memory.Memory, addModBuiltinAddr memory.MemoryAddress, nAddModsIndex uint64, mulModBuiltinAddr memory.MemoryAddress, nMulModsIndex uint64) error {
	if nAddModsIndex > MAX_N {
		return fmt.Errorf("AddMod builtin: n must be <= {MAX_N}")
	}
	if nMulModsIndex > MAX_N {
		return fmt.Errorf("MulMod builtin: n must be <= {MAX_N}")
	}

	addModBuiltinSegment, ok := mem.FindSegmentWithBuiltin("AddMod")
	if !ok {
		return fmt.Errorf("AddMod builtin segment doesn't exist")
	}
	mulModBuiltinSegment, ok := mem.FindSegmentWithBuiltin("MulMod")
	if !ok {
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

	addModBuiltinInputs, err := addModBuiltinRunner.readInputs(mem, addModBuiltinAddr, true)
	if err != nil {
		return err
	}
	if err := addModBuiltinRunner.fillInputs(mem, addModBuiltinAddr, addModBuiltinInputs); err != nil {
		return err
	}
	if err := addModBuiltinRunner.fillOffsets(mem, addModBuiltinInputs.offsetsPtr, nAddModsIndex, addModBuiltinInputs.n-nAddModsIndex); err != nil {
		return err
	}

	mulModBuiltinInputs, err := mulModBuiltinRunner.readInputs(mem, mulModBuiltinAddr, true)
	if err != nil {
		return err
	}
	if err := mulModBuiltinRunner.fillInputs(mem, mulModBuiltinAddr, mulModBuiltinInputs); err != nil {
		return err
	}
	if err := mulModBuiltinRunner.fillOffsets(mem, mulModBuiltinInputs.offsetsPtr, nMulModsIndex, mulModBuiltinInputs.n-nMulModsIndex); err != nil {
		return err
	}

	addModIndex, mulModIndex := uint64(0), uint64(0)
	for addModIndex < nAddModsIndex {
		ok, err := addModBuiltinRunner.fillValue(mem, addModBuiltinInputs, int(addModIndex), Add)
		if err != nil {
			return err
		}
		if ok {
			addModIndex++
		}
	}

	for mulModIndex < nMulModsIndex {
		ok, err = mulModBuiltinRunner.fillValue(mem, mulModBuiltinInputs, int(mulModIndex), Mul)
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
