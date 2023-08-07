package memory

import (
	"testing"

	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/assert"
)

func TestAddFelt(t *testing.T) {
	r := new(MemoryAddress)
	r1 := CreateMemoryAddress(2, new(f.Element).SetUint64(10))
	r2 := new(MemoryAddress).SetUint64(2)
	expected := CreateMemoryAddress(2, new(f.Element).SetUint64(12))

	res, err := r.Add(r1, r2)

	assert.NoError(t, err)

	assert.Equal(t, r, res)
	assert.Equal(t, *res, *expected)
}

func TestAddRelocatable(t *testing.T) {
	r := new(MemoryAddress)
	r1 := CreateMemoryAddress(2, new(f.Element).SetUint64(10))
	r2 := CreateMemoryAddress(2, new(f.Element).SetUint64(2))

	r, err := r.Add(r1, r2)

	assert.Nil(t, r)
	assert.Error(t, err)
}

func TestSubFelt(t *testing.T) {
	r := new(MemoryAddress)
	r1 := CreateMemoryAddress(2, new(f.Element).SetUint64(10))
	r2 := new(MemoryAddress).SetUint64(2)
	expected := CreateMemoryAddress(2, new(f.Element).SetUint64(8))

	res, err := r.Sub(r1, r2)

	assert.NoError(t, err)

	assert.Equal(t, r, res)
	assert.Equal(t, *res, *expected)
}

func TestSubSameSegment(t *testing.T) {
	r := new(MemoryAddress)
	r1 := CreateMemoryAddress(2, new(f.Element).SetUint64(10))
	r2 := CreateMemoryAddress(2, new(f.Element).SetUint64(2))
	expected := new(MemoryAddress).SetUint64(8)

	res, err := r.Sub(r1, r2)

	assert.NoError(t, err)

	assert.Equal(t, r, res)
	assert.Equal(t, *res, *expected)
}

func TestSubDifferentSegment(t *testing.T) {
	r := new(MemoryAddress)
	r1 := CreateMemoryAddress(2, new(f.Element).SetUint64(10))
	r2 := CreateMemoryAddress(1, new(f.Element).SetUint64(2))

	r, err := r.Sub(r1, r2)

	assert.Nil(t, r)
	assert.Error(t, err)
}

func TestRelocate1(t *testing.T) {
	r := new(MemoryAddress)
	r1 := CreateMemoryAddress(2, new(f.Element).SetUint64(10))
	expected := new(MemoryAddress).SetUint64(52)

	res, err := r.Relocate(r1, &map[uint64]*MemoryAddress{
		2: new(MemoryAddress).SetUint64(42),
	})

	assert.NoError(t, err)
	assert.Equal(t, res, r)
	assert.Equal(t, *res, *expected)
}

func TestRelocate2(t *testing.T) {
	r := new(MemoryAddress)
	r1 := CreateMemoryAddress(2, new(f.Element).SetUint64(10))
	expected := CreateMemoryAddress(10, new(f.Element).SetUint64(11))

	res, err := r.Relocate(r1, &map[uint64]*MemoryAddress{
		2: CreateMemoryAddress(10, new(f.Element).SetUint64(1)),
	})

	assert.NoError(t, err)
	assert.Equal(t, res, r)
	assert.Equal(t, *res, *expected)
}

func TestRelocateMissingRule(t *testing.T) {
	r := new(MemoryAddress)
	r1 := CreateMemoryAddress(2, new(f.Element).SetUint64(10))

	res, err := r.Relocate(r1, &map[uint64]*MemoryAddress{
		3: CreateMemoryAddress(10, new(f.Element).SetUint64(1)),
	})

	assert.Error(t, err)
	assert.Nil(t, res)
}
