package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSegmentRead(t *testing.T) {
	segment := Segment{Data: []MemoryValue{
		MemoryValueFromInt(3),
		MemoryValueFromInt(5),
		{},
	}}

	assert.Equal(t, segment.Read(0), MemoryValueFromInt(3))
	assert.Equal(t, segment.Read(1), MemoryValueFromInt(5))
	assert.False(t, segment.Data[2].Known())
	assert.Equal(t, segment.Read(2), EmptyMemoryValueAsFelt())
	assert.True(t, segment.Data[0].Known()) //Segment read should mark cell as accessed
	assert.True(t, segment.Data[1].Known())
	assert.True(t, segment.Data[2].Known())

	assert.Equal(t, len(segment.Data), 3)
	//Check if we can read offsets higher than segment len
	assert.Equal(t, segment.Read(100), EmptyMemoryValueAsFelt())
	assert.Equal(t, len(segment.Data), 101) //Verify that segment len was increased
}

func TestSegmentPeek(t *testing.T) {
	segment := Segment{Data: []MemoryValue{
		MemoryValueFromInt(2),
		{},
	}}
	assert.Equal(t, segment.Peek(0), MemoryValueFromInt(2))
	assert.Equal(t, segment.Peek(1), MemoryValue{})
	assert.True(t, segment.Data[0].Known())  //Cell that was already accessed should stay accessed
	assert.False(t, segment.Data[1].Known()) //Peek should not mark the cell as accessed

	assert.Equal(t, len(segment.Data), 2)
	//Check if we can peek offsets higher than segment len
	assert.Equal(t, segment.Peek(30), MemoryValue{})
	assert.Equal(t, len(segment.Data), 31) //Verify that segment len was increased
}

func TestSegmentWrite(t *testing.T) {
	segment := Segment{Data: make([]MemoryValue, 2)}

	err := segment.Write(0, UseInTestOnlyMemoryValuePointerFromInt(100))
	assert.NoError(t, err)
	assert.Equal(t, segment.Data[0], MemoryValueFromInt(100))
	assert.True(t, segment.Data[0].Known())
	assert.False(t, segment.Data[1].Known()) //Check that the other cell wasn't marked as accessed

	err = segment.Write(1, UseInTestOnlyMemoryValuePointerFromInt(15))
	assert.NoError(t, err)
	assert.Equal(t, segment.Data[1], MemoryValueFromInt(15))
	assert.True(t, segment.Data[1].Known())

	//Atempt to write twice
	err = segment.Write(0, UseInTestOnlyMemoryValuePointerFromInt(590))
	assert.Error(t, err)

	//Check that memory wasn't modified
	assert.Equal(t, segment.Read(0), MemoryValueFromInt(100))
	assert.True(t, segment.Data[0].Known())
}

func TestSegmentReadAndWrite(t *testing.T) {
	segment := Segment{Data: make([]MemoryValue, 1)}
	err := segment.Write(0, UseInTestOnlyMemoryValuePointerFromInt(48))
	assert.NoError(t, err)
	assert.Equal(t, segment.Read(0), MemoryValueFromInt(48))
	assert.True(t, segment.Data[0].Known())
}

func TestIncreaseSegmentSizeSmallerSize(t *testing.T) {
	segment := Segment{Data: []MemoryValue{
		MemoryValueFromInt(1),
		MemoryValueFromInt(2),
	}}
	// Panic if we decrase the size
	require.Panics(t, func() { segment.IncreaseSegmentSize(0) })
	// Panic if the size remains the same
	require.Panics(t, func() { segment.IncreaseSegmentSize(1) })
}

func TestIncreaseSegmentSizeMaxNewSize(t *testing.T) {
	segment := Segment{Data: []MemoryValue{
		MemoryValueFromInt(1),
		MemoryValueFromInt(2),
		MemoryValueFromInt(3),
	}}

	segment.IncreaseSegmentSize(1000)
	assert.True(t, len(segment.Data) == 1000)
	assert.True(t, cap(segment.Data) == 1000)

	//Make sure no data was lost after incrase
	assert.Equal(t, segment.Read(0), MemoryValueFromInt(1))
	assert.Equal(t, segment.Read(1), MemoryValueFromInt(2))
	assert.Equal(t, segment.Read(2), MemoryValueFromInt(3))
}

func TestIncreaseSegmentSizeDouble(t *testing.T) {
	segment := Segment{Data: []MemoryValue{
		MemoryValueFromInt(1),
		MemoryValueFromInt(2),
	}}

	segment.IncreaseSegmentSize(3)
	assert.True(t, len(segment.Data) == 4)
	assert.True(t, cap(segment.Data) == 4)

	//Make sure no data was lost after incrase
	assert.Equal(t, segment.Read(0), MemoryValueFromInt(1))
	assert.Equal(t, segment.Read(1), MemoryValueFromInt(2))
}
func TestMemoryWriteAndRead(t *testing.T) {
	memory := InitializeEmptyMemory()
	memory.AllocateEmptySegment()

	err := memory.Write(0, 0, UseInTestOnlyMemoryValuePointerFromInt(123))
	assert.NoError(t, err)
	val, err := memory.Read(0, 0)
	assert.NoError(t, err)
	assert.Equal(t, val, MemoryValueFromInt(123))

	//Attempt to write twice segment and offset
	err = memory.Write(0, 0, UseInTestOnlyMemoryValuePointerFromInt(321))
	assert.Error(t, err)

	//Attempt to write twice using address
	err = memory.WriteToAddress(&MemoryAddress{0, 0}, UseInTestOnlyMemoryValuePointerFromInt(542))
	assert.Error(t, err)

	//Verify data wasn't modified
	val, err = memory.Read(0, 0)
	assert.NoError(t, err)
	assert.Equal(t, val, MemoryValueFromInt(123))

	addr := MemoryAddress{0, 6}
	err = memory.WriteToAddress(&addr, UseInTestOnlyMemoryValuePointerFromInt(31))
	assert.NoError(t, err)
	val, err = memory.Read(0, 6)
	assert.NoError(t, err)
	assert.Equal(t, val, MemoryValueFromInt(31))
	val, err = memory.ReadFromAddress(&addr)
	assert.NoError(t, err)
	assert.Equal(t, val, MemoryValueFromInt(31))
}

func TestMemoryReadOutOfRange(t *testing.T) {
	memory := InitializeEmptyMemory()
	memory.AllocateEmptySegment()
	_, err := memory.Read(2, 2)
	assert.Error(t, err)
}

func TestMemoryPeek(t *testing.T) {
	memory := InitializeEmptyMemory()
	memory.AllocateEmptySegment()
	err := memory.Write(0, 1, UseInTestOnlyMemoryValuePointerFromInt(412))
	assert.NoError(t, err)

	cell, err := memory.Peek(0, 1)
	assert.NoError(t, err)
	assert.Equal(t, cell, MemoryValueFromInt(412))

	cell, err = memory.PeekFromAddress(&MemoryAddress{0, 1})
	assert.NoError(t, err)
	assert.Equal(t, cell, MemoryValueFromInt(412))
}
