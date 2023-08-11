package memory

import (
	"testing"

	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/assert"
)

func TestFeltPlusFelt(t *testing.T) {
	memVal := EmptyMemoryValueAsFelt()
	lhs := MemoryValueFromFieldElement(new(f.Element).SetUint64(3))
	rhs := MemoryValueFromFieldElement(new(f.Element).SetUint64(7))

	expected := MemoryValueFromUint64(10)

	res, err := memVal.Add(lhs, rhs)
	assert.NoError(t, err)

	assert.Equal(t, memVal, res)
	assert.Equal(t, *expected, *res)
}

func TestMemoryAddressPlusFelt(t *testing.T) {
	memVal := EmptyMemoryValueAsAddress()
	lhs := MemoryValueFromMemoryAddress(&MemoryAddress{
		SegmentIndex: 2,
		Offset:       10,
	})
	rhs := MemoryValueFromFieldElement(new(f.Element).SetUint64(2))

	expected := MemoryValueFromMemoryAddress(&MemoryAddress{
		SegmentIndex: 2,
		Offset:       12,
	})

	res, err := memVal.Add(lhs, rhs)
	assert.NoError(t, err)

	assert.Equal(t, memVal, res)
	assert.Equal(t, *expected, *res)
}

func TestFeltPlusMemoryAddress(t *testing.T) {
	memVal := EmptyMemoryValueAsAddress()
	lhs := MemoryValueFromFieldElement(new(f.Element).SetUint64(2))
	rhs := MemoryValueFromMemoryAddress(&MemoryAddress{
		SegmentIndex: 2,
		Offset:       10,
	})

	expected := MemoryValueFromMemoryAddress(&MemoryAddress{
		SegmentIndex: 2,
		Offset:       12,
	})

	res, err := memVal.Add(lhs, rhs)
	assert.NoError(t, err)

	assert.Equal(t, memVal, res)
	assert.Equal(t, *expected, *res)

}

func TestMemoryAddressPlusMemoryAddress(t *testing.T) {
	memVal := EmptyMemoryValueAsAddress()
	lhs := MemoryValueFromMemoryAddress(&MemoryAddress{
		SegmentIndex: 2,
		Offset:       10,
	})
	rhs := MemoryValueFromMemoryAddress(&MemoryAddress{
		SegmentIndex: 2,
		Offset:       2,
	})
	memVal, err := memVal.Add(lhs, rhs)

	assert.Nil(t, memVal)
	assert.Error(t, err)
}

func TestFeltSubFelt(t *testing.T) {
	memVal := EmptyMemoryValueAsFelt()
	lhs := MemoryValueFromFieldElement(new(f.Element).SetUint64(8))
	rhs := MemoryValueFromFieldElement(new(f.Element).SetUint64(7))

	expected := MemoryValueFromUint64(1)

	res, err := memVal.Sub(lhs, rhs)
	assert.NoError(t, err)

	assert.Equal(t, memVal, res)
	assert.Equal(t, *expected, *res)
}

func TestMemoryAddressSubFelt(t *testing.T) {
	memVal := EmptyMemoryValueAsAddress()
	lhs := MemoryValueFromMemoryAddress(&MemoryAddress{
		SegmentIndex: 2,
		Offset:       10,
	})
	rhs := MemoryValueFromFieldElement(new(f.Element).SetUint64(2))

	expected := MemoryValueFromMemoryAddress(&MemoryAddress{
		SegmentIndex: 2,
		Offset:       8,
	})

	res, err := memVal.Sub(lhs, rhs)

	assert.NoError(t, err)
	assert.Equal(t, memVal, res)
	assert.Equal(t, *expected, *res)
}

func TestFeltSubMemoryAddress(t *testing.T) {
	memVal := EmptyMemoryValueAsAddress()
	lhs := MemoryValueFromFieldElement(new(f.Element).SetUint64(15))
	rhs := MemoryValueFromMemoryAddress(&MemoryAddress{
		SegmentIndex: 2,
		Offset:       10,
	})

	memVal, err := memVal.Sub(lhs, rhs)

	assert.Nil(t, memVal)
	assert.Error(t, err)
}

func TestMemoryAddressSubMemoryAddressSameSegment(t *testing.T) {
	memVal := EmptyMemoryValueAsAddress()
	lhs := MemoryValueFromMemoryAddress(&MemoryAddress{
		SegmentIndex: 2,
		Offset:       10,
	})
	rhs := MemoryValueFromMemoryAddress(&MemoryAddress{
		SegmentIndex: 2,
		Offset:       2,
	})
	expected := MemoryValueFromMemoryAddress(&MemoryAddress{
		SegmentIndex: 2,
		Offset:       8,
	})

	res, err := memVal.Sub(lhs, rhs)

	assert.NoError(t, err)

	assert.Equal(t, memVal, res)
	assert.Equal(t, *expected, *res)
}

func TestMemoryAddressSubMemoryAddressDiffSegment(t *testing.T) {
	memVal := EmptyMemoryValueAsAddress()
	lhs := MemoryValueFromMemoryAddress(&MemoryAddress{
		SegmentIndex: 2,
		Offset:       10,
	})
	rhs := MemoryValueFromMemoryAddress(&MemoryAddress{
		SegmentIndex: 5,
		Offset:       2,
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
