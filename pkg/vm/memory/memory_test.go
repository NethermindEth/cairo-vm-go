package memory

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSegmentRead(t *testing.T) {
	segment := defaultSegment(3, 5, nil)

	// read accordingly first known values
	noErrorAndEqualSegmentRead(t, &segment, 0, MemoryValueFromInt(3))
	noErrorAndEqualSegmentRead(t, &segment, 1, MemoryValueFromInt(5))

	// third value is unknown and should error
	assert.False(t, segment.Data[2].Known())
	mv, err := segment.Read(2)
	assert.Equal(t, UnknownValue, mv)
	assert.ErrorContains(t, err, "reading unknown value")

	// reading out of bound shouldn't increase the Len()
	assert.Equal(t, uint64(2), segment.Len())
	mv, err = segment.Read(100)
	assert.Equal(t, UnknownValue, mv)
	assert.ErrorContains(t, err, "reading unknown value")
	assert.Equal(t, uint64(2), segment.Len())
}

func TestSegmentPeek(t *testing.T) {
	segment := defaultSegment(2, nil)

	assert.Equal(t, MemoryValueFromInt(2), segment.Peek(0))
	assert.Equal(t, UnknownValue, segment.Peek(1))

	assert.Equal(t, uint64(1), segment.Len())
	//Check if we can peek offsets higher than segment len
	assert.Equal(t, UnknownValue, segment.Peek(30))
	assert.Equal(t, uint64(1), segment.Len()) //Verify that segment len was increased
}

func TestSegmentWrite(t *testing.T) {
	segment := defaultSegment(nil, nil)

	err := segment.Write(0, memoryValuePointerFromInt(100))
	assert.NoError(t, err)
	assert.Equal(t, MemoryValueFromInt(100), segment.Data[0])
	assert.False(t, segment.Data[1].Known())

	err = segment.Write(1, memoryValuePointerFromInt(15))
	assert.NoError(t, err)
	assert.Equal(t, MemoryValueFromInt(15), segment.Data[1])
	assert.True(t, segment.Data[1].Known())

	//Atempt to write twice
	err = segment.Write(0, memoryValuePointerFromInt(590))
	assert.Error(t, err)

	//Check that memory wasn't modified
	noErrorAndEqualSegmentRead(t, &segment, 0, MemoryValueFromInt(100))
	assert.True(t, segment.Data[0].Known())
}

func TestSegmentReadAndWrite(t *testing.T) {
	segment := defaultSegment(nil)

	err := segment.Write(0, memoryValuePointerFromInt(48))
	assert.NoError(t, err)
	noErrorAndEqualSegmentRead(t, &segment, 0, MemoryValueFromInt(48))
	assert.True(t, segment.Data[0].Known())
}

func TestIncreaseSegmentSizeSmallerSize(t *testing.T) {
	segment := defaultSegment(1, 2)
	// Panic if we decrase the size
	require.Panics(t, func() { segment.IncreaseSegmentSize(0) })
	// Panic if the size remains the same
	require.Panics(t, func() { segment.IncreaseSegmentSize(1) })
}

func TestIncreaseSegmentSizeMaxNewSize(t *testing.T) {
	segment := defaultSegment(1, 2, 3)

	segment.IncreaseSegmentSize(1000)
	assert.True(t, len(segment.Data) == 1000)
	assert.True(t, cap(segment.Data) == 1000)

	// Make sure no data was lost after incrase
	noErrorAndEqualSegmentRead(t, &segment, 0, MemoryValueFromInt(1))
	noErrorAndEqualSegmentRead(t, &segment, 1, MemoryValueFromInt(2))
	noErrorAndEqualSegmentRead(t, &segment, 2, MemoryValueFromInt(3))
}

func TestIncreaseSegmentSizeDouble(t *testing.T) {
	segment := defaultSegment(1, 2)
	segment.IncreaseSegmentSize(3)

	assert.True(t, len(segment.Data) == 4)
	assert.True(t, cap(segment.Data) == 4)

	//Make sure no data was lost after incrase
	noErrorAndEqualSegmentRead(t, &segment, 0, MemoryValueFromInt(1))
	noErrorAndEqualSegmentRead(t, &segment, 1, MemoryValueFromInt(2))
}
func TestMemoryWriteAndRead(t *testing.T) {
	memory := InitializeEmptyMemory()
	memory.AllocateEmptySegment()

	err := memory.Write(0, 0, memoryValuePointerFromInt(123))
	assert.NoError(t, err)
	val, err := memory.Read(0, 0)
	assert.NoError(t, err)
	assert.Equal(t, val, MemoryValueFromInt(123))

	//Attempt to write twice segment and offset
	err = memory.Write(0, 0, memoryValuePointerFromInt(321))
	assert.Error(t, err)

	//Attempt to write twice using address
	err = memory.WriteToAddress(&MemoryAddress{0, 0}, memoryValuePointerFromInt(542))
	assert.Error(t, err)

	//Verify data wasn't modified
	val, err = memory.Read(0, 0)
	assert.NoError(t, err)
	assert.Equal(t, val, MemoryValueFromInt(123))

	addr := MemoryAddress{0, 6}
	err = memory.WriteToAddress(&addr, memoryValuePointerFromInt(31))
	assert.NoError(t, err)
	val, err = memory.Read(0, 6)
	assert.NoError(t, err)
	assert.Equal(t, val, MemoryValueFromInt(31))
	val, err = memory.ReadFromAddress(&addr)
	assert.NoError(t, err)
	assert.Equal(t, val, MemoryValueFromInt(31))
}

func TestMemoryReadUnallocated(t *testing.T) {
	memory := InitializeEmptyMemory()
	memory.AllocateEmptySegment()
	_, err := memory.Read(1, 0)
	require.ErrorContains(t, err, "unallocated")
}

func TestMemoryPeek(t *testing.T) {
	memory := InitializeEmptyMemory()
	memory.AllocateEmptySegment()
	err := memory.Write(0, 1, memoryValuePointerFromInt(412))
	assert.NoError(t, err)

	mv, err := memory.Peek(0, 1)
	assert.NoError(t, err)
	assert.Equal(t, MemoryValueFromInt(412), mv)

	mv, err = memory.PeekFromAddress(&MemoryAddress{0, 1})
	assert.NoError(t, err)
	assert.Equal(t, MemoryValueFromInt(412), mv)
}

type testBuiltin struct{}

func (b *testBuiltin) CheckWrite(segment *Segment, offset uint64, value *MemoryValue) error {
	if offset%2 == 1 {
		return fmt.Errorf("write error")
	}
	return nil
}

func (b *testBuiltin) InferValue(segment *Segment, offset uint64) error {
	if offset%2 == 1 {
		return fmt.Errorf("infer error")
	}
	segment.Data[offset] = MemoryValueFromInt(offset)
	return nil
}

func (b *testBuiltin) String() string {
	return "test_builtin"
}

func TestSegmentBuiltin(t *testing.T) {
	segment := EmptySegment().WithBuiltinRunner(&testBuiltin{})

	err := segment.Write(100, memoryValuePointerFromInt(3))
	require.NoError(t, err)

	t.Run("inference fails", func(t *testing.T) {
		_, err := segment.Read(1)
		require.ErrorContains(t, err, "infer error")
	})
	t.Run("inference succeeds", func(t *testing.T) {
		read, err := segment.Read(2)
		require.NoError(t, err)
		assert.Equal(t, MemoryValueFromInt(2), read)
	})

	empty := EmptyMemoryValueAsFelt()
	t.Run("write check fails", func(t *testing.T) {
		err := segment.Write(3, &empty)
		require.ErrorContains(t, err, "write error")
	})
	t.Run("write check succeeds", func(t *testing.T) {
		err := segment.Write(4, &empty)
		require.NoError(t, err)
	})

	t.Run("no inference on known memory values", func(t *testing.T) {
		v, err := segment.Read(4)
		require.NoError(t, err)
		assert.Equal(t, empty, v)
	})
}

func TestSegmentsOffsets(t *testing.T) {
	memory := InitializeEmptyMemory()
	memory.AllocateEmptySegment() //Program
	memory.AllocateEmptySegment() //Execution
	memory.AllocateEmptySegment()

	err := memory.Segments[1].Write(0, memoryValuePointerFromInt(1))
	assert.NoError(t, err)
	err = memory.Segments[1].Write(1, memoryValuePointerFromInt(2))
	assert.NoError(t, err)
	err = memory.Segments[1].Write(2, memoryValuePointerFromInt(3))
	assert.NoError(t, err)
	err = memory.Segments[1].Write(3, memoryValuePointerFromInt(4))
	assert.NoError(t, err)

	err = memory.Segments[2].Write(0, memoryValuePointerFromInt(5))
	assert.NoError(t, err)
	err = memory.Segments[2].Write(1, memoryValuePointerFromInt(6))
	assert.NoError(t, err)

	// segmentsOffsets[0] = 1
	// segmentsOffsets[1] = 1
	// segmentsOffsets[2] = 1+4
	// segmentsOffsets[3] = 1+4+2
	expected_offsets := []uint64{1, 1, 5, 7}

	offsets, memoryUsed := memory.SegmentsOffsets()
	for i, v := range offsets {
		assert.Equal(t, expected_offsets[i], v)

	}
	assert.Equal(t, memoryUsed, uint64(7))
}

// compares the memory value match an expected value at the given segment and offset
func noErrorAndEqualSegmentRead(t *testing.T, s *Segment, offset uint64, expected MemoryValue) {
	v, err := s.Read(offset)
	require.NoError(t, err)
	assert.Equal(t, expected, v)
}

// creates a default segment with any given data. nil value represents unknown values
func defaultSegment(anyData ...any) Segment {
	data := make([]MemoryValue, len(anyData))
	lastIndex := -1
	var err error

	for i, any := range anyData {
		if any == nil {
			continue
		}
		data[i], err = MemoryValueFromAny(any)
		lastIndex = i
		if err != nil {
			panic(err)
		}
	}
	return Segment{
		Data:          data,
		LastIndex:     lastIndex,
		BuiltinRunner: &NoBuiltin{},
	}
}
