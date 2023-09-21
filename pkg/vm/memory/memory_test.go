package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCellWrite(t *testing.T) {
	cell := Cell{}

	err := cell.Write(MemoryValueFromInt(1)) // Write 1 to a new cell

	assert.NoError(t, err)
	assert.True(t, cell.Accessed)
	assert.Equal(t, cell.Value, MemoryValueFromInt(1))

	//Attemp to write again to the same cell
	err = cell.Write(MemoryValueFromInt(51))
	assert.Error(t, err)
	assert.True(t, cell.Accessed)
	assert.Equal(t, cell.Value, MemoryValueFromInt(1)) //check that the value didn't change
}

func TestCellRead(t *testing.T) {
	cell := Cell{Accessed: false, Value: nil}
	assert.Equal(t, cell.Read(), EmptyMemoryValueAsFelt()) //Read from empty cell
	assert.True(t, cell.Accessed)

	cell = Cell{Accessed: false, Value: MemoryValueFromInt(51)}
	assert.Equal(t, cell.Read(), MemoryValueFromInt(51))
	assert.True(t, cell.Accessed)
}

func TestCellWriteAndRead(t *testing.T) {
	cell := Cell{}

	err := cell.Write(MemoryValueFromInt(82))

	assert.NoError(t, err)
	assert.True(t, cell.Accessed)
	assert.Equal(t, cell.Read(), MemoryValueFromInt(82))
}

func TestSegmentRead(t *testing.T) {
	segment := Segment{Data: []Cell{
		{Accessed: false, Value: MemoryValueFromInt(3)},
		{Accessed: false, Value: MemoryValueFromInt(5)},
		{Accessed: true, Value: MemoryValueFromInt(9)},
	}}

	assert.Equal(t, segment.Read(0), MemoryValueFromInt(3))
	assert.Equal(t, segment.Read(1), MemoryValueFromInt(5))
	assert.Equal(t, segment.Read(2), MemoryValueFromInt(9))
	assert.True(t, segment.Data[0].Accessed) //Segment read should mark cell as accessed
	assert.True(t, segment.Data[1].Accessed)
	assert.True(t, segment.Data[2].Accessed)

	assert.Equal(t, len(segment.Data), 3)
	//Check if we can read offsets higher than segment len
	assert.Equal(t, segment.Read(100), EmptyMemoryValueAsFelt())
	assert.Equal(t, len(segment.Data), 101) //Verify that segment len was increased
}

func TestSegmentPeek(t *testing.T) {
	segment := Segment{Data: []Cell{
		{Accessed: false, Value: MemoryValueFromInt(2)},
		{Accessed: true, Value: MemoryValueFromInt(4)},
	}}
	assert.Equal(t, segment.Peek(0).Value, MemoryValueFromInt(2))
	assert.Equal(t, segment.Peek(1).Value, MemoryValueFromInt(4))
	assert.False(t, segment.Data[0].Accessed) //Peek should not mark the cell as accessed
	assert.True(t, segment.Data[1].Accessed)  //Cell that was already accessed should stay accessed

	assert.Equal(t, len(segment.Data), 2)
	//Check if we can peek offsets higher than segment len
	assert.Equal(t, segment.Peek(30).Read(), EmptyMemoryValueAsFelt())
	assert.Equal(t, len(segment.Data), 31) //Verify that segment len was increased
}

func TestSegmentWrite(t *testing.T) {

}

func TestSegmentReadAndWrite(t *testing.T) {

}
