package vm

import (
	"fmt"

	f "github.com/NethermindEth/juno/core/felt"
)

// The default segment that represents ordinary felts.
const feltSegment = 0

// A virtual memory address of the form SegmentIndex:Offset.
type Relocatable struct {
	SegmentIndex uint64
	Offset       f.Felt
}

// Creates a relocatable out of its components.
func NewRelocatable(segment uint64, offset *f.Felt) *Relocatable {
	return &Relocatable{SegmentIndex: segment, Offset: *offset}
}

// Adds two relocatables if the second one is a Felt (relocatable of the form 0:off).
func (r *Relocatable) Add(r1 *Relocatable, r2 *Relocatable) (*Relocatable, error) {
	if r2.SegmentIndex != feltSegment {
		return nil, fmt.Errorf("cannot add two relocatables")
	}

	r.SegmentIndex = r1.SegmentIndex
	r.Offset.Add(&r1.Offset, &r2.Offset)

	return r, nil
}

// Subs two relocatables if they're in the same segment or the rhs is a Felt.
func (r *Relocatable) Sub(r1 *Relocatable, r2 *Relocatable) (*Relocatable, error) {
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
func (r Relocatable) String() string {
	return fmt.Sprintf("%d:%d", r.SegmentIndex, r.Offset)
}

// Given a map of segment relocations relocates the Relocatable.
func (r *Relocatable) Relocate(r1 *Relocatable, segmentsOffsets *map[uint64]*Relocatable) (*Relocatable, error) {
	if (*segmentsOffsets)[r1.SegmentIndex] == nil {
		return nil, fmt.Errorf("missing segment %d relocation rule", r.SegmentIndex)
	}

	r, err := r.Add((*segmentsOffsets)[r1.SegmentIndex], &Relocatable{0, r1.Offset})

	return r, err
}

// Turns a relocatable of the form 0:offset into offset.Uint64()
// otherwise fails.
func (r *Relocatable) Uint64() (uint64, error) {
	if r.SegmentIndex == feltSegment {
		return r.Offset.Uint64(), nil
	}
	return 0, fmt.Errorf("cannot convert a relocatable '%s' into uint64", *r)
}

// Sets a relocatable into the relocatable of the form 0:r.Offset.SetUint64(v).
func (r *Relocatable) SetUint64(v uint64) *Relocatable {
	r.SegmentIndex = feltSegment
	r.Offset.SetUint64(v)

	return r
}

// Sets a relocatable into the relocatable of the form 0:v.
func (r *Relocatable) SetFelt(v *f.Felt) *Relocatable {
	r.SegmentIndex = feltSegment
	r.Offset.Set(v)

	return r
}
