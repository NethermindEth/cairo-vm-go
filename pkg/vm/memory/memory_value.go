package memory

import (
	"errors"
	"fmt"

	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"golang.org/x/exp/constraints"
)

// Represents a Memory Address of the Cairo VM. Because memory is split between different segments
// during execution, addresses has two locators: the segment they belong to and the location
// inside that segment
type MemoryAddress struct {
	SegmentIndex uint64
	Offset       uint64
}

// Creates a new memory address
func NewMemoryAddress(segment uint64, offset uint64) *MemoryAddress {
	return &MemoryAddress{SegmentIndex: segment, Offset: offset}
}

func (address *MemoryAddress) Equal(other *MemoryAddress) bool {
	return address.SegmentIndex == other.SegmentIndex && address.Offset == other.Offset
}

// Adds a memory address and a field element

func (address *MemoryAddress) Add(lhs *MemoryAddress, rhs *f.Element) (*MemoryAddress, error) {
	lhsOffset := new(f.Element).SetUint64(lhs.Offset)
	newOffset := new(f.Element).Add(lhsOffset, rhs)
	if !newOffset.IsUint64() {
		return nil, fmt.Errorf("new offset bigger than uint64: %s", rhs.Text(10))
	}
	address.SegmentIndex = lhs.SegmentIndex
	address.Offset = newOffset.Uint64()
	return address, nil
}

// Subs from a memory address a felt or another memory address in the same segment
func (address *MemoryAddress) Sub(lhs *MemoryAddress, rhs any) (*MemoryAddress, error) {
	// First match segment index
	address.SegmentIndex = lhs.SegmentIndex

	// Then update offset accordingly
	switch rhs := rhs.(type) {
	case uint64:
		if rhs > lhs.Offset {
			return nil, errors.New("rhs is greater than lhs offset")
		}
		address.Offset = lhs.Offset - rhs
		return address, nil
	case *f.Element:
		if !rhs.IsUint64() {
			return nil, fmt.Errorf("rhs field element does not fit in uint64: %s", rhs)
		}
		feltRhs64 := rhs.Uint64()
		if feltRhs64 > lhs.Offset {
			return nil, fmt.Errorf("rhs %d is greater than lhs offset %d", feltRhs64, lhs.Offset)
		}
		address.Offset = lhs.Offset - feltRhs64
		return address, nil
	case *MemoryAddress:
		if lhs.SegmentIndex != rhs.SegmentIndex {
			return nil, fmt.Errorf("addresses are in different segments: rhs is in %d, lhs is in %d",
				rhs.SegmentIndex, lhs.SegmentIndex)
		}
		if rhs.Offset > lhs.Offset {
			return nil, fmt.Errorf("rhs offset %d is greater than lhs offset %d", rhs.Offset, lhs.Offset)
		}
		address.Offset = lhs.Offset - rhs.Offset
		return address, nil
	default:
		return nil, fmt.Errorf("unknown rhs type: %T", rhs)
	}
}

func (address *MemoryAddress) Relocate(segmentsOffset []uint64) *f.Element {
	// no risk overflow because this sizes exists in actual Memory
	// so if by chance the uint64 addition overflowed, then we have
	// a machine with more than 2**64 bytes of memory (quite a lot!)
	return new(f.Element).SetUint64(
		segmentsOffset[address.SegmentIndex] + address.Offset,
	)
}

func (address MemoryAddress) String() string {
	return fmt.Sprintf(
		"%d:%d", address.SegmentIndex, address.Offset,
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

func MemoryValueFromInt[T constraints.Integer](v T) *MemoryValue {
	if v >= 0 {
		return MemoryValueFromUint(uint64(v))
	}
	lhs := &f.Element{}
	rhs := new(f.Element).SetUint64(uint64(-v))
	return &MemoryValue{
		felt: new(f.Element).Sub(lhs, rhs),
	}
}

func MemoryValueFromUint[T constraints.Unsigned](v T) *MemoryValue {
	newElement := f.NewElement(uint64(v))
	return &MemoryValue{
		felt: &newElement,
	}
}

func MemoryValueFromSegmentAndOffset[T constraints.Integer](segmentIndex, offset T) *MemoryValue {
	return &MemoryValue{
		address: &MemoryAddress{SegmentIndex: uint64(segmentIndex), Offset: uint64(offset)},
	}
}

func MemoryValueFromAny(anyType any) (*MemoryValue, error) {
	switch anyType := anyType.(type) {
	case int:
		return MemoryValueFromInt(anyType), nil
	case uint64:
		return MemoryValueFromUint(anyType), nil
	case *f.Element:
		return MemoryValueFromFieldElement(anyType), nil
	case *MemoryAddress:
		return MemoryValueFromMemoryAddress(anyType), nil
	default:
		return nil, fmt.Errorf("invalid type to convert to a MemoryValue: %T", anyType)
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

func EmptyMemoryValueAs(address bool) *MemoryValue {
	if address {
		return EmptyMemoryValueAsAddress()
	}
	return EmptyMemoryValueAsFelt()
}

func (mv *MemoryValue) ToMemoryAddress() (*MemoryAddress, error) {
	if mv.address == nil {
		return nil, errors.New("memory value is not an address")
	}
	return mv.address, nil
}

func (mv *MemoryValue) ToFieldElement() (*f.Element, error) {
	if mv.felt == nil {
		return nil, fmt.Errorf("memory value is not a field element")
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

func (mv *MemoryValue) Equal(other *MemoryValue) bool {
	if mv.IsAddress() && other.IsAddress() {
		return mv.address.Equal(other.address)
	}
	if mv.IsFelt() && other.IsFelt() {
		return mv.felt.Equal(other.felt)
	}
	return false
}

func (mv *MemoryValue) Add(lhs, rhs *MemoryValue) (*MemoryValue, error) {
	var err error

	// If both lhs and rhs are felts, perform a simple addition
	if lhs.IsFelt() {
		if rhs.IsFelt() { // Felt + Felt
			mv.felt = MemoryValueFromUint(lhs.felt.Uint64() + rhs.felt.Uint64()).felt
		} else { // Felt + Address
			mv.address, err = mv.address.Add(rhs.address, lhs.felt)
		}
	} else if rhs.IsFelt() { // Address + Felt
		mv.address, err = mv.address.Add(lhs.address, rhs.felt)
	} else { // Address + Address
		return nil, errors.New("addition of two addresses is not supported")
	}

	if err != nil {
		return nil, err
	}

	return mv, nil
}

// Subs two memory values if they're in the same segment or the rhs is a Felt.
func (mv *MemoryValue) Sub(lhs, rhs *MemoryValue) (*MemoryValue, error) {
	var err error
	if lhs.IsAddress() {
		mv.address, err = mv.address.Sub(lhs.address, rhs.ToAny())
	} else {
		if rhs.IsAddress() {
			return nil, errors.New("cannot substract an address from a felt")
		} else {
			mv.felt = mv.felt.Sub(lhs.felt, rhs.felt)
		}
	}

	if err != nil {
		return nil, err
	}

	return mv, nil
}

func (mv *MemoryValue) Mul(lhs, rhs *MemoryValue) (*MemoryValue, error) {
	if lhs.IsAddress() || rhs.IsAddress() {
		return nil, errors.New("cannot multiply memory addresses")
	}
	mv.felt.Mul(lhs.felt, rhs.felt)
	return mv, nil
}

func (mv *MemoryValue) Div(lhs, rhs *MemoryValue) (*MemoryValue, error) {
	if lhs.IsAddress() || rhs.IsAddress() {
		return nil, errors.New("cannot divide memory addresses")
	}

	mv.felt.Div(lhs.felt, rhs.felt)
	return mv, nil
}

func (mv MemoryValue) String() string {
	if mv.IsAddress() {
		return mv.address.String()
	}
	return mv.felt.String()
}

// Retuns a MemoryValue holding a felt as uint if it fits
func (mv *MemoryValue) Uint64() (uint64, error) {
	if mv.IsAddress() {
		return 0, fmt.Errorf("cannot convert a memory address into uint64: %s", *mv)
	}
	if !mv.felt.IsUint64() {
		return 0, fmt.Errorf("field element does not fit in uint64: %s", mv.String())
	}

	return mv.felt.Uint64(), nil
}
