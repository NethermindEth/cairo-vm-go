package memory

import (
	"fmt"
	"testing"

	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/assert"
)

func TestAddFelt(t *testing.T) {
	memVal := EmptyMemoryValue()
	lhs := MemoryValueFromMemoryAddress(&MemoryAddress{
		SegmentIndex: 2,
		Offset:       new(f.Element).SetUint64(10),
	})
	rhs := MemoryValueFromFieldElement(new(f.Element).SetUint64(2))

	expected := MemoryValueFromMemoryAddress(&MemoryAddress{
		SegmentIndex: 2,
		Offset:       new(f.Element).SetUint64(12),
	})
	fmt.Println("a-1")

	res, err := memVal.Add(lhs, rhs)

	fmt.Println("a-2")
	assert.NoError(t, err)

	fmt.Println("a0")

	assert.Equal(t, memVal, res)
	fmt.Println("a1")
	assert.Equal(t, *res, *expected)
	fmt.Println("a2")
}

func TestAddRelocatable(t *testing.T) {
	memVal := EmptyMemoryValue()
	lhs := MemoryValueFromMemoryAddress(&MemoryAddress{
		SegmentIndex: 2,
		Offset:       new(f.Element).SetUint64(10),
	})
	rhs := MemoryValueFromMemoryAddress(&MemoryAddress{
		SegmentIndex: 2,
		Offset:       new(f.Element).SetUint64(2),
	})
	memVal, err := memVal.Add(lhs, rhs)

	assert.Nil(t, memVal)
	assert.Error(t, err)
}

func TestSubFelt(t *testing.T) {
	memVal := EmptyMemoryValue()
	lhs := MemoryValueFromMemoryAddress(&MemoryAddress{
		SegmentIndex: 2,
		Offset:       new(f.Element).SetUint64(10),
	})
	rhs := MemoryValueFromFieldElement(new(f.Element).SetUint64(2))

	expected := MemoryValueFromMemoryAddress(&MemoryAddress{
		SegmentIndex: 2,
		Offset:       new(f.Element).SetUint64(8),
	})

	res, err := memVal.Sub(lhs, rhs)

	assert.NoError(t, err)
	assert.Equal(t, memVal, res)
	assert.Equal(t, *res, *expected)
}

func TestSubSameSegment(t *testing.T) {
	memVal := EmptyMemoryValue()
	lhs := MemoryValueFromMemoryAddress(&MemoryAddress{
		SegmentIndex: 2,
		Offset:       new(f.Element).SetUint64(10),
	})
	rhs := MemoryValueFromMemoryAddress(&MemoryAddress{
		SegmentIndex: 2,
		Offset:       new(f.Element).SetUint64(2),
	})
	expected := MemoryValueFromMemoryAddress(&MemoryAddress{
		SegmentIndex: 2,
		Offset:       new(f.Element).SetUint64(8),
	})

	res, err := memVal.Sub(lhs, rhs)

	assert.NoError(t, err)

	assert.Equal(t, memVal, res)
	assert.Equal(t, *res, *expected)
}

func TestSubDifferentSegment(t *testing.T) {
	memVal := EmptyMemoryValue()
	lhs := MemoryValueFromMemoryAddress(&MemoryAddress{
		SegmentIndex: 2,
		Offset:       new(f.Element).SetUint64(10),
	})
	rhs := MemoryValueFromMemoryAddress(&MemoryAddress{
		SegmentIndex: 5,
		Offset:       new(f.Element).SetUint64(2),
	})

	memVal, err := memVal.Sub(lhs, rhs)

	assert.Nil(t, memVal)
	assert.Error(t, err)
}

// Note: Leaving relocation logic for later
//func TestRelocate1(t *testing.T) {
//	r := new(MemoryAddress)
//	r1 := CreateMemoryAddress(2, new(f.Element).SetUint64(10))
//	expected := new(MemoryAddress).SetUint64(52)
//
//	res, err := r.Relocate(r1, &map[uint64]*MemoryAddress{
//		2: new(MemoryAddress).SetUint64(42),
//	})
//
//	assert.NoError(t, err)
//	assert.Equal(t, res, r)
//	assert.Equal(t, *res, *expected)
//}
//
//func TestRelocate2(t *testing.T) {
//	r := new(MemoryAddress)
//	r1 := CreateMemoryAddress(2, new(f.Element).SetUint64(10))
//	expected := CreateMemoryAddress(10, new(f.Element).SetUint64(11))
//
//	res, err := r.Relocate(r1, &map[uint64]*MemoryAddress{
//		2: CreateMemoryAddress(10, new(f.Element).SetUint64(1)),
//	})
//
//	assert.NoError(t, err)
//	assert.Equal(t, res, r)
//	assert.Equal(t, *res, *expected)
//}
//
//func TestRelocateMissingRule(t *testing.T) {
//	r := new(MemoryAddress)
//	r1 := CreateMemoryAddress(2, new(f.Element).SetUint64(10))
//
//	res, err := r.Relocate(r1, &map[uint64]*MemoryAddress{
//		3: CreateMemoryAddress(10, new(f.Element).SetUint64(1)),
//	})
//
//	assert.Error(t, err)
//	assert.Nil(t, res)
//}
