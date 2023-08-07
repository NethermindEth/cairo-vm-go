package memory

import (
	"fmt"

	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

// The default segment that represents ordinary felts.
const feltSegment = 0

// Represents a Memory Address of the Cairo VM. Because memory is split between different segments
// during execution, addresses has two locators: the segment they belong to and the location
// inside that segment
type MemoryAddress struct {
	SegmentIndex uint64
	Offset       f.Element
}

// Creates a new memory address
func CreateMemoryAddress(segment uint64, offset *f.Element) *MemoryAddress {
	return &MemoryAddress{SegmentIndex: segment, Offset: *offset}
}

// Adds two relocatables if the second one is a Felt (relocatable of the form 0:off).
func (r *MemoryAddress) Add(r1 *MemoryAddress, r2 *MemoryAddress) (*MemoryAddress, error) {
	if r2.SegmentIndex != feltSegment {
		return nil, fmt.Errorf("cannot add two relocatables")
	}

	r.SegmentIndex = r1.SegmentIndex
	r.Offset.Add(&r1.Offset, &r2.Offset)

	return r, nil
}

// Subs two relocatables if they're in the same segment or the rhs is a Felt.
func (r *MemoryAddress) Sub(r1 *MemoryAddress, r2 *MemoryAddress) (*MemoryAddress, error) {
	if r2.SegmentIndex == feltSegment {
		r.SegmentIndex = r1.SegmentIndex
		r.Offset.Sub(&r1.Offset, &r2.Offset)
		return r, nil
	}

	if r1.SegmentIndex != r2.SegmentIndex {
		return nil, fmt.Errorf("cannot subtract relocatables from different segments: %d != %d", r1.SegmentIndex, r2.SegmentIndex)
	}

	r.SegmentIndex = feltSegment
	r.Offset.Sub(&r1.Offset, &r2.Offset)

	return r, nil
}

// String representation of a relocatable.
func (r MemoryAddress) String() string {
	return fmt.Sprintf("%d:%d", r.SegmentIndex, r.Offset)
}

// Given a map of segment relocation, update a memory address location
func (r *MemoryAddress) Relocate(r1 *MemoryAddress, segmentsOffsets *map[uint64]*MemoryAddress) (*MemoryAddress, error) {
	if (*segmentsOffsets)[r1.SegmentIndex] == nil {
		return nil, fmt.Errorf("missing segment %d relocation rule", r.SegmentIndex)
	}

	r, err := r.Add((*segmentsOffsets)[r1.SegmentIndex], &MemoryAddress{0, r1.Offset})

	return r, err
}

// Turns a relocatable of the form 0:offset into offset.Uint64()
// otherwise fails.
func (r *MemoryAddress) Uint64() (uint64, error) {
	if r.SegmentIndex == feltSegment {
		return r.Offset.Uint64(), nil
	}
	return 0, fmt.Errorf("cannot convert a relocatable '%s' into uint64", *r)
}

// Sets a relocatable into the relocatable of the form 0:r.Offset.SetUint64(v).
func (r *MemoryAddress) SetUint64(v uint64) *MemoryAddress {
	r.SegmentIndex = feltSegment
	r.Offset.SetUint64(v)

	return r
}

// Sets a relocatable into the relocatable of the form 0:v.
func (r *MemoryAddress) SetFelt(v *f.Element) *MemoryAddress {
	r.SegmentIndex = feltSegment
	r.Offset.Set(v)

	return r
}
