package memory

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSegmentRead(t *testing.T) {
	segment := Segment{Data: []MemoryValue{
		MemoryValueFromInt(3),
		MemoryValueFromInt(5),
		{},
	}, BuiltinRunner: &NoBuiltin{}}

	noErrorAndEqualSegmentRead(t, &segment, 0, MemoryValueFromInt(3))
	noErrorAndEqualSegmentRead(t, &segment, 1, MemoryValueFromInt(5))
	assert.False(t, segment.Data[2].Known())
	noErrorAndEqualSegmentRead(t, &segment, 2, EmptyMemoryValueAsFelt())
	assert.True(t, segment.Data[0].Known()) //Segment read should mark cell as accessed
	assert.True(t, segment.Data[1].Known())
	assert.True(t, segment.Data[2].Known())

	assert.Equal(t, len(segment.Data), 3)
	//Check if we can read offsets higher than segment len
	noErrorAndEqualSegmentRead(t, &segment, 100, EmptyMemoryValueAsFelt())
	assert.Equal(t, len(segment.Data), 101) //Verify that segment len was increased
}

func TestSegmentPeek(t *testing.T) {
	segment := Segment{Data: []MemoryValue{
		MemoryValueFromInt(2),
		{},
	}, BuiltinRunner: &NoBuiltin{}}
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
	segment := Segment{
		Data:          make([]MemoryValue, 2),
		BuiltinRunner: &NoBuiltin{},
	}

	err := segment.Write(0, memoryValuePointerFromInt(100))
	assert.NoError(t, err)
	assert.Equal(t, segment.Data[0], MemoryValueFromInt(100))
	assert.True(t, segment.Data[0].Known())
	assert.False(t, segment.Data[1].Known()) //Check that the other cell wasn't marked as accessed

	err = segment.Write(1, memoryValuePointerFromInt(15))
	assert.NoError(t, err)
	assert.Equal(t, segment.Data[1], MemoryValueFromInt(15))
	assert.True(t, segment.Data[1].Known())

	//Atempt to write twice
	err = segment.Write(0, memoryValuePointerFromInt(590))
	assert.Error(t, err)

	//Check that memory wasn't modified
	noErrorAndEqualSegmentRead(t, &segment, 0, MemoryValueFromInt(100))
	assert.True(t, segment.Data[0].Known())
}

func TestSegmentReadAndWrite(t *testing.T) {
	segment := Segment{
		Data:          make([]MemoryValue, 1),
		BuiltinRunner: &NoBuiltin{},
	}
	err := segment.Write(0, memoryValuePointerFromInt(48))
	assert.NoError(t, err)
	noErrorAndEqualSegmentRead(t, &segment, 0, MemoryValueFromInt(48))
	assert.True(t, segment.Data[0].Known())
}

func TestIncreaseSegmentSizeSmallerSize(t *testing.T) {
	segment := Segment{
		Data: []MemoryValue{
			MemoryValueFromInt(1),
			MemoryValueFromInt(2),
		},
		BuiltinRunner: &NoBuiltin{},
	}
	// Panic if we decrase the size
	require.Panics(t, func() { segment.IncreaseSegmentSize(0) })
	// Panic if the size remains the same
	require.Panics(t, func() { segment.IncreaseSegmentSize(1) })
}

func TestIncreaseSegmentSizeMaxNewSize(t *testing.T) {
	segment := Segment{
		Data: []MemoryValue{
			MemoryValueFromInt(1),
			MemoryValueFromInt(2),
			MemoryValueFromInt(3),
		},
		BuiltinRunner: &NoBuiltin{},
	}

	segment.IncreaseSegmentSize(1000)
	assert.True(t, len(segment.Data) == 1000)
	assert.True(t, cap(segment.Data) == 1000)

	// Make sure no data was lost after incrase
	noErrorAndEqualSegmentRead(t, &segment, 0, MemoryValueFromInt(1))
	noErrorAndEqualSegmentRead(t, &segment, 1, MemoryValueFromInt(2))
	noErrorAndEqualSegmentRead(t, &segment, 2, MemoryValueFromInt(3))
}

func TestIncreaseSegmentSizeDouble(t *testing.T) {
	segment := Segment{Data: []MemoryValue{
		MemoryValueFromInt(1),
		MemoryValueFromInt(2)},
		BuiltinRunner: &NoBuiltin{},
	}

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

func TestMemoryReadOutOfRange(t *testing.T) {
	memory := InitializeEmptyMemory()
	memory.AllocateEmptySegment()
	_, err := memory.Read(2, 2)
	assert.Error(t, err)
}

func TestMemoryPeek(t *testing.T) {
	memory := InitializeEmptyMemory()
	memory.AllocateEmptySegment()
	err := memory.Write(0, 1, memoryValuePointerFromInt(412))
	assert.NoError(t, err)

	cell, err := memory.Peek(0, 1)
	assert.NoError(t, err)
	assert.Equal(t, cell, MemoryValueFromInt(412))

	cell, err = memory.PeekFromAddress(&MemoryAddress{0, 1})
	assert.NoError(t, err)
	assert.Equal(t, cell, MemoryValueFromInt(412))
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
		return fmt.Errorf("deduce error")
	}
	segment.Data[offset] = MemoryValueFromInt(offset)
	return nil
}

func TestSegmentBuiltin(t *testing.T) {
	segment := EmptySegment().WithBuiltinRunner(&testBuiltin{})

	t.Run("deduction fails", func(t *testing.T) {
		_, err := segment.Read(1)
		require.EqualError(t, err, "deduce error")
	})
	t.Run("deduction succeeds", func(t *testing.T) {
		read, err := segment.Read(2)
		require.NoError(t, err)
		assert.Equal(t, MemoryValueFromInt(2), read)
	})

	empty := EmptyMemoryValueAsFelt()
	t.Run("write check fails", func(t *testing.T) {
		err := segment.Write(3, &empty)
		require.EqualError(t, err, "write error")
	})
	t.Run("write check succeeds", func(t *testing.T) {
		err := segment.Write(4, &empty)
		require.NoError(t, err)
	})

	t.Run("no deduction on known cells", func(t *testing.T) {
		v, err := segment.Read(4)
		require.NoError(t, err)
		assert.Equal(t, empty, v)
	})
}

func noErrorAndEqualSegmentRead(t *testing.T, s *Segment, offset uint64, expected MemoryValue) {
	v, err := s.Read(offset)
	require.NoError(t, err)
	assert.Equal(t, expected, v)
}
