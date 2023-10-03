package memory

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/constraints"

	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/assert"
)

func UseInTestOnlyMemoryValuePointerFromInt[T constraints.Integer](v T) *MemoryValue {
	mv := MemoryValueFromInt(v)
	return &mv
}

func TestFeltPlusFelt(t *testing.T) {
	memVal := EmptyMemoryValueAsFelt()
	lhs := MemoryValueFromFieldElement(new(f.Element).SetUint64(3))
	rhs := MemoryValueFromFieldElement(new(f.Element).SetUint64(7))

	expected := MemoryValueFromInt(10)

	err := memVal.Add(&lhs, &rhs)
	require.NoError(t, err)
	assert.Equal(t, expected, memVal)
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

	err := memVal.Add(&lhs, &rhs)
	require.NoError(t, err)
	assert.Equal(t, expected, memVal)
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

	err := memVal.Add(&lhs, &rhs)
	require.NoError(t, err)
	assert.Equal(t, expected, memVal)
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
	err := memVal.Add(&lhs, &rhs)
	assert.Error(t, err)
}

func TestFeltSubFelt(t *testing.T) {
	memVal := EmptyMemoryValueAsFelt()
	lhs := MemoryValueFromFieldElement(new(f.Element).SetUint64(8))
	rhs := MemoryValueFromFieldElement(new(f.Element).SetUint64(7))

	expected := MemoryValueFromInt(1)

	err := memVal.Sub(&lhs, &rhs)
	require.NoError(t, err)
	assert.Equal(t, expected, memVal)
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

	err := memVal.Sub(&lhs, &rhs)
	require.NoError(t, err)
	assert.Equal(t, expected, memVal)
}

func TestFeltSubMemoryAddress(t *testing.T) {
	memVal := EmptyMemoryValueAsAddress()
	lhs := MemoryValueFromFieldElement(new(f.Element).SetUint64(15))
	rhs := MemoryValueFromMemoryAddress(&MemoryAddress{
		SegmentIndex: 2,
		Offset:       10,
	})

	err := memVal.Sub(&lhs, &rhs)
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

	err := memVal.Sub(&lhs, &rhs)
	require.NoError(t, err)
	assert.Equal(t, expected, memVal)
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

	err := memVal.Sub(&lhs, &rhs)
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
//	require.NoError(t, err)
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
//	require.NoError(t, err)
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
