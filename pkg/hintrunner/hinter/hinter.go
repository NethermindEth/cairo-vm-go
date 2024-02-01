package hinter

import (
	"fmt"

	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

type Hinter interface {
	fmt.Stringer

	Execute(vm *VM.VirtualMachine, ctx *HintRunnerContext) error
}

// Used to keep track of all dictionaries data
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

// Used to manage dictionaries creation
type DictionaryManager struct {
	// a map that links a segment index to a dictionary
	dictionaries map[uint64]Dictionary
}

func InitializeDictionaryManager(ctx *HintRunnerContext) {
	if ctx.DictionaryManager.dictionaries == nil {
		ctx.DictionaryManager.dictionaries = make(map[uint64]Dictionary)
	}
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
	if ok {
		return dict, nil
	}
	return Dictionary{}, fmt.Errorf("no dictionary at address %s", dictAddr)
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

func InitializeSquashedDictionaryManager(ctx *HintRunnerContext) error {
	if ctx.SquashedDictionaryManager.KeyToIndices != nil ||
		ctx.SquashedDictionaryManager.Keys != nil {
		return fmt.Errorf("squashed dictionary manager already initialized")
	}
	ctx.SquashedDictionaryManager.KeyToIndices = make(map[f.Element][]uint64, 100)
	ctx.SquashedDictionaryManager.Keys = make([]f.Element, 0, 100)
	return nil
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
func (sdm *SquashedDictionaryManager) LastKey() (f.Element, error) {
	if len(sdm.Keys) == 0 {
		return f.Element{}, fmt.Errorf("no keys left")
	}
	return sdm.Keys[len(sdm.Keys)-1], nil
}

// It pops out the smallest key in the key list
func (sdm *SquashedDictionaryManager) PopKey() (f.Element, error) {
	key, err := sdm.LastKey()
	if err != nil {
		return key, err
	}

	sdm.Keys = sdm.Keys[:len(sdm.Keys)-1]
	return key, nil
}

// It returns the list of indices associated to the smallest key
func (sdm *SquashedDictionaryManager) LastIndices() ([]uint64, error) {
	key, err := sdm.LastKey()
	if err != nil {
		return nil, err
	}

	return sdm.KeyToIndices[key], nil
}

// It returns smallest index associated with the smallest key
func (sdm *SquashedDictionaryManager) LastIndex() (uint64, error) {
	key, err := sdm.LastKey()
	if err != nil {
		return 0, err
	}

	indices := sdm.KeyToIndices[key]
	if len(indices) == 0 {
		return 0, fmt.Errorf("no indices for key %s", &key)
	}

	return indices[len(indices)-1], nil
}

// It pops out smallest index associated with the smallest key
func (sdm *SquashedDictionaryManager) PopIndex() (uint64, error) {
	key, err := sdm.LastKey()
	if err != nil {
		return 0, err
	}

	indices := sdm.KeyToIndices[key]
	if len(indices) == 0 {
		return 0, fmt.Errorf("no indices for key %s", &key)
	}

	index := indices[len(indices)-1]
	sdm.KeyToIndices[key] = indices[:len(indices)-1]
	return index, nil
}

// Global context to keep track of different results across different
// hints execution.
type HintRunnerContext struct {
	DictionaryManager         DictionaryManager
	SquashedDictionaryManager SquashedDictionaryManager
	ExcludedArc               int
	// points towards free memory of a segment
	ConstantSizeSegment mem.MemoryAddress
}
