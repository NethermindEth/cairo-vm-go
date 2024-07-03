package hinter

import (
	"fmt"

	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

// Used to keep track of all Dictionaries data
type ZeroDictionary struct {
	// The Data contained in a dictionary
	Data *map[fp.Element]mem.MemoryValue
	// Default value for key not present in the dictionary
	DefaultValue *mem.MemoryValue
	// first free offset in memory segment of dictionary
	FreeOffset *uint64
}

// Gets the memory value at certain key
func (d *ZeroDictionary) at(key fp.Element) (mem.MemoryValue, error) {
	if value, ok := (*d.Data)[key]; ok {
		return value, nil
	}
	if *d.DefaultValue != mem.UnknownValue {
		return *d.DefaultValue, nil
	}
	return mem.UnknownValue, fmt.Errorf("no value for key: %v", key)
}

// Given a key and a value, it sets the value at the given key
func (d *ZeroDictionary) set(key fp.Element, value mem.MemoryValue) {
	(*d.Data)[key] = value
}

// Given a incrementBy value, it increments the freeOffset field of dictionary by it
func (d *ZeroDictionary) incrementFreeOffset(freeOffset uint64) {
	*d.FreeOffset += freeOffset
}

// Given a freeOffset value, it sets the freeOffset field of dictionary to it
func (d *ZeroDictionary) setFreeOffset(freeOffset uint64) {
	*d.FreeOffset = freeOffset
}

// Used to manage dictionaries creation
type ZeroDictionaryManager struct {
	// a map that links a segment index to a dictionary
	Dictionaries map[uint64]ZeroDictionary
}

func NewZeroDictionaryManager() ZeroDictionaryManager {
	return ZeroDictionaryManager{
		Dictionaries: make(map[uint64]ZeroDictionary),
	}
}

// It creates a new segment which will hold dictionary values. It links this
// segment with the current dictionary and returns the address that points
// to the start of this segment. initial dictionary data is set from the data argument.
func (dm *ZeroDictionaryManager) NewDictionary(vm *VM.VirtualMachine, data map[fp.Element]mem.MemoryValue) mem.MemoryAddress {
	newDictAddr := vm.Memory.AllocateEmptySegment()
	freeOffset := uint64(0)
	dm.Dictionaries[newDictAddr.SegmentIndex] = ZeroDictionary{
		Data:         &data,
		DefaultValue: &mem.UnknownValue,
		FreeOffset:   &freeOffset,
	}
	return newDictAddr
}

// It creates a new segment which will hold dictionary values. It links this
// segment with the current dictionary and returns the address that points
// to the start of this segment. If key not present in the dictionary during
// querying the defaultValue will be returned instead.
func (dm *ZeroDictionaryManager) NewDefaultDictionary(vm *VM.VirtualMachine, defaultValue mem.MemoryValue) mem.MemoryAddress {
	newDefaultDictAddr := vm.Memory.AllocateEmptySegment()
	newData := make(map[fp.Element]mem.MemoryValue)
	freeOffset := uint64(0)
	dm.Dictionaries[newDefaultDictAddr.SegmentIndex] = ZeroDictionary{
		Data:         &newData,
		DefaultValue: &defaultValue,
		FreeOffset:   &freeOffset,
	}
	return newDefaultDictAddr
}

// Given a memory address, it looks for the right dictionary using the segment index. If no
// segment is associated with the given segment index, it errors
func (dm *ZeroDictionaryManager) GetDictionary(dictAddr mem.MemoryAddress) (ZeroDictionary, error) {
	dict, ok := dm.Dictionaries[dictAddr.SegmentIndex]
	if !ok {
		return ZeroDictionary{}, fmt.Errorf("no dictionary at address: %s", dictAddr)
	}
	if *dict.FreeOffset != dictAddr.Offset {
		return ZeroDictionary{}, fmt.Errorf("no dictionary at address: %s", dictAddr)
	}
	return dict, nil
}

// Given a memory address and a key it returns the value held at that position. The address is used
// to locate the correct dictionary and the key to index on it
func (dm *ZeroDictionaryManager) At(dictAddr mem.MemoryAddress, key fp.Element) (mem.MemoryValue, error) {
	dict, err := dm.GetDictionary(dictAddr)
	if err != nil {
		return mem.UnknownValue, err
	}
	value, err := dict.at(key)
	if err != nil {
		return mem.UnknownValue, err
	}
	return value, nil
}

// Given a memory address,a key and a value it stores the value at the correct position.
func (dm *ZeroDictionaryManager) Set(dictAddr mem.MemoryAddress, key fp.Element, value mem.MemoryValue) error {
	dict, err := dm.GetDictionary(dictAddr)
	if err != nil {
		return err
	}
	dict.set(key, value)
	return nil
}

// Given a memory address and a incrementBy, it increments the freeOffset field of dictionary by it.
func (dm *ZeroDictionaryManager) IncrementFreeOffset(dictAddr mem.MemoryAddress, incrementBy uint64) error {
	dict, err := dm.GetDictionary(dictAddr)
	if err != nil {
		return err
	}
	dict.incrementFreeOffset(incrementBy)
	return nil
}

// Given a memory address and a freeOffset, it sets the freeOffset field of dictionary to it.
func (dm *ZeroDictionaryManager) SetFreeOffset(dictAddr mem.MemoryAddress, freeOffset uint64) error {
	dict, err := dm.GetDictionary(dictAddr)
	if err != nil {
		return err
	}
	dict.setFreeOffset(freeOffset)
	return nil
}
