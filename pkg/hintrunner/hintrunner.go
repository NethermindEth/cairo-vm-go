package hintrunner

import (
	"fmt"

	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

// Used to keep track of all dictionaries
type Dictionary struct {
	// The data contained on a dictionary
	data map[f.Element]*mem.MemoryValue
	// Unique id assigned at the moment of creation
	idx uint64
}

// Gets the memory value at certain key
func (d *Dictionary) At(key *f.Element) (*mem.MemoryValue, error) {
	if value, ok := d.data[*key]; ok {
		return value, nil
	}
	return nil, fmt.Errorf("no value for key %s", key)
}

// Given a key and a value, it sets the value at the given key
func (d *Dictionary) Set(key *f.Element, value *mem.MemoryValue) {
	d.data[*key] = value
}

// Returns the initialization number when the dictionary was created
func (d *Dictionary) InitNumber() uint64 {
	return d.idx
}

type DictionaryManager struct {
	// a map that links a segment index to a dictionary
	dictionaries map[uint64]Dictionary
}

// It creates a new segment which will hold dictionary values. It links this
// segment with the current dictionary and returns the address that points
// to the start of this segment
func (dm *DictionaryManager) NewDictionary(vm *VM.VirtualMachine) mem.MemoryAddress {
	newDictAddr := vm.Memory.AllocateEmptySegment()
	dm.dictionaries[newDictAddr.SegmentIndex] = Dictionary{
		data: make(map[f.Element]*mem.MemoryValue),
		idx:  uint64(len(dm.dictionaries)),
	}
	return newDictAddr
}

// Given a memory address, it looks for the right dictionary using the segment index. If no
// segment is associated with the given segment index, it errors
func (dm *DictionaryManager) GetDictionary(dictAddr *mem.MemoryAddress) (Dictionary, error) {
	dict, ok := dm.dictionaries[dictAddr.SegmentIndex]
	if !ok {
		return Dictionary{}, fmt.Errorf("no dictionary at address %s", dictAddr)
	}
	return dict, nil
}

// Given a memory address and a key it returns the value held at that position. The address is used
// to locate the correct dictionary and the key to index on it
func (dm *DictionaryManager) At(dictAddr *mem.MemoryAddress, key *f.Element) (*mem.MemoryValue, error) {
	if dict, ok := dm.dictionaries[dictAddr.SegmentIndex]; ok {
		return dict.At(key)
	}
	return nil, fmt.Errorf("no dictionary at address %s", dictAddr)
}

// Given a memory address,a key and a value it stores the value at the correct position.
func (dm *DictionaryManager) Set(dictAddr *mem.MemoryAddress, key *f.Element, value *mem.MemoryValue) error {
	if dict, ok := dm.dictionaries[dictAddr.SegmentIndex]; ok {
		dict.Set(key, value)
		return nil
	}
	return fmt.Errorf("no dictionary at address %s", dictAddr)
}

// Used to keep track of squashed dictionaries
type SquashedDictionaryManager struct {
	// A map from each key to a list of indices where the key is present
	// the list in reversed order.
	// Note: The indices should be Felts, but current memory limitations
	// make it impossible to use an index that big so we use uint64 instead
	KeyToIndices map[f.Element][]uint64

	// A descending list of keys
	Keys []f.Element
}

// It adds another index to the list of indices associated to the given key
// If the key is not present, it creates a new entry
func (sdm *SquashedDictionaryManager) Insert(key *f.Element, index uint64) {
	keyIndex := *key
	if indices, ok := sdm.KeyToIndices[keyIndex]; ok {
		sdm.KeyToIndices[keyIndex] = append(indices, index)
	} else {
		sdm.KeyToIndices[keyIndex] = []uint64{index}
	}
}

// It returns the smallest key in the key list
func (sdm *SquashedDictionaryManager) LastKey() f.Element {
	return sdm.Keys[len(sdm.Keys)-1]
}

// It pops out the smallest key in the key list
func (sdm *SquashedDictionaryManager) PopKey() f.Element {
	key := sdm.LastKey()
	sdm.Keys = sdm.Keys[:len(sdm.Keys)-1]
	return key
}

// It returns the list of indices associated to the smallest key
func (sdm *SquashedDictionaryManager) LastIndices() []uint64 {
	key := sdm.LastKey()
	return sdm.KeyToIndices[key]
}

// It returns smallest index associated with the smallest key
func (sdm *SquashedDictionaryManager) LastIndex() uint64 {
	key := sdm.LastKey()
	indices := sdm.KeyToIndices[key]
	return indices[len(indices)-1]
}

// It pops out smallest index associated with the smallest key
func (sdm *SquashedDictionaryManager) PopIndex() uint64 {
	key := sdm.LastKey()
	indices := sdm.KeyToIndices[key]
	index := indices[len(indices)-1]
	sdm.KeyToIndices[key] = indices[:len(indices)-1]
	return index
}

// Global context to keep track of different results across different
// hints execution.
type HintRunnerContext struct {
	DictionaryManager         DictionaryManager
	SquashedDictionaryManager SquashedDictionaryManager
}

type HintRunner struct {
	// Execution context required by certain hints such as dictionaires
	context HintRunnerContext
	// A mapping from program counter to hint implementation
	hints map[uint64]Hinter
}

func NewHintRunner(hints map[uint64]Hinter) HintRunner {
	return HintRunner{
		context: HintRunnerContext{
			DictionaryManager{},
			SquashedDictionaryManager{},
		},
		hints: hints,
	}
}

func (hr *HintRunner) RunHint(vm *VM.VirtualMachine) error {
	hint := hr.hints[vm.Context.Pc.Offset]
	if hint == nil {
		return nil
	}

	err := hint.Execute(vm, &hr.context)
	if err != nil {
		return fmt.Errorf("execute hint %s: %v", hint, err)
	}
	return nil
}
