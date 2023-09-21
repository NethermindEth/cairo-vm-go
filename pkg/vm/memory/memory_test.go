package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrite(t *testing.T) {
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

func TestRead(t *testing.T) {
	cell := Cell{Accessed: false, Value: nil}
	assert.Equal(t, cell.Read(), EmptyMemoryValueAsFelt()) //Read from empty cell
	assert.True(t, cell.Accessed)

	cell = Cell{Accessed: false, Value: MemoryValueFromInt(51)}
	assert.Equal(t, cell.Read(), MemoryValueFromInt(51))
	assert.True(t, cell.Accessed)
}

func TestWriteAndRead(t *testing.T) {
	cell := Cell{}

	err := cell.Write(MemoryValueFromInt(82))

	assert.NoError(t, err)
	assert.True(t, cell.Accessed)
	assert.Equal(t, cell.Read(), MemoryValueFromInt(82))

}
