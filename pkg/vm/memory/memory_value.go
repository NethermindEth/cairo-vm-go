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
	Offset       uint64
}

// Creates a new memory address
func CreateMemoryAddress(segment uint64, offset uint64) *MemoryAddress {
	return &MemoryAddress{SegmentIndex: segment, Offset: offset}
}

// Adds a memory address and a field element
func (address *MemoryAddress) Add(lhs *MemoryAddress, rhs *f.Element) (*MemoryAddress, error) {
	if !rhs.IsUint64() {
		return nil, fmt.Errorf(
			"adding to %s a field element %s greater than uint64",
			lhs.String(),
			rhs.String(),
		)
	}

	address.SegmentIndex = lhs.SegmentIndex
	address.Offset = lhs.Offset + rhs.Uint64()
	return address, nil
}

// Subs from a memory address a felt or another memory address in the same segment
func (address *MemoryAddress) Sub(lhs *MemoryAddress, rhs any) (*MemoryAddress, error) {
	// First match segment index
	address.SegmentIndex = lhs.SegmentIndex

	// Then update offset accordingly
	switch t := rhs.(type) {
	case *f.Element:
		feltRhs := rhs.(*f.Element)
		if !feltRhs.IsUint64() {
			return nil, fmt.Errorf(
				"substracting from %s a field element %s greater than uint64",
				lhs.String(),
				feltRhs.String(),
			)
		}
		address.Offset = lhs.Offset - feltRhs.Uint64()
		return address, nil
	case *MemoryAddress:
		addressRhs := rhs.(*MemoryAddress)
		if lhs.SegmentIndex != addressRhs.SegmentIndex {
			return nil, fmt.Errorf(
				"cannot substract %s from %s due to different segment location",
				addressRhs.String(),
				lhs.String(),
			)
		}
		address.Offset = lhs.Offset - addressRhs.Offset
		return address, nil
	default:
		return nil,
			fmt.Errorf(
				"cannot substract from %s, invalid rhs type: %v. Expected a felt or another memory address",
				address.String(),
				t,
			)

	}
}

func (address MemoryAddress) String() string {
	return fmt.Sprintf(
		"Memory Address: segment: %d, offset: %d", address.SegmentIndex, address.Offset,
	)
}

// Stores all posible types that can be stored in a Memory cell,
//
//   - either a Felt value (an `f.Element`),
//   - or a pointer to another Memory Cell (a `MemoryAddress`)
//
// Both members cannot be non-nil at the same time
type MemoryValue struct {
	felt    *f.Element
	address *MemoryAddress
}

func MemoryValueFromMemoryAddress(address *MemoryAddress) *MemoryValue {
	return &MemoryValue{
		address: address,
	}
}

func MemoryValueFromFieldElement(felt *f.Element) *MemoryValue {
	return &MemoryValue{
		felt: felt,
	}
}

func MemoryValueFromUint64(v uint64) *MemoryValue {
	newElement := f.NewElement(v)
	return &MemoryValue{
		felt: &newElement,
	}
}

func MemoryValueFromAny(anyType any) (*MemoryValue, error) {
	switch t := anyType.(type) {
	case *f.Element:
		return MemoryValueFromFieldElement(anyType.(*f.Element)), nil
	case *MemoryAddress:
		return MemoryValueFromMemoryAddress(anyType.(*MemoryAddress)), nil
	default:
		return nil, fmt.Errorf("invalid type to convert a memory value: %v", t)
	}
}

func EmptyMemoryValueAsFelt() *MemoryValue {
	return &MemoryValue{
		felt: new(f.Element),
	}
}
func EmptyMemoryValueAsAddress() *MemoryValue {
	return &MemoryValue{
		address: new(MemoryAddress),
	}
}

func (mv *MemoryValue) ToMemoryAddress() (*MemoryAddress, error) {
	if mv.address == nil {
		return nil, fmt.Errorf("error trying to read a memory value as an address")
	}
	return mv.address, nil
}

func (mv *MemoryValue) ToFieldElement() (*f.Element, error) {
	if mv.felt == nil {
		return nil, fmt.Errorf("error trying to read a memory value as a field element")
	}
	return mv.felt, nil
}

func (mv *MemoryValue) ToAny() any {
	if mv.felt != nil {
		return mv.felt
	}
	return mv.address
}

func (mv *MemoryValue) IsAddress() bool {
	return mv.address != nil
}

func (mv *MemoryValue) IsFelt() bool {
	return mv.felt != nil
}

// Adds two memory values is the second one is a Felt
func (memVal *MemoryValue) Add(lhs, rhs *MemoryValue) (*MemoryValue, error) {
	if !lhs.IsAddress() {
		return nil, fmt.Errorf("memory value addition requires an address in the lhs")
	}
	if !rhs.IsFelt() {
		return nil, fmt.Errorf("memory value addition requires a felt in the rhs")
	}

	_, err := memVal.address.Add(lhs.address, rhs.felt)
	if err != nil {
		return nil, err
	}

	return memVal, nil
}

// Subs two memory values if they're in the same segment or the rhs is a Felt.
func (memVal *MemoryValue) Sub(lhs, rhs *MemoryValue) (*MemoryValue, error) {
	if !lhs.IsAddress() {
		return nil, fmt.Errorf("memory value substraction requires an address in the lhs")
	}

	_, err := memVal.address.Sub(lhs.address, rhs.ToAny())
	if err != nil {
		return nil, err
	}

	return memVal, nil
}

func (memVal MemoryValue) String() string {
	if memVal.IsAddress() {
		return memVal.address.String()
	}
	return memVal.felt.String()
}

// Retuns a MemoryValue holding a felt as uint if it fits
func (memVal *MemoryValue) Uint64() (uint64, error) {
	if memVal.IsAddress() {
		return 0, fmt.Errorf("cannot convert a memory address '%s' into uint64", *memVal)
	}
	if !memVal.felt.IsUint64() {
		return 0, fmt.Errorf("cannot convert a field element '%s' into uint64", *memVal)
	}

	return memVal.felt.Uint64(), nil
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
