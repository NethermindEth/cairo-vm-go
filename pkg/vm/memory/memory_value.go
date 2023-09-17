package memory

import (
	"errors"
	"fmt"

	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	mo "github.com/samber/mo"
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
	if !rhs.IsUint64() {
		return nil, fmt.Errorf("field element does not fit in uint64: %s", rhs.String())
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
	return new(f.Element).SetUint64(segmentsOffset[address.SegmentIndex] + address.Offset)
}

func (address MemoryAddress) String() string {
	return fmt.Sprintf(
		"Memory Address: segment: %d, offset: %d", address.SegmentIndex, address.Offset,
	)
}

// This is an abreviation for simplicity
type memoryValue = mo.Either[*f.Element, *MemoryAddress]

// Stores all posible types that can be stored in a Memory cell,
//
//   - either a Felt value (an `f.Element`),
//   - or a pointer to another Memory Cell (a `MemoryAddress`)
type MemoryValue memoryValue

func MemoryValueFromMemoryAddress(address *MemoryAddress) *MemoryValue {
	var mv MemoryValue = MemoryValue(mo.Right[*f.Element, *MemoryAddress](address))
	return &mv
}

func MemoryValueFromFieldElement(felt *f.Element) *MemoryValue {
	mv := MemoryValue(mo.Left[*f.Element, *MemoryAddress](felt))
	return &mv
}

func MemoryValueFromInt[T constraints.Integer](v T) *MemoryValue {
	newElement := f.NewElement(uint64(v))
	return MemoryValueFromFieldElement(&newElement)
}

func MemoryValueFromSegmentAndOffset[T constraints.Integer](segmentIndex, offset T) *MemoryValue {
	return MemoryValueFromMemoryAddress(&MemoryAddress{SegmentIndex: uint64(segmentIndex), Offset: uint64(offset)})
}

func MemoryValueFromAny(anyType any) (*MemoryValue, error) {
	switch anyType := anyType.(type) {
	case uint64:
		return MemoryValueFromInt(anyType), nil
	case *f.Element:
		return MemoryValueFromFieldElement(anyType), nil
	case *MemoryAddress:
		return MemoryValueFromMemoryAddress(anyType), nil
	default:
		return nil, fmt.Errorf("invalid type to convert to a MemoryValue: %T", anyType)
	}
}

func EmptyMemoryValueAsFelt() *MemoryValue {
	return MemoryValueFromFieldElement(new(f.Element))
}

func EmptyMemoryValueAsAddress() *MemoryValue {
	return MemoryValueFromMemoryAddress(new(MemoryAddress))
}

func EmptyMemoryValueAs(address bool) *MemoryValue {
	if address {
		return EmptyMemoryValueAsAddress()
	}
	return EmptyMemoryValueAsFelt()
}

func (mv *MemoryValue) toMemoryAddress() *MemoryAddress {
	return memoryValue(*mv).RightOrEmpty()
}

func (mv *MemoryValue) ToMemoryAddress() (*MemoryAddress, error) {
	address, isAddress := memoryValue(*mv).Right()
	if !isAddress {
		return nil, errors.New("memory value is not an address")
	}
	return address, nil
}

func (mv *MemoryValue) toFieldElement() *f.Element {
	return memoryValue(*mv).LeftOrEmpty()
}

func (mv *MemoryValue) ToFieldElement() (*f.Element, error) {
	felt, isFelt := memoryValue(*mv).Left()
	if !isFelt {
		return nil, fmt.Errorf("memory value is not a field element")
	}
	return felt, nil
}

func (mv *MemoryValue) ToAny() any {
	felt, isFelt := memoryValue(*mv).Left()

	if isFelt {
		return felt
	}

	return memoryValue(*mv).RightOrEmpty()
}

func (mv *MemoryValue) IsAddress() bool {
	return memoryValue(*mv).IsRight()
}

func (mv *MemoryValue) IsFelt() bool {
	return memoryValue(*mv).IsLeft()
}

func (mv *MemoryValue) Equal(other *MemoryValue) (isEqual bool) {
	if mv.IsAddress() && other.IsAddress() {
		return mv.toMemoryAddress().Equal(other.toMemoryAddress())
	} else if mv.IsFelt() && other.IsFelt() {
		return mv.toFieldElement().Equal(other.toFieldElement())
	}
	return false
}

// Adds two memory values is the second one is a Felt
func (mv *MemoryValue) Add(lhs, rhs *MemoryValue) (res *MemoryValue, err error) {
	memoryValue(*lhs).MapLeft(func(e *f.Element) mo.Either[*f.Element, *MemoryAddress] {
		if rhs.IsAddress() {
			_, err = mv.toMemoryAddress().Add(rhs.toMemoryAddress(), e)
		} else {
			mv.toFieldElement().Add(e, rhs.toFieldElement())
		}
		return mo.Left[*f.Element, *MemoryAddress](nil)
	}).MapRight(func(ma *MemoryAddress) mo.Either[*f.Element, *MemoryAddress] {
		if !rhs.IsFelt() {
			err = errors.New("rhs is not a felt")
		} else {
			_, err = mv.toMemoryAddress().Add(ma, rhs.toFieldElement())
		}

		return memoryValue(*mv)
	})

	res = mv
	if err != nil {
		res = nil
	}

	return
}

// Subs two memory values if they're in the same segment or the rhs is a Felt.
func (mv *MemoryValue) Sub(lhs, rhs *MemoryValue) (res *MemoryValue, err error) {
	memoryValue(*lhs).MapLeft(func(e *f.Element) mo.Either[*f.Element, *MemoryAddress] {
		if rhs.IsAddress() {
			err = errors.New("cannot substract an address from a felt")
		} else {
			mv.toFieldElement().Sub(e, rhs.toFieldElement())
		}
		return mo.Left[*f.Element, *MemoryAddress](nil)
	}).MapRight(func(ma *MemoryAddress) mo.Either[*f.Element, *MemoryAddress] {
		_, err = mv.toMemoryAddress().Sub(ma, rhs.ToAny())
		return memoryValue(*mv)
	})

	res = mv

	if err != nil {
		res = nil
	}

	return
}

func (mv *MemoryValue) Mul(lhs, rhs *MemoryValue) (*MemoryValue, error) {
	if lhs.IsAddress() || rhs.IsAddress() {
		return nil, errors.New("cannot multiply memory addresses")
	}
	mv.toFieldElement().Mul(lhs.toFieldElement(), rhs.toFieldElement())
	return mv, nil
}

func (mv *MemoryValue) Div(lhs, rhs *MemoryValue) (*MemoryValue, error) {
	if lhs.IsAddress() || rhs.IsAddress() {
		return nil, errors.New("cannot divide memory addresses")
	}

	mv.toFieldElement().Div(lhs.toFieldElement(), rhs.toFieldElement())
	return mv, nil
}

func (mv MemoryValue) String() string {
	if mv.IsAddress() {
		return mv.toMemoryAddress().String()
	}
	return mv.toFieldElement().String()
}

// Retuns a MemoryValue holding a felt as uint if it fits
func (mv *MemoryValue) Uint64() (uint64, error) {
	if mv.IsAddress() {
		return 0, fmt.Errorf("cannot convert a memory address into uint64: %s", *mv)
	}

	felt := mv.toFieldElement()

	if !felt.IsUint64() {
		return 0, fmt.Errorf("field element does not fit in uint64: %s", mv.String())
	}

	return felt.Uint64(), nil
}
