package memory

import (
	"errors"
	"fmt"
	"strings"

	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

type BuiltinRunner interface {
	fmt.Stringer
	CheckWrite(segment *Segment, offset uint64, value *MemoryValue) error
	InferValue(segment *Segment, offset uint64) error
	GetAllocatedSize(segmentUsedSize uint64, vmCurrentStep uint64) (uint64, error)
	GetCellsPerInstance() uint64
	GetStopPointer() uint64
	SetStopPointer(stopPointer uint64)
}

type NoBuiltin struct{}

func (b *NoBuiltin) CheckWrite(segment *Segment, offset uint64, value *MemoryValue) error {
	return nil
}

func (b *NoBuiltin) InferValue(segment *Segment, offset uint64) error {
	return errors.New("reading unknown value")
}

func (b *NoBuiltin) String() string {
	return "no builtin"
}

func (b *NoBuiltin) GetAllocatedSize(segmentUsedSize uint64, vmCurrentStep uint64) (uint64, error) {
	return 0, nil
}

func (b *NoBuiltin) GetCellsPerInstance() uint64 {
	return 0
}

func (b *NoBuiltin) GetStopPointer() uint64 {
	return 0
}

func (b *NoBuiltin) SetStopPointer(stopPointer uint64) {}

type Segment struct {
	Data []MemoryValue
	// the max index where a value was written
	LastIndex           int
	BuiltinRunner       BuiltinRunner
	PublicMemoryOffsets []PublicMemoryOffset
}

func (segment *Segment) WithBuiltinRunner(builtinRunner BuiltinRunner) *Segment {
	segment.BuiltinRunner = builtinRunner
	return segment
}

func EmptySegment() *Segment {
	// empty segments have capacity 100 as a default
	return &Segment{
		Data:          make([]MemoryValue, 0, 100),
		LastIndex:     -1,
		BuiltinRunner: &NoBuiltin{},
	}
}

func EmptySegmentWithCapacity(capacity int) *Segment {
	return &Segment{
		Data:          make([]MemoryValue, 0, capacity),
		LastIndex:     -1,
		BuiltinRunner: &NoBuiltin{},
	}
}

func EmptySegmentWithLength(length int) *Segment {
	return &Segment{
		Data:          make([]MemoryValue, length),
		LastIndex:     length - 1,
		BuiltinRunner: &NoBuiltin{},
	}
}

// returns the effective size of a segment length
// i.e the rightmost element index + 1
func (segment *Segment) Len() uint64 {
	return uint64(segment.LastIndex + 1)
}

// returns the real length that a segment has
func (segment *Segment) RealLen() uint64 {
	return uint64(len(segment.Data))
}

// Writes a new memory value to a specified offset, errors in case of overwriting a
// different memory value
func (segment *Segment) Write(offset uint64, value *MemoryValue) error {
	if offset >= segment.RealLen() {
		segment.IncreaseSegmentSize(offset + 1)
	}
	if offset >= segment.Len() {
		segment.LastIndex = int(offset)
	}

	mv := &segment.Data[offset]
	if mv.Known() && !mv.Equal(value) {
		return fmt.Errorf("rewriting value: old value: %s, new value: %s", mv, value)
	}
	segment.Data[offset] = *value
	if err := segment.BuiltinRunner.CheckWrite(segment, offset, value); err != nil {
		return fmt.Errorf("%s: %w", segment.BuiltinRunner, err)
	}

	return nil
}

// Reads a memory value from a specified offset at the segment
func (segment *Segment) Read(offset uint64) (MemoryValue, error) {
	if offset >= segment.RealLen() {
		segment.IncreaseSegmentSize(offset + 1)
	}

	mv := &segment.Data[offset]
	if !mv.Known() {
		if err := segment.BuiltinRunner.InferValue(segment, offset); err != nil {
			return UnknownValue, fmt.Errorf("%s: %w", segment.BuiltinRunner, err)
		}
	}

	if offset > segment.Len() {
		segment.LastIndex = int(offset)
	}
	return *mv, nil
}

func (segment *Segment) Peek(offset uint64) MemoryValue {
	if offset >= segment.RealLen() {
		return UnknownValue
	}
	return segment.Data[offset]
}

// Increase a segment allocated space. Panics if the new size is smaller
func (segment *Segment) IncreaseSegmentSize(newSize uint64) {
	segmentData := segment.Data
	if len(segmentData) > int(newSize) {
		panic(fmt.Sprintf(
			"cannot decrease segment size: %d -> %d",
			len(segmentData),
			newSize,
		))
	}

	var newSegmentData []MemoryValue
	if cap(segmentData) > int(newSize) {
		newSegmentData = segmentData[:cap(segmentData)]
	} else {
		newSegmentData = make([]MemoryValue, utils.Max(newSize, uint64(len(segmentData)*2)))
		copy(newSegmentData, segmentData)
	}
	segment.Data = newSegmentData
}

func (segment *Segment) Finalize(newSize uint64, publicMemoryOffsets []PublicMemoryOffset) {
	if newSize > 0 {
		segment.LastIndex = int(newSize - 1)
	}
	segment.PublicMemoryOffsets = append(segment.PublicMemoryOffsets, publicMemoryOffsets...)
}

type PublicMemoryOffset struct {
	Address uint16
	Page    uint16
}

//func (segment *Segment) String() string {
//	repr := make([]string, len(segment.Data))
//	for i, cell := range segment.Data {
//		if i < len(segment.Data)-5 {
//			continue
//		}
//		if cell.Accessed {
//			repr[i] = cell.Value.String()
//		} else {
//			repr[i] = "-"
//		}
//	}
//	return strings.Join(repr, ", ")
//}

func (segment *Segment) String() string {
	var header strings.Builder
	header.WriteString(fmt.Sprintf(
		"%s real len: %d real cap: %d len: %d\n",
		segment.BuiltinRunner,
		len(segment.Data),
		cap(segment.Data),
		segment.Len(),
	))

	for i := range segment.Data {
		if i < int(segment.Len())-5 {
			continue
		}
		if segment.Data[i].Known() {
			header.WriteString(fmt.Sprintf("[%d]-> %s\n", i, segment.Data[i].String()))
		}
	}
	return header.String()
}

// Represents the whole VM memory divided into segments
type Memory struct {
	Segments []*Segment
	// TemporarySegments is a map of temporary segments, key is the segment index, value is the segment
	TemporarySegments []*Segment
	relocationRules   map[int]MemoryAddress
}

// todo(rodro): can the amount of segments be known before hand?
func InitializeEmptyMemory() *Memory {
	return &Memory{
		// capacity 4 should be enough for the minimum amount of segments
		Segments: make([]*Segment, 0, 4),
		// allocate 1 empty temporary segment, so that the real first segment index is 1, indexed by -1 in temporary address
		TemporarySegments: []*Segment{EmptySegment()},
		relocationRules:   make(map[int]MemoryAddress),
	}
}

// Allocates a new segment providing its initial data and returns its index
func (memory *Memory) AllocateSegment(data []*f.Element) (MemoryAddress, error) {
	newSegment := EmptySegmentWithLength(len(data))
	for i := range data {
		memVal := MemoryValueFromFieldElement(data[i])
		err := newSegment.Write(uint64(i), &memVal)
		if err != nil {
			return UnknownAddress, err
		}
	}
	memory.Segments = append(memory.Segments, newSegment)
	return MemoryAddress{
		SegmentIndex: len(memory.Segments) - 1,
		Offset:       0,
	}, nil
}

// Allocates an empty segment and returns its index
func (memory *Memory) AllocateEmptySegment() MemoryAddress {
	memory.Segments = append(memory.Segments, EmptySegment())
	return MemoryAddress{
		SegmentIndex: len(memory.Segments) - 1,
		Offset:       0,
	}
}

// Allocates an empty temporary segment and returns its index
func (memory *Memory) AllocateEmptyTemporarySegment() MemoryAddress {
	memory.TemporarySegments = append(memory.TemporarySegments, EmptySegment())
	return MemoryAddress{
		SegmentIndex: -(len(memory.TemporarySegments) - 1),
		Offset:       0,
	}
}

// Allocate a Builtin segment
func (memory *Memory) AllocateBuiltinSegment(builtinRunner BuiltinRunner) MemoryAddress {
	builtinSegment := EmptySegment().WithBuiltinRunner(builtinRunner)
	memory.Segments = append(memory.Segments, builtinSegment)
	return MemoryAddress{
		SegmentIndex: len(memory.Segments) - 1,
		Offset:       0,
	}
}

// Writes to a given segment index and offset a new memory value. Errors if writing
// to an unallocated segment or if overwriting a different memory value
func (memory *Memory) Write(segmentIndex int, offset uint64, value *MemoryValue) error {
	if segmentIndex >= 0 {
		if segmentIndex >= len(memory.Segments) {
			return fmt.Errorf("segment %d: unallocated", segmentIndex)
		}
		if err := memory.Segments[segmentIndex].Write(offset, value); err != nil {
			return fmt.Errorf("segment %d, offset %d: %w", segmentIndex, offset, err)
		}
		return nil
	} else {
		segmentIndex = -segmentIndex
		if segmentIndex >= len(memory.TemporarySegments) {
			return fmt.Errorf("temporary segment %d: unallocated", segmentIndex)
		}
		if err := memory.TemporarySegments[segmentIndex].Write(offset, value); err != nil {
			return fmt.Errorf("temporary segment %d, offset %d: %w", segmentIndex, offset, err)
		}
		return nil
	}
}

// Writes to a memory address a new memory value. Errors if writing to an unallocated
// segment or if overwriting a different memory value
func (memory *Memory) WriteToAddress(address *MemoryAddress, value *MemoryValue) error {
	return memory.Write(address.SegmentIndex, address.Offset, value)
}

// Reads a memory value given the segment index and offset. Errors if reading from
// an unallocated segment or if reading an unknown memory value
func (memory *Memory) Read(segmentIndex int, offset uint64) (MemoryValue, error) {
	if segmentIndex >= 0 {
		if segmentIndex >= len(memory.Segments) {
			return MemoryValue{}, fmt.Errorf("segment %d: unallocated", segmentIndex)
		}
		mv, err := memory.Segments[segmentIndex].Read(offset)
		if err != nil {
			return MemoryValue{}, fmt.Errorf("segment %d, offset %d: %w", segmentIndex, offset, err)
		}
		return mv, nil
	} else {
		segmentIndex = -segmentIndex
		if segmentIndex >= len(memory.TemporarySegments) {
			return MemoryValue{}, fmt.Errorf("temporary segment %d: unallocated", segmentIndex)
		}
		mv, err := memory.TemporarySegments[segmentIndex].Read(offset)
		if err != nil {
			return MemoryValue{}, fmt.Errorf("temporary segment %d, offset %d: %w", segmentIndex, offset, err)
		}
		return mv, nil
	}
}

// Reads a memory value given an address. Errors if reading from
// an unallocated segment or if reading an unknown memory value
func (memory *Memory) ReadFromAddress(address *MemoryAddress) (MemoryValue, error) {
	return memory.Read(address.SegmentIndex, address.Offset)
}

// Works the same as `Read` but `MemoryValue` is converted to `Element` first
func (memory *Memory) ReadAsElement(segmentIndex int, offset uint64) (f.Element, error) {
	mv, err := memory.Read(segmentIndex, offset)
	if err != nil {
		return f.Element{}, err
	}
	felt, err := mv.FieldElement()
	if err != nil {
		return f.Element{}, err
	}
	return *felt, nil
}

// Works the same as `ReadFromAddress` but `MemoryValue` is converted to `Element` first
func (memory *Memory) ReadFromAddressAsElement(address *MemoryAddress) (f.Element, error) {
	return memory.ReadAsElement(address.SegmentIndex, address.Offset)
}

// Works the same as `Read` but `MemoryValue` is converted to `MemoryAddress` first
func (memory *Memory) ReadAsAddress(address *MemoryAddress) (MemoryAddress, error) {
	mv, err := memory.Read(address.SegmentIndex, address.Offset)
	if err != nil {
		return UnknownAddress, err
	}
	addr, err := mv.MemoryAddress()
	if err != nil {
		return UnknownAddress, err
	}
	return *addr, nil
}

// Works the same as `ReadFromAddress` but `MemoryValue` is converted to `MemoryAddress` first
func (memory *Memory) ReadFromAddressAsAddress(address *MemoryAddress) (MemoryAddress, error) {
	return memory.ReadAsAddress(address)
}

// Given a segment index and offset, returns the memory value at that position, without
// modifying it in any way. Errors if peeking from an unallocated segment
func (memory *Memory) Peek(segmentIndex int, offset uint64) (MemoryValue, error) {
	if segmentIndex >= 0 {
		if segmentIndex >= len(memory.Segments) {
			return MemoryValue{}, fmt.Errorf("segment %d: unallocated", segmentIndex)
		}
		return memory.Segments[segmentIndex].Peek(offset), nil
	} else {
		segmentIndex = -segmentIndex
		if segmentIndex >= len(memory.TemporarySegments) {
			return MemoryValue{}, fmt.Errorf("temporary segment %d: unallocated", segmentIndex)
		}
		return memory.TemporarySegments[segmentIndex].Peek(offset), nil
	}
}

// Given an address returns the memory value at that position, without
// modifying it in any way. Errors if peeking from an unallocated segment
func (memory *Memory) PeekFromAddress(address *MemoryAddress) (MemoryValue, error) {
	return memory.Peek(address.SegmentIndex, address.Offset)
}

// Given a segment index and offset returns true if the value at that address
// is known
func (memory *Memory) KnownValue(segment int, offset uint64) bool {
	if segment >= 0 {
		if segment >= len(memory.Segments) ||
			offset >= uint64(len(memory.Segments[segment].Data)) {
			return false
		}
		return memory.Segments[segment].Data[offset].Known()
	} else {
		segment = -segment
		if segment >= len(memory.TemporarySegments) ||
			offset >= uint64(len(memory.TemporarySegments[segment].Data)) {
			return false
		}
		return memory.TemporarySegments[segment].Data[offset].Known()
	}
}

// Given an address returns true if it contains a known value
func (memory *Memory) KnownValueAtAddress(address *MemoryAddress) bool {
	return memory.KnownValue(address.SegmentIndex, address.Offset)
}

// It returns all segment offsets and max memory used
func (memory *Memory) RelocationOffsets() ([]uint64, uint64) {
	// Prover expects maxMemoryUsed to start at one
	var maxMemoryUsed uint64 = 1

	// segmentsOffsets[0] = 1
	// segmentsOffsets[1] = 1 + len(segment[0])
	// segmentsOffsets[N] = 1 + len(segment[n-1]) + sum of segments[n-1-i] for i in [1, n-1]
	segmentsOffsets := make([]uint64, uint64(len(memory.Segments))+1)
	segmentsOffsets[0] = 1
	for i, segment := range memory.Segments {
		segmentLength := segment.Len()
		maxMemoryUsed += segmentLength
		segmentsOffsets[i+1] = segmentsOffsets[i] + segmentLength
	}
	return segmentsOffsets, maxMemoryUsed
}

// It finds a segment with a given builtin name, it returns the segment and true if found
func (memory *Memory) FindSegmentWithBuiltin(builtinName string) (*Segment, bool) {
	for i := range memory.Segments {
		if memory.Segments[i].BuiltinRunner.String() == builtinName {
			return memory.Segments[i], true
		}
	}
	return nil, false
}

func (memory *Memory) WriteUint256ToAddress(addr MemoryAddress, low, high *f.Element) error {
	lowMemoryValue := MemoryValueFromFieldElement(low)
	err := memory.WriteToAddress(&addr, &lowMemoryValue)
	if err != nil {
		return err
	}
	return memory.WriteToNthStructField(addr, MemoryValueFromFieldElement(high), 1)
}

func (memory *Memory) WriteToNthStructField(addr MemoryAddress, value MemoryValue, field int16) error {
	nAddr, err := addr.AddOffset(field)
	if err != nil {
		return err
	}
	return memory.WriteToAddress(&nAddr, &value)
}

func (memory *Memory) AddRelocationRule(segmentIndex int, addr MemoryAddress) {
	memory.relocationRules[segmentIndex] = addr
}

func (memory *Memory) RelocateTemporarySegments() error {
	// We check if the length of the temporary segments is 1 because the first temporary is added during initialization
	// for proper indexing, and is always empty
	if len(memory.relocationRules) == 0 || len(memory.TemporarySegments)-1 == 0 {
		return nil
	}
	for i, segment := range memory.Segments {
		for j := uint64(0); j < segment.RealLen(); j++ {
			if !segment.Data[j].Known() {
				continue
			}

			if segment.Data[j].IsAddress() {
				addr, _ := segment.Data[j].MemoryAddress()
				if addr.SegmentIndex < 0 {
					if rule, ok := memory.relocationRules[-addr.SegmentIndex]; ok {
						newAddr := MemoryAddress{SegmentIndex: rule.SegmentIndex, Offset: rule.Offset + addr.Offset}
						memory.Segments[i].Data[j] = MemoryValueFromMemoryAddress(&newAddr)
					}
				}
			}
		}
	}

	for index := 0; index < len(memory.TemporarySegments); index++ {
		baseAddr, ok := memory.relocationRules[index]
		if !ok {
			continue
		}

		dataSegment := memory.TemporarySegments[index]

		for _, cell := range dataSegment.Data {
			if cell.Known() {
				if err := memory.Write(baseAddr.SegmentIndex, baseAddr.Offset, &cell); err != nil {
					return err
				}
				baseAddr.Offset++
			}
		}
	}
	return nil
}
