package memory

import (
	"fmt"
	"testing"

	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/require"
)

func TestMemoryRelocationWithFelt(t *testing.T) {
	// segment 0: [2, -, -, 3]
	// segment 3: [5, -, 7, -, 11, 13]
	// relocated: [2, -, -, 3, 5, -, 7, -, 11, 13]

	manager := CreateMemoryManager()
	updateMemoryWithValues(
		manager.Memory,
		[]memoryWrite{
			// segment zero
			{0, 0, uint64(2)},
			{0, 3, uint64(3)},
			// segment three
			{3, 0, uint64(5)},
			{3, 2, uint64(7)},
			{3, 4, uint64(11)},
			{3, 5, uint64(13)},
		},
	)

	res := manager.RelocateMemory()

	expected := []*f.Element{
		// segment zero
		new(f.Element).SetUint64(2),
		nil,
		nil,
		new(f.Element).SetUint64(3),
		// segment three
		new(f.Element).SetUint64(5),
		nil,
		new(f.Element).SetUint64(7),
		nil,
		new(f.Element).SetUint64(11),
		new(f.Element).SetUint64(13),
	}

	require.Equal(t, len(expected), len(res))
	require.Equal(t, expected, res)
}

func TestMemoryRelocationWithAddress(t *testing.T) {
	// segment 0: [-, 1, -, 1:5] (4)
	// segment 1: [1, 4:3, 7, -, -, 13] (10)
	// segment 2: [0:1] (11)
	// segment 3: [2:0] (12)
	// segment 4: [0:0, 1:1, 1:5, 15] (16)
	// relocated: [
	//      zero:   -,  1, -,  9,
	//      one:    1, 15, 7,  -, -, 13,
	//      two:    1,
	//      three: 10,
	//      four:   0,  5, 9, 15,
	// ]

	manager := CreateMemoryManager()
	updateMemoryWithValues(
		manager.Memory,
		[]memoryWrite{
			// segment zero
			{0, 1, uint64(1)},
			{0, 3, NewMemoryAddress(1, 5)},
			// segment one
			{1, 0, uint64(1)},
			{1, 1, NewMemoryAddress(4, 3)},
			{1, 2, uint64(7)},
			{1, 5, uint64(13)},
			// segment two
			{2, 0, NewMemoryAddress(0, 1)},
			// segment three
			{3, 0, NewMemoryAddress(2, 0)},
			// segment four
			{4, 0, NewMemoryAddress(0, 0)},
			{4, 1, NewMemoryAddress(1, 1)},
			{4, 2, NewMemoryAddress(1, 5)},
			{4, 3, uint64(15)},
		},
	)

	res := manager.RelocateMemory()

	expected := []*f.Element{
		// segment zero
		nil,
		new(f.Element).SetUint64(1),
		nil,
		new(f.Element).SetUint64(9),
		// segment one
		new(f.Element).SetUint64(1),
		new(f.Element).SetUint64(15),
		new(f.Element).SetUint64(7),
		nil,
		nil,
		new(f.Element).SetUint64(13),
		// segment two
		new(f.Element).SetUint64(1),
		// segment three
		new(f.Element).SetUint64(10),
		// segment 4
		new(f.Element).SetUint64(0),
		new(f.Element).SetUint64(5),
		new(f.Element).SetUint64(9),
		new(f.Element).SetUint64(15),
	}

	require.Equal(t, len(expected), len(res))
	require.Equal(t, expected, res)
}

type memoryWrite struct {
	SegmentIndex uint64
	Offset       uint64
	Value        any
}

func updateMemoryWithValues(memory *Memory, valuesToWrite []memoryWrite) {
	var max_segment uint64 = 0
	for _, toWrite := range valuesToWrite {
		// wrap any inside a memory value
		val, err := MemoryValueFromAny(toWrite.Value)
		if err != nil {
			panic(err)
		}

		// if the destination segment does not exist, create it
		for toWrite.SegmentIndex >= max_segment {
			max_segment += 1
			memory.AllocateEmptySegment()
		}

		fmt.Println("c")
		// write the memory val
		err = memory.Write(toWrite.SegmentIndex, toWrite.Offset, val)
		if err != nil {
			panic(err)
		}

	}
}
