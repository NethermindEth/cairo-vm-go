package memory

import (
	"fmt"

	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

// Represents a write-once Memory Cell
type Cell struct {
	Value    *MemoryValue
	Accessed bool
}

type Segment struct {
	Data []Cell
}

func EmptySegment() *Segment {
	return &Segment{
		Data: make([]Cell, 0),
	}
}

func EmptySegmentWithCapacity(capacity int) Segment {
	return Segment{
		Data: make([]Cell, 0, capacity),
	}
}

func EmptySegmentWithLength(length int) Segment {
	return Segment{
		Data: make([]Cell, length),
	}
}
func (segment *Segment) Write(index uint64, value *MemoryValue) error {
	cell := segment.Data[index]
	if cell.Accessed {
		return fmt.Errorf("rewriting cell at %d, old value: %d new value: %d", index, &cell.Value, &value)
	}
	cell.Accessed = true
	cell.Value = value
	return nil
}

func (segment *Segment) Read(index uint64) *MemoryValue {
	cell := segment.Data[index]
	cell.Accessed = true
	return cell.Value
}

// todo(rodro): Check out temprary segments
// Represents the whole VM memory divided into segments
type Memory struct {
	Segments []*Segment
}

// todo(rodro): can the amount of segments be known before hand?
func InitializeEmptyMemory() *Memory {
	return &Memory{
		// capacity 4 should be enough for the minimum amount of segments
		Segments: make([]*Segment, 0, 4),
	}
}

func (memory *Memory) AllocateSegment(data *[]f.Element) (int, error) {
	newSegment := EmptySegmentWithLength(len(*data))
	for i := range *data {
		memVal := MemoryValueFromFieldElement(&(*data)[i])
		err := newSegment.Write(uint64(i), memVal)
		if err != nil {
			return 0, fmt.Errorf("cannot allocate new segment: %w", err)
		}
	}
	memory.Segments = append(memory.Segments, &newSegment)
	return len(memory.Segments), nil
}

// Allocates a new segment and returns its index
func (memory *Memory) AllocateEmptySegment() int {
	memory.Segments = append(memory.Segments, EmptySegment())
	return len(memory.Segments) - 1
}

func (memory *Memory) Write(address *MemoryAddress, value *MemoryValue) error {
	if address.SegmentIndex > uint64(len(memory.Segments)) {
		return fmt.Errorf("writing to unallocated segment %d", address.SegmentIndex)
	}

	return memory.Segments[address.SegmentIndex].Write(address.Offset, value)
}

func (memory *Memory) Read(address *MemoryAddress) (*MemoryValue, error) {
	if address.SegmentIndex > uint64(len(memory.Segments)) {
		return nil, fmt.Errorf("reading from unallocated segment %d", address.SegmentIndex)
	}

	return memory.Segments[address.SegmentIndex].Read(address.Offset), nil
}
