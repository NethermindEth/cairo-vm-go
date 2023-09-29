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

var UnknownValue = MemoryAddress{}

func (address *MemoryAddress) Equal(other *MemoryAddress) bool {
	return address.SegmentIndex == other.SegmentIndex && address.Offset == other.Offset
}

// Adds a memory address and a field element
func (address *MemoryAddress) Add(lhs *MemoryAddress, rhs *f.Element) error {
	lhsOffset := new(f.Element).SetUint64(lhs.Offset)
	newOffset := new(f.Element).Add(lhsOffset, rhs)

	if !newOffset.IsUint64() {
		return fmt.Errorf("new offset bigger than uint64: %s", rhs.Text(10))
	}

	address.SegmentIndex = lhs.SegmentIndex
	address.Offset = newOffset.Uint64()
	return nil
}

// Subs from a memory address a felt or another memory address in the same segment
func (address *MemoryAddress) Sub(lhs *MemoryAddress, rhs any) error {
	// First match segment index
	address.SegmentIndex = lhs.SegmentIndex

	// Then update offset accordingly
	switch rhs := rhs.(type) {
	case uint64:
		if rhs > lhs.Offset {
			return errors.New("rhs is greater than lhs offset")
		}
		address.Offset = lhs.Offset - rhs
		return nil
	case *f.Element:
		if !rhs.IsUint64() {
			return fmt.Errorf("rhs field element does not fit in uint64: %s", rhs)
		}
		feltRhs64 := rhs.Uint64()
		if feltRhs64 > lhs.Offset {
			return fmt.Errorf("rhs %d is greater than lhs offset %d", feltRhs64, lhs.Offset)
		}
		address.Offset = lhs.Offset - feltRhs64
		return nil
	case *MemoryAddress:
		if lhs.SegmentIndex != rhs.SegmentIndex {
			return fmt.Errorf("addresses are in different segments: rhs is in %d, lhs is in %d",
				rhs.SegmentIndex, lhs.SegmentIndex)
		}
		if rhs.Offset > lhs.Offset {
			return fmt.Errorf("rhs offset %d is greater than lhs offset %d", rhs.Offset, lhs.Offset)
		}
		address.Offset = lhs.Offset - rhs.Offset
		return nil
	default:
		return fmt.Errorf("unknown rhs type: %T", rhs)
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
type MemoryValue struct {
	felt      f.Element
	address   MemoryAddress
	isFelt    bool
	isAddress bool
}

func MemoryValueFromMemoryAddress(address *MemoryAddress) MemoryValue {
	return MemoryValue{
		address:   *address,
		isAddress: true,
	}
}

func MemoryValueFromFieldElement(felt *f.Element) MemoryValue {
	return MemoryValue{
		felt:   *felt,
		isFelt: true,
	}
}

func MemoryValueFromInt[T constraints.Integer](v T) MemoryValue {
	if v >= 0 {
		return MemoryValueFromUint(uint64(v))
	}

	value := MemoryValue{isFelt: true}
	rhs := f.NewElement(uint64(-v))
	value.felt.Sub(&value.felt, &rhs)
	return value
}

func MemoryValueFromUint[T constraints.Unsigned](v T) MemoryValue {
	return MemoryValue{
		felt:   f.NewElement(uint64(v)),
		isFelt: true,
	}
}

func MemoryValueFromSegmentAndOffset[T constraints.Integer](segmentIndex, offset T) MemoryValue {
	return MemoryValue{
		address:   MemoryAddress{SegmentIndex: uint64(segmentIndex), Offset: uint64(offset)},
		isAddress: true,
	}
}

func MemoryValueFromAny(anyType any) (MemoryValue, error) {
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
		return MemoryValue{}, fmt.Errorf("invalid type to convert to a MemoryValue: %T", anyType)
	}
}

func EmptyMemoryValueAsFelt() MemoryValue {
	return MemoryValue{
		isFelt: true,
	}
}

func EmptyMemoryValueAsAddress() MemoryValue {
	return MemoryValue{
		isAddress: true,
	}
}

func EmptyMemoryValueAs(address bool) MemoryValue {
	return MemoryValue{
		isAddress: address,
		isFelt:    !address,
	}
}

func (mv *MemoryValue) ToMemoryAddress() (*MemoryAddress, error) {
	if !mv.isAddress {
		return nil, errors.New("memory value is not an address")
	}
	return &mv.address, nil
}

func (mv *MemoryValue) ToFieldElement() (*f.Element, error) {
	if !mv.isFelt {
		return nil, fmt.Errorf("memory value is not a field element")
	}
	return &mv.felt, nil
}

func (mv *MemoryValue) ToAny() any {
	if mv.isAddress {
		return &mv.address
	}
	return &mv.felt
}

func (mv *MemoryValue) IsAddress() bool {
	return mv.isAddress
}

func (mv *MemoryValue) IsFelt() bool {
	return mv.isFelt
}

func (mv *MemoryValue) Known() bool {
	return mv.isAddress || mv.isFelt
}

func (mv *MemoryValue) Equal(other *MemoryValue) bool {
	if mv.IsAddress() && other.IsAddress() {
		return mv.address.Equal(&other.address)
	}
	if mv.IsFelt() && other.IsFelt() {
		return mv.felt.Equal(&other.felt)
	}
	return false
}

// Adds two memory values is the second one is a Felt
func (mv *MemoryValue) Add(lhs, rhs *MemoryValue) error {
	if lhs.IsAddress() {
		if !rhs.IsFelt() {
			return errors.New("rhs is not a felt")
		}
		return mv.address.Add(&lhs.address, &rhs.felt)
	}

	if rhs.IsAddress() {
		return mv.address.Add(&rhs.address, &lhs.felt)
	}
	mv.felt.Add(&lhs.felt, &rhs.felt)
	return nil
}

// Subs two memory values if they're in the same segment or the rhs is a Felt.
func (mv *MemoryValue) Sub(lhs, rhs *MemoryValue) error {
	if lhs.IsAddress() {
		return mv.address.Sub(&lhs.address, rhs.ToAny())
	}

	if rhs.IsAddress() {
		return errors.New("cannot substract an address from a felt")
	}

	mv.felt.Sub(&lhs.felt, &rhs.felt)
	return nil
}

func (mv *MemoryValue) Mul(lhs, rhs *MemoryValue) error {
	if lhs.IsAddress() || rhs.IsAddress() {
		return errors.New("cannot multiply memory addresses")
	}
	mv.felt.Mul(&lhs.felt, &rhs.felt)
	return nil
}

func (mv *MemoryValue) Div(lhs, rhs *MemoryValue) error {
	if lhs.IsAddress() || rhs.IsAddress() {
		return errors.New("cannot divide memory addresses")
	}
	mv.felt.Div(&lhs.felt, &rhs.felt)
	return nil
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
