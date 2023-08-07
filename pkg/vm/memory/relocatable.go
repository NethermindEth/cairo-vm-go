package memory

import (
	"fmt"

	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

// Represents a Memory Address of the Cairo VM. Because memory is split between different segments
// during execution, addresses has two locators: the segment they belong to and the location
// inside that segment
type MemoryAddress struct {
	SegmentIndex uint64
	Offset       *f.Element
}

// Creates a new memory address
func CreateMemoryAddress(segment uint64, offset *f.Element) *MemoryAddress {
	return &MemoryAddress{SegmentIndex: segment, Offset: offset}
}

func (address MemoryAddress) String() string {
	return fmt.Sprintf("Memory Address: segment: %d, offset: %s", address.SegmentIndex, address.Offset.Text(10))
}

// Wraps all posible types that can be stored in a Memory cell,
//
//   - either a Felt value (an `f.Element`),
//   - or a pointer to another Memory Cell (a `MemoryAddress`)
type MemoryValue struct {
	// When isAddress is true, this indicates the segment of
	// the memory address
	segmentIndex uint64
	// This represents the offset if it is an address or the
	// value if it is a field element
	value     *f.Element
	isAddress bool
}

func EmptyMemoryValue() *MemoryValue {
	return &MemoryValue{
		value: new(f.Element),
	}
}

func (mv *MemoryValue) ToMemoryAddress() (*MemoryAddress, error) {
	if !mv.isAddress {
		return nil, fmt.Errorf("error trying to read a memory address as field element")
	}
	return &MemoryAddress{
		SegmentIndex: mv.segmentIndex,
		Offset:       mv.value,
	}, nil
}

func (mv *MemoryValue) ToFieldElement() (*f.Element, error) {
	if mv.isAddress {
		return nil, fmt.Errorf("error trying to read a field element as a memory address")
	}
	return mv.value, nil
}

func MemoryValueFromMemoryAddress(address *MemoryAddress) *MemoryValue {
	return &MemoryValue{
		segmentIndex: address.SegmentIndex,
		value:        address.Offset,
		isAddress:    true,
	}
}

func MemoryValueFromFieldElement(felt *f.Element) *MemoryValue {
	return &MemoryValue{
		value:     felt,
		isAddress: false,
	}
}

func MemoryValueFromUint64(v uint64) *MemoryValue {
	newElement := f.NewElement(v)
	return &MemoryValue{
		value: &newElement,
	}
}

// Adds two memory value is the second one is a Felt
func (memVal *MemoryValue) Add(lhs *MemoryValue, rhs *MemoryValue) (*MemoryValue, error) {
	if rhs.isAddress {
		return nil, fmt.Errorf("cannot add two memory addresses")
	}

	fmt.Println(memVal)
	fmt.Println("x0")
	fmt.Println(lhs)
	fmt.Println("x1")
	fmt.Println(rhs)

	// todo(rodro): is lhs always a memory address?
	if lhs.isAddress {
		memVal.isAddress = true
		memVal.segmentIndex = lhs.segmentIndex
	}
	fmt.Println("x2", lhs.value, rhs.value)
	memVal.value.Add(lhs.value, rhs.value)
	fmt.Println("x3")

	return memVal, nil
}

// Subs two relocatables if they're in the same segment or the rhs is a Felt.
func (memVal *MemoryValue) Sub(lhs *MemoryValue, rhs *MemoryValue) (*MemoryValue, error) {
	if !rhs.isAddress {
		// todo(rodro): is lhs always a memory address?
		if lhs.isAddress {
			memVal.isAddress = true
			memVal.segmentIndex = lhs.segmentIndex
		}
		memVal.value.Sub(lhs.value, rhs.value)
		return memVal, nil
	}

	// todo(rodro): can lhs not be a memory address?
	if !lhs.isAddress {
		return nil, fmt.Errorf("sub not implemented for lhs as non address")
	}

	if lhs.segmentIndex != rhs.segmentIndex {
		return nil, fmt.Errorf("cannot subtract relocatables from different segments: %d != %d", lhs.segmentIndex, rhs.segmentIndex)
	}

	memVal.isAddress = true
	memVal.segmentIndex = lhs.segmentIndex
	memVal.value.Sub(lhs.value, rhs.value)

	return memVal, nil
}

func (memVal MemoryValue) String() string {
	if memVal.isAddress {
		return MemoryAddress{
			memVal.segmentIndex,
			memVal.value,
		}.String()
	}
	return memVal.value.String()
}

// Note: Commenting this function since relocation is possibly going to look
// different.
// Given a map of segment relocation, update a memory address location
//func (r *MemoryAddress) Relocate(r1 *MemoryAddress, segmentsOffsets *map[uint64]*MemoryAddress) (*MemoryAddress, error) {
//	if (*segmentsOffsets)[r1.SegmentIndex] == nil {
//		return nil, fmt.Errorf("missing segment %d relocation rule", r.SegmentIndex)
//	}
//
//	r, err := r.Add((*segmentsOffsets)[r1.SegmentIndex], &MemoryAddress{0, r1.Offset})
//
//	return r, err
//}

// Turns a relocatable of the form 0:offset into offset.Uint64()
// otherwise fails.
func (memVal *MemoryValue) Uint64() (uint64, error) {
	if memVal.isAddress {
		return 0, fmt.Errorf("cannot convert a memory address '%s' into uint64", *memVal)
	}
	return memVal.value.Uint64(), nil
}
