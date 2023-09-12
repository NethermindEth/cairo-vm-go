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

func (cell *Cell) Write(value *MemoryValue) error {
	if cell.Accessed && cell.Value != nil && !cell.Value.Equal(value) {
		return fmt.Errorf("rewriting cell: old value \"%d\", new value \"%d\"", &cell.Value, &value)
	}

	cell.Accessed = true
	cell.Value = value
	return nil
}

func (cell *Cell) Read() *MemoryValue {
	cell.Accessed = true
	if cell.Value == nil {
		cell.Value = EmptyMemoryValueAsFelt()
	}
	return cell.Value
}

func (cell *Cell) String() string {
	if !cell.Accessed {
		return "-"
	}
	return cell.Value.String()
}

type Segment struct {
	Data []Cell
}

func EmptySegment() *Segment {
	return &Segment{
		Data: make([]Cell, 0),
	}
}

func EmptySegmentWithCapacity(capacity int) *Segment {
	return &Segment{
		Data: make([]Cell, 0, capacity),
	}
}

func EmptySegmentWithLength(length int) *Segment {
	return &Segment{
		Data: make([]Cell, length),
	}
}

func (segment *Segment) Len() uint64 {
	return uint64(len(segment.Data))
}

func (segment *Segment) Cap() uint64 {
	return uint64(cap(segment.Data))
}

// Writes a new memory value to a specified offset, errors in case of overwriting an existing cell
func (segment *Segment) Write(offset uint64, value *MemoryValue) error {
	if offset >= uint64(len(segment.Data)) {
		segment.IncreaseSegmentSize(offset + 1)
	}

	err := segment.Data[offset].Write(value)
	if err != nil {
		return fmt.Errorf("write cell at segment offset %d: %v", offset, err)
	}
	return nil
}

// Reads a memory value from a specified offset at the segment
func (segment *Segment) Read(offset uint64) *MemoryValue {
	if offset >= uint64(len(segment.Data)) {
		segment.IncreaseSegmentSize(offset + 1)
	}

	return segment.Data[offset].Read()
}

func (segment *Segment) Peek(offset uint64) *Cell {
	if offset >= uint64(len(segment.Data)) {
		segment.IncreaseSegmentSize(offset + 1)
	}

	return &segment.Data[offset]
}

// Increase a segment allocated space. Panics if the new size is smaller
func (segment *Segment) IncreaseSegmentSize(newSize uint64) {
	segmentData := segment.Data
	if len(segmentData) > int(newSize) {
		panic("cannot increase segment size to s smaller value")
	}

	var newSegmentData []Cell
	if cap(segmentData) > int(newSize) {
		// If there is enough capacity just reslice
		newSegmentData = segmentData[:newSize]
	} else {
		// If not enough capacity then re allocate and copy
		newSegmentData = make([]Cell, newSize, newSize+100)
		copy(newSegmentData, segmentData)
	}
	segment.Data = newSegmentData
}

func (segment *Segment) String() string {
	header := fmt.Sprintf("len: %d cap: %d\n", len(segment.Data), cap(segment.Data))
	for i := range segment.Data {
		if segment.Data[i].Accessed {
			header += fmt.Sprintf("[%d]-> %s\n", i, segment.Data[i].String())
		}
	}
	return header
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

// Allocates a new segment providing its initial data and returns its index
func (memory *Memory) AllocateSegment(data []*f.Element) (int, error) {
	newSegment := EmptySegmentWithLength(len(data))
	for i := range data {
		memVal := MemoryValueFromFieldElement(data[i])
		err := newSegment.Write(uint64(i), memVal)
		if err != nil {
			return 0, err
		}
	}
	memory.Segments = append(memory.Segments, newSegment)
	return len(memory.Segments), nil
}

// Allocates an empty segment and returns its index
func (memory *Memory) AllocateEmptySegment() int {
	memory.Segments = append(memory.Segments, EmptySegment())
	return len(memory.Segments) - 1
}

// Writes to a memory address a new memory value. Errors if writing to an unallocated
// space or if rewriting a specific cell
func (memory *Memory) Write(segmentIndex uint64, offset uint64, value *MemoryValue) error {
	if segmentIndex > uint64(len(memory.Segments)) {
		return fmt.Errorf("unallocated segment at index %d", segmentIndex)
	}

	return memory.Segments[segmentIndex].Write(offset, value)
}

func (memory *Memory) WriteToAddress(address *MemoryAddress, value *MemoryValue) error {
	return memory.Write(address.SegmentIndex, address.Offset, value)
}

// Reads a memory value given the segment index and offset. Errors if reading from
// an unallocated space. If reading a cell which hasn't been accesed before, it is
// initalized with its default zero value
func (memory *Memory) Read(segmentIndex uint64, offset uint64) (*MemoryValue, error) {
	if segmentIndex > uint64(len(memory.Segments)) {
		return nil, fmt.Errorf("unallocated segment at index %d", segmentIndex)
	}
	return memory.Segments[segmentIndex].Read(offset), nil
}

// Reads a memory value from a memory address. Errors if reading from an unallocated
// space. If reading a cell which hasn't been accesed before, it is initalized with
// its default zero value
func (memory *Memory) ReadFromAddress(address *MemoryAddress) (*MemoryValue, error) {
	return memory.Read(address.SegmentIndex, address.Offset)
}

// Given a segment index and offset returns a pointer to the Memory Cell
func (memory *Memory) Peek(segmentIndex uint64, offset uint64) (*Cell, error) {
	if segmentIndex > uint64(len(memory.Segments)) {
		return nil, fmt.Errorf("unallocated segment at index %d", segmentIndex)
	}
	return memory.Segments[segmentIndex].Peek(offset), nil
}

// Given a Memory Address returns a pointer to the Memory Cell
func (memory *Memory) PeekFromAddress(address *MemoryAddress) (*Cell, error) {
	return memory.Peek(address.SegmentIndex, address.Offset)
}
